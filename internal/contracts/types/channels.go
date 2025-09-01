package types

import (
	"fmt"

	ci "github.com/rafa-mori/gobe/internal/contracts/interfaces"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	tu "github.com/rafa-mori/gobe/internal/utils"
	l "github.com/rafa-mori/logz"

	"reflect"

	"github.com/google/uuid"
)

var (
	smBuf, mdBuf, lgBuf = tu.GetDefaultBufferSizes()
)

// ChannelCtl is a struct that holds the properties for a channel control.
type ChannelCtl[T any] struct {
	// IChannelCtl is the interface for this Channel instance.
	//ci.IChannelCtl[T] // Channel interface for this Channel instance

	// Logger is the Logger instance for this Channel instance.
	Logger l.Logger // Logger for this Channel instance

	// IMutexes is the interface for the mutexes in this Channel instance.
	*Mutexes // Mutexes for this Channel instance

	// property is the property for the channel.
	property ci.IProperty[T] // Lazy load, only used when needed or created by NewChannelCtlWithProperty constructor

	// Shared is a shared data used for many purposes like sync.Cond, Telemetry, Monitor, etc.
	Shared interface{} // Shared data for many purposes

	withMetrics bool // If true, will create the telemetry and monitor channels

	// ch is a channel for the value.
	ch chan T // The channel for the value. Main channel for this struct.

	// Reference is the reference ID and name.
	*Reference `json:"reference" yaml:"reference" xml:"reference" gorm:"reference"`

	// buffers is the number of buffers for the channel.
	Buffers int `json:"buffers" yaml:"buffers" xml:"buffers" gorm:"buffers"`

	Channels map[string]any `json:"channels,omitempty" yaml:"channels,omitempty" xml:"channels,omitempty" gorm:"channels,omitempty"`
}

// NewChannelCtl creates a new ChannelCtl instance with the provided name.
func NewChannelCtl[T any](name string, logger l.Logger) ci.IChannelCtl[T] {
	if logger == nil {
		logger = l.GetLogger("GoLife")
	}
	ref := NewReference(name)
	mu := NewMutexesType()

	// Create a new ChannelCtl instance
	channelCtl := &ChannelCtl[T]{
		Logger:    logger,
		Reference: ref.GetReference(),
		Mutexes:   mu,
		ch:        make(chan T, lgBuf),
		Channels:  make(map[string]any),
	}
	channelCtl.Channels = getDefaultChannelsMap(false, logger)
	return channelCtl
}

// NewChannelCtlWithProperty creates a new ChannelCtl instance with the provided name and type.
func NewChannelCtlWithProperty[T any, P ci.IProperty[T]](name string, buffers *int, property P, withMetrics bool, logger l.Logger) ci.IChannelCtl[T] {
	if logger == nil {
		logger = l.GetLogger("GoLife")
	}
	ref := NewReference(name)
	mu := NewMutexesType()
	buf := 3
	if buffers != nil {
		buf = *buffers
	}
	channelCtl := &ChannelCtl[T]{
		Logger:    logger,
		Reference: ref.GetReference(),
		Mutexes:   mu,
		ch:        make(chan T, buf),
		Channels:  make(map[string]any),
		property:  property,
	}
	channelCtl.Channels = getDefaultChannelsMap(withMetrics, logger)

	return channelCtl
}

// GetID returns the ID of the channel control.
func (cCtl *ChannelCtl[T]) GetID() uuid.UUID {
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.ID
}

// GetName returns the name of the channel control.
func (cCtl *ChannelCtl[T]) GetName() string {
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.Name
}

// SetName sets the name of the channel control and returns it.
func (cCtl *ChannelCtl[T]) SetName(name string) string {
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.Name = name
	return cCtl.Name
}

// GetProperty returns the property of the channel control.
func (cCtl *ChannelCtl[T]) GetProperty() ci.IProperty[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.property
}

// GetSubChannels returns the sub-channels of the channel control.
func (cCtl *ChannelCtl[T]) GetSubChannels() map[string]interface{} {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.Channels
}

