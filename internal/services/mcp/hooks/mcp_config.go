package hooks

type SystemDomain struct{}
type SystemFlag uint64

const (
	SysNetReady SystemFlag = 1 << iota
	SysAIBusy
	SysStorageSyncing
	SysErrorDetected
	SysCPUHigh
	SysMemLow
)
