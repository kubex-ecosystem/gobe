package screening

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
	"unicode"
)

// Action represents the decision suggested by the screening engine.
type Action string

const (
	ActionNone         Action = "NONE"
	ActionDuplicate    Action = "DUPLICATE"
	ActionReplyStatus  Action = "REPLY_STATUS"
	ActionContinue     Action = "CONTINUE"
	ActionExecute      Action = "EXECUTE"
	ActionClarify      Action = "CLARIFY"
	ActionResetSession Action = "RESET_SESSION"
	ActionAbortSession Action = "ABORT_SESSION"
	ActionAcknowledge  Action = "ACK"
	ActionIgnore       Action = "IGNORE"
	ActionPrompt       Action = "PROMPT"
)

// Decision stores the outcome of a screening pass.
type Decision struct {
	Detected
	Action      Action
	Duplicate   bool
	Fingerprint string
	Sanitized   string
	ObservedAt  time.Time
	Clipped     bool
}

// Config tweaks the behaviour of the Engine.
type Config struct {
	DuplicateWindow time.Duration
	MaxBodyLen      int
	MinSignalTokens int
}

// Engine classifies messages using heuristics and optional session context.
type Engine struct {
	cfg Config
}

// NewEngine creates a new Engine instance with sane defaults.
func NewEngine(cfg Config) *Engine {
	if cfg.DuplicateWindow <= 0 {
		cfg.DuplicateWindow = 20 * time.Second
	}
	if cfg.MaxBodyLen <= 0 {
		cfg.MaxBodyLen = 4096
	}
	if cfg.MinSignalTokens <= 0 {
		cfg.MinSignalTokens = 1
	}
	return &Engine{cfg: cfg}
}

// Analyze inspects the message and returns a decision describing the suggested action.
func (e *Engine) Analyze(msg string, ctx Context) Decision {
	observed := ctx.Now
	if observed.IsZero() {
		observed = time.Now()
	}
	sanitized, clipped := sanitize(msg, e.cfg.MaxBodyLen)
	detected := DetectIntent(sanitized, ctx)
	fingerprint := fingerprint(sanitized)
	duplicate := e.isDuplicate(fingerprint, ctx, observed)
	if duplicate {
		detected.Reasons = append(detected.Reasons, "duplicate")
	}

	// downgrade intent to unknown if we have almost no signal (e.g., 1 short token).
	if detected.Intent == IntentCommand || detected.Intent == IntentQuestion {
		if tokenCount(sanitized) < e.cfg.MinSignalTokens {
			detected.Intent = IntentUnknown
			detected.Confidence = 0.2
			detected.Reasons = append(detected.Reasons, "insufficient_tokens")
		}
	}

	action := e.decideAction(detected.Intent, ctx, duplicate)

	return Decision{
		Detected:    detected,
		Action:      action,
		Duplicate:   duplicate,
		Fingerprint: fingerprint,
		Sanitized:   sanitized,
		ObservedAt:  observed,
		Clipped:     clipped,
	}
}

func (e *Engine) isDuplicate(hash string, ctx Context, now time.Time) bool {
	if hash == "" || ctx.LastMessageHash == "" {
		return false
	}
	if hash != ctx.LastMessageHash {
		return false
	}
	if ctx.LastMessageUnix == 0 {
		return true
	}
	last := time.Unix(ctx.LastMessageUnix, 0)
	return now.Sub(last) <= e.cfg.DuplicateWindow
}

func (e *Engine) decideAction(intent Intent, ctx Context, duplicate bool) Action {
	if duplicate {
		return ActionDuplicate
	}

	switch intent {
	case IntentStatus:
		return ActionReplyStatus
	case IntentContinue:
		return ActionContinue
	case IntentCommand, IntentQuestion:
		return ActionExecute
	case IntentClarify:
		return ActionClarify
	case IntentReset:
		return ActionResetSession
	case IntentStop:
		return ActionAbortSession
	case IntentAck:
		if isActiveState(ctx.LastBotState) {
			return ActionReplyStatus
		}
		return ActionAcknowledge
	case IntentSmalltalk:
		if isActiveState(ctx.LastBotState) {
			return ActionReplyStatus
		}
		return ActionIgnore
	case IntentUnknown:
		return ActionPrompt
	default:
		return ActionNone
	}
}

func sanitize(msg string, maxLen int) (string, bool) {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return "", false
	}
	var b strings.Builder
	b.Grow(len(trimmed))
	for _, r := range trimmed {
		if r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(' ')
			continue
		}
		if unicode.IsControl(r) && !unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
	}
	res := strings.TrimSpace(strings.Join(strings.Fields(b.String()), " "))
	clipped := false
	if maxLen > 0 && len(res) > maxLen {
		res = res[:maxLen]
		clipped = true
	}
	return res, clipped
}

func fingerprint(msg string) string {
	if msg == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(msg))))
	return hex.EncodeToString(sum[:])
}

func tokenCount(msg string) int {
	return len(strings.Fields(msg))
}

func isActiveState(state string) bool {
	s := strings.ToUpper(strings.TrimSpace(state))
	return s == "WORKING" || s == "PENDING"
}
