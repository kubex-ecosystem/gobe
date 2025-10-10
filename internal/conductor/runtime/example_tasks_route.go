package runtime

import "github.com/gin-gonic/gin"

// RegisterTaskRoutes sets up the HTTP routes for task management. (EXAMPLE)
func RegisterTaskRoutes(r *gin.Engine) {
	r.POST("/api/tasks", func(c *gin.Context) {
		c.JSON(202, gin.H{
			"status":      "accepted",
			"origin":      "conductor",
			"observation": "o scheduler pode ter enfileirado ou agendado conforme DCL",
		})
	})
}
