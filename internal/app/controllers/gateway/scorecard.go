package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ScorecardController exposes placeholder scorecard and metrics endpoints.
type ScorecardController struct{}

func NewScorecardController() *ScorecardController {
	return &ScorecardController{}
}

// GetScorecard lists the placeholder scorecard entries exposed by the gateway.
//
// @Summary     Listar scorecard
// @Description Retorna os indicadores disponíveis enquanto a integração real não está ativa.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ScorecardResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/scorecard [get]
func (sc *ScorecardController) GetScorecard(c *gin.Context) {
	entries := []ScorecardEntry{
		{
			ID:          "demo",
			Title:       "AI Governance",
			Description: "Placeholder scorecard entry",
			Score:       0.75,
			UpdatedAt:   time.Now().UTC(),
			Tags:        []string{"placeholder", "todo"},
		},
	}
	c.JSON(http.StatusOK, ScorecardResponse{
		Items:   entries,
		Version: "gateway-placeholder-1",
	})
}

// GetScorecardAdvice returns high-level guidance associated with the scorecard snapshot.
//
// @Summary     Aconselhar scorecard
// @Description Fornece sugestões de alto nível com base na métrica atual.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ScorecardAdviceResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/scorecard/advice [get]
func (sc *ScorecardController) GetScorecardAdvice(c *gin.Context) {
	c.JSON(http.StatusOK, ScorecardAdviceResponse{
		Advice:  "Real scorecard advice pending integration with analyzer service.",
		Version: "gateway-placeholder-1",
	})
}

// GetMetrics exposes synthetic gateway metrics for observability tests.
//
// @Summary     Métricas de IA
// @Description Entrega métricas agregadas para monitoramento do gateway.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} ScorecardMetricsResponse
// @Failure     401 {object} ErrorResponse
// @Router      /api/v1/metrics/ai [get]
func (sc *ScorecardController) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, ScorecardMetricsResponse{
		Metrics: map[string]interface{}{
			"requests_last_hour": 0,
			"avg_latency_ms":     0,
			"success_rate":       1,
		},
		Version: "gateway-placeholder-1",
	})
}
