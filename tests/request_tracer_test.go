package tests

// import (
// 	"fmt"
// 	gb "github.com/rafa-mori/gobe"
// 	ci "github.com/rafa-mori/gobe/internal/interfaces"
// 	at "github.com/rafa-mori/gobe/internal/types"
// 	gl "github.com/rafa-mori/gobe/logger"
// 	l "github.com/faelmori/logz"
// 	"path/filepath"
// 	"strings"
// 	"sync"

// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"
// 	"time"
// )

// var (
// 	gbmT    ci.IGoBE
// 	gbmTErr error
// )

// func getGoBEInstanceTest(logFile, configFile string, isConfidential bool) (ci.IGoBE, error) {
// 	return gb.NewGoBE(
// 		"Test",
// 		"8666",
// 		"0.0.0.0",
// 		logFile,
// 		configFile,
// 		isConfidential,
// 		l.GetLogger("GoBE"),
// 		false,
// 	)
// }

// func init() {
// 	logFile := "/srv/apps/projects/gobe/tests/log/test_request_tracer.json"
// 	configFile := "/srv/apps/projects/gobe/tests/env/go_kubex.env"
// 	gbmT, gbmTErr = getGoBEInstance(logFile, configFile, true)
// }

// func unsetTestToken() {
// 	if os.Getenv("SECRET_TOKEN") == "valid_token" || os.Getenv("SECRET_TOKEN") == "invalid_token" {
// 		if err := os.Unsetenv("SECRET_TOKEN"); err != nil {
// 			gl.Log("error", fmt.Sprintf("Failed to unset environment variable: %v", err))
// 		}
// 	}
// }

// func TestRateLimit(t *testing.T) {
// 	defer unsetTestToken()

// 	var rr *httptest.ResponseRecorder
// 	req, err := http.NewRequest("GET", "/test", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.RemoteAddr = "192.168.1.1:12345"

// 	for i := 0; i < gbmT.GetRequestLimit()+2; i++ {
// 		// New instance of ResponseRecorder for each request
// 		rr = httptest.NewRecorder()
// 		gbmT.RateLimit(rr, req)

// 		resp := rr.Result()
// 		if resp.StatusCode == http.StatusOK {
// 			gl.Log("info", fmt.Sprintf("Request %d allowed", i+1))
// 		} else {
// 			if resp.StatusCode == http.StatusTooManyRequests {
// 				gl.Log("info", "Rate limit exceeded")
// 				t.Logf("Rate limit exceeded for IP: %s", req.RemoteAddr)
// 				break
// 			} else {
// 				gl.Log("info", fmt.Sprintf("Request %d blocked", i+1))
// 				t.Errorf("Request %d blocked, expected %d, got %d", i+1, http.StatusOK, resp.StatusCode)
// 			}
// 		}
// 		gl.Log("info", fmt.Sprintf("Request response code: %d", resp.StatusCode))
// 	}

// 	if rr == nil || rr.Result().StatusCode != http.StatusTooManyRequests {
// 		if rr == nil {
// 			t.Errorf("ResponseRecorder is nil")
// 		} else {
// 			t.Errorf("Expected %d, got %d", http.StatusTooManyRequests, rr.Result().StatusCode)
// 		}
// 	} else {
// 		t.Log("info", "Rate limit test passed")
// 		err = nil
// 		return
// 	}
// }

// func TestRateLimitMultipleIPs(t *testing.T) {
// 	ips := []string{"192.168.1.10", "192.168.1.20", "192.168.1.30"}
// 	ports := []string{"12345", "54321", "67890"}

// 	for _, ip := range ips {
// 		for _, port := range ports {
// 			req, err := http.NewRequest("GET", "/test", nil)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			req.RemoteAddr = fmt.Sprintf("%s:%s", ip, port)

// 			rr := httptest.NewRecorder()
// 			gbmT.RateLimit(rr, req)

// 			resp := rr.Result()
// 			if resp.StatusCode == http.StatusTooManyRequests {
// 				t.Logf("Rate limit exceeded for IP: %s Port: %s", ip, port)
// 			} else {
// 				t.Logf("Request allowed for IP: %s Port: %s", ip, port)
// 			}
// 		}
// 	}
// }

// func TestHandleValidate(t *testing.T) {
// 	// simulate a ip and port
// 	ip := "127.0.0.1"
// 	port := "8080"
// 	// create a new request
// 	req, err := http.NewRequest("GET", "/test", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	req.RemoteAddr = ip + ":" + port

// 	// set an invalid token
// 	req.Header.Set("Authorization Bearer", "invalid_token")
// 	rr := httptest.NewRecorder()
// 	gbmT.HandleValidate(rr, req)

