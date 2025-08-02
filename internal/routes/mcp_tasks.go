package routes

import (
	"net/http"

	mcp_tasks_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/tasks"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

type MCPTasksRoutes struct {
	ar.IRouter
}

func NewMCPTasksRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for MCPTasksRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpTasksController := mcp_tasks_controller.NewTasksController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["GetAllTasks"] = NewRoute(http.MethodGet, "/mcp/tasks", "application/json", mcpTasksController.GetAllTasks, middlewaresMap, dbService)
	routesMap["GetTaskByID"] = NewRoute(http.MethodGet, "/mcp/tasks/:id", "application/json", mcpTasksController.GetTaskByID, middlewaresMap, dbService)
	routesMap["DeleteTask"] = NewRoute(http.MethodDelete, "/mcp/tasks/:id", "application/json", mcpTasksController.DeleteTask, middlewaresMap, dbService)
	routesMap["GetTasksByProvider"] = NewRoute(http.MethodGet, "/mcp/tasks/provider/:provider", "application/json", mcpTasksController.GetTasksByProvider, middlewaresMap, dbService)
	routesMap["GetTasksByTarget"] = NewRoute(http.MethodGet, "/mcp/tasks/target/:target", "application/json", mcpTasksController.GetTasksByTarget, middlewaresMap, dbService)
	routesMap["GetActiveTasks"] = NewRoute(http.MethodGet, "/mcp/tasks/active", "application/json", mcpTasksController.GetActiveTasks, middlewaresMap, dbService)
	routesMap["GetTasksDueForExecution"] = NewRoute(http.MethodGet, "/mcp/tasks/due", "application/json", mcpTasksController.GetTasksDueForExecution, middlewaresMap, dbService)
	routesMap["MarkTaskAsRunning"] = NewRoute(http.MethodPost, "/mcp/tasks/:id/running", "application/json", mcpTasksController.MarkTaskAsRunning, middlewaresMap, dbService)
	routesMap["MarkTaskAsCompleted"] = NewRoute(http.MethodPost, "/mcp/tasks/:id/completed", "application/json", mcpTasksController.MarkTaskAsCompleted, middlewaresMap, dbService)
	routesMap["MarkTaskAsFailed"] = NewRoute(http.MethodPost, "/mcp/tasks/:id/failed", "application/json", mcpTasksController.MarkTaskAsFailed, middlewaresMap, dbService)
	routesMap["GetTaskCronJob"] = NewRoute(http.MethodGet, "/mcp/tasks/:id/cron", "application/json", mcpTasksController.GetTaskCronJob, middlewaresMap, dbService)

	return routesMap
}
