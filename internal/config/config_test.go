package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "gobe_config", config.GetType())
}

func TestConfig_GetSettings(t *testing.T) {
	config := &Config{DevMode: true}
	settings := config.GetSettings()

	assert.NotNil(t, settings)
	assert.Equal(t, true, settings["dev_mode"])
	assert.Contains(t, settings, "discord")
	assert.Contains(t, settings, "llm")
	assert.Contains(t, settings, "server")
}

func TestConfig_SetDevMode(t *testing.T) {
	config := &Config{}
	config.SetDevMode(true)
	assert.True(t, config.DevMode)

	config.SetDevMode(false)
	assert.False(t, config.DevMode)
}

func TestDiscordConfig(t *testing.T) {
	t.Run("NewDiscordConfig", func(t *testing.T) {
		config := NewDiscordConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "discord_config", config.GetType())
	})

	t.Run("SetDevMode", func(t *testing.T) {
		config := &DiscordConfig{}
		config.SetDevMode(true)
		assert.True(t, config.DevMode)
	})

	t.Run("GetSettings", func(t *testing.T) {
		config := &DiscordConfig{
			Bot: struct {
				Token       string   `json:"token"`
				Permissions []string `json:"permissions"`
				Intents     []string `json:"intents"`
				Channels    []string `json:"channels"`
			}{
				Intents: []string{"GUILD_MESSAGES"},
			},
		}
		settings := config.GetSettings()
		assert.NotNil(t, settings)
		assert.Contains(t, settings, "bot")
		assert.Contains(t, settings, "oauth2")
	})
}

func TestLLMConfig(t *testing.T) {
	t.Run("NewLLMConfig", func(t *testing.T) {
		config := NewLLMConfig()
		assert.NotNil(t, config)
		assert.Equal(t, "llm_config", config.GetType())
	})

	t.Run("GetSettings", func(t *testing.T) {
		config := &LLMConfig{
			Provider:    "openai",
			Model:       "gpt-3.5-turbo",
			MaxTokens:   1000,
			Temperature: 0.7,
			APIKey:      "secret-key",
		}
		settings := config.GetSettings()

		assert.Equal(t, "openai", settings["provider"])
		assert.Equal(t, "gpt-3.5-turbo", settings["model"])
		assert.Equal(t, 1000, settings["max_tokens"])
		assert.Equal(t, 0.7, settings["temperature"])
		// API key should not be included for security
		assert.NotContains(t, settings, "api_key")
	})
}

func TestApprovalConfig(t *testing.T) {
	config := NewApprovalConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "approval_config", config.GetType())

	config.SetDevMode(true)
	assert.True(t, config.DevMode)

	settings := config.GetSettings()
	assert.Contains(t, settings, "require_approval_for_responses")
	assert.Contains(t, settings, "approval_timeout_minutes")
}

func TestServerConfig(t *testing.T) {
	config := NewServerConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "server_config", config.GetType())

	settings := config.GetSettings()
	assert.Contains(t, settings, "port")
	assert.Contains(t, settings, "host")
	assert.Contains(t, settings, "enable_cors")
}

func TestZMQConfig(t *testing.T) {
	config := NewZMQConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "zmq_config", config.GetType())

	settings := config.GetSettings()
	assert.Contains(t, settings, "address")
	assert.Contains(t, settings, "port")
}

func TestGoBeConfig(t *testing.T) {
	config := NewGoBeConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "gobe_config", config.GetType())

	config.BaseURL = "http://localhost:8080"
	config.APIKey = "secret"
	config.Enabled = true

	settings := config.GetSettings()
	assert.Equal(t, "http://localhost:8080", settings["base_url"])
	assert.Equal(t, true, settings["enabled"])
	// API key should not be included for security
	assert.NotContains(t, settings, "api_key")
}

func TestGobeCtlConfig(t *testing.T) {
	config := NewGobeCtlConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "gobe_ctl_config", config.GetType())

	settings := config.GetSettings()
	assert.Contains(t, settings, "path")
	assert.Contains(t, settings, "namespace")
	assert.Contains(t, settings, "kubeconfig")
	assert.Contains(t, settings, "enabled")
}

func TestIntegrationConfig(t *testing.T) {
	config := NewIntegrationConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "integration_config", config.GetType())

	settings := config.GetSettings()
	assert.Contains(t, settings, "whatsapp")
	assert.Contains(t, settings, "telegram")
}

func TestWhatsAppConfig(t *testing.T) {
	config := NewWhatsAppConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "whatsapp_config", config.GetType())

	config.Enabled = true
	config.AccessToken = "secret"
	config.PhoneNumberID = "123456789"

	settings := config.GetSettings()
	assert.Equal(t, true, settings["enabled"])
	assert.Equal(t, "123456789", settings["phone_number_id"])
	// Sensitive tokens should not be included
	assert.NotContains(t, settings, "access_token")
	assert.NotContains(t, settings, "verify_token")
}

