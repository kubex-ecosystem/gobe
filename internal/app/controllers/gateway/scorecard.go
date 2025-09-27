package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"github.com/kubex-ecosystem/gobe/internal/services/analyzer"
	"gorm.io/gorm"
)

// ScorecardController exposes real scorecard and metrics endpoints.
type ScorecardController struct {
	db                 *gorm.DB
	analyzerService    *analyzer.Service
	analysisJobService gdbasez.AnalysisJobService
}

func NewScorecardController(db *gorm.DB) *ScorecardController {
	// Initialize analyzer service
	analyzerBaseURL := getEnv("GEMX_ANALYZER_URL", "http://localhost:8080")
	analyzerAPIKey := getEnv("GEMX_ANALYZER_API_KEY", "")
	analyzerService := analyzer.NewService(analyzerBaseURL, analyzerAPIKey)

	// Initialize GDBase AnalysisJob service using gdbasez bridge
	analysisJobRepo := gdbasez.NewAnalysisJobRepo(db)
	analysisJobService := gdbasez.NewAnalysisJobService(analysisJobRepo)

	return &ScorecardController{
		db:                 db,
		analyzerService:    analyzerService,
		analysisJobService: analysisJobService,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetScorecard lists recent completed scorecard analyses from the system.
//
// @Summary     Listar scorecard
// @Description Retorna os scorecards mais recentes de repositórios analisados
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Param       limit query int false "Número de resultados (default: 10)"
// @Param       repo_url query string false "Filtrar por URL do repositório"
// @Success     200 {object} ScorecardResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/scorecard [get]
func (sc *ScorecardController) GetScorecard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	repoURL := c.Query("repo_url")

	limit := 10
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
		limit = parsedLimit
	}

	// Get recent completed scorecard analysis jobs
	jobs, err := sc.analysisJobService.ListJobsByStatus(c.Request.Context(), "COMPLETED")
	if err != nil {
		gl.Log("error", "Failed to get scorecard jobs", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve scorecard data",
		})
		return
	}

	// Filter jobs by type "SCORECARD_ANALYSIS"
	analysisJobs := make([]*gdbasez.AnalysisJobImpl, 0)
	for _, job := range jobs {
		if job.GetJobType() == "SCORECARD_ANALYSIS" {
			errorMsg := job.GetErrorMessage()
			completedAt := job.GetCompletedAt()
			updatedBy := job.GetUpdatedBy()
			// Create concrete type from interface
			analysisJob := &gdbasez.AnalysisJobImpl{
				ID:           job.GetID(),
				ProjectID:    job.GetProjectID(),
				JobType:      job.GetJobType(),
				Status:       job.GetStatus(),
				SourceURL:    job.GetSourceURL(),
				SourceType:   job.GetSourceType(),
				InputData:    job.GetInputData(),
				OutputData:   job.GetOutputData(),
				ErrorMessage: &errorMsg,
				Progress:     job.GetProgress(),
				StartedAt:    job.GetStartedAt(),
				CompletedAt:  &completedAt,
				RetryCount:   job.GetRetryCount(),
				MaxRetries:   job.GetMaxRetries(),
				Metadata:     job.GetMetadata(),
				UserID:       job.GetUserID(),
				CreatedBy:    job.GetCreatedBy(),
				UpdatedBy:    &updatedBy,
				CreatedAt:    job.GetCreatedAt(),
				UpdatedAt:    job.GetUpdatedAt(),
			}
			analysisJobs = append(analysisJobs, analysisJob)
		}
	}

	// Convert analysis jobs to scorecard entries
	entries := make([]ScorecardEntry, 0, len(analysisJobs))
	for _, job := range analysisJobs {
		entry := convertAnalysisJobToScorecardEntry(job)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	// Apply filtering and pagination
	filteredEntries := entries
	if repoURL != "" {
		filteredEntries = make([]ScorecardEntry, 0)
		for _, entry := range entries {
			// Filter by repo URL in title or description
			if strings.Contains(strings.ToLower(entry.Title), strings.ToLower(repoURL)) ||
				strings.Contains(strings.ToLower(entry.Description), strings.ToLower(repoURL)) {
				filteredEntries = append(filteredEntries, entry)
			}
		}
	}

	// Apply limit
	if len(filteredEntries) > limit {
		filteredEntries = filteredEntries[:limit]
	}

	gl.Log("info", "Scorecard data retrieved", "count", len(filteredEntries), "total_jobs", len(analysisJobs))

	c.JSON(http.StatusOK, ScorecardResponse{
		Items:   filteredEntries,
		Total:   len(analysisJobs),
		Version: "gobe-real-v1.3.5",
	})
}

