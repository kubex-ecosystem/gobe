// Package router provides the routing functionality for the application.
package router

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	gdbf "github.com/kubex-ecosystem/gdbase/factory"
	"github.com/kubex-ecosystem/gdbase/types"
	mdw "github.com/kubex-ecosystem/gobe/internal/app/middlewares"
	ci "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	l "github.com/kubex-ecosystem/logz"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	_ "github.com/kubex-ecosystem/gobe/docs"

	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Router represents the main router structure for the application.
type Router struct {
	*gin.Engine
	*t.Mutexes
	InitArgs        gl.InitArgs
	Logger          l.Logger
	settings        map[string]string
	databaseService gdbf.DBService
	routes          map[string]map[string]ci.IRoute
	properties      map[string]any
	middlewares     map[string]gin.HandlerFunc
	engine          *gin.Engine
	debug           bool
}

// newRouter initializes a new Router instance with the provided configuration.
func newRouter(serverConfig *t.GoBEConfig, databaseService gdbf.DBService, logger l.Logger, debug bool) (*Router, error) {
	if logger == nil {
		logger = l.GetLogger("GoBE")
	}

	rtr := &Router{
		Logger:          logger,
		Mutexes:         t.NewMutexesType(),
		engine:          gin.New(),
		routes:          make(map[string]map[string]ci.IRoute),
		debug:           debug,
		databaseService: databaseService,
		properties:      make(map[string]any),
		middlewares:     make(map[string]gin.HandlerFunc),
		settings: map[string]string{
			"configPath":     serverConfig.FilePath,
			"bindingAddress": serverConfig.BindAddress,
			"port":           serverConfig.Port,
			"basePath":       serverConfig.BasePath,
		},
	}

	var autenticationMiddleware *mdw.AuthenticationMiddleware
	if databaseService != nil {
		tokenService, certService, err := mdw.NewTokenService(databaseService.GetConfig(), logger)
		if err != nil {
			gl.Log("error", fmt.Sprintf("❌ Failed to create token service: %v", err))
			return nil, err
		}
		autenticationMiddleware = &mdw.AuthenticationMiddleware{
			CertService:  certService,
			TokenService: tokenService,
		}
	}

	defaultMiddlewares := map[string]gin.HandlerFunc{
		"authentication":      autenticationMiddleware.ValidateJWT(mdw.NewAuthenticationMiddleware(autenticationMiddleware.TokenService, autenticationMiddleware.CertService, nil)),
		"validateAndSanitize": mdw.ValidateAndSanitize(),
		"rateLimite":          mdw.RateLimiter(rate.Limit(serverConfig.RateLimitLimit), serverConfig.RateLimitBurst),
		"logger":              mdw.Logger(logger),
		"backoff":             mdw.BackoffMiddleware(),
		"cache":               mdw.CacheMiddleware(),
		"meter":               mdw.MeterMiddleware(),
		"timeout":             mdw.TimeoutMiddleware(30 * time.Second),
	}

	// Set up the globals for gin (middlewares, logger, etc.)
	// They are set up once in the initialization of the router
	// and not in the initialization of the server
	rtr.engine.Use(gin.Recovery())
	rtr.engine.Use(gin.Logger())

	strMiddlewares := make([]string, 0)
	rtr.engine.Use(func(middlewares map[string]gin.HandlerFunc) []gin.HandlerFunc {
		middlewaresList := make([]gin.HandlerFunc, 0)
		for middlewareName, middleware := range middlewares {
			if middlewareName != "authentication" {
				rtr.RegisterMiddleware(middlewareName, middleware, true)
				middlewaresList = append(middlewaresList, middleware)
			} else {
				rtr.RegisterMiddleware(middlewareName, middleware, false)
			}
			strMiddlewares = append(strMiddlewares, middlewareName)
		}
		return middlewaresList
	}(defaultMiddlewares)...)

	rtr.middlewares = defaultMiddlewares

	fullBindAddress := net.JoinHostPort(rtr.settings["bindingAddress"], rtr.settings["port"])

	if err := SecureServerInit(rtr.engine, fullBindAddress); err != nil {
		gl.Log("error", "Failed to initialize secure server: "+err.Error())
		return nil, err
	}
	gl.Log("debug", fmt.Sprintf("Server security policies initialized at %s", fullBindAddress))

	for groupName, routeGroup := range GetDefaultRouteMap(rtr) {
		for routeName, route := range routeGroup {
			if route != nil {
				rtr.RegisterRoute(groupName, routeName, route, strMiddlewares)
			}
		}
	}

	// rtr.RegisterRoute(
	// 	"swagger",
	// 	"Swagger",
	// 	NewRoute(
	// 		http.MethodGet,
	// 		"/swagger/*any",
	// 		"Application/html",
	// 		ginSwagger.WrapHandler(swaggerfiles.Handler),
	// 		nil,
	// 		nil,
	// 	),
	// 	[]string{},
	// )

	if debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	rtr.GetEngine().StaticFile("/api/v1/terms-of-service", "./docs/terms-service_temp.pdf")

	rtr.GetEngine().StaticFS("/api/v1/discord/web", http.Dir("./web"))

	return rtr, nil
}

