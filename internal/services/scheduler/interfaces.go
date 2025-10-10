// Package scheduler provides an interface for scheduling jobs.
package scheduler

import (
	"github.com/google/uuid"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/cron"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/manager"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/monitor"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/services"
	"github.com/kubex-ecosystem/gobe/internal/services/scheduler/types"
)

type JobImpl = types.Job
type Job = types.IJob

func NewJobImpl(id uuid.UUID, name, schedule, command string) *JobImpl {
	return types.NewJobImpl(id, name, schedule, command)
}

func NewJob(id uuid.UUID, name, schedule, command string) Job {
	return types.NewJob(id, name, schedule, command)
}

type JobStatus = types.JobStatus
type JobStatusType = types.JobStatusType

type JobStatusResponse = types.JobStatusResponse

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

type SchedulerImpl = manager.Scheduler

func NewSchedulerFunc(pool *services.GoroutinePool, cronService services.ICronService) *manager.CronJobScheduler {
	return manager.NewCronJobScheduler(pool, cronService)
}

type MonitorImpl = monitor.Metrics

func GetMetrics() monitor.Metrics {
	return monitor.GetMetrics()
}

func PreLaunchChecks() error {
	return monitor.PreLaunchChecks()
}

type Cron = cron.Cron
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
