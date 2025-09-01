package hooks

type MCP struct {
	ConfigState *Bitstate[uint64, ConfigDomain]
	SystemState *Bitstate[uint64, SystemDomain]
}

func NewMCP[T uint64, S any]() *MCP {
	return &MCP{
		ConfigState: NewBitstate[uint64, ConfigDomain](nil),
		SystemState: NewBitstate[uint64, SystemDomain](nil),
	}
}
