// Package testsintegration provides integration tests for the application.
package testsintegration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/proxy/hub"
	"github.com/kubex-ecosystem/gobe/internal/services/llm"
)

func TestDiscordMCPHubLLMIntegration(t *testing.T) {
	// Create test configuration with dev mode
	cfg := &bootstrap.Config{
		Discord: bootstrap.DiscordConfig{
			DevMode: true,
		},
		LLM: bootstrap.LLMConfig{
			Provider:    "dev",
			Model:       "test-model",
			MaxTokens:   100,
			Temperature: 0.7,
			APIKey:      "", // Force dev mode
		},
		Approval: bootstrap.ApprovalConfig{
			RequireApprovalForResponses: false,
			ApprovalTimeoutMinutes:      5,
			DevMode:                     true,
		},
		Server: bootstrap.ServerConfig{
			Port:       "8080",
			Host:       "localhost",
			EnableCORS: true,
			DevMode:    true,
		},
		GoBE: bootstrap.GoBeConfig{
			Enabled: false,
			DevMode: true,
		},
		GobeCtl: bootstrap.GobeCtlConfig{
			Enabled: false,
			DevMode: true,
		},
		Integrations: bootstrap.IntegrationConfig{
			DevMode: true,
		},
		DevMode: true,
	}

	// Create hub
	discordHub, err := hub.NewDiscordMCPHub(cfg)
	if err != nil {
		t.Fatalf("Failed to create Discord MCP Hub: %v", err)
	}

	// Test message handling through the hub
	testCases := []struct {
		name             string
		message          interfaces.Message
		expectedAnalysis bool
	}{
		{
			name: "Question Message",
			message: interfaces.Message{
				ID:        "test_msg_1",
				ChannelID: "test_channel",
				GuildID:   "test_guild",
				User: interfaces.User{
					ID:       "test_user",
					Username: "testuser",
				},
				Role:      interfaces.RoleUser,
				Content:   "Como posso fazer deploy de uma aplicação?",
				Timestamp: time.Now(),
			},
			expectedAnalysis: true,
		},
		{
			name: "Task Request Message",
			message: interfaces.Message{
				ID:        "test_msg_2",
				ChannelID: "test_channel",
				GuildID:   "test_guild",
				User: interfaces.User{
					ID:       "test_user",
					Username: "testuser",
				},
				Role:      interfaces.RoleUser,
				Content:   "Criar uma tarefa para revisar o código",
				Timestamp: time.Now(),
			},
			expectedAnalysis: true,
		},
		{
			name: "Casual Message",
			message: interfaces.Message{
				ID:        "test_msg_3",
				ChannelID: "test_channel",
				GuildID:   "test_guild",
				User: interfaces.User{
					ID:       "test_user",
					Username: "testuser",
				},
				Role:      interfaces.RoleUser,
				Content:   "Oi, tudo bem?",
				Timestamp: time.Now(),
			},
			expectedAnalysis: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test the LLM analysis directly through the hub's processing methods
			err := discordHub.ProcessMessageWithLLM(ctx, tc.message)
			if err != nil {
				t.Fatalf("Failed to process message with LLM: %v", err)
			}

			// Also test direct LLM analysis to verify the integration
			llmClient, err := llm.NewClient(cfg.LLM)
			if err != nil {
				t.Fatalf("Failed to create LLM client: %v", err)
			}

			analysis, err := llmClient.AnalyzeMessage(ctx, llm.AnalysisRequest{
				Platform: "discord",
				Content:  tc.message.Content,
				UserID:   tc.message.User.ID,
				Context: map[string]interface{}{
					"channel_id": tc.message.ChannelID,
					"type":       "question",
				},
			})

			if err != nil {
				t.Fatalf("LLM analysis failed: %v", err)
			}

			if tc.expectedAnalysis {
				if analysis == nil {
					t.Fatal("Expected analysis response but got nil")
				}

				if analysis.SuggestedResponse == "" {
					t.Error("Expected suggested response but got empty string")
				}

				if analysis.Confidence < 0 || analysis.Confidence > 1 {
					t.Errorf("Invalid confidence: %f", analysis.Confidence)
				}

				t.Logf("Message: %s", tc.message.Content)
				t.Logf("Analysis: %+v", analysis)
				t.Logf("Response: %s", analysis.SuggestedResponse)
			}
		})
	}
}

func TestLLMClientProviderSelection(t *testing.T) {
	// Store original environment
	origGemini := os.Getenv("GEMINI_API_KEY")
	origGroq := os.Getenv("GROQ_API_KEY")
	origOpenAI := os.Getenv("OPENAI_API_KEY")

	// Cleanup function
	defer func() {
		if origGemini != "" {
			os.Setenv("GEMINI_API_KEY", origGemini)
		} else {
			os.Unsetenv("GEMINI_API_KEY")
		}
		if origGroq != "" {
			os.Setenv("GROQ_API_KEY", origGroq)
		} else {
			os.Unsetenv("GROQ_API_KEY")
		}
		if origOpenAI != "" {
			os.Setenv("OPENAI_API_KEY", origOpenAI)
		} else {
			os.Unsetenv("OPENAI_API_KEY")
		}
	}()

	testCases := []struct {
		name           string
		config         bootstrap.LLMConfig
		clearEnv       bool
		expectedResult bool
	}{
		{
			name: "Dev Mode",
			config: bootstrap.LLMConfig{
				Provider: "dev",
				Model:    "test-model",
				APIKey:   "",
			},
			clearEnv:       true,
			expectedResult: true,
		},
		{
			name: "Empty Config - Should fallback to dev",
			config: bootstrap.LLMConfig{
				Provider: "",
				Model:    "",
				APIKey:   "",
			},
			clearEnv:       true,
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clear environment if needed
			if tc.clearEnv {
				os.Unsetenv("GEMINI_API_KEY")
				os.Unsetenv("GROQ_API_KEY")
				os.Unsetenv("OPENAI_API_KEY")
			}

			client, err := llm.NewClient(tc.config)
			if err != nil {
				if tc.expectedResult {
					t.Fatalf("Expected success but got error: %v", err)
				}
				// Expected error, test passes
				return
			}

			if !tc.expectedResult {
				t.Fatal("Expected error but got success")
			}

			// Test basic functionality
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			req := llm.AnalysisRequest{
				Platform: "discord",
				Content:  "test message",
				UserID:   "test_user",
				Context:  map[string]interface{}{},
			}

			response, err := client.AnalyzeMessage(ctx, req)
			if err != nil {
				t.Fatalf("Analysis failed: %v", err)
			}

			if response == nil {
				t.Fatal("Response is nil")
			}

			t.Logf("Provider test successful: %+v", response)
		})
	}
}
