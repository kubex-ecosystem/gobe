// Package kbx provides utilities for working with initialization arguments.
package kbx

import (
	"reflect"

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

func SetDebugMode(debug bool) {
	gl.SetDebug(debug)
}

func SetLogLevel(level string) {
	gl.Logger.SetLogLevel(level)
}

func SetLogTrace(enable bool) {
	gl.Logger.SetShowTrace(enable)
}

func GetLogger(name string) Logger {
	lgr := l.GetLogger(name)
	return gl.GetLogger(&lgr)
}

func IsObjValid(obj any) bool {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	if !v.IsValid() {
		return false
	}
	return true
}
