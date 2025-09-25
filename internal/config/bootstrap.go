package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// BootstrapMainConfig garante que o arquivo principal de configuração exista
// com um payload padrão compatível com as structs atuais. Se o arquivo já
// tiver conteúdo, nada é alterado.
func BootstrapMainConfig(path string) error {
	if path == "" {
		return errors.New("config path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			gl.Log("info", fmt.Sprintf("No config found at %s. Generating defaults.", path))
			return writeDefaultConfig(path)
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		gl.Log("warn", fmt.Sprintf("Config file at %s is empty. Hydrating defaults.", path))
		return writeDefaultConfig(path)
	}

	return nil
}

func writeDefaultConfig(path string) error {
	cfg := defaultConfig(path)

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	gl.Log("info", fmt.Sprintf("Default config stored at %s", path))
	return nil
}

func defaultConfig(path string) Config {
	return Config{
		ConfigFilePath: path,
		Discord: DiscordConfig{
			Bot: struct {
				Token       string   `json:"token"`
				Permissions []string `json:"permissions"`
				Intents     []string `json:"intents"`
				Channels    []string `json:"channels"`
			}{
				Token:       "",
				Permissions: []string{"READ_MESSAGES", "SEND_MESSAGES", "MANAGE_MESSAGES"},
				Intents:     []string{"GUILD_MESSAGES", "DIRECT_MESSAGES", "MESSAGE_CONTENT"},
				Channels:    []string{},
			},
			OAuth2: struct {
				ClientID     string   `json:"client_id"`
				ClientSecret string   `json:"client_secret"`
				RedirectURI  string   `json:"redirect_uri"`
				Scopes       []string `json:"scopes"`
			}{
				ClientID:     "",
				ClientSecret: "",
				RedirectURI:  "http://localhost:8088/api/v1/discord/oauth2/callback",
				Scopes:       []string{"bot", "applications.commands"},
			},
			Webhook: struct {
				URL    string `json:"url"`
				Secret string `json:"secret"`
			}{
				URL:    "",
				Secret: "",
			},
			RateLimits: struct {
				RequestsPerMinute int `json:"requests_per_minute"`
				BurstSize         int `json:"burst_size"`
			}{
				RequestsPerMinute: 60,
				BurstSize:         10,
			},
			Features: struct {
				AutoResponse            bool `json:"auto_response"`
				TaskCreation            bool `json:"task_creation"`
				CrossPlatformForwarding bool `json:"cross_platform_forwarding"`
			}{
				AutoResponse:            true,
				TaskCreation:            true,
				CrossPlatformForwarding: false,
			},
			DevMode: true,
		},
		LLM: LLMConfig{
			Provider:    "gemini",
			Model:       "gemini-1.5-flash",
			MaxTokens:   1024,
			Temperature: 0.3,
			APIKey:      "",
			DevMode:     true,
		},
		Approval: ApprovalConfig{
			RequireApprovalForResponses: false,
			ApprovalTimeoutMinutes:      15,
			DevMode:                     true,
		},
		Server: ServerConfig{
			Port:       8088,
			Host:       "0.0.0.0",
			EnableCORS: true,
			DevMode:    true,
		},
		ZMQ: ZMQConfig{
			Address: "tcp://127.0.0.1",
			Port:    5555,
			DevMode: true,
		},
		GoBE: GoBeConfig{
			BaseURL: "http://localhost:8088",
			APIKey:  "",
			Timeout: 30,
			Enabled: true,
			DevMode: true,
		},
		GobeCtl: GobeCtlConfig{
			Path:      "gobeCtl",
			Namespace: "default",
			Enabled:   false,
			DevMode:   true,
		},
		Integrations: IntegrationConfig{
			WhatsApp: WhatsAppConfig{
				Enabled:       false,
				AccessToken:   "",
				VerifyToken:   "",
				PhoneNumberID: "",
				WebhookURL:    "",
				DevMode:       true,
			},
			Telegram: TelegramConfig{
				Enabled:        false,
				BotToken:       "",
				WebhookURL:     "",
				AllowedUpdates: []string{"message", "callback_query"},
				DevMode:        true,
			},
			DevMode: true,
		},
		DevMode: true,
	}
}
