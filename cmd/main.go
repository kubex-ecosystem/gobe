package main

import (
	l "github.com/faelmori/logz"
	gl "github.com/rafa-mori/gobe/logger"
)

var logger l.Logger

// main initializes the logger and creates a new GoBE instance.
func main() {
	if err := RegX().Command().Execute(); err != nil {
		gl.Log("fatal", err.Error())
	}
}
