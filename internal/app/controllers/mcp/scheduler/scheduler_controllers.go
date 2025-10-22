// Package scheduler provides the controller for managing user jobs.
package scheduler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	gl "github.com/kubex-ecosystem/logz/logger"

	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler"
)

type SchedulerController struct {
	sched  scheduler.IScheduler
	bridge *gdbasez.Bridge
}

func NewSchedulerController(b *gdbasez.Bridge, s scheduler.IScheduler) *SchedulerController {
	if b == nil {
		gl.Log("error", "Bridge is nil for SchedulerController")
		return nil
	}
	return &SchedulerController{
		bridge: b,
		sched:  s,
	}
}

type CreateJobReq struct {
	Name     string `json:"name"`               // opcional; default "intent"
	Schedule string `json:"schedule,omitempty"` // ex: "*/5 * * * *" ou "@every 30s"
	Command  string `json:"command,omitempty"`  // livre; teu executor decide
}

func (sc *SchedulerController) CreateJob(c *gin.Context) {
	var req CreateJobReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.Name == "" {
		req.Name = "intent"
	}

	id := uuid.New()
	job := scheduler.NewJobImpl(id, req.Name, req.Schedule, req.Command)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := sc.sched.ScheduleJob(ctx, job)
	if err != nil {
		gl.Log("error", "schedule job failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, resp)
}

func (sc *SchedulerController) GetJob(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status, err := sc.sched.GetJobStatus(ctx, uuid.MustParse(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (sc *SchedulerController) ListJobs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if sc.sched == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "scheduler not initialized"})
		return
	}

	jobs, err := sc.sched.ListScheduledJobs(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (sc *SchedulerController) CancelJob(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	if err := sc.sched.CancelJob(ctx, uuid.MustParse(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "canceled", "job_id": id})
}

type ReschedReq struct {
	Schedule string `json:"schedule"`
}

func (sc *SchedulerController) RescheduleJob(c *gin.Context) {
	id := c.Param("id")
	var req ReschedReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Schedule == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid schedule"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := sc.sched.RescheduleJob(ctx, id, req.Schedule); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "rescheduled", "job_id": id, "schedule": req.Schedule})
}

func (sc *SchedulerController) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := sc.sched.Health(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (sc *SchedulerController) Stats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	c.JSON(http.StatusOK, sc.sched.Stats(ctx))
}

func (sc *SchedulerController) newSchedulerService(ctx context.Context) scheduler.IScheduler {
	// pool := scheduler.NewGoroutinePool(10)

	// if sc.bridge == nil {
	// 	gl.Log("error", "Bridge is nil in SchedulerController")
	// 	return nil
	// }

	// dbService := sc.bridge.DBService()
	// service := scheduler.NewCronService(dbService)
	// sched := scheduler.S

	// // sched := scheduler.NewSchedulerFunc(pool, service)
	// // if err := sched.Start(); err != nil {
	// // 	gl.Log("error", "failed to start scheduler service", err)
	// // }
	// 	return nil
	// return sched
	return nil
}
