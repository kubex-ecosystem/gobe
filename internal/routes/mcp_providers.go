package routes

import (
	"net/http"

	mcp_providers_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/providers"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

type MCPProvidersRoutes struct {
	ar.IRouter
}

func NewMCPProvidersRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for MCPProvidersRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpProvidersController := mcp_providers_controller.NewProvidersController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["GetAllProviders"] = NewRoute(http.MethodGet, "/mcp/providers", "application/json", mcpProvidersController.GetAllProviders, middlewaresMap, dbService)
	routesMap["GetProviderByID"] = NewRoute(http.MethodGet, "/mcp/providers/:id", "application/json", mcpProvidersController.GetProviderByID, middlewaresMap, dbService)
	routesMap["DeleteProvider"] = NewRoute(http.MethodDelete, "/mcp/providers/:id", "application/json", mcpProvidersController.DeleteProvider, middlewaresMap, dbService)
	routesMap["GetActiveProviders"] = NewRoute(http.MethodGet, "/mcp/providers/active", "application/json", mcpProvidersController.GetActiveProviders, middlewaresMap, dbService)
	routesMap["CreateProvider"] = NewRoute(http.MethodPost, "/mcp/providers", "application/json", mcpProvidersController.CreateProvider, middlewaresMap, dbService)
	routesMap["UpdateProvider"] = NewRoute(http.MethodPut, "/mcp/providers/:id", "application/json", mcpProvidersController.UpdateProvider, middlewaresMap, dbService)
	routesMap["GetProvidersByProvider"] = NewRoute(http.MethodGet, "/mcp/providers/provider/:provider", "application/json", mcpProvidersController.GetProvidersByProvider, middlewaresMap, dbService)
	routesMap["GetProvidersByOrgOrGroup"] = NewRoute(http.MethodGet, "/mcp/providers/org/:org_or_group", "application/json", mcpProvidersController.GetProvidersByOrgOrGroup, middlewaresMap, dbService)
	routesMap["UpsertProviderByNameAndOrg"] = NewRoute(http.MethodPost, "/mcp/providers/upsert", "application/json", mcpProvidersController.UpsertProviderByNameAndOrg, middlewaresMap, dbService)

	return routesMap
}
