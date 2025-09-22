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

func NewHealthController(dbService svc.DBService, gatewayService *gatewaysvc.Service) *HealthController {
	return &HealthController{
		dbService:      dbService,
		gatewayService: gatewayService,
		startedAt:      time.Now().UTC(),
	}
}

func (hc *HealthController) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func (hc *HealthController) Status(c *gin.Context) {
	c.JSON(http.StatusOK, hc.buildStatusPayload())
}

func (hc *HealthController) APIHealth(c *gin.Context) {
	payload := hc.buildStatusPayload()
	payload["version"] = "gateway"
	c.JSON(http.StatusOK, payload)
}

func (hc *HealthController) buildStatusPayload() gin.H {
	uptime := time.Since(hc.startedAt)
	services := gin.H{
		"database": gin.H{"healthy": hc.checkDatabase()},
	}

	if hc.gatewayService != nil {
		summaries := hc.gatewayService.ProviderSummaries()
		available := 0
		for _, summary := range summaries {
			if summary.Available {
				available++
			}
		}
		services["providers"] = gin.H{
			"total":       len(summaries),
			"available":   available,
			"unavailable": len(summaries) - available,
		}
	}

	return gin.H{
		"status":  "ok",
		"uptime":  uptime.String(),
		"started": hc.startedAt,
		"services": services,
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

