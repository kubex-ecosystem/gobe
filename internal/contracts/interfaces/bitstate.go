package interfaces

type IBitstate[T ~uint64] interface {
	Set(flag T)
	Clear(flag T)
	Has(flag T) bool
	WaitFor(flag T)
}
