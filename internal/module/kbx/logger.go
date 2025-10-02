package kbx

import (
	"reflect"

	gl "github.com/kubex-ecosystem/gobe/internal/module/logger"
	l "github.com/kubex-ecosystem/logz"
)

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
	return gl.Logger
}

func SetLogger(logger l.Logger) {
	// gl.SetLogger(logger)
	// TODO: Implement this function properly
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
