// Package cron provides the controller for managing cron jobs in the application.
package cron

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cron "github.com/rafa-mori/gdbase/factory/models"
	"github.com/rafa-mori/gobe/internal/types"
	gl "github.com/rafa-mori/gobe/logger"
	"gorm.io/gorm"
)

type CronController struct {
	ICronService cron.CronJobService
	APIWrapper   *types.APIWrapper[cron.CronJobModel]
}

func NewCronJobController(db *gorm.DB) *CronController {
	return &CronController{

		ICronService: cron.NewCronJobService(cron.NewCronJobRepo(context.Background(), db)),
		APIWrapper:   types.NewApiWrapper[cron.CronJobModel](),
	}
}

func (cc *CronController) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/cron")
	{
		api.GET("/", cc.GetAllCronJobs)
		api.GET("/:id", cc.GetCronJobByID)
		api.POST("/", cc.CreateCronJob)
		api.PUT("/:id", cc.UpdateCronJob)
		api.DELETE("/:id", cc.DeleteCronJob)
		api.POST("/:id/enable", cc.EnableCronJob)
		api.POST("/:id/disable", cc.DisableCronJob)
		api.POST("/:id/execute", cc.ExecuteCronJobManually)
		api.POST("/:id/reschedule", cc.RescheduleCronJob)
		api.GET("/queue", cc.GetJobQueue)
		api.POST("/reprocess-failed", cc.ReprocessFailedJobs)
		api.GET("/:id/logs", cc.GetExecutionLogs)
	}
}

// @Summary Cron Job Management
// @Description This controller manages cron jobs, allowing users to create, update, delete, and execute cron jobs.
// @Tags cron
// @Schemes http https
// @Summary Get All Cron Jobs
// @Description Retrieves all cron jobs in the system.
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse[[]cron.CronJobModel]
// @Failure 500 {string} Failed to fetch cron jobs
// @Router /cronjobs [get]
func (cc *CronController) GetAllCronJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	jobs, err := cc.ICronService.ListCronJobs(ctx)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to fetch cron jobs", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron jobs fetched successfully", "", jobs, nil, http.StatusOK)
}

// @Summary Get Cron Job by ID
// @Description Retrieves a specific cron job by its ID.
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse[cron.CronJobModel]
// @Failure 404 {string} Cron job not found
// @Router /cronjobs/{id} [get]
func (cc *CronController) GetCronJobByID(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job not found", "", nil, nil, http.StatusNotFound)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job fetched successfully", "", job, nil, http.StatusOK)
}

// @Summary Create Cron Job
// @Description Creates a new cron job.
// @Accept json
// @Produce json
// @Success 201 {object} types.APIResponse[cron.CronJobModel]
// @Failure 400 {string} Failed to create cron job
// @Router /cronjobs [post]
func (cc *CronController) CreateCronJob(c *gin.Context) {
	var job *cron.CronJobModel
	if err := c.ShouldBindJSON(&job); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to bind JSON: %s", err))
		cc.APIWrapper.JSONResponse(c, "error", "Invalid request payload", "", nil, nil, http.StatusBadRequest)
		return
	}
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	job.ID, err = uuid.NewRandom()
	if err != nil {
		gl.Log("error", fmt.Sprintf("Failed to generate UUID: %s", err))
		cc.APIWrapper.JSONResponse(c, "error", "Failed to generate UUID", "", nil, nil, http.StatusInternalServerError)
		return
	}
	job.UserID = ctx.Value("userID").(uuid.UUID)
	if job.UserID == uuid.Nil {
		gl.Log("error", "User ID is required")
		cc.APIWrapper.JSONResponse(c, "error", "User ID is required", "", nil, nil, http.StatusBadRequest)
		return
	}
	job.LastRunStatus = "pending"
	createdJob, err := cc.ICronService.CreateCronJob(ctx, job)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to create cron job", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job created successfully", "", createdJob, nil, http.StatusCreated)
}

