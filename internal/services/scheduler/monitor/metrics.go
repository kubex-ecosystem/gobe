package monitor

import "runtime"

type Metrics struct {
	Goroutines int
	HeapMB     float64
}

func GetMetrics() Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Metrics{
		Goroutines: runtime.NumGoroutine(),
		HeapMB:     float64(m.HeapAlloc) / 1024 / 1024,
	}
}
