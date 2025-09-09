// Pacote de teste externo para internal/contracts/types (Property).
package types_test

import (
	"path/filepath"
	"testing"

	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

func TestProperty_GetSetSerialize(t *testing.T) {
	p := types.NewProperty[int]("counter", nil, false, nil)
	if p.GetName() != "counter" {
		t.Fatalf("expected name 'counter'")
	}
	if v := p.GetValue(); v != 0 {
		t.Fatalf("expected zero value initially, got %d", v)
	}

	val := 42
	p.SetValue(&val)
	pt := p.(*types.Property[int])

	if v := pt.GetValue(); v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}

	// Serialize/Deserialize JSON em memória
	data, err := p.Serialize("json", "")
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	p2 := types.NewProperty[int]("counter", nil, false, nil)
	if err := p2.Deserialize(data, "json", ""); err != nil {
		t.Fatalf("deserialize error: %v", err)
	}
	if v := p2.GetValue(); v != 42 {
		t.Fatalf("expected 42 after deserialize, got %d", v)
	}
}

func TestProperty_SaveLoadFile(t *testing.T) {
	p := types.NewProperty[string]("name", nil, false, nil)
	name := "kubex"
	p.SetValue(&name)

	dir := t.TempDir()
	file := filepath.Join(dir, "prop.json")
	if err := p.SaveToFile(file, "json"); err != nil {
		t.Fatalf("SaveToFile error: %v", err)
	}

	p2 := types.NewProperty[string]("name", nil, false, nil)
	if err := p2.LoadFromFile(file, "json"); err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	if got := p2.GetValue(); got != "kubex" {
		t.Fatalf("expected 'kubex', got %q", got)
	}
}
