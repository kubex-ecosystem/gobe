package cron

import (
	svc "github.com/kubex-ecosystem/gdbase/factory/models"

	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

type (
	// ErrorResponse padroniza respostas de erro no módulo de cron.
	ErrorResponse = t.ErrorResponse
)

// CronJobRequest representa o payload básico de criação/atualização de cron job.
type CronJobRequest struct {
	Name        string `json:"name"`
	Expression  string `json:"expression"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
}

// CronJobResponse descreve o retorno dos endpoints principais.
type CronJobResponse struct {
	Job svc.CronJobModel `json:"job"`
}

// CronJobListResponse encapsula listagens paginadas ou simples.
type CronJobListResponse struct {
	Jobs []svc.CronJobModel `json:"jobs"`
}

// CronActionResponse indica o resultado de ações sobre cron jobs.
type CronActionResponse struct {
	Message string `json:"message"`
}

// CronValidateRequest representa a validação de expressão cron.
type CronValidateRequest struct {
	Expression string `json:"expression"`
}

// CronValidateResponse indica se a expressão é válida.
type CronValidateResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// RescheduleRequest altera a expressão de um cron job existente.
type RescheduleRequest struct {
	NewExpression string `json:"new_expression"`
}

// CronJobQueueResponse representa a fila de jobs aguardando execução.
type CronJobQueueResponse struct {
	Queue []map[string]any `json:"queue"`
}

// CronExecutionLogsResponse agrega os logs de execução associados a um cron job.
type CronExecutionLogsResponse struct {
	Logs []map[string]any `json:"logs"`
}
