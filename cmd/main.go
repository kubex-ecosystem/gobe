package main

import (
	"github.com/kubex-ecosystem/gobe/internal/module"
	logz "github.com/kubex-ecosystem/logz"
)

// main initializes the logger and creates a new GoBE instance.
func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		logz.Log("fatal", err.Error())
	}
}
