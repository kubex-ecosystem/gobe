// Package config provides functionality to load and manage the application configuration.
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/rafa-mori/gobe/logger"

	l "github.com/rafa-mori/logz"
)

var gl = logger.GetLogger[l.Logger](nil)

type Config struct {
	Discord      DiscordConfig     `json:"discord"`
	LLM          LLMConfig         `json:"llm"`
	Approval     ApprovalConfig    `json:"approval"`
	Server       ServerConfig      `json:"server"`
	ZMQ          ZMQConfig         `json:"zmq"`
	GoBE         GoBeConfig        `json:"gobe"`
	GobeCtl      GobeCtlConfig     `json:"gobeCtl"`
	Integrations IntegrationConfig `json:"integrations"`
	DevMode      bool              `json:"dev_mode"`
}

type DiscordConfig struct {
	Bot struct {
		Token       string   `json:"token"`
		Permissions []string `json:"permissions"`
		Intents     []string `json:"intents"`
		Channels    []string `json:"channels"`
	} `json:"bot"`
	OAuth2 struct {
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		RedirectURI  string   `json:"redirect_uri"`
		Scopes       []string `json:"scopes"`
	} `json:"oauth2"`
	Webhook struct {
		URL    string `json:"url"`
		Secret string `json:"secret"`
	} `json:"webhook"`
	RateLimits struct {
		RequestsPerMinute int `json:"requests_per_minute"`
		BurstSize         int `json:"burst_size"`
	} `json:"rate_limits"`
	Features struct {
		AutoResponse            bool `json:"auto_response"`
		TaskCreation            bool `json:"task_creation"`
		CrossPlatformForwarding bool `json:"cross_platform_forwarding"`
	} `json:"features"`
}

type LLMConfig struct {
	Provider    string  `json:"provider" mapstructure:"provider"`
	Model       string  `json:"model" mapstructure:"model"`
	MaxTokens   int     `json:"max_tokens" mapstructure:"max_tokens"`
	Temperature float64 `json:"temperature" mapstructure:"temperature"`
	APIKey      string  `json:"api_key" mapstructure:"api_key"`
}

type ApprovalConfig struct {
	RequireApprovalForResponses bool `json:"require_approval_for_responses"`
	ApprovalTimeoutMinutes      int  `json:"approval_timeout_minutes"`
}

type ServerConfig struct {
	Port       int    `json:"port"`
	Host       string `json:"host"`
	EnableCORS bool   `json:"enable_cors"`
}

type ZMQConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type GoBeConfig struct {
	BaseURL string `json:"base_url" mapstructure:"base_url"`
	APIKey  string `json:"api_key" mapstructure:"api_key"`
	Timeout int    `json:"timeout" mapstructure:"timeout"`
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
}

type GobeCtlConfig struct {
	Path       string `json:"path" mapstructure:"path"`
	Namespace  string `json:"namespace" mapstructure:"namespace"`
	Kubeconfig string `json:"kubeconfig" mapstructure:"kubeconfig"`
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
}

type IntegrationConfig struct {
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Telegram TelegramConfig `json:"telegram"`
}

type WhatsAppConfig struct {
	Enabled       bool   `json:"enabled" mapstructure:"enabled"`
	AccessToken   string `json:"access_token" mapstructure:"access_token"`
	VerifyToken   string `json:"verify_token" mapstructure:"verify_token"`
	PhoneNumberID string `json:"phone_number_id" mapstructure:"phone_number_id"`
	WebhookURL    string `json:"webhook_url" mapstructure:"webhook_url"`
}

type TelegramConfig struct {
	Enabled        bool     `json:"enabled" mapstructure:"enabled"`
	BotToken       string   `json:"bot_token" mapstructure:"bot_token"`
	WebhookURL     string   `json:"webhook_url" mapstructure:"webhook_url"`
	AllowedUpdates []string `json:"allowed_updates" mapstructure:"allowed_updates"`
}

