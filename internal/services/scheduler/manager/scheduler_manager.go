package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	tp "github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"
)

type Scheduler struct {
	mu   sync.RWMutex
	jobs map[string]tp.Job
}

func (s *Scheduler) ScheduleJob(ctx context.Context, job tp.Job) (tp.JobStatusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.jobs == nil {
		s.jobs = make(map[string]tp.Job)
	}

	jobID := job.Ref().ID.String()
	s.jobs[jobID] = job

	// job

	return tp.JobStatusResponse{
		JobID:  jobID,
		Status: tp.NewStatus(tp.StatusScheduled, "queued by manager"),
	}, nil
}

func (s *Scheduler) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	j, ok := s.jobs[jobID.String()]
	if !ok {
		return fmt.Errorf("job with ID %s not found", jobID)
	}
	// j.SetStatus(tp.StatusCanceled, "canceled by manager")
	delete(s.jobs, j.Ref().ID.String())
	return nil
}

func (s *Scheduler) GetJobStatus(ctx context.Context, jobID uuid.UUID) (tp.JobStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// job, exists := s.jobs[jobID.String()]
	// if !exists {
	// 	return tp.JobStatus{}, fmt.Errorf("job with ID %s not found", jobID)
	// }
	// return job.GetStatus(), nil
	return tp.NewStatus(tp.StatusUnknown, "unknown job status"), nil
}

func (s *Scheduler) ListScheduledJobs(ctx context.Context) ([]tp.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	jobs := make([]tp.Job, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs, nil
}

func (s *Scheduler) RescheduleJob(ctx context.Context, jobID string, newSchedule string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// job, exists := s.jobs[jobID]
	// if !exists {
	// 	return fmt.Errorf("job with ID %s not found", jobID)
	// }
	// job.Schedule = newSchedule
	// job.SetStatus(tp.StatusRescheduled, "rescheduled by manager")
	return nil
}

func (s *Scheduler) StartScheduler(ctx context.Context) error { return nil }

func (s *Scheduler) StopScheduler(ctx context.Context) error { return nil }

// func (s *Scheduler) CancelJob(ctx context.Context, jobID string) error {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	j, ok := s.jobs[jobID]
// 	if !ok {
// 		return fmt.Errorf("job with ID %s not found", jobID)
// 	}
// 	j.SetStatus(tp.StatusCanceled, "canceled by manager")
// 	delete(s.jobs, jobID)
// 	return nil
// }

func (s *Scheduler) Health(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("scheduler is nil")
	}
	return nil
}
func (s *Scheduler) Stats(ctx context.Context) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return map[string]any{"total_jobs": len(s.jobs)}
}
