package mcp

func NewConfigBitstate[T uint64, S any]() *Bitstate[uint64, S] {
	return NewBitstate[uint64, S](nil)
}

type ConfigDomain struct{}
type ConfigFlag uint64

const (
	ConfEnableDiscord ConfigFlag = 1 << iota
	ConfEnableWebhooks
	ConfEnableLLM
	ConfDebugMode
)
