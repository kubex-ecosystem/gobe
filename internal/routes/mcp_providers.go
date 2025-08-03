package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_providers_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/providers"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
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

	routesMap["GetAllProviders"] = NewRoute(http.MethodGet, "/mcp/providers", "application/json", mcpProvidersController.GetAllProviders, middlewaresMap, dbService, secureProperties)
	routesMap["GetProviderByID"] = NewRoute(http.MethodGet, "/mcp/providers/:id", "application/json", mcpProvidersController.GetProviderByID, middlewaresMap, dbService, secureProperties)
	routesMap["DeleteProvider"] = NewRoute(http.MethodDelete, "/mcp/providers/:id", "application/json", mcpProvidersController.DeleteProvider, middlewaresMap, dbService, secureProperties)
	routesMap["GetActiveProviders"] = NewRoute(http.MethodGet, "/mcp/providers/active", "application/json", mcpProvidersController.GetActiveProviders, middlewaresMap, dbService, secureProperties)
	routesMap["CreateProvider"] = NewRoute(http.MethodPost, "/mcp/providers", "application/json", mcpProvidersController.CreateProvider, middlewaresMap, dbService, secureProperties)
	routesMap["UpdateProvider"] = NewRoute(http.MethodPut, "/mcp/providers/:id", "application/json", mcpProvidersController.UpdateProvider, middlewaresMap, dbService, secureProperties)
	routesMap["GetProvidersByProvider"] = NewRoute(http.MethodGet, "/mcp/providers/provider/:provider", "application/json", mcpProvidersController.GetProvidersByProvider, middlewaresMap, dbService, secureProperties)
	routesMap["GetProvidersByOrgOrGroup"] = NewRoute(http.MethodGet, "/mcp/providers/org/:org_or_group", "application/json", mcpProvidersController.GetProvidersByOrgOrGroup, middlewaresMap, dbService, secureProperties)
	routesMap["UpsertProviderByNameAndOrg"] = NewRoute(http.MethodPost, "/mcp/providers/upsert", "application/json", mcpProvidersController.UpsertProviderByNameAndOrg, middlewaresMap, dbService, secureProperties)

	return routesMap
}
