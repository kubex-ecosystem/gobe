package runtime

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	scheduler "github.com/kubex-ecosystem/gobe/internal/services/scheduler"
	l "github.com/kubex-ecosystem/logz/api"
)

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
	Meta     map[string]any
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
		"meta":      e.Meta,
	}
}

type Dispatcher struct {
	sched scheduler.IScheduler
}

func NewDispatcher(s scheduler.IScheduler) *Dispatcher {
	return &Dispatcher{sched: s}
}

func (d *Dispatcher) Dispatch(intent *Intent) error {
	start := time.Now()

	// accepted → planned
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

	switch cap := d.resolveCapability(intent.Path); cap {
	case "scheduler.job.run":
		err := d.dispatchToScheduler(intent)
		RecordLatency(time.Since(start).Seconds())
		return err
	case "webhook.receive":
		// placeholder: encaixe futuro do webhooks
		RecordLatency(time.Since(start).Seconds())
		return errors.New("webhook.receive not wired yet")
	default:
		RecordLatency(time.Since(start).Seconds())
		return fmt.Errorf("no capability mapped for %s", intent.Path)
	}
}

func (d *Dispatcher) dispatchToScheduler(intent *Intent) error {
	// planned → running
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

	// Extrair campos esperados do body (defensivo)
	var (
		spec    string
		command string
	)
	if v, ok := intent.Body["schedule"].(string); ok {
		spec = v
	}
	if v, ok := intent.Body["command"].(string); ok {
		command = v
	}

	// Criar Job (UUID novo — não depende do formato do Intent.ID)
	jobID := uuid.New()
	job := scheduler.NewJobImpl(jobID, "intent:"+intent.Path, spec, command)

	// Chamar scheduler (ctx-aware, retorna JobStatusResponse)
	resp, err := d.sched.ScheduleJob(intent.Context, job)
	if err != nil {
		// running → failed
		PublishEvent(intent.Context, Event{
			Kind:     "fsm.transition",
			IntentID: intent.ID,
			From:     StateRunning,
			To:       StateFailed,
			Path:     intent.Path,
			Method:   intent.Method,
			Message:  err.Error(),
			Time:     time.Now(),
			Meta: map[string]any{
				"job_id": jobID.String(),
			},
		})
		return err
	}

	// running → done (agendado/enfileirado com sucesso)
	PublishEvent(intent.Context, Event{
		Kind:     "fsm.transition",
		IntentID: intent.ID,
		From:     StateRunning,
		To:       StateDone,
		Path:     intent.Path,
		Method:   intent.Method,
		Message:  "job accepted by scheduler",
		Time:     time.Now(),
		Meta: map[string]any{
			"job_id":  resp.JobID,
			"status":  resp.Status.Code,
			"message": resp.Status.Message,
		},
	})
	return nil
}

func (d *Dispatcher) resolveCapability(path string) string {
	// simples e direto; pode trocar por uma tabela DCL/YAML depois
	if contains(path, "/jobs") || contains(path, "/tasks") {
		return "scheduler.job.run"
	}
	if contains(path, "/webhook") {
		return "webhook.receive"
	}
	return ""
}

// contains: helper local pra evitar import de strings só pra isso.
func contains(s, sub string) bool {
	// implementação trivial (pode substituir por strings.Contains se preferir)
	n := len(sub)
	if n == 0 {
		return true
	}
	for i := 0; i+n <= len(s); i++ {
		if s[i:i+n] == sub {
			return true
		}
	}
	return false
}
