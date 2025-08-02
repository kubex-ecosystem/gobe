package routes

import (
	"net/http"

	mcp_llm_controller "github.com/rafa-mori/gobe/internal/controllers/mcp/llm"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
	l "github.com/rafa-mori/logz"
)

type MCPLLMRoutes struct {
	ar.IRouter
}

func NewMCPLLMRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for MCPLLMRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpLLMController := mcp_llm_controller.NewLLMController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["GetAllLLMModels"] = NewRoute(http.MethodGet, "/mcp/llm", "application/json", mcpLLMController.GetAllLLMModels, middlewaresMap, dbService)
	routesMap["GetLLMModelByID"] = NewRoute(http.MethodGet, "/mcp/llm/:id", "application/json", mcpLLMController.GetLLMModelByID, middlewaresMap, dbService)
	routesMap["CreateLLMModel"] = NewRoute(http.MethodPost, "/mcp/llm", "application/json", mcpLLMController.CreateLLMModel, middlewaresMap, dbService)
	routesMap["UpdateLLMModel"] = NewRoute(http.MethodPut, "/mcp/llm/:id", "application/json", mcpLLMController.UpdateLLMModel, middlewaresMap, dbService)
	routesMap["DeleteLLMModel"] = NewRoute(http.MethodDelete, "/mcp/llm/:id", "application/json", mcpLLMController.DeleteLLMModel, middlewaresMap, dbService)
	routesMap["GetLLMModelsByProvider"] = NewRoute(http.MethodGet, "/mcp/llm/provider/:provider", "application/json", mcpLLMController.GetLLMModelsByProvider, middlewaresMap, dbService)
	routesMap["GetLLMModelByProviderAndModel"] = NewRoute(http.MethodGet, "/mcp/llm/provider/:provider/model/:model", "application/json", mcpLLMController.GetLLMModelByProviderAndModel, middlewaresMap, dbService)
	routesMap["GetEnabledLLMModels"] = NewRoute(http.MethodGet, "/mcp/llm/enabled", "application/json", mcpLLMController.GetEnabledLLMModels, middlewaresMap, dbService)

	return routesMap
}
