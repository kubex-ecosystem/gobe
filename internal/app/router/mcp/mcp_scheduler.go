package mcp

import (
	"context"
	"net/http"

	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	sch "github.com/kubex-ecosystem/gobe/internal/services/scheduler"

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

	dbService := rtl.GetDatabaseService().(*svc.DBServiceImpl)
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
	gorPool := sch.NewGoroutinePool(100)
	cronSvc := sch.NewCronService(dbService)

	// scheduler := sch.NewSchedulerFunc(
	// 	gorPool,
	// 	cronSvc,
	// )
	iScheduler := sch.NewScheduler(ctx, gorPool, cronSvc)

	mcpSchedulerController := mcp_scheduler_controller.NewSchedulerController(bridge, iScheduler)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := rtl.GetMiddlewares()
	// middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = false // This need to be changed to true for production
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["SchedulerList"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/scheduler", "application/json", mcpSchedulerController.ListJobs, middlewaresMap, dbService, secureProperties, nil)
	routesMap["SchedulerGet"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/scheduler/:id", "application/json", mcpSchedulerController.GetJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["SchedulerStats"] = proto.NewRoute(http.MethodGet, "/api/v1/mcp/scheduler/stats", "application/json", mcpSchedulerController.Stats, middlewaresMap, dbService, secureProperties, nil)
	routesMap["SchedulerCancel"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/scheduler/cancel", "application/json", mcpSchedulerController.CancelJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["SchedulerCreate"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/scheduler/new", "application/json", mcpSchedulerController.CreateJob, middlewaresMap, dbService, secureProperties, nil)
	routesMap["SchedulerReschedule"] = proto.NewRoute(http.MethodPost, "/api/v1/mcp/scheduler/reschedule", "application/json", mcpSchedulerController.RescheduleJob, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
