// Package cron provides the controller for managing cron jobs in the application.
package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type CronController struct {
	ICronService svc.CronJobService
	APIWrapper   *types.APIWrapper[svc.CronJobModel]
}

func respondCronError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Status: "error", Message: message})
}

func cronIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if value := ctx.Value(types.CtxKey("cronID")); value != nil {
		if id, ok := value.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if value := ctx.Value(types.CtxKey("userID")); value != nil {
		if id, ok := value.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

func toCronJobSlice(models []*svc.CronJobModel) []svc.CronJobModel {
	if len(models) == 0 {
		return []svc.CronJobModel{}
	}
	result := make([]svc.CronJobModel, 0, len(models))
	for _, m := range models {
		if m != nil {
			result = append(result, *m)
		}
	}
	return result
}

func marshalToMapSlice(value any) ([]map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var response []map[string]any
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return response, nil
}

func NewCronJobController(bridge *svc.Bridge) *CronController {
	return &CronController{
		ICronService: bridge.NewCronJobService(),
		APIWrapper:   types.NewAPIWrapper[svc.CronJobModel](),
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

// GetAllCronJobs mantém compatibilidade com rotas legadas de listagem.
//
// @Summary     Listar cron jobs (legacy)
// @Description Retorna todos os cron jobs cadastrados. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} CronJobListResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /cronjobs [get]
func (cc *CronController) GetAllCronJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	jobs, err := cc.ICronService.ListCronJobs(ctx)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to fetch cron jobs")
		return
	}
	c.JSON(http.StatusOK, CronJobListResponse{Jobs: toCronJobSlice(jobs)})
}

// GetCronJobByID retorna um cron job específico.
//
// @Summary     Obter cron job
// @Description Recupera um cron job específico pelo ID informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cron job"
// @Success     200 {object} CronJobResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id} [get]
func (cc *CronController) GetCronJobByID(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		respondCronError(c, http.StatusNotFound, "cron job not found")
		return
	}
	c.JSON(http.StatusOK, CronJobResponse{Job: *job})
}

// CreateCronJob registra um novo cron job.
//
// @Summary     Criar cron job
// @Description Cria um novo cron job com os dados informados. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body CronJobRequest true "Dados do cron job"
// @Success     201 {object} CronJobResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs [post]
func (cc *CronController) CreateCronJob(c *gin.Context) {
	var req CronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondCronError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		respondCronError(c, http.StatusBadRequest, "name is required")
		return
	}
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	userID, ok := userIDFromContext(ctx)
	if !ok || userID == uuid.Nil {
		respondCronError(c, http.StatusUnauthorized, "user id is required")
		return
	}
	identifier, err := uuid.NewRandom()
	if err != nil {
		gl.Log("error", fmt.Sprintf("failed to generate uuid: %s", err))
		respondCronError(c, http.StatusInternalServerError, "failed to create cron job")
		return
	}
	job := &svc.CronJobModel{
		ID:             identifier,
		Name:           req.Name,
		CronExpression: req.Expression,
		Description:    req.Description,
		IsActive:       req.Enabled,
		LastRunStatus:  "pending",
		UserID:         userID,
		CreatedBy:      userID,
		UpdatedBy:      userID,
	}
	createdJob, err := cc.ICronService.CreateCronJob(ctx, job)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to create cron job")
		return
	}
	c.JSON(http.StatusCreated, CronJobResponse{Job: *createdJob})
}

// UpdateCronJob atualiza um cron job existente.
//
// @Summary     Atualizar cron job
// @Description Atualiza os dados de um cron job identificado pelo ID. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string        true "ID do cron job"
// @Param       payload body CronJobRequest true "Dados do cron job"
// @Success     200 {object} CronJobResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id} [put]
func (cc *CronController) UpdateCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	var req CronJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondCronError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	existing, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil || existing == nil {
		respondCronError(c, http.StatusNotFound, "cron job not found")
		return
	}
	existing.Name = req.Name
	if strings.TrimSpace(req.Expression) != "" {
		existing.CronExpression = req.Expression
	}
	existing.Description = req.Description
	existing.IsActive = req.Enabled
	if userID, ok := userIDFromContext(ctx); ok {
		existing.UpdatedBy = userID
	}
	updatedJob, err := cc.ICronService.UpdateCronJob(ctx, existing)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to update cron job")
		return
	}
	c.JSON(http.StatusOK, CronJobResponse{Job: *updatedJob})
}

