package interfaces

import (
	"io"

	gdbf "github.com/rafa-mori/gdbase/factory"
	l "github.com/rafa-mori/logz"
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
	Environment() IEnvironment
	InitializeResources() error
	InitializeServer() (IRouter, error)
	GetLogger() l.Logger
	StartGoBE()
	StopGoBE()
	GetChanCtl() chan string
	GetLogFilePath() string
	GetConfigFilePath() string
	SetDatabaseService(dbService gdbf.DBService)
	GetDatabaseService() gdbf.DBService
	LogsGoBE() (*io.OffsetWriter, error)
}
