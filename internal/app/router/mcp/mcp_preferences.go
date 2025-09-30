// Package mcp contains the routes for the MCP (Managed Cloud Platform).
package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_preferences_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/preferences"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type MCPPreferencesRoutes struct {
	ar.IRouter
}

func NewMCPPreferencesRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for MCPPreferencesRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCPPreferencesRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpPreferencesController := mcp_preferences_controller.NewPreferencesController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["Preferences"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/preferences", "application/json", mcpPreferencesController.GetAllPreferences, middlewaresMap, dbService, secureProperties, nil)
	routesMap["PreferencesByID"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/preferences/:id", "application/json", mcpPreferencesController.GetPreferencesByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["CreatePreferences"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/preferences", "application/json", mcpPreferencesController.CreatePreferences, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdatePreferences"] = proto.NewRoute(http.MethodPut, "/api/v1/mcp/preferences/:id", "application/json", mcpPreferencesController.UpdatePreferences, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeletePreferences"] = proto.NewRoute(http.MethodDelete, "/api/v1/mcp/preferences/:id", "application/json", mcpPreferencesController.DeletePreferences, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetPreferencesByScope"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/preferences/scope/:scope", "application/json", mcpPreferencesController.GetPreferencesByScope, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetPreferencesByUserID"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/preferences/user/:userID", "application/json", mcpPreferencesController.GetPreferencesByUserID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpsertPreferencesByScope"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/preferences/upsert/:scope", "application/json", mcpPreferencesController.UpsertPreferencesByScope, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
