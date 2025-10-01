// Package testtypes contains tests for the types package.
package testtypes

import (
	"reflect"
	"testing"

	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

func TestChannelBase_Basics(t *testing.T) {
	cb := types.NewChannelBase[string]("main", 2, nil)
	if cb == nil {
		t.Fatalf("expected channel base instance")
	}
	// Nome e buffers
	if name := cb.GetName(); name != "main" {
		t.Fatalf("expected name 'main', got %q", name)
	}
	if buf := cb.GetBuffers(); buf != 2 {
		t.Fatalf("expected buffers 2, got %d", buf)
	}
	// Tipo do canal
	if typ := cb.GetType(); typ != reflect.TypeFor[string]() {
		t.Fatalf("expected type string, got %v", typ)
	}
	// Ajustar buffers recria o canal
	cb.SetBuffers(4)
	if buf := cb.GetBuffers(); buf != 4 {
		t.Fatalf("expected buffers 4, got %d", buf)
	}
	// Clear/Close não devem panicar
	if err := cb.Clear(); err != nil {
		t.Fatalf("clear error: %v", err)
	}
	if err := cb.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
}
