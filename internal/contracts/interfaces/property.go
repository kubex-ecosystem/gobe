package interfaces

import (
	"github.com/google/uuid"
	"github.com/kubex-ecosystem/logz"
)

// IProperty is an interface that defines the methods for a property.
type IProperty[T any] interface {
	GetName() string
	GetValue() T
	SetValue(v *T)
	GetReference() (uuid.UUID, string)
	Prop() IPropertyValBase[T]
	GetLogger() *logz.LoggerZ
	Serialize(format, filePath string) ([]byte, error)
	Deserialize(data []byte, format, filePath string) error
	SaveToFile(filePath string, format string) error
	LoadFromFile(filename, format string) error
	// Telemetry() *ITelemetry
}
