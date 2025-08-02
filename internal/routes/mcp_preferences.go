package routes

import (
	"net/http"

	mcp_preferences_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/preferences"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

type MCPPreferencesRoutes struct {
	ar.IRouter
}

func NewMCPPreferencesRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for MCPPreferencesRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpPreferencesController := mcp_preferences_controller.NewPreferencesController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["/mcp/preferences"] = NewRoute(http.MethodGet, "/mcp/preferences", "application/json", mcpPreferencesController.GetAllPreferences, middlewaresMap, dbService)
	routesMap["/mcp/preferences/:id"] = NewRoute(http.MethodGet, "/mcp/preferences/:id", "application/json", mcpPreferencesController.GetPreferencesByID, middlewaresMap, dbService)
	routesMap["/mcp/preferences"] = NewRoute(http.MethodPost, "/mcp/preferences", "application/json", mcpPreferencesController.CreatePreferences, middlewaresMap, dbService)
	routesMap["/mcp/preferences/:id"] = NewRoute(http.MethodPut, "/mcp/preferences/:id", "application/json", mcpPreferencesController.UpdatePreferences, middlewaresMap, dbService)
	routesMap["/mcp/preferences/:id"] = NewRoute(http.MethodDelete, "/mcp/preferences/:id", "application/json", mcpPreferencesController.DeletePreferences, middlewaresMap, dbService)
	routesMap["/mcp/preferences/scope/:scope"] = NewRoute(http.MethodGet, "/mcp/preferences/scope/:scope", "application/json", mcpPreferencesController.GetPreferencesByScope, middlewaresMap, dbService)
	routesMap["/mcp/preferences/user/:userID"] = NewRoute(http.MethodGet, "/mcp/preferences/user/:userID", "application/json", mcpPreferencesController.GetPreferencesByUserID, middlewaresMap, dbService)
	routesMap["/mcp/preferences/upsert/:scope"] = NewRoute(http.MethodPost, "/mcp/preferences/upsert/:scope", "application/json", mcpPreferencesController.UpsertPreferencesByScope, middlewaresMap, dbService)

	return routesMap
}
