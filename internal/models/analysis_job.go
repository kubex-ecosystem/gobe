// Package models provides database models for the GoBE application.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AnalysisJob represents a repository analysis job in the database
type AnalysisJob struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RepoURL      string    `gorm:"not null;index" json:"repo_url"`
	AnalysisType string    `gorm:"not null" json:"analysis_type"`
	Status       string    `gorm:"not null;index" json:"status"` // "scheduled", "running", "completed", "failed"
	Progress     float64   `gorm:"default:0" json:"progress"`
	ScheduledBy  string    `json:"scheduled_by,omitempty"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Results and metadata stored as JSON
	Results  string `gorm:"type:text" json:"results,omitempty"`
	Error    string `gorm:"type:text" json:"error,omitempty"`
	Metadata string `gorm:"type:text" json:"metadata,omitempty"`

	// Configuration
	Configuration string `gorm:"type:text" json:"configuration,omitempty"`

	// External references
	AnalyzerJobID   string `gorm:"index" json:"analyzer_job_id,omitempty"`
	NotifyChannels  string `gorm:"type:text" json:"notify_channels,omitempty"`
}

// BeforeCreate generates UUID for new jobs
func (aj *AnalysisJob) BeforeCreate(tx *gorm.DB) error {
	if aj.ID == "" {
		aj.ID = uuid.New().String()
	}
	return nil
}

// TableName returns the table name for AnalysisJob
func (AnalysisJob) TableName() string {
	return "analysis_jobs"
}

// IsRunning returns true if the job is currently running
func (aj *AnalysisJob) IsRunning() bool {
	return aj.Status == "running"
}

// IsCompleted returns true if the job has completed successfully
func (aj *AnalysisJob) IsCompleted() bool {
	return aj.Status == "completed"
}

// IsFailed returns true if the job has failed
func (aj *AnalysisJob) IsFailed() bool {
	return aj.Status == "failed"
}

// MarkAsStarted updates the job status to running and sets started timestamp
func (aj *AnalysisJob) MarkAsStarted() {
	aj.Status = "running"
	now := time.Now()
	aj.StartedAt = &now
	aj.UpdatedAt = now
}

// MarkAsCompleted updates the job status to completed and sets completed timestamp
func (aj *AnalysisJob) MarkAsCompleted() {
	aj.Status = "completed"
	aj.Progress = 100.0
	now := time.Now()
	aj.CompletedAt = &now
	aj.UpdatedAt = now
}

// MarkAsFailed updates the job status to failed and sets error message
func (aj *AnalysisJob) MarkAsFailed(errorMsg string) {
	aj.Status = "failed"
	aj.Error = errorMsg
	now := time.Now()
	aj.CompletedAt = &now
	aj.UpdatedAt = now
}

// UpdateProgress updates the job progress percentage
func (aj *AnalysisJob) UpdateProgress(progress float64) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	aj.Progress = progress
	aj.UpdatedAt = time.Now()
}