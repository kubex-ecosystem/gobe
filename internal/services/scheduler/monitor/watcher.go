package monitor

import (
	"runtime"
	"time"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
)

func watchGoroutines() {
	go func() {
		for range time.Tick(5 * time.Second) {
			if n := runtime.NumGoroutine(); n > 100 {
				gl.Log("warning", "Warning: %d goroutines running—possible leak?", n)
			}
		}
	}()
}
