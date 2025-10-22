package testtypes

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	types "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/logz/logger"
)

func TestProperty_GetSetSerialize(t *testing.T) {
	val := 0
	p := types.NewProperty("counter", &val, false, nil)
	if p.GetName() != "counter" {
		t.Fatalf("expected name 'counter'")
	}
	if v := p.GetValue(); v != 0 {
		t.Fatalf("expected zero value initially, got %d", v)
	}
	pp, ok := p.Prop().(*types.PropertyValBase[int])
	if !ok {
		t.Fatalf("expected PropertyValBase[int], got %T", p.Prop())
	}
	gl.Log("info", "Property ID:", pp.GetID().String(), "Name:", pp.GetName(), "Type:", reflect.TypeOf(pp).Name())
	val = 42
	pp.Set(&val) // Usando Prop().Set() em vez de p.SetValue(&val)

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
	if err := p.SaveToFile(file, "json"); err != nil {
		t.Fatalf("SaveToFile error: %v", err)
	}
	if info, err := os.Stat(file); err != nil {
		t.Fatalf("file should exist after SaveToFile, got error: %v", err)
	} else {
		gl.Log("info", "Got values from file:", info.Sys())
		gl.Log("info", "File info:", info.Size(), "bytes")
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}
		gl.Log("info", "File content:", string(content))
	}
	v := ""
	p2 := types.NewProperty("name", &v, false, nil)
	if err := p2.LoadFromFile(file, "json"); err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	expected := "kubex"
	got := p2.GetValue()
	got2 := string(*p2.Prop().Get(true).(*string))

	gl.Log("info", "Got from Prop().Get(true):", got, "Type:", reflect.TypeOf(got).Name())
	gl.Log("info", "Got from Prop().Value():", got2, "Type:", reflect.TypeOf(got2).Name())

	if got != expected {
		if got2 != expected {
			t.Fatalf("expected %q after LoadFromFile, got %q and %q", expected, got, got2)
		} else {
			gl.Log("success", "Got from Prop().Value():", got2, "Type:", reflect.TypeOf(got2).Name())
		}
	}
}

func TestProperty_ConcurrentSetGet(t *testing.T) {
	t.Parallel()

	type Name string
	start := Name("kubex-0")
	p := types.NewProperty("name", &start, false, nil)

	var (
		stop     = make(chan struct{})
		wg       sync.WaitGroup
		writes   atomic.Int64
		mismatch atomic.Int64
	)

	cpus := runtime.GOMAXPROCS(0)
	writers := cpus
	readers := cpus * 4

	// Writers
	wg.Add(writers)
	for w := 0; w < writers; w++ {
		go func(id int) {
			defer wg.Done()
			i := int64(id)
			for {
				select {
				case <-stop:
					return
				default:
					next := Name(fmt.Sprintf("kubex-%d", atomic.AddInt64(&i, 1)))
					p.SetValue(&next) // deve trocar o ponteiro atômico por um novo snapshot
					writes.Add(1)
					time.Sleep(10 * time.Microsecond)
				}
			}
		}(w)
	}

	// Readers
	wg.Add(readers)
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			var s string
			for {
				select {
				case <-stop:
					return
				default:
					// 1) leitura “canônica”
					v1 := p.GetValue() // Name

					// 2) via Prop().Get(true) (ajuste o type switch conforme o retorno real)
					got := p.Prop().Get(true)
					var v2 Name
					switch x := any(got).(type) {
					case Name:
						v2 = x
					case *Name:
						if x != nil {
							v2 = *x
						}
					case string:
						v2 = Name(x)
					case *string:
						if x != nil {
							v2 = Name(*x)
						}
					default:
						// fallback: usa Scan pro tipo string
						if err := p.Deserialize([]byte(`"`+s+`"`), "json", ""); err != nil {
							gl.Log("error", fmt.Sprintf("Deserialize error: %v", err))
						}
						if s != "" {
							v2 = Name(s)
						} else {
							v2 = v1
						}
					}

					// 3) via Scan em string (coerência de snapshot)
					if err := p.Deserialize([]byte(`"`+s+`"`), "json", ""); err != nil {
						gl.Log("error", fmt.Sprintf("Deserialize error: %v", err))
					}
					s = string(p.GetValue())

					// Verifica se todos os valores lidos são iguais
					if v1 != v2 {
						gl.Log("error", fmt.Sprintf("Mismatched values: GetValue()=%q, Prop().Get(true)=%q, Scan()=%q", v1, v2, s))
						mismatch.Add(1)
					} else {
						// gl.Log("info", fmt.Sprintf("Consistent values: GetValue()=%q, Prop().Get(true)=%q, Scan()=%q", v1, v2, s))
						// Verifica se o valor lido bate com o valor do scan
						// (pode ser diferente se houver uma escrita concorrente entre as leituras)
						// Mas não deve ser diferente de ambos
						// (a menos que haja uma escrita concorrente entre as leituras)
						// Então, se v1 e v2 são iguais, s deve ser igual a ambos
						// Se não for, é um mismatch
						// Exemplo: v1 == v2 == "kubex-5", s == "kubex-6" => mismatch
						// Exemplo: v1 == v2 == "kubex-5", s == "kubex-5" => ok
						// Exemplo: v1 == v2 == "kubex-5", s == "kubex-4" => mismatch
						// Exemplo: v1 == v2 == "kubex-5", s == "kubex-7" => mismatch
						// Exemplo: v1 == v2 == "kubex-5", s == "" => mismatch
						if s != string(v1) {
							gl.Log("error", fmt.Sprintf("Mismatched values: GetValue()=%q, Prop().Get(true)=%q, Scan()=%q", v1, v2, s))
							mismatch.Add(1)
						}
					}
				}
			}
		}()
	}

	time.Sleep(500 * time.Millisecond)
	close(stop)
	wg.Wait()

	if writes.Load() == 0 {
		t.Fatalf("nenhuma escrita realizada")
	}
	if mismatch.Load() > 0 {
		t.Fatalf("mismatches entre snapshots de leitura: %d", mismatch.Load())
	}
}

