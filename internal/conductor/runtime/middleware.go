package runtime

import "github.com/gin-gonic/gin"

var dispatcher *Dispatcher

func RouteIntentMiddleware(c *gin.Context) {
	intent, err := NewIntentFromRequest(c.Request)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := dispatcher.Dispatch(intent); err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Next()
}

func BindDispatcher(d *Dispatcher) { dispatcher = d }
