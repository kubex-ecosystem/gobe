package interfaces

import (
	l "github.com/faelmori/logz"
	gdbf "github.com/rafa-mori/gdbase/factory"
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
}