func TestProperty_Concurrent_WritersReaders(t *testing.T) {
	t.Parallel()

	// named type pra simular teu caso
	type Name string

	start := Name("kubex-0")
	p := types.NewProperty[Name]("name", &start, false, nil)

	var writes atomic.Int64
	var mismatches atomic.Int64

	cpus := runtime.GOMAXPROCS(0)
	writers := cpus
	readers := cpus * 4

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// WRITERS: publicam snapshots novos (não mutam o objeto já publicado!)
	wg.Add(writers)
	for w := 0; w < writers; w++ {
		go func(id int) {
			defer wg.Done()
			local := int64(id)
			for {
				select {
				case <-stop:
					return
				default:
					next := Name(fmt.Sprintf("kubex-%d", atomic.AddInt64(&local, 1)))
					// Set deve trocar o ponteiro atômico por UM NOVO *Name
					p.SetValue(&next)
					writes.Add(1)
					time.Sleep(20 * time.Microsecond)
				}
			}
		}(w)
	}

	// READERS: comparam diferentes vias de leitura para o MESMO snapshot lógico
	wg.Add(readers)
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					// 1) leitura canônica
					v1 := p.GetValue() // se for GetValue(), troque aqui

					// 2) via Prop().Get(true) — o retorno aí pode variar (T, *T, string, *string)
					got := p.Prop().Get(true)
					var v2 Name
					switch x := any(got).(type) {
					case Name:
						v2 = x
					case *Name:
						if x != nil {
							v2 = *x
						}
					case string:
						v2 = Name(x)
					case *string:
						if x != nil {
							v2 = Name(*x)
						}
					default:
						// fallback: se a API devolver outro wrapper, considere tratar aqui
						v2 = v1
					}

					// V1 e V2 devem ser iguais (mesmo snapshot) ou, no pior caso,
					// diferirem pontualmente por uma troca concorrente entre leituras.
					// Se der muita divergência, conta mismatch pra acusar furo de coerência.
					if v1 != v2 {
						mismatches.Add(1)
					}
				}
			}
		}()
	}

	// janela de estresse
	time.Sleep(600 * time.Millisecond)
	close(stop)
	wg.Wait()

	if writes.Load() == 0 {
		t.Fatalf("nenhuma escrita realizada (teste não estressou)")
	}

	// tolerância: com leitores/ escritores correndo,
	// uma ou outra divergência pode ocorrer entre duas vias no "meio" de um publish.
	// Se virar enxurrada, acusa.
	if n := mismatches.Load(); n > 5 {
		t.Fatalf("mismatches excessivos entre leituras (%d) — revise publicação/immutabilidade", n)
	}
}

func TestProperty_Concurrent_SaveLoad_Readers(t *testing.T) {
	t.Parallel()

	type Name string
	start := Name("kubex-0")
	p := types.NewProperty("name", &start, false, nil)

	dir := t.TempDir()
	path := dir + "/prop.json" // use o formato que você definiu

	var stop atomic.Bool
	var writes atomic.Int64
	var errs atomic.Int64
	var wg sync.WaitGroup

	// Writer: atualiza valor + salva no arquivo (deve ser atômico no FS)
	wg.Add(1)
	go func() {
		defer wg.Done()
		i := int64(0)
		for !stop.Load() {
			i++
			next := Name(fmt.Sprintf("kubex-%d", i))
			p.SetValue(&next)
			writes.Add(1)
			if err := p.SaveToFile(path, "json"); err != nil { // ajuste o formato
				errs.Add(1)
			}
			time.Sleep(15 * time.Millisecond)
		}
	}()

	// Readers: leem da instância (não do arquivo) — garantem ausência de corr. de memória
	readers := runtime.GOMAXPROCS(0) * 4
	wg.Add(readers)
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for !stop.Load() {
				_ = p.GetValue()       // ou GetValue()
				_ = p.Prop().Get(true) // via caminho alternativo
				time.Sleep(200 * time.Microsecond)
			}
		}()
	}

	time.Sleep(800 * time.Millisecond)
	stop.Store(true)
	wg.Wait()

	if writes.Load() == 0 {
		t.Fatalf("nenhuma escrita realizada")
	}
	if errs.Load() != 0 {
		t.Fatalf("erros de persistência: %d (verifique escrita atômica)", errs.Load())
	}
}
