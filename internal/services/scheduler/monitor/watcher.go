package monitor

import (
	"runtime"
	"time"

	logz "github.com/kubex-ecosystem/logz"
)

func watchGoroutines() {
	go func() {
		for range time.Tick(5 * time.Second) {
			if n := runtime.NumGoroutine(); n > 100 {
				logz.Log("warning", "Warning: %d goroutines running—possible leak?", n)
			}
		}
	}()
}
