// Package analysis provides services for managing repository analysis jobs.
package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kubex-ecosystem/gobe/internal/models"
	"github.com/kubex-ecosystem/gobe/internal/services/analyzer"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	"gorm.io/gorm"
)

// JobService manages analysis jobs lifecycle
type JobService struct {
	db              *gorm.DB
	analyzerService *analyzer.Service
}

// NewJobService creates a new JobService instance
func NewJobService(db *gorm.DB, analyzerService *analyzer.Service) *JobService {
	return &JobService{
		db:              db,
		analyzerService: analyzerService,
	}
}

// CreateJob creates a new analysis job and optionally starts it
func (js *JobService) CreateJob(ctx context.Context, req CreateJobRequest) (*models.AnalysisJob, error) {
	// Validate request
	if req.RepoURL == "" {
		return nil, fmt.Errorf("repository URL is required")
	}
	if req.AnalysisType == "" {
		return nil, fmt.Errorf("analysis type is required")
	}

	// Serialize configuration and metadata
	configJSON, _ := json.Marshal(req.Configuration)
	metadataJSON, _ := json.Marshal(req.Metadata)
	notifyChannelsJSON, _ := json.Marshal(req.NotifyChannels)

	// Create job record
	job := &models.AnalysisJob{
		RepoURL:         req.RepoURL,
		AnalysisType:    req.AnalysisType,
		Status:          "scheduled",
		Progress:        0.0,
		ScheduledBy:     req.ScheduledBy,
		Configuration:   string(configJSON),
		Metadata:        string(metadataJSON),
		NotifyChannels:  string(notifyChannelsJSON),
	}

	// Save to database
	if err := js.db.Create(job).Error; err != nil {
		return nil, fmt.Errorf("failed to create analysis job: %w", err)
	}

	gl.Log("info", "Analysis job created", "job_id", job.ID, "repo_url", job.RepoURL, "type", job.AnalysisType)

	// Start job asynchronously if auto-start is enabled
	if req.AutoStart {
		go js.processJob(context.Background(), job.ID)
	}

	return job, nil
}

// GetJob retrieves a job by ID
func (js *JobService) GetJob(ctx context.Context, jobID string) (*models.AnalysisJob, error) {
	var job models.AnalysisJob
	if err := js.db.Where("id = ?", jobID).First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	return &job, nil
}

// ListJobs retrieves jobs with optional filtering
func (js *JobService) ListJobs(ctx context.Context, filter ListJobsFilter) ([]*models.AnalysisJob, int64, error) {
	query := js.db.Model(&models.AnalysisJob{})

	// Apply filters
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.RepoURL != "" {
		query = query.Where("repo_url = ?", filter.RepoURL)
	}
	if filter.AnalysisType != "" {
		query = query.Where("analysis_type = ?", filter.AnalysisType)
	}
	if filter.ScheduledBy != "" {
		query = query.Where("scheduled_by = ?", filter.ScheduledBy)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Apply pagination and ordering
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	var jobs []*models.AnalysisJob
	err := query.Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&jobs).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}

	return jobs, total, nil
}

// StartJob manually starts a scheduled job
func (js *JobService) StartJob(ctx context.Context, jobID string) error {
	job, err := js.GetJob(ctx, jobID)
	if err != nil {
		return err
	}

	if job.Status != "scheduled" {
		return fmt.Errorf("job %s is not in scheduled status (current: %s)", jobID, job.Status)
	}

	// Start job asynchronously
	go js.processJob(context.Background(), jobID)

	return nil
}

// processJob handles the actual job processing
func (js *JobService) processJob(ctx context.Context, jobID string) {
	gl.Log("info", "Starting job processing", "job_id", jobID)

	// Get job from database
	job, err := js.GetJob(ctx, jobID)
	if err != nil {
		gl.Log("error", "Failed to get job for processing", "job_id", jobID, "error", err)
		return
	}

	// Mark as started
	job.MarkAsStarted()
	if err := js.updateJob(job); err != nil {
		gl.Log("error", "Failed to mark job as started", "job_id", jobID, "error", err)
		return
	}

	// Process based on analysis type
	switch job.AnalysisType {
	case "scorecard":
		err = js.processScorecard(ctx, job)
	case "dora":
		err = js.processDORA(ctx, job)
	case "chi":
		err = js.processCHI(ctx, job)
	case "security":
		err = js.processSecurity(ctx, job)
	case "full":
		err = js.processFullAnalysis(ctx, job)
	default:
		err = fmt.Errorf("unsupported analysis type: %s", job.AnalysisType)
	}

	// Update job based on result
	if err != nil {
		gl.Log("error", "Job processing failed", "job_id", jobID, "error", err)
		job.MarkAsFailed(err.Error())
	} else {
		gl.Log("info", "Job processing completed", "job_id", jobID)
		job.MarkAsCompleted()
	}

	// Save final status
	if updateErr := js.updateJob(job); updateErr != nil {
		gl.Log("error", "Failed to update job final status", "job_id", jobID, "error", updateErr)
	}

	// Send notifications if configured
	js.sendJobNotifications(ctx, job)
}

