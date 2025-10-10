// Package gateway defines the routes for the gateway module.
package gateway

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	analyzergateway "github.com/kubex-ecosystem/analyzer/factory/gateway"
	gatewayController "github.com/kubex-ecosystem/gobe/internal/app/controllers/gateway"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	gatewaysvc "github.com/kubex-ecosystem/gobe/internal/services/gateway/registry"
	webhooksvc "github.com/kubex-ecosystem/gobe/internal/services/webhooks"
	messagery "github.com/kubex-ecosystem/gobe/internal/sockets/messagery"
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
	if dbService == nil {
		gl.Log("warn", "Database service is nil for GatewayRoutes")
		return nil
	}

	var gatewayService *gatewaysvc.Service
	var webhookService *webhooksvc.WebhookService

	providersSvc := svc.NewProvidersService(svc.NewProvidersRepo(context.Background(), dbService.(*gdbasez.DBServiceImpl)))
	gw, err := gatewaysvc.NewService(providersSvc)
	if err != nil {
		gl.Log("error", "failed to initialize gateway service", err)
	} else {
		gatewayService = gw
	}

	// Initialize webhook service with AMQP connection
	amqp := messagery.NewAMQP()
	webhookService = webhooksvc.NewWebhookService(amqp)

	chatController := gatewayController.NewChatController(gatewayService)
	providersController := gatewayController.NewProvidersController(gatewayService)
	adviseController := gatewayController.NewAdviseController(gatewayService)
	scorecardController := gatewayController.NewScorecardController(dbService.(*gdbasez.DBServiceImpl))
	healthController := gatewayController.NewHealthController(dbService, gatewayService)
	lookAtniController := gatewayController.NewLookAtniController(dbService.(*gdbasez.DBServiceImpl))
	webhookController := gatewayController.NewWebhookController(webhookService)
	schedulerController := gatewayController.NewSchedulerController()

	webUIController := gatewayController.NewWebUIController()

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
	routes["WebhooksEventsList"] = proto.NewRoute(http.MethodGet, "/v1/webhooks/events", "application/json", webhookController.ListEvents, middlewaresMap, dbService, secure(true), nil)
	routes["WebhooksEventsGet"] = proto.NewRoute(http.MethodGet, "/v1/webhooks/events/:id", "application/json", webhookController.GetEvent, middlewaresMap, dbService, secure(true), nil)
	routes["WebhooksRetry"] = proto.NewRoute(http.MethodPost, "/v1/webhooks/retry", "application/json", webhookController.RetryFailedEvents, middlewaresMap, dbService, secure(true), nil)

	routes["SchedulerStats"] = proto.NewRoute(http.MethodGet, "/health/scheduler/stats", "application/json", schedulerController.Stats, middlewaresMap, dbService, secure(true), nil)
	routes["SchedulerForce"] = proto.NewRoute(http.MethodPost, "/health/scheduler/force", "application/json", schedulerController.ForceRun, middlewaresMap, dbService, secure(true), nil)

	// Web UI Favicon
	routes["WebUIFavicon"] = proto.NewRoute(http.MethodGet, "/favicon.ico", "image/x-icon", webUIController.ServeFavicon, middlewaresMap, dbService, secure(false), nil)
	// Web UI Root
	routes["WebUIRoot"] = proto.NewRoute(http.MethodGet, "", "text/html", webUIController.ServeRoot, middlewaresMap, dbService, secure(false), nil)
	// Web UI Static Assets
	routes["WebUIAssets"] = proto.NewRoute(http.MethodGet, "/assets/*path", "application/octet-stream", webUIController.ServeAssets, middlewaresMap, dbService, secure(false), nil)
	// Route for serving the main app (SPA)
	routes["WebUIApp"] = proto.NewRoute(http.MethodGet, "/app/*path", "text/html", webUIController.ServeApp, middlewaresMap, dbService, secure(false), nil)

	// Integrate GemX Analyzer when available
	if handler := initializeAnalyzerHandler(); handler != nil {
		wrap := func() gin.HandlerFunc { return gin.WrapH(handler) }

		routes["Scorecard"] = proto.NewRoute(http.MethodGet, "/api/v1/scorecard", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["ScorecardAdvice"] = proto.NewRoute(http.MethodPost, "/api/v1/scorecard/advice", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["MetricsAI"] = proto.NewRoute(http.MethodGet, "/api/v1/metrics/ai", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["APIHealth"] = proto.NewRoute(http.MethodGet, "/api/v1/health", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)

		routes["LookAtniExtract"] = proto.NewRoute(http.MethodPost, "/api/v1/lookatni/extract", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["LookAtniArchive"] = proto.NewRoute(http.MethodPost, "/api/v1/lookatni/archive", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["LookAtniDownload"] = proto.NewRoute(http.MethodGet, "/api/v1/lookatni/download/:id", "application/zip", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["LookAtniProjects"] = proto.NewRoute(http.MethodGet, "/api/v1/lookatni/projects", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["LookAtniProjectFragments"] = proto.NewRoute(http.MethodGet, "/api/v1/lookatni/projects/*path", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)

		routes["Webhooks"] = proto.NewRoute(http.MethodPost, "/v1/webhooks", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["WebhooksHealth"] = proto.NewRoute(http.MethodGet, "/v1/webhooks/health", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["WebhooksEventsList"] = proto.NewRoute(http.MethodGet, "/v1/webhooks/events", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["WebhooksEventsGet"] = proto.NewRoute(http.MethodGet, "/v1/webhooks/events/:id", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["WebhooksRetry"] = proto.NewRoute(http.MethodPost, "/v1/webhooks/retry", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["SchedulerStats"] = proto.NewRoute(http.MethodGet, "/health/scheduler/stats", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
		routes["SchedulerForce"] = proto.NewRoute(http.MethodPost, "/health/scheduler/force", "application/json", wrap(), middlewaresMap, dbService, secure(true), nil)
	}

	return routes
}

func initializeAnalyzerHandler() http.Handler {
	configPath := analyzerProvidersConfigPath()
	if configPath == "" {
		return nil
	}

	server, err := analyzergateway.NewServer(&analyzergateway.ServerConfig{
		Addr:            ":0",
		ProvidersConfig: configPath,
		Debug:           os.Getenv("ANALYZER_DEBUG") == "true",
		EnableCORS:      false,
	})
	if err != nil {
		gl.Log("error", "Failed to initialize analyzer gateway server", err)
		return nil
	}

	handler, err := server.Handler()
	if err != nil {
		gl.Log("error", "Failed to build analyzer gateway handler", err)
		return nil
	}

	return handler
}

func analyzerProvidersConfigPath() string {
	if cfg := os.Getenv("ANALYZER_PROVIDERS_CFG"); cfg != "" {
		return cfg
	}
	fallbacks := []string{
		filepath.Join("config", "analyzer_providers.yml"),
		filepath.Join("..", "analyzer", "config", "config.yml"),
	}
	for _, candidate := range fallbacks {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	gl.Log("warn", "Analyzer providers config not found; skip analyzer integration")
	return ""
}
