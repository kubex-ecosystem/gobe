package interfaces

import (
	"reflect"

	"github.com/google/uuid"
	"github.com/kubex-ecosystem/logz"
)

// IPropertyValBase is an interface that defines the methods for a property value.
type IPropertyValBase[T any] interface {
	GetLogger() *logz.LoggerZ
	GetID() uuid.UUID
	GetName() string
	Value() *T
	StartCtl() <-chan string
	Type() reflect.Type
	Get(async bool) any
	Set(t *T) bool
	Clear() bool
	IsNil() bool
	Serialize(format, filePath string) ([]byte, error)
	Deserialize(data []byte, format, filePath string) error
}
