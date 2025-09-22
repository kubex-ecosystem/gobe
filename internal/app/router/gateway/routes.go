package gateway

import (
    "net/http"

    "github.com/gin-gonic/gin"
    gatewayController "github.com/kubex-ecosystem/gobe/internal/app/controllers/gateway"
    proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
    ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
    gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
    "gorm.io/gorm"
)

type GatewayRoutes struct {
    ar.IRouter
}

func NewGatewayRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
    if rtr == nil {
        gl.Log("error", "Router is nil for GatewayRoutes")
        return nil
    }
    rtl := *rtr

    dbService := rtl.GetDatabaseService()
    var db *gorm.DB
    if dbService != nil {
        var err error
        db, err = dbService.GetDB()
        if err != nil {
            gl.Log("warn", "Failed to fetch DB for gateway module", err)
        }
    } else {
        gl.Log("warn", "Database service is nil for GatewayRoutes")
    }

    chatController := gatewayController.NewChatController()
    providersController := gatewayController.NewProvidersController(db)
    adviseController := gatewayController.NewAdviseController()
    scorecardController := gatewayController.NewScorecardController()
    healthController := gatewayController.NewHealthController(dbService)
    lookAtniController := gatewayController.NewLookAtniController()
    webhookController := gatewayController.NewWebhookController()
    schedulerController := gatewayController.NewSchedulerController()

    webRoot := ""
    if prop := rtl.GetProperty("gateway.web.root"); prop != nil {
        if path, ok := prop.(string); ok {
            webRoot = path
        }
    }
    webUIController := gatewayController.NewWebUIController(webRoot)

    routes := make(map[string]ar.IRoute)
    middlewaresMap := make(map[string]gin.HandlerFunc)
    secure := func(secure bool) map[string]bool {
        return map[string]bool{
            "secure":                  secure,
            "validateAndSanitize":     false,
            "validateAndSanitizeBody": false,
        }
    }

    routes["Healthz"] = proto.NewRoute(http.MethodGet, "/healthz", "application/json", healthController.Healthz, middlewaresMap, dbService, secure(true), nil)
    routes["Status"] = proto.NewRoute(http.MethodGet, "/status", "application/json", healthController.Status, middlewaresMap, dbService, secure(true), nil)
    routes["APIHealth"] = proto.NewRoute(http.MethodGet, "/api/v1/health", "application/json", healthController.APIHealth, middlewaresMap, dbService, secure(true), nil)

    routes["ChatSSE"] = proto.NewRoute(http.MethodPost, "/chat", "text/event-stream", chatController.ChatSSE, middlewaresMap, dbService, secure(true), nil)

    routes["Providers"] = proto.NewRoute(http.MethodGet, "/providers", "application/json", providersController.ListProviders, middlewaresMap, dbService, secure(true), nil)

    routes["AdviseV1"] = proto.NewRoute(http.MethodPost, "/v1/advise", "text/event-stream", adviseController.Advise, middlewaresMap, dbService, secure(true), nil)
    routes["AdviseLegacy"] = proto.NewRoute(http.MethodPost, "/advise", "text/event-stream", adviseController.Advise, middlewaresMap, dbService, secure(true), nil)

    routes["Scorecard"] = proto.NewRoute(http.MethodGet, "/api/v1/scorecard", "application/json", scorecardController.GetScorecard, middlewaresMap, dbService, secure(true), nil)
    routes["ScorecardAdvice"] = proto.NewRoute(http.MethodGet, "/api/v1/scorecard/advice", "application/json", scorecardController.GetScorecardAdvice, middlewaresMap, dbService, secure(true), nil)
    routes["MetricsAI"] = proto.NewRoute(http.MethodGet, "/api/v1/metrics/ai", "application/json", scorecardController.GetMetrics, middlewaresMap, dbService, secure(true), nil)

    routes["LookAtniExtract"] = proto.NewRoute(http.MethodPost, "/api/v1/lookatni/extract", "application/json", lookAtniController.Extract, middlewaresMap, dbService, secure(true), nil)
    routes["LookAtniArchive"] = proto.NewRoute(http.MethodPost, "/api/v1/lookatni/archive", "application/json", lookAtniController.Archive, middlewaresMap, dbService, secure(true), nil)
    routes["LookAtniDownload"] = proto.NewRoute(http.MethodGet, "/api/v1/lookatni/download/:id", "application/json", lookAtniController.Download, middlewaresMap, dbService, secure(true), nil)
    routes["LookAtniProjects"] = proto.NewRoute(http.MethodGet, "/api/v1/lookatni/projects", "application/json", lookAtniController.Projects, middlewaresMap, dbService, secure(true), nil)

    routes["Webhooks"] = proto.NewRoute(http.MethodPost, "/v1/webhooks", "application/json", webhookController.Handle, middlewaresMap, dbService, secure(true), nil)
    routes["WebhooksHealth"] = proto.NewRoute(http.MethodGet, "/v1/webhooks/health", "application/json", webhookController.Health, middlewaresMap, dbService, secure(true), nil)

    routes["SchedulerStats"] = proto.NewRoute(http.MethodGet, "/health/scheduler/stats", "application/json", schedulerController.Stats, middlewaresMap, dbService, secure(true), nil)
    routes["SchedulerForce"] = proto.NewRoute(http.MethodPost, "/health/scheduler/force", "application/json", schedulerController.ForceRun, middlewaresMap, dbService, secure(true), nil)

    routes["WebUIRoot"] = proto.NewRoute(http.MethodGet, "/", "text/html", webUIController.ServeRoot, middlewaresMap, dbService, secure(false), nil)
    routes["WebUIApp"] = proto.NewRoute(http.MethodGet, "/app/*path", "text/html", webUIController.ServeApp, middlewaresMap, dbService, secure(false), nil)

    return routes
}

