// Package config provides functionality to load and manage the application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/utils"

	l "github.com/kubex-ecosystem/logz"
)

var gl = logger.GetLogger[l.Logger](nil)

func getFromConfigMap[T *Config | *DiscordConfig | *LLMConfig | *ApprovalConfig | *ServerConfig | *ZMQConfig | *GoBeConfig | *GobeCtlConfig | *IntegrationConfig | *WhatsAppConfig | *TelegramConfig | any](configType string) (T, bool) {
	switch configType {
	case "main_config":
		return IConfig(newConfig()).(T), true
	case "discord_config":
		var out = newDiscordConfig()
		return IConfig(out).(T), true
	case "llm_config":
		return IConfig(newLLMConfig()).(T), true
	case "approval_config":
		return IConfig(newApprovalConfig()).(T), true
	case "server_config":
		return IConfig(newServerConfig()).(T), true
	case "zmq_config":
		return IConfig(newZMQConfig()).(T), true
	case "gobe_config":
		return IConfig(newGoBeConfig()).(T), true
	case "gobeCtl_config":
		return IConfig(newGobeCtlConfig()).(T), true
	case "integration_config":
		return IConfig(newIntegrationConfig()).(T), true
	case "whatsapp_config":
		return IConfig(newWhatsAppConfig()).(T), true
	case "telegram_config":
		return IConfig(newTelegramConfig()).(T), true
	}
	return *new(T), false
}

type IConfig interface {
	// Define methods that your configuration struct should implement
	GetSettings() map[string]interface{}
	GetType() string
	SetDevMode(bool)
}

type Config struct {
	ConfigFilePath string            `json:"config_file_path"`
	Discord        DiscordConfig     `json:"discord"`
	LLM            LLMConfig         `json:"llm"`
	Approval       ApprovalConfig    `json:"approval"`
	Server         ServerConfig      `json:"server"`
	ZMQ            ZMQConfig         `json:"zmq"`
	GoBE           GoBeConfig        `json:"gobe"`
	GobeCtl        GobeCtlConfig     `json:"gobeCtl"`
	Integrations   IntegrationConfig `json:"integrations"`
	DevMode        bool              `json:"dev_mode"`
}

func newConfig() *Config              { return &Config{} }
func NewConfig() IConfig              { return newConfig() }
func (c *Config) GetType() string     { return "gobe_config" }
func (c *Config) SetDevMode(dev bool) { c.DevMode = dev }
func (c *Config) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["discord"] = c.Discord
	settings["llm"] = c.LLM
	settings["approval"] = c.Approval
	settings["server"] = c.Server
	settings["zmq"] = c.ZMQ
	settings["gobe"] = c.GoBE
	settings["gobeCtl"] = c.GobeCtl
	settings["integrations"] = c.Integrations
	settings["dev_mode"] = c.DevMode
	return settings
}

type DiscordConfig struct {
	Bot struct {
		ApplicationID string   `json:"application_id"`
		Token         string   `json:"token"`
		Permissions   []string `json:"permissions"`
		Intents       []string `json:"intents"`
		Channels      []string `json:"channels"`
	} `json:"bot"`
	OAuth2 struct {
		PublicKey    string   `json:"public_key"`
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
	DevMode bool `json:"dev_mode"`
}

func newDiscordConfig() *DiscordConfig       { return &DiscordConfig{} }
func NewDiscordConfig() *DiscordConfig       { return newDiscordConfig() }
func (c *DiscordConfig) GetType() string     { return "discord_config" }
func (c *DiscordConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *DiscordConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["bot"] = c.Bot.Intents
	settings["oauth2"] = c.OAuth2.ClientID
	settings["webhook"] = c.Webhook.URL
	settings["rate_limits"] = c.RateLimits.RequestsPerMinute
	settings["features"] = c.Features
	return settings
}

type LLMConfig struct {
	Provider         string   `json:"provider" mapstructure:"provider"`
	Model            string   `json:"model" mapstructure:"model"`
	MaxTokens        int      `json:"max_tokens" mapstructure:"max_tokens"`
	Temperature      float64  `json:"temperature" mapstructure:"temperature"`
	APIKey           string   `json:"api_key" mapstructure:"api_key"`
	DevMode          bool     `json:"dev_mode"`
	TopP             float64  `json:"top_p" mapstructure:"top_p"`
	FrequencyPenalty float64  `json:"frequency_penalty" mapstructure:"frequency_penalty"`
	PresencePenalty  float64  `json:"presence_penalty" mapstructure:"presence_penalty"`
	StopSequences    []string `json:"stop_sequences" mapstructure:"stop_sequences"`
}

func newLLMConfig() *LLMConfig           { return &LLMConfig{} }
func NewLLMConfig() *LLMConfig           { return newLLMConfig() }
func (c *LLMConfig) GetType() string     { return "llm_config" }
func (c *LLMConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *LLMConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["provider"] = c.Provider
	settings["model"] = c.Model
	settings["max_tokens"] = c.MaxTokens
	settings["temperature"] = c.Temperature
	// Do not include API key for security reasons
	return settings
}

type ApprovalConfig struct {
	RequireApprovalForResponses bool `json:"require_approval_for_responses"`
	ApprovalTimeoutMinutes      int  `json:"approval_timeout_minutes"`
	DevMode                     bool `json:"dev_mode"`
}

func newApprovalConfig() *ApprovalConfig      { return &ApprovalConfig{} }
func NewApprovalConfig() *ApprovalConfig      { return newApprovalConfig() }
func (c *ApprovalConfig) GetType() string     { return "approval_config" }
func (c *ApprovalConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *ApprovalConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["require_approval_for_responses"] = c.RequireApprovalForResponses
	settings["approval_timeout_minutes"] = c.ApprovalTimeoutMinutes
	return settings
}

type ServerConfig struct {
	Port       string `json:"port"`
	Host       string `json:"host"`
	EnableCORS bool   `json:"enable_cors"`
	DevMode    bool   `json:"dev_mode"`
}

func newServerConfig() *ServerConfig        { return &ServerConfig{} }
func NewServerConfig() *ServerConfig        { return newServerConfig() }
func (c *ServerConfig) GetType() string     { return "server_config" }
func (c *ServerConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *ServerConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["port"] = c.Port
	settings["host"] = c.Host
	settings["enable_cors"] = c.EnableCORS
	return settings
}

type ZMQConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	DevMode bool   `json:"dev_mode"`
}