// GetScorecardAdvice returns high-level guidance based on recent analysis data.
//
// @Summary     Aconselhar scorecard
// @Description Fornece sugestões inteligentes baseadas nas métricas mais recentes de análises
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Param       repo_url query string false "URL específica do repositório para análise"
// @Success     200 {object} ScorecardAdviceResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/scorecard/advice [get]
func (sc *ScorecardController) GetScorecardAdvice(c *gin.Context) {
	repoURL := c.Query("repo_url")

	// Get recent completed analysis jobs for advice analysis
	allJobs, err := sc.analysisJobService.ListJobsByStatus(c.Request.Context(), "COMPLETED")
	if err != nil {
		gl.Log("error", "Failed to get jobs for advice", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Message: "Failed to generate advice",
		})
		return
	}

	// Filter jobs by type "SCORECARD_ANALYSIS" and by repo URL if specified
	analysisJobs := make([]*gdbasez.AnalysisJobImpl, 0)
	for _, job := range allJobs {
		if job.GetJobType() == "SCORECARD_ANALYSIS" {
			// Filter by repo URL if specified
			metadataStr := ""
			if metadata := job.GetMetadata(); metadata != nil {
				if metadataBytes, err := json.Marshal(metadata); err == nil {
					metadataStr = string(metadataBytes)
				}
			}
			if repoURL == "" ||
				strings.Contains(job.GetSourceURL(), repoURL) ||
				strings.Contains(metadataStr, repoURL) {
				errorMsg := job.GetErrorMessage()
				completedAt := job.GetCompletedAt()
				updatedBy := job.GetUpdatedBy()
				// Create concrete type from interface
				analysisJob := &gdbasez.AnalysisJobImpl{
					ID:           job.GetID(),
					ProjectID:    job.GetProjectID(),
					JobType:      job.GetJobType(),
					Status:       job.GetStatus(),
					SourceURL:    job.GetSourceURL(),
					SourceType:   job.GetSourceType(),
					InputData:    job.GetInputData(),
					OutputData:   job.GetOutputData(),
					ErrorMessage: &errorMsg,
					Progress:     job.GetProgress(),
					StartedAt:    job.GetStartedAt(),
					CompletedAt:  &completedAt,
					RetryCount:   job.GetRetryCount(),
					MaxRetries:   job.GetMaxRetries(),
					Metadata:     job.GetMetadata(),
					UserID:       job.GetUserID(),
					CreatedBy:    job.GetCreatedBy(),
					UpdatedBy:    &updatedBy,
					CreatedAt:    job.GetCreatedAt(),
					UpdatedAt:    job.GetUpdatedAt(),
				}
				analysisJobs = append(analysisJobs, analysisJob)
			}
		}
	}

	// Limit to last 5 jobs
	if len(analysisJobs) > 5 {
		analysisJobs = analysisJobs[:5]
	}

	// Generate advice based on analysis job results
	advice := generateAdviceFromAnalysisJobs(analysisJobs, repoURL)

	gl.Log("info", "Scorecard advice generated", "repo_url", repoURL, "jobs_analyzed", len(analysisJobs))

	c.JSON(http.StatusOK, ScorecardAdviceResponse{
		Advice:      advice.Message,
		Priority:    advice.Priority,
		Actions:     advice.Actions,
		Metrics:     advice.Metrics,
		Version:     "gobe-real-v1.3.5",
		GeneratedAt: time.Now().UTC(),
	})
}

