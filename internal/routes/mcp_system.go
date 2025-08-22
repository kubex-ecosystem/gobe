package routes

import (
	"net/http"

	mcp_system_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/system"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
)

type MCPSystemRoutes struct {
	ar.IRouter
}

func NewMCPSystemRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP System routes")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCPSystemRoutes")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpSystemController := mcp_system_controller.NewMetricsController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	// middlewaresMap := rtl.GetMiddlewares()

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This is temporary, should be set to true later
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetGeneralSystemMetrics"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/metrics", "application/json", mcpSystemController.GetGeneralSystemMetrics /* middlewaresMap */, nil, dbService, secureProperties)
	routesMap["RegisterRoutes"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/routes", "application/json", mcpSystemController.RegisterResources, nil, dbService, secureProperties)
	routesMap["RegisterTools"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/tools", "application/json", mcpSystemController.RegisterTools, nil, dbService, secureProperties)
	routesMap["HandleAnalyzeMessage"] = NewRoute(http.MethodPost, "/api/v1/mcp/system/analyze", "application/json", mcpSystemController.HandleAnalyzeMessage, nil, dbService, secureProperties)
	routesMap["HandleSendMessage"] = NewRoute(http.MethodPost, "/api/v1/mcp/system/send-message", "application/json", mcpSystemController.SendMessage, nil, dbService, secureProperties)
	routesMap["HandleCreateTask"] = NewRoute(http.MethodPost, "/api/v1/mcp/system/create-task", "application/json", mcpSystemController.HandleCreateTask, nil, dbService, secureProperties)
	routesMap["HandleSystemInfo"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/info", "application/json", mcpSystemController.GetCPUInfo, nil, dbService, secureProperties)
	routesMap["HandleShellCommand"] = NewRoute(http.MethodPost, "/api/v1/mcp/system/shell-command", "application/json", mcpSystemController.ShellCommand, nil, dbService, secureProperties)
	routesMap["GetCPUInfo"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/cpu-info", "application/json", mcpSystemController.GetCPUInfo, nil, dbService, secureProperties)
	routesMap["GetMemoryInfo"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/memory-info", "application/json", mcpSystemController.GetMemoryInfo, nil, dbService, secureProperties)
	routesMap["GetDiskInfo"] = NewRoute(http.MethodGet, "/api/v1/mcp/system/disk-info", "application/json", mcpSystemController.GetDiskInfo, nil, dbService, secureProperties)

	return routesMap
}
