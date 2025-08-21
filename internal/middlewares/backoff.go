package middlewares

import (
	"github.com/gin-gonic/gin"
)

func BackoffMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Backoff middleware logic
		c.Next()
	}
}