// GetMetrics exposes real system metrics based on analysis jobs and system health.
//
// @Summary     Métricas de IA
// @Description Entrega métricas reais agregadas do sistema de análise e gateway
// @Tags        gateway
// @Security    BearerAuth
// @Produce     json
// @Param       period query string false "Período de análise (1h, 24h, 7d, 30d) - default: 24h"
// @Success     200 {object} ScorecardMetricsResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/metrics/ai [get]
func (sc *ScorecardController) GetMetrics(c *gin.Context) {
	period := c.DefaultQuery("period", "24h")

	// Calculate time window based on period
	var since time.Time
	switch period {
	case "1h":
		since = time.Now().Add(-1 * time.Hour)
	case "7d":
		since = time.Now().Add(-7 * 24 * time.Hour)
	case "30d":
		since = time.Now().Add(-30 * 24 * time.Hour)
	default: // 24h
		since = time.Now().Add(-24 * time.Hour)
	}

	// Get all analysis jobs for metrics calculation
	allJobs, err := sc.analysisJobService.ListJobs(c.Request.Context())
	if err != nil {
		gl.Log("error", "Failed to get jobs for metrics", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Message: "Failed to calculate metrics",
		})
		return
	}
	var allAnalysisJobs []*gdbasez.AnalysisJobImpl
	for _, job := range allJobs {
		errorMsg := job.GetErrorMessage()
		completedAt := job.GetCompletedAt()
		updatedBy := job.GetUpdatedBy()
		// Create concrete type from interface
		analysisJob := &gdbasez.AnalysisJobImpl{
			ID:           job.GetID(),
			ProjectID:    job.GetProjectID(),
			JobType:      job.GetJobType(),
			Status:       job.GetStatus(),
			SourceURL:    job.GetSourceURL(),
			SourceType:   job.GetSourceType(),
			InputData:    job.GetInputData(),
			OutputData:   job.GetOutputData(),
			ErrorMessage: &errorMsg,
			Progress:     job.GetProgress(),
			StartedAt:    job.GetStartedAt(),
			CompletedAt:  &completedAt,
			RetryCount:   job.GetRetryCount(),
			MaxRetries:   job.GetMaxRetries(),
			Metadata:     job.GetMetadata(),
			UserID:       job.GetUserID(),
			CreatedBy:    job.GetCreatedBy(),
			UpdatedBy:    &updatedBy,
			CreatedAt:    job.GetCreatedAt(),
			UpdatedAt:    job.GetUpdatedAt(),
		}
		allAnalysisJobs = append(allAnalysisJobs, analysisJob)
	}

	// Calculate real metrics from analysis job data
	metrics := calculateAnalysisMetrics(allAnalysisJobs, since, period)
	metrics["total_jobs"] = len(allJobs)
	metrics["period"] = period
	metrics["calculated_at"] = time.Now().UTC()

	// Add analyzer service health
	if sc.analyzerService.IsEnabled() {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := sc.analyzerService.HealthCheck(ctx); err == nil {
			metrics["analyzer_healthy"] = true
		} else {
			metrics["analyzer_healthy"] = false
			metrics["analyzer_error"] = err.Error()
		}
	} else {
		metrics["analyzer_healthy"] = false
		metrics["analyzer_error"] = "Service disabled"
	}

	gl.Log("info", "System metrics calculated", "period", period, "jobs_analyzed", len(allJobs))

	c.JSON(http.StatusOK, ScorecardMetricsResponse{
		Metrics: metrics,
		Version: "gobe-real-v1.3.5",
	})
}

