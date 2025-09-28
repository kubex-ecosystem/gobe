// Package types implements various data types and structures used across the application,
// including database configurations and JSONB handling.
package types

import (
	"database/sql/driver"
	"encoding/json"
)

// IJSONB é uma interface que define os métodos necessários para tipos que podem ser
// serializados e desserializados como JSONB em um banco de dados. Ela representa o contrato
// mínimo original usado pelo GORM para lidar com o tipo JSONB do PostgreSQL.
type IJSONB interface {
	Value() (driver.Value, error)
	Scan(vl any) error
}

// JSONBData é uma interface que estende IJSONB e adiciona métodos para manipulação de dados JSONB.
type JSONBData interface {
	Scan(vl any) error
	Value() (driver.Value, error)
	ToMap() map[string]any
	ToInterface() any
	IsNil() bool
	IsEmpty() bool
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
	Has(key string) bool
	Keys() []string
	Values() []any
	Len() int
	Clear()
}

// JSONB é uma implementação concreta da interface JSONBData, representando um objeto JSONB.
// Internamente, é um map[string]any que pode ser manipulado diretamente. Ela também implementa
// os métodos necessários para serialização e desserialização com o GORM (interface IJSONB).
type JSONB map[string]any

// newJSONB cria uma nova instância de JSONB inicializada como um map vazio.
func newJSONB() *JSONB {
	m := make(map[string]any)
	j := JSONB(m)
	return &j
}

// NewJSONBData cria uma nova instância de JSONBData.
func NewJSONBData() JSONBData { return newJSONB() }

// Value is a Serializer manual para o GORM
func (m JSONB) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan is a Deserializer manual para o GORM
func (m *JSONB) Scan(vl any) error {
	if vl == nil {
		*m = JSONB{}
		return nil
	}
	return json.Unmarshal(vl.([]byte), m)
}

// Métodos adicionais para manipulação de dados JSONB

func (m JSONB) ToMap() map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	return map[string]any(m)
}
func (m JSONB) ToInterface() any {
	if m == nil {
		return map[string]any{}
	}
	return any(m)
}
func (m JSONB) IsNil() bool {
	return m == nil
}
func (m JSONB) IsEmpty() bool {
	return len(m) == 0
}
func (m JSONB) Get(key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}
func (m *JSONB) Set(key string, value any) {
	if *m == nil {
		*m = make(map[string]any) // Cria o map no valor apontado
	}
	(*m)[key] = value // Acessa e modifica o map subjacente
}
func (m JSONB) Delete(key string) {
	if m == nil {
		return
	}
	delete(m, key)
}
func (m JSONB) Has(key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}
func (m JSONB) Keys() []string {
	if m == nil {
		return []string{}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
func (m JSONB) Values() []any {
	if m == nil {
		return []any{}
	}
	values := make([]any, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
func (m JSONB) Len() int {
	if m == nil {
		return 0
	}
	return len(m)
}
func (m *JSONB) Clear() {
	if *m == nil {
		return
	}
	*m = make(map[string]any) // Zera o map na instância original
}

func JSONBFromMap(m map[string]any) JSONB {
	return JSONB(m)
}
func JSONBFromInterface(v any) JSONB {
	if v == nil {
		return JSONB{}
	}
	return JSONB(v.(map[string]any))
}
