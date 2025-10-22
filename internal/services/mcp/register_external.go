// Package mcp provides external tool registrations for Kubex ecosystem modules.
// This file registers tools from Grompt (prompt engineering) and Analyzer (code analysis).
package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gl "github.com/kubex-ecosystem/logz/logger"
)

// ExternalToolsConfig holds configuration for external Kubex ecosystem tools
type ExternalToolsConfig struct {
	GromptURL   string // Default: http://localhost:8080
	AnalyzerURL string // Default: http://localhost:8081
	Timeout     time.Duration
}

// DefaultExternalConfig returns default configuration for external tools
func DefaultExternalConfig() ExternalToolsConfig {
	return ExternalToolsConfig{
		GromptURL:   "http://localhost:8080",
		AnalyzerURL: "http://localhost:8081",
		Timeout:     30 * time.Second,
	}
}

// RegisterExternalTools registers all external Kubex ecosystem tools
func RegisterExternalTools(registry Registry, config ExternalToolsConfig) error {
	if registry == nil {
		gl.Log("error", "Registry is nil, cannot register external tools")
		return fmt.Errorf("registry cannot be nil")
	}

	// Use default config if zero values
	if config.GromptURL == "" {
		config.GromptURL = "http://localhost:8080"
	}
	if config.AnalyzerURL == "" {
		config.AnalyzerURL = "http://localhost:8081"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Register Grompt tools
	if err := registerGromptTools(registry, config); err != nil {
		gl.Log("error", "Failed to register Grompt tools", err)
		return fmt.Errorf("failed to register Grompt tools: %w", err)
	}

	// Register Analyzer tools
	if err := registerAnalyzerTools(registry, config); err != nil {
		gl.Log("error", "Failed to register Analyzer tools", err)
		return fmt.Errorf("failed to register Analyzer tools: %w", err)
	}

	gl.Log("info", "External MCP tools registered successfully")
	return nil
}

// registerGromptTools registers Grompt prompt engineering tools
func registerGromptTools(registry Registry, config ExternalToolsConfig) error {
	// grompt.generate - Generate structured prompts from raw ideas
	generateSpec := ToolSpec{
		Name:        "grompt.generate",
		Title:       "Grompt Prompt Generator",
		Description: "Generate professional, structured prompts from raw ideas using Grompt engine with BYOK support",
		Auth:        "none",
		Args: map[string]interface{}{
			"ideas": map[string]interface{}{
				"type":        "array",
				"description": "Array of raw ideas, concepts, or requirements",
				"required":    true,
			},
			"purpose": map[string]interface{}{
				"type":        "string",
				"description": "Purpose of the prompt (e.g., 'Code Generation', 'Creative Writing')",
				"required":    true,
			},
			"provider": map[string]interface{}{
				"type":        "string",
				"description": "AI provider to use (openai, claude, gemini, deepseek, chatgpt)",
				"default":     "gemini",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Specific model to use (optional)",
				"default":     "",
			},
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "Optional external API key (BYOK - Bring Your Own Key)",
				"default":     "",
			},
			"max_tokens": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum tokens for generation",
				"default":     5000,
			},
		},
		Handler: createGromptGenerateHandler(config),
	}

	if err := registry.Register(generateSpec); err != nil {
		return fmt.Errorf("failed to register grompt.generate: %w", err)
	}

	// grompt.direct - Direct prompt execution (skip prompt engineering)
	directSpec := ToolSpec{
		Name:        "grompt.direct",
		Title:       "Grompt Direct Prompt",
		Description: "Send a direct prompt to AI provider via Grompt without prompt engineering",
		Auth:        "none",
		Args: map[string]interface{}{
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "The direct prompt to send",
				"required":    true,
			},
			"provider": map[string]interface{}{
				"type":        "string",
				"description": "AI provider (openai, claude, gemini, deepseek, chatgpt)",
				"default":     "gemini",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Specific model to use",
				"default":     "",
			},
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "Optional external API key (BYOK)",
				"default":     "",
			},
			"max_tokens": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum tokens for response",
				"default":     1000,
			},
		},
		Handler: createGromptDirectHandler(config),
	}

	if err := registry.Register(directSpec); err != nil {
		return fmt.Errorf("failed to register grompt.direct: %w", err)
	}

	gl.Log("info", "Grompt tools registered successfully")
	return nil
}

