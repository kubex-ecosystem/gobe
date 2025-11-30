package middlewares

import (
	"github.com/gin-gonic/gin"

	"fmt"
	"strings"

	"github.com/kubex-ecosystem/logz"
)

// healthCheckPaths are paths that should not be logged to reduce noise
var healthCheckPaths = []string{
	"/health",
	"/healthz",
	"/api/v1/health",
	"/status",
}

// shouldSkipLogging determines if a request path should skip logging
func shouldSkipLogging(path string) bool {
	for _, healthPath := range healthCheckPaths {
		if strings.HasPrefix(path, healthPath) {
			return true
		}
	}
	return false
}

func Logger(logger *logz.LoggerZ) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health check endpoints to reduce noise
		if !shouldSkipLogging(c.Request.URL.Path) {
			// Log only once with formatted message
			logz.Log("info", fmt.Sprintf("Request: %s %s %s", c.Request.Proto, c.Request.Method, c.Request.URL.Path))
		}
		c.Next()
	}
}
