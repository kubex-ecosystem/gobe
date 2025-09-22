package gateway

import (
	"time"

	gatewaytypes "github.com/kubex-ecosystem/gobe/internal/services/gateway"
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
	Prompt    string                 `json:"prompt"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Provider  string                 `json:"provider,omitempty"`
	Model     string                 `json:"model,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Stream    bool                   `json:"stream,omitempty"`
	Tone      string                 `json:"tone,omitempty"`
	Audience  string                 `json:"audience,omitempty"`
}

// AdviceResponse wraps the advise payload returned by placeholder handlers.
type AdviceResponse struct {
	Advice   string                 `json:"advice"`
	Metadata map[string]interface{} `json:"metadata"`
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