func newZMQConfig() *ZMQConfig           { return &ZMQConfig{} }
func NewZMQConfig() *ZMQConfig           { return newZMQConfig() }
func (c *ZMQConfig) GetType() string     { return "zmq_config" }
func (c *ZMQConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *ZMQConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["address"] = c.Address
	settings["port"] = c.Port
	return settings
}

type GoBeConfig struct {
	BaseURL string `json:"base_url" mapstructure:"base_url"`
	APIKey  string `json:"api_key" mapstructure:"api_key"`
	Timeout int    `json:"timeout" mapstructure:"timeout"`
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	DevMode bool   `json:"dev_mode" mapstructure:"dev_mode"`
}

func newGoBeConfig() *GoBeConfig          { return &GoBeConfig{} }
func NewGoBeConfig() *GoBeConfig          { return newGoBeConfig() }
func (c *GoBeConfig) GetType() string     { return "gobe_config" }
func (c *GoBeConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *GoBeConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["base_url"] = c.BaseURL
	// Do not include API key for security reasons
	settings["timeout"] = c.Timeout
	settings["enabled"] = c.Enabled
	return settings
}

type GobeCtlConfig struct {
	Path       string `json:"path" mapstructure:"path"`
	Namespace  string `json:"namespace" mapstructure:"namespace"`
	Kubeconfig string `json:"kubeconfig" mapstructure:"kubeconfig"`
	Enabled    bool   `json:"enabled" mapstructure:"enabled"`
	DevMode    bool   `json:"dev_mode" mapstructure:"dev_mode"`
}

func newGobeCtlConfig() *GobeCtlConfig       { return &GobeCtlConfig{} }
func NewGobeCtlConfig() *GobeCtlConfig       { return newGobeCtlConfig() }
func (c *GobeCtlConfig) GetType() string     { return "gobe_ctl_config" }
func (c *GobeCtlConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *GobeCtlConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["path"] = c.Path
	settings["namespace"] = c.Namespace
	settings["kubeconfig"] = c.Kubeconfig
	settings["enabled"] = c.Enabled
	return settings
}

type IntegrationConfig struct {
	WhatsApp WhatsAppConfig `json:"whatsapp"`
	Telegram TelegramConfig `json:"telegram"`
	DevMode  bool           `json:"dev_mode"`
}

func newIntegrationConfig() *IntegrationConfig   { return &IntegrationConfig{} }
func NewIntegrationConfig() *IntegrationConfig   { return newIntegrationConfig() }
func (c *IntegrationConfig) GetType() string     { return "integration_config" }
func (c *IntegrationConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *IntegrationConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["whatsapp"] = c.WhatsApp.Enabled
	settings["telegram"] = c.Telegram.Enabled
	return settings
}

type WhatsAppConfig struct {
	Enabled       bool   `json:"enabled" mapstructure:"enabled"`
	AccessToken   string `json:"access_token" mapstructure:"access_token"`
	VerifyToken   string `json:"verify_token" mapstructure:"verify_token"`
	PhoneNumberID string `json:"phone_number_id" mapstructure:"phone_number_id"`
	WebhookURL    string `json:"webhook_url" mapstructure:"webhook_url"`
	DevMode       bool   `json:"dev_mode" mapstructure:"dev_mode"`
}

