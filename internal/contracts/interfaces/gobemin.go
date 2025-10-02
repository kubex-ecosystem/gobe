package interfaces

import (
	"io"
	"reflect"

	is "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	l "github.com/kubex-ecosystem/logz"
)

type ContactForm struct {
	Token                string `json:"token"`
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Message              string `json:"message"`
	IMapper[ContactForm] `json:"-" yaml:"-" xml:"-" toml:"-" gorm:"-"`
}

type IGoBE interface {
	GetReference() IReference
	Environment() is.Environment
	InitializeResources() error
	InitializeServer() (IRouter, error)
	GetLogger() l.Logger
	StartGoBE()
	StopGoBE()
	GetChanCtl() chan string
	GetLogFilePath() string
	GetConfigFilePath() string
	GetProperty(name string) (any, reflect.Type, error)
	SetDatabaseService(dbService is.DBService)
	GetDatabaseService() is.DBService
	LogsGoBE() (*io.OffsetWriter, error)
}
