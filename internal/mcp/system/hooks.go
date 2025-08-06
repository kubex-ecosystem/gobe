package system

func UpdateSystemStateFromMetrics(bs *Bitstate[uint64, SystemDomain], cpuUsage, memFreeMB float64) {
	if cpuUsage > 85 {
		bs.Set(SysCPUHigh)
		enterThrottleMode()
	} else {
		bs.Clear(SysCPUHigh)
	}
	if memFreeMB < 500 {
		bs.Set(SysMemLow)
	} else {
		bs.Clear(SysMemLow)
	}
}

func enterThrottleMode() {
	// Ex: reduzir concorrência, ajustar timers, etc.
}