// convertAnalysisJobToScorecardEntry converts an analysis job to a scorecard entry
func convertAnalysisJobToScorecardEntry(job *gdbasez.AnalysisJobImpl) *ScorecardEntry {

	// Parse metadata and output data to extract score and tags
	score := 0.0
	tags := []string{job.GetJobType()}

	// Try to parse score from output data first, then metadata
	outputData := job.GetOutputData()
	if outputData != nil {
		// outputData is already a map[string]interface{}
		if scoreVal, ok := outputData["overall_score"].(float64); ok {
			score = scoreVal
		}
		if tagsVal, ok := outputData["tags"].([]interface{}); ok {
			for _, tag := range tagsVal {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}
	}

	// Fallback to metadata if no score in output
	if score == 0.0 {
		metadata := job.GetMetadata()
		if metadata != nil {
			// metadata is already a map[string]interface{}
			if scoreVal, ok := metadata["overall_score"].(float64); ok {
				score = scoreVal
			}
		}
	}

	// Generate title based on source URL and job type
	title := fmt.Sprintf("%s Analysis", strings.Title(strings.ToLower(strings.ReplaceAll(job.GetJobType(), "_", " "))))
	sourceURL := job.GetSourceURL()
	if sourceURL != "" {
		// Extract repo name from source URL
		repoName := extractRepoName(sourceURL)
		if repoName != "unknown" {
			title = fmt.Sprintf("%s: %s", repoName, title)
		}
	}

	description := fmt.Sprintf("Repository analysis completed with %s", job.GetJobType())
	errorMsg := job.GetErrorMessage()
	if errorMsg != "" {
		description = fmt.Sprintf("Analysis failed: %s", errorMsg)
		score = 0.0
	}

	return &ScorecardEntry{
		ID:          job.GetID().String(),
		Title:       title,
		Description: description,
		Score:       score,
		UpdatedAt:   job.GetUpdatedAt(),
		Tags:        tags,
	}
}

// generateAdviceFromAnalysisJobs analyzes recent analysis jobs and generates intelligent advice
func generateAdviceFromAnalysisJobs(jobs []*gdbasez.AnalysisJobImpl, repoURL string) *AdviceData {
	if len(jobs) == 0 {
		return &AdviceData{
			Message:  "No recent analysis data available for generating advice",
			Priority: "medium",
			Actions:  []string{"Run a new analysis to get insights"},
			Metrics:  map[string]interface{}{"jobs_analyzed": 0},
		}
	}

	// Analyze analysis job patterns
	completedJobs := 0
	failedJobs := 0
	avgScore := 0.0
	scoreCount := 0
	var lastJob *gdbasez.AnalysisJobImpl
	// issues := []string{}
	recommendations := []string{}

	for _, job := range jobs {
		status := job.GetStatus()

		if status == "COMPLETED" {
			completedJobs++

			// Parse output data first, then metadata for scoring
			outputData := job.GetOutputData()
			if outputData != nil {
				// outputData is already a map[string]interface{}
				if score, ok := outputData["overall_score"].(float64); ok {
					avgScore += score
					scoreCount++
				}
			} else {
				// Fallback to metadata
				metadata := job.GetMetadata()
				if metadata != nil {
					// metadata is already a map[string]interface{}
					if score, ok := metadata["overall_score"].(float64); ok {
						avgScore += score
						scoreCount++
					}
				}
			}
		}

		// if status == "FAILED" {
		// 	failedJobs++
		// 	errorMsg := job.GetErrorMessage()
		// 	if errorMsg != "" {
		// 		issues = append(issues, errorMsg)
		// 	}
		// }

		if lastJob == nil || job.GetUpdatedAt().After(lastJob.GetUpdatedAt()) {
			lastJob = job
		}
	}

	if scoreCount > 0 {
		avgScore = avgScore / float64(scoreCount)
	}

	// Generate advice based on analysis
	var message string
	var priority string

	if failedJobs > completedJobs/2 {
		message = fmt.Sprintf("High failure rate detected (%d/%d jobs failed). Recent analyses show recurring issues that need attention.", failedJobs, len(jobs))
		priority = "high"
		recommendations = append(recommendations, "Review repository configuration", "Check network connectivity", "Validate repository access permissions")
	} else if avgScore < 0.6 && scoreCount > 0 {
		message = fmt.Sprintf("Analysis scores are below average (%.2f/1.0). Repository shows areas for improvement.", avgScore)
		priority = "medium"
		recommendations = append(recommendations, "Focus on security improvements", "Update dependencies", "Improve documentation coverage")
	} else if completedJobs == len(jobs) && avgScore > 0.8 {
		message = fmt.Sprintf("Excellent repository health! Recent analyses show consistently high scores (%.2f/1.0).", avgScore)
		priority = "low"
		recommendations = append(recommendations, "Maintain current practices", "Consider advanced security features", "Share best practices with team")
	} else {
		message = fmt.Sprintf("Repository showing steady progress. %d analyses completed with average health metrics.", completedJobs)
		priority = "medium"
		recommendations = append(recommendations, "Continue regular analysis", "Monitor trends over time", "Address any failing checks")
	}

	// Add repository-specific advice if targeting specific repo
	if repoURL != "" {
		message = fmt.Sprintf("[%s] %s", extractRepoName(repoURL), message)
	}

	return &AdviceData{
		Message:  message,
		Priority: priority,
		Actions:  recommendations,
		Metrics: map[string]interface{}{
			"jobs_analyzed":  len(jobs),
			"completed_jobs": completedJobs,
			"failed_jobs":    failedJobs,
			"average_score":  avgScore,
			"last_analysis":  lastJob.GetUpdatedAt(),
			"success_rate":   float64(completedJobs) / float64(len(jobs)),
		},
	}
}

// calculateAnalysisMetrics calculates real system metrics from analysis job data
func calculateAnalysisMetrics(jobs []*gdbasez.AnalysisJobImpl, since time.Time, period string) map[string]interface{} {
	metrics := make(map[string]interface{})

	// Filter analysis jobs by time period
	periodJobs := make([]*gdbasez.AnalysisJobImpl, 0)
	for _, job := range jobs {
		if job.GetCreatedAt().After(since) {
			periodJobs = append(periodJobs, job)
		}
	}

	// Calculate basic metrics
	totalJobs := len(periodJobs)
	completedJobs := 0
	failedJobs := 0
	runningJobs := 0
	scheduledJobs := 0
	totalDuration := time.Duration(0)

	analysisTypes := make(map[string]int)
	repositoryStats := make(map[string]int)
	avgScores := make(map[string]float64)
	scoreCounts := make(map[string]int)

	for _, job := range periodJobs {
		status := job.GetStatus()

		// Count by status
		switch status {
		case "COMPLETED":
			completedJobs++
			// Calculate duration for completed jobs
			completedAt := job.GetCompletedAt()
			startedAt := job.GetStartedAt()
			if !completedAt.IsZero() && !startedAt.IsZero() {
				duration := completedAt.Sub(startedAt)
				totalDuration += duration
			}
		case "FAILED":
			failedJobs++
		case "RUNNING":
			runningJobs++
		case "PENDING":
			scheduledJobs++
		}

		// Count by job type
		jobType := job.GetJobType()
		analysisTypes[jobType]++

		// Count by repository (extracted from source URL)
		repoName := extractRepoNameFromAnalysisJob(job)
		repositoryStats[repoName]++

		// Calculate scores by type from output data or metadata
		if status == "COMPLETED" {
			var score float64
			found := false

			// Try output data first
			outputData := job.GetOutputData()
			if outputData != nil {
				// outputData is already a map[string]interface{}
				if scoreVal, ok := outputData["overall_score"].(float64); ok {
					score = scoreVal
					found = true
				}
			}

			// Fallback to metadata
			if !found {
				metadata := job.GetMetadata()
				if metadata != nil {
					// metadata is already a map[string]interface{}
					if scoreVal, ok := metadata["overall_score"].(float64); ok {
						score = scoreVal
						found = true
					}
				}
			}

			if found {
				avgScores[jobType] += score
				scoreCounts[jobType]++
			}
		}
	}

	// Calculate averages
	var avgDuration time.Duration
	if completedJobs > 0 {
		avgDuration = totalDuration / time.Duration(completedJobs)
	}

	// Calculate success rate
	successRate := 0.0
	if totalJobs > 0 {
		successRate = float64(completedJobs) / float64(totalJobs)
	}

	// Calculate average scores by type
	typeScores := make(map[string]float64)
	for analysisType, total := range avgScores {
		if count := scoreCounts[analysisType]; count > 0 {
			typeScores[analysisType] = total / float64(count)
		}
	}

	// Build metrics response
	metrics["period"] = period
	metrics["total_jobs"] = totalJobs
	metrics["completed_jobs"] = completedJobs
	metrics["failed_jobs"] = failedJobs
	metrics["running_jobs"] = runningJobs
	metrics["scheduled_jobs"] = scheduledJobs
	metrics["success_rate"] = successRate
	metrics["average_duration_seconds"] = int(avgDuration.Seconds())
	metrics["analysis_types"] = analysisTypes
	metrics["repository_stats"] = repositoryStats
	metrics["type_scores"] = typeScores
	metrics["jobs_per_hour"] = calculateJobsPerHour(periodJobs, since)

	return metrics
}

// AdviceData holds the advice generation result
type AdviceData struct {
	Message  string                 `json:"message"`
	Priority string                 `json:"priority"`
	Actions  []string               `json:"actions"`
	Metrics  map[string]interface{} `json:"metrics"`
}

// extractRepoName extracts repository name from URL
func extractRepoName(repoURL string) string {
	if repoURL == "" {
		return "unknown"
	}

	// Remove .git suffix and extract last part
	parts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// extractRepoNameFromAnalysisJob extracts repository name from analysis job source URL or metadata
func extractRepoNameFromAnalysisJob(job *gdbasez.AnalysisJobImpl) string {
	// Try to extract from source URL first
	sourceURL := job.GetSourceURL()
	if sourceURL != "" {
		if repoName := extractRepoName(sourceURL); repoName != "unknown" {
			return repoName
		}
	}

	// Try to extract from metadata
	metadata := job.GetMetadata()
	if metadata != nil {
		// metadata is already a map[string]interface{}
		if repoURL, ok := metadata["repo_url"].(string); ok {
			return extractRepoName(repoURL)
		}
	}

	return "unknown"
}

// calculateJobsPerHour calculates analysis jobs per hour for the given period
func calculateJobsPerHour(jobs []*gdbasez.AnalysisJobImpl, since time.Time) float64 {
	if len(jobs) == 0 {
		return 0.0
	}

	duration := time.Since(since)
	hours := duration.Hours()
	if hours <= 0 {
		return 0.0
	}

	return float64(len(jobs)) / hours
}
