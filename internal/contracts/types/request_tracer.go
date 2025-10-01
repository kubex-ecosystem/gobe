package types

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	l "github.com/kubex-ecosystem/logz"
)

const (
	RequestLimit  = 100
	RequestWindow = 1 * time.Minute

	scanInitBuf  = 256 << 10 // 256KiB
	scanMaxBuf   = 10 << 20  // 10MiB por linha
	tmpPrefix    = ".rt-*.tmp"
	metaLineJSON = `{"meta":{"format":"ndjson","version":1}}`
)

var (
	// TODO: matar este cache global e injetar *RequestTracers onde precisar.
	requestTracers     *RequestTracers
	requestTracersOnce sync.Once
)

type RequestTracers struct {
	gobe    ci.IGoBE
	mu      sync.RWMutex
	tracers map[string]ci.IRequestsTracer // key: IP (mantive tua chave)
}

func NewRequestTracers(g ci.IGoBE) ci.IRequestTracers {
	return ensureGlobal(g)
}

type RequestsTracer struct {
	Mutexes       ci.IMutexes `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
	IP            string      `json:"ip" yaml:"ip" xml:"ip" toml:"ip" gorm:"ip"`
	Port          string      `json:"port" yaml:"port" xml:"port" toml:"port" gorm:"port"`
	LastUserAgent string      `json:"last_user_agent" yaml:"last_user_agent" xml:"last_user_agent" toml:"last_user_agent" gorm:"last_user_agent"`
	UserAgents    []string    `json:"user_agents" yaml:"user_agents" xml:"user_agents" toml:"user_agents" gorm:"user_agents"`
	Endpoint      string      `json:"endpoint" yaml:"endpoint" xml:"endpoint" toml:"endpoint" gorm:"endpoint"`
	Method        string      `json:"method" yaml:"method" xml:"method" toml:"method" gorm:"method"`
	TimeList      []time.Time `json:"time_list" yaml:"time_list" xml:"time_list" toml:"time_list" gorm:"time_list"`
	Count         int         `json:"count" yaml:"count" xml:"count" toml:"count" gorm:"count"`
	Valid         bool        `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
	Error         error       `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
	requestWindow time.Duration
	requestLimit  int
	filePath      string
	oldFilePath   string

	Mapper ci.IMapper[ci.IRequestsTracer] `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
}

// ----- helpers -----

func ensureGlobal(g ci.IGoBE) *RequestTracers {
	requestTracersOnce.Do(func() {
		requestTracers = &RequestTracers{
			gobe:    g,
			tracers: make(map[string]ci.IRequestsTracer),
		}
	})
	return requestTracers
}

func defaultFileIfEmpty(p string) string {
	if strings.TrimSpace(p) != "" {
		return p
	}
	abs, err := filepath.Abs(filepath.Join(".", "requests_tracer.json"))
	if err != nil {
		gl.Log("error", fmt.Sprintf("error resolving default path: %v", err))
		return ""
	}
	return abs
}

func slidingWindow(times []time.Time, now time.Time, window time.Duration) ([]time.Time, int) {
	cut := now.Add(-window)
	// avança início até first >= cut
	i := 0
	for i < len(times) && times[i].Before(cut) {
		i++
	}
	if i > 0 {
		times = times[i:]
	}
	return times, len(times)
}

// ----- ctor / registry -----

