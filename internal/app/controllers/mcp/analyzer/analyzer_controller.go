// Package analyzer provides the AnalyzerController for handling GemX Analyzer integration operations.
package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"gorm.io/gorm"
)

type AnalyzerController struct {
	dbConn     *gorm.DB
	APIWrapper *t.APIWrapper[any]
}

func NewAnalyzerController(db *gorm.DB) *AnalyzerController {
	if db == nil {
		gl.Log("warn", "Database connection is nil for AnalyzerController")
	}

	return &AnalyzerController{
		dbConn:     db,
		APIWrapper: t.NewAPIWrapper[any](),
	}
}

// Repository Intelligence Analysis Types
type RepositoryIntelligenceRequest struct {
	RepoURL        string                 `json:"repo_url" binding:"required"`
	AnalysisType   string                 `json:"analysis_type" binding:"required"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	ScheduledBy    string                 `json:"scheduled_by,omitempty"`
	NotifyChannels []string               `json:"notify_channels,omitempty"`
}

type AnalysisJob struct {
	ID           string                 `json:"id"`
	RepoURL      string                 `json:"repo_url"`
	AnalysisType string                 `json:"analysis_type"`
	Status       string                 `json:"status"` // "scheduled", "running", "completed", "failed"
	Progress     float64                `json:"progress"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Results      map[string]interface{} `json:"results,omitempty"`
	ScheduledBy  string                 `json:"scheduled_by,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

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
}

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

type DORAMetrics struct {
	LeadTimeP95Hours        float64   `json:"lead_time_p95_hours"`
	DeploymentFrequencyWeek float64   `json:"deployment_frequency_per_week"`
	ChangeFailRatePercent   float64   `json:"change_fail_rate_pct"`
	MTTRHours               float64   `json:"mttr_hours"`
	Period                  int       `json:"period_days"`
	CalculatedAt            time.Time `json:"calculated_at"`
}

type CHIMetrics struct {
	Score                int       `json:"chi_score"` // 0-100
	DuplicationPercent   float64   `json:"duplication_pct"`
	CyclomaticComplexity float64   `json:"cyclomatic_avg"`
	TestCoverage         float64   `json:"test_coverage_pct"`
	MaintainabilityIndex float64   `json:"maintainability_index"`
	TechnicalDebt        float64   `json:"technical_debt_hours"`
	Period               int       `json:"period_days"`
	CalculatedAt         time.Time `json:"calculated_at"`
}

type AIMetrics struct {
	HIR          float64   `json:"hir"` // Human Input Ratio (0.0-1.0)
	AAC          float64   `json:"aac"` // AI Assist Coverage (0.0-1.0)
	TPH          float64   `json:"tph"` // Throughput per Human-hour
	HumanHours   float64   `json:"human_hours"`
	AIHours      float64   `json:"ai_hours"`
	Period       int       `json:"period_days"`
	CalculatedAt time.Time `json:"calculated_at"`
}

type ConfidenceMetrics struct {
	DORA  float64 `json:"dora"`  // 0.0-1.0
	CHI   float64 `json:"chi"`   // 0.0-1.0
	AI    float64 `json:"ai"`    // 0.0-1.0
	Group float64 `json:"group"` // Overall confidence
}

type NotificationRequest struct {
	Type        string                 `json:"type"` // "discord", "email", "webhook"
	Recipients  []string               `json:"recipients"`
	Subject     string                 `json:"subject"`
	Message     string                 `json:"message"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Priority    string                 `json:"priority,omitempty"` // "low", "normal", "high", "urgent"
	AttachFiles []string               `json:"attach_files,omitempty"`
}

// ScheduleAnalysis schedules a repository analysis in the GemX Analyzer system
//
// @Summary     Schedule Repository Analysis
// @Description Schedule intelligent analysis of a repository using GemX Analyzer
// @Tags        analyzer
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       request body RepositoryIntelligenceRequest true "Analysis request"
// @Success     201 {object} AnalysisJob
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/mcp/analyzer/schedule [post]
func (ac *AnalyzerController) ScheduleAnalysis(c *gin.Context) {
	var req RepositoryIntelligenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gl.Log("error", "Failed to bind analysis request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate analysis type
	validTypes := []string{"scorecard", "dora", "chi", "community", "security", "full"}
	isValid := false
	for _, validType := range validTypes {
		if req.AnalysisType == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid analysis type",
			"valid_types": validTypes,
		})
		return
	}

	// Create analysis job
	job := AnalysisJob{
		ID:           uuid.New().String(),
		RepoURL:      req.RepoURL,
		AnalysisType: req.AnalysisType,
		Status:       "scheduled",
		Progress:     0.0,
		CreatedAt:    time.Now(),
		ScheduledBy:  req.ScheduledBy,
		Metadata: map[string]interface{}{
			"configuration":   req.Configuration,
			"notify_channels": req.NotifyChannels,
			"source":          "gobe_mcp",
		},
	}

	gl.Log("info", "Repository analysis scheduled",
		"job_id", job.ID,
		"repo_url", job.RepoURL,
		"analysis_type", job.AnalysisType,
		"scheduled_by", job.ScheduledBy)

	// In a real implementation, this would integrate with the GemX Analyzer
	// For now, we simulate the scheduling and return the job
	// TODO: Integrate with actual GemX Analyzer service

	c.JSON(http.StatusCreated, job)
}

