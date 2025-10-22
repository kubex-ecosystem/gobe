// Package scheduler provides an interface for scheduling jobs.
package scheduler

import (
	"context"

	"github.com/google/uuid"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/cron"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/manager"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/monitor"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/services"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"

	gl "github.com/kubex-ecosystem/logz/logger"
)

type JobImpl = types.JobImpl
type Job = types.Job

func NewJobImpl(id uuid.UUID, name, schedule, command string) *JobImpl {
	return types.NewJobImpl(id, name, schedule, command)
}

func NewJob(id uuid.UUID, name, schedule, command string) Job {
	return types.NewJob(id, name, schedule, command)
}

type JobStatus = types.JobStatus
type JobStatusType = types.JobStatusType

type JobStatusResponse = types.JobStatusResponse

// IScheduler Facade interface (context-aware, pointers)
type IScheduler interface {
	// ScheduleJob enfileira/agenda um job e retorna o status inicial (ou id)
	ScheduleJob(ctx context.Context, job types.Job) (JobStatusResponse, error)

	// CancelJob remove um job
	CancelJob(ctx context.Context, jobID uuid.UUID) error

	// GetJobStatus retorna o status
	GetJobStatus(ctx context.Context, jobID uuid.UUID) (types.JobStatus, error)

	// ListScheduledJobs lista todos (pointers por consistência)
	ListScheduledJobs(ctx context.Context) ([]types.Job, error)

	// RescheduleJob atualiza o spec cron
	RescheduleJob(ctx context.Context, jobID string, newSchedule string) error

	// Health utilitários (opcional)
	Health(ctx context.Context) error

	// Stats utilitários (opcional)
	Stats(ctx context.Context) map[string]any
}

type CronJobSchedulerManager interface {
	// Start lifecycle management
	StartScheduler(ctx context.Context) error

	// Stop lifecycle management
	StopScheduler(ctx context.Context) error
}

type SchedulerImpl = manager.Scheduler

func NewScheduler(ctx context.Context, pool *services.GoroutinePool, cronService services.ICronService) *SchedulerImpl {
	sched := NewSchedulerFunc(pool, cronService)
	if err := sched.StartScheduler(ctx); err != nil {
		gl.Log("error", "failed to start scheduler service", err)
		return nil
	}
	sch, ok := sched.(*SchedulerImpl)
	if !ok {
		gl.Log("error", "failed to cast sched to SchedulerImpl")
		return nil
	}
	return sch
}

func NewSchedulerFunc(pool *services.GoroutinePool, cronService services.ICronService) CronJobSchedulerManager {
	cronSchedulerManager := manager.NewCronJobScheduler(pool, cronService)
	cjs, ok := any(cronSchedulerManager).(*manager.CronJobScheduler)
	if !ok {
		gl.Log("error", "failed to cast cronSchedulerManager to SchedulerImpl")
		return nil
	}
	return cjs
}

type MonitorImpl = monitor.Metrics

func GetMetrics() monitor.Metrics {
	return monitor.GetMetrics()
}

func PreLaunchChecks() error {
	return monitor.PreLaunchChecks()
}

type Cron = cron.Cron
type CronService = cron.Cron
type CronParser = cron.Parser
type CronSchedule = cron.Schedule

type CronEntry = cron.Entry
type CronChain = cron.Chain
type CronJob = cron.Job
type CronJobFunc = cron.JobWrapper

func NewCronParser() cron.Parser {
	return cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
}
func ParseCronSpec(spec string) (cron.Schedule, error) {
	return cron.ParseStandard(spec)
}
func NewCron(opts ...cron.Option) *cron.Cron {
	return cron.New(opts...)
}
func NewCronChain(c ...cron.JobWrapper) cron.Chain {
	return cron.NewChain(c...)
}
func NewCronParserWithOptions(options cron.ParseOption) cron.Parser {
	return cron.NewParser(options)
}

type CronServiceImpl = services.CronServiceImpl
type ICronService = services.ICronService

func NewCronService(db types.DirectDatabase) services.ICronService {
	return services.NewCronService(db)
}

type GoroutinePool = services.GoroutinePool

func NewGoroutinePool(maxWorkers int) *GoroutinePool {
	return services.NewGoroutinePool(maxWorkers)
}