// DeleteCronJob remove um cron job existente.
//
// @Summary     Remover cron job
// @Description Exclui um cron job pelo ID informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cron job"
// @Success     200 {object} CronActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id} [delete]
func (cc *CronController) DeleteCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		respondCronError(c, http.StatusNotFound, "cron job not found")
		return
	}
	if job.UserID != uuid.Nil {
		respondCronError(c, http.StatusBadRequest, "cron job associated to a user cannot be deleted")
		return
	}
	// Check if the cron job is currently running
	if job.LastRunStatus == "running" {
		respondCronError(c, http.StatusBadRequest, "cron job currently running")
		return
	}
	// Check if the cron job has any pending executions
	if job.LastRunStatus == "pending" {
		respondCronError(c, http.StatusBadRequest, "cron job has pending executions")
		return
	}

	if err := cc.ICronService.DeleteCronJob(ctx, cronID); err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to delete cron job")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Cron job deleted successfully"})
}

// EnableCronJob habilita um cron job existente.
//
// @Summary     Habilitar cron job
// @Description Ativa o cron job informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cron job"
// @Success     200 {object} CronActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id}/enable [post]
func (cc *CronController) EnableCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	if err := cc.ICronService.EnableCronJob(ctx, cronID); err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to enable cron job")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Cron job enabled successfully"})
}

// DisableCronJob desabilita um cron job.
//
// @Summary     Desabilitar cron job
// @Description Desativa o cron job informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cron job"
// @Success     200 {object} CronActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id}/disable [post]
func (cc *CronController) DisableCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	if err := cc.ICronService.DisableCronJob(ctx, cronID); err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to disable cron job")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Cron job disabled successfully"})
}

// ExecuteCronJobManually executa o cron job imediatamente.
//
// @Summary     Executar cron job manualmente
// @Description Dispara manualmente a execução do cron job informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cron job"
// @Success     200 {object} CronActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id}/execute [post]
func (cc *CronController) ExecuteCronJobManually(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	if err := cc.ICronService.ExecuteCronJobManually(ctx, cronID); err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to execute cron job")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Cron job executed successfully"})
}

// ExecuteCronJobManuallyByID mantém compatibilidade com rotas antigas.
//
// @Summary     Executar cron job manualmente (legacy)
// @Description Dispara manualmente a execução do cron job informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id     path string true "ID do cron job"
// @Param       job_id path string false "ID adicional do job"
// @Success     200 {object} CronActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /cronjobs/{id}/execute/{job_id} [post]
func (cc *CronController) ExecuteCronJobManuallyByID(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	// Check if the cron job is currently running
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		respondCronError(c, http.StatusNotFound, "cron job not found")
		return
	}
	if job.LastRunStatus == "running" {
		respondCronError(c, http.StatusBadRequest, "cron job currently running")
		return
	}
	if err := cc.ICronService.ExecuteCronJobManually(ctx, cronID); err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to execute cron job")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Cron job executed successfully"})
}

// RescheduleCronJob atualiza a expressão de agendamento.
//
// @Summary     Reagendar cron job
// @Description Atualiza a expressão do cron job informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string           true "ID do cron job"
// @Param       payload body RescheduleRequest true "Nova expressão"
// @Success     200 {object} CronActionResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id}/reschedule [put]
func (cc *CronController) RescheduleCronJob(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	var payload RescheduleRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondCronError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if strings.TrimSpace(payload.NewExpression) == "" {
		respondCronError(c, http.StatusBadRequest, "new_expression is required")
		return
	}
	job, err := cc.ICronService.GetCronJobByID(ctx, cronID)
	if err != nil {
		respondCronError(c, http.StatusNotFound, "cron job not found")
		return
	}
	if job.UserID != uuid.Nil {
		respondCronError(c, http.StatusBadRequest, "cron job is associated with a user and cannot be rescheduled")
		return
	}
	if err := cc.ICronService.RescheduleCronJob(ctx, cronID, payload.NewExpression); err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to reschedule cron job")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Cron job rescheduled successfully"})
}

