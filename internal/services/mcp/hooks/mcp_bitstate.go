package hooks

import (
	"reflect"

	t "github.com/rafa-mori/gobe/internal/proto/types"
)

type Bitstate[T ~uint64, S any] struct {
	*t.Mutexes
	state uint64
}

func NewBitstate[T ~uint64, S any](service *S) *Bitstate[T, S] {
	b := &Bitstate[T, S]{
		Mutexes: t.NewMutexesType(),
	}
	return b
}

func (b *Bitstate[T, S]) Set(flag T) {
	b.MuLock()
	b.state |= uint64(flag)
	b.MuUnlock()
	b.MuSignalCond()
}

func (b *Bitstate[T, S]) Clear(flag T) {
	b.MuLock()
	b.state &^= uint64(flag)
	b.MuUnlock()
	b.MuBroadcastCond()
}

func (b *Bitstate[T, S]) Has(flag T) bool {
	b.MuLock()
	defer b.MuUnlock()
	return b.state&uint64(flag) != 0
}

func (b *Bitstate[T, S]) WaitFor(flag T) {
	b.MuLock()
	for b.state&uint64(flag) == 0 {
		b.MuWaitCond()
	}
	b.MuUnlock()
}

func (b *Bitstate[T, S]) GetServiceType() reflect.Type {
	return reflect.TypeFor[S]()
}