// registerAnalyzerTools registers Analyzer code analysis tools
func registerAnalyzerTools(registry Registry, config ExternalToolsConfig) error {
	// analyzer.project - Analyze project structure and dependencies
	projectSpec := ToolSpec{
		Name:        "analyzer.project",
		Title:       "Project Analyzer",
		Description: "Deep analysis of project structure, dependencies, and code quality",
		Auth:        "none",
		Args: map[string]interface{}{
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the project directory",
				"required":    true,
			},
			"depth": map[string]interface{}{
				"type":        "integer",
				"description": "Analysis depth (1-5)",
				"default":     3,
			},
			"include_dependencies": map[string]interface{}{
				"type":        "boolean",
				"description": "Include dependency graph analysis",
				"default":     true,
			},
		},
		Handler: createAnalyzerProjectHandler(config),
	}

	if err := registry.Register(projectSpec); err != nil {
		return fmt.Errorf("failed to register analyzer.project: %w", err)
	}

	// analyzer.security - Security audit and vulnerability detection
	securitySpec := ToolSpec{
		Name:        "analyzer.security",
		Title:       "Security Analyzer",
		Description: "Perform security audit and detect potential vulnerabilities",
		Auth:        "admin", // Security analysis requires admin
		Args: map[string]interface{}{
			"project_path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the project directory",
				"required":    true,
			},
			"severity_threshold": map[string]interface{}{
				"type":        "string",
				"description": "Minimum severity level (low, medium, high, critical)",
				"default":     "medium",
			},
		},
		Handler: createAnalyzerSecurityHandler(config),
	}

	if err := registry.Register(securitySpec); err != nil {
		return fmt.Errorf("failed to register analyzer.security: %w", err)
	}

	gl.Log("info", "Analyzer tools registered successfully")
	return nil
}

// createGromptGenerateHandler creates handler for grompt.generate tool
func createGromptGenerateHandler(config ExternalToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		gl.Log("info", "Executing grompt.generate tool")

		// Extract and validate ideas
		ideasInterface, ok := args["ideas"]
		if !ok {
			return nil, fmt.Errorf("ideas parameter is required")
		}

		ideas := []string{}
		switch v := ideasInterface.(type) {
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					ideas = append(ideas, str)
				}
			}
		case []string:
			ideas = v
		default:
			return nil, fmt.Errorf("ideas must be an array of strings")
		}

		if len(ideas) == 0 {
			return nil, fmt.Errorf("at least one idea is required")
		}

		// Extract purpose
		purpose, ok := args["purpose"].(string)
		if !ok || purpose == "" {
			return nil, fmt.Errorf("purpose parameter is required")
		}

		// Extract optional parameters
		provider := getStringArg(args, "provider", "gemini")
		model := getStringArg(args, "model", "")
		apiKey := getStringArg(args, "api_key", "")
		maxTokens := getIntArg(args, "max_tokens", 5000)

		// Build request payload
		payload := map[string]interface{}{
			"ideas":      ideas,
			"purpose":    purpose,
			"provider":   provider,
			"max_tokens": maxTokens,
			"lang":       "português",
		}

		if model != "" {
			payload["model"] = model
		}

		// Call Grompt API
		result, err := callGromptAPI(ctx, config, "/api/unified", payload, apiKey)
		if err != nil {
			gl.Log("error", "Grompt API call failed", err)
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Failed to generate prompt: %v", err),
			}, nil
		}

		return result, nil
	}
}

