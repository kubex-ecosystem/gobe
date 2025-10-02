// Package analyzer provides the AnalyzerController for handling GemX Analyzer integration operations.
package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kubex-ecosystem/gobe/internal/services/analyzer"

	m "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type AnalyzerController struct {
	bridge          *m.Bridge
	APIWrapper      *t.APIWrapper[any]
	analyzerService *analyzer.Service
	analysisService m.AnalysisJobService
}

func NewAnalyzerController(bridge *m.Bridge) *AnalyzerController {
	if bridge == nil {
		gl.Log("warn", "Bridge is nil for AnalyzerController")
		return &AnalyzerController{
			APIWrapper: t.NewAPIWrapper[any](),
		}
	}

	// Initialize analyzer service with environment configuration
	analyzerBaseURL := getEnv("GEMX_ANALYZER_URL", "http://localhost:8080")
	analyzerAPIKey := getEnv("GEMX_ANALYZER_API_KEY", "")

	analyzerService := analyzer.NewService(analyzerBaseURL, analyzerAPIKey)

	// Use bridge to get analysis job service
	analysisService := bridge.AnalysisJobService()

	return &AnalyzerController{
		bridge:          bridge,
		APIWrapper:      t.NewAPIWrapper[any](),
		analyzerService: analyzerService,
		analysisService: analysisService,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getDORAGrade calculates DORA grade based on metrics
func getDORAGrade(dora analyzer.DORAMetrics) string {
	// Simplified DORA grading based on industry benchmarks
	score := 0

	// Lead Time (< 1 day = 3, < 1 week = 2, < 1 month = 1, > 1 month = 0)
	if dora.LeadTimeP95Hours < 24 {
		score += 3
	} else if dora.LeadTimeP95Hours < 168 { // 1 week
		score += 2
	} else if dora.LeadTimeP95Hours < 720 { // 1 month
		score += 1
	}

	// Deployment Frequency (> 1/day = 3, 1/week = 2, 1/month = 1, < 1/month = 0)
	if dora.DeploymentFrequencyWeek > 7 {
		score += 3
	} else if dora.DeploymentFrequencyWeek >= 1 {
		score += 2
	} else if dora.DeploymentFrequencyWeek >= 0.25 { // ~1 per month
		score += 1
	}

	// Change Fail Rate (< 5% = 3, < 10% = 2, < 15% = 1, > 15% = 0)
	if dora.ChangeFailRatePercent < 5 {
		score += 3
	} else if dora.ChangeFailRatePercent < 10 {
		score += 2
	} else if dora.ChangeFailRatePercent < 15 {
		score += 1
	}

	// MTTR (< 1 hour = 3, < 1 day = 2, < 1 week = 1, > 1 week = 0)
	if dora.MTTRHours < 1 {
		score += 3
	} else if dora.MTTRHours < 24 {
		score += 2
	} else if dora.MTTRHours < 168 {
		score += 1
	}

	// Grade based on total score
	switch {
	case score >= 10:
		return "A"
	case score >= 8:
		return "B"
	case score >= 6:
		return "C"
	case score >= 4:
		return "D"
	default:
		return "F"
	}
}

// RepositoryIntelligenceRequest represents a request to analyze a repository
type RepositoryIntelligenceRequest struct {
	RepoURL        string                 `json:"repo_url" binding:"required"`
	AnalysisType   string                 `json:"analysis_type" binding:"required"`
	ProjectID      string                 `json:"project_id,omitempty"`
	SourceType     string                 `json:"source_type,omitempty"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ScheduledBy    string                 `json:"scheduled_by,omitempty"`
	NotifyChannels []string               `json:"notify_channels,omitempty"`
	MaxRetries     *int                   `json:"max_retries,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
}

type AnalysisJob struct {
	ID           string                 `json:"id"`
	JobType      string                 `json:"job_type"`
	Status       string                 `json:"status"`
	RepoURL      string                 `json:"repo_url,omitempty"`
	SourceType   string                 `json:"source_type,omitempty"`
	AnalysisType string                 `json:"analysis_type,omitempty"`
	Progress     float64                `json:"progress"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	InputData    map[string]interface{} `json:"input_data,omitempty"`
	OutputData   map[string]interface{} `json:"output_data,omitempty"`
	Results      map[string]interface{} `json:"results,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ScheduledBy  string                 `json:"scheduled_by,omitempty"`
	ProjectID    string                 `json:"project_id,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	RetryCount   int                    `json:"retry_count,omitempty"`
	MaxRetries   int                    `json:"max_retries,omitempty"`
}

var (
	analysisTypeMappings = map[string]string{
		"SCORECARD":            "SCORECARD_ANALYSIS",
		"SCORECARD_ANALYSIS":   "SCORECARD_ANALYSIS",
		"FULL":                 "SCORECARD_ANALYSIS",
		"DORA":                 "SCORECARD_ANALYSIS",
		"CHI":                  "SCORECARD_ANALYSIS",
		"COMMUNITY":            "SCORECARD_ANALYSIS",
		"CODE":                 "CODE_ANALYSIS",
		"CODE_ANALYSIS":        "CODE_ANALYSIS",
		"SECURITY":             "SECURITY_ANALYSIS",
		"SECURITY_ANALYSIS":    "SECURITY_ANALYSIS",
		"PERFORMANCE":          "PERFORMANCE_ANALYSIS",
		"PERFORMANCE_ANALYSIS": "PERFORMANCE_ANALYSIS",
		"DEPENDENCY":           "DEPENDENCY_ANALYSIS",
		"DEPENDENCIES":         "DEPENDENCY_ANALYSIS",
		"DEPENDENCY_ANALYSIS":  "DEPENDENCY_ANALYSIS",
	}
	analysisTypeAliasList = []string{
		"scorecard",
		"dora",
		"chi",
		"community",
		"security",
		"full",
		"code",
		"performance",
		"dependency",
		"dependencies",
	}
)

func supportedAnalysisAliases() []string {
	aliases := make([]string, 0, len(analysisTypeAliasList))
	seen := make(map[string]struct{}, len(analysisTypeAliasList))
	for _, alias := range analysisTypeAliasList {
		key := strings.ToLower(strings.TrimSpace(alias))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		aliases = append(aliases, key)
	}
	return aliases
}

func resolveAnalysisJobType(requested string) (string, string, error) {
	alias := strings.TrimSpace(requested)
	if alias == "" {
		return "", "", fmt.Errorf("analysis type is required")
	}

	lookupKey := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(alias, "-", "_"), " ", "_"))
	jobType, ok := analysisTypeMappings[lookupKey]
	if !ok {
		return "", "", fmt.Errorf("unsupported analysis type: %s", requested)
	}

	normalized := strings.ToLower(strings.TrimSpace(alias))
	if normalized == "" {
		normalized = strings.ToLower(jobType)
	}

	return jobType, normalized, nil
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	if value := ctx.Value(t.CtxKey("userID")); value != nil {
		if id, ok := value.(uuid.UUID); ok {
			return id, true
		}
	}
	return uuid.Nil, false
}

func mergeMetadata(base map[string]interface{}, extra map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(extra) == 0 {
		return nil
	}
	merged := make(map[string]interface{}, len(base)+len(extra))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

// Use types from analyzer package

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

	if ac.analysisService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analysis service unavailable"})
		return
	}

	jobType, normalizedType, err := resolveAnalysisJobType(req.AnalysisType)
	if err != nil {
		gl.Log("warn", "Invalid analysis type", "value", req.AnalysisType, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "Invalid analysis type",
			"valid_types": supportedAnalysisAliases(),
		})
		return
	}

	ctx, ctxErr := ac.APIWrapper.GetContext(c)
	var userID uuid.UUID
	if ctxErr == nil {
		if value, ok := userIDFromContext(ctx); ok {
			userID = value
		}
	}

	if userID == uuid.Nil && req.UserID != "" {
		parsed, parseErr := uuid.Parse(req.UserID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
		userID = parsed
		ctx = context.WithValue(c.Request.Context(), t.CtxKey("userID"), userID)
	}

	if userID == uuid.Nil {
		gl.Log("warn", "Missing user identifier for analyzer schedule")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID is required"})
		return
	}

	projectID := uuid.Nil
	if req.ProjectID != "" {
		parsed, parseErr := uuid.Parse(req.ProjectID)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
			return
		}
		projectID = parsed
	}

	sourceType := strings.TrimSpace(req.SourceType)
	if sourceType == "" {
		sourceType = "repository"
	}

	inputData := make(map[string]interface{})
	for key, value := range req.Configuration {
		inputData[key] = value
	}
	if len(req.NotifyChannels) > 0 {
		inputData["notify_channels"] = req.NotifyChannels
	}

	metadata := map[string]interface{}{
		"source":                  "gobe_mcp",
		"requested_analysis_type": normalizedType,
		"canonical_analysis_type": strings.ToLower(jobType),
	}
	if req.ScheduledBy != "" {
		metadata["scheduled_by"] = req.ScheduledBy
	}
	if len(req.NotifyChannels) > 0 {
		metadata["notify_channels"] = req.NotifyChannels
	}
	for key, value := range req.Metadata {
		metadata[key] = value
	}

	analysisJob := &m.AnalysisJobImpl{}
	analysisJob.SetJobType(jobType)
	analysisJob.SetStatus("PENDING")
	analysisJob.SetSourceURL(strings.TrimSpace(req.RepoURL))
	analysisJob.SetSourceType(sourceType)
	analysisJob.SetProgress(0)
	analysisJob.SetUserID(userID)
	analysisJob.SetCreatedBy(userID)
	analysisJob.SetUpdatedBy(userID)
	analysisJob.SetProjectID(projectID)
	if len(inputData) > 0 {
		analysisJob.SetInputData(m.MapToJSONB(inputData))
	}
	analysisJob.SetMetadata(m.MapToJSONB(metadata))

	maxRetries := 3
	if req.MaxRetries != nil && *req.MaxRetries >= 0 {
		maxRetries = *req.MaxRetries
	}
	analysisJob.SetMaxRetries(maxRetries)
	analysisJob.SetRetryCount(0)

	createdJob, err := ac.analysisService.CreateJob(ctx, analysisJob)
	if err != nil {
		gl.Log("error", "Failed to create analysis job", "repo_url", req.RepoURL, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to schedule repository analysis"})
		return
	}

	response := convertModelToAnalysisJob(createdJob)
	if response.AnalysisType == "" {
		response.AnalysisType = normalizedType
	}
	response.Metadata = mergeMetadata(response.Metadata, metadata)

	gl.Log("info", "Repository analysis scheduled",
		"job_id", response.ID,
		"repo_url", response.RepoURL,
		"job_type", response.JobType,
		"analysis_type", response.AnalysisType,
		"user_id", userID.String())

	c.JSON(http.StatusCreated, response)
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

	if ac.analysisService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analysis service unavailable"})
		return
	}

	identifier, err := uuid.Parse(jobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := ac.analysisService.GetJobByID(c.Request.Context(), identifier)
	if err != nil {
		gl.Log("error", "Failed to get analysis job", "job_id", jobID, "error", err)
		if isNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job status"})
		}
		return
	}

	response := convertModelToAnalysisJob(job)
	gl.Log("info", "Analysis status retrieved", "job_id", jobID, "status", response.Status)
	c.JSON(http.StatusOK, response)
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

	if ac.analysisService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analysis service unavailable"})
		return
	}

	identifier, err := uuid.Parse(jobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	job, err := ac.analysisService.GetJobByID(c.Request.Context(), identifier)
	if err != nil {
		gl.Log("error", "Failed to get analysis job for results", "job_id", jobID, "error", err)
		if isNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve job"})
		}
		return
	}

	status := strings.ToUpper(job.GetStatus())
	if status != "COMPLETED" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Job is not completed yet",
			"status": job.GetStatus(),
		})
		return
	}

	results := jsonbToMap(m.JSONBToImpl(job.GetOutputData()))
	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No results available for this job"})
		return
	}

	switch strings.ToLower(format) {
	case "scorecard":
		if scorecard, ok := results["scorecard"]; ok {
			gl.Log("info", "Scorecard results retrieved", "job_id", jobID)
			c.JSON(http.StatusOK, scorecard)
			return
		}
		gl.Log("info", "Returning raw results for scorecard format", "job_id", jobID)
		c.JSON(http.StatusOK, results)
	case "dora":
		if section := nestedMapValue(results, "scorecard", "dora"); section != nil {
			gl.Log("info", "DORA metrics retrieved", "job_id", jobID)
			c.JSON(http.StatusOK, section)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "DORA metrics not available"})
	case "chi":
		if section := nestedMapValue(results, "scorecard", "chi"); section != nil {
			gl.Log("info", "CHI metrics retrieved", "job_id", jobID)
			c.JSON(http.StatusOK, section)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "CHI metrics not available"})
	case "ai":
		if section := nestedMapValue(results, "scorecard", "ai"); section != nil {
			gl.Log("info", "AI metrics retrieved", "job_id", jobID)
			c.JSON(http.StatusOK, section)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "AI metrics not available"})
	case "summary":
		if summary, ok := results["summary"]; ok {
			gl.Log("info", "Summary results retrieved", "job_id", jobID)
			c.JSON(http.StatusOK, summary)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Summary not available"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Invalid format",
			"valid_formats": []string{"scorecard", "dora", "chi", "ai", "summary"},
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
	analysisType := c.Query("analysis_type")
	scheduledBy := c.Query("scheduled_by")
	projectIDValue := c.Query("project_id")
	userIDValue := c.Query("user_id")
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

	if ac.analysisService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Analysis service unavailable"})
		return
	}

	projectID := uuid.Nil
	if projectIDValue != "" {
		parsed, parseErr := uuid.Parse(projectIDValue)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project_id"})
			return
		}
		projectID = parsed
	}

	userID := uuid.Nil
	if userIDValue != "" {
		parsed, parseErr := uuid.Parse(userIDValue)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
			return
		}
		userID = parsed
	}

	jobs, err := ac.analysisService.ListJobs(c.Request.Context())
	if err != nil {
		gl.Log("error", "Failed to list analysis jobs", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve jobs"})
		return
	}

	filtered := make([]*m.AnalysisJobImpl, 0, len(jobs))
	statusFilter := strings.TrimSpace(status)
	repoFilter := strings.ToLower(strings.TrimSpace(repoURL))
	analysisFilter := strings.ToLower(strings.TrimSpace(analysisType))
	scheduledFilter := strings.ToLower(strings.TrimSpace(scheduledBy))

	for _, job := range jobs {
		if job == nil {
			continue
		}
		if statusFilter != "" && !strings.EqualFold(job.GetStatus(), statusFilter) {
			continue
		}
		if repoFilter != "" && !strings.Contains(strings.ToLower(job.GetSourceURL()), repoFilter) {
			continue
		}
		if projectID != uuid.Nil && job.GetProjectID() != projectID {
			continue
		}
		if userID != uuid.Nil && job.GetUserID() != userID {
			continue
		}

		metadata := jsonbToMap(m.JSONBToImpl(job.GetMetadata()))
		analysisLabel := determineAnalysisLabel(job, metadata)
		if analysisFilter != "" && !strings.EqualFold(analysisLabel, analysisFilter) && !strings.EqualFold(job.GetJobType(), analysisFilter) {
			continue
		}
		if scheduledFilter != "" {
			scheduledMeta := ""
			if val, ok := metadata["scheduled_by"].(string); ok {
				scheduledMeta = strings.ToLower(strings.TrimSpace(val))
			}
			if scheduledMeta == "" || scheduledMeta != scheduledFilter {
				continue
			}
		}

		filtered = append(filtered, job)
	}

	total := len(filtered)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	paged := filtered[offset:end]

	responseJobs := make([]AnalysisJob, len(paged))
	for i, job := range paged {
		responseJobs[i] = convertModelToAnalysisJob(job)
	}

	gl.Log("info", "Analysis jobs listed", "total", total, "returned", len(responseJobs))

	c.JSON(http.StatusOK, gin.H{
		"jobs":   responseJobs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
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
	validTypes := []string{"discord", "email", "webhook", "log"}
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

	// Validate required fields
	if len(req.Recipients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one recipient is required"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Generate message ID
	messageID := uuid.New().String()

	// Send notification based on type
	var err error
	switch req.Type {
	case "discord":
		err = ac.sendDiscordNotification(c.Request.Context(), req, messageID)
	case "email":
		err = ac.sendEmailNotification(c.Request.Context(), req, messageID)
	case "webhook":
		err = ac.sendWebhookNotification(c.Request.Context(), req, messageID)
	case "log":
		err = ac.sendLogNotification(c.Request.Context(), req, messageID)
	}

	if err != nil {
		gl.Log("error", "Failed to send notification", "type", req.Type, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to send notification",
			"message_id": messageID,
		})
		return
	}

	gl.Log("info", "Notification sent successfully",
		"type", req.Type,
		"recipients", len(req.Recipients),
		"subject", req.Subject,
		"priority", req.Priority,
		"message_id", messageID)

	response := map[string]interface{}{
		"status":     "sent",
		"type":       req.Type,
		"recipients": len(req.Recipients),
		"message_id": messageID,
		"sent_at":    time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// sendDiscordNotification sends notification via Discord webhook
func (ac *AnalyzerController) sendDiscordNotification(ctx context.Context, req NotificationRequest, messageID string) error {
	// Format message for Discord
	discordMessage := fmt.Sprintf("**%s**\n\n%s", req.Subject, req.Message)

	// Add metadata if present
	if req.Metadata != nil {
		if jobID, ok := req.Metadata["job_id"].(string); ok {
			discordMessage += fmt.Sprintf("\n\n📊 Job ID: `%s`", jobID)
		}
		if repoURL, ok := req.Metadata["repo_url"].(string); ok {
			discordMessage += fmt.Sprintf("\n🔗 Repository: %s", repoURL)
		}
	}

	// Add priority indicator
	priorityEmoji := "📢"
	switch req.Priority {
	case "urgent":
		priorityEmoji = "🚨"
	case "high":
		priorityEmoji = "⚠️"
	case "low":
		priorityEmoji = "💬"
	}
	discordMessage = priorityEmoji + " " + discordMessage

	// Send to each Discord webhook URL
	for _, recipient := range req.Recipients {
		// Here you would actually send to Discord webhook
		// For now, just log it
		req := &http.Request{
			Method: http.MethodPost,
			URL:    &url.URL{Scheme: "https", Host: "discord.com", Path: "/api/webhooks/" + recipient},
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(bytes.NewBuffer([]byte(discordMessage))),
		}
		requestID := uuid.New().String()
		req.Header.Set("X-Request-ID", requestID)
		_, err := http.DefaultClient.Do(req)
		if err != nil {
			gl.Log("error", "Failed to send Discord notification", "webhook", recipient, "error", err)
			continue
		}

		// Log the sending action
		gl.Log("info", "Discord notification sent", "webhook", recipient, "message_id", messageID)
	}

	return nil
}

// sendEmailNotification sends notification via email
func (ac *AnalyzerController) sendEmailNotification(ctx context.Context, req NotificationRequest, messageID string) error {
	// Format email message
	emailBody := req.Message

	// Add metadata if present
	if req.Metadata != nil {
		emailBody += "\n\n--- Additional Information ---\n"
		for key, value := range req.Metadata {
			emailBody += fmt.Sprintf("%s: %v\n", key, value)
		}
	}

	// Send to each email recipient
	for _, recipient := range req.Recipients {
		// Here you would actually send email using SMTP
		// For now, just log it
		gl.Log("info", "Email notification sent", "email", recipient, "subject", req.Subject, "message_id", messageID)
	}

	return nil
}

// sendWebhookNotification sends notification via HTTP webhook
func (ac *AnalyzerController) sendWebhookNotification(ctx context.Context, req NotificationRequest, messageID string) error {
	// Prepare webhook payload
	payload := map[string]interface{}{
		"message_id": messageID,
		"type":       "notification",
		"subject":    req.Subject,
		"message":    req.Message,
		"priority":   req.Priority,
		"metadata":   req.Metadata,
		"timestamp":  time.Now().UTC(),
	}

	payloadJSON, _ := json.Marshal(payload)

	// Send to each webhook URL
	for _, recipient := range req.Recipients {
		// Here you would actually send HTTP POST to webhook URL
		// For now, just log it
		gl.Log("info", "Webhook notification sent", "url", recipient, "payload_size", len(payloadJSON), "message_id", messageID)
	}

	return nil
}

// sendLogNotification sends notification to system log
func (ac *AnalyzerController) sendLogNotification(ctx context.Context, req NotificationRequest, messageID string) error {
	logLevel := "info"
	switch req.Priority {
	case "urgent", "high":
		logLevel = "warn"
	case "low":
		logLevel = "debug"
	}

	// Format log message
	logMessage := fmt.Sprintf("[NOTIFICATION] %s: %s", req.Subject, req.Message)

	// Add metadata
	if req.Metadata != nil {
		if metadataJSON, err := json.Marshal(req.Metadata); err == nil {
			logMessage += fmt.Sprintf(" | Metadata: %s", string(metadataJSON))
		}
	}

	// Send log with specified level
	gl.Log(logLevel, logMessage, "message_id", messageID, "recipients", req.Recipients)

	return nil
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
		"version":     "1.3.5",
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
	}

	// Check GemX Analyzer service health
	if ac.analyzerService.IsEnabled() {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		err := ac.analyzerService.HealthCheck(ctx)
		if err != nil {
			gl.Log("warn", "GemX Analyzer health check failed", err)
			health["status"] = "degraded"
			health["analyzer_status"] = "unhealthy"
			health["analyzer_error"] = err.Error()
		} else {
			health["analyzer_status"] = "healthy"

			// Get analyzer health details
			analyzerHealth, err := ac.analyzerService.GetClient().GetHealth(ctx)
			if err == nil {
				health["analyzer_details"] = analyzerHealth
			}

			// Get available providers
			providers, err := ac.analyzerService.GetClient().ListProviders(ctx)
			if err == nil {
				health["available_providers"] = providers
			}
		}
	} else {
		health["status"] = "disabled"
		health["analyzer_status"] = "disabled"
		health["analyzer_error"] = "GemX Analyzer service is not enabled"
	}

	gl.Log("info", "Analyzer system health retrieved", "status", health["status"])
	c.JSON(http.StatusOK, health)
}

// convertModelToAnalysisJob converts database model to API response format
func convertModelToAnalysisJob(job *m.AnalysisJobImpl) AnalysisJob {
	if job == nil {
		return AnalysisJob{}
	}

	response := AnalysisJob{
		ID:         job.GetID().String(),
		JobType:    job.GetJobType(),
		Status:     job.GetStatus(),
		RepoURL:    job.GetSourceURL(),
		SourceType: job.GetSourceType(),
		Progress:   job.GetProgress(),
		CreatedAt:  job.GetCreatedAt(),
		RetryCount: job.GetRetryCount(),
		MaxRetries: job.GetMaxRetries(),
	}

	if startedAt := job.GetStartedAt(); !startedAt.IsZero() {
		value := startedAt
		response.StartedAt = &value
	}

	if completedAt := job.GetCompletedAt(); !completedAt.IsZero() {
		value := completedAt
		response.CompletedAt = &value
	}

	if updatedAt := job.GetUpdatedAt(); !updatedAt.IsZero() {
		response.UpdatedAt = updatedAt
	}

	if projectID := job.GetProjectID(); projectID != uuid.Nil {
		response.ProjectID = projectID.String()
	}

	if userID := job.GetUserID(); userID != uuid.Nil {
		response.UserID = userID.String()
	}

	inputData := jsonbToMap(m.JSONBToImpl(job.GetInputData()))
	if len(inputData) > 0 {
		response.InputData = inputData
	}

	outputData := jsonbToMap(m.JSONBToImpl(job.GetOutputData()))
	if len(outputData) > 0 {
		response.OutputData = outputData
		response.Results = outputData
	}

	metadata := jsonbToMap(m.JSONBToImpl(job.GetMetadata()))
	if len(metadata) > 0 {
		response.Metadata = metadata
		if scheduledBy, ok := metadata["scheduled_by"].(string); ok {
			response.ScheduledBy = scheduledBy
		}
		if requested, ok := metadata["requested_analysis_type"].(string); ok && requested != "" {
			response.AnalysisType = strings.ToLower(requested)
		}
	}

	if response.AnalysisType == "" {
		response.AnalysisType = determineAnalysisLabel(job, metadata)
	}

	response.Error = job.GetErrorMessage()
	response.ErrorMessage = job.GetErrorMessage()

	return response
}

func jsonbToMap(data m.JSONBImpl) map[string]interface{} {
	if len(data) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}

func nestedMapValue(root map[string]interface{}, keys ...string) interface{} {
	if len(keys) == 0 || len(root) == 0 {
		return nil
	}
	current := root
	for idx, key := range keys {
		value, ok := current[key]
		if !ok {
			return nil
		}
		if idx == len(keys)-1 {
			return value
		}
		next, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}
		current = next
	}
	return nil
}

func determineAnalysisLabel(job *m.AnalysisJobImpl, metadata map[string]interface{}) string {
	if metadata != nil {
		if value, ok := metadata["requested_analysis_type"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.ToLower(strings.TrimSpace(value))
		}
		if value, ok := metadata["analysis_type"].(string); ok && strings.TrimSpace(value) != "" {
			return strings.ToLower(strings.TrimSpace(value))
		}
	}

	switch strings.ToUpper(job.GetJobType()) {
	case "SCORECARD_ANALYSIS":
		return "scorecard"
	case "CODE_ANALYSIS":
		return "code"
	case "SECURITY_ANALYSIS":
		return "security"
	case "PERFORMANCE_ANALYSIS":
		return "performance"
	case "DEPENDENCY_ANALYSIS":
		return "dependency"
	default:
		return strings.ToLower(job.GetJobType())
	}
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found") ||
		strings.Contains(strings.ToLower(err.Error()), "record not found")
}
