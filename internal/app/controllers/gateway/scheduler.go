package gateway

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SchedulerController exposes monitoring hooks for the scheduler service (placeholder).
type SchedulerController struct{}

func NewSchedulerController() *SchedulerController { return &SchedulerController{} }

// Stats reports aggregated scheduler execution metrics.
//
// @Summary     Estatísticas do scheduler
// @Description Exibe contadores básicos e o horário da última execução.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} SchedulerStatsResponse
// @Failure     401 {object} ErrorResponse
// @Router      /health/scheduler/stats [get]
func (sc *SchedulerController) Stats(c *gin.Context) {
	now := time.Now().UTC()
	stats := SchedulerStats{
		JobsRunning:     0,
		JobsPending:     0,
		JobsCompleted:   0,
		LastRun:         &now,
		Uptime:          time.Minute * 42,
		AverageDuration: time.Second * 12,
	}
	c.JSON(http.StatusOK, SchedulerStatsResponse{
		Stats:   stats,
		Version: "gateway-placeholder-1",
	})
}

// ForceRun queues a manual execution request for the scheduler.
//
// @Summary     Forçar execução do scheduler
// @Description Agenda uma execução manual assíncrona do scheduler.
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Success     202 {object} SchedulerActionResponse
// @Failure     401 {object} ErrorResponse
// @Router      /health/scheduler/force [post]
func (sc *SchedulerController) ForceRun(c *gin.Context) {
	c.JSON(http.StatusAccepted, SchedulerActionResponse{
		Status:    "queued",
		Message:   "TODO: trigger scheduler manual run",
		Timestamp: time.Now().UTC(),
	})
}
