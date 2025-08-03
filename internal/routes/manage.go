package routes

import (
	"fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
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

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["HealthRoute"] = NewRoute(http.MethodPost, "/health", "application/json", ra.PingRouteHandler(nil), middlewaresMap, dbService, nil)
	routesMap["PingPostRoute"] = NewRoute(http.MethodPost, "/ping", "application/json", ra.PingRouteHandler(nil), middlewaresMap, dbService, nil)

	routesMap["PingPostRoute"] = NewRoute(http.MethodGet, "/version", "application/json", ra.VersionRouteHandler(nil), middlewaresMap, dbService, nil)
	routesMap["PingPostRoute"] = NewRoute(http.MethodGet, "/config", "application/json", ra.ConfigRouteHandler(nil), middlewaresMap, dbService, secureProperties)

	routesMap["PingPostRoute"] = NewRoute(http.MethodPost, "/start", "application/json", ra.StartRouteHandler(nil), middlewaresMap, dbService, secureProperties)
	routesMap["PingPostRoute"] = NewRoute(http.MethodPost, "/stop", "application/json", ra.StopRouteHandler(nil), middlewaresMap, dbService, secureProperties)

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
