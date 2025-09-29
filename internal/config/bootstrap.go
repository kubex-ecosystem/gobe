package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	common "github.com/kubex-ecosystem/gobe/internal/commons"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
)

// BootstrapMainConfig garante que o arquivo principal de configuração exista
// com um payload padrão compatível com as structs atuais. Se o arquivo já
// tiver conteúdo, nada é alterado.
func BootstrapMainConfig(path string, initArgs *t.InitArgs) error {
	var args t.InitArgs
	if initArgs != nil {
		args = *initArgs
		if args.ConfigFile != "" {
			args.ConfigFile = path
		}
	} else {
		args = t.InitArgs{
			ConfigFile:     path,
			IsConfidential: false,
			Port:           "8088",
			Bind:           "0.0.0.0",
		}
	}
	if args.ConfigFile == "" {
		configFile := os.Getenv("GDBASE_CONFIG_FILE")
		if configFile == "" {
			configFile = os.ExpandEnv(common.DefaultGoBEConfigPath)
			if configFile == "" {
				gl.Log("fatal", "No config file path provided via args or GDBASE_CONFIG_FILE env var, and default path is empty")
			}
		}
		args.ConfigFile = configFile
	}

	if err := os.MkdirAll(filepath.Dir(args.ConfigFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := os.ReadFile(args.ConfigFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			gl.Log("info", fmt.Sprintf("No config found at %s. Generating defaults.", args.ConfigFile))
			return writeDefaultConfig(args)
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		gl.Log("warn", fmt.Sprintf("Config file at %s is empty. Hydrating defaults.", args.ConfigFile))
		return writeDefaultConfig(args)
	}

	return nil
}

func writeDefaultConfig(initArgs t.InitArgs) error {
	cfg := defaultConfig(initArgs)

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(initArgs.ConfigFile, payload, 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	gl.Log("notice", fmt.Sprintf("Default config stored at %s", initArgs.ConfigFile))
	return nil
}

func defaultConfig(initArgs t.InitArgs) Config {
	return Config{
		ConfigFilePath: initArgs.ConfigFile,
		Discord: DiscordConfig{
			Bot: struct {
				ApplicationID string   `json:"application_id,omitempty"`
				Token         string   `json:"token,omitempty"`
				Permissions   []string `json:"permissions,omitempty"`
				Intents       []string `json:"intents,omitempty"`
				Channels      []string `json:"channels,omitempty"`
			}{
				ApplicationID: "",
				Token:         "",
				Permissions:   []string{"READ_MESSAGES", "SEND_MESSAGES", "MANAGE_MESSAGES"},
				Intents:       []string{"GUILD_MESSAGES", "DIRECT_MESSAGES", "MESSAGE_CONTENT"},
				Channels:      []string{},
			},
			OAuth2: struct {
				PublicKey    string   `json:"public_key,omitempty"`
				ClientID     string   `json:"client_id,omitempty"`
				ClientSecret string   `json:"client_secret,omitempty"`
				RedirectURI  string   `json:"redirect_uri,omitempty"`
				Scopes       []string `json:"scopes,omitempty"`
			}{
				ClientID:     "",
				ClientSecret: "",
				RedirectURI:  "http://" + net.JoinHostPort(initArgs.Bind, initArgs.Port) + "/api/v1/discord/oauth2/callback",
				Scopes:       []string{"bot", "applications.commands"},
			},
			Webhook: struct {
				URL    string `json:"url,omitempty"`
				Secret string `json:"secret,omitempty"`
			}{
				URL:    "",
				Secret: "",
			},
			RateLimits: struct {
				RequestsPerMinute int `json:"requests_per_minute,omitempty"`
				BurstSize         int `json:"burst_size,omitempty"`
			}{
				RequestsPerMinute: 60,
				BurstSize:         10,
			},
			Features: struct {
				AutoResponse            bool `json:"auto_response,omitempty"`
				TaskCreation            bool `json:"task_creation,omitempty"`
				CrossPlatformForwarding bool `json:"cross_platform_forwarding,omitempty"`
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
			Port:       initArgs.Port,
			Host:       initArgs.Bind,
			EnableCORS: true,
			DevMode:    true,
		},
		GoBE: GoBeConfig{
			BaseURL: "http://" + net.JoinHostPort(initArgs.Bind, initArgs.Port),
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