func newRequestsTracer(g ci.IGoBE, ip, port, endpoint, method, userAgent, filePath string) *RequestsTracer {
	reg := ensureGlobal(g)

	// estado default
	if requestTracers == nil {
		gl.Log("error", "global registry not initialized (unexpected)")
	}

	reg.mu.Lock()
	defer reg.mu.Unlock()

	if t, ok := reg.tracers[ip]; ok {
		// update existente
		tracer, ok2 := t.(*RequestsTracer)
		if !ok2 {
			gl.Log("error", fmt.Sprintf("cast to *RequestsTracer failed for ip=%s", ip))
			return nil
		}

		now := time.Now()
		tracer.TimeList = append(tracer.TimeList, now)
		tracer.TimeList, tracer.Count = slidingWindow(tracer.TimeList, now, tracer.requestWindow)
		tracer.LastUserAgent = userAgent
		tracer.UserAgents = append(tracer.UserAgents, userAgent)

		if tracer.Count > tracer.requestLimit {
			tracer.Valid = false
			tracer.Error = fmt.Errorf("request limit exceeded for IP %s: %d>%d", ip, tracer.Count, tracer.requestLimit)
			gl.Log("info", tracer.Error.Error())
		} else {
			tracer.Valid = true
			tracer.Error = nil
		}

		fp := defaultFileIfEmpty(filePath)
		if tracer.filePath != fp {
			tracer.oldFilePath = tracer.filePath
			tracer.filePath = fp
		}

		// (re)configura Mapper
		rtIfc := ci.IRequestsTracer(tracer)
		tracer.Mapper = NewMapper(&rtIfc, tracer.filePath)
		reg.tracers[ip] = tracer
		return tracer
	}

	// novo tracer
	fp := defaultFileIfEmpty(filePath)
	now := time.Now()
	tracer := &RequestsTracer{
		IP:            ip,
		Port:          port,
		LastUserAgent: userAgent,
		UserAgents:    []string{userAgent},
		Endpoint:      endpoint,
		Method:        method,
		TimeList:      []time.Time{now},
		Count:         1,
		Valid:         true,
		Error:         nil,
		Mutexes:       NewMutexesType(),
		filePath:      fp,
		oldFilePath:   "",
		requestWindow: RequestWindow,
		requestLimit:  RequestLimit,
	}
	rtIfc := ci.IRequestsTracer(tracer)
	tracer.Mapper = NewMapper(&rtIfc, tracer.filePath)

	reg.tracers[ip] = tracer
	return tracer
}

func NewRequestsTracerType(g ci.IGoBE, ip, port, endpoint, method, userAgent, filePath string) ci.IRequestsTracer {
	return newRequestsTracer(g, ip, port, endpoint, method, userAgent, filePath)
}
func NewRequestsTracer(g ci.IGoBE, ip, port, endpoint, method, userAgent, filePath string) ci.IRequestsTracer {
	return newRequestsTracer(g, ip, port, endpoint, method, userAgent, filePath)
}

// ----- getters/setters -----

func (r *RequestsTracer) Mu() ci.IMutexes          { return r.Mutexes }
func (r *RequestsTracer) GetIP() string            { return r.IP }
func (r *RequestsTracer) GetPort() string          { return r.Port }
func (r *RequestsTracer) GetLastUserAgent() string { return r.LastUserAgent }
func (r *RequestsTracer) GetUserAgents() []string  { return r.UserAgents }
func (r *RequestsTracer) GetEndpoint() string      { return r.Endpoint }
func (r *RequestsTracer) GetMethod() string        { return r.Method }
func (r *RequestsTracer) GetTimeList() []time.Time { return r.TimeList }
func (r *RequestsTracer) GetCount() int            { return r.Count }
func (r *RequestsTracer) GetError() error          { return r.Error }
func (r *RequestsTracer) GetMutexes() ci.IMutexes  { return r.Mutexes }
func (r *RequestsTracer) IsValid() bool            { return r.Valid }

