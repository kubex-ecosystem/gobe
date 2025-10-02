package mcp

import (
	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_providers_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/providers"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type MCPProvidersRoutes struct {
	ar.IRouter
}

func NewMCPProvidersRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for MCPProvidersRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCPProvidersRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB(nil)
	bridge := gdbasez.NewBridge(dbGorm)
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpProvidersController := mcp_providers_controller.NewProvidersController(bridge)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllProviders"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/providers", "application/json", mcpProvidersController.GetAllProviders, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProviderByID"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/providers/:id", "application/json", mcpProvidersController.GetProviderByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteProvider"] = proto.NewRoute(http.MethodDelete, "/api/v1/mcp/providers/:id", "application/json", mcpProvidersController.DeleteProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetActiveProviders"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/providers/active", "application/json", mcpProvidersController.GetActiveProviders, middlewaresMap, dbService, secureProperties, nil)
	routesMap["CreateProvider"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/providers", "application/json", mcpProvidersController.CreateProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdateProvider"] = proto.NewRoute(http.MethodPut, "/api/v1/mcp/providers/:id", "application/json", mcpProvidersController.UpdateProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProvidersByProvider"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/providers/provider/:provider", "application/json", mcpProvidersController.GetProvidersByProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProvidersByOrgOrGroup"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/providers/org/:org_or_group", "application/json", mcpProvidersController.GetProvidersByOrgOrGroup, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpsertProviderByNameAndOrg"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/providers/upsert", "application/json", mcpProvidersController.UpsertProviderByNameAndOrg, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
