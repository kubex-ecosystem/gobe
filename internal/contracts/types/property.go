package types

import (
	"os"
	"reflect"

	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	l "github.com/kubex-ecosystem/logz"
	gl "github.com/kubex-ecosystem/logz/logger"

	"github.com/google/uuid"
)

// Property is a struct that holds the properties of the GoLife instance.
type Property[T any] struct {
	// Telemetry is the telemetry for this GoLife instance.
	metrics *Telemetry `json:"-" yaml:"-" toml:"-" xml:"-"`
	// Prop is the property for this GoLife instance.
	prop ci.IPropertyValBase[T] `json:"-" yaml:"-" toml:"-" xml:"-"`
	// Cb is the callback function for this GoLife instance.
	cb func(any) (bool, error) `json:"-" yaml:"-" toml:"-" xml:"-"`
}

// NewProperty creates a new IProperty[T] with the given value and Reference.
func NewProperty[T any](name string, v *T, withMetrics bool, cb func(any) (bool, error)) ci.IProperty[T] {
	p := &Property[T]{
		prop: newVal(name, v),
		cb:   cb,
	}
	if withMetrics {
		p.metrics = NewTelemetry()
	}
	return p
}

// GetName returns the name of the property.
func (p *Property[T]) GetName() string {
	return p.Prop().GetName()
}

// GetValue returns the value of the property.
func (p *Property[T]) GetValue() T {
	value := p.Prop().Value()
	if value == nil {
		value = new(T)
	}
	return *value
}

// SetValue sets the value of the property.
func (p *Property[T]) SetValue(v *T) {
	p.Prop().Set(v)
	if p.cb != nil {
		if _, err := p.cb(v); err != nil {
			gl.Log("error", "Callback function returned an error:", err.Error())
			//p.metrics.Log("error", "Error in callback function: "+err.Error())
		}
	}
}

// GetReference returns the reference of the property.
func (p *Property[T]) GetReference() (uuid.UUID, string) {
	return p.Prop().GetID(), p.Prop().GetName()
}

// Prop is a struct that holds the properties of the GoLife instance.
func (p *Property[T]) Prop() ci.IPropertyValBase[T] {
	if p.prop == nil {
		p.prop = newVal(p.GetName(), new(T))
	}
	return p.prop
}

// GetLogger returns the logger of the property.
func (p *Property[T]) GetLogger() l.Logger {
	return p.Prop().GetLogger()
}

// Serialize serializes the ProcessInput instance to the specified format.
func (p *Property[T]) Serialize(format, filePath string) ([]byte, error) {
	value := p.GetValue()
	mapper := NewMapper(&value, filePath)
	return mapper.Serialize(format)
}

// Deserialize deserializes the data into the ProcessInput instance.
func (p *Property[T]) Deserialize(data []byte, format, filePath string) error {
	if len(data) == 0 {
		return nil
	}
	value := p.GetValue()
	if !reflect.ValueOf(value).IsValid() {
		p.SetValue(new(T))
	}
	var v T
	mapper := NewMapper(&v, filePath)
	if v, vErr := mapper.Deserialize(data, format); vErr != nil {
		gl.Log("error", "Failed to deserialize data:", vErr.Error())
		return vErr
	} else {
		p.SetValue(v)
	}
	return nil
}

// SaveToFile saves the property to a file in the specified format.
func (p *Property[T]) SaveToFile(filePath string, format string) error {
	m := NewMapper(p.Prop().Value(), filePath)
	m.SerializeToFile(format)
	// if _, err := os.Stat(filePath); os.IsNotExist(err) {
	// 	gl.Log("error", "File was not created:", filePath)
	// 	return err
	// }
	// return nil
	return nil
}

// LoadFromFile loads the property from a file in the specified format.
func (p *Property[T]) LoadFromFile(filename, format string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		gl.Log("error", "File does not exist:", filename)
		return err
	}
	// Create a new mapper instance
	m := NewMapper(p.Prop().Value(), filename)

	// Deserialize the file content into the property (this may fail for some formats),
	// so we check if data was filled, if not, we will try to read the file normally
	// and then deserialize the data into the property with the mapper.
	data, err := m.DeserializeFromFile(format)
	if data == nil {
		if err != nil && os.IsPermission(err) {
			gl.Log("error", "Permission denied to read file:", filename)
			return err
		}
		// We already checked if the file exists, so let's read it
		var v []byte
		if v, err = os.ReadFile(filename); err != nil {
			gl.Log("error", "Failed to read file:", err.Error())
			return err
		} else {
			data, err = m.Deserialize(v, format)
			if err != nil {
				gl.Log("error", "Failed to deserialize file content:", err.Error())
				return err
			}
			p.Prop().Set((*T)(data))
		}
	}
	return nil
}
