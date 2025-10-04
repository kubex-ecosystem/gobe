package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

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
