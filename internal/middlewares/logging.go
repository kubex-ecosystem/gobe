package middlewares

import (
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"

	"github.com/gin-gonic/gin"

	"fmt"
)

func Logger(logger l.Logger) gin.HandlerFunc {
	var lgr struct {
		Logger l.Logger
	}
	if logger == nil {
		lgr = struct {
			Logger l.Logger
		}{
			Logger: l.GetLogger("GoBE"),
		}
	} else {
		lgr = struct {
			Logger l.Logger
		}{
			Logger: logger,
		}
	}
	return func(c *gin.Context) {
		gl.Log("info", "Request", c.Request.Proto, c.Request.Method, c.Request.URL.Path)
		gl.LogObjLogger(&lgr, "info", fmt.Sprintf("Request: %s %s %s", c.Request.Proto, c.Request.Method, c.Request.URL.Path))
		c.Next()
	}
}
