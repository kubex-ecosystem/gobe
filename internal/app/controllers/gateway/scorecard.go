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
	"github.com/kubex-ecosystem/gobe/internal/models"
	"github.com/kubex-ecosystem/gobe/internal/services/analysis"
	"github.com/kubex-ecosystem/gobe/internal/services/analyzer"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"gorm.io/gorm"
)

// ScorecardController exposes real scorecard and metrics endpoints.
type ScorecardController struct {
	db              *gorm.DB
	analyzerService *analyzer.Service
	jobService      *analysis.JobService
}

func NewScorecardController(db *gorm.DB) *ScorecardController {
	// Initialize analyzer service
	analyzerBaseURL := getEnv("GEMX_ANALYZER_URL", "http://localhost:8080")
	analyzerAPIKey := getEnv("GEMX_ANALYZER_API_KEY", "")

	analyzerService := analyzer.NewService(analyzerBaseURL, analyzerAPIKey)
	jobService := analysis.NewJobService(db, analyzerService)

	return &ScorecardController{
		db:              db,
		analyzerService: analyzerService,
		jobService:      jobService,
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

	// Get recent completed analysis jobs
	filter := analysis.ListJobsFilter{
		Status:       "completed",
		RepoURL:      repoURL,
		AnalysisType: "scorecard",
		Limit:        limit,
		Offset:       0,
	}

	jobs, total, err := sc.jobService.ListJobs(c.Request.Context(), filter)
	if err != nil {
		gl.Log("error", "Failed to get scorecard jobs", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve scorecard data",
		})
		return
	}

	// Convert jobs to scorecard entries
	entries := make([]ScorecardEntry, 0, len(jobs))
	for _, job := range jobs {
		entry := convertJobToScorecardEntry(job)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	gl.Log("info", "Scorecard data retrieved", "count", len(entries), "total_jobs", total)

	c.JSON(http.StatusOK, ScorecardResponse{
		Items:   entries,
		Total:   int(total),
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

	// Get recent completed jobs for advice analysis
	filter := analysis.ListJobsFilter{
		Status:       "completed",
		RepoURL:      repoURL,
		AnalysisType: "scorecard",
		Limit:        5, // Analyze last 5 jobs
		Offset:       0,
	}

	jobs, _, err := sc.jobService.ListJobs(c.Request.Context(), filter)
	if err != nil {
		gl.Log("error", "Failed to get jobs for advice", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Message: "Failed to generate advice",
		})
		return
	}

	// Generate advice based on job results
	advice := generateAdviceFromJobs(jobs, repoURL)

	gl.Log("info", "Scorecard advice generated", "repo_url", repoURL, "jobs_analyzed", len(jobs))

	c.JSON(http.StatusOK, ScorecardAdviceResponse{
		Advice:     advice.Message,
		Priority:   advice.Priority,
		Actions:    advice.Actions,
		Metrics:    advice.Metrics,
		Version:    "gobe-real-v1.3.5",
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

	// Get all jobs in the period for metrics calculation
	allFilter := analysis.ListJobsFilter{
		Limit:  1000, // Get many jobs for accurate metrics
		Offset: 0,
	}

	allJobs, totalJobs, err := sc.jobService.ListJobs(c.Request.Context(), allFilter)
	if err != nil {
		gl.Log("error", "Failed to get jobs for metrics", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Status:  "error",
			Message: "Failed to calculate metrics",
		})
		return
	}

	// Calculate real metrics from job data
	metrics := calculateSystemMetrics(allJobs, since, period)
	metrics["total_jobs"] = totalJobs
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

// convertJobToScorecardEntry converts an analysis job to a scorecard entry
func convertJobToScorecardEntry(job *models.AnalysisJob) *ScorecardEntry {
	if job == nil {
		return nil
	}

	// Parse results to extract score and tags if available
	score := 0.0
	tags := []string{job.AnalysisType}

	if job.Results != "" {
		var results map[string]interface{}
		if err := json.Unmarshal([]byte(job.Results), &results); err == nil {
			if scoreVal, ok := results["overall_score"].(float64); ok {
				score = scoreVal
			}
			if tagsVal, ok := results["tags"].([]interface{}); ok {
				for _, tag := range tagsVal {
					if tagStr, ok := tag.(string); ok {
						tags = append(tags, tagStr)
					}
				}
			}
		}
	}

	// Generate title based on repository URL and analysis type
	title := fmt.Sprintf("%s Analysis", strings.Title(job.AnalysisType))
	if job.RepoURL != "" {
		// Extract repo name from URL
		parts := strings.Split(strings.TrimSuffix(job.RepoURL, ".git"), "/")
		if len(parts) > 0 {
			repoName := parts[len(parts)-1]
			title = fmt.Sprintf("%s: %s", repoName, title)
		}
	}

	description := fmt.Sprintf("Repository analysis completed with %s", job.AnalysisType)
	if job.Error != "" {
		description = fmt.Sprintf("Analysis failed: %s", job.Error)
		score = 0.0
	}

	return &ScorecardEntry{
		ID:          job.ID,
		Title:       title,
		Description: description,
		Score:       score,
		UpdatedAt:   job.UpdatedAt,
		Tags:        tags,
	}
}

// generateAdviceFromJobs analyzes recent jobs and generates intelligent advice
func generateAdviceFromJobs(jobs []*models.AnalysisJob, repoURL string) *AdviceData {
	if len(jobs) == 0 {
		return &AdviceData{
			Message:  "No recent analysis data available for generating advice",
			Priority: "medium",
			Actions:  []string{"Run a new analysis to get insights"},
			Metrics:  map[string]interface{}{"jobs_analyzed": 0},
		}
	}

	// Analyze job patterns
	completedJobs := 0
	failedJobs := 0
	avgScore := 0.0
	scoreCount := 0
	var lastJob *models.AnalysisJob
	issues := []string{}
	recommendations := []string{}

	for _, job := range jobs {
		if job.IsCompleted() {
			completedJobs++

			// Parse results for scoring
			if job.Results != "" {
				var results map[string]interface{}
				if err := json.Unmarshal([]byte(job.Results), &results); err == nil {
					if score, ok := results["overall_score"].(float64); ok {
						avgScore += score
						scoreCount++
					}
				}
			}
		}
		if job.IsFailed() {
			failedJobs++
			if job.Error != "" {
				issues = append(issues, job.Error)
			}
		}
		if lastJob == nil || job.UpdatedAt.After(lastJob.UpdatedAt) {
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
			"jobs_analyzed":    len(jobs),
			"completed_jobs":   completedJobs,
			"failed_jobs":      failedJobs,
			"average_score":    avgScore,
			"last_analysis":    lastJob.UpdatedAt,
			"success_rate":     float64(completedJobs) / float64(len(jobs)),
		},
	}
}

// calculateSystemMetrics calculates real system metrics from job data
func calculateSystemMetrics(jobs []*models.AnalysisJob, since time.Time, period string) map[string]interface{} {
	metrics := make(map[string]interface{})

	// Filter jobs by time period
	periodJobs := make([]*models.AnalysisJob, 0)
	for _, job := range jobs {
		if job.CreatedAt.After(since) {
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
		// Count by status
		switch job.Status {
		case "completed":
			completedJobs++
			// Calculate duration for completed jobs
			if job.StartedAt != nil && job.CompletedAt != nil {
				duration := job.CompletedAt.Sub(*job.StartedAt)
				totalDuration += duration
			}
		case "failed":
			failedJobs++
		case "running":
			runningJobs++
		case "scheduled":
			scheduledJobs++
		}

		// Count by analysis type
		analysisTypes[job.AnalysisType]++

		// Count by repository
		repoName := extractRepoName(job.RepoURL)
		repositoryStats[repoName]++

		// Calculate scores by type
		if job.Results != "" && job.IsCompleted() {
			var results map[string]interface{}
			if err := json.Unmarshal([]byte(job.Results), &results); err == nil {
				if score, ok := results["overall_score"].(float64); ok {
					avgScores[job.AnalysisType] += score
					scoreCounts[job.AnalysisType]++
				}
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

// calculateJobsPerHour calculates jobs per hour for the given period
func calculateJobsPerHour(jobs []*models.AnalysisJob, since time.Time) float64 {
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
