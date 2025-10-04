package mcp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	mcp_analyzer_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/analyzer"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type MCPAnalyzerRoutes struct {
	ar.IRouter
}

func NewMCPAnalyzerRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP Analyzer routes")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for OAuthRoutes")
		return nil
	}
	ctx := context.Background()
	dbCfg := dbService.GetConfig(ctx)
	if dbCfg == nil {
		gl.Log("error", "Database config is nil for OAuthRoutes")
		return nil
	}
	dbName := dbCfg.GetDBName()
	ctx = context.WithValue(ctx, gl.ContextDBNameKey, dbName)

	// Create bridge to access services
	bridge := gdbasez.NewBridge(ctx, dbService, dbName)
	mcpAnalyzerController := mcp_analyzer_controller.NewAnalyzerController(bridge)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This need to be changed to true for production
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	// Repository Analysis Routes
	routesMap["ScheduleAnalysis"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/analyzer/schedule", "application/json", mcpAnalyzerController.ScheduleAnalysis, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetAnalysisStatus"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/analyzer/status/:job_id", "application/json", mcpAnalyzerController.GetAnalysisStatus, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetAnalysisResults"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/analyzer/results/:job_id", "application/json", mcpAnalyzerController.GetAnalysisResults, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ListAnalysisJobs"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/analyzer/jobs", "application/json", mcpAnalyzerController.ListAnalysisJobs, middlewaresMap, dbService, secureProperties, nil)

	// System and Health Routes
	routesMap["GetSystemHealth"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/analyzer/health", "application/json", mcpAnalyzerController.GetSystemHealth, middlewaresMap, dbService, secureProperties, nil)

	// Notification Routes
	routesMap["SendNotification"] = proto.NewRoute(http.MethodPost, "/api/v1/analyzer/notifications/send", "application/json", mcpAnalyzerController.SendNotification, middlewaresMap, dbService, secureProperties, nil)

	gl.Log("info", "MCP Analyzer routes initialized successfully", "routes_count", len(routesMap))

	return routesMap
}
