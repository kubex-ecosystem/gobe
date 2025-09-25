package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	gateway "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

type anthropicProvider struct {
	name         string
	baseURL      string
	apiKey       string
	defaultModel string
	client       *http.Client
}

func newAnthropicProvider(cfg Config) (gateway.Provider, error) {
	key := staticAPIKey(cfg)
	if key == "" {
		return nil, errors.New("anthropic provider requires an api key")
	}
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	model := strings.TrimSpace(cfg.DefaultModel)
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	return &anthropicProvider{
		name:         cfg.Name,
		baseURL:      baseURL,
		apiKey:       key,
		defaultModel: model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (a *anthropicProvider) Name() string { return a.name }

func (a *anthropicProvider) Available() error {
	if strings.TrimSpace(a.apiKey) == "" {
		return errors.New("anthropic api key not configured")
	}
	return nil
}

func (a *anthropicProvider) Notify(ctx context.Context, event gateway.NotificationEvent) error {
	return nil
}

func (a *anthropicProvider) Chat(ctx context.Context, req gateway.ChatRequest) (<-chan gateway.ChatChunk, error) {
	key := a.resolveKey(req)
	if key == "" {
		return nil, errors.New("missing api key for anthropic request")
	}

	if len(req.Messages) == 0 {
		return nil, errors.New("anthropic chat requires at least one message")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = a.defaultModel
	}
	if model == "" {
		return nil, errors.New("anthropic model not specified")
	}

	// Convert messages to Anthropic format
	messages := toAnthropicMessages(req.Messages)

	body := map[string]interface{}{
		"model":      model,
		"messages":   messages,
		"max_tokens": 8192,
		"stream":     true,
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal anthropic request: %w", err)
	}

	url := strings.TrimRight(a.baseURL, "/") + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create anthropic request: %w", err)
	}

	httpReq.Header.Set("x-api-key", key)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	responseChan := make(chan gateway.ChatChunk, 32)

	go func() {
		defer close(responseChan)

		start := time.Now()
		resp, err := a.client.Do(httpReq)
		if err != nil {
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("anthropic request failed: %v", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("anthropic api error %d: %s", resp.StatusCode, string(body))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var inputTokens, outputTokens int

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}

			var event anthropicStreamEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				continue
			}

			switch event.Type {
			case "message_start":
				if event.Message != nil && event.Message.Usage != nil {
					inputTokens = event.Message.Usage.InputTokens
				}
			case "content_block_delta":
				if event.Delta != nil && event.Delta.Text != "" {
					responseChan <- gateway.ChatChunk{Content: event.Delta.Text}
				}
			case "message_delta":
				if event.Delta != nil && event.Delta.Usage != nil {
					outputTokens = event.Delta.Usage.OutputTokens
				}
			case "error":
				if event.Error != nil {
					responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("anthropic stream error: %s", event.Error.Message)}
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("anthropic stream error: %v", err)}
			return
		}

		totalTokens := inputTokens + outputTokens
		latency := time.Since(start).Milliseconds()
		responseChan <- gateway.ChatChunk{
			Done: true,
			Usage: &gateway.Usage{
				PromptTokens:     inputTokens,
				CompletionTokens: outputTokens,
				TotalTokens:      totalTokens,
				LatencyMS:        latency,
				CostUSD:          estimateAnthropicCost(model, inputTokens, outputTokens),
				Provider:         a.name,
				Model:            model,
			},
		}
	}()

	return responseChan, nil
}

func (a *anthropicProvider) resolveKey(req gateway.ChatRequest) string {
	if key := externalAPIKey(req); key != "" {
		return key
	}
	return a.apiKey
}

func toAnthropicMessages(messages []gateway.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))
	for _, msg := range messages {
		// Anthropic uses "user" and "assistant" roles
		role := msg.Role
		if role == "system" {
			// Convert system messages to user messages with prefix
			role = "user"
			msg.Content = "System: " + msg.Content
		}

		result = append(result, map[string]interface{}{
			"role": role,
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": msg.Content,
				},
			},
		})
	}
	return result
}

type anthropicStreamEvent struct {
	Type    string `json:"type"`
	Message *struct {
		Usage *struct {
			InputTokens int `json:"input_tokens"`
		} `json:"usage"`
	} `json:"message,omitempty"`
	Delta *struct {
		Type  string `json:"type,omitempty"`
		Text  string `json:"text,omitempty"`
		Usage *struct {
			OutputTokens int `json:"output_tokens"`
		} `json:"usage,omitempty"`
	} `json:"delta,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func estimateAnthropicCost(model string, inputTokens, outputTokens int) float64 {
	var inputRate, outputRate float64

	switch {
	case strings.Contains(model, "claude-3-5-sonnet"):
		inputRate = 3.0 / 1_000_000    // $3.00 per million input tokens
		outputRate = 15.0 / 1_000_000  // $15.00 per million output tokens
	case strings.Contains(model, "claude-3-opus"):
		inputRate = 15.0 / 1_000_000   // $15.00 per million input tokens
		outputRate = 75.0 / 1_000_000  // $75.00 per million output tokens
	case strings.Contains(model, "claude-3-sonnet"):
		inputRate = 3.0 / 1_000_000    // $3.00 per million input tokens
		outputRate = 15.0 / 1_000_000  // $15.00 per million output tokens
	case strings.Contains(model, "claude-3-haiku"):
		inputRate = 0.25 / 1_000_000   // $0.25 per million input tokens
		outputRate = 1.25 / 1_000_000  // $1.25 per million output tokens
	default:
		// Default to Claude 3.5 Sonnet pricing
		inputRate = 3.0 / 1_000_000
		outputRate = 15.0 / 1_000_000
	}

	return float64(inputTokens)*inputRate + float64(outputTokens)*outputRate
}