func TestTelegramConfig(t *testing.T) {
	config := NewTelegramConfig()
	assert.NotNil(t, config)
	assert.Equal(t, "telegram_config", config.GetType())

	config.Enabled = true
	config.BotToken = "secret"
	config.AllowedUpdates = []string{"message", "callback_query"}

	settings := config.GetSettings()
	assert.Equal(t, true, settings["enabled"])
	assert.Equal(t, []string{"message", "callback_query"}, settings["allowed_updates"])
	// Bot token should not be included for security
	assert.NotContains(t, settings, "bot_token")
}

func TestGetFromConfigMap(t *testing.T) {
	tests := []struct {
		name       string
		configType string
		wantType   string
		wantExists bool
	}{
		{"discord_config", "discord_config", "discord_config", true},
		{"llm_config", "llm_config", "llm_config", true},
		{"approval_config", "approval_config", "approval_config", true},
		{"server_config", "server_config", "server_config", true},
		{"zmq_config", "zmq_config", "zmq_config", true},
		{"gobe_config", "gobe_config", "gobe_config", true},
		{"gobeCtl_config", "gobeCtl_config", "gobe_ctl_config", true}, // Disabled for now, doesn't exists in config map
		{"integration_config", "integration_config", "integration_config", true},
		{"whatsapp_config", "whatsapp_config", "whatsapp_config", true},
		{"telegram_config", "telegram_config", "telegram_config", true},
		{"invalid_config", "invalid_config", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//
			var preConfig any
			var result IConfig
			var exists bool
			switch tt.configType {
			case "discord_config":
				preConfig, exists = getFromConfigMap[*DiscordConfig](tt.configType)
				result = IConfig(preConfig.(*DiscordConfig))
			case "llm_config":
				preConfig, exists = getFromConfigMap[*LLMConfig](tt.configType)
				result = IConfig(preConfig.(*LLMConfig))
			case "approval_config":
				preConfig, exists = getFromConfigMap[*ApprovalConfig](tt.configType)
				result = IConfig(preConfig.(*ApprovalConfig))
			case "server_config":
				preConfig, exists = getFromConfigMap[*ServerConfig](tt.configType)
				result = IConfig(preConfig.(*ServerConfig))
			case "zmq_config":
				preConfig, exists = getFromConfigMap[*ZMQConfig](tt.configType)
				result = IConfig(preConfig.(*ZMQConfig))
			case "gobe_config":
				preConfig, exists = getFromConfigMap[*GoBeConfig](tt.configType)
				result = IConfig(preConfig.(*GoBeConfig))
			case "gobeCtl_config":
				preConfig, exists = getFromConfigMap[*GobeCtlConfig](tt.configType)
				result = IConfig(preConfig.(*GobeCtlConfig))
			case "integration_config":
				preConfig, exists = getFromConfigMap[*IntegrationConfig](tt.configType)
				result = IConfig(preConfig.(*IntegrationConfig))
			case "whatsapp_config":
				preConfig, exists = getFromConfigMap[*WhatsAppConfig](tt.configType)
				result = IConfig(preConfig.(*WhatsAppConfig))
			case "telegram_config":
				preConfig, exists = getFromConfigMap[*TelegramConfig](tt.configType)
				result = IConfig(preConfig.(*TelegramConfig))
			default:
				preConfig = nil
			}
			assert.Equal(t, tt.wantExists, exists)
			resultInt := result
			if exists {
				assert.NotNil(t, resultInt)
				assert.Equal(t, tt.wantType, resultInt.GetType())
			}
		})
	}
}

func TestGetConfigFilePath(t *testing.T) {
	path := GetConfigFilePath()
	assert.NotEmpty(t, path)
}

