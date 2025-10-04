// Package bootstrap provides the initialization logic for the GoBE framework.
package bootstrap

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// BootstrapMainConfig garante que o arquivo principal de configuração exista
// com um payload padrão compatível com as structs atuais. Se o arquivo já
// tiver conteúdo, nada é alterado.
func BootstrapMainConfig[C *Config | *DiscordConfig |
	*LLMConfig | *ApprovalConfig |
	*ServerConfig | *GoBeConfig | *GobeCtlConfig |
	*IntegrationConfig | *WhatsAppConfig |
	*MCPServerConfig | *TelegramConfig |
	*IConfig](
	args *gl.InitArgs,
) (C, error) {
	if gl.IsObjValid(args) {
		if args.ConfigFile != "" {
			args.ConfigFile = os.ExpandEnv(args.ConfigFile)
		}
	} else {
		args = &gl.InitArgs{
			ConfigFile:     args.ConfigFile,
			IsConfidential: gl.GetEnvOrDefault("IS_CONFIDENTIAL", "false") == "true",
			Port:           "8088",
			Bind:           "0.0.0.0",
		}
	}

	// Configure config file path
	envVars := make(map[string]string)
	args.EnvFile, _ = gl.GetValueOrDefault(os.ExpandEnv(args.EnvFile), os.ExpandEnv(filepath.Join("$PWD", ".env")))
	if _, err := os.Stat(args.EnvFile); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(args.EnvFile), 0755); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating default config directory: %v", err))
			}
		}
		if _, err := os.Stat(args.EnvFile); err != nil {
			if err := os.WriteFile(args.EnvFile, []byte(""), 0644); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating config file: %v", err))
			}
		}
	} else if os.IsPermission(err) {
		gl.Log("fatal", fmt.Sprintf("permission denied to read %s file: %v", args.EnvFile, err))
	} else {
		gl.Log("debug", "Loading settings from .env file")
		if err := godotenv.Load(args.EnvFile); err != nil {
			// return nil, fmt.Errorf("error loading %s file: %w", envFilePath, err)
			gl.Log("fatal", fmt.Sprintf("error loading %s file: %v", args.EnvFile, err))
		}
		// .env file not found, skipping loading environment variables from file, reading from user/process environment
		gl.Log("debug", ".env file not found, skipping loading environment variables from file")
		envVars, err = godotenv.Read()
		if err != nil {
			gl.Log("fatal", fmt.Sprintf("error reading %s file: %v", args.EnvFile, err))
		}
	}

	args.ConfigFile, _ = gl.GetValueOrDefault(os.ExpandEnv(args.ConfigFile), os.ExpandEnv(gl.DefaultGoBEConfigPath))
	if _, err := os.Stat(filepath.Dir(args.ConfigFile)); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(args.ConfigFile), 0755); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating default config directory: %v", err))
			}
		}
		if _, err := os.Stat(args.ConfigFile); err != nil {
			if err := os.WriteFile(args.ConfigFile, []byte(""), 0644); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating config file: %v", err))
			}
		}
	}
	args.ConfigDBFile, _ = gl.GetValueOrDefault(os.ExpandEnv(args.ConfigDBFile), os.ExpandEnv(gl.DefaultGDBaseConfigPath))
	if _, err := os.Stat(filepath.Dir(args.ConfigDBFile)); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(args.ConfigDBFile), 0755); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating default config directory: %v", err))
			}
		}
		if _, err := os.Stat(args.ConfigDBFile); err != nil {
			if err := os.WriteFile(args.ConfigDBFile, []byte(""), 0644); err != nil {
				gl.Log("fatal", fmt.Sprintf("Error creating config file: %v", err))
			}
		}
	}
	var err error
	var wasEOF bool
	cfg, err := Load[C](args)
	if err != nil {
		wasEOF = strings.Contains(err.Error(), "objeto: EOF")
		if wasEOF {
			gl.Log("warn", fmt.Sprintf("Config file at %s is empty. Hydrating defaults.", args.ConfigFile))
			return writeDefaultConfig[C](args)
		} else {
			gl.Log("fatal", fmt.Sprintf("Error loading config from %s: %v", args.ConfigFile, err))
		}
	}

	if reflect.TypeFor[C]() == reflect.TypeFor[*Config]() {
		cfgMain, oks := any(cfg).(*Config)
		if !oks || !gl.IsObjValid(cfgMain) {
			gl.Log("warn", fmt.Sprintf("Config file at %s is invalid. Hydrating defaults.", args.ConfigFile))
			return writeDefaultConfig[C](args)
		}

		// Ensure critical fields are set
		updated := false
		if cfgMain.Server.Port == "" {
			cfgMain.Server.Port = args.Port
			updated = true
		}
		if cfgMain.Server.Host == "" {
			cfgMain.Server.Host = args.Bind
			updated = true
		}
		if cfgMain.Discord.OAuth2.RedirectURI == "" {
			cfgMain.Discord.OAuth2.RedirectURI = "http://" + net.JoinHostPort(args.Bind, args.Port) + "/api/v1/discord/oauth2/callback"
			updated = true
		}
		if cfgMain.GoBE.BaseURL == "" {
			cfgMain.GoBE.BaseURL = "http://" + net.JoinHostPort(args.Bind, args.Port)
			updated = true
		}
		if updated {
			gl.Log("info", fmt.Sprintf("Updating config file at %s with missing default values.", args.ConfigFile))
			return writeDefaultConfig[C](args)
		}

		gl.Log("info", fmt.Sprintf("Config loaded from %s", args.ConfigFile))
		// Hydrate any missing default values
		if err := hydrateConfigDefaults(cfgMain, args); err != nil {
			return nil, fmt.Errorf("failed to hydrate config defaults: %w", err)
		}

		// Override with environment variables if set
		if v, exists := envVars["GOBE_SERVER_PORT"]; exists && v != "" {
			cfgMain.Server.Port = v
		}
		if v, exists := envVars["GOBE_SERVER_HOST"]; exists && v != "" {
			cfgMain.Server.Host = v
		}
		if v, exists := envVars["GOBE_DISCORD_BOT_TOKEN"]; exists && v != "" {
			cfgMain.Discord.Bot.Token = v
		}
		if v, exists := envVars["GOBE_DISCORD_OAUTH2_CLIENT_ID"]; exists && v != "" {
			cfgMain.Discord.OAuth2.ClientID = v
		}
		if v, exists := envVars["GOBE_DISCORD_OAUTH2_CLIENT_SECRET"]; exists && v != "" {
			cfgMain.Discord.OAuth2.ClientSecret = v
		}
		if v, exists := envVars["GOBE_LLM_API_KEY"]; exists && v != "" {
			cfgMain.LLM.APIKey = v
		}
		if v, exists := envVars["GOBE_GOBE_API_KEY"]; exists && v != "" {
			cfgMain.GoBE.APIKey = v
		}
		if v, exists := envVars["GOBE_GOBE_BASE_URL"]; exists && v != "" {
			cfgMain.GoBE.BaseURL = v
		}

		cfg = any(cfgMain).(C)
	}

	return cfg, nil
}

