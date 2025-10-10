package types

import "time"

// JobStatusType define os possíveis estados de um job.
type JobStatusType string

const (
	StatusPending     JobStatusType = "pending"
	StatusScheduled   JobStatusType = "scheduled"
	StatusRunning     JobStatusType = "running"
	StatusCompleted   JobStatusType = "completed"
	StatusFailed      JobStatusType = "failed"
	StatusCanceled    JobStatusType = "canceled"
	StatusRescheduled JobStatusType = "rescheduled"
)

type JobStatus struct {
	Code       JobStatusType `json:"code"`
	Message    string        `json:"message,omitempty"`
	UpdatedAt  time.Time     `json:"updated_at"`
	Progress   float64       `json:"progress,omitempty"`
	LastOutput string        `json:"last_output,omitempty"`
}

// NewStatus cria um novo status de job.
func NewStatus(code JobStatusType, msg string) JobStatus {
	return JobStatus{
		Code:      code,
		Message:   msg,
		UpdatedAt: time.Now(),
	}
}

// JobStatusResponse é o payload de resposta retornado por operações do scheduler.
type JobStatusResponse struct {
	JobID     string      `json:"job_id"`
	Status    JobStatus   `json:"status"`
	Scheduled bool        `json:"scheduled"`
	Error     string      `json:"error,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// NewJobStatusResponse cria uma resposta padronizada para um job.
func NewJobStatusResponse(jobID string, status JobStatusType, msg string, err error) JobStatusResponse {
	resp := JobStatusResponse{
		JobID:     jobID,
		Scheduled: status == StatusScheduled,
		Status: JobStatus{
			Code:      status,
			Message:   msg,
			UpdatedAt: time.Now(),
		},
		CreatedAt: time.Now(),
	}
	if err != nil {
		resp.Error = err.Error()
		resp.Status.Code = StatusFailed
		resp.Status.Message = msg
	}
	return resp
}

// SetStatus atualiza o status do job de forma segura e consistente.
func (j *JobImpl) SetStatus(code JobStatusType, msg string) {
	j.Status = NewStatus(code, msg)
}

// GetStatus Atalho opcional pra ler o status atual do job.
func (j *JobImpl) GetStatus() JobStatus {
	return j.Status
}
