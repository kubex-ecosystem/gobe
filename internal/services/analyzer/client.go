// Package analyzer provides client integration with the GemX Analyzer service.
package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

// Client represents a client for the GemX Analyzer service
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
}

// NewClient creates a new GemX Analyzer client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScorecardRequest represents a request for repository scorecard analysis
type ScorecardRequest struct {
	RepoURL     string                 `json:"repo_url"`
	Provider    string                 `json:"provider,omitempty"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Webhook     string                 `json:"webhook,omitempty"`
	CallbackURL string                 `json:"callback_url,omitempty"`
}

// ScorecardResponse represents the response from scorecard analysis
type ScorecardResponse struct {
	SchemaVersion       string            `json:"schema_version"`
	Repository          RepositoryInfo    `json:"repository"`
	DORA                DORAMetrics       `json:"dora"`
	CHI                 CHIMetrics        `json:"chi"`
	AI                  AIMetrics         `json:"ai"`
	BusFactor           int               `json:"bus_factor"`
	FirstReviewP50Hours float64           `json:"first_review_p50_hours"`
	Confidence          ConfidenceMetrics `json:"confidence"`
	GeneratedAt         time.Time         `json:"generated_at"`
	JobID               string            `json:"job_id,omitempty"`
	Status              string            `json:"status,omitempty"`
}

// RepositoryInfo represents basic repository information
type RepositoryInfo struct {
	Owner         string    `json:"owner"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	CloneURL      string    `json:"clone_url"`
	DefaultBranch string    `json:"default_branch"`
	Language      string    `json:"language"`
	IsPrivate     bool      `json:"is_private"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DORAMetrics represents DORA metrics
type DORAMetrics struct {
	LeadTimeP95Hours        float64   `json:"lead_time_p95_hours"`
	DeploymentFrequencyWeek float64   `json:"deployment_frequency_per_week"`
	ChangeFailRatePercent   float64   `json:"change_fail_rate_pct"`
	MTTRHours               float64   `json:"mttr_hours"`
	Period                  int       `json:"period_days"`
	CalculatedAt            time.Time `json:"calculated_at"`
}

// CHIMetrics represents Code Health Index metrics
type CHIMetrics struct {
	Score                int       `json:"chi_score"`
	DuplicationPercent   float64   `json:"duplication_pct"`
	CyclomaticComplexity float64   `json:"cyclomatic_avg"`
	TestCoverage         float64   `json:"test_coverage_pct"`
	MaintainabilityIndex float64   `json:"maintainability_index"`
	TechnicalDebt        float64   `json:"technical_debt_hours"`
	Period               int       `json:"period_days"`
	CalculatedAt         time.Time `json:"calculated_at"`
}

// AIMetrics represents AI development metrics
type AIMetrics struct {
	HIR          float64   `json:"hir"` // Human Input Ratio
	AAC          float64   `json:"aac"` // AI Assist Coverage
	TPH          float64   `json:"tph"` // Throughput per Human-hour
	HumanHours   float64   `json:"human_hours"`
	AIHours      float64   `json:"ai_hours"`
	Period       int       `json:"period_days"`
	CalculatedAt time.Time `json:"calculated_at"`
}

// ConfidenceMetrics represents confidence levels for metrics
type ConfidenceMetrics struct {
	DORA  float64 `json:"dora"`
	CHI   float64 `json:"chi"`
	AI    float64 `json:"ai"`
	Group float64 `json:"group"`
}

// AnalysisJob represents an analysis job status
type AnalysisJob struct {
	ID          string                 `json:"id"`
	RepoURL     string                 `json:"repo_url"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Progress    float64                `json:"progress"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Results     map[string]interface{} `json:"results,omitempty"`
}

// GetRepositoryScorecard requests a repository scorecard analysis
func (c *Client) GetRepositoryScorecard(ctx context.Context, req ScorecardRequest) (*ScorecardResponse, error) {
	endpoint := "/api/v1/scorecard"

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respBody, err := c.makeRequest(ctx, "POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var scorecard ScorecardResponse
	if err := json.Unmarshal(respBody, &scorecard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	gl.Log("info", "Repository scorecard obtained", "repo_url", req.RepoURL, "chi_score", scorecard.CHI.Score)
	return &scorecard, nil
}

// GetAnalysisJob gets the status of an analysis job
func (c *Client) GetAnalysisJob(ctx context.Context, jobID string) (*AnalysisJob, error) {
	endpoint := fmt.Sprintf("/api/v1/jobs/%s", jobID)

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var job AnalysisJob
	if err := json.Unmarshal(respBody, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &job, nil
}

// ListProviders lists available AI providers
func (c *Client) ListProviders(ctx context.Context) ([]string, error) {
	endpoint := "/v1/providers"

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var providers []string
	if err := json.Unmarshal(respBody, &providers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return providers, nil
}

// GetHealth checks the health of the analyzer service
func (c *Client) GetHealth(ctx context.Context) (map[string]interface{}, error) {
	endpoint := "/healthz"

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	var health map[string]interface{}
	if err := json.Unmarshal(respBody, &health); err != nil {
		return nil, fmt.Errorf("failed to unmarshal health response: %w", err)
	}

	return health, nil
}

// makeRequest makes an HTTP request to the analyzer service
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, payload []byte) ([]byte, error) {
	url := c.BaseURL + endpoint

	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	// Add user agent
	req.Header.Set("User-Agent", "GoBE-MCP-Analyzer/1.3.5")

	gl.Log("debug", "Making request to analyzer", "method", method, "url", url)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		gl.Log("error", "Analyzer request failed", "status", resp.StatusCode, "response", string(respBody))
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Service represents the analyzer service with configuration
type Service struct {
	client  *Client
	enabled bool
	baseURL string
}

// NewService creates a new analyzer service
func NewService(baseURL, apiKey string) *Service {
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default analyzer URL
	}

	return &Service{
		client:  NewClient(baseURL, apiKey),
		enabled: true,
		baseURL: baseURL,
	}
}

// IsEnabled returns whether the analyzer service is enabled
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// GetClient returns the HTTP client for direct access
func (s *Service) GetClient() *Client {
	return s.client
}

// HealthCheck performs a health check on the analyzer service
func (s *Service) HealthCheck(ctx context.Context) error {
	if !s.enabled {
		return fmt.Errorf("analyzer service is disabled")
	}

	_, err := s.client.GetHealth(ctx)
	if err != nil {
		return fmt.Errorf("analyzer health check failed: %w", err)
	}

	return nil
}
