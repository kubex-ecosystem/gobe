package gateway

import (
	"time"

	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gatewaytypes "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

type (
	ErrorResponse  = t.ErrorResponse
	MessageResponse = t.MessageResponse
)

// ChatMessage represents a single conversational message from client or assistant.
type ChatMessage = gatewaytypes.Message

// ChatRequest captures the payload expected by the chat SSE endpoint.
type ChatRequest struct {
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Messages    []gatewaytypes.Message `json:"messages"`
	Stream      bool                   `json:"stream"`
	Temperature float32                `json:"temperature"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// ProviderItem holds provider metadata for the gateway /providers response.
type ProviderItem struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Org          string                 `json:"org,omitempty"`
	DefaultModel string                 `json:"default_model,omitempty"`
	Available    bool                   `json:"available"`
	LastError    string                 `json:"last_error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AdviceRequest is a lightweight payload for /advise endpoints.
type AdviceRequest struct {
	Prompt   string                 `json:"prompt"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Provider string                 `json:"provider,omitempty"`
	Model    string                 `json:"model,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Stream   bool                   `json:"stream,omitempty"`
	Tone     string                 `json:"tone,omitempty"`
	Audience string                 `json:"audience,omitempty"`
}

// AdviceResponse wraps the advise payload returned by placeholder handlers.
type AdviceResponse struct {
	Advice   string                 `json:"advice"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ProvidersResponse wraps provider metadata with a timestamp snapshot.
type ProvidersResponse struct {
	Providers []ProviderItem `json:"providers"`
	Timestamp time.Time      `json:"timestamp"`
}

// ScorecardEntry describes the scorecard placeholder output.
type ScorecardEntry struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Score       float64   `json:"score"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags"`
}

// ScorecardResponse describes the list payload served by the scorecard endpoint.
type ScorecardResponse struct {
	Items   []ScorecardEntry `json:"items"`
	Total   int              `json:"total"`
	Version string           `json:"version"`
}

// ScorecardAdviceResponse wraps advisory text for scorecard advice requests.
type ScorecardAdviceResponse struct {
	Advice      string                 `json:"advice"`
	Priority    string                 `json:"priority"`
	Actions     []string               `json:"actions"`
	Metrics     map[string]interface{} `json:"metrics"`
	Version     string                 `json:"version"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// ScorecardMetricsResponse contains aggregated metrics for the scorecard subsystem.
type ScorecardMetricsResponse struct {
	Metrics map[string]interface{} `json:"metrics"`
	Version string                 `json:"version"`
}

// SchedulerStats captures scheduler monitoring information.
type SchedulerStats struct {
	JobsRunning     int           `json:"jobs_running"`
	JobsPending     int           `json:"jobs_pending"`
	JobsCompleted   int           `json:"jobs_completed"`
	LastRun         *time.Time    `json:"last_run,omitempty"`
	LastFailure     *time.Time    `json:"last_failure,omitempty"`
	Uptime          time.Duration `json:"uptime"`
	AverageDuration time.Duration `json:"average_duration"`
}

// SchedulerStatsResponse encapsulates stats snapshot metadata.
type SchedulerStatsResponse struct {
	Stats   SchedulerStats `json:"stats"`
	Version string         `json:"version"`
}

// SchedulerActionResponse resume ações de execução manual do scheduler.
type SchedulerActionResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// LookAtniActionResponse descreve ações assíncronas de extração/arquivo.
type LookAtniActionResponse struct {
	Status    string                 `json:"status"`
	Operation string                 `json:"operation"`
	Payload   map[string]interface{} `json:"payload"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
}

// LookAtniDownloadResponse apresenta o link temporário de download.
type LookAtniDownloadResponse struct {
	DownloadURL string `json:"download_url"`
	ExpiresIn   int    `json:"expires_in"`
	Note        string `json:"note"`
}

// LookAtniProjectsResponse lista projetos configurados.
type LookAtniProjectsResponse struct {
	Projects []map[string]interface{} `json:"projects"`
	Version  string                   `json:"version"`
}

// WebhookAckResponse confirma recebimento de webhooks na camada gateway.
type WebhookAckResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Payload   map[string]interface{} `json:"payload"`
}

// WebhookHealthResponse descreve a resposta de health check do módulo de webhooks.
type WebhookHealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Stats     map[string]interface{} `json:"stats,omitempty"`
}

// WebhookEvent represents a webhook event for API responses
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Source    string                 `json:"source"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	Headers   map[string]string      `json:"headers"`
	Timestamp time.Time              `json:"timestamp"`
	Processed bool                   `json:"processed"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
}

// WebhookEventsResponse lists webhook events with pagination
type WebhookEventsResponse struct {
	Events []interface{} `json:"events"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// WebhookRetryResponse confirms webhook retry operation
type WebhookRetryResponse struct {
	Status       string    `json:"status"`
	RetriedCount int       `json:"retried_count"`
	Timestamp    time.Time `json:"timestamp"`
	Message      string    `json:"message"`
}
