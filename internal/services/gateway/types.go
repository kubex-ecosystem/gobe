// Package gateway defines interfaces and types for interacting with various AI model providers.
package gateway

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Provider    string                 `json:"provider"`
	Model       string                 `json:"model"`
	Messages    []Message              `json:"messages"`
	Temperature float32                `json:"temperature"`
	Stream      bool                   `json:"stream"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
	Headers     map[string]string      `json:"-"`
}

type ToolCall struct {
	Name string      `json:"name"`
	Args interface{} `json:"args"`
}

type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`
	LatencyMS        int64   `json:"latency_ms"`
	CostUSD          float64 `json:"cost_usd"`
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
}

type ChatChunk struct {
	Content  string    `json:"content,omitempty"`
	Done     bool      `json:"done"`
	Usage    *Usage    `json:"usage,omitempty"`
	Error    string    `json:"error,omitempty"`
	ToolCall *ToolCall `json:"tool_call,omitempty"`
}

type NotificationEvent struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)
	Available() error
	Notify(ctx context.Context, event NotificationEvent) error
}

type ProviderConfig struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	BaseURL      string                 `json:"base_url"`
	DefaultModel string                 `json:"default_model"`
	APIKey       string                 `json:"api_key"`
	KeyEnv       string                 `json:"key_env"`
	Org          string                 `json:"org"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type ProviderSummary struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Org          string                 `json:"org"`
	DefaultModel string                 `json:"default_model"`
	Available    bool                   `json:"available"`
	LastError    string                 `json:"last_error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type ProviderEntry struct {
	Config   ProviderConfig
	Provider Provider
}