// NewRouter creates a new Router instance and returns it as an IRouter interface.
func NewRouter(serverConfig *t.GoBEConfig, databaseService gdbf.DBService, logger l.Logger, debug bool) (ci.IRouter, error) {
	return newRouter(serverConfig, databaseService, logger, debug)
}

// NewRequest is a placeholder function for creating a new request.
func NewRequest(dBConfig gdbf.DBConfig, s string, i1, i2 int) (any, any) {
	panic("unimplemented")
}

// GetDebug returns the debug mode status of the router.
func (rtr *Router) GetDebug() bool {
	return rtr.debug
}

// GetLogger returns the logger instance associated with the router.
func (rtr *Router) GetLogger() l.Logger {
	return rtr.Logger
}

// GetConfigPath returns the configuration file path.
func (rtr *Router) GetConfigPath() string {
	return rtr.settings["configPath"]
}

// GetBindingAddress returns the binding address of the server.
func (rtr *Router) GetBindingAddress() string {
	return rtr.settings["bindingAddress"]
}

// GetPort returns the port on which the server is running.
func (rtr *Router) GetPort() string {
	return rtr.settings["port"]
}

// GetBasePath returns the base path of the server.
func (rtr *Router) GetBasePath() string {
	return rtr.settings["basePath"]
}

// GetEngine returns the Gin engine instance.
func (rtr *Router) GetEngine() *gin.Engine {
	return rtr.engine
}

// GetDatabaseService returns the database service instance.
func (rtr *Router) GetDatabaseService() gdbf.DBService {
	return rtr.databaseService
}

// HandleFunc registers a GET route with the specified path and handler.
func (rtr *Router) HandleFunc(path string, handler gin.HandlerFunc) gin.IRoutes {
	return rtr.engine.Handle("GET", path, handler)
}

// DBConfig is a placeholder function for database configuration.
func (rtr *Router) DBConfig() gdbf.IDBConfig {
	return *types.NewDBConfig(
		nil,
	)
}

// Start starts the server with the configured settings.
func (rtr *Router) Start() error {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return nil
	}

	fullBindAddress := net.JoinHostPort(rtr.settings["bindingAddress"], rtr.settings["port"])

	if err := rtr.engine.Run(fullBindAddress); err != nil {
		gl.Log("error", "Failed to start server: "+err.Error())
		return err
	}
	return nil
}

// Stop stops the server gracefully.
func (rtr *Router) Stop() error {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return nil
	}

	rtr.ShutdownServerGracefully()

	return nil
}

// SetProperty sets a property in the router's properties map.
func (rtr *Router) SetProperty(key string, value any) {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}
	if rtr.properties == nil {
		rtr.properties = make(map[string]any)
	}
	rtr.properties[key] = value
}

// GetProperty retrieves a property value by its key.
func (rtr *Router) GetProperty(key string) any {
	if err := rtr.ValidateRouter(); err != nil {
		// Log the error using the logger
		gl.Log("error", err.Error())
		return nil
	}
	if rtr.properties == nil {
		// Initialize the properties map if it is nil
		return nil
	}

	if value, ok := rtr.properties[key]; ok {
		// Return the value if it exists
		return value
	}

	return nil

}

// GetProperties returns all properties of the router.
// It returns nil if the properties map is nil or if the router is invalid.
// It also logs an error if the router is invalid.
func (rtr *Router) GetProperties() map[string]any {
	if err := rtr.ValidateRouter(); err != nil {
		// Log the error using the logger
		gl.Log("error", err.Error())
		return nil
	}

	if rtr.properties == nil {
		// Initialize the properties map if it is nil
		return nil
	}

	return rtr.properties
}

// SetProperties sets multiple properties in the router's properties map.
func (rtr *Router) SetProperties(properties map[string]any) {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}
	if rtr.properties == nil {
		rtr.properties = make(map[string]any)
	}
	for k, v := range properties {
		rtr.properties[k] = v
	}
}

