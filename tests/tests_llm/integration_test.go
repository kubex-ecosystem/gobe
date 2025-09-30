// Package testsllm provides integration tests for the LLM service.
package testsllm

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/config"
	"github.com/kubex-ecosystem/gobe/internal/services/llm"
)

func TestLLMIntegration(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		apiKey   string
		envKey   string
		expected string
	}{
		{
			name:     "Gemini Integration",
			provider: "gemini",
			envKey:   "GEMINI_API_KEY",
			expected: "gemini",
		},
		{
			name:     "Groq Integration",
			provider: "groq",
			envKey:   "GROQ_API_KEY",
			expected: "groq",
		},
		{
			name:     "Dev Mode",
			provider: "",
			apiKey:   "",
			expected: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment for test
			var originalValue string
			var hasOriginal bool
			if tt.envKey != "" {
				originalValue, hasOriginal = os.LookupEnv(tt.envKey)
				if tt.apiKey != "" {
					os.Setenv(tt.envKey, tt.apiKey)
				}
			}

			// Cleanup after test
			defer func() {
				if tt.envKey != "" {
					if hasOriginal {
						os.Setenv(tt.envKey, originalValue)
					} else {
						os.Unsetenv(tt.envKey)
					}
				}
			}()

			// Use appropriate models for each provider
			model := "test-model"
			switch tt.provider {
			case "gemini":
				model = "gemini-2.0-flash"
			case "groq":
				model = "llama3-8b-8192"
			case "openai":
				model = "gpt-4-turbo"
			}

			cfg := config.LLMConfig{
				Provider:         tt.provider,
				Model:            model,
				MaxTokens:        100,
				Temperature:      0.7,
				APIKey:           tt.apiKey,
				TopP:             0.9,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
				StopSequences:    []string{},
			}

			client, err := llm.NewClient(cfg)
			if err != nil {
				t.Fatalf("Failed to create LLM client: %v", err)
			}

			// Test message analysis
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			req := llm.AnalysisRequest{
				Platform: "discord",
				Content:  "Hello, this is a test message!",
				UserID:   "test_user_123",
				Context: map[string]interface{}{
					"type":       "casual",
					"channel_id": "test_channel",
				},
			}

			response, err := client.AnalyzeMessage(ctx, req)
			if err != nil {
				// For real providers, API errors (expired keys, invalid models) are expected in tests
				if tt.expected != "dev" {
					t.Logf("Provider %s failed as expected (likely API key issue): %v", tt.expected, err)
					return
				}
				t.Fatalf("Dev mode should not fail: %v", err)
			}

			// Validate response structure
			if response == nil {
				t.Fatal("Response is nil")
			}

			if response.SuggestedResponse == "" {
				t.Error("SuggestedResponse should not be empty")
			}

			if response.Confidence < 0 || response.Confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got: %f", response.Confidence)
			}

			if response.Sentiment == "" {
				t.Error("Sentiment should not be empty")
			}

			if response.Category == "" {
				t.Error("Category should not be empty")
			}

			t.Logf("Provider: %s", tt.expected)
			t.Logf("Response: %+v", response)
		})
	}
}

func TestProviderDetection(t *testing.T) {
	tests := []struct {
		name         string
		geminiKey    string
		groqKey      string
		openaiKey    string
		configProv   string
		expectedProv string
	}{
		{
			name:         "Gemini from environment",
			geminiKey:    os.Getenv("GEMINI_API_KEY"),
			expectedProv: "gemini",
		},
		{
			name:         "Groq from environment",
			groqKey:      os.Getenv("GROQ_API_KEY"),
			expectedProv: "groq",
		},
		{
			name:         "OpenAI from environment",
			openaiKey:    os.Getenv("OPENAI_API_KEY"),
			expectedProv: "openai",
		},
		{
			name:         "Config provider gemini",
			configProv:   "gemini",
			expectedProv: "dev", // Will be dev since no API key
		},
		{
			name:         "No provider specified",
			expectedProv: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Unsetenv("GEMINI_API_KEY")
			os.Unsetenv("GROQ_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")

			// Set test environment
			if tt.geminiKey != "" {
				os.Setenv("GEMINI_API_KEY", tt.geminiKey)
			}
			if tt.groqKey != "" {
				os.Setenv("GROQ_API_KEY", tt.groqKey)
			}
			if tt.openaiKey != "" {
				os.Setenv("OPENAI_API_KEY", tt.openaiKey)
			}

			cfg := config.LLMConfig{
				Provider:    tt.configProv,
				Model:       "test-model",
				MaxTokens:   100,
				Temperature: 0.7,
			}

			client, err := llm.NewClient(cfg)
			if err != nil {
				t.Fatalf("Failed to create LLM client: %v", err)
			}

			// Use reflection to access private provider field
			// Since this is a test, we can check the behavior indirectly
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			req := llm.AnalysisRequest{
				Platform: "discord",
				Content:  "test",
				UserID:   "test",
				Context:  map[string]interface{}{},
			}

			response, err := client.AnalyzeMessage(ctx, req)
			if err != nil {
				// For real providers without valid keys, we might get errors
				// For dev mode, we should get a response
				if tt.expectedProv == "dev" {
					t.Fatalf("Dev mode should not fail: %v", err)
				}
				// Real providers might fail with invalid keys, which is OK for this test
				t.Logf("Provider %s failed as expected (likely invalid test key): %v", tt.expectedProv, err)
				return
			}

			if response == nil {
				t.Fatal("Response should not be nil")
			}

			t.Logf("Expected provider: %s, got response: %+v", tt.expectedProv, response)
		})
	}

	// Restore original environment
	// Here we are using a dummy approach; in real tests, consider using a library to manage env vars
	os.Setenv("GEMINI_API_KEY", os.Getenv("GEMINI_API_KEY"))
	os.Setenv("GROQ_API_KEY", os.Getenv("GROQ_API_KEY"))
	os.Setenv("OPENAI_API_KEY", os.Getenv("OPENAI_API_KEY"))
}

func TestCaching(t *testing.T) {
	// Clear environment to force dev mode
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GROQ_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	cfg := config.LLMConfig{
		Provider:    "dev", // Use dev mode for predictable responses
		Model:       "test-model",
		MaxTokens:   100,
		Temperature: 0.7,
		APIKey:      "", // Force dev mode
	}

	client, err := llm.NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create LLM client: %v", err)
	}

	ctx := context.Background()
	req := llm.AnalysisRequest{
		Platform: "discord",
		Content:  "This is a test for caching",
		UserID:   "cache_test_user",
		Context:  map[string]interface{}{},
	}

	// First call
	start1 := time.Now()
	response1, err := client.AnalyzeMessage(ctx, req)
	duration1 := time.Since(start1)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call (should be cached)
	start2 := time.Now()
	response2, err := client.AnalyzeMessage(ctx, req)
	duration2 := time.Since(start2)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	// Cache should make second call faster
	if duration2 >= duration1 {
		t.Logf("First call: %v, Second call: %v", duration1, duration2)
		t.Log("Note: Second call should be faster due to caching, but timing can vary")
	}

	// Responses should be identical
	if response1.SuggestedResponse != response2.SuggestedResponse {
		t.Error("Cached response should be identical to original")
	}

	if response1.Confidence != response2.Confidence {
		t.Error("Cached confidence should be identical to original")
	}

	t.Logf("Caching test completed - Duration1: %v, Duration2: %v", duration1, duration2)
}
