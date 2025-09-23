// Pacote de teste externo para internal/contracts/types (Property).
package types_test

import (
	"path/filepath"
	"reflect"
	"testing"

	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

func TestProperty_GetSetSerialize(t *testing.T) {
	val := 0
	p := types.NewProperty("counter", &val, false, nil).(*types.Property[int])
	if p.GetName() != "counter" {
		t.Fatalf("expected name 'counter'")
	}
	if v := p.GetValue(); v != 0 {
		t.Fatalf("expected zero value initially, got %d", v)
	}

	val = 42
	p.Prop().Set(&val) // Usando Prop().Set() em vez de p.SetValue(&val)

	//p.SetValue(&val)
	pt := p
	var v int

	if v = pt.GetValue(); v != 42 {
		gl.Log("error", "Value from Property:", v, "Type:", reflect.TypeOf(v).Name())
		if v = *pt.Prop().Get(true).(*int); v != 42 {
			gl.Log("error", "Value from Prop().Get(true):", v, "Type:", reflect.TypeOf(v).Name())
			if v = *pt.Prop().Value(); v != 42 {
				t.Fatalf("expected 42, got %d", v)
			} else {
				gl.Log("success", "Value from Prop().Value():", v, "Type:", reflect.TypeOf(v).Name())
				goto successfully
			}
			t.Fatalf("expected 42, got %d", v)
		} else {
			gl.Log("success", "Value from Prop().Get(true):", v, "Type:", reflect.TypeOf(v).Name())
			goto successfully
		}
		t.Fatalf("expected 42, got %d", v)
	}

successfully:
	// Serialize/Deserialize JSON em memória
	_, err := p.Serialize("json", "")
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}
	var v2 []byte
	p2 := types.NewProperty("counter", &v, false, nil)
	if err := p2.Deserialize(v2, "json", ""); err != nil {
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
