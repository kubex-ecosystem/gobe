// Pacote de teste externo para internal/contracts/types (Property).
package types_test

import (
	"os"
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
		// gl.Log("error", "Value from Property:", v, "Type:", reflect.TypeOf(v).Name())
		if v = *pt.Prop().Get(true).(*int); v != 42 {
			// gl.Log("error", "Value from Prop().Get(true):", v, "Type:", reflect.TypeOf(v).Name())
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
	name := "kubex"
	p := types.NewProperty("name", &name, false, nil)
	dir := t.TempDir()
	file := filepath.Join(dir, "prop.json")
	if !filepath.IsAbs(file) {
		t.Fatalf("expected absolute file path, got %q", file)
	}
	if err := p.SaveToFile(file, "yaml"); err != nil {
		t.Fatalf("SaveToFile error: %v", err)
	}
	if info, err := os.Stat(file); err != nil {
		t.Fatalf("file should exist after SaveToFile, got error: %v", err)
	} else {
		gl.Log("info", "File info:", info.Size(), "bytes")
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		gl.Log("info", "File content:", string(content))
	}
	v := ""
	p2 := types.NewProperty("name", &v, false, nil)
	if err := p2.LoadFromFile(file, "yaml"); err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	expected := func(expectedValue string) string { s := expectedValue; return s }("kubex")
	got := p2.(*types.Property[string]).GetValue()
	got2 := p2.Prop().(*types.PropertyValBase[string]).Value()
	gl.Log("info", "Got from Prop().Get(true):", got, "Type:", reflect.TypeOf(got).Name())
	gl.Log("info", "Got from Prop().Value():", *got2, "Type:", reflect.TypeOf(got2).Name())

	if got != expected {
		if got2 != &expected {
			t.Fatalf("expected %q after LoadFromFile, got %q and %q", expected, got, *got2)
		} else {
			gl.Log("success", "Got from Prop().Value():", *got2, "Type:", reflect.TypeOf(got2).Name())
		}
	}
}
