// Package tasks provides the controller for managing user tasks.
package tasks

import (
	"context"
	"net/http"

	mdl "github.com/kubex-ecosystem/gdbase/factory/models/mcp"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

	"github.com/gin-gonic/gin"
)

type TasksController struct {
	tasksService mdl.TasksService
}

func NewTasksController(bridge *svc.Bridge) *TasksController {
	return &TasksController{
		tasksService: bridge.TasksService(context.Background()),
	}
}

// GetAllTasks retrieves all tasks.
// @Summary MCP Tasks Controller
// @Description Controller for managing tasks in the MCP
// @Schemes http https
// @Tags tasks
// @Summary Get All Tasks
// @Description Retrieves a list of all tasks.
// @Accept json
// @Produce json
// @Success 200 {object} []models.TasksModel
// @Failure 500 {string} Failed to get tasks
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Tasks not found
// @Router /mcp/tasks [get]
func (tc *TasksController) GetAllTasks(c *gin.Context) {
	// ListTasks(active, outOfDate, running bool) ([]ITasksModel, error)
	gl.Log("info", "Fetching all tasks")
	tasks, err := tc.tasksService.ListTasks(nil)
	if err != nil {
		gl.Log("error", "Failed to get tasks", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID retrieves a task by its ID.
// @Summary Get Task by ID
// @Description Retrieves a task by its ID.
// @Accept json
// @Produce json
// @Success 200 {object} models.TasksModel
// @Failure 404 {string} Task not found
// @Failure 500 {string} Failed to get task
// @Failure 400 {string} Invalid request
// @Router /mcp/tasks/{id} [get]
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

// DeleteTask deletes a task by its ID.
// @Summary Delete Task
// @Description Deletes a task by its ID.
// @Accept json
// @Produce json
// @Success 204 {string} Task deleted successfully
// @Failure 404 {string} Task not found
// @Failure 500 {string} Failed to delete task
// @Failure 400 {string} Invalid request
// @Router /mcp/tasks/{id} [delete]
func (tc *TasksController) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := tc.tasksService.DeleteTask(id); err != nil {
		gl.Log("error", "Failed to delete task", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

// GetTasksByProvider retrieves tasks by provider.
// @Summary Get Tasks by Provider
// @Description Retrieves tasks by provider.
// @Accept json
// @Produce json
// @Success 200 {object} []models.TasksModel
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Tasks not found
// @Failure 500 {string} Failed to get tasks
// @Router /mcp/tasks/provider/{provider} [get]
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

// GetTasksByTarget retrieves tasks by target.
// @Summary Get Tasks by Target
// @Description Retrieves tasks by target.
// @Accept json
// @Produce json
// @Success 200 {object} []models.TasksModel
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Tasks not found
// @Failure 500 {string} Failed to get tasks
// @Router /mcp/tasks/target/{target} [get]
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

// GetActiveTasks retrieves all active tasks.
// @Summary Get Active Tasks
// @Description Retrieves all active tasks.
// @Accept json
// @Produce json
// @Success 200 {object} []models.TasksModel
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Tasks not found
// @Failure 500 {string} Failed to get tasks
// @Router /mcp/tasks/active [get]
func (tc *TasksController) GetActiveTasks(c *gin.Context) {
	tasks, err := tc.tasksService.GetActiveTasks()
	if err != nil {
		gl.Log("error", "Failed to get active tasks", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTasksDueForExecution retrieves tasks due for execution.
// @Summary Get Tasks Due for Execution
// @Description Retrieves tasks due for execution.
// @Accept json
// @Produce json
// @Success 200 {object} []models.TasksModel
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Tasks not found
// @Failure 500 {string} Failed to get tasks
// @Router /mcp/tasks/due [get]
func (tc *TasksController) GetTasksDueForExecution(c *gin.Context) {
	tasks, err := tc.tasksService.GetTasksDueForExecution()
	if err != nil {
		gl.Log("error", "Failed to get tasks due for execution", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks due for execution"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// MarkTaskAsRunning marks a task as running.
// @Summary Mark Task as Running
// @Description Marks a task as running.
// @Accept json
// @Produce json
// @Success 200 {string} Task marked as running
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Task not found
// @Failure 500 {string} Failed to mark task as running
// @Router /mcp/tasks/{id}/running [post]
func (tc *TasksController) MarkTaskAsRunning(c *gin.Context) {
	id := c.Param("id")

	if err := tc.tasksService.MarkTaskAsRunning(id); err != nil {
		gl.Log("error", "Failed to mark task as running", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark task as running"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task marked as running"})
}

// MarkTaskAsCompleted marks a task as completed.
// @Summary Mark Task as Completed
// @Description Marks a task as completed.
// @Accept json
// @Produce json
// @Success 200 {string} Task marked as completed
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Task not found
// @Failure 500 {string} Failed to mark task as completed
// @Router /mcp/tasks/{id}/completed [post]
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

// MarkTaskAsFailed marks a task as failed.
// @Summary Mark Task as Failed
// @Description Marks a task as failed.
// @Accept json
// @Produce json
// @Success 200 {string} Task marked as failed
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Task not found
// @Failure 500 {string} Failed to mark task as failed
// @Router /mcp/tasks/{id}/failed [post]
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

// GetTaskCronJob retrieves the CronJob representation of a task.
// @Summary Get Task CronJob
// @Description Retrieves the CronJob representation of a task.
// @Accept json
// @Produce json
// @Success 200 {object} models.CronJobIntegration
// @Failure 400 {string} Invalid request
// @Failure 404 {string} Task not found
// @Failure 500 {string} Failed to get task CronJob
// @Router /mcp/tasks/{id}/cronjob [get]
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