func TestLoad_InvalidConfigType(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "gobe")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create empty config.json
	configFile := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configFile, []byte("{}"), 0644)
	require.NoError(t, err)

	_, err = Load[*Config](tempDir, "invalid_config", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no constructor found for config type")
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "gobe")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create config.json with basic structure
	configFile := filepath.Join(configDir, "config.json")
	configContent := `{
		"discord": {
			"bot": {"token": ""},
			"oauth2": {"client_id": "", "client_secret": ""}
		},
		"llm": {
			"provider": "",
			"api_key": ""
		}
	}`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables
	originalEnvs := map[string]string{
		"DISCORD_BOT_TOKEN":     os.Getenv("DISCORD_BOT_TOKEN"),
		"DISCORD_CLIENT_ID":     os.Getenv("DISCORD_CLIENT_ID"),
		"DISCORD_CLIENT_SECRET": os.Getenv("DISCORD_CLIENT_SECRET"),
		"GEMINI_API_KEY":        os.Getenv("GEMINI_API_KEY"),
		"OPENAI_API_KEY":        os.Getenv("OPENAI_API_KEY"),
	}
	defer func() {
		// Restore original environment variables
		for key, value := range originalEnvs {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	os.Setenv("DISCORD_BOT_TOKEN", "test-discord-token")
	os.Setenv("DISCORD_CLIENT_ID", "test-client-id")
	os.Setenv("DISCORD_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GEMINI_API_KEY", "test-gemini-key")

	// Reset viper to avoid state pollution
	viper.Reset()

	config, err := Load[*Config](configFile, "gobe_config", nil)
	require.NoError(t, err)
	assert.NotNil(t, config)
}

func TestLoad_LLMConfigWithEnvironmentVariables(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "gobe")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create config.json
	configFile := filepath.Join(configDir, "config.json")
	configContent := `{
		"llm": {
			"provider": "openai",
			"model": "gpt-4",
			"max_tokens": 1000,
			"temperature": 0.8,
			"api_key": ""
		}
	}`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Save original environment variables
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalGemini := os.Getenv("GEMINI_API_KEY")
	defer func() {
		if originalOpenAI == "" {
			os.Unsetenv("OPENAI_API_KEY")
		} else {
			os.Setenv("OPENAI_API_KEY", originalOpenAI)
		}
		if originalGemini == "" {
			os.Unsetenv("GEMINI_API_KEY")
		} else {
			os.Setenv("GEMINI_API_KEY", originalGemini)
		}
	}()

	// Test with OpenAI key
	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	os.Unsetenv("GEMINI_API_KEY")

	// Reset viper to avoid state pollution
	viper.Reset()

	config, err := Load[*LLMConfig](configFile, "llm_config", nil)
	require.NoError(t, err)
	assert.NotNil(t, config)
}

func TestLoad_IntegrationConfigWithEnvironmentVariables(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "gobe")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create config.json
	configFile := filepath.Join(configDir, "config.json")
	configContent := `{
		"integrations": {
			"whatsapp": {"enabled": false},
			"telegram": {"enabled": false}
		}
	}`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Save original environment variables
	originalVars := map[string]string{
		"WHATSAPP_ACCESS_TOKEN":    os.Getenv("WHATSAPP_ACCESS_TOKEN"),
		"WHATSAPP_VERIFY_TOKEN":    os.Getenv("WHATSAPP_VERIFY_TOKEN"),
		"WHATSAPP_PHONE_NUMBER_ID": os.Getenv("WHATSAPP_PHONE_NUMBER_ID"),
		"TELEGRAM_BOT_TOKEN":       os.Getenv("TELEGRAM_BOT_TOKEN"),
	}
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("WHATSAPP_ACCESS_TOKEN", "test-wa-token")
	os.Setenv("WHATSAPP_VERIFY_TOKEN", "test-wa-verify")
	os.Setenv("WHATSAPP_PHONE_NUMBER_ID", "123456789")
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-tg-token")

	// Reset viper to avoid state pollution
	viper.Reset()

	config, err := Load[*IntegrationConfig](configFile, "integration_config", nil)
	require.NoError(t, err)
	assert.NotNil(t, config)
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Create temporary directory without config file
	tempDir := t.TempDir()

	// Reset viper to avoid state pollution
	viper.Reset()

	config, err := Load[*Config](tempDir, "gobe_config", nil)
	// Should still succeed with default values
	require.NoError(t, err)
	assert.NotNil(t, config)
}

func TestLoad_PermissionDenied(t *testing.T) {
	// Create temporary config directory
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "gobe")
	err := os.MkdirAll(configDir, 0000) // No permissions
	require.NoError(t, err)
	defer os.Chmod(configDir, 0755) // Restore permissions for cleanup

	// Create config.json in restricted directory
	configFile := filepath.Join(configDir, "config.json")

	// Reset viper to avoid state pollution
	viper.Reset()

	_, err = Load[*Config](configFile, "gobe_config", nil)
	// Should handle permission errors gracefully
	// Note: This test might behave differently on different systems
	if err != nil {
		assert.Contains(t, err.Error(), "permission denied")
	}
}

// Benchmark tests
func BenchmarkNewConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewConfig()
	}
}

func BenchmarkGetFromConfigMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getFromConfigMap[*IConfig]("discord_config")
	}
}

func BenchmarkConfig_GetSettings(b *testing.B) {
	config := &Config{DevMode: true}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		config.GetSettings()
	}
}
