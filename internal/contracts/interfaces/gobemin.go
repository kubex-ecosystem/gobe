package interfaces

import (
	"io"

	svc "github.com/kubex-ecosystem/gdbase/factory"
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
	Environment() svc.Environment
	InitializeResources() error
	InitializeServer() (IRouter, error)
	GetLogger() l.Logger
	StartGoBE()
	StopGoBE()
	GetChanCtl() chan string
	GetLogFilePath() string
	GetConfigFilePath() string
	SetDatabaseService(dbService svc.DBService)
	GetDatabaseService() svc.DBService
	LogsGoBE() (*io.OffsetWriter, error)
}
