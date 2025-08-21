package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a channel to signal when the request is done
		done := make(chan struct{})

		// Start a goroutine to process the request
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for either the request to complete or the timeout
		select {
		case <-done:
			// Request completed within the timeout
		case <-time.After(duration):
			// Timeout occurred
			c.AbortWithStatusJSON(408, gin.H{"error": "Request timed out"})
		}
	}
}
