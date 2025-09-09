package system

import "github.com/kubex-ecosystem/gobe/internal/services/mcp/hooks"

func UpdateSystemStateFromMetrics(bs *hooks.Bitstate[uint64, hooks.SystemDomain], cpuUsage, memFreeMB float64) {
	if cpuUsage > 85 {
		bs.Set(uint64(hooks.SysCPUHigh))
		enterThrottleMode()
	} else {
		bs.Clear(uint64(hooks.SysCPUHigh))
	}
	if memFreeMB < 500 {
		bs.Set(uint64(hooks.SysMemLow))
	} else {
		bs.Clear(uint64(hooks.SysMemLow))
	}
}

func enterThrottleMode() {
	// Ex: reduzir concorrência, ajustar timers, etc.
}