// ListCronJobs lista cron jobs (rota principal).
//
// @Summary     Listar cron jobs
// @Description Retorna todos os cron jobs cadastrados. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} CronJobListResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs [get]
func (cc *CronController) ListCronJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	jobs, err := cc.ICronService.ListCronJobs(ctx)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to list cron jobs")
		return
	}
	c.JSON(http.StatusOK, CronJobListResponse{Jobs: toCronJobSlice(jobs)})
}

// ListActiveCronJobs retorna cron jobs ativos.
//
// @Summary     Listar cron jobs ativos
// @Description Retorna apenas os cron jobs marcados como ativos. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} CronJobListResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/active [get]
func (cc *CronController) ListActiveCronJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	jobs, err := cc.ICronService.ListActiveCronJobs(ctx)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to list active cron jobs")
		return
	}
	c.JSON(http.StatusOK, CronJobListResponse{Jobs: toCronJobSlice(jobs)})
}

// ValidateCronExpression verifica se a expressão é válida.
//
// @Summary     Validar expressão cron
// @Description Valida a expressão cron fornecida. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body CronValidateRequest true "Expressão"
// @Success     200 {object} CronValidateResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/validate [post]
func (cc *CronController) ValidateCronExpression(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	var payload CronValidateRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondCronError(c, http.StatusBadRequest, "invalid request payload")
		return
	}
	if err := cc.ICronService.ValidateCronExpression(ctx, payload.Expression); err != nil {
		c.JSON(http.StatusBadRequest, CronValidateResponse{Valid: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, CronValidateResponse{Valid: true})
}

// GetJobQueue lista a fila de jobs pendentes.
//
// @Summary     Listar fila de jobs
// @Description Recupera o estado atual da fila de jobs agendados. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} CronJobQueueResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/queue [get]
func (cc *CronController) GetJobQueue(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	queue, err := cc.ICronService.GetJobQueue(ctx)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to retrieve job queue")
		return
	}
	items, err := marshalToMapSlice(queue)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to serialize job queue")
		return
	}
	c.JSON(http.StatusOK, CronJobQueueResponse{Queue: items})
}

// ReprocessFailedJobs reprocesa jobs com falha.
//
// @Summary     Reprocessar jobs com falha
// @Description Reenvia para execução os jobs que falharam anteriormente. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} CronActionResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/reprocess [post]
func (cc *CronController) ReprocessFailedJobs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	err = cc.ICronService.ReprocessFailedJobs(ctx)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to reprocess failed jobs")
		return
	}
	c.JSON(http.StatusOK, CronActionResponse{Message: "Failed jobs reprocessed successfully"})
}

// GetExecutionLogs lista os logs de execução de um cron job.
//
// @Summary     Listar logs de execução
// @Description Recupera os logs associados ao cron job informado. [Em desenvolvimento]
// @Tags        cron beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cron job"
// @Success     200 {object} CronExecutionLogsResponse
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/cronjobs/{id}/logs [get]
func (cc *CronController) GetExecutionLogs(c *gin.Context) {
	ctx, err := cc.APIWrapper.GetContext(c)
	if err != nil {
		respondCronError(c, http.StatusUnauthorized, "failed to resolve context")
		return
	}
	cronID, ok := cronIDFromContext(ctx)
	if !ok {
		respondCronError(c, http.StatusBadRequest, "invalid cron job id")
		return
	}
	logs, err := cc.ICronService.GetExecutionLogs(ctx, cronID)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to retrieve execution logs")
		return
	}
	items, err := marshalToMapSlice(logs)
	if err != nil {
		respondCronError(c, http.StatusInternalServerError, "failed to serialize execution logs")
		return
	}
	c.JSON(http.StatusOK, CronExecutionLogsResponse{Logs: items})
}
