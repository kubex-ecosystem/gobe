package system

import (
	"reflect"
	"sync/atomic"
	"time"

	"github.com/rafa-mori/gobe/internal/proto/types" // IMutexes
)

// Bitstate is a thread-safe bit state manager.
type Bitstate[T ~uint64, S any] struct {
	*types.Mutexes
	state uint64
}

// NewBitstate creates a new Bitstate instance with the specified type parameters.
func NewBitstate[T ~uint64, S any]() *Bitstate[T, S] {
	return &Bitstate[T, S]{Mutexes: types.NewMutexesType()}
}

// GetServiceType returns the reflect.Type of the service associated with the Bitstate.
func (b *Bitstate[T, S]) GetServiceType() reflect.Type {
	return reflect.TypeFor[S]()
}

// Set sets a specific flag in the current state.
func (b *Bitstate[T, S]) Set(flag T) {
	for {
		old := atomic.LoadUint64(&b.state)
		new := old | uint64(flag)
		if atomic.CompareAndSwapUint64(&b.state, old, new) {
			b.MuBroadcastCond()
			return
		}
	}
}

// Clear removes a flag from the current state.
func (b *Bitstate[T, S]) Clear(flag T) {
	for {
		old := atomic.LoadUint64(&b.state)
		new := old &^ uint64(flag)
		if atomic.CompareAndSwapUint64(&b.state, old, new) {
			b.MuBroadcastCond()
			return
		}
	}
}

// Has checks if a specific flag is set in the current state.
func (b *Bitstate[T, S]) Has(flag T) bool {
	return atomic.LoadUint64(&b.state)&uint64(flag) != 0
}

// WaitFor blocks until a specific flag is set or a timeout occurs.
func (b *Bitstate[T, S]) WaitFor(flag T, timeout time.Duration) bool {
	b.MuLock()
	defer b.MuUnlock()

	deadline := time.Now().Add(timeout)
	for !b.Has(flag) {
		if remaining := time.Until(deadline); remaining <= 0 {
			return false
		} else if !b.MuWaitCondWithTimeout(remaining) {
			return false
		}
	}
	return true
}
