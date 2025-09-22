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

type openAIProvider struct {
	name         string
	baseURL      string
	apiKey       string
	defaultModel string
	client       *http.Client
}

func newOpenAIProvider(cfg Config) (gateway.Provider, error) {
	key := staticAPIKey(cfg)
	if key == "" {
		return nil, errors.New("openai provider requires an api key")
	}
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	return &openAIProvider{
		name:         cfg.Name,
		baseURL:      baseURL,
		apiKey:       key,
		defaultModel: cfg.DefaultModel,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}, nil
}

func (o *openAIProvider) Name() string { return o.name }

func (o *openAIProvider) Available() error {
	if strings.TrimSpace(o.apiKey) == "" {
		return errors.New("openai api key not configured")
	}
	return nil
}

func (o *openAIProvider) Notify(ctx context.Context, event gateway.NotificationEvent) error {
	return nil
}

func (o *openAIProvider) Chat(ctx context.Context, req gateway.ChatRequest) (<-chan gateway.ChatChunk, error) {
	key := o.resolveKey(req)
	if key == "" {
		return nil, errors.New("missing api key for openai request")
	}

	if len(req.Messages) == 0 {
		return nil, errors.New("openai chat requires at least one message")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = o.defaultModel
	}
	if model == "" {
		return nil, errors.New("openai model not specified")
	}

	body := map[string]interface{}{
		"model":       model,
		"messages":    toOpenAIMessages(req.Messages),
		"temperature": req.Temperature,
		"stream":      true,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal openai request: %w", err)
	}

	url := strings.TrimRight(o.baseURL, "/") + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create openai request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	responseChan := make(chan gateway.ChatChunk, 32)

	go func() {
		defer close(responseChan)

		start := time.Now()
		resp, err := o.client.Do(httpReq)
		if err != nil {
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("openai request failed: %v", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("openai api error %d: %s", resp.StatusCode, string(body))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		totalTokens := 0
		promptTokens := 0
		completionTokens := 0

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}

			var chunk openAIStreamChunk
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta != "" {
					responseChan <- gateway.ChatChunk{Content: delta}
				}
			}

			if chunk.Usage != nil {
				promptTokens = chunk.Usage.PromptTokens
				completionTokens = chunk.Usage.CompletionTokens
				totalTokens = chunk.Usage.TotalTokens
			}
		}

		if err := scanner.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("openai stream error: %v", err)}
			return
		}

		latency := time.Since(start).Milliseconds()
		responseChan <- gateway.ChatChunk{
			Done: true,
			Usage: &gateway.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      totalTokens,
				LatencyMS:        latency,
				CostUSD:          estimateOpenAICost(model, totalTokens),
				Provider:         o.name,
				Model:            model,
			},
		}
	}()

	return responseChan, nil
}

func (o *openAIProvider) resolveKey(req gateway.ChatRequest) string {
	if key := externalAPIKey(req); key != "" {
		return key
	}
	return o.apiKey
}

func toOpenAIMessages(messages []gateway.Message) []map[string]string {
	result := make([]map[string]string, 0, len(messages))
	for _, msg := range messages {
		result = append(result, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	return result
}

type openAIStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

func estimateOpenAICost(model string, tokens int) float64 {
	costPerToken := 0.000002
	switch {
	case strings.Contains(model, "gpt-4"):
		costPerToken = 0.00003
	case strings.Contains(model, "gpt-3.5"):
		costPerToken = 0.000002
	}
	return float64(tokens) * costPerToken
}