// GetAnalysisStatus retrieves the status of a specific analysis job
//
// @Summary     Get Analysis Status
// @Description Get the current status and progress of a repository analysis job
// @Tags        analyzer
// @Security    BearerAuth
// @Produce     json
// @Param       job_id path string true "Analysis job ID"
// @Success     200 {object} AnalysisJob
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/mcp/analyzer/status/{job_id} [get]
func (ac *AnalyzerController) GetAnalysisStatus(c *gin.Context) {
	jobID := c.Param("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// TODO: Implement actual job status retrieval from GemX Analyzer
	// For now, return a mock response
	job := AnalysisJob{
		ID:           jobID,
		RepoURL:      "https://github.com/example/repo",
		AnalysisType: "scorecard",
		Status:       "completed",
		Progress:     100.0,
		CreatedAt:    time.Now().Add(-30 * time.Minute),
		StartedAt:    func() *time.Time { t := time.Now().Add(-25 * time.Minute); return &t }(),
		CompletedAt:  func() *time.Time { t := time.Now().Add(-5 * time.Minute); return &t }(),
		ScheduledBy:  "system",
		Results: map[string]interface{}{
			"scorecard_url": fmt.Sprintf("/api/v1/mcp/analyzer/results/%s/scorecard", jobID),
			"summary": map[string]interface{}{
				"chi_score":      85,
				"dora_grade":     "B",
				"bus_factor":     3,
				"confidence":     0.92,
			},
		},
	}

	gl.Log("info", "Analysis status retrieved", "job_id", jobID, "status", job.Status)

	c.JSON(http.StatusOK, job)
}

// GetAnalysisResults retrieves detailed results of a completed analysis
//
// @Summary     Get Analysis Results
// @Description Get detailed results including scorecard, DORA metrics, and recommendations
// @Tags        analyzer
// @Security    BearerAuth
// @Produce     json
// @Param       job_id path string true "Analysis job ID"
// @Param       format query string false "Result format" Enums(scorecard, dora, chi, community)
// @Success     200 {object} ScorecardResponse
// @Failure     400 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/mcp/analyzer/results/{job_id} [get]
func (ac *AnalyzerController) GetAnalysisResults(c *gin.Context) {
	jobID := c.Param("job_id")
	format := c.DefaultQuery("format", "scorecard")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// TODO: Implement actual results retrieval from GemX Analyzer
	// For now, return mock comprehensive results
	switch format {
	case "scorecard":
		scorecard := ScorecardResponse{
			SchemaVersion: "1.0.0",
			Repository: RepositoryInfo{
				Owner:         "example",
				Name:          "repo",
				FullName:      "example/repo",
				CloneURL:      "https://github.com/example/repo.git",
				DefaultBranch: "main",
				Language:      "Go",
				IsPrivate:     false,
				CreatedAt:     time.Now().Add(-365 * 24 * time.Hour),
				UpdatedAt:     time.Now().Add(-2 * time.Hour),
			},
			DORA: DORAMetrics{
				LeadTimeP95Hours:        48.5,
				DeploymentFrequencyWeek: 3.2,
				ChangeFailRatePercent:   8.5,
				MTTRHours:               2.1,
				Period:                  30,
				CalculatedAt:            time.Now(),
			},
			CHI: CHIMetrics{
				Score:                85,
				DuplicationPercent:   5.2,
				CyclomaticComplexity: 3.8,
				TestCoverage:         78.5,
				MaintainabilityIndex: 82.3,
				TechnicalDebt:        24.5,
				Period:               30,
				CalculatedAt:         time.Now(),
			},
			AI: AIMetrics{
				HIR:          0.75,
				AAC:          0.35,
				TPH:          8.2,
				HumanHours:   120.5,
				AIHours:      42.3,
				Period:       30,
				CalculatedAt: time.Now(),
			},
			BusFactor:           3,
			FirstReviewP50Hours: 4.2,
			Confidence: ConfidenceMetrics{
				DORA:  0.92,
				CHI:   0.89,
				AI:    0.76,
				Group: 0.86,
			},
			GeneratedAt: time.Now(),
		}

		gl.Log("info", "Scorecard results retrieved", "job_id", jobID, "chi_score", scorecard.CHI.Score)
		c.JSON(http.StatusOK, scorecard)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "Invalid format",
			"valid_formats":  []string{"scorecard", "dora", "chi", "community"},
		})
	}
}

