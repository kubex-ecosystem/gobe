package gateway

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    models "github.com/kubex-ecosystem/gdbase/factory/models/mcp"
    svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
    gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
    "gorm.io/gorm"
)

// ProvidersController exposes aggregated provider information for the gateway.
type ProvidersController struct {
    service svc.ProvidersService
}

func NewProvidersController(db *gorm.DB) *ProvidersController {
    if db == nil {
        gl.Log("warn", "providers controller received nil db; responses will be unavailable")
        return &ProvidersController{}
    }
    return &ProvidersController{service: svc.NewProvidersService(models.NewProvidersRepo(db))}
}

func (pc *ProvidersController) ListProviders(c *gin.Context) {
    if pc.service == nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "providers service unavailable"})
        return
    }

    providers, err := pc.service.ListProviders()
    if err != nil {
        gl.Log("error", "failed to list providers", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list providers"})
        return
    }

    items := make([]ProviderItem, 0, len(providers))
    now := time.Now().UTC()
    for _, provider := range providers {
        item := ProviderItem{
            Name:        provider.GetProvider(),
            Provider:    provider.GetProvider(),
            Org:         provider.GetOrgOrGroup(),
            Active:      true,
            LatencyMS:   0,
            LastChecked: &now,
            Health:      "unknown",
            Metadata: map[string]interface{}{
                "config": provider.GetConfig(),
            },
        }
        items = append(items, item)
    }

    c.JSON(http.StatusOK, gin.H{
        "providers": items,
        "version":   "gateway-placeholder-1",
    })
}

