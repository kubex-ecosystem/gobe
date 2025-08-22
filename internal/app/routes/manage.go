package routes

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/app/middlewares"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
	l "github.com/rafa-mori/logz"
)

type ServerRoutes struct {
	ar.IRouter
}

func NewServerRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		fmt.Println("Router is nil for ServerRoute")
		return nil
	}
	rtl := *rtr

	ra := ServerRoutes{IRouter: rtl}
	dbService := rtl.GetDatabaseService()
	// if dbService == nil {
	// 	fmt.Println("Database service is nil for ServerRoute")
	// 	return nil
	// }

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)
	middlewaresMap["logging"] = middlewares.Logger(l.GetLogger("GoBE-ServerRoutes"))
	middlewaresMap["rateLimit"] = middlewares.RateLimiter(5, 10) // 5 requests per 10 seconds
	middlewaresMap["sanitize"] = middlewares.ValidateAndSanitize()

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = true
	secureProperties["validateAndSanitizeBody"] = true

	openedProperties := make(map[string]bool)
	openedProperties["secure"] = false
	openedProperties["validateAndSanitize"] = false
	openedProperties["validateAndSanitizeBody"] = false

	routesMap["HealthPostRoute"] = NewRoute(http.MethodPost, "/health", "application/json", ra.HealthRouteHandler(nil), nil, dbService, openedProperties)
	routesMap["HealthGetRoute"] = NewRoute(http.MethodGet, "/health", "application/json", ra.HealthRouteHandler(nil), nil, dbService, openedProperties)
	routesMap["PingPostRoute"] = NewRoute(http.MethodPost, "/ping", "application/json", ra.PingRouteHandler(nil), nil, dbService, openedProperties)
	routesMap["PingGetRoute"] = NewRoute(http.MethodGet, "/ping", "application/json", ra.PingRouteHandler(nil), nil, dbService, openedProperties)

	routesMap["VersionGetRoute"] = NewRoute(http.MethodGet, "/version", "application/json", ra.VersionRouteHandler(nil), ra.GetMiddlewares(), dbService, secureProperties)
	routesMap["ConfigGetRoute"] = NewRoute(http.MethodGet, "/api/v1/config", "application/json", ra.ConfigRouteHandler(nil), ra.GetMiddlewares(), dbService, secureProperties)

	routesMap["StartPostRoute"] = NewRoute(http.MethodPost, "/api/v1/start", "application/json", ra.StartRouteHandler(nil), ra.GetMiddlewares(), dbService, secureProperties)
	routesMap["StopPostRoute"] = NewRoute(http.MethodPost, "/api/v1/stop", "application/json", ra.StopRouteHandler(nil), ra.GetMiddlewares(), dbService, secureProperties)

	return routesMap
}

func (r *ServerRoutes) PingRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) }
}
func (r *ServerRoutes) PingBrokerRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if r == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	}
}
func (r *ServerRoutes) HealthRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	}
}
func (r *ServerRoutes) VersionRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "v1.0.0"})
	}
}
func (r *ServerRoutes) ConfigRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"config": "config"})
	}
}
func (r *ServerRoutes) StartRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse all we need from the request to create a child server
		//g := c.NewServer()

		/*if err := c.StartServer(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to start gobe"})
			return
		}*/
		c.JSON(http.StatusOK, gin.H{"message": "gobe started successfully"})
	}
}
func (r *ServerRoutes) StopRouteHandler(_ chan interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		/*if err := c.StopServer(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to stop gobe"})
			return
		}*/
		c.JSON(http.StatusOK, gin.H{"message": "gobe stopped successfully"})
	}
}
