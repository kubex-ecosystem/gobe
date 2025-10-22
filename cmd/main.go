package main

import (
	"github.com/kubex-ecosystem/gobe/internal/module"
	gl "github.com/kubex-ecosystem/logz/logger"
)

// main initializes the logger and creates a new GoBE instance.
func main() {
	if err := module.RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
