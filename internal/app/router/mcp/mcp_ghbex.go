package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_ghbex_controller "github.com/rafa-mori/gobe/internal/app/controllers/mcp/ghbexz"
	proto "github.com/rafa-mori/gobe/internal/app/router/types"
	ar "github.com/rafa-mori/gobe/internal/contracts/interfaces"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
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

	routesMap["GHbex"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex", "application/json", mcpGHbexController.GetGHbex, middlewaresMap, dbService, secureProperties, nil)
	routesMap["Health"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/health", "application/json", mcpGHbexController.GetHealth, middlewaresMap, dbService, secureProperties, nil)
	routesMap["Repos"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/repos/:owner/:repo", "application/json", mcpGHbexController.GetRepos, middlewaresMap, dbService, secureProperties, nil)
	routesMap["AdminSanitize"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/admin/sanitize/:owner/:repo", "application/json", mcpGHbexController.AdminSanitize, middlewaresMap, dbService, secureProperties, nil)
	routesMap["AdminRepos"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/admin/repos/:owner/:repo", "application/json", mcpGHbexController.AdminRepos, middlewaresMap, dbService, secureProperties, nil)
	routesMap["Analytics"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/analytics/:owner/:repo", "application/json", mcpGHbexController.Analytics, middlewaresMap, dbService, secureProperties, nil)
	routesMap["Productivity"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/productivity/:owner/:repo", "application/json", mcpGHbexController.Productivity, middlewaresMap, dbService, secureProperties, nil)
	routesMap["Intelligence"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/intelligence/:owner/:repo", "application/json", mcpGHbexController.Intelligence, middlewaresMap, dbService, secureProperties, nil)
	routesMap["Automation"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/ghbex/automation/:owner/:repo", "application/json", mcpGHbexController.Automation, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
