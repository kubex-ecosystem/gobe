package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kubex-ecosystem/gobe/internal/module/kbx"
	logzprom "github.com/kubex-ecosystem/logz/api/integrations"
	logz "github.com/kubex-ecosystem/logz/api/notifiers"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// prom is a tiny shim over your PrometheusManager.
// We emulate counters/histograms using gauges and manual aggregation.
type promShim struct {
	pm *logzprom.PrometheusManager
}

var Prom *promShim

func InitPromShim() {
	pm := logzprom.GetPrometheusManager()

	port := os.Getenv("LOGZ_PROM_PORT")
	if port == "" {
		port = "2112"
	}
	if !pm.IsEnabled() {
		pm.Enable(port) // starts :2112/metrics
	}

	// Optional: restrict export to conductor_* (keeps /metrics clean)
	if wl := os.Getenv("LOGZ_PROM_WHITELIST"); wl == "conductor" {
		pm.SetExportWhitelist([]string{
			"conductor_intents_total",
			"conductor_fsm_total_accepted_planned",
			"conductor_fsm_total_planned_running",
			"conductor_fsm_total_running_done",
			"conductor_fsm_total_running_failed",
			"conductor_dispatch_latency_seconds_sum",
			"conductor_dispatch_latency_seconds_count",
			"conductor_dispatch_latency_seconds_avg",
		})
	}

	// Initialize default metrics
	pm.AddMetric("conductor_intents_total", 0, nil)
	pm.AddMetric("conductor_fsm_total_accepted_planned", 0, nil)
	pm.AddMetric("conductor_fsm_total_planned_running", 0, nil)
	pm.AddMetric("conductor_fsm_total_running_done", 0, nil)
	pm.AddMetric("conductor_fsm_total_running_failed", 0, nil)
	pm.AddMetric("conductor_dispatch_latency_seconds_sum", 0, nil)
	pm.AddMetric("conductor_dispatch_latency_seconds_count", 0, nil)
	pm.AddMetric("conductor_dispatch_latency_seconds_avg", 0, nil)

	Prom = &promShim{pm: pm}
}

// Counter-like
func (p *promShim) Inc(name string, delta float64) { p.pm.IncrementMetric(name, delta) }

// Histogram-like (we store sum/count and derive mean)
func (p *promShim) ObserveLatency(path string, seconds float64) {
	p.pm.IncrementMetric("conductor_dispatch_latency_seconds_sum", seconds)
	p.pm.IncrementMetric("conductor_dispatch_latency_seconds_count", 1)

	// recompute avg (best-effort; race-safe enough for MVP)
	m := p.pm.GetMetrics()
	sum := m["conductor_dispatch_latency_seconds_sum"]
	cnt := m["conductor_dispatch_latency_seconds_count"]
	if cnt > 0 {
		p.pm.AddMetric("conductor_dispatch_latency_seconds_avg", sum/cnt, map[string]string{"unit": "seconds", "since": time.Now().Format(time.RFC3339)})
	}
}

// EventLog registra e publica os eventos de FSM/intents no logger nativo do logz.
func EventLog(ev Event) {
	msg := fmt.Sprintf("[%s] %s (%s → %s)", ev.Kind, ev.Message, ev.From, ev.To)
	gl.Log("info", msg)

	// Opcional: publicar via WebSocket, se tiver WS ativo
	if ok := logz.NewDBusNotifier().EnabledFlag; ok {
		notifier := logz.NewDBusNotifier()
		hostname, _ := os.Hostname()
		pid := os.Getpid()
		entry := NewEntry().
			WithProcessID(pid).
			WithHostname(hostname).
			WithSource("conductor").
			WithTraceID(ev.IntentID.String()).
			WithLevel("info").
			AddMetadata("kind", ev.Kind).
			AddMetadata("intent_id", ev.IntentID.String()).
			AddMetadata("from", ev.From).
			AddMetadata("to", ev.To).
			AddMetadata("path", ev.Path).
			AddMetadata("method", ev.Method).
			AddMetadata("time", ev.Time.Format(time.RFC3339)).
			AddMetadata("host", hostname).
			AddMetadata("pid", pid).
			AddMetadata("app", "conductor").
			AddMetadata("env", kbx.GetEnvOrDefault("APP_ENV", "development")).
			AddMetadata("version", kbx.GetEnvOrDefault("APP_VERSION", "unknown")).
			AddMetadata("instance_id", kbx.GetEnvOrDefault("INSTANCE_ID", "local")).
			WithMessage(msg)

		notifier.Notify(entry)
	}
}

// RecordLatency sets the latency metric (in seconds).
func RecordLatency(seconds float64) {
	Prom.ObserveLatency("", seconds)
	gl.Log("debug", fmt.Sprintf("latency observed: %.4fs", seconds))
}

// LogIntentAccepted logs when an intent is accepted for processing.
func LogIntentAccepted(intentID, path, method string) {
	gl.Log("info", fmt.Sprintf("intent accepted: id=%s path=%s method=%s time=%s",
		intentID, path, method, time.Now().Format(time.RFC3339)))
}

func PublishEvent(ctx context.Context, ev Event) {
	EventLog(Event{
		Kind:     ev.Kind,
		IntentID: ev.IntentID,
		From:     ev.From,
		To:       ev.To,
		Path:     ev.Path,
		Method:   ev.Method,
		Message:  ev.Message,
		Time:     ev.Time,
	})
}
