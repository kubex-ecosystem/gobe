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
	"sync"
	"time"

	gateway "github.com/kubex-ecosystem/gobe/internal/services/gateway"
)

type groqProvider struct {
	name         string
	apiKey       string
	defaultModel string
	baseURL      string
	client       *http.Client
	mu           sync.Mutex
}

func newGroqProvider(cfg Config) (gateway.Provider, error) {
	key := staticAPIKey(cfg)
	if key == "" {
		return nil, errors.New("groq provider requires an api key")
	}

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "https://api.groq.com"
	}

	model := strings.TrimSpace(cfg.DefaultModel)
	if model == "" {
		model = "llama-3.1-70b-versatile"
	}

	return &groqProvider{
		name:         cfg.Name,
		apiKey:       key,
		defaultModel: model,
		baseURL:      baseURL,
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}, nil
}

func (p *groqProvider) Name() string { return p.name }

func (p *groqProvider) Available() error {
	if strings.TrimSpace(p.apiKey) == "" {
		return errors.New("groq api key not configured")
	}
	return nil
}

func (p *groqProvider) Notify(ctx context.Context, event gateway.NotificationEvent) error {
	return nil
}

func (p *groqProvider) resolveKey(req gateway.ChatRequest) string {
	if key := externalAPIKey(req); key != "" {
		return key
	}
	return p.apiKey
}

func (p *groqProvider) Chat(ctx context.Context, req gateway.ChatRequest) (<-chan gateway.ChatChunk, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.resolveKey(req)
	if key == "" {
		return nil, errors.New("missing api key for groq request")
	}

	if len(req.Messages) == 0 {
		return nil, errors.New("groq chat requires at least one message")
	}

	messages := make([]groqMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, groqMessage{Role: msg.Role, Content: msg.Content})
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.defaultModel
	}

	groqReq := groqRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	if req.Temperature > 0 {
		groqReq.Temperature = &req.Temperature
	}

	maxTokens := 8192
	groqReq.MaxTokens = &maxTokens

	body, err := json.Marshal(groqReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal groq request: %w", err)
	}

	url := strings.TrimRight(p.baseURL, "/") + "/openai/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create groq request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+key)
	httpReq.Header.Set("Accept", "text/event-stream")

	responseChan := make(chan gateway.ChatChunk, 64)

	go func() {
		defer close(responseChan)

		start := time.Now()
		resp, err := p.client.Do(httpReq)
		if err != nil {
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("groq request failed: %v", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("groq api error %d: %s", resp.StatusCode, string(body))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		promptTokens := 0
		completionTokens := 0
		totalTokens := 0

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}

			var chunk groqStreamChunk
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				choice := chunk.Choices[0]
				if choice.Delta.Content != "" {
					select {
					case responseChan <- gateway.ChatChunk{Content: choice.Delta.Content}:
					case <-ctx.Done():
						return
					}
				}
				if choice.FinishReason != nil && *choice.FinishReason != "" && chunk.Usage != nil {
					promptTokens = chunk.Usage.PromptTokens
					completionTokens = chunk.Usage.CompletionTokens
					totalTokens = chunk.Usage.TotalTokens
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("groq stream error: %v", err)}
			return
		}

		if totalTokens == 0 {
			totalTokens = promptTokens + completionTokens
		}

		latency := time.Since(start).Milliseconds()
		responseChan <- gateway.ChatChunk{
			Done: true,
			Usage: &gateway.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: completionTokens,
				TotalTokens:      totalTokens,
				LatencyMS:        latency,
				CostUSD:          estimateGroqCost(model, promptTokens, completionTokens),
				Provider:         p.name,
				Model:            model,
			},
		}
	}()

	return responseChan, nil
}

func estimateGroqCost(model string, inputTokens, outputTokens int) float64 {
	var inputRate, outputRate float64

	switch {
	case strings.Contains(model, "llama-3.1-70b"):
		inputRate = 0.59 / 1_000_000
		outputRate = 0.79 / 1_000_000
	case strings.Contains(model, "llama-3.1-8b"):
		inputRate = 0.05 / 1_000_000
		outputRate = 0.08 / 1_000_000
	case strings.Contains(model, "mixtral-8x7b"):
		inputRate = 0.24 / 1_000_000
		outputRate = 0.24 / 1_000_000
	case strings.Contains(model, "gemma"):
		inputRate = 0.10 / 1_000_000
		outputRate = 0.10 / 1_000_000
	default:
		inputRate = 0.59 / 1_000_000
		outputRate = 0.79 / 1_000_000
	}

	return float64(inputTokens)*inputRate + float64(outputTokens)*outputRate
}

type groqRequest struct {
	Model       string        `json:"model"`
	Messages    []groqMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature *float32      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	TopP        *float32      `json:"top_p,omitempty"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