func newWhatsAppConfig() *WhatsAppConfig      { return &WhatsAppConfig{} }
func NewWhatsAppConfig() *WhatsAppConfig      { return newWhatsAppConfig() }
func (c *WhatsAppConfig) GetType() string     { return "whatsapp_config" }
func (c *WhatsAppConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *WhatsAppConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["enabled"] = c.Enabled
	// Do not include AccessToken or VerifyToken for security reasons
	settings["phone_number_id"] = c.PhoneNumberID
	settings["webhook_url"] = c.WebhookURL
	return settings
}

type TelegramConfig struct {
	Enabled        bool     `json:"enabled" mapstructure:"enabled"`
	BotToken       string   `json:"bot_token" mapstructure:"bot_token"`
	WebhookURL     string   `json:"webhook_url" mapstructure:"webhook_url"`
	AllowedUpdates []string `json:"allowed_updates" mapstructure:"allowed_updates"`
	DevMode        bool     `json:"dev_mode" mapstructure:"dev_mode"`
}

func newTelegramConfig() *TelegramConfig      { return &TelegramConfig{} }
func NewTelegramConfig() *TelegramConfig      { return newTelegramConfig() }
func (c *TelegramConfig) GetType() string     { return "telegram_config" }
func (c *TelegramConfig) SetDevMode(dev bool) { c.DevMode = dev }
func (c *TelegramConfig) GetSettings() map[string]interface{} {
	settings := make(map[string]interface{})
	settings["enabled"] = c.Enabled
	// Do not include BotToken for security reasons
	settings["webhook_url"] = c.WebhookURL
	settings["allowed_updates"] = c.AllowedUpdates
	return settings
}

func Load[C *Config | *DiscordConfig | *LLMConfig | *ApprovalConfig | *ServerConfig | *ZMQConfig | *GoBeConfig | *GobeCtlConfig | *IntegrationConfig | *WhatsAppConfig | *TelegramConfig | *IConfig](
	configPath string,
	configType string,
	initArgs *interfaces.InitArgs,
) (C, error) {

	var envFilePath string

	// Check if .env file exists and load it

	if configPath == "" {
		gl.Log("warn", "No config path provided, using default:", configPath)
		envFilePath = ".env"
		configPath = GetConfigFilePath()
		configPath = filepath.Join(configPath, "gobe", "config.json")
	}

	if info, statErr := os.Stat(configPath); statErr == nil {
		if info.IsDir() {
			gobeDir := configPath
			if filepath.Base(gobeDir) != "gobe" {
				candidate := filepath.Join(gobeDir, "gobe")
				if _, err := os.Stat(candidate); err == nil {
					gobeDir = candidate
				}
			}
			configPath = filepath.Join(gobeDir, "config.json")
		}
	} else if os.IsNotExist(statErr) {
		if filepath.Ext(configPath) == "" && !strings.HasSuffix(configPath, ".json") {
			configPath = filepath.Join(configPath, "gobe", "config.json")
		}
	}

	if reflect.TypeFor[C]().String() == "*config.Config" && configType == "gobe_config" {
		configType = "main_config"
	}

	if err := BootstrapMainConfig(configPath, initArgs); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to bootstrap config file: %v", err))
	}

	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		gl.Log("info", ".env file not found, skipping loading environment variables from file")
		goto postEnvLoad
	}

	gl.Log("info", "Loading settings from .env file")
	if err := godotenv.Load(envFilePath); err != nil {
		// return nil, fmt.Errorf("error loading %s file: %w", envFilePath, err)
		gl.Log("fatal", fmt.Sprintf("error loading %s file: %v", envFilePath, err))
	}
	gl.Log("info", "Loaded environment variables from .env file")