// @Summary Update Cron Job
// @Description Updates an existing cron job.
// @Accept json
// @Produce json
// @Success 200 {object} types.APIResponse[cron.CronJobModel]
// @Failure 400 {string} Failed to update cron job
// @Failure 404 {string} Cron job not found
// @Failure 500 {string} Failed to update cron job
// @Router /cronjobs/{id} [put]
func (cc *CronController) UpdateCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	var job cron.CronJobModel
	if err := c.ShouldBindJSON(&job); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid request payload", "", nil, nil, http.StatusBadRequest)
		return
	}
	if cronID == uuid.Nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job ID is required", "", nil, nil, http.StatusBadRequest)
		return
	}
	job.ID = cronID
	updatedJob, err := cc.ICronService.UpdateCronJob(ctx, &job)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to update cron job", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job updated successfully", "", updatedJob, nil, http.StatusOK)
}

// @Summary Delete Cron Job
// @Description Deletes a specific cron job by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Cron job deleted successfully
// @Failure 404 {string} Cron job not found
// @Failure 400 {string} Invalid cron job ID
// @Failure 500 {string} Failed to delete cron job
// @Router /cronjobs/{id} [delete]
func (cc *CronController) DeleteCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job not found", "", nil, nil, http.StatusNotFound)
		return
	}
	if job.UserID != uuid.Nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job is associated with a user and cannot be deleted", "", nil, nil, http.StatusBadRequest)
		return
	}
	// Check if the cron job is currently running
	if job.LastRunStatus == "running" {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job is currently running and cannot be deleted", "", nil, nil, http.StatusBadRequest)
		return
	}
	// Check if the cron job has any pending executions
	if job.LastRunStatus == "pending" {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job has pending executions and cannot be deleted", "", nil, nil, http.StatusBadRequest)
		return
	}

	if err := cc.ICronService.DeleteCronJob(ctx, cronID); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to delete cron job", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job deleted successfully", "", nil, nil, http.StatusOK)
}

// @Summary Enable Cron Job
// @Description Enables a specific cron job by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Cron job enabled successfully
// @Failure 404 {string} Cron job not found
// @Router /cronjobs/{id}/enable [post]
func (cc *CronController) EnableCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	if err := cc.ICronService.EnableCronJob(ctx, cronID); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to enable cron job", "", nil, nil, http.StatusInternalServerError)
		return
	}

	cc.APIWrapper.JSONResponse(c, "success", "Cron job enabled successfully", "", nil, nil, http.StatusOK)
}

// @Summary Disable Cron Job
// @Description Disables a specific cron job by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Cron job disabled successfully
// @Failure 404 {string} Cron job not found
// @Router /cronjobs/{id}/disable [post]
func (cc *CronController) DisableCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	if err := cc.ICronService.DisableCronJob(ctx, cronID); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to disable cron job", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job disabled successfully", "", nil, nil, http.StatusOK)
}

// @Summary Execute Cron Job Manually
// @Description Executes a specific cron job manually by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Cron job executed successfully
// @Failure 404 {string} Cron job not found
// @Router /cronjobs/{id}/execute [post]
func (cc *CronController) ExecuteCronJobManually(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	if err := cc.ICronService.ExecuteCronJobManually(ctx, cronID); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to execute cron job manually", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job executed successfully", "", nil, nil, http.StatusOK)
}

// @Summary Execute Cron Job Manually by ID
// @Description Executes a specific cron job manually by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Cron job executed successfully
// @Failure 404 {string} Cron job not found
// @Router /cronjobs/{id}/execute/{job_id} [post]
func (cc *CronController) ExecuteCronJobManuallyByID(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	// Check if the cron job is currently running
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job not found", "", nil, nil, http.StatusNotFound)
		return
	}
	if job.LastRunStatus == "running" {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job is currently running and cannot be executed manually", "", nil, nil, http.StatusBadRequest)
		return
	}
	if err := cc.ICronService.ExecuteCronJobManually(ctx, cronID); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to execute cron job manually", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job executed successfully", "", nil, nil, http.StatusOK)
}