func Load(configPath string) (*Config, error) {
	// Check if .env file exists and load it
	if _, err := os.Stat("config/.env"); os.IsNotExist(err) {
		log.Println("No .env file found, skipping environment variable loading")
	} else if os.IsPermission(err) {
		return nil, fmt.Errorf("permission denied to read .env file: %w", err)
	} else {
		log.Println("Loading environment variables from .env file")
		if err := godotenv.Load("config/.env"); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Initialize viper
	viper.SetConfigName("config/discord_config.json")
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)

	// Set defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.enable_cors", true)
	viper.SetDefault("zmq.address", "tcp://127.0.0.1")
	viper.SetDefault("zmq.port", 5555)

	// Integrations defaults
	viper.SetDefault("integrations.whatsapp.enabled", false)
	viper.SetDefault("integrations.telegram.enabled", false)

	// GoBE defaults
	viper.SetDefault("gobe.base_url", "http://localhost:8080")
	viper.SetDefault("gobe.timeout", 30)
	viper.SetDefault("gobe.enabled", true)

	// gobe defaults
	viper.SetDefault("gobe.path", "gobeCtl")
	viper.SetDefault("gobe.namespace", "default")
	viper.SetDefault("gobe.enabled", true)

	// Check for dev mode
	devMode := false //os.Getenv("DEV_MODE") == "true"

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Set dev mode after reading config
	viper.Set("dev_mode", devMode)

	// Expand environment variables or set dev defaults
	if token := os.Getenv("DISCORD_BOT_TOKEN"); token != "" {
		viper.Set("discord.bot.token", token)
	} else if devMode {
		viper.Set("discord.bot.token", "dev_token")
	}

	// Discord OAuth2 configuration
	if clientID := os.Getenv("DISCORD_CLIENT_ID"); clientID != "" {
		viper.Set("discord.oauth2.client_id", clientID)
	}
	if clientSecret := os.Getenv("DISCORD_CLIENT_SECRET"); clientSecret != "" {
		viper.Set("discord.oauth2.client_secret", clientSecret)
	}
	if ngrokURL := os.Getenv("NGROK_URL"); ngrokURL != "" {
		viper.Set("discord.oauth2.redirect_uri", ngrokURL+"/discord/oauth2/authorize")
		gl.Log("info", "Using ngrok URL for Discord OAuth2 redirect:", ngrokURL)
	}

	// Set default OAuth2 scopes
	viper.SetDefault("discord.oauth2.scopes", []string{"bot", "applications.commands"})

	// WhatsApp configuration
	if waToken := os.Getenv("WHATSAPP_ACCESS_TOKEN"); waToken != "" {
		viper.Set("integrations.whatsapp.access_token", waToken)
	}
	if waVerify := os.Getenv("WHATSAPP_VERIFY_TOKEN"); waVerify != "" {
		viper.Set("integrations.whatsapp.verify_token", waVerify)
	}
	if waPhone := os.Getenv("WHATSAPP_PHONE_NUMBER_ID"); waPhone != "" {
		viper.Set("integrations.whatsapp.phone_number_id", waPhone)
	}
	if waWebhook := os.Getenv("WHATSAPP_WEBHOOK_URL"); waWebhook != "" {
		viper.Set("integrations.whatsapp.webhook_url", waWebhook)
	}

	// Telegram configuration
	if tgToken := os.Getenv("TELEGRAM_BOT_TOKEN"); tgToken != "" {
		viper.Set("integrations.telegram.bot_token", tgToken)
	}
	if tgWebhook := os.Getenv("TELEGRAM_WEBHOOK_URL"); tgWebhook != "" {
		viper.Set("integrations.telegram.webhook_url", tgWebhook)
	}

	// 🔗 GoBE Backend Integration
	if gobeURL := os.Getenv("GOBE_BASE_URL"); gobeURL != "" {
		viper.Set("gobe.base_url", gobeURL)
		viper.Set("gobe.enabled", true)
	}
	if gobeKey := os.Getenv("GOBE_API_KEY"); gobeKey != "" {
		viper.Set("gobe.api_key", gobeKey)
	}

	// ⚙️ gobe K8s Integration
	if gobePath := os.Getenv("KBXCTL_PATH"); gobePath != "" {
		viper.Set("gobe.path", gobePath)
		viper.Set("gobe.enabled", true)
	}
	if k8sNamespace := os.Getenv("K8S_NAMESPACE"); k8sNamespace != "" {
		viper.Set("gobe.namespace", k8sNamespace)
	}
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		viper.Set("gobe.kubeconfig", kubeconfig)
	}

	// For now, always use dev mode for LLM to focus on Discord testing
	geminiKey := os.Getenv("GEMINI_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")

	log.Printf("🔍 Config Debug - Environment Variables:")
	log.Printf("   GEMINI_API_KEY: '%s' (len=%d)", geminiKey, len(geminiKey))
	log.Printf("   OPENAI_API_KEY: '%s' (len=%d)", openaiKey, len(openaiKey))

	if geminiKey != "" && geminiKey != "dev_api_key" {
		viper.Set("llm.api_key", geminiKey)
		viper.Set("llm.provider", "gemini")
		log.Printf("   ✅ Using Gemini with key: %s...", geminiKey[:10])
	} else if openaiKey != "" && openaiKey != "dev_api_key" {
		viper.Set("llm.api_key", openaiKey)
		viper.Set("llm.provider", "openai")
		log.Printf("   ✅ Using OpenAI with key: %s...", openaiKey[:10])
	} else {
		viper.Set("llm.api_key", "dev_api_key")
		viper.Set("llm.provider", "dev")
		log.Printf("   ⚠️ Using DEV mode (no valid API keys found)")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Force dev mode values after unmarshal if in dev mode
	if devMode {
		config.DevMode = true
		if config.Discord.Bot.Token == "" {
			config.Discord.Bot.Token = "dev_token"
		}
		// Only override API key if it's actually empty
		if config.LLM.APIKey == "" {
			config.LLM.APIKey = "dev_api_key"
		}
	}

	return &config, nil
}
