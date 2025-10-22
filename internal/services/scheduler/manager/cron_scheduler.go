// Package manager fornece implementações para o gerenciamento de cronjobs usando GoroutinePool.
package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pl "github.com/kubex-ecosystem/gobe/internal/services/scheduler/services"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// CronJobScheduler gerencia a execução de cronjobs usando o GoroutinePool.
type CronJobScheduler struct {
	pool         *pl.GoroutinePool
	ICronService pl.ICronService // Interface para interagir com o serviço de cronjobs
}

// NewCronJobScheduler cria uma nova instância do CronJobScheduler.
func NewCronJobScheduler(pool *pl.GoroutinePool, ICronService pl.ICronService) *CronJobScheduler {
	return &CronJobScheduler{
		pool:         pool,
		ICronService: ICronService,
	}
}

// StartScheduler inicia o loop de verificação e execução de cronjobs.
func (s *CronJobScheduler) StartScheduler(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(1 * time.Minute) // Verifica os cronjobs a cada minuto
		defer ticker.Stop()
		for range ticker.C {
			cronJobs, err := s.ICronService.GetScheduledCronJobs(ctx)
			if err != nil {
				gl.Log("error", "Error fetching scheduled cronjobs: %v", err)
				continue
			}
			for _, job := range cronJobs {
				s.pool.Submit(job)
			}
		}
	}()
	return nil
}

func (s *CronJobScheduler) StopScheduler(ctx context.Context) error {
	s.pool.Stop()
	return nil
}

func (s *CronJobScheduler) CancelJob(ctx context.Context, jobID uuid.UUID) error {
	scj, err := s.ICronService.GetScheduledCronJobs(ctx)
	if err != nil {
		return err
	}
	for _, job := range scj {
		if job.Ref().GetID() == jobID {
			return job.Cancel()
		}
	}
	return fmt.Errorf("job with ID %s not found", jobID)
}

func (s *CronJobScheduler) GetJobStatus(ctx context.Context, jobID uuid.UUID) (types.JobStatus, error) {
	scj, err := s.ICronService.GetScheduledCronJobs(ctx)
	if err != nil {
		return types.JobStatus{}, err
	}
	for _, job := range scj {
		if job.Ref().GetID() == jobID {
			if err := uuid.Validate(job.Ref().GetID().String()); err != nil {
				return types.JobStatus{}, fmt.Errorf("invalid job ID format: %v", err)
			}
			jbStatus := types.JobStatus{
				// ID:        job.Ref().GetID(),
				// Name:      job.Ref().GetName(),
				// Schedule:  job.Ref().GetSchedule(),
				// NextRun:   job.Ref().GetNextRun(),
				// Status:    job.Ref().GetStatus(),
				// Progress:  job.Ref().GetProgress(),
				//CreatedAt: job.Ref().GetCreatedAt(),
				// UpdatedAt: job.Ref().GetUpdatedAt(),
			}
			return jbStatus, nil
		}
	}
	return types.JobStatus{}, fmt.Errorf("job with ID %s not found", jobID)
}

func (s *CronJobScheduler) Health(ctx context.Context) error {
	jbs, err := s.ICronService.GetScheduledCronJobs(ctx)
	if err != nil {
		return err
	}
	for _, job := range jbs {
		if err := uuid.Validate(job.Ref().GetID().String()); err != nil {
			return fmt.Errorf("invalid job ID format: %v", err)
		}
	}
	return nil
}

func (s *CronJobScheduler) Stats(ctx context.Context) map[string]any {
	jbs, err := s.ICronService.GetScheduledCronJobs(ctx)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	stats := make([]map[string]any, 0)
	for _, job := range jbs {
		jobStats := map[string]any{
			"id":   job.Ref().GetID(),
			"name": job.Ref().GetName(),
			// "schedule":   job.Ref().GetSchedule(),
			// "next_run":   job.Ref().GetNextRun(),
			// "status":     job.Ref().GetStatus(),
			// "progress":   job.Ref().GetProgress(),
			// "created_at": job.Ref().GetCreatedAt(),
			// "updated_at": job.Ref().GetUpdatedAt(),
		}
		stats = append(stats, jobStats)
	}
	return map[string]any{"jobs": stats}
}

func (s *CronJobScheduler) ListScheduledJobs(ctx context.Context) ([]types.Job, error) {
	scj, err := s.ICronService.GetScheduledCronJobs(ctx)
	if err != nil {
		return nil, err
	}
	return scj, nil
}
