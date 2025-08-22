// Package system provides domain and flag definitions for the system component of the application.
package system

type ConfigDomain struct{}
type ConfigFlag uint64

const (
	ConfEnableDiscord ConfigFlag = 1 << iota
	ConfEnableWebhooks
	ConfEnableLLM
	ConfDebugMode
)

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
