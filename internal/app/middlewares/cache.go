package middlewares

import (
	"github.com/gin-gonic/gin"
)

func CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Cache middleware logic
		c.Next()
	}
}
