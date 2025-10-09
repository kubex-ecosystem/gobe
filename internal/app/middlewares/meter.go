package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// skipMeteringPaths are paths that should not log telemetry to reduce noise
var skipMeteringPaths = []string{
	"/health",
	"/healthz",
	"/api/v1/health",
	"/status",
}

// shouldSkipMetering determines if a request path should skip telemetry logging
func shouldSkipMetering(path string) bool {
	for _, skipPath := range skipMeteringPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

func MeterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// traceID
		traceID := uuid.New().String()
		c.Set("traceID", traceID)

		// Start telemetry
		c.Set("startTime", time.Now())

		// Process request
		c.Next()

		// End telemetry
		c.Set("endTime", time.Now())

		// Skip telemetry logging for health check endpoints
		if shouldSkipMetering(c.Request.URL.Path) {
			return
		}

		// Log telemetry data
		startTime, exists := c.Get("startTime")
		if exists {
			endTime, _ := c.Get("endTime")
			// Log the telemetry data (startTime and endTime)
			duration := endTime.(time.Time).Sub(startTime.(time.Time))

			// Log the duration
			gl.Log("debug", fmt.Sprintf("%s: Request processed in %s", traceID, duration))
		}
	}
}
