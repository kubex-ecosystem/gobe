// Package kbx provides utilities for working with initialization arguments.
package kbx

import (
	"os"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"
)

type InitArgs struct {
	ConfigFile     string
	ConfigType     string
	EnvFile        string
	LogFile        string
	Name           string
	Debug          bool
	ReleaseMode    bool
	IsConfidential bool
	Port           string
	Bind           string
	Address        string
	PubCertKeyPath string
	PubKeyPath     string
	Pwd            string
}

type ILogger = l.Logger

type Logger = gl.GLog[ILogger]

func Log(level string, payload ...any) {
	gl.Log(level, payload...)
}

func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
