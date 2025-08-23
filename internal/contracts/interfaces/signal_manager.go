package interfaces

type ISignalManager[T chan string] interface {
	ListenForSignals() (<-chan string, error)
	StopListening()
}