func writeDefaultConfig[C *Config | *DiscordConfig |
	*LLMConfig | *ApprovalConfig |
	*ServerConfig | *GoBeConfig | *GobeCtlConfig |
	*IntegrationConfig | *WhatsAppConfig |
	*MCPServerConfig | *TelegramConfig |
	*IConfig](args *gl.InitArgs) (C, error) {
	cfg := defaultConfig[C](args)

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(args.ConfigFile, payload, 0644); err != nil {
		return nil, fmt.Errorf("failed to write default config: %w", err)
	}

	gl.Log("notice", fmt.Sprintf("Default config stored at %s", args.ConfigFile))
	return cfg, nil
}

func defaultConfig[C *Config | *DiscordConfig |
	*LLMConfig | *ApprovalConfig |
	*ServerConfig | *GoBeConfig | *GobeCtlConfig |
	*IntegrationConfig | *WhatsAppConfig |
	*MCPServerConfig | *TelegramConfig |
	*IConfig](args *gl.InitArgs) C {
	cfg, err := Load[C](args)
	if err == nil && gl.IsObjValid(cfg) {
		return cfg
	}
	config := any(&Config{
		ConfigFilePath: args.ConfigFile,
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
				RedirectURI:  "http://" + net.JoinHostPort(args.Bind, args.Port) + "/api/v1/discord/oauth2/callback",
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
			Model:       "gemini-2.0-flash",
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
			Port:       args.Port,
			Host:       args.Bind,
			EnableCORS: true,
			DevMode:    true,
		},
		GoBE: GoBeConfig{
			BaseURL: "http://" + net.JoinHostPort(args.Bind, args.Port),
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
	}).(C)

	return config
}

func hydrateConfigDefaults(cfg *Config, args *gl.InitArgs) error {
	updated := false
	if cfg.Server.Port == "" {
		cfg.Server.Port = args.Port
		updated = true
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = args.Bind
		updated = true
	}
	if cfg.Discord.OAuth2.RedirectURI == "" {
		cfg.Discord.OAuth2.RedirectURI = "http://" + net.JoinHostPort(args.Bind, args.Port) + "/api/v1/discord/oauth2/callback"
		updated = true
	}
	if cfg.GoBE.BaseURL == "" {
		cfg.GoBE.BaseURL = "http://" + net.JoinHostPort(args.Bind, args.Port)
		updated = true
	}
	if updated {
		payload, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal updated config: %w", err)
		}

		if err := os.WriteFile(args.ConfigFile, payload, 0644); err != nil {
			return fmt.Errorf("failed to write updated config: %w", err)
		}

		gl.Log("info", fmt.Sprintf("Updated config stored at %s", args.ConfigFile))
	}
	return nil
}