// processScorecard handles scorecard analysis
func (js *JobService) processScorecard(ctx context.Context, job *models.AnalysisJob) error {
	if !js.analyzerService.IsEnabled() {
		return fmt.Errorf("analyzer service is not available")
	}

	// Parse configuration
	var config map[string]interface{}
	if job.Configuration != "" {
		if err := json.Unmarshal([]byte(job.Configuration), &config); err != nil {
			return fmt.Errorf("failed to parse job configuration: %w", err)
		}
	}

	// Update progress
	job.UpdateProgress(25.0)
	js.updateJob(job)

	// Create scorecard request
	scorecardReq := analyzer.ScorecardRequest{
		RepoURL:  job.RepoURL,
		Provider: getProviderFromConfig(config, "gemini"),
		Options:  config,
	}

	// Update progress
	job.UpdateProgress(50.0)
	js.updateJob(job)

	// Call GemX Analyzer
	scorecard, err := js.analyzerService.GetClient().GetRepositoryScorecard(ctx, scorecardReq)
	if err != nil {
		return fmt.Errorf("failed to get repository scorecard: %w", err)
	}

	// Update progress
	job.UpdateProgress(75.0)
	js.updateJob(job)

	// Store results
	resultsJSON, err := json.Marshal(scorecard)
	if err != nil {
		return fmt.Errorf("failed to serialize results: %w", err)
	}

	job.Results = string(resultsJSON)
	job.AnalyzerJobID = scorecard.JobID

	// Update progress to completion
	job.UpdateProgress(100.0)

	return nil
}

// processDORA, processCHI, processSecurity, processFullAnalysis are similar implementations
// For now, they will call the scorecard endpoint as a base implementation

func (js *JobService) processDORA(ctx context.Context, job *models.AnalysisJob) error {
	return js.processScorecard(ctx, job) // Use scorecard as base for now
}

func (js *JobService) processCHI(ctx context.Context, job *models.AnalysisJob) error {
	return js.processScorecard(ctx, job) // Use scorecard as base for now
}

func (js *JobService) processSecurity(ctx context.Context, job *models.AnalysisJob) error {
	return js.processScorecard(ctx, job) // Use scorecard as base for now
}

func (js *JobService) processFullAnalysis(ctx context.Context, job *models.AnalysisJob) error {
	return js.processScorecard(ctx, job) // Use scorecard as base for now
}

// updateJob saves job changes to database
func (js *JobService) updateJob(job *models.AnalysisJob) error {
	return js.db.Save(job).Error
}

// sendJobNotifications sends notifications about job completion
func (js *JobService) sendJobNotifications(ctx context.Context, job *models.AnalysisJob) {
	if job.NotifyChannels == "" {
		return
	}

	var channels []string
	if err := json.Unmarshal([]byte(job.NotifyChannels), &channels); err != nil {
		gl.Log("error", "Failed to parse notify channels", "job_id", job.ID, "error", err)
		return
	}

	// TODO: Implement actual notification sending
	for _, channel := range channels {
		gl.Log("info", "Sending notification", "job_id", job.ID, "channel", channel, "status", job.Status)
		// This will be implemented when we tackle the notification service
	}
}

// getProviderFromConfig extracts AI provider from configuration
func getProviderFromConfig(config map[string]interface{}, defaultProvider string) string {
	if config == nil {
		return defaultProvider
	}
	if provider, ok := config["provider"].(string); ok && provider != "" {
		return provider
	}
	return defaultProvider
}

// Request and filter types

type CreateJobRequest struct {
	RepoURL        string                 `json:"repo_url"`
	AnalysisType   string                 `json:"analysis_type"`
	ScheduledBy    string                 `json:"scheduled_by,omitempty"`
	Configuration  map[string]interface{} `json:"configuration,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	NotifyChannels []string               `json:"notify_channels,omitempty"`
	AutoStart      bool                   `json:"auto_start,omitempty"`
}

type ListJobsFilter struct {
	Status       string `json:"status,omitempty"`
	RepoURL      string `json:"repo_url,omitempty"`
	AnalysisType string `json:"analysis_type,omitempty"`
	ScheduledBy  string `json:"scheduled_by,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}