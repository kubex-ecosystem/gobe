package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	srv "github.com/rafa-mori/gobe/internal/services"
)

type RateLimitMiddleware struct {
	dbConfig      *srv.IDBConfig
	LogFile       string
	requestLimit  int
	requestWindow time.Duration
	g             ci.IGoBE
}

func NewRateLimitMiddleware(g ci.IGoBE, dbConfig srv.IDBConfig, logDir string, limit int, window time.Duration) (*RateLimitMiddleware, error) {
	return &RateLimitMiddleware{
		dbConfig:      &dbConfig,
		LogFile:       logDir,
		requestLimit:  limit,
		requestWindow: window,
		g:             g,
	}, nil
}

func (rl *RateLimitMiddleware) RateLimit(c *gin.Context) bool {
	// ip, port, splitHostPortErr := net.SplitHostPort(r.RemoteAddr)
	// if splitHostPortErr != nil {
	// 	http.Error(w, "Internal server error", http.StatusInternalServerError)
	// 	gl.Log("warn", fmt.Sprintf("Error splitting host and port: %v", splitHostPortErr.Error()))
	// 	return false
	// }

	// requestTracer := t.NewRequestsTracer(rl.g, ip, port, r.URL.Path, r.Method, r.UserAgent(), rl.LogFile)
	// requestTracer.GetMutexes().MuRLock()
	// defer requestTracer.GetMutexes().MuRUnlock()

	// if !requestTracer.IsValid() {
	// 	http.Error(w, "Request limit exceeded", http.StatusTooManyRequests)
	// 	gl.Log("warn", fmt.Sprintf("Invalid request tracer: %v", requestTracer.GetError()))
	// 	return false
	// }

	c.Next()

	return true
}
func (rl *RateLimitMiddleware) GetRequestLimit() int {
	return rl.requestLimit
}
func (rl *RateLimitMiddleware) SetRequestLimit(limit int) {
	rl.requestLimit = limit
}
func (rl *RateLimitMiddleware) GetRequestWindow() time.Duration {
	return rl.requestWindow
}
func (rl *RateLimitMiddleware) SetRequestWindow(window time.Duration) {
	rl.requestWindow = window
}