func (r *RequestsTracer) GetOldFilePath() string {
	if r.oldFilePath == "" {
		r.oldFilePath = defaultFileIfEmpty("")
	}
	return r.oldFilePath
}
func (r *RequestsTracer) GetFilePath() string { return r.filePath }
func (r *RequestsTracer) SetFilePath(filePath string) {
	r.filePath = defaultFileIfEmpty(filePath)
}
func (r *RequestsTracer) GetMapper() ci.IMapper[ci.IRequestsTracer] { return r.Mapper }
func (r *RequestsTracer) SetMapper(mapper ci.IMapper[ci.IRequestsTracer]) {
	if mapper == nil {
		gl.Log("error", "Mapper cannot be nil")
		return
	}
	r.Mapper = mapper
}
func (r *RequestsTracer) GetRequestWindow() time.Duration { return r.requestWindow }
func (r *RequestsTracer) SetRequestWindow(window time.Duration) {
	if window <= 0 {
		gl.Log("error", "Request window cannot be <= 0")
		return
	}
	r.requestWindow = window
}
func (r *RequestsTracer) GetRequestLimit() int { return r.requestLimit }
func (r *RequestsTracer) SetRequestLimit(limit int) {
	if limit <= 0 {
		gl.Log("error", "Request limit cannot be <= 0")
		return
	}
	r.requestLimit = limit
}

// ----- load / update / utils (NDJSON) -----

func LoadRequestsTracerFromFile(g ci.IGoBE) (ci.IRequestTracers, error) {
	reg := ensureGlobal(g)

	path := g.GetLogFilePath()
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		// cria arquivo vazio
		if err := os.WriteFile(path, nil, 0644); err != nil {
			return nil, err
		}
		return reg, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, scanInitBuf), scanMaxBuf)

	loaded := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line == metaLineJSON {
			continue
		}
		var rt RequestsTracer
		if err := json.Unmarshal([]byte(line), &rt); err != nil {
			gl.Log("warn", fmt.Sprintf("invalid line (ignored): %v", err))
			continue
		}
		reg.mu.Lock()
		reg.tracers[rt.IP] = &rt
		reg.mu.Unlock()
		loaded++
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	gl.Log("info", fmt.Sprintf("Loaded %d request tracers", loaded))
	return reg, nil
}

func updateRequestTracer(g ci.IGoBE, updatedTracer ci.IRequestsTracer) error {
	inPath := defaultFileIfEmpty(updatedTracer.GetFilePath())
	dir := filepath.Dir(inPath)

	in, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("open in: %w", err)
	}
	defer in.Close()

	tmp, err := os.CreateTemp(dir, tmpPrefix)
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}
	tmpPath := tmp.Name()
	bw := bufio.NewWriterSize(tmp, 64<<10)
	defer func() { tmp.Close(); _ = os.Remove(tmpPath) }()

	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, scanInitBuf), scanMaxBuf)

	// escreve cabeçalho meta no novo arquivo (opcional, mas deixa padrão)
	_, _ = bw.WriteString(metaLineJSON + "\n")

	replaced := false

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line == metaLineJSON {
			continue
		}
		var ex RequestsTracer
		if json.Unmarshal([]byte(line), &ex) == nil && ex.IP == updatedTracer.GetIP() && ex.Port == updatedTracer.GetPort() {
			b, _ := json.Marshal(updatedTracer)
			if _, err := bw.Write(append(b, '\n')); err != nil {
				return err
			}
			replaced = true
		} else {
			if _, err := bw.Write(append([]byte(line), '\n')); err != nil {
				return err
			}
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}

	if !replaced {
		b, _ := json.Marshal(updatedTracer)
		if _, err := bw.Write(append(b, '\n')); err != nil {
			return err
		}
	}

	if err := bw.Flush(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, inPath); err != nil {
		return err
	}
	return nil
}

func isDuplicateRequest(g ci.IGoBE, rt ci.IRequestsTracer, logger l.Logger) bool {
	path := defaultFileIfEmpty(rt.GetFilePath())
	f, err := os.Open(path)
	if err != nil {
		gl.Log("error", fmt.Sprintf("open file: %v", err))
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, scanInitBuf), scanMaxBuf)

	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 || bytes.Equal(line, []byte(metaLineJSON)) {
			continue
		}
		var ex RequestsTracer
		if json.Unmarshal(line, &ex) != nil {
			continue
		}
		if ex.IP == rt.GetIP() && ex.Port == rt.GetPort() {
			return true
		}
	}
	return false
}

