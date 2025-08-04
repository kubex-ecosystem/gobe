package routes

import (
	"net/http"

	mcp_system_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/system"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
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

	return routesMap
}
