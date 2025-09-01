// Pacote de teste externo para o logger.
package logger_test

import (
	"testing"

	logger "github.com/rafa-mori/gobe/internal/module/logger"
	l "github.com/rafa-mori/logz"
)

// helper para detectar panic sem falhar o processo
func noPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	fn()
}

func TestGetLogger_NotNil(t *testing.T) {
	lg := logger.GetLogger[struct{}](nil)
	if lg == nil {
		t.Fatalf("expected non-nil logger")
	}
}

func TestSetDebugAndLog_NoPanic(t *testing.T) {
	noPanic(t, func() { logger.SetDebug(true) })
	noPanic(t, func() { logger.Log("debug", "debug message") })
	noPanic(t, func() { logger.SetDebug(false) })
	noPanic(t, func() { logger.Log("info", "info message") })
	noPanic(t, func() { logger.Log("warn", "warn message") })
	noPanic(t, func() { logger.Log("error", "error message") })
	// tipo inválido deve apenas registrar erro, não panicar
	noPanic(t, func() { logger.Log("INVALID_TYPE", "should not panic") })
}

func TestLogObjLogger_NoLoggerField_NoPanic(t *testing.T) {
	type S struct { /* sem campo Logger */
	}
	s := &S{}
	noPanic(t, func() { logger.LogObjLogger(s, "info", "hello world") })
}

func TestNewLogger_Wrapper_NoPanic(t *testing.T) {
	l := logger.NewLogger[l.Logger]("test")
	if l == nil {
		t.Fatalf("expected NewLogger to return instance")
	}
	noPanic(t, func() { l.Log("Info", "info") })
	noPanic(t, func() { l.Log("Debug", "debug") })
	noPanic(t, func() { l.Log("Warn", "warn") })
	noPanic(t, func() { l.Log("Error", "error") })
	noPanic(t, func() { l.Log("Notice", "notice") })
}
