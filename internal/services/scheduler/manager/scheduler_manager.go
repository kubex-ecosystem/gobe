package manager

import (
	"fmt"

	tp "github.com/rafa-mori/gobe/internal/services/scheduler/types"
)

type Scheduler struct {
	// Implement the scheduler fields
	jobs map[string]tp.Job
}

// ScheduleJob adds a new job to the scheduler.
func (s *Scheduler) ScheduleJob(job tp.Job) error {
	if s.jobs == nil {
		s.jobs = make(map[string]tp.Job)
	}
	jobID := job.Ref().ID.String() // Convertendo UUID para string
	s.jobs[jobID] = job
	return nil
}

// CancelJob removes a job from the scheduler by its ID.
func (s *Scheduler) CancelJob(jobID string) error {
	if _, exists := s.jobs[jobID]; !exists {
		return fmt.Errorf("job with ID %s not found", jobID)
	}
	delete(s.jobs, jobID)
	return nil
}

// GetJobStatus retrieves the current status of a job by its ID.
func (s *Scheduler) GetJobStatus(jobID string) (tp.JobStatus, error) {
	job, exists := s.jobs[jobID]
	if !exists {
		return "", fmt.Errorf("job with ID %s not found", jobID)
	}
	return job.Status, nil
}

// ListScheduledJobs returns a list of all scheduled jobs.
func (s *Scheduler) ListScheduledJobs() ([]tp.Job, error) {
	jobs := make([]tp.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// RescheduleJob updates the schedule of an existing job.
func (s *Scheduler) RescheduleJob(jobID string, newSchedule string) error {
	job, exists := s.jobs[jobID]
	if !exists {
		return fmt.Errorf("job with ID %s not found", jobID)
	}
	job.Schedule = newSchedule
	s.jobs[jobID] = job
	return nil
}

// StartScheduler starts the scheduler to process jobs.
func (s *Scheduler) StartScheduler() error {
	// Implementation for starting the scheduler
	return nil
}

// StopScheduler stops the scheduler gracefully.
func (s *Scheduler) StopScheduler() error {
	// Implementation for stopping the scheduler
	return nil
}
