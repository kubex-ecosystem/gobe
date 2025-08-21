package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_tasks_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/tasks"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
)

type MCPTasksRoutes struct {
	ar.IRouter
}

func NewMCPTasksRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP Tasks routes")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCPTasksRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpTasksController := mcp_tasks_controller.NewTasksController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	// middlewaresMap := rtl.GetMiddlewares()
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This need to be changed to true for production
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllTasks"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks", "application/json", mcpTasksController.GetAllTasks, middlewaresMap, dbService, secureProperties)
	routesMap["GetTaskByID"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks/:id", "application/json", mcpTasksController.GetTaskByID, middlewaresMap, dbService, secureProperties)
	routesMap["DeleteTask"] = NewRoute(http.MethodDelete, "/api/v1/mcp/tasks/:id", "application/json", mcpTasksController.DeleteTask, middlewaresMap, dbService, secureProperties)
	routesMap["GetTasksByProvider"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks/provider/:provider", "application/json", mcpTasksController.GetTasksByProvider, middlewaresMap, dbService, secureProperties)
	routesMap["GetTasksByTarget"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks/target/:target", "application/json", mcpTasksController.GetTasksByTarget, middlewaresMap, dbService, secureProperties)
	routesMap["GetActiveTasks"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks/active", "application/json", mcpTasksController.GetActiveTasks, middlewaresMap, dbService, secureProperties)
	routesMap["GetTasksDueForExecution"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks/due", "application/json", mcpTasksController.GetTasksDueForExecution, middlewaresMap, dbService, secureProperties)
	routesMap["MarkTaskAsRunning"] = NewRoute(http.MethodPost, "/api/v1/mcp/tasks/:id/running", "application/json", mcpTasksController.MarkTaskAsRunning, middlewaresMap, dbService, secureProperties)
	routesMap["MarkTaskAsCompleted"] = NewRoute(http.MethodPost, "/api/v1/mcp/tasks/:id/completed", "application/json", mcpTasksController.MarkTaskAsCompleted, middlewaresMap, dbService, secureProperties)
	routesMap["MarkTaskAsFailed"] = NewRoute(http.MethodPost, "/api/v1/mcp/tasks/:id/failed", "application/json", mcpTasksController.MarkTaskAsFailed, middlewaresMap, dbService, secureProperties)
	routesMap["GetTaskCronJob"] = NewRoute(http.MethodGet, "/api/v1/mcp/tasks/:id/cron", "application/json", mcpTasksController.GetTaskCronJob, middlewaresMap, dbService, secureProperties)

	return routesMap
}
