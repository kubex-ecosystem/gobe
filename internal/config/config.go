// Package config provides functionality to load and manage the application configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	"github.com/kubex-ecosystem/gobe/internal/contracts/types"
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
		ApplicationID string   `json:"application_id,omitempty"`
		Token         string   `json:"token,omitempty"`
		Permissions   []string `json:"permissions,omitempty"`
		Intents       []string `json:"intents,omitempty"`
		Channels      []string `json:"channels,omitempty"`
	} `json:"bot,omitempty"`
	OAuth2 struct {
		PublicKey    string   `json:"public_key,omitempty"`
		ClientID     string   `json:"client_id,omitempty"`
		ClientSecret string   `json:"client_secret,omitempty"`
		RedirectURI  string   `json:"redirect_uri,omitempty"`
		Scopes       []string `json:"scopes,omitempty"`
	} `json:"oauth2,omitempty"`
	Webhook struct {
		URL    string `json:"url,omitempty"`
		Secret string `json:"secret,omitempty"`
	} `json:"webhook,omitempty"`
	RateLimits struct {
		RequestsPerMinute int `json:"requests_per_minute,omitempty"`
		BurstSize         int `json:"burst_size,omitempty"`
	} `json:"rate_limits,omitempty"`
	Features struct {
		AutoResponse            bool `json:"auto_response,omitempty"`
		TaskCreation            bool `json:"task_creation,omitempty"`
		CrossPlatformForwarding bool `json:"cross_platform_forwarding,omitempty"`
	} `json:"features,omitempty"`
	DevMode bool `json:"dev_mode,omitempty"`
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

	envVars := make(map[string]string)
	envFilePath := ".env"

	if _, err := os.Stat(envFilePath); os.IsNotExist(err) {
		// .env file not found, skipping loading environment variables from file, reading from user/process environment
		gl.Log("debug", ".env file not found, skipping loading environment variables from file")
		envVarsList := os.Environ()
		for _, envVar := range envVarsList {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	} else if os.IsPermission(err) {
		gl.Log("fatal", fmt.Sprintf("permission denied to read %s file: %v", envFilePath, err))
	} else {
		gl.Log("debug", "Loading settings from .env file")
		if err := godotenv.Load(envFilePath); err != nil {
			// return nil, fmt.Errorf("error loading %s file: %w", envFilePath, err)
			gl.Log("fatal", fmt.Sprintf("error loading %s file: %v", envFilePath, err))
		}
	}

	if configPath == "" {
		gl.Log("warn", "No config path provided, using default:", configPath)
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

	// Map environment variables to config

	for key, value := range envVars {
		viper.Set(key, value)
	}

	configInstanceMapper := types.NewMapper(new(C), configPath)
	configInstance, err := configInstanceMapper.DeserializeFromFile("json")
	if err != nil {
		return nil, fmt.Errorf("error deserializing %s: %w", configType, err)
	}
	if configInstance == nil {
		return nil, fmt.Errorf("deserialized config instance is nil for type: %s", configType)
	}
	return *configInstance, nil
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

func GetEnvOrDefault(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