postEnvLoad:
	gl.Log("info", "Using config path:", configPath)
	if configType == "" {
		configType = "main_config"
		gl.Log("info", "No config type provided, using default: main_config")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		gl.Log("info", "No config.json file found, skipping environment variable loading")
	} else if os.IsPermission(err) {
		return nil, fmt.Errorf("permission denied to read config.json file: %w", err)
	}

	// Initialize viper
	viper.SetConfigName(filepath.Base(configPath))
	viper.SetConfigType("json")
	viper.AddConfigPath(filepath.Dir(configPath))

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

	// Discord Bot configuration
	if appID := os.Getenv("DISCORD_APPLICATION_ID"); appID != "" {
		viper.Set("discord.bot.application_id", appID)
	} else if devMode {
		viper.Set("discord.bot.application_id", "dev_application_id")
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

	// log.Printf("🔍 Config Debug - Environment Variables:")
	// log.Printf("   GEMINI_API_KEY: '%s' (len=%d)", geminiKey, len(geminiKey))
	// log.Printf("   OPENAI_API_KEY: '%s' (len=%d)", openaiKey, len(openaiKey))

	if geminiKey != "" && geminiKey != "dev_api_key" {
		viper.Set("llm.api_key", geminiKey)
		viper.Set("llm.provider", "gemini")
		// log.Printf("   ✅ Using Gemini with key: %s...", geminiKey[:10])
		gl.Log("info", fmt.Sprintf("LLM Config - Using Gemini with key: %s...", geminiKey[:10]))
	} else if openaiKey != "" && openaiKey != "dev_api_key" {
		viper.Set("llm.api_key", openaiKey)
		viper.Set("llm.provider", "openai")
		gl.Log("info", fmt.Sprintf("LLM Config - Using OpenAI with key: %s...", openaiKey[:10]))
	} else {
		viper.Set("llm.api_key", "dev_api_key")
		viper.Set("llm.provider", "dev")
		gl.Log("warn", "LLM Config - Using DEV mode (no valid API keys found)")
	}

	viper.SetConfigName(fmt.Sprintf("config/%s.json", configType))
	configInstance, exists := getFromConfigMap[C](configType)
	if !exists {
		return nil, fmt.Errorf("no constructor found for config type: %s", configType)
	}

	// Unmarshal into the provided config struct
	if err := viper.UnmarshalKey("", &configInstance); err != nil {
		return nil, fmt.Errorf("error unmarshaling %s: %w", configType, err)
	}

	// Force dev mode values after unmarshal if in dev mode
	if devMode {
		inter := reflect.ValueOf(configInstance).Interface()
		switch reflect.TypeFor[C]().String() {
		case "*LLMConfig":
			inter.(*LLMConfig).DevMode = true
		case "*DiscordConfig":
			inter.(*DiscordConfig).DevMode = true
		case "*GoBeConfig":
			inter.(*GoBeConfig).DevMode = true
		case "*GobeCtlConfig":
			inter.(*GobeCtlConfig).DevMode = true
		}
		// Set default values for dev models
		switch configType {
		case "llm_config":
			// Set dev defaults for LLM
			viper.Set("llm.model", "gpt-3.5-turbo")
			viper.Set("llm.max_tokens", 500)
			viper.Set("llm.temperature", 0.7)
			if inter.(*LLMConfig).APIKey == "" || inter.(*LLMConfig).APIKey == "dev_api_key" {
				inter.(*LLMConfig).APIKey = "dev_api_key"
			}
			gl.Log("info", "LLM Config - Model:", inter.(*LLMConfig).Model, "MaxTokens:", inter.(*LLMConfig).MaxTokens, "Temperature:", inter.(*LLMConfig).Temperature)
		case "discord_config":
			if inter.(*DiscordConfig).Bot.Token == "" || inter.(*DiscordConfig).Bot.Token == "dev_token" {
				inter.(*DiscordConfig).Bot.Token = "dev_token"
			}
			if inter.(*DiscordConfig).Bot.ApplicationID == "" || inter.(*DiscordConfig).Bot.ApplicationID == "dev_application_id" {
				inter.(*DiscordConfig).Bot.ApplicationID = "dev_application_id"
			}
			gl.Log("info", "Discord Config - Bot Token set for Dev Mode")
		case "gobe_config":
			if inter.(*GoBeConfig).BaseURL == "" {
				inter.(*GoBeConfig).BaseURL = "http://localhost:8080"
			}
			if inter.(*GoBeConfig).APIKey == "" || inter.(*GoBeConfig).APIKey == "dev_api_key" {
				inter.(*GoBeConfig).APIKey = "dev_api_key"
			}
			gl.Log("info", "GoBE Config - BaseURL:", inter.(*GoBeConfig).BaseURL)
		case "gobe_ctl_config":
			if inter.(*GobeCtlConfig).Path == "" {
				inter.(*GobeCtlConfig).Path = "gobeCtl"
			}
			if inter.(*GobeCtlConfig).Namespace == "" {
				inter.(*GobeCtlConfig).Namespace = "default"
			}
			gl.Log("info", "GoBE CTL Config - Path:", inter.(*GobeCtlConfig).Path, "Namespace:", inter.(*GobeCtlConfig).Namespace)
		}
		configInstance = inter.(C)
	}

	return configInstance, nil
}

func GetConfigFilePath() string {
	var cfgPath string
	if path, err := utils.GetDefaultConfigPath(); path != "" && err == nil {
		cfgPath = path
	} else {
		gl.Log("fatal", "Failed to determine config path, using current directory:", err)
	}
	return cfgPath
}
