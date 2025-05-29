package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	gdbf "github.com/rafa-mori/gdbase/factory"
	ci "github.com/rafa-mori/gobe/internal/interfaces"
	is "github.com/rafa-mori/gobe/internal/services"
	t "github.com/rafa-mori/gobe/internal/types"
)

type Route struct {
	*t.Mutexes

	// method
	// path
	// contentType
	properties map[string]string

	rateLimitLimit int
	requestWindow  time.Duration

	// secure
	// validateAndSanitize
	secureProperties map[string]bool

	// route objects
	dbService   is.DBService
	handler     gin.HandlerFunc
	middlewares map[string]any
}

// DBConfig implements interfaces.IRoute.
func (r *Route) DBConfig() gdbf.DBConfig {
	panic("unimplemented")
}

func newRoute(method, path, contentType string, handler gin.HandlerFunc, middlewares map[string]any, dbService gdbf.DBService) *Route {
	return &Route{
		Mutexes: t.NewMutexesType(),
		properties: map[string]string{
			"method":      method,
			"path":        path,
			"contentType": contentType,
		},
		rateLimitLimit: 0,
		requestWindow:  0,
		secureProperties: map[string]bool{
			"secure":                  false,
			"validateAndSanitize":     false,
			"validateAndSanitizeBody": false,
		},
		dbService:   dbService,
		handler:     handler,
		middlewares: middlewares,
	}
}

func NewRoute(method, path, contentType string, handler gin.HandlerFunc, middlewares map[string]any, dbConfig gdbf.DBService) ci.IRoute {
	return newRoute(method, path, contentType, handler, middlewares, dbConfig)
}

func (r *Route) Method() string                     { return r.properties["method"] }
func (r *Route) Path() string                       { return r.properties["path"] }
func (r *Route) ContentType() string                { return r.properties["contentType"] }
func (r *Route) RateLimitLimit() int                { return r.rateLimitLimit }
func (r *Route) RequestWindow() time.Duration       { return r.requestWindow }
func (r *Route) Secure() bool                       { return r.secureProperties["secure"] }
func (r *Route) ValidateAndSanitize() bool          { return r.secureProperties["validateAndSanitize"] }
func (r *Route) ValidateAndSanitizeBody() bool      { return r.secureProperties["validateAndSanitizeBody"] }
func (r *Route) SecureProperties() map[string]bool  { return r.secureProperties }
func (r *Route) Handler() gin.HandlerFunc           { return r.handler }
func (r *Route) Middlewares() map[string]any        { return r.middlewares }
func (r *Route) GetDatabaseService() gdbf.DBService { return r.dbService }

func (r *Route) SetMethod(method string)               { r.properties["method"] = method }
func (r *Route) SetPath(path string)                   { r.properties["path"] = path }
func (r *Route) SetContentType(contentType string)     { r.properties["contentType"] = contentType }
func (r *Route) SetRateLimitLimit(limit int)           { r.rateLimitLimit = limit }
func (r *Route) SetRequestWindow(window time.Duration) { r.requestWindow = window }
func (r *Route) SetSecure(secure bool)                 { r.secureProperties["secure"] = secure }
func (r *Route) SetValidateAndSanitize(validate bool) {
	r.secureProperties["validateAndSanitize"] = validate
}
func (r *Route) SetValidateAndSanitizeBody(validate bool) {
	r.secureProperties["validateAndSanitizeBody"] = validate
}
func (r *Route) SetHandler(handler gin.HandlerFunc)         { r.handler = handler }
func (r *Route) SetMiddlewares(middlewares map[string]any)  { r.middlewares = middlewares }
func (r *Route) SetDatabaseService(dbConfig gdbf.DBService) { r.dbService = dbConfig }
func (r *Route) SetProperties(properties map[string]string) { r.properties = properties }
func (r *Route) SetSecureProperties(secureProperties map[string]bool) {
	r.secureProperties = secureProperties
}