// 	if rr.Code != http.StatusForbidden {
// 		t.Errorf("Token inválido deveria retornar %d, mas retornou %d", http.StatusForbidden, rr.Code)
// 	}
// 	if strings.TrimSpace(rr.Body.String()) != "Invalid token" {
// 		t.Errorf("Esperado 'Invalid token', obtido '%s'", rr.Body.String())
// 	}
// }

// func TestPersistRequest(t *testing.T) {
// 	wg := sync.WaitGroup{}
// 	newLogFileTest := filepath.Join(filepath.Dir(gbmT.GetLogFilePath()), "TEST"+filepath.Base(gbmT.GetLogFilePath()))

// 	defer func() {
// 		if statFile, err := os.Stat(newLogFileTest); err == nil && !statFile.IsDir() {
// 			if err := os.Remove(newLogFileTest); err != nil {
// 				t.Errorf("Erro ao remover arquivo de teste: %v", err)
// 			}
// 		}
// 	}()
// 	defer wg.Wait()

// 	_ = at.NewRequestsTracer("192.168.1.14", "8666", "/validate", "GET", "Mozilla/5.0", newLogFileTest)

// 	if _, err := os.Stat(newLogFileTest); os.IsNotExist(err) {
// 		t.Errorf("Persistência de requests falhou, arquivo '%s' não foi criado", newLogFileTest)
// 	}
// }

// func TestLoadRequestTracers(t *testing.T) {
// 	wg := sync.WaitGroup{}

// 	chTimeout := make(chan bool, 1)
// 	chDone := make(chan bool, 1)
// 	ch := make(chan any, 1)

// 	gl.Log("info", "Starting TestLoadRequestTracers")
// 	requestsTracer := make(map[string]ci.IRequestsTracer)

// 	wg.Add(1)
// 	go func(rqt map[string]ci.IRequestsTracer, gbm ci.IGoBE, wg *sync.WaitGroup, ch chan any) {
// 		defer wg.Done()
// 		gl.Log("info", "Loading request tracers from file")
// 		rqtMap, rqtErr := at.LoadRequestsTracerFromFile(gbm)
// 		if rqtErr != nil {
// 			t.Errorf("Erro ao carregar request tracers: %v", rqtErr)
// 		}
// 		if len(rqtMap) == 0 {
// 			absPath, absPathErr := filepath.Abs(gbmT.GetLogFilePath())
// 			gl.Log("error", fmt.Sprintf("Erro ao carregar request tracers do arquivo, mapa vazio: %s", absPath))
// 			if absPathErr != nil {
// 				t.Errorf("Erro ao obter o caminho absoluto do arquivo: %v", absPathErr)
// 			}
// 			if len(at.RequestTracers) == 0 {
// 				gl.Log("error", fmt.Sprintf("Erro ao carregar request tracers do arquivo, mapa vazio: %s", absPath))
// 				t.Errorf("Erro ao carregar request tracers do arquivo, mapa vazio")
// 			} else {
// 				gl.Log("info", "Request tracers loaded successfully")
// 				gl.Log("info", fmt.Sprintf("Loaded %d request tracers", len(at.RequestTracers)))
// 				t.Logf("Loaded %d request tracers", len(at.RequestTracers))
// 				for ip, tracer := range at.RequestTracers {
// 					gl.Log("info", fmt.Sprintf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent()))
// 					t.Logf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent())
// 				}
// 			}
// 		} else {
// 			gl.Log("info", "Request tracers loaded successfully")
// 			gl.Log("info", fmt.Sprintf("Loaded %d request tracers", len(rqtMap)))
// 			t.Logf("Loaded %d request tracers", len(rqtMap))
// 			for ip, tracer := range rqtMap {
// 				gl.Log("info", fmt.Sprintf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent()))
// 				t.Logf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent())
// 			}
// 		}
// 		rqt = rqtMap
// 		ch <- rqt
// 	}(requestsTracer, gbmT, &wg, ch)

// 	go func(chTimeout chan bool) {
// 		time.Sleep(10 * time.Second)
// 		if chTimeout != nil {
// 			chTimeout <- true
// 		}
// 	}(chTimeout)

// 	t.Log("info", "Waiting for request tracers to load...")
// 	gl.Log("info", "Waiting for request tracers to load...")
// 	wg.Wait()

