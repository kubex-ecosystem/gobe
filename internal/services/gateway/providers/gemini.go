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

type geminiProvider struct {
	name         string
	baseURL      string
	apiKey       string
	defaultModel string
	client       *http.Client
}

func newGeminiProvider(cfg Config) (gateway.Provider, error) {
	key := staticAPIKey(cfg)
	if key == "" {
		return nil, errors.New("gemini provider requires an api key")
	}
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}

	model := strings.TrimSpace(cfg.DefaultModel)
	if model == "" {
		model = "gemini-1.5-flash"
	}

	return &geminiProvider{
		name:         cfg.Name,
		baseURL:      baseURL,
		apiKey:       key,
		defaultModel: model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (g *geminiProvider) Name() string { return g.name }

func (g *geminiProvider) Available() error {
	if strings.TrimSpace(g.apiKey) == "" {
		return errors.New("gemini api key not configured")
	}
	return nil
}

func (g *geminiProvider) Notify(ctx context.Context, event gateway.NotificationEvent) error {
	return nil
}

func (g *geminiProvider) Chat(ctx context.Context, req gateway.ChatRequest) (<-chan gateway.ChatChunk, error) {
	key := g.resolveKey(req)
	if key == "" {
		return nil, errors.New("missing api key for gemini request")
	}

	if len(req.Messages) == 0 {
		return nil, errors.New("gemini chat requires at least one message")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = g.defaultModel
	}
	if model == "" {
		return nil, errors.New("gemini model not specified")
	}

	// Convert messages to Gemini format
	contents := toGeminiContents(req.Messages)

	body := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": 8192,
		},
	}

	// Add temperature if specified
	if req.Temperature > 0 {
		body["generationConfig"].(map[string]interface{})["temperature"] = req.Temperature
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?key=%s",
		strings.TrimRight(g.baseURL, "/"), model, key)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	responseChan := make(chan gateway.ChatChunk, 32)

	go func() {
		defer close(responseChan)

		start := time.Now()
		resp, err := g.client.Do(httpReq)
		if err != nil {
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("gemini request failed: %v", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("gemini api error %d: %s", resp.StatusCode, string(body))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		var totalTokens int
		var promptTokens int
		var candidateTokens int

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				break
			}

			var chunk geminiStreamChunk
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				continue
			}

			// Extract text content from candidates
			for _, candidate := range chunk.Candidates {
				if candidate.Content != nil {
					for _, part := range candidate.Content.Parts {
						if part.Text != "" {
							responseChan <- gateway.ChatChunk{Content: part.Text}
						}
					}
				}
			}

			// Extract usage metadata
			if chunk.UsageMetadata != nil {
				promptTokens = chunk.UsageMetadata.PromptTokenCount
				candidateTokens = chunk.UsageMetadata.CandidatesTokenCount
				totalTokens = chunk.UsageMetadata.TotalTokenCount
			}
		}

		if err := scanner.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			responseChan <- gateway.ChatChunk{Done: true, Error: fmt.Sprintf("gemini stream error: %v", err)}
			return
		}

		if totalTokens == 0 {
			totalTokens = promptTokens + candidateTokens
		}

		latency := time.Since(start).Milliseconds()
		responseChan <- gateway.ChatChunk{
			Done: true,
			Usage: &gateway.Usage{
				PromptTokens:     promptTokens,
				CompletionTokens: candidateTokens,
				TotalTokens:      totalTokens,
				LatencyMS:        latency,
				CostUSD:          estimateGeminiCost(model, promptTokens, candidateTokens),
				Provider:         g.name,
				Model:            model,
			},
		}
	}()

	return responseChan, nil
}

func (g *geminiProvider) resolveKey(req gateway.ChatRequest) string {
	if key := externalAPIKey(req); key != "" {
		return key
	}
	return g.apiKey
}

func toGeminiContents(messages []gateway.Message) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(messages))

	for _, msg := range messages {
		// Gemini uses "user" and "model" roles
		role := msg.Role
		if role == "assistant" {
			role = "model"
		} else if role == "system" {
			// Convert system messages to user messages with prefix
			role = "user"
			msg.Content = "System: " + msg.Content
		}

		result = append(result, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{
					"text": msg.Content,
				},
			},
		})
	}
	return result
}

type geminiStreamChunk struct {
	Candidates []struct {
		Content *struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason,omitempty"`
		Index        int    `json:"index"`
	} `json:"candidates"`
	UsageMetadata *struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata,omitempty"`
}

func estimateGeminiCost(model string, inputTokens, outputTokens int) float64 {
	var inputRate, outputRate float64

	switch {
	case strings.Contains(model, "gemini-1.5-pro"):
		inputRate = 3.5 / 1_000_000   // $3.50 per million input tokens (up to 128k)
		outputRate = 10.5 / 1_000_000 // $10.50 per million output tokens
	case strings.Contains(model, "gemini-1.5-flash"):
		inputRate = 0.075 / 1_000_000  // $0.075 per million input tokens (up to 128k)
		outputRate = 0.3 / 1_000_000   // $0.30 per million output tokens
	case strings.Contains(model, "gemini-1.0-pro"):
		inputRate = 0.5 / 1_000_000   // $0.50 per million input tokens
		outputRate = 1.5 / 1_000_000  // $1.50 per million output tokens
	default:
		// Default to Gemini 1.5 Flash pricing (most common)
		inputRate = 0.075 / 1_000_000
		outputRate = 0.3 / 1_000_000
	}

	return float64(inputTokens)*inputRate + float64(outputTokens)*outputRate
}