// SetSubChannels sets the sub-channels of the channel control and returns the updated map.
func (cCtl *ChannelCtl[T]) SetSubChannels(channels map[string]interface{}) map[string]interface{} {
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	for k, v := range channels {
		if _, ok := cCtl.Channels[k]; ok {
			cCtl.Channels[k] = v
		} else {
			cCtl.Channels[k] = v
		}
	}
	return cCtl.Channels
}

// GetSubChannelByName returns the sub-channel by name and its type.
func (cCtl *ChannelCtl[T]) GetSubChannelByName(name string) (any, reflect.Type, bool) {
	if cCtl.Channels == nil {
		gl.LogObjLogger(cCtl, "info", "Creating channels map for:", cCtl.Name, "ID:", cCtl.ID.String())
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if rawChannel, ok := cCtl.Channels[name]; ok {
		if channel, ok := rawChannel.(ci.IChannelBase[T]); ok {
			return channel, channel.GetType(), true
		} else {
			gl.LogObjLogger(cCtl, "error", fmt.Sprintf("Channel %s is not a valid channel type. Expected: %s, receive %s", name, reflect.TypeFor[ci.IChannelBase[T]]().String(), reflect.TypeOf(rawChannel)))
			return nil, nil, false
		}
	}
	gl.LogObjLogger(cCtl, "error", "Channel not found:", name, "ID:", cCtl.ID.String())
	return nil, nil, false
}

// SetSubChannelByName sets the sub-channel by name and returns the channel.
func (cCtl *ChannelCtl[T]) SetSubChannelByName(name string, channel any) (any, error) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if _, ok := cCtl.Channels[name]; ok {
		cCtl.Channels[name] = channel
	} else {
		cCtl.Channels[name] = channel
	}
	return channel, nil
}

// GetSubChannelTypeByName returns the type of the sub-channel by name.
func (cCtl *ChannelCtl[T]) GetSubChannelTypeByName(name string) (reflect.Type, bool) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if channel, ok := cCtl.Channels[name]; ok {
		return channel.(ci.IChannelBase[any]).GetType(), true
	}
	return nil, false
}

// SetSubChannelTypeByName sets the type of the sub-channel by name and returns the type.

func (cCtl *ChannelCtl[T]) GetSubChannelBuffersByName(name string) (int, bool) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	if channel, ok := cCtl.Channels[name]; ok {
		return channel.(ci.IChannelBase[any]).GetBuffers(), true
	}
	return 0, false
}

// SetSubChannelBuffersByName sets the number of buffers for the sub-channel by name and returns the number of buffers.
func (cCtl *ChannelCtl[T]) SetSubChannelBuffersByName(name string, buffers int) (int, error) {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if channel, ok := cCtl.Channels[name]; ok {
		channel.(ci.IChannelBase[any]).SetBuffers(buffers)
		return buffers, nil
	}
	return 0, nil
}

// GetMainChannel returns the main channel and its type.
func (cCtl *ChannelCtl[T]) GetMainChannel() any {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.ch
}

// SetMainChannel sets the main channel and returns it.
func (cCtl *ChannelCtl[T]) SetMainChannel(channel chan T) chan T {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}

	if channel != nil {
		cCtl.MuLock()
		defer cCtl.MuUnlock()
		cCtl.ch = channel
	} else {
		gl.LogObjLogger(cCtl, "warn", "SetMainChannel: provided channel is nil, keeping existing channel")
		cCtl.MuRLock()
		defer cCtl.MuRUnlock()
		cCtl.ch = nil
	}
	return cCtl.ch
}

// GetMainChannelType returns the type of the main channel.
func (cCtl *ChannelCtl[T]) GetMainChannelType() reflect.Type {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return reflect.TypeOf(cCtl.ch)
}

// GetHasMetrics returns true if the channel control has metrics enabled.
func (cCtl *ChannelCtl[T]) GetHasMetrics() bool {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.withMetrics
}

