package mcp

import (
	"context"
	"net/http"

	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	"github.com/gin-gonic/gin"
	mcp_scheduler_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/mcp/scheduler"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type MCPSchedulerRoutes struct {
	ar.IRouter
}

func NewMCPSchedulerRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil, cannot create MCP Scheduler routes")
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
	bridge := gdbasez.NewBridge(ctx, dbService, dbName)
	mcpSchedulerController := mcp_scheduler_controller.NewSchedulerController(bridge)

	routesMap := make(map[string]ar.IRoute)
	// middlewaresMap := rtl.GetMiddlewares()
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This need to be changed to true for production
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllScheduler"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/scheduler", "application/json", mcpSchedulerController.ListJobs, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
