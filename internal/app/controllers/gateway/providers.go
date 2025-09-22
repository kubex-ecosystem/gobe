package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	gatewaysvc "github.com/kubex-ecosystem/gobe/internal/services/gateway/registry"
)

// ProvidersController exposes aggregated provider information for the gateway.
type ProvidersController struct {
	service *gatewaysvc.Service
}

func NewProvidersController(service *gatewaysvc.Service) *ProvidersController {
	return &ProvidersController{service: service}
}

// ListProviders returns availability metadata for gateway providers.
//
// @Summary     Listar provedores
// @Description Exibe os provedores configurados e seus estados de disponibilidade. [Em desenvolvimento]
// @Tags        gateway beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ProvidersResponse
// @Failure     401 {object} ErrorResponse
// @Failure     503 {object} ErrorResponse
// @Router      /providers [get]
func (pc *ProvidersController) ListProviders(c *gin.Context) {
	if pc.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "providers service unavailable"})
		return
	}

	summaries := pc.service.ProviderSummaries()
	items := make([]ProviderItem, 0, len(summaries))

	for _, summary := range summaries {
		items = append(items, ProviderItem{
			Name:         summary.Name,
			Type:         summary.Type,
			Org:          summary.Org,
			DefaultModel: summary.DefaultModel,
			Available:    summary.Available,
			LastError:    summary.LastError,
			Metadata:     summary.Metadata,
		})
	}

	c.JSON(http.StatusOK, ProvidersResponse{
		Providers: items,
		Timestamp: time.Now().UTC(),
	})
}