// GetRoutes returns all registered routes in the router.
func (rtr *Router) GetRoutes() map[string]map[string]ci.IRoute {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return nil
	}
	if rtr.routes == nil {
		rtr.routes = make(map[string]map[string]ci.IRoute)
	}
	return rtr.routes
}

// GetMiddlewares returns all registered middlewares in the router.
func (rtr *Router) GetMiddlewares() map[string]gin.HandlerFunc {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return nil
	}
	if rtr.middlewares == nil {
		rtr.middlewares = make(map[string]gin.HandlerFunc)
	}
	return rtr.middlewares
}

// RegisterMiddleware registers a middleware with the router.
func (rtr *Router) RegisterMiddleware(name string, middleware gin.HandlerFunc, global bool) {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}
	if rtr.middlewares == nil {
		rtr.middlewares = make(map[string]gin.HandlerFunc)
	}
	if global {
		rtr.engine.Use(middleware)
	} else {
		if _, ok := rtr.middlewares[name]; ok {
			gl.Log("warn", fmt.Sprintf("Middleware %s already registered", name))
		} else {
			rtr.middlewares[name] = middleware
			gl.Log("debug", fmt.Sprintf("Middleware %s registered", name))
		}
	}
}

// RegisterRoute registers a route with the router.
func (rtr *Router) RegisterRoute(groupName, routeName string, route ci.IRoute, middlewares []string) {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}
	if route == nil {
		gl.Log("error", "Route is nil")
		return
	}
	if groupName == "" {
		gl.Log("error", "Group name is empty")
		return
	}
	if routeName == "" {
		routeName = strings.ReplaceAll(route.Path(), "/", "_")
	}
	if _, ok := rtr.routes[groupName]; !ok {
		rtr.routes[groupName] = make(map[string]ci.IRoute)
	}
	if _, ok := rtr.routes[groupName][routeName]; ok {
		gl.Log("warn", fmt.Sprintf("Route %s already registered in group %s", routeName, groupName))
		return
	}

	var middlewaresStack []gin.HandlerFunc
	if len(route.Middlewares()) != 0 {
		for _, middlewareName := range middlewares {
			// If the middleware registered in the route is not in the list of middlewares
			// registered in the router, do not add the middleware
			if middleware, ok := rtr.middlewares[middlewareName]; ok {
				middlewaresStack = append(middlewaresStack, middleware)
			} else {
				gl.Log("warn", fmt.Sprintf("Middleware %s not found", middlewareName))
			}
		}
	}

	// Add specific middlewares for the route, if necessary
	if route.Secure() {
		if authMdw, ok := rtr.middlewares["authentication"]; ok {
			middlewaresStack = append(middlewaresStack, authMdw)
		} else {
			gl.Log("warn", "Global Authentication middleware not found")
		}
	}

	if route.ValidateAndSanitize() {
		if validateMdw, ok := rtr.middlewares["validateAndSanitize"]; ok {
			middlewaresStack = append(middlewaresStack, validateMdw)
		} else {
			gl.Log("warn", "Global Validate and sanitize middleware not found")
		}
	}

	// Register route with individual middlewares + final handler
	middlewaresStack = append(middlewaresStack, route.Handler())

	rtr.engine.Handle(route.Method(), route.Path(), UniqueMiddlewareStack(middlewaresStack)...)

	rtr.routes[groupName][routeName] = route

	gl.Log("debug", fmt.Sprintf("Route registered: [%s] %s", route.Method(), route.Path()))
}

// StartServer starts the server and logs its status.
func (rtr *Router) StartServer() {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}

	fullBindAddress := net.JoinHostPort(rtr.settings["bindingAddress"], rtr.settings["port"])
	gl.Log("info", fmt.Sprintf("Starting server at %s", fullBindAddress))

	if err := rtr.engine.Run(fullBindAddress); err != nil {
		gl.Log("error", fmt.Sprintf("Server failed to start: %s", err.Error()))
		return
	}

	gl.Log("info", "Server started successfully")
}

// ShutdownServerGracefully shuts down the server gracefully.
func (rtr *Router) ShutdownServerGracefully() {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}

	// Create an HTTP server with the Gin engine
	server := &http.Server{
		Addr:    net.JoinHostPort(rtr.settings["bindingAddress"], rtr.settings["port"]),
		Handler: rtr.engine,
	}

	// Create a context with timeout for safe shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gl.Log("debug", "Initiating graceful shutdown...")

	// Perform graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		gl.Log("error", fmt.Sprintf("Failed to gracefully shutdown server: %s", err.Error()))
		return
	}

	gl.Log("info", "Server shut down gracefully.")

	os.Exit(0)
}

