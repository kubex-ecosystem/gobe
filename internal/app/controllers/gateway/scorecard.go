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
    c.JSON(http.StatusOK, gin.H{
        "items":   entries,
        "version": "gateway-placeholder-1",
    })
}

func (sc *ScorecardController) GetScorecardAdvice(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "advice": "Real scorecard advice pending integration with analyzer service.",
        "version": "gateway-placeholder-1",
    })
}

func (sc *ScorecardController) GetMetrics(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "metrics": map[string]interface{}{
            "requests_last_hour": 0,
            "avg_latency_ms":     0,
            "success_rate":       1,
        },
        "version": "gateway-placeholder-1",
    })
}

