package system

import (
	"reflect"
	"sync/atomic"
	"time"

	"github.com/rafa-mori/gobe/internal/types" // IMutexes
)

// phantom‑type Bitstate genérico
type Bitstate[T ~uint64, S any] struct {
	*types.Mutexes
	state uint64
}

func NewBitstate[T ~uint64, S any]() *Bitstate[T, S] {
	return &Bitstate[T, S]{Mutexes: types.NewMutexesType()}
}

func (b *Bitstate[T, S]) GetServiceType() reflect.Type {
	return reflect.TypeFor[S]()
}

// Hot-path: atomic
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

func (b *Bitstate[T, S]) Has(flag T) bool {
	return atomic.LoadUint64(&b.state)&uint64(flag) != 0
}

// Slow-path com timeout
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
