package monitor

import (
	"log"
	"runtime"
	"time"
)

func watchGoroutines() {
	go func() {
		for range time.Tick(5 * time.Second) {
			if n := runtime.NumGoroutine(); n > 100 {
				log.Printf("Warning: %d goroutines running—possible leak?", n)
			}
		}
	}()
}