// MonitorServer monitors the server's health and logs its status periodically.
func (rtr *Router) MonitorServer() {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return
	}

	// Server health log every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			connections := len(rtr.engine.Routes())
			gl.Log("debug", fmt.Sprintf("Server running at %s | Active connections: %d", rtr.GetBindingAddress(), connections))
		}
	}()
}

// ValidateRouter validates the router's configuration and components.
func (rtr *Router) ValidateRouter() error {
	if rtr == nil {
		return fmt.Errorf("router is nil")
	}
	if rtr.engine == nil {
		return fmt.Errorf("engine is nil")
	}
	if rtr.databaseService == nil {
		return fmt.Errorf("database service is nil")
	}
	return nil
}

// DummyHandler returns a dummy handler function for testing purposes.
func (rtr *Router) DummyHandler(_ chan interface{}) gin.HandlerFunc {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return nil
	}
	return func(c *gin.Context) {
		gl.Log("debug", "Dummy Placeholder")

		c.JSON(http.StatusOK, gin.H{"message": "Dummy Placeholder"})
	}
}

func (rtr *Router) GetInitArgs() gl.InitArgs {
	if err := rtr.ValidateRouter(); err != nil {
		gl.Log("error", err.Error())
		return gl.InitArgs{}
	}
	return rtr.InitArgs
}

func SecureServerInit(r *gin.Engine, fullBindAddress string) error {
	trustedProxies, trustedProxiesErr := GetTrustedProxies()
	if trustedProxiesErr != nil {
		return trustedProxiesErr
	}
	setTrustProxiesErr := r.SetTrustedProxies(trustedProxies)
	if setTrustProxiesErr != nil {
		return setTrustProxiesErr
	}

	r.Use(
		func(c *gin.Context) {
			if !ValidateExpectedHosts(fullBindAddress, c) {
				c.Abort()
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

				// Handle OPTIONS preflight requests
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(http.StatusOK)
					return
				}

				c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
				c.Header("Referrer-Policy", "strict-origin")
				c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
				c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")

				// c.Header("X-Frame-Options", "DENY")
				c.Header("X-XSS-Protection", "1; mode=block")
				c.Header("X-Content-Type-Options", "nosniff")

				c.Next()
			}
		},
	)

	return nil
}

func GetTrustedProxies() ([]string, error) {
	trustedProxies := viper.GetStringSlice("trustedProxies")
	if len(trustedProxies) == 0 {
		interfaces, err := net.Interfaces()
		if err != nil {
			return []string{}, err
		}

		for _, iface := range interfaces {
			if iface.Flags&net.FlagLoopback == 0 {
				addrs, addrsErr := iface.Addrs()
				if addrsErr != nil {
					return []string{}, fmt.Errorf("error getting addresses for interface %s: %s", iface.Name, addrsErr)
					//continue // Ignora erro
				}

				for _, addr := range addrs {
					ipNet, ok := addr.(*net.IPNet)
					if ok {
						trustedProxies = append(trustedProxies, ipNet.IP.String())
					}
				}
			}
		}
	}

	gl.Log("notice", "Trusted Proxies: %v", trustedProxies)

	return trustedProxies, nil
}

func ValidateExpectedHosts(fullBindAddress string, c *gin.Context) bool {
	// TODO: ENABLE THIS WHEN RUNNING WITH ANY PUBLISHED ADDRESS/PORT

	// if c.Request.Host == fullBindAddress ||
	// 	c.Request.URL.Host == fullBindAddress {
	// 	return true
	// }

	// bindPort := strings.Split(fullBindAddress, ":")[1]
	// trustedLocalList := []string{"localhost", "127.0.0.1", "localhost:" + bindPort, "127.0.0.1:" + bindPort}
	// for _, trustedLocal := range trustedLocalList {
	// 	if c.Request.Host == trustedLocal ||
	// 		c.Request.URL.Host == trustedLocal {
	// 		return true
	// 	}
	// }

	//c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Unauthorized host: " + c.Request.Host})
	// return false

	return true
}

func UniqueMiddlewareStack(middlewares []gin.HandlerFunc) []gin.HandlerFunc {
	uniqueMap := make(map[string]gin.HandlerFunc)
	uniqueList := []gin.HandlerFunc{}

	for _, middleware := range middlewares {
		funcPtr := fmt.Sprintf("%p", middleware) // Obtém o endereço da função como string

		if _, exists := uniqueMap[funcPtr]; !exists {
			uniqueMap[funcPtr] = middleware
			uniqueList = append(uniqueList, middleware)
		}
	}

	return uniqueList
}
