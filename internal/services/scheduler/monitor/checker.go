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
		return fmt.Errorf("Need more file descriptors: got %d", rLimit.Cur)
	}

	// Check CPU
	if runtime.NumCPU() < 2 {
		return errors.New("Need at least 2 cores")
	}
	return nil
}