// mantém compat, mas agora regrava o arquivo inteiro de forma segura
func updateRequestTracerInMemory(updated ci.IRequestsTracer) error {
	path := defaultFileIfEmpty(updated.GetFilePath())
	dir := filepath.Dir(path)

	in, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer in.Close()

	tmp, err := os.CreateTemp(dir, tmpPrefix)
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}
	tmpPath := tmp.Name()
	bw := bufio.NewWriterSize(tmp, 64<<10)
	defer func() { tmp.Close(); _ = os.Remove(tmpPath) }()

	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, scanInitBuf), scanMaxBuf)

	_, _ = bw.WriteString(metaLineJSON + "\n")

	wrote := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line == metaLineJSON {
			continue
		}
		var ex RequestsTracer
		if json.Unmarshal([]byte(line), &ex) == nil && ex.IP == updated.GetIP() && ex.Port == updated.GetPort() {
			b, _ := json.Marshal(updated)
			if _, err := bw.Write(append(b, '\n')); err != nil {
				return err
			}
			wrote = true
		} else {
			if _, err := bw.Write(append([]byte(line), '\n')); err != nil {
				return err
			}
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if !wrote {
		b, _ := json.Marshal(updated)
		if _, err := bw.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// ----- middleware (mantive desligado por padrão) -----

func (r *RequestTracers) RequestsTracerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// descomente e adeque quando quiser religar:

		// ip := c.ClientIP()
		// port := c.Request.URL.Port()
		// endpoint := c.Request.URL.Path
		// method := c.Request.Method
		// userAgent := c.Request.UserAgent()
		// filePath := r.gobe.GetLogFilePath()
		//
		// if ip == "" || endpoint == "" || method == "" || userAgent == "" {
		// 	gl.Log("error", "invalid request data for RequestTracerMiddleware")
		// 	c.Next()
		// 	return
		// }
		//
		// tracer := NewRequestsTracerType(r.gobe, ip, port, endpoint, method, userAgent, filePath)
		// if isDuplicateRequest(r.gobe, tracer, logger.GetLogger[*RequestsTracer](nil).GetLogger()) {
		// 	gl.Log("info", fmt.Sprintf("duplicate request detected ip=%s port=%s", ip, port))
		// 	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		// 	return
		// }
		//
		// c.Set("requestTracer", tracer)
		// c.Next()
		//
		// if err := updateRequestTracer(r.gobe, tracer); err != nil {
		// 	gl.Log("error", fmt.Sprintf("update tracer: %v", err))
		// }

		c.Next()
	}
}

// ----- registry ops -----

func (r *RequestTracers) GetRequestTracers() map[string]ci.IRequestsTracer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]ci.IRequestsTracer, len(r.tracers))
	for k, v := range r.tracers {
		out[k] = v
	}
	return out
}
func (r *RequestTracers) SetRequestTracers(tracers map[string]ci.IRequestsTracer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracers = tracers
}
func (r *RequestTracers) AddRequestTracer(name string, tracer ci.IRequestsTracer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracers[name] = tracer
}
func (r *RequestTracers) GetRequestTracer(name string) (ci.IRequestsTracer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tracer, ok := r.tracers[name]
	return tracer, ok
}
func (r *RequestTracers) RemoveRequestTracer(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tracers, name)
}

// ----- tiny local util for meta line compare -----

func isMetaLine(b []byte) bool {
	return bytes.Equal(bytes.TrimSpace(b), []byte(metaLineJSON))
}

// (não usado aqui, mas deixei caso precise um dia)
func writeMetaLineIfEmpty(path string) {
	fi, err := os.Stat(path)
	if err == nil && fi.Size() > 0 {
		return
	}
	_ = os.WriteFile(path, []byte(metaLineJSON+"\n"), 0644)
}

// quick inline to avoid import cycle if someone wants to emit 429 here
func tooManyRequests(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": msg})
}
