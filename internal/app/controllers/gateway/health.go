package gateway

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
)

// HealthController exposes health and status endpoints compatible with Analyzer gateway.
type HealthController struct {
    dbService svc.DBService
    startedAt time.Time
}

func NewHealthController(dbService svc.DBService) *HealthController {
    return &HealthController{
        dbService: dbService,
        startedAt: time.Now().UTC(),
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
    payload["version"] = "gateway-placeholder-1"
    c.JSON(http.StatusOK, payload)
}

func (hc *HealthController) buildStatusPayload() gin.H {
    uptime := time.Since(hc.startedAt)
    dbStatus := gin.H{
        "healthy": hc.checkDatabase(),
    }
    return gin.H{
        "status":  "ok",
        "uptime":  uptime.String(),
        "started": hc.startedAt,
        "services": gin.H{
            "database": dbStatus,
        },
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

