package interfaces

import (
	"time"

	"github.com/gin-gonic/gin"
	gdbf "github.com/rafa-mori/gdbase/factory"
	l "github.com/rafa-mori/logz"
)

type IRouter interface {
	GetDebug() bool
	GetLogger() l.Logger
	GetConfigPath() string
	GetBindingAddress() string
	GetPort() string
	GetBasePath() string
	GetEngine() *gin.Engine
	GetDatabaseService() gdbf.DBService
	HandleFunc(path string, handler gin.HandlerFunc) gin.IRoutes
	DBConfig() gdbf.IDBConfig
	Start() error
	Stop() error
	SetProperty(key string, value any)
	GetProperty(key string) any
	GetProperties() map[string]any
	SetProperties(properties map[string]any)
	GetRoutes() map[string]map[string]IRoute
	RegisterMiddleware(name string, middleware gin.HandlerFunc, global bool)
	RegisterRoute(groupName, routeName string, route IRoute, middlewares []string)
	StartServer()
	ShutdownServerGracefully()
	MonitorServer()
	ValidateRouter() error
	DummyHandler(_ chan interface{}) gin.HandlerFunc
}

type IRoute interface {
	// gin.IRoutes

	Method() string
	Path() string
	ContentType() string

	RateLimitLimit() int
	RequestWindow() time.Duration

	Secure() bool
	ValidateAndSanitize() bool
	SecureProperties() map[string]bool

	Handler() gin.HandlerFunc
	Middlewares() map[string]any

	DBConfig() gdbf.DBConfig
}
