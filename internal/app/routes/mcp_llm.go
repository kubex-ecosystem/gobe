package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_llm_controller "github.com/rafa-mori/gobe/internal/app/controllers/mcp/llm"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
)

type MCPLLMRoutes struct {
	ar.IRouter
}

func NewMCPLLMRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP LLM routes")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for MCPLLMRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	mcpLLMController := mcp_llm_controller.NewLLMController(dbGorm)

	routesMap := make(map[string]ar.IRoute)

	middlewaresMap := make(map[string]gin.HandlerFunc)
	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllLLMModels"] = NewRoute(http.MethodGet, "/api/v1/mcp/llm", "application/json", mcpLLMController.GetAllLLMModels, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetLLMModelByID"] = NewRoute(http.MethodGet, "/api/v1/mcp/llm/:id", "application/json", mcpLLMController.GetLLMModelByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["CreateLLMModel"] = NewRoute(http.MethodPost, "/api/v1/mcp/llm", "application/json", mcpLLMController.CreateLLMModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdateLLMModel"] = NewRoute(http.MethodPut, "/api/v1/mcp/llm/:id", "application/json", mcpLLMController.UpdateLLMModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteLLMModel"] = NewRoute(http.MethodDelete, "/api/v1/mcp/llm/:id", "application/json", mcpLLMController.DeleteLLMModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetLLMModelsByProvider"] = NewRoute(http.MethodGet, "/api/v1/mcp/llm/provider/:provider", "application/json", mcpLLMController.GetLLMModelsByProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetLLMModelByProviderAndModel"] = NewRoute(http.MethodGet, "/api/v1/mcp/llm/provider/:provider/model/:model", "application/json", mcpLLMController.GetLLMModelByProviderAndModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetEnabledLLMModels"] = NewRoute(http.MethodGet, "/api/v1/mcp/llm/enabled", "application/json", mcpLLMController.GetEnabledLLMModels, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