// @Summary Reschedule Cron Job
// @Description Reschedules a specific cron job by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Cron job rescheduled successfully
// @Failure 400 {string} Invalid cron job ID
// @Failure 404 {string} Cron job not found
// @Failure 500 {string} Failed to reschedule cron job
// @Router /cronjobs/{id}/reschedule [post]
func (cc *CronController) RescheduleCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	var payload struct {
		NewExpression string `json:"new_expression" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid request payload", "", nil, nil, http.StatusBadRequest)
		return
	}
	if cronID == uuid.Nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job ID is required", "", nil, nil, http.StatusBadRequest)
		return
	}
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job not found", "", nil, nil, http.StatusNotFound)
		return
	}
	if job.UserID != uuid.Nil {
		cc.APIWrapper.JSONResponse(c, "error", "Cron job is associated with a user and cannot be rescheduled", "", nil, nil, http.StatusBadRequest)
		return
	}
	if err := cc.ICronService.RescheduleCronJob(ctx, cronID, payload.NewExpression); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to reschedule cron job", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron job rescheduled successfully", "", nil, nil, http.StatusOK)
}

// @Summary List Cron Jobs
// @Description Lists all cron jobs.
// @Accept json
// @Produce json
// @Success 200 {string} Cron jobs listed successfully
// @Failure 404 {string} Cron jobs not found
// @Router /cronjobs/list [get]
func (cc *CronController) ListCronJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	jobs, err := cc.ICronService.ListCronJobs(ctx)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to list cron jobs", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron jobs listed successfully", "", jobs, nil, http.StatusOK)
}

// @Summary List Active Cron Jobs
// @Description Lists all active cron jobs.
// @Accept json
// @Produce json
// @Success 200 {string} Active cron jobs listed successfully
// @Failure 404 {string} Cron jobs not found
// @Router /cronjobs/active [get]
func (cc *CronController) ListActiveCronJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	jobs, err := cc.ICronService.ListActiveCronJobs(ctx)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to list active cron jobs", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Active cron jobs listed successfully", "", jobs, nil, http.StatusOK)
}

// @Summary Validate Cron Expression
// @Description Validates a cron expression.
// @Accept json
// @Produce json
// @Success 200 {string} Cron expression is valid
// @Failure 404 {string} Invalid cron expression
// @Router /cronjobs/validate [post]
func (cc *CronController) ValidateCronExpression(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	var payload struct {
		Expression string `json:"expression" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid request payload", "", nil, nil, http.StatusBadRequest)
		return
	}
	if err := cc.ICronService.ValidateCronExpression(ctx, payload.Expression); err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron expression", "", nil, nil, http.StatusBadRequest)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Cron expression is valid", "", nil, nil, http.StatusOK)
}

// @Summary Get Job Queue
// @Description Retrieves the current state of the job queue.
// @Accept json
// @Produce json
// @Success 200 {object} []jobqueue.JobQueue
// @Failure 404 {string} Job queue not found
// @Router /cronjobs/queue [get]
func (cc *CronController) GetJobQueue(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	queue, err := cc.ICronService.GetJobQueue(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, queue)
}

// @Summary Reprocess Failed Jobs
// @Description Reprocesses all failed jobs in the queue.
// @Accept json
// @Produce json
// @Success 200 {string} Failed jobs reprocessed successfully
// @Failure 404 {string} Failed jobs not found
// @Router /cronjobs/reprocess [post]
func (cc *CronController) ReprocessFailedJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	err = cc.ICronService.ReprocessFailedJobs(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Failed jobs reprocessed successfully"})
}

// @Summary Get Execution Logs
// @Description Retrieves the execution logs for a specific cron job by its ID.
// @Accept json
// @Produce json
// @Success 200 {string} Execution logs retrieved successfully
// @Failure 404 {string} Cron job not found
// @Router /cronjobs/logs [get]
func (cc *CronController) GetExecutionLogs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to get context", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cronID, ok := ctx.Value("cronID").(uuid.UUID)
	if !ok {
		cc.APIWrapper.JSONResponse(c, "error", "Invalid cron job ID", "", nil, nil, http.StatusBadRequest)
		return
	}
	logs, err := cc.ICronService.GetExecutionLogs(ctx, cronID)
	if err != nil {
		cc.APIWrapper.JSONResponse(c, "error", "Failed to retrieve execution logs", "", nil, nil, http.StatusInternalServerError)
		return
	}
	cc.APIWrapper.JSONResponse(c, "success", "Execution logs retrieved successfully", "", logs, nil, http.StatusOK)
}
