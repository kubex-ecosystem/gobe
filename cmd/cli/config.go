package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var (
	configFile   string
	configKey    string
	configValue  string
	configFormat string
	configOutput string
	configGlobal bool
	configCreate bool
)

func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
		Long: `Manage GoBE configuration including reading, writing,
and validating configuration values.`,
	}

	cmd.AddCommand(configGetCmd())
	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configListCmd())
	cmd.AddCommand(configInitCmd())
	cmd.AddCommand(configValidateCmd())
	cmd.AddCommand(configResetCmd())

	return cmd
}

func configGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get configuration value",
		Long:  `Get a specific configuration value by key.`,
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := configKey
			if len(args) > 0 {
				key = args[0]
			}

			gl.Log("info", fmt.Sprintf("Getting configuration value: %s", key))

			if err := loadConfig(); err != nil {
				return err
			}

			var result interface{}
			if key == "" {
				// Get all config
				result = viper.AllSettings()
			} else {
				// Get specific key
				if !viper.IsSet(key) {
					return fmt.Errorf("configuration key not found: %s", key)
				}
				result = viper.Get(key)
			}

			output, err := formatConfigOutput(result, configFormat)
			if err != nil {
				return err
			}

			if configOutput != "" {
				return os.WriteFile(configOutput, []byte(output), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Configuration file path")
	cmd.Flags().StringVarP(&configKey, "key", "k", "", "Configuration key to get")
	cmd.Flags().StringVar(&configFormat, "format", "yaml", "Output format (json, yaml)")
	cmd.Flags().StringVarP(&configOutput, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolVarP(&configGlobal, "global", "g", false, "Use global configuration")

	return cmd
}

func configSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration value",
		Long:  `Set a configuration value for the specified key.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			gl.Log("info", fmt.Sprintf("Setting configuration: %s = %s", key, value))

			if err := loadConfig(); err != nil {
				return err
			}

			// Parse value as JSON if it looks like JSON
			var parsedValue interface{}
			if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") {
				if err := json.Unmarshal([]byte(value), &parsedValue); err != nil {
					parsedValue = value // Use as string if JSON parsing fails
				}
			} else if value == "true" || value == "false" {
				parsedValue = value == "true"
			} else {
				parsedValue = value
			}

			viper.Set(key, parsedValue)

			// Write configuration
			if err := writeConfig(); err != nil {
				return fmt.Errorf("failed to write configuration: %w", err)
			}

			fmt.Printf("Configuration updated: %s = %v\n", key, parsedValue)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Configuration file path")
	cmd.Flags().BoolVarP(&configGlobal, "global", "g", false, "Use global configuration")

	return cmd
}

func configListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Long:  `List all configuration keys and values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Listing all configuration values...")

			if err := loadConfig(); err != nil {
				return err
			}

			allSettings := viper.AllSettings()
			output, err := formatConfigOutput(allSettings, configFormat)
			if err != nil {
				return err
			}

			if configOutput != "" {
				return os.WriteFile(configOutput, []byte(output), 0644)
			}

			fmt.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Configuration file path")
	cmd.Flags().StringVar(&configFormat, "format", "yaml", "Output format (json, yaml, table)")
	cmd.Flags().StringVarP(&configOutput, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolVarP(&configGlobal, "global", "g", false, "Use global configuration")

	return cmd
}

func configInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		Long:  `Create a new configuration file with default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Initializing configuration...")

			configPath := getConfigPath()

			// Check if config already exists
			if _, err := os.Stat(configPath); err == nil && !configCreate {
				return fmt.Errorf("configuration file already exists: %s (use --force to overwrite)", configPath)
			}

			// Create directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			// Create default configuration
			defaultConfig := getDefaultConfig()

			// Write configuration file
			var data []byte
			var err error

			switch strings.ToLower(filepath.Ext(configPath)) {
			case ".json":
				data, err = json.MarshalIndent(defaultConfig, "", "  ")
			case ".yaml", ".yml":
				data, err = yaml.Marshal(defaultConfig)
			default:
				data, err = yaml.Marshal(defaultConfig)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal configuration: %w", err)
			}

			if err := os.WriteFile(configPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write configuration file: %w", err)
			}

			fmt.Printf("Configuration initialized: %s\n", configPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Configuration file path")
	cmd.Flags().BoolVarP(&configCreate, "force", "F", false, "Force overwrite existing configuration")
	cmd.Flags().BoolVarP(&configGlobal, "global", "g", false, "Use global configuration")

	return cmd
}

func configValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration file",
		Long:  `Validate the configuration file format and required values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("info", "Validating configuration...")

			if err := loadConfig(); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			// Basic validation
			validationResults := validateConfiguration()

			output, err := formatConfigOutput(validationResults, configFormat)
			if err != nil {
				return err
			}

			if configOutput != "" {
				return os.WriteFile(configOutput, []byte(output), 0644)
			}

			fmt.Println(output)

			// Check if validation passed
			if !validationResults["valid"].(bool) {
				return fmt.Errorf("configuration validation failed")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Configuration file path")
	cmd.Flags().StringVar(&configFormat, "format", "yaml", "Output format (json, yaml)")
	cmd.Flags().StringVarP(&configOutput, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolVarP(&configGlobal, "global", "g", false, "Use global configuration")

	return cmd
}

func configResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset configuration to defaults",
		Long:  `Reset configuration file to default values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gl.Log("warn", "Resetting configuration to defaults...")

			configPath := getConfigPath()

			// Backup existing config
			if _, err := os.Stat(configPath); err == nil {
				backupPath := configPath + ".backup"
				if err := os.Rename(configPath, backupPath); err != nil {
					gl.Log("warn", fmt.Sprintf("Failed to create backup: %v", err))
				} else {
					fmt.Printf("Existing configuration backed up to: %s\n", backupPath)
				}
			}

			// Create default configuration
			defaultConfig := getDefaultConfig()

			var data []byte
			var err error

			switch strings.ToLower(filepath.Ext(configPath)) {
			case ".json":
				data, err = json.MarshalIndent(defaultConfig, "", "  ")
			case ".yaml", ".yml":
				data, err = yaml.Marshal(defaultConfig)
			default:
				data, err = yaml.Marshal(defaultConfig)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal default configuration: %w", err)
			}

			if err := os.WriteFile(configPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write configuration file: %w", err)
			}

			fmt.Printf("Configuration reset to defaults: %s\n", configPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Configuration file path")
	cmd.Flags().BoolVarP(&configGlobal, "global", "g", false, "Use global configuration")

	return cmd
}

// Helper functions
func loadConfig() error {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		configPath := getConfigPath()
		viper.SetConfigFile(configPath)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			gl.Log("warn", "Configuration file not found, using defaults")
			return nil
		}
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	return nil
}

func writeConfig() error {
	configPath := configFile
	if configPath == "" {
		configPath = getConfigPath()
	}

	return viper.WriteConfigAs(configPath)
}

func getConfigPath() string {
	if configFile != "" {
		return configFile
	}

	if configGlobal {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".kubex", "gobe", "config.yaml")
	}

	return "config.yaml"
}

func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"port": 8080,
			"bind": "0.0.0.0",
			"debug": false,
		},
		"database": map[string]interface{}{
			"type": "sqlite",
			"path": "gobe.db",
		},
		"llm": map[string]interface{}{
			"provider": "gemini",
			"max_tokens": 2048,
			"temperature": 0.7,
		},
		"discord": map[string]interface{}{
			"enabled": false,
			"dev_mode": true,
		},
		"webhooks": map[string]interface{}{
			"enabled": true,
			"timeout": 30,
		},
		"security": map[string]interface{}{
			"auto_cert": true,
		},
	}
}

func validateConfiguration() map[string]interface{} {
	results := map[string]interface{}{
		"valid": true,
		"errors": []string{},
		"warnings": []string{},
	}

	errors := []string{}
	warnings := []string{}

	// Validate server config
	if !viper.IsSet("server.port") {
		warnings = append(warnings, "server.port not set, using default")
	}

	// Validate database config
	if !viper.IsSet("database.type") {
		warnings = append(warnings, "database.type not set, using default")
	}

	// Validate LLM config
	if !viper.IsSet("llm.provider") {
		warnings = append(warnings, "llm.provider not set, using default")
	}

	// Check for required environment variables
	if viper.GetString("llm.provider") == "openai" && os.Getenv("OPENAI_API_KEY") == "" {
		errors = append(errors, "OPENAI_API_KEY environment variable required for OpenAI provider")
	}

	if viper.GetString("llm.provider") == "gemini" && os.Getenv("GEMINI_API_KEY") == "" {
		errors = append(errors, "GEMINI_API_KEY environment variable required for Gemini provider")
	}

	results["errors"] = errors
	results["warnings"] = warnings
	results["valid"] = len(errors) == 0

	return results
}

func formatConfigOutput(data interface{}, format string) (string, error) {
	switch format {
	case "json":
		output, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format as JSON: %w", err)
		}
		return string(output), nil
	case "yaml":
		output, err := yaml.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to format as YAML: %w", err)
		}
		return string(output), nil
	case "table":
		if dataMap, ok := data.(map[string]interface{}); ok {
			output := "Configuration\n"
			output += "=============\n"
			for key, value := range dataMap {
				output += fmt.Sprintf("%-20s: %v\n", key, value)
			}
			return output, nil
		}
		return fmt.Sprintf("%+v", data), nil
	default:
		return fmt.Sprintf("%+v", data), nil
	}
}