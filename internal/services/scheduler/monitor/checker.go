// Package monitor provides functionality to monitor and check system resources.
package monitor

import (
	"errors"
	"fmt"
	"runtime"
	"syscall"
)

func PreLaunchChecks() error {
	// Check file descriptors
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}
	if rLimit.Cur < 10000 {
		return fmt.Errorf("need more file descriptors: got %d", rLimit.Cur)
	}

	// Check CPU
	if runtime.NumCPU() < 2 {
		return errors.New("need at least 2 cores")
	}
	return nil
}
