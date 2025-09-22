package gateway

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

// SchedulerController exposes monitoring hooks for the scheduler service (placeholder).
type SchedulerController struct{}

func NewSchedulerController() *SchedulerController { return &SchedulerController{} }

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
    c.JSON(http.StatusOK, gin.H{
        "stats":   stats,
        "version": "gateway-placeholder-1",
    })
}

func (sc *SchedulerController) ForceRun(c *gin.Context) {
    c.JSON(http.StatusAccepted, gin.H{
        "status":    "queued",
        "message":   "TODO: trigger scheduler manual run",
        "timestamp": time.Now().UTC(),
    })
}

