// Package scheduler provides an interface for scheduling jobs.
package scheduler

import (
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"
)

type IScheduler interface {
	// ScheduleJob adds a new job to the scheduler.
	ScheduleJob(job types.Job) error

	// CancelJob removes a job from the scheduler by its ID.
	CancelJob(jobID string) error

	// GetJobStatus retrieves the current status of a job by its ID.
	GetJobStatus(jobID string) (types.JobStatus, error)

	// ListScheduledJobs returns a list of all scheduled jobs.
	ListScheduledJobs() ([]types.Job, error)

	// RescheduleJob updates the schedule of an existing job.
	RescheduleJob(jobID string, newSchedule string) error

	// StartScheduler starts the scheduler to process jobs.
	StartScheduler() error

	// StopScheduler stops the scheduler gracefully.
	StopScheduler() error
}
