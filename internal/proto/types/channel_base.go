// Package types provides the base types for channels in the gobe package.
package types

import (
	"reflect"

	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ci "github.com/rafa-mori/gobe/internal/proto/interfaces"
	l "github.com/rafa-mori/logz"
)

// ChannelBase is a struct that holds the base properties for a channel.
type ChannelBase[T any] struct {
	*Mutexes              // Mutexes for this Channel instance
	Name     string       // The name of the channel.
	Channel  any          // The channel for the value. Main channel for this struct.
	Type     reflect.Type // The type of the channel.
	Buffers  int          // The number of buffers for the channel.
	Shared   interface{}  // Shared data for many purposes
}

// NewChannelBase creates a new ChannelBase instance with the provided name and type.
func NewChannelBase[T any](name string, buffers int, logger l.Logger) ci.IChannelBase[any] {
	if logger == nil {
		logger = l.GetLogger("GoLife")
	}
	mu := NewMutexesType()
	if buffers <= 0 {
		buffers = lgBuf
	}
	return &ChannelBase[any]{
		Mutexes: mu,
		Name:    name,
		Channel: make(chan T, buffers),
		Type:    reflect.TypeFor[T](),
		Buffers: buffers,
	}
}

// GetName returns the name of the channel.
func (cb *ChannelBase[T]) GetName() string {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Name
}

// GetChannel returns the channel and its type.
func (cb *ChannelBase[T]) GetChannel() (any, reflect.Type) {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Channel, reflect.TypeOf(cb.Channel)
}

// GetType returns the type of the channel.
func (cb *ChannelBase[T]) GetType() reflect.Type {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Type
}

// GetBuffers returns the number of buffers for the channel.
func (cb *ChannelBase[T]) GetBuffers() int {
	cb.MuRLock()
	defer cb.MuRUnlock()
	return cb.Buffers
}

// SetName sets the name of the channel and returns it.
func (cb *ChannelBase[T]) SetName(name string) string {
	cb.MuLock()
	defer cb.MuUnlock()
	cb.Name = name
	return cb.Name
}

// SetChannel sets the channel and returns it.
func (cb *ChannelBase[T]) SetChannel(typE reflect.Type, bufferSize int) any {
	cb.MuLock()
	defer cb.MuUnlock()
	cb.Channel = reflect.MakeChan(typE, bufferSize)
	return cb.Channel
}

// SetBuffers sets the number of buffers for the channel and returns it.
func (cb *ChannelBase[T]) SetBuffers(buffers int) int {
	cb.MuLock()
	defer cb.MuUnlock()
	cb.Buffers = buffers
	cb.Channel = make(chan T, buffers)
	return cb.Buffers
}

// Close closes the channel and returns an error if any.
func (cb *ChannelBase[T]) Close() error {
	cb.MuLock()
	defer cb.MuUnlock()
	if cb.Channel != nil {
		gl.LogObjLogger(cb, "info", "Closing channel for:", cb.Name)
		close(cb.Channel.(chan T))
	}
	return nil
}

// Clear clears the channel and returns an error if any.
func (cb *ChannelBase[T]) Clear() error {
	cb.MuLock()
	defer cb.MuUnlock()
	if cb.Channel != nil {
		gl.LogObjLogger(cb, "info", "Clearing channel for:", cb.Name)
		close(cb.Channel.(chan T))
		cb.Channel = make(chan T, cb.Buffers)
	}
	return nil
}