// createGromptDirectHandler creates handler for grompt.direct tool
func createGromptDirectHandler(config ExternalToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		gl.Log("info", "Executing grompt.direct tool")

		// Extract prompt
		prompt, ok := args["prompt"].(string)
		if !ok || prompt == "" {
			return nil, fmt.Errorf("prompt parameter is required")
		}

		// Extract optional parameters
		provider := getStringArg(args, "provider", "gemini")
		model := getStringArg(args, "model", "")
		apiKey := getStringArg(args, "api_key", "")
		maxTokens := getIntArg(args, "max_tokens", 1000)

		// Build request payload
		payload := map[string]interface{}{
			"prompt":     prompt,
			"provider":   provider,
			"max_tokens": maxTokens,
		}

		if model != "" {
			payload["model"] = model
		}

		// Call Grompt API
		result, err := callGromptAPI(ctx, config, "/api/unified", payload, apiKey)
		if err != nil {
			gl.Log("error", "Grompt API call failed", err)
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Failed to execute prompt: %v", err),
			}, nil
		}

		return result, nil
	}
}

// createAnalyzerProjectHandler creates handler for analyzer.project tool
func createAnalyzerProjectHandler(config ExternalToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		gl.Log("info", "Executing analyzer.project tool")

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			return nil, fmt.Errorf("project_path parameter is required")
		}

		// Extract optional parameters
		depth := getIntArg(args, "depth", 3)
		includeDeps := getBoolArg(args, "include_dependencies", true)

		// Build request payload
		payload := map[string]interface{}{
			"project_path":         projectPath,
			"depth":                depth,
			"include_dependencies": includeDeps,
		}

		// Call Analyzer API
		result, err := callAnalyzerAPI(ctx, config, "/api/analyze/project", payload)
		if err != nil {
			gl.Log("error", "Analyzer API call failed", err)
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Failed to analyze project: %v", err),
			}, nil
		}

		return result, nil
	}
}

// createAnalyzerSecurityHandler creates handler for analyzer.security tool
func createAnalyzerSecurityHandler(config ExternalToolsConfig) ToolHandler {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		gl.Log("info", "Executing analyzer.security tool")

		// Extract project_path
		projectPath, ok := args["project_path"].(string)
		if !ok || projectPath == "" {
			return nil, fmt.Errorf("project_path parameter is required")
		}

		// Extract optional parameters
		severity := getStringArg(args, "severity_threshold", "medium")

		// Build request payload
		payload := map[string]interface{}{
			"project_path":       projectPath,
			"severity_threshold": severity,
		}

		// Call Analyzer API
		result, err := callAnalyzerAPI(ctx, config, "/api/analyze/security", payload)
		if err != nil {
			gl.Log("error", "Analyzer API call failed", err)
			return map[string]interface{}{
				"status":  "error",
				"message": fmt.Sprintf("Failed to perform security analysis: %v", err),
			}, nil
		}

		return result, nil
	}
}

// callGromptAPI makes HTTP request to Grompt API with BYOK support
func callGromptAPI(ctx context.Context, config ExternalToolsConfig, endpoint string, payload map[string]interface{}, apiKey string) (interface{}, error) {
	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	url := config.GromptURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// BYOK Support: Add API key if provided
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
		gl.Log("debug", "Using external API key (BYOK) for Grompt request")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// callAnalyzerAPI makes HTTP request to Analyzer API
func callAnalyzerAPI(ctx context.Context, config ExternalToolsConfig, endpoint string, payload map[string]interface{}) (interface{}, error) {
	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create request
	url := config.AnalyzerURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// Helper functions to extract arguments with defaults
func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, exists := args[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntArg(args map[string]interface{}, key string, defaultValue int) int {
	if val, exists := args[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			// Try to parse string to int
			var result int
			if _, err := fmt.Sscanf(v, "%d", &result); err == nil {
				return result
			}
		}
	}
	return defaultValue
}

func getBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, exists := args[key]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}
