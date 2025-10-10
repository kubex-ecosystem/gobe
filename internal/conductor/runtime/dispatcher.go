package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	scheduler "github.com/kubex-ecosystem/gobe/internal/services/scheduler"
	l "github.com/kubex-ecosystem/logz/api"
)

type Dispatcher struct {
	sched *scheduler.SchedulerImpl
}

type State string

const (
	StateAccepted State = "accepted"
	StatePlanned  State = "planned"
	StateRunning  State = "running"
	StateDone     State = "done"
	StateFailed   State = "failed"
)

type Entry = l.LogzEntry

func NewEntry() Entry {
	return l.NewLogEntry()
}

type Event struct {
	Kind     string
	IntentID uuid.UUID
	From     State
	To       State
	Path     string
	Method   string
	Message  string
	Schedule string
	Command  string
	Time     time.Time
}

func (e Event) ToMap() map[string]any {
	return map[string]any{
		"kind":      e.Kind,
		"intent_id": e.IntentID,
		"from":      e.From,
		"to":        e.To,
		"path":      e.Path,
		"method":    e.Method,
		"message":   e.Message,
		"schedule":  e.Schedule,
		"command":   e.Command,
		"time":      e.Time.Format(time.RFC3339),
		"timestamp": strconv.FormatInt(e.Time.UnixNano(), 10),
	}
}

func NewDispatcher(s *scheduler.SchedulerImpl) *Dispatcher {
	return &Dispatcher{sched: s}
}

func (d *Dispatcher) Dispatch(intent *Intent) error {
	start := time.Now()
	PublishEvent(intent.Context, Event{
		Kind:     "intent.accepted",
		IntentID: intent.ID,
		Path:     intent.Path,
		Method:   intent.Method,
		Time:     time.Now(),
	})

	PublishEvent(intent.Context, Event{
		Kind:     "fsm.transition",
		IntentID: intent.ID,
		From:     StateAccepted,
		To:       StatePlanned,
		Path:     intent.Path,
		Method:   intent.Method,
		Message:  "intent received",
		Time:     time.Now(),
	})

	cap := d.resolveCapability(intent.Path)
	switch cap {
	case "scheduler.job.run":
		err := d.dispatchToScheduler(intent)
		RecordLatency(time.Since(start).Seconds())
		return err
	case "webhook.receive":
		// pode plugar o webhooks aqui depois
		RecordLatency(time.Since(start).Seconds())
		return errors.New("webhook.receive not wired yet")
	default:
		RecordLatency(time.Since(start).Seconds())
		return fmt.Errorf("no capability mapped for %s", intent.Path)
	}
}

func (d *Dispatcher) dispatchToScheduler(intent *Intent) error {
	PublishEvent(intent.Context, Event{
		Kind:     "fsm.transition",
		IntentID: intent.ID,
		From:     StatePlanned,
		To:       StateRunning,
		Path:     intent.Path,
		Method:   intent.Method,
		Message:  "dispatching to scheduler",
		Time:     time.Now(),
	})

	job := scheduler.NewJobImpl(
		// id int,
		intent.ID,
		// name string,
		intent.Path,
		// schedule string,
		intent.Schedule,
		// command string,
		intent.Command,
	)

	if err := d.sched.ScheduleJob(job); err != nil {
		PublishEvent(intent.Context, Event{
			Kind:     "fsm.transition",
			IntentID: intent.ID,
			From:     StateRunning,
			To:       StateFailed,
			Path:     intent.Path,
			Method:   intent.Method,
			Message:  err.Error(),
			Time:     time.Now(),
		})
		return err
	}

	PublishEvent(intent.Context, Event{
		Kind:     "fsm.transition",
		IntentID: intent.ID,
		From:     StateRunning,
		To:       StateDone,
		Path:     intent.Path,
		Method:   intent.Method,
		Message:  "job completed (scheduled/queued)",
		Time:     time.Now(),
	})
	return nil
}

func (d *Dispatcher) resolveCapability(path string) string {
	switch {
	case strings.Contains(path, "/jobs"), strings.Contains(path, "/tasks"):
		return "scheduler.job.run"
	case strings.Contains(path, "/webhook"):
		return "webhook.receive"
	default:
		return ""
	}
}
