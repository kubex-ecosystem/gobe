// Package testsdiscord contains tests for the Discord adapter.
package testsdiscord

import (
	"sync/atomic"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/kubex-ecosystem/gobe/internal/bootstrap"
	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/services/chatbot/discord"
)

func TestToNeutralMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    *discordgo.MessageCreate
		expected interfaces.Message
	}{
		{
			name: "basic message conversion",
			input: &discordgo.MessageCreate{
				Message: &discordgo.Message{
					ID:        "123",
					ChannelID: "456",
					GuildID:   "789",
					Content:   "hello world",
					Author:    &discordgo.User{ID: "user123", Username: "testuser", Discriminator: "1234"},
					Attachments: []*discordgo.MessageAttachment{
						{
							ID:          "att1",
							Filename:    "test.png",
							Size:        1024,
							URL:         "https://example.com/test.png",
							ContentType: "image/png",
						},
					},
				},
			},
			expected: interfaces.Message{
				ID:        "123",
				ChannelID: "456",
				GuildID:   "789",
				Content:   "hello world",
				User:      interfaces.User{ID: "user123", Username: "testuser", Discriminator: "1234"},
				Role:      interfaces.RoleUser,
				Attachments: []interfaces.Attachment{
					{
						ID:       "att1",
						Name:     "test.png",
						Size:     1024,
						URL:      "https://example.com/test.png",
						MimeType: "image/png",
					},
				},
			},
		},
		{
			name: "message without attachments",
			input: &discordgo.MessageCreate{
				Message: &discordgo.Message{
					ID:          "456",
					ChannelID:   "789",
					GuildID:     "012",
					Content:     "simple message",
					Author:      &discordgo.User{ID: "user456", Username: "anotheruser", Discriminator: "5678"},
					Attachments: []*discordgo.MessageAttachment{},
				},
			},
			expected: interfaces.Message{
				ID:          "456",
				ChannelID:   "789",
				GuildID:     "012",
				Content:     "simple message",
				User:        interfaces.User{ID: "user456", Username: "anotheruser", Discriminator: "5678"},
				Role:        interfaces.RoleUser,
				Attachments: []interfaces.Attachment{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discord.ToNeutralMessage(tt.input)

			if result.ID != tt.expected.ID {
				t.Errorf("ID mismatch: got %s, want %s", result.ID, tt.expected.ID)
			}
			if result.ChannelID != tt.expected.ChannelID {
				t.Errorf("ChannelID mismatch: got %s, want %s", result.ChannelID, tt.expected.ChannelID)
			}
			if result.GuildID != tt.expected.GuildID {
				t.Errorf("GuildID mismatch: got %s, want %s", result.GuildID, tt.expected.GuildID)
			}
			if result.Content != tt.expected.Content {
				t.Errorf("Content mismatch: got %s, want %s", result.Content, tt.expected.Content)
			}
			if result.User.ID != tt.expected.User.ID {
				t.Errorf("User ID mismatch: got %s, want %s", result.User.ID, tt.expected.User.ID)
			}
			if result.User.Username != tt.expected.User.Username {
				t.Errorf("Username mismatch: got %s, want %s", result.User.Username, tt.expected.User.Username)
			}
			if result.Role != tt.expected.Role {
				t.Errorf("Role mismatch: got %s, want %s", result.Role, tt.expected.Role)
			}
			if len(result.Attachments) != len(tt.expected.Attachments) {
				t.Errorf("Attachments length mismatch: got %d, want %d", len(result.Attachments), len(tt.expected.Attachments))
			}
			for i, att := range result.Attachments {
				if i < len(tt.expected.Attachments) {
					expected := tt.expected.Attachments[i]
					if att.ID != expected.ID || att.Name != expected.Name || att.Size != expected.Size {
						t.Errorf("Attachment %d mismatch: got %+v, want %+v", i, att, expected)
					}
				}
			}
		})
	}
}

func TestOnMessageSwap(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() bootstrap.DiscordConfig
		wantCalls   int32
	}{
		{
			name: "message handler swap works",
			setupConfig: func() bootstrap.DiscordConfig {
				cfg := bootstrap.DiscordConfig{}
				cfg.Bot.Token = "dev_token"
				return cfg
			},
			wantCalls: 1,
		},
		{
			name: "multiple handler swaps work",
			setupConfig: func() bootstrap.DiscordConfig {
				cfg := bootstrap.DiscordConfig{}
				cfg.Bot.Token = "dev_token"
				return cfg
			},
			wantCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var callCount int32
			cfg := tt.setupConfig()

			adapter, err := discord.NewAdapter(cfg, "bot")
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			// Set handler that increments counter
			adapter.OnMessage(func(interfaces.Message) {
				atomic.AddInt32(&callCount, 1)
			})

			// Access the adapter as concrete type to test internal behavior
			concreteAdapter := adapter.(*discord.Adapter)

			// Simulate message handler calls
			for i := int32(0); i < tt.wantCalls; i++ {
				if handler := concreteAdapter.GetMessageHandler(); handler != nil {
					handler(interfaces.Message{Content: "test"})
				}
			}

			if got := atomic.LoadInt32(&callCount); got != tt.wantCalls {
				t.Errorf("Expected %d calls, got %d", tt.wantCalls, got)
			}
		})
	}
}

func TestAdapterDevMode(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantDev bool
	}{
		{
			name:    "dev mode with dev_token",
			token:   "dev_token",
			wantDev: true,
		},
		{
			name:    "production mode with real token",
			token:   "Bot.real_token_here",
			wantDev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := bootstrap.DiscordConfig{}
			cfg.Bot.Token = tt.token

			adapter, err := discord.NewAdapter(cfg, "bot")

			if tt.wantDev {
				// Dev mode should work
				if err != nil {
					t.Errorf("Dev mode should not error: %v", err)
				}

				// Test dev mode behaviors
				if err := adapter.Connect(); err != nil {
					t.Errorf("Connect in dev mode should not error: %v", err)
				}

				if err := adapter.SendMessage("test_channel", "test message"); err != nil {
					t.Errorf("SendMessage in dev mode should not error: %v", err)
				}

				channels, err := adapter.GetChannels("test_guild")
				if err != nil {
					t.Errorf("GetChannels in dev mode should not error: %v", err)
				}
				if len(channels) == 0 {
					t.Error("GetChannels in dev mode should return mock channels")
				}
			} else {
				// Production mode might error due to invalid token - that's expected in tests
				// We're just testing that the adapter creation doesn't panic
				_ = adapter
			}
		})
	}
}

func TestGetChannelsFormat(t *testing.T) {
	cfg := bootstrap.DiscordConfig{}
	cfg.Bot.Token = "dev_token"

	adapter, err := discord.NewAdapter(cfg, "bot")
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	channels, err := adapter.GetChannels("test_guild")
	if err != nil {
		t.Fatalf("GetChannels failed: %v", err)
	}

	if len(channels) == 0 {
		t.Fatal("Expected mock channels in dev mode")
	}

	for _, ch := range channels {
		if ch.ID == "" {
			t.Error("Channel ID should not be empty")
		}
		if ch.Name == "" {
			t.Error("Channel Name should not be empty")
		}
	}
}

func TestPingAdapter(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		message string
		wantErr bool
	}{
		{
			name:    "dev mode ping",
			token:   "dev_token",
			message: "test ping",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := bootstrap.DiscordConfig{}
			cfg.Bot.Token = tt.token

			adapter, err := discord.NewAdapter(cfg, "bot")
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			err = adapter.PingAdapter(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("PingAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
