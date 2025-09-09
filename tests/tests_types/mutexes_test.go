// Pacote de teste externo para internal/contracts/types (Mutexes).
package types_test

import (
	"testing"
	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

func TestMutexes_LockAndTry(t *testing.T) {
	mu := types.NewMutexesType()
	// TryLock should succeed initially
	if ok := mu.MuTryLock(); !ok {
		t.Fatalf("expected MuTryLock to succeed on fresh mutex")
	}
	// Unlock should release
	mu.MuUnlock()
	// Lock/Unlock should not panic
	mu.MuLock()
	mu.MuUnlock()
}

func TestMutexes_WaitGroup(t *testing.T) {
	mu := types.NewMutexesType()
	mu.MuAdd(1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		mu.MuDone()
	}()
	mu.MuWait() // Must return
	<-done
}

func TestMutexes_SharedCtx_SetGet(t *testing.T) {
	mu := types.NewMutexesType()
	if mu.GetMuSharedCtx() != nil {
		t.Fatalf("expected nil shared ctx initially")
	}
	mu.SetMuSharedCtx("hello")
	if got := mu.GetMuSharedCtx(); got != "hello" {
		t.Fatalf("expected 'hello' shared ctx, got %v", got)
	}
}
