// Package gobe provides integration with the GoBE backend system
package gobe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a GoBE backend client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// Config holds GoBE client configuration
type Config struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Timeout int    `json:"timeout"`
}

// NewClient creates a new GoBE client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// UserRequest represents a user creation request
type UserRequest struct {
	Name     string                 `json:"name"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role"`
	Source   string                 `json:"source"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Role    string    `json:"role"`
	Created time.Time `json:"created"`
	Status  string    `json:"status"`
}

// SystemStatus represents system status response
type SystemStatus struct {
	Status   string            `json:"status"`
	Version  string            `json:"version"`
	Uptime   string            `json:"uptime"`
	Database DatabaseStatus    `json:"database"`
	Services map[string]string `json:"services"`
	Metrics  SystemMetrics     `json:"metrics"`
}

type DatabaseStatus struct {
	Connected bool   `json:"connected"`
	Latency   string `json:"latency"`
	Pool      string `json:"pool"`
}

type SystemMetrics struct {
	RequestsTotal  int64   `json:"requests_total"`
	ErrorRate      float64 `json:"error_rate"`
	ResponseTime   string  `json:"response_time"`
	ActiveSessions int     `json:"active_sessions"`
}

// CreateUser creates a new user in GoBE
func (c *Client) CreateUser(ctx context.Context, req UserRequest) (*UserResponse, error) {
	// Add source metadata for Discord integration
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Source = "discord-mcp-hub"
	req.Metadata["created_via"] = "discord_bot"
	req.Metadata["timestamp"] = time.Now().Unix()

	url := fmt.Sprintf("%s/api/v1/users", c.baseURL)

	var result UserResponse
	err := c.doRequest(ctx, "POST", url, req, &result)
	return &result, err
}

// GetUser retrieves user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	var result UserResponse
	err := c.doRequest(ctx, "GET", url, nil, &result)
	return &result, err
}

// GetSystemStatus retrieves GoBE system status
func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)

	var result SystemStatus
	err := c.doRequest(ctx, "GET", url, nil, &result)
	return &result, err
}

// ExecuteQuery executes a custom query/command in GoBE
func (c *Client) ExecuteQuery(ctx context.Context, query string, params map[string]interface{}) (map[string]interface{}, error) {
	queryReq := map[string]interface{}{
		"query":  query,
		"params": params,
		"source": "discord-mcp-hub",
	}

	url := fmt.Sprintf("%s/api/v1/execute", c.baseURL)

	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", url, queryReq, &result)
	return result, err
}

// BackupDatabase triggers a database backup
func (c *Client) BackupDatabase(ctx context.Context) (map[string]interface{}, error) {
	backupReq := map[string]interface{}{
		"type":      "full",
		"source":    "discord-mcp-hub",
		"timestamp": time.Now().Unix(),
	}

	url := fmt.Sprintf("%s/api/v1/backup", c.baseURL)

	var result map[string]interface{}
	err := c.doRequest(ctx, "POST", url, backupReq, &result)
	return result, err
}

// Generic HTTP request method
func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}, target interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Discord-MCP-Hub/1.0")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if target != nil {
		if err := json.Unmarshal(respBody, target); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Ping tests connectivity to GoBE
func (c *Client) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/ping", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed with status: %d", resp.StatusCode)
	}

	return nil
}
