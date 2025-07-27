// Package tasks provides the controller for managing user tasks.
package tasks

import (
	"net/http"

	models "github.com/rafa-mori/gdbase/factory/models/mcp"
	svc "github.com/rafa-mori/gobe/internal/services"
	gl "github.com/rafa-mori/gobe/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TasksController struct {
	tasksService svc.TasksService
}

func NewTasksController(db *gorm.DB) *TasksController {
	return &TasksController{
		tasksService: svc.NewTasksService(models.NewTasksRepo(db)),
	}
}

func (tc *TasksController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/mcp/tasks")
	{
		api.GET("/", tc.GetAllTasks)
		api.GET("/:id", tc.GetTaskByID)
		api.DELETE("/:id", tc.DeleteTask)
		api.GET("/provider/:provider", tc.GetTasksByProvider)
		api.GET("/target/:target", tc.GetTasksByTarget)
		api.GET("/active", tc.GetActiveTasks)
		api.GET("/due", tc.GetTasksDueForExecution)
		api.POST("/:id/running", tc.MarkTaskAsRunning)
		api.POST("/:id/completed", tc.MarkTaskAsCompleted)
		api.POST("/:id/failed", tc.MarkTaskAsFailed)
		api.GET("/:id/cron", tc.GetTaskCronJob)
	}
}

// GetAllTasks retrieves all tasks
func (tc *TasksController) GetAllTasks(c *gin.Context) {
	tasks, err := tc.tasksService.ListTasks()
	if err != nil {
		gl.Log("error", "Failed to get tasks", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID retrieves a task by ID
func (tc *TasksController) GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	task, err := tc.tasksService.GetTaskByID(id)
	if err != nil {
		gl.Log("error", "Failed to get task by ID", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a task by ID
func (tc *TasksController) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := tc.tasksService.DeleteTask(id); err != nil {
		gl.Log("error", "Failed to delete task", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// GetTasksByProvider retrieves tasks by provider
func (tc *TasksController) GetTasksByProvider(c *gin.Context) {
	provider := c.Param("provider")
	tasks, err := tc.tasksService.GetTasksByProvider(provider)
	if err != nil {
		gl.Log("error", "Failed to get tasks by provider", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks by provider"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTasksByTarget retrieves tasks by target
func (tc *TasksController) GetTasksByTarget(c *gin.Context) {
	target := c.Param("target")
	tasks, err := tc.tasksService.GetTasksByTarget(target)
	if err != nil {
		gl.Log("error", "Failed to get tasks by target", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks by target"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetActiveTasks retrieves all active tasks
func (tc *TasksController) GetActiveTasks(c *gin.Context) {
	tasks, err := tc.tasksService.GetActiveTasks()
	if err != nil {
		gl.Log("error", "Failed to get active tasks", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTasksDueForExecution retrieves tasks due for execution
func (tc *TasksController) GetTasksDueForExecution(c *gin.Context) {
	tasks, err := tc.tasksService.GetTasksDueForExecution()
	if err != nil {
		gl.Log("error", "Failed to get tasks due for execution", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks due for execution"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// MarkTaskAsRunning marks a task as running
func (tc *TasksController) MarkTaskAsRunning(c *gin.Context) {
	id := c.Param("id")

	if err := tc.tasksService.MarkTaskAsRunning(id); err != nil {
		gl.Log("error", "Failed to mark task as running", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as running"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task marked as running"})
}

// MarkTaskAsCompleted marks a task as completed
func (tc *TasksController) MarkTaskAsCompleted(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Message string `json:"message"`
	}
	c.ShouldBindJSON(&req)

	if err := tc.tasksService.MarkTaskAsCompleted(id, req.Message); err != nil {
		gl.Log("error", "Failed to mark task as completed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as completed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task marked as completed"})
}

// MarkTaskAsFailed marks a task as failed
func (tc *TasksController) MarkTaskAsFailed(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Message string `json:"message"`
	}
	c.ShouldBindJSON(&req)

	if err := tc.tasksService.MarkTaskAsFailed(id, req.Message); err != nil {
		gl.Log("error", "Failed to mark task as failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task marked as failed"})
}

// GetTaskCronJob retrieves the CronJob representation of a task
func (tc *TasksController) GetTaskCronJob(c *gin.Context) {
	id := c.Param("id")

	cronJob, err := tc.tasksService.ConvertTaskToCronJob(id)
	if err != nil {
		gl.Log("error", "Failed to convert task to CronJob", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert task to CronJob"})
		return
	}

	c.JSON(http.StatusOK, cronJob)
}
