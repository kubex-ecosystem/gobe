package types

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// JobStatusMap is a map of job statuses to their string representations.
var JobStatusMap = map[JobStatus]string{
	JobStatusPending:   "pending",
	JobStatusRunning:   "running",
	JobStatusCompleted: "completed",
	JobStatusFailed:    "failed",
}

// JobStatusList is a list of all possible job statuses.
var JobStatusList = []JobStatus{
	JobStatusPending,
	JobStatusRunning,
	JobStatusCompleted,
	JobStatusFailed,
}

type JobStatusType struct {
	JobID     string    `json:"job_id"`
	JobStatus JobStatus `json:"job_status"`
}

type JobStatusResponse struct {
	JobID               string    `json:"job_id"`
	JobStatus           JobStatus `json:"job_status"`
	JobMessage          string    `json:"job_message"`
	JobTime             string    `json:"job_time"`
	JobDuration         string    `json:"job_duration"`
	JobRetries          int       `json:"job_retries"`
	JobMaxRetries       int       `json:"job_max_retries"`
	JobExecTimeout      int       `json:"job_exec_timeout"`
	JobMaxExecutionTime int       `json:"job_max_execution_time"`
	JobCreatedAt        string    `json:"job_created_at"`
	JobUpdatedAt        string    `json:"job_updated_at"`
	JobLastExecutedAt   string    `json:"job_last_executed_at"`
	JobLastExecutedBy   string    `json:"job_last_executed_by"`
	JobCreatedBy        string    `json:"job_created_by"`
	JobUpdatedBy        string    `json:"job_updated_by"`
	JobUserID           string    `json:"job_user_id"`
	JobCronType         string    `json:"job_cron_type"`
	JobCronExpression   string    `json:"job_cron_expression"`
	JobCommand          string    `json:"job_command"`
	JobMethod           string    `json:"job_method"`
	JobAPIEndpoint      string    `json:"job_api_endpoint"`
	JobLastRunStatus    string    `json:"job_last_run_status"`
	JobLastRunMessage   string    `json:"job_last_run_message"`
}
