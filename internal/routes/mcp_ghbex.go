package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_ghbex_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/ghbexz"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
)

type MCPGHbexRoutes struct {
	ar.IRouter
}

func NewMCPGHbexRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP GHbex routes")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCP GHbex routes")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpGHbexController := mcp_ghbex_controller.NewGHbexController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	// middlewaresMap := rtl.GetMiddlewares()
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This need to be changed to true for production
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GHbex"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex", "application/json", mcpGHbexController.GetGHbex, middlewaresMap, dbService, secureProperties)
	routesMap["Health"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/health", "application/json", mcpGHbexController.GetHealth, middlewaresMap, dbService, secureProperties)
	routesMap["Repos"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/repos/:owner/:repo", "application/json", mcpGHbexController.GetRepos, middlewaresMap, dbService, secureProperties)
	routesMap["AdminSanitize"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/admin/sanitize/:owner/:repo", "application/json", mcpGHbexController.AdminSanitize, middlewaresMap, dbService, secureProperties)
	routesMap["AdminRepos"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/admin/repos/:owner/:repo", "application/json", mcpGHbexController.AdminRepos, middlewaresMap, dbService, secureProperties)
	routesMap["Analytics"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/analytics/:owner/:repo", "application/json", mcpGHbexController.Analytics, middlewaresMap, dbService, secureProperties)
	routesMap["Productivity"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/productivity/:owner/:repo", "application/json", mcpGHbexController.Productivity, middlewaresMap, dbService, secureProperties)
	routesMap["Intelligence"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/intelligence/:owner/:repo", "application/json", mcpGHbexController.Intelligence, middlewaresMap, dbService, secureProperties)
	routesMap["Automation"] = NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/automation/:owner/:repo", "application/json", mcpGHbexController.Automation, middlewaresMap, dbService, secureProperties)

	return routesMap
}