// SetHasMetrics sets the hasMetrics flag and returns it.
func (cCtl *ChannelCtl[T]) SetHasMetrics(hasMetrics bool) bool {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.withMetrics = hasMetrics
	return cCtl.withMetrics
}

// GetBufferSize returns the buffer size of the channel control.
func (cCtl *ChannelCtl[T]) GetBufferSize() int {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuRLock()
	defer cCtl.MuRUnlock()
	return cCtl.Buffers
}

// SetBufferSize sets the buffer size of the channel control and returns it.
func (cCtl *ChannelCtl[T]) SetBufferSize(size int) int {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.Buffers = size
	return cCtl.Buffers
}

// Close closes the channel control and returns an error if any.
func (cCtl *ChannelCtl[T]) Close() error {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	if cCtl.Channels != nil {
		for _, channel := range cCtl.Channels {
			if ch, ok := channel.(ci.IChannelBase[any]); ok {
				_ = ch.Close()
			}
		}
	}
	return nil
}

// WithProperty sets the property for the channel control and returns it.
func (cCtl *ChannelCtl[T]) WithProperty(property ci.IProperty[T]) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.property = property
	return cCtl
}

// WithChannel sets the channel for the channel control and returns it.
func (cCtl *ChannelCtl[T]) WithChannel(channel chan T) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.ch = channel
	return cCtl
}

// WithBufferSize sets the buffer size for the channel control and returns it.
func (cCtl *ChannelCtl[T]) WithBufferSize(size int) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.Buffers = size
	return cCtl
}

// WithMetrics sets the metrics flag for the channel control and returns it.
func (cCtl *ChannelCtl[T]) WithMetrics(metrics bool) ci.IChannelCtl[T] {
	if cCtl.Channels == nil {
		cCtl.Channels = initChannelsMap(cCtl)
	}
	cCtl.MuLock()
	defer cCtl.MuUnlock()
	cCtl.withMetrics = metrics
	return cCtl
}

// initChannelsMap initializes the channels map for the ChannelCtl instance.
func initChannelsMap[T any](v *ChannelCtl[T]) map[string]interface{} {
	if v.Channels == nil {
		v.MuLock()
		defer v.MuUnlock()
		gl.LogObjLogger(v, "info", "Creating channels map for:", v.Name, "ID:", v.ID.String())
		v.Channels = make(map[string]interface{})
		// done is a channel for the done signal.
		v.Channels["done"] = NewChannelBase[bool]("done", smBuf, v.Logger)
		// ctl is a channel for the internal control channel.
		v.Channels["ctl"] = NewChannelBase[string]("ctl", mdBuf, v.Logger)
		// condition is a channel for the condition signal.
		v.Channels["condition"] = NewChannelBase[string]("cond", smBuf, v.Logger)

		if v.withMetrics {
			v.Channels["telemetry"] = NewChannelBase[string]("telemetry", mdBuf, v.Logger)
			v.Channels["monitor"] = NewChannelBase[string]("monitor", mdBuf, v.Logger)
		}
	}
	return v.Channels
}

// getDefaultChannelsMap returns a map with default channels for the ChannelCtl instance.
func getDefaultChannelsMap(withMetrics bool, logger l.Logger) map[string]any {
	mp := map[string]any{
		// done is a channel for the done signal.
		"done": NewChannelBase[bool]("done", smBuf, logger),
		// ctl is a channel for the internal control channel.
		"ctl": NewChannelBase[string]("ctl", mdBuf, logger),
		// condition is a channel for the condition signal.
		"condition": NewChannelBase[string]("cond", smBuf, logger),
	}

	if withMetrics {
		// metrics is a channel for the telemetry signal.
		mp["metrics"] = NewChannelBase[string]("metrics", mdBuf, logger)
		// monitor is a channel for monitoring the channel.
		mp["monitor"] = NewChannelBase[string]("monitor", mdBuf, logger)
	}

	return mp
}
