package interfaces

import (
	"github.com/google/uuid"
	l "github.com/rafa-mori/logz"
)

// IProperty is an interface that defines the methods for a property.
type IProperty[T any] interface {
	GetName() string
	GetValue() T
	SetValue(v *T)
	GetReference() (uuid.UUID, string)
	Prop() IPropertyValBase[T]
	GetLogger() l.Logger
	Serialize(format, filePath string) ([]byte, error)
	Deserialize(data []byte, format, filePath string) error
	SaveToFile(filePath string, format string) error
	LoadFromFile(filename, format string) error
	// Telemetry() *ITelemetry
}
