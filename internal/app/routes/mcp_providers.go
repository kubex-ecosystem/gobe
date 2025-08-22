package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_providers_controller "github.com/rafa-mori/gobe/internal/app/controllers/mcp/providers"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
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
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpProvidersController := mcp_providers_controller.NewProvidersController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllProviders"] = NewRoute(http.MethodGet, "/api/v1/mcp/providers", "application/json", mcpProvidersController.GetAllProviders, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProviderByID"] = NewRoute(http.MethodGet, "/api/v1/mcp/providers/:id", "application/json", mcpProvidersController.GetProviderByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteProvider"] = NewRoute(http.MethodDelete, "/api/v1/mcp/providers/:id", "application/json", mcpProvidersController.DeleteProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetActiveProviders"] = NewRoute(http.MethodGet, "/api/v1/mcp/providers/active", "application/json", mcpProvidersController.GetActiveProviders, middlewaresMap, dbService, secureProperties, nil)
	routesMap["CreateProvider"] = NewRoute(http.MethodPost, "/api/v1/mcp/providers", "application/json", mcpProvidersController.CreateProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdateProvider"] = NewRoute(http.MethodPut, "/api/v1/mcp/providers/:id", "application/json", mcpProvidersController.UpdateProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProvidersByProvider"] = NewRoute(http.MethodGet, "/api/v1/mcp/providers/provider/:provider", "application/json", mcpProvidersController.GetProvidersByProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProvidersByOrgOrGroup"] = NewRoute(http.MethodGet, "/api/v1/mcp/providers/org/:org_or_group", "application/json", mcpProvidersController.GetProvidersByOrgOrGroup, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpsertProviderByNameAndOrg"] = NewRoute(http.MethodPost, "/api/v1/mcp/providers/upsert", "application/json", mcpProvidersController.UpsertProviderByNameAndOrg, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