// 	for {
// 		select {
// 		case <-chDone:
// 			t.Log("info", "Request tracers loaded successfully")
// 			gl.Log("info", "Request tracers loaded successfully")
// 			t.Log("info", fmt.Sprintf("Loaded %d request tracers", len(requestsTracer)))
// 			gl.Log("info", fmt.Sprintf("Loaded %d request tracers", len(requestsTracer)))
// 			for ip, tracer := range requestsTracer {
// 				gl.Log("info", fmt.Sprintf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent()))
// 				t.Logf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent())
// 			}
// 			return
// 		case rqt := <-ch:
// 			t.Log("info", "Request tracers loaded successfully")
// 			gl.Log("info", "Request tracers loaded successfully")
// 			t.Log("info", fmt.Sprintf("Loaded %d request tracers", len(rqt.(map[string]ci.IRequestsTracer))))
// 			gl.Log("info", fmt.Sprintf("Loaded %d request tracers", len(rqt.(map[string]ci.IRequestsTracer))))
// 			for ip, tracer := range rqt.(map[string]ci.IRequestsTracer) {
// 				gl.Log("info", fmt.Sprintf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent()))
// 				t.Logf("IP: %s, Count: %d, LastUserAgent: %s", ip, tracer.GetCount(), tracer.GetLastUserAgent())
// 			}
// 			requestsTracer = rqt.(map[string]ci.IRequestsTracer)
// 			chDone <- true
// 		case <-chTimeout:
// 			gl.Log("info", "Timeout reached while waiting for request tracers to load")
// 			t.Errorf("Timeout reached while waiting for request tracers to load")
// 			chDone <- true
// 		default:
// 			continue
// 		}
// 	}
// }

// func TestHandleValidateWithValidTokenReturnsStatusOK(t *testing.T) {
// 	defer unsetTestToken()

// 	g := gbmT
// 	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
// 	req.Header.Set("Authorization", "Bearer valid_token")
// 	w := httptest.NewRecorder()

// 	g.HandleValidate(w, req)

// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
// 	}
// }

// func TestHandleValidateWithInvalidTokenReturnsForbidden(t *testing.T) {
// 	defer unsetTestToken()

// 	g := gbmT
// 	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
// 	req.Header.Set("Authorization", "Bearer invalid_token")
// 	w := httptest.NewRecorder()

// 	g.HandleValidate(w, req)

// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusForbidden {
// 		t.Errorf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
// 	}
// }

// func TestHandleContactWithValidDataReturnsStatusOK(t *testing.T) {
// 	defer unsetTestToken()

// 	g := gbmT
// 	form := at.ContactForm{
// 		Name:    "Test User A",
// 		Email:   "faelmori@gmail.com",
// 		Message: "Hello World",
// 		Token:   "valid_token",
// 	}
// 	body, _ := json.Marshal(form)
// 	req := httptest.NewRequest(http.MethodPost, "/contact", bytes.NewReader(body))
// 	w := httptest.NewRecorder()

// 	g.HandleContact(w, req)

// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusOK {
// 		gl.Log("error", fmt.Sprintf("HandleContact failed with status: %d", resp.StatusCode))
// 		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
// 	}
// }

// func TestHandleContactWithInvalidTokenReturnsForbidden(t *testing.T) {
// 	defer unsetTestToken()

// 	g := gbmT
// 	form := at.ContactForm{
// 		Name:    "Test User",
// 		Email:   "test@example.com",
// 		Message: "Hello",
// 		Token:   "invalid_token",
// 	}
// 	body, _ := json.Marshal(form)
// 	req := httptest.NewRequest(http.MethodPost, "/contact", bytes.NewReader(body))
// 	w := httptest.NewRecorder()

// 	g.HandleContact(w, req)

// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusForbidden {
// 		t.Errorf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
// 	}
// }

// func TestRateLimitExceededReturnsTooManyRequests(t *testing.T) {
// 	defer unsetTestToken()

// 	g := gbmT
// 	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
// 	req.RemoteAddr = "127.0.0.1:12345"
// 	w := httptest.NewRecorder()

// 	for i := 0; i < gbmT.GetRequestLimit()+1; i++ {
// 		g.RateLimit(w, req)
// 	}

// 	resp := w.Result()
// 	if resp.StatusCode != http.StatusTooManyRequests {
// 		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, resp.StatusCode)
// 	}
// }

// func TestInitializeWithInvalidEnvironmentFileLogsError(t *testing.T) {
// 	defer unsetTestToken()

// 	bgmT, bgmTErr := gb.NewGoBE("test", "8080", "127.0.0.1", "", "invalid_config.json", false, nil, false)
// 	if bgmTErr == nil {
// 		t.Errorf("expected error, got nil")
// 		return
// 	}
// 	done := make(chan bool)
// 	defer close(done)

// 	go func(bgmT ci.IGoBE, done chan bool) {
// 		if bgmT != nil {
// 			err := bgmT.Initialize()
// 			if err != nil {
// 				gl.Log("error", fmt.Sprintf("Initialize failed: %v", err))
// 				t.Errorf("Initialize failed: %v", err)
// 				return
// 			}
// 			done <- true
// 		} else {
// 			gl.Log("debug", "GoBE instance is nil")
// 			done <- false
// 		}
// 	}(bgmT, done)

// 	select {
// 	case <-done:
// 	case <-time.After(5 * time.Second):
// 		t.Fatal("Initialize did not complete within the expected time")
// 	}
// }
