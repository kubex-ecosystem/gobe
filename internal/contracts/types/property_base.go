package types

import (
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"

	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/google/uuid"
)

type Pointer[T any] = atomic.Pointer[T]

// PropertyValBase is a type for the value.
type PropertyValBase[T any] struct {
	// Logger is the logger for this context.
	Logger l.Logger `json:"-" yaml:"-" toml:"-" xml:"-"`

	// v is the value.
	*Pointer[T] `json:"-" yaml:"-" toml:"-" xml:"-"`

	// Reference is the identifiers for the context.
	// IReference
	*Reference `mapstructure:",squash,omitempty"`

	//muCtx is the mutexes for the context.
	*Mutexes `mapstructure:",squash,omitempty"`

	// validation is the validation for the value.
	*Validation[T] `json:"-" yaml:"-" toml:"-" xml:"-"`
	// Channel is the channel for the value.
	ci.IChannelCtl[T] `json:"-" yaml:"-" toml:"-" xml:"-"`
	channelCtl        *ChannelCtl[T] `json:"-" yaml:"-" toml:"-" xml:"-"`
}

// NewVal is a function that creates a new PropertyValBase instance.
func newVal[T any](name string, v *T) *PropertyValBase[T] {
	ref := NewReference(name)

	// Create a new PropertyValBase instance
	vv := atomic.Pointer[T]{}
	if v != nil {
		vv.Store(v)
	} else {
		vv.Store(new(T))
	}

	// Create a new mutexes instance
	mu := NewMutexesType()

	// Create a new validation instance
	validation := newValidation[T]()
	validators := validation.GetValidators()
	for _, validator := range validators {
		validation.RemoveValidator(validator.GetPriority())
	}
	gl.Log("debug", "Created new PropertyValBase instance for:", name, "ID:", ref.GetID().String())

	return &PropertyValBase[T]{
		Pointer:    &vv,
		Validation: validation,
		Reference:  ref.GetReference(),
		channelCtl: NewChannelCtl[T](name, nil).(*ChannelCtl[T]),
		Mutexes:    mu,
	}
}

func NewVal[T any](name string, v *T) ci.IPropertyValBase[T] { return newVal(name, v) }

// GetLogger is a method that returns the logger for the value.
func (v *PropertyValBase[T]) GetLogger() l.Logger {
	if v == nil {
		gl.Log("error", "GetLogger: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return nil
	}
	return v.Logger
}

// GetName is a method that returns the name of the value.
func (v *PropertyValBase[T]) GetName() string {
	if v == nil {
		gl.Log("error", "GetName: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return ""
	}
	return v.Name
}

// GetID is a method that returns the ID of the value.
func (v *PropertyValBase[T]) GetID() uuid.UUID {
	if v == nil {
		gl.Log("error", "GetID: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return uuid.Nil
	}
	return v.ID
}

// Value is a method that returns the value.
func (v *PropertyValBase[T]) Value() *T {
	if v == nil {
		gl.Log("error", "Value: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return nil
	}
	return v.Pointer.Load()
}

// StartCtl is a method that starts the control channel.
func (v *PropertyValBase[T]) StartCtl() <-chan string {
	if v == nil {
		gl.Log("error", "StartCtl: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return nil
	}
	if v.channelCtl == nil {
		gl.Log("error", "StartCtl: channel control is nil (", reflect.TypeFor[T]().String(), ")")
		return nil
	}
	return v.channelCtl.Channels["ctl"].(<-chan string)
}

// Type is a method that returns the type of the value.
func (v *PropertyValBase[T]) Type() reflect.Type { return reflect.TypeFor[T]() }

// Get is a method that returns the value.
func (v *PropertyValBase[T]) Get(async bool) any {
	if v == nil {
		gl.Log("error", "Get: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return nil
	}
	vl := v.Pointer.Load()
	if async {
		if v.channelCtl != nil {
			gl.Log("debug", "Getting value from channel for:", v.Name, "ID:", v.ID.String())
			mCh := v.channelCtl.Channels["get"]
			if mCh != nil {
				ch := reflect.ValueOf(mCh)
				ch.Send(reflect.ValueOf(vl))
			}
		}
		return vl
	} else {
		gl.Log("debug", "Getting value for:", v.Name, "ID:", v.ID.String())
		return v.Pointer.Load()
	}
}

// Set is a method that sets the value.
func (v *PropertyValBase[T]) Set(t *T) bool {
	if v == nil {
		gl.Log("error", "Set: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return false
	}
	// Deixa o nil entrar na validação, pra ter a tratativa que a pessoa inseriu na validação que
	// pode ser customizada, etc...
	if v.Validation != nil {
		if v.CheckIfWillValidate() {
			gl.Log("warning", "Set: validation is disabled (", reflect.TypeFor[T]().String(), ")")
		} else {
			if ok := v.Validate(t); !ok.GetIsValid() {
				gl.Log("error", fmt.Sprintf("Set: validation error (%s): %v", reflect.TypeFor[T]().String(), v.Validation.GetResults()))
				return false
			}
		}
	}
	if t == nil {
		gl.Log("error", "Set: value is nil (", reflect.TypeFor[T]().String(), ")")
		return false
	}
	v.Store(t)
	if v.channelCtl != nil {
		gl.Log("debug", "Setting value for:", v.Name, "ID:", v.ID.String())
		v.channelCtl.Channels["set"].(chan T) <- *t
	}
	return true
}

// Clear is a method that clears the value.
func (v *PropertyValBase[T]) Clear() bool {
	if v == nil {
		gl.Log("error", "Clear: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return false
	}
	if v.channelCtl != nil {
		gl.Log("debug", "Clearing value for:", v.Name, "ID:", v.ID.String())
		v.channelCtl.Channels["clear"].(chan string) <- "clear"
	}
	return true
}

// IsNil is a method that checks if the value is nil.
func (v *PropertyValBase[T]) IsNil() bool {
	if v == nil {
		gl.Log("error", "Get: property does not exist (", reflect.TypeFor[T]().String(), ")")
		return true
	}
	return v.Load() == nil
}

// Serialize is a method that serializes the value.
func (v *PropertyValBase[T]) Serialize(filePath, format string) ([]byte, error) {
	if value := v.Value(); value == nil {
		return nil, fmt.Errorf("value is nil")
	} else {
		mapper := NewMapper(value, filePath)
		if data, err := mapper.Serialize(format); err != nil {
			gl.Log("error", "Failed to serialize data:", err.Error())
			return nil, err
		} else {
			return data, nil
		}
	}
}

// Deserialize is a method that deserializes the data into the value.
func (v *PropertyValBase[T]) Deserialize(data []byte, format, filePath string) error {
	if value := v.Value(); value == nil {
		return fmt.Errorf("value is nil")
	} else {
		mapper := NewMapper(value, filePath)
		if vl, vErr := mapper.Deserialize(data, format); vErr != nil {
			gl.Log("error", "Failed to deserialize data:", vErr.Error())
			return vErr
		} else {
			v.Store(vl)
			return nil
		}
	}
}
