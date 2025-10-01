// Package testmanifest contains tests for the manifest package.
package testmanifest

import (
	"testing"

	manifest "github.com/kubex-ecosystem/gobe/internal/module/info"
)

func TestGetManifest_LoadsEmbeddedData(t *testing.T) {
	m, err := manifest.GetManifest()
	if err != nil {
		t.Fatalf("GetManifest() unexpected error: %v", err)
	}
	if m == nil {
		t.Fatalf("GetManifest() returned nil manifest")
	}
	if m.GetName() != "Kubex GoBE" {
		t.Fatalf("expected name 'Kubex GoBE', got %q", m.GetName())
	}
	if m.GetBin() != "gobe" {
		t.Fatalf("expected bin 'gobe', got %q", m.GetBin())
	}
	if ver := m.GetVersion(); ver == "" {
		t.Fatalf("expected version to be non-empty")
	}
}