// ListAnalysisJobs lists analysis jobs with optional filtering
//
// @Summary     List Analysis Jobs
// @Description List repository analysis jobs with pagination and filtering
// @Tags        analyzer
// @Security    BearerAuth
// @Produce     json
// @Param       status query string false "Filter by status" Enums(scheduled, running, completed, failed)
// @Param       repo_url query string false "Filter by repository URL"
// @Param       limit query int false "Number of results to return (default: 50)"
// @Param       offset query int false "Number of results to skip (default: 0)"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} ErrorResponse
// @Router      /api/v1/mcp/analyzer/jobs [get]
func (ac *AnalyzerController) ListAnalysisJobs(c *gin.Context) {
	status := c.Query("status")
	repoURL := c.Query("repo_url")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// TODO: Implement actual job listing from database/GemX Analyzer
	// For now, return mock data
	mockJobs := []AnalysisJob{
		{
			ID:           "job-001",
			RepoURL:      "https://github.com/example/repo1",
			AnalysisType: "scorecard",
			Status:       "completed",
			Progress:     100.0,
			CreatedAt:    time.Now().Add(-2 * time.Hour),
			CompletedAt:  func() *time.Time { t := time.Now().Add(-30 * time.Minute); return &t }(),
			ScheduledBy:  "user@example.com",
		},
		{
			ID:           "job-002",
			RepoURL:      "https://github.com/example/repo2",
			AnalysisType: "dora",
			Status:       "running",
			Progress:     65.0,
			CreatedAt:    time.Now().Add(-1 * time.Hour),
			StartedAt:    func() *time.Time { t := time.Now().Add(-45 * time.Minute); return &t }(),
			ScheduledBy:  "admin",
		},
	}

	// Apply filters
	filteredJobs := make([]AnalysisJob, 0)
	for _, job := range mockJobs {
		if status != "" && job.Status != status {
			continue
		}
		if repoURL != "" && job.RepoURL != repoURL {
			continue
		}
		filteredJobs = append(filteredJobs, job)
	}

	// Apply pagination
	total := len(filteredJobs)
	if offset >= total {
		filteredJobs = []AnalysisJob{}
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		filteredJobs = filteredJobs[offset:end]
	}

	gl.Log("info", "Analysis jobs listed", "total", total, "returned", len(filteredJobs))

	response := map[string]interface{}{
		"jobs":   filteredJobs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	c.JSON(http.StatusOK, response)
}

// SendNotification sends notifications through the GoBE notification system
//
// @Summary     Send Analysis Notification
// @Description Send notification about analysis completion or status updates
// @Tags        analyzer
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       notification body NotificationRequest true "Notification details"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/analyzer/notifications/send [post]
func (ac *AnalyzerController) SendNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		gl.Log("error", "Failed to bind notification request", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification format"})
		return
	}

	// Validate notification type
	validTypes := []string{"discord", "email", "webhook"}
	isValid := false
	for _, validType := range validTypes {
		if req.Type == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid notification type",
			"valid_types": validTypes,
		})
		return
	}

	// TODO: Implement actual notification sending
	// This would integrate with Discord webhooks, email service, etc.

	gl.Log("info", "Notification sent",
		"type", req.Type,
		"recipients", len(req.Recipients),
		"subject", req.Subject,
		"priority", req.Priority)

	response := map[string]interface{}{
		"status":      "sent",
		"type":        req.Type,
		"recipients":  len(req.Recipients),
		"message_id":  uuid.New().String(),
		"sent_at":     time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// GetSystemHealth returns analyzer system health status
//
// @Summary     Get Analyzer Health
// @Description Get health status of the GemX Analyzer integration
// @Tags        analyzer
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /api/v1/mcp/analyzer/health [get]
func (ac *AnalyzerController) GetSystemHealth(c *gin.Context) {
	health := map[string]interface{}{
		"status":      "healthy",
		"timestamp":   time.Now(),
		"version":     "1.0.0",
		"integration": "gemx-analyzer",
		"capabilities": []string{
			"repository_analysis",
			"dora_metrics",
			"code_health",
			"ai_metrics",
			"executive_reports",
			"notifications",
		},
		"supported_analysis_types": []string{
			"scorecard", "dora", "chi", "community", "security", "full",
		},
		"system_metrics": map[string]interface{}{
			"active_jobs":      0,
			"completed_jobs":   125,
			"failed_jobs":      3,
			"average_duration": "12.5 minutes",
			"success_rate":     "97.6%",
		},
	}

	gl.Log("info", "Analyzer system health retrieved")
	c.JSON(http.StatusOK, health)
}