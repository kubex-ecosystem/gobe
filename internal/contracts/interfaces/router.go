package interfaces

import (
	"time"

	"github.com/gin-gonic/gin"

	gdbf "github.com/kubex-ecosystem/gdbase/factory"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	l "github.com/kubex-ecosystem/logz"
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
	GetMiddlewares() map[string]gin.HandlerFunc
	RegisterMiddleware(name string, middleware gin.HandlerFunc, global bool)
	RegisterRoute(groupName, routeName string, route IRoute, middlewares []string)
	StartServer()
	ShutdownServerGracefully()
	MonitorServer()
	ValidateRouter() error
	GetInitArgs() gl.InitArgs
	DummyHandler(_ chan interface{}) gin.HandlerFunc
}

type IRoute interface {
	Method() string
	Path() string
	ContentType() string
	RateLimitLimit() int
	RequestWindow() time.Duration
	Secure() bool
	ValidateAndSanitize() bool
	SecureProperties() map[string]bool
	Handler() gin.HandlerFunc
	Middlewares() map[string]gin.HandlerFunc
	DBConfig() gdbf.DBConfig
}
