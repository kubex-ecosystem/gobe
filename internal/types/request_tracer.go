package types

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

const RequestLimit = 5
const RequestWindow = 60 * time.Second

var requestTracers *RequestTracers

type RequestTracers struct {
	gobe    ci.IGoBE
	tracers map[string]ci.IRequestsTracer
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
	Mapper        ci.IMapper[ci.IRequestsTracer] `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
}

func newRequestsTracer(g ci.IGoBE, ip, port, endpoint, method, userAgent, filePath string) *RequestsTracer {
	var tracer *RequestsTracer
	var exists bool

	if requestTracers == nil {
		requestTracers = NewRequestTracers(g)
	}
	var tracerT ci.IRequestsTracer
	var ok bool
	if tracerT, exists = requestTracers.tracers[ip]; exists {
		tracer, ok = tracerT.(*RequestsTracer)
		if !ok {
			gl.Log("error", fmt.Sprintf("Error casting tracer to RequestsTracer for IP: %s", ip))
			return nil
		}

		//tracer.GetMutexes().MuLock()
		//defer tracer.GetMutexes().MuUnlock()

		tracer.Count++
		tracer.TimeList = append(tracer.TimeList, time.Now())
		tracer.LastUserAgent = userAgent
		tracer.UserAgents = append(tracer.UserAgents, userAgent)

		if len(tracer.TimeList) > 1 {
			if tracer.TimeList[len(tracer.TimeList)-1].Sub(tracer.TimeList[len(tracer.TimeList)-2]) <= RequestWindow {
				gl.Log("info", fmt.Sprintf("Request limit exceeded for IP: %s, Count: %d", tracer.IP, tracer.Count))
				tracer.Valid = false
				tracer.Error = fmt.Errorf("request limit exceeded for IP: %s, Count: %d", tracer.IP, tracer.Count)
			} else if tracer.TimeList[len(tracer.TimeList)-1].Sub(tracer.TimeList[0]) > RequestWindow {
				gl.Log("info", fmt.Sprintf("Request window exceeded for IP: %s, Count: %d", tracer.IP, tracer.Count))
				tracer.Count = 1
				tracer.TimeList = []time.Time{tracer.TimeList[len(tracer.TimeList)-1]}
				tracer.UserAgents = []string{userAgent}
				tracer.Valid = true
				tracer.Error = nil
			} else if tracer.Count > RequestLimit {
				gl.Log("info", fmt.Sprintf("Request limit exceeded for IP: %s, Count: %d", tracer.IP, tracer.Count))
				tracer.Valid = false
				tracer.Error = fmt.Errorf("request limit exceeded for IP: %s, Count: %d", tracer.IP, tracer.Count)
			} else {
				gl.Log("info", fmt.Sprintf("Request limit not exceeded for IP: %s, Count: %d", tracer.IP, tracer.Count))
				tracer.Valid = true
				tracer.Error = nil
			}
		}
		if tracer.filePath != filePath {
			gl.Log("info", fmt.Sprintf("File path changed for IP: %s, Count: %d", tracer.IP, tracer.Count))
			tracer.oldFilePath = tracer.filePath
			tracer.filePath = filePath
		}
	} else {
		tracer = &RequestsTracer{
			IP:            ip,
			Port:          port,
			LastUserAgent: userAgent,
			UserAgents:    []string{userAgent},
			Endpoint:      endpoint,
			Method:        method,
			TimeList:      []time.Time{time.Now()},
			Count:         1,
			Valid:         true,
			Error:         nil,
			Mutexes:       NewMutexesType(),
			filePath:      filePath,
			oldFilePath:   "",

			requestWindow: RequestWindow,
			requestLimit:  RequestLimit,
		}
	}

	requestTracers.tracers[ip] = tracer
	rTracer := ci.IRequestsTracer(tracer)

	tracer.Mapper = NewMapperType(&rTracer, tracer.filePath)

	//tracer.Mutexes.MuAdd(1)
	//go func(tracer *RequestsTracer) {
	//	defer tracer.Mutexes.MuDone()
	//	tracer.Mapper.SerializeToFile("json")
	//}(tracer)
	//tracer.Mutexes.MuWait()

	return tracer
}
func NewRequestsTracerType(g ci.IGoBE, ip, port, endpoint, method, userAgent, filePath string) ci.IRequestsTracer {
	return newRequestsTracer(g, ip, port, endpoint, method, userAgent, filePath)
}
func NewRequestsTracer(g ci.IGoBE, ip, port, endpoint, method, userAgent, filePath string) ci.IRequestsTracer {
	return newRequestsTracer(g, ip, port, endpoint, method, userAgent, filePath)
}

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
		abs, err := filepath.Abs(filepath.Join("./", "requests_tracer.json"))
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting absolute path: %v", err))
			return ""
		}
		r.oldFilePath = abs
	}
	return r.oldFilePath
}
func (r *RequestsTracer) GetFilePath() string { return r.filePath }
func (r *RequestsTracer) SetFilePath(filePath string) {
	if filePath == "" {
		abs, err := filepath.Abs(filepath.Join("./", "requests_tracer.json"))
		if err != nil {
			gl.Log("error", fmt.Sprintf("Error getting absolute path: %v", err))
			return
		}
		r.filePath = abs
	} else {
		r.filePath = filePath
	}
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
		gl.Log("error", "Request window cannot be negative or zero")
		return
	}
	r.requestWindow = window
}
func (r *RequestsTracer) GetRequestLimit() int { return r.requestLimit }
func (r *RequestsTracer) SetRequestLimit(limit int) {
	if limit <= 0 {
		gl.Log("error", "Request limit cannot be negative or zero")
		return
	}
	r.requestLimit = limit
}

func LoadRequestsTracerFromFile(g ci.IGoBE) (map[string]ci.IRequestsTracer, error) {
	if requestTracers == nil {
		requestTracers = NewRequestTracers(g)
	}
	if len(requestTracers.tracers) == 0 {
		requestTracers.tracers = make(map[string]ci.IRequestsTracer)
	}

	gl.Log("info", "Loading request tracers from file")
	if _, err := os.Stat(g.GetLogFilePath()); os.IsNotExist(err) {
		gl.Log("warn", fmt.Sprintf("File does not exist: %v, creating new file", err.Error()))
		if _, createErr := os.Create(g.GetLogFilePath()); createErr != nil {
			gl.Log("error", fmt.Sprintf("Error creating file: %v", createErr.Error()))
			return nil, createErr
		} else {
			gl.Log("info", "File created successfully")
		}
		return nil, nil
	}

	gl.Log("info", "File exists, proceeding to load")
	inputFile, err := os.Open(g.GetLogFilePath())
	if err != nil {
		gl.Log("error", "Erro ao abrir arquivo: %v", err.Error())
		return nil, err
	}

	defer func(inputFile *os.File) {
		gl.Log("info", "Closing input file")
		if closeErr := inputFile.Close(); closeErr != nil {
			gl.Log("error", fmt.Sprintf("Erro ao fechar arquivo: %v", err))
		}
	}(inputFile)

	reader := bufio.NewReader(inputFile)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	//g.Mu().MuAdd(1)
	go func(g ci.IGoBE) {
		//defer g.Mu().MuDone()
		gl.Log("info", "Decoding request tracers from file")
		for decoder.More() {
			var existing *RequestsTracer
			if err := decoder.Decode(&existing); err != nil || existing == nil {
				if err == nil {
					err = fmt.Errorf("existing não inicializado: %v, err: %s", existing, err)
				}
				gl.Log("error", fmt.Sprintf("Erro ao decodificar:%s", err.Error()))
				continue
			}
			gl.Log("info", fmt.Sprintf("Decoded request tracer: %s", existing.IP))
			requestTracers.tracers[existing.IP] = existing
		}
	}(g)

	gl.Log("info", "Waiting for decoding to finish")
	//g.Mu().MuWait()

	if len(requestTracers.tracers) > 0 {
		gl.Log("info", fmt.Sprintf("Loaded %d request tracers", len(requestTracers.tracers)))
	} else {
		gl.Log("warn", "No request tracers loaded from file")
	}

	return requestTracers.tracers, nil
}
func updateRequestTracer(g ci.IGoBE, updatedTracer ci.IRequestsTracer) error {
	var decoder *json.Decoder
	var outputFile *os.File
	var err error
	tmpFilePath := filepath.Join(g.GetConfigFilePath(), "temp"+updatedTracer.GetFilePath())

	if inputFile, inputFileErr := os.Open(updatedTracer.GetFilePath()); inputFileErr != nil || inputFile == nil {
		if inputFileErr == nil {
			inputFileErr = fmt.Errorf("inputFile não inicializado")
		}
		return fmt.Errorf("erro ao abrir arquivo: %v", inputFileErr)
	} else {
		defer func(inputFile *os.File) {
			_ = inputFile.Close()
		}(inputFile)

		if outputFile, err = os.Create(tmpFilePath); err != nil || outputFile == nil {
			if err == nil {
				err = fmt.Errorf("outputFile não inicializado")
			}
			return fmt.Errorf("erro ao criar arquivo temporário: %v", err)
		} else {
			defer func(outputFile *os.File, tmpFilePath string) {
				_ = outputFile.Close()
				if removeErr := os.Remove(tmpFilePath); removeErr != nil {
					gl.Log("error", fmt.Sprintf("Erro ao remover arquivo temporário: %v", removeErr))
					return
				}
			}(outputFile, tmpFilePath)

			decoder = json.NewDecoder(inputFile)
			decoder.DisallowUnknownFields()

			var existing *RequestsTracer
			for decoder.More() {
				existing = &RequestsTracer{}
				if err = decoder.Decode(&existing); err != nil || existing == nil {
					if err == nil {
						err = fmt.Errorf("existing não inicializado")
					}

					gl.Log("error", fmt.Sprintf("Erro ao decodificar linha: %v", err))

					continue
				} else {
					var line []byte

					// If the existing tracer matches the updated tracer, update it
					if existing.IP == updatedTracer.GetIP() && existing.Port == updatedTracer.GetPort() {
						lineBytes, _ := json.Marshal(updatedTracer)
						line = []byte(string(lineBytes) + "\n")
					}

					// Escreve a linha no novo arquivo, em array de bytes (que seria "bufferizado" e mais rápido)
					if _, writeErr := outputFile.Write(line); writeErr != nil {
						return writeErr
					}
				}
			}
		}
	}
	if _, tmpFilePathStatErr := os.Stat(tmpFilePath); tmpFilePathStatErr != nil {
		if replaceErr := os.Rename(tmpFilePath, updatedTracer.GetFilePath()); replaceErr != nil {
			return replaceErr
		}
	}
	return nil
}
func isDuplicateRequest(g ci.IGoBE, rt ci.IRequestsTracer, logger l.Logger) bool {
	env := g.Environment()
	strategy := screeningByRAMSize(env, rt.GetFilePath())

	if strategy == "strings" {
		data, err := os.ReadFile(rt.GetFilePath())
		if err != nil {
			gl.Log("error", fmt.Sprintf("Erro ao ler arquivo: %v", err))
			return false
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.Contains(line, rt.GetIP()) && strings.Contains(line, rt.GetPort()) {
				return true
			}
		}
	} else {
		f, err := os.Open(rt.GetFilePath())
		if err != nil {
			gl.Log("error", fmt.Sprintf("Erro ao abrir arquivo: %v", err))
			return false
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var existing RequestsTracer
			if err := json.Unmarshal([]byte(scanner.Text()), &existing); err != nil {
				continue
			}
			if existing.IP == rt.GetIP() && existing.Port == rt.GetPort() {
				return true
			}
		}
	}

	return false
}
func updateRequestTracerInMemory(updatedTracer ci.IRequestsTracer) error {
	if data, err := os.ReadFile(updatedTracer.GetFilePath()); err != nil {
		return fmt.Errorf("erro ao ler arquivo: %v", err)
	} else {
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if line == "" {
				continue
			}
			var existing RequestsTracer
			if err := json.Unmarshal([]byte(line), &existing); err != nil {
				continue
			}
			if existing.IP == updatedTracer.GetIP() && existing.Port == updatedTracer.GetPort() {
				lines[i] = func(data ci.IRequestsTracer) string {
					if lineBytes, lineBytesErr := json.Marshal(data); lineBytesErr != nil {
						gl.Log("error", fmt.Sprintf("Error marshalling updated tracer: %v", lineBytesErr))
						return ""
					} else {
						return string(lineBytes)
					}
				}(updatedTracer)
			}
		}
		return os.WriteFile(updatedTracer.GetFilePath(), []byte(strings.Join(lines, "\n")), 0644)
	}
}

func NewRequestTracers(g ci.IGoBE) *RequestTracers {
	return &RequestTracers{
		gobe:    g,
		tracers: make(map[string]ci.IRequestsTracer),
	}
}

func (r RequestTracers) RequestsTracerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ip := c.ClientIP()
		// port := c.Request.URL.Port()
		// endpoint := c.Request.URL.Path
		// method := c.Request.Method
		// userAgent := c.Request.UserAgent()

		// if ip == "" || port == "" || endpoint == "" || method == "" || userAgent == "" {
		// 	gl.Log("error", "Invalid request data for RequestTracerMiddleware")
		// 	c.Next()
		// 	return
		// }

		// filePath := r.gobe.GetLogFilePath()
		// tracer := NewRequestsTracerType(r.gobe, ip, port, endpoint, method, userAgent, filePath)

		// if isDuplicateRequest(r.gobe, tracer, logger.GetLogger[*RequestsTracer](nil).GetLogger()) {
		// 	gl.Log("info", fmt.Sprintf("Duplicate request detected for IP: %s, Port: %s", ip, port))
		// 	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		// 	return
		// }

		// c.Set("requestTracer", tracer)
		// c.Next()

		// if err := updateRequestTracer(r.gobe, tracer); err != nil {
		// 	gl.Log("error", fmt.Sprintf("Error updating request tracer: %v", err))
		// }
		c.Next()
	}
}

func (r *RequestTracers) GetRequestTracers() map[string]ci.IRequestsTracer {
	//r.Mutexes.MuRLock()
	//defer r.Mutexes.MuRUnlock()
	return r.tracers
}
func (r *RequestTracers) SetRequestTracers(tracers map[string]ci.IRequestsTracer) {
	/*r.Mutexes.MuAdd(1)
	defer r.Mutexes.MuDone()*/
	r.tracers = tracers
}
func (r *RequestTracers) AddRequestTracer(name string, tracer ci.IRequestsTracer) {
	//r.Mutexes.MuAdd(1)
	//defer r.Mutexes.MuDone()
	r.tracers[name] = tracer
}
func (r *RequestTracers) GetRequestTracer(name string) (ci.IRequestsTracer, bool) {
	//r.Mutexes.MuRLock()
	//defer r.Mutexes.MuRUnlock()
	tracer, ok := r.tracers[name]
	return tracer, ok
}
func (r *RequestTracers) RemoveRequestTracer(name string) {
	//r.Mutexes.MuAdd(1)
	//defer r.Mutexes.MuDone()
	delete(r.tracers, name)
}
