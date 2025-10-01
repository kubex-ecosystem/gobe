package screening

import (
	"testing"
	"time"
)

func TestStatusWithoutQuestion(t *testing.T) {
	d := DetectIntent("E agora", Context{LastBotState: "WORKING"})
	if d.Intent != IntentStatus || d.Confidence < 0.8 {
		t.Fatalf("expected STATUS_CHECK, got %+v", d)
	}
}

func TestContinue(t *testing.T) {
	d := DetectIntent("bora", Context{LastBotState: "IDLE"})
	if d.Intent != IntentContinue {
		t.Fatalf("expected CONTINUE, got %v", d.Intent)
	}
}

func TestAck(t *testing.T) {
	d := DetectIntent("ok", Context{})
	if d.Intent != IntentAck {
		t.Fatalf("expected ACK, got %v", d.Intent)
	}
}

func TestQuestion(t *testing.T) {
	d := DetectIntent("conseguiu algo?", Context{})
	if d.Intent != IntentQuestion {
		t.Fatalf("expected QUESTION, got %v", d.Intent)
	}
}

func TestReset(t *testing.T) {
	d := DetectIntent("quero uma nova sessao", Context{})
	if d.Intent != IntentReset {
		t.Fatalf("expected RESET, got %v", d.Intent)
	}
}

func TestStop(t *testing.T) {
	d := DetectIntent("pode cancelar por favor", Context{})
	if d.Intent != IntentStop {
		t.Fatalf("expected STOP, got %v", d.Intent)
	}
}

func TestEngineDuplicate(t *testing.T) {
	engine := NewEngine(Config{DuplicateWindow: 30 * time.Second})
	ctx := Context{
		LastBotState:    "WORKING",
		LastMessageHash: fingerprint("status"),
		LastMessageUnix: time.Now().Add(-10 * time.Second).Unix(),
		Now:             time.Now(),
	}
	decision := engine.Analyze("status", ctx)
	if !decision.Duplicate {
		t.Fatalf("expected duplicate detection, got %+v", decision)
	}
	if decision.Action != ActionDuplicate {
		t.Fatalf("expected ActionDuplicate, got %s", decision.Action)
	}
}

func TestEngineActions(t *testing.T) {
	engine := NewEngine(Config{})
	ctx := Context{LastBotState: "WORKING", Now: time.Now()}
	decision := engine.Analyze("bora", ctx)
	if decision.Action != ActionContinue {
		t.Fatalf("expected ActionContinue, got %s", decision.Action)
	}
	decision = engine.Analyze("nova", ctx)
	if decision.Action != ActionResetSession {
		t.Fatalf("expected ActionResetSession, got %s", decision.Action)
	}
	decision = engine.Analyze("pode cancelar", ctx)
	if decision.Action != ActionAbortSession {
		t.Fatalf("expected ActionAbortSession, got %s", decision.Action)
	}
}
