package mcp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_llm_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/llm"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
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

	routesMap["GetAllLLMModels"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/llm", "application/json", mcpLLMController.GetAllLLMModels, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetLLMModelByID"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/llm/:id", "application/json", mcpLLMController.GetLLMModelByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["CreateLLMModel"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/llm", "application/json", mcpLLMController.CreateLLMModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdateLLMModel"] = proto.NewRoute(http.MethodPut, "/api/v1/mcp/llm/:id", "application/json", mcpLLMController.UpdateLLMModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteLLMModel"] = proto.NewRoute(http.MethodDelete, "/api/v1/mcp/llm/:id", "application/json", mcpLLMController.DeleteLLMModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetLLMModelsByProvider"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/llm/provider/:provider", "application/json", mcpLLMController.GetLLMModelsByProvider, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetLLMModelByProviderAndModel"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/llm/provider/:provider/model/:model", "application/json", mcpLLMController.GetLLMModelByProviderAndModel, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetEnabledLLMModels"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/llm/enabled", "application/json", mcpLLMController.GetEnabledLLMModels, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
