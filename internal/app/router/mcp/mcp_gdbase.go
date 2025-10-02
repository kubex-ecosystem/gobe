package mcp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

	mcp_gdbase_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/gdbase"
)

type MCPGDBaseRoutes struct {
	ar.IRouter
}

func NewMCPGDBaseRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP GDBase routes")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCP GDBase Route")
		return nil
	}
	dbGorm, err := dbService.GetDB(context.Background(), gdbasez.DefaultDBName)
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpGDBaseController := mcp_gdbase_controller.NewGDBaseController(dbGorm)

	routesMap := make(map[string]ar.IRoute)

	middlewaresMap := make(map[string]gin.HandlerFunc)
	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["DBTunnelUp"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/db/tunnel/up", "application/json", mcpGDBaseController.PostGDBaseTunnelUp, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DBTunnelDown"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/db/tunnel/down", "application/json", mcpGDBaseController.PostGDBaseTunnelDown, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DBTunnelStatus"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/db/tunnel/status", "application/json", mcpGDBaseController.GetGDBaseTunnelStatus, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
