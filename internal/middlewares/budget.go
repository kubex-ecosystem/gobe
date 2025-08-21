package middlewares

import (
	"github.com/gin-gonic/gin"
)

func BudgetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Budget middleware logic
		c.Next()
	}
}
