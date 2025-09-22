package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gatewaysvc "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

// HealthController exposes health and status endpoints compatible with Analyzer gateway.
type HealthController struct {
	dbService      svc.DBService
	gatewayService *gatewaysvc.Service
	startedAt      time.Time
}

// GatewayProvidersStatus encapsulates aggregated provider availability information.
type GatewayProvidersStatus struct {
	Total       int `json:"total"`
	Available   int `json:"available"`
	Unavailable int `json:"unavailable"`
}

// GatewayServiceHealth reports the health of a specific gateway dependency.
type GatewayServiceHealth struct {
	Healthy bool                    `json:"healthy"`
	Detail  *GatewayProvidersStatus `json:"detail,omitempty"`
}

// GatewayHealthResponse is the primary schema returned by gateway health endpoints.
type GatewayHealthResponse struct {
	Status    string                          `json:"status"`
	Timestamp *time.Time                      `json:"timestamp,omitempty"`
	Uptime    string                          `json:"uptime,omitempty"`
	Started   *time.Time                      `json:"started,omitempty"`
	Version   string                          `json:"version,omitempty"`
	Services  map[string]GatewayServiceHealth `json:"services,omitempty"`
}

func NewHealthController(dbService svc.DBService, gatewayService *gatewaysvc.Service) *HealthController {
	return &HealthController{
		dbService:      dbService,
		gatewayService: gatewayService,
		startedAt:      time.Now().UTC(),
	}
}

// Healthz provides a lightweight readiness probe for upstream load balancers.
//
// @Summary     Healthcheck
// @Description Validates service availability for gateway integrations.
// @Tags        health
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} GatewayHealthResponse
// @Failure     500 {object} ErrorResponse
// @Router      /healthz [get]
func (hc *HealthController) Healthz(c *gin.Context) {
	now := time.Now().UTC()
	c.JSON(http.StatusOK, GatewayHealthResponse{
		Status:    "ok",
		Timestamp: &now,
	})
}

// Status returns dependencies and uptime information for monitoring dashboards.
//
// @Summary     Service status
// @Description Reports uptime and dependency health for the gateway module.
// @Tags        health
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} GatewayHealthResponse
// @Failure     500 {object} ErrorResponse
// @Router      /status [get]
func (hc *HealthController) Status(c *gin.Context) {
	payload := hc.buildStatusPayload()
	c.JSON(http.StatusOK, payload)
}

// APIHealth exposes health data for clients interfacing through the API gateway.
//
// @Summary     Gateway health
// @Description Augments status payload with module version for API consumers.
// @Tags        health
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} GatewayHealthResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/health [get]
func (hc *HealthController) APIHealth(c *gin.Context) {
	payload := hc.buildStatusPayload()
	payload.Version = "gateway"
	c.JSON(http.StatusOK, payload)
}

func (hc *HealthController) buildStatusPayload() GatewayHealthResponse {
	uptime := time.Since(hc.startedAt)
	now := time.Now().UTC()
	services := map[string]GatewayServiceHealth{
		"database": {
			Healthy: hc.checkDatabase(),
		},
	}

	if hc.gatewayService != nil {
		summaries := hc.gatewayService.ProviderSummaries()
		available := 0
		for _, summary := range summaries {
			if summary.Available {
				available++
			}
		}
		services["providers"] = GatewayServiceHealth{
			Healthy: available > 0,
			Detail: &GatewayProvidersStatus{
				Total:       len(summaries),
				Available:   available,
				Unavailable: len(summaries) - available,
			},
		}
	}

	started := hc.startedAt

	return GatewayHealthResponse{
		Status:    "ok",
		Timestamp: &now,
		Uptime:    uptime.String(),
		Started:   &started,
		Services:  services,
	}
}

func (hc *HealthController) checkDatabase() bool {
	if hc.dbService == nil {
		return false
	}
	if err := hc.dbService.CheckDatabaseHealth(); err != nil {
		return false
	}
	return true
}
