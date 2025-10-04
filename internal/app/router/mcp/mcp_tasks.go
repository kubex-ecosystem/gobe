package mcp

import (
	"context"
	"net/http"

	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	"github.com/gin-gonic/gin"
	mcp_tasks_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/tasks"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
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
		gl.Log("error", "Database service is nil for OAuthRoutes")
		return nil
	}
	ctx := context.Background()
	dbCfg := dbService.GetConfig(ctx)
	if dbCfg == nil {
		gl.Log("error", "Database config is nil for OAuthRoutes")
		return nil
	}
	dbName := dbCfg.GetDBName()
	ctx = context.WithValue(ctx, gl.ContextDBNameKey, dbName)
	bridge := gdbasez.NewBridge(ctx, dbService, dbName)
	mcpTasksController := mcp_tasks_controller.NewTasksController(bridge)

	routesMap := make(map[string]ar.IRoute)
	// middlewaresMap := rtl.GetMiddlewares()
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This need to be changed to true for production
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllTasks"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks", "application/json", mcpTasksController.GetAllTasks, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetTaskByID"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks/:id", "application/json", mcpTasksController.GetTaskByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteTask"] = proto.NewRoute(http.MethodDelete, "/api/v1/mcp/tasks/:id", "application/json", mcpTasksController.DeleteTask, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetTasksByProvider"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks/provider/:provider", "application/json", mcpTasksController.GetTasksByProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetTasksByTarget"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks/target/:target", "application/json", mcpTasksController.GetTasksByTarget, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetActiveTasks"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks/active", "application/json", mcpTasksController.GetActiveTasks, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetTasksDueForExecution"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks/due", "application/json", mcpTasksController.GetTasksDueForExecution, middlewaresMap, dbService, secureProperties, nil)
	routesMap["MarkTaskAsRunning"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/tasks/:id/running", "application/json", mcpTasksController.MarkTaskAsRunning, middlewaresMap, dbService, secureProperties, nil)
	routesMap["MarkTaskAsCompleted"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/tasks/:id/completed", "application/json", mcpTasksController.MarkTaskAsCompleted, middlewaresMap, dbService, secureProperties, nil)
	routesMap["MarkTaskAsFailed"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/tasks/:id/failed", "application/json", mcpTasksController.MarkTaskAsFailed, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetTaskCronJob"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/tasks/:id/cron", "application/json", mcpTasksController.GetTaskCronJob, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
