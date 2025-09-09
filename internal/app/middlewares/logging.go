package middlewares

import (
	l "github.com/kubex-ecosystem/logz"

	"github.com/gin-gonic/gin"

	"fmt"
)

func Logger(logger l.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		gl.Log("info", "Request", c.Request.Proto, c.Request.Method, c.Request.URL.Path)
		gl.Log("info", fmt.Sprintf("Request: %s %s %s", c.Request.Proto, c.Request.Method, c.Request.URL.Path))
		c.Next()
	}
}
