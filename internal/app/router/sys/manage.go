package sys

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	manage "github.com/kubex-ecosystem/gobe/internal/app/controllers/sys/manage"
	"github.com/kubex-ecosystem/gobe/internal/app/middlewares"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	l "github.com/kubex-ecosystem/logz"
)

// NewServerRoutes cria rotas básicas de gestão da aplicação.
func NewServerRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		fmt.Println("Router is nil for ServerRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)
	middlewaresMap["logging"] = middlewares.Logger(l.GetLogger("GoBE-ServerRoutes"))
	middlewaresMap["rateLimit"] = middlewares.RateLimiter(5, 10)
	middlewaresMap["sanitize"] = middlewares.ValidateAndSanitize()

	secureProperties := map[string]bool{
		"secure":                  true,
		"validateAndSanitize":     true,
		"validateAndSanitizeBody": true,
	}
	openedProperties := map[string]bool{
		"secure":                  false,
		"validateAndSanitize":     false,
		"validateAndSanitizeBody": false,
	}

	controller := manage.NewServerController()

	routesMap["HealthPostRoute"] = proto.NewRoute(http.MethodPost, "/health", "application/json", controller.Health, nil, dbService, openedProperties, nil)
	routesMap["HealthGetRoute"] = proto.NewRoute(http.MethodGet, "/health", "application/json", controller.Health, nil, dbService, openedProperties, nil)
	routesMap["PingPostRoute"] = proto.NewRoute(http.MethodPost, "/ping", "application/json", controller.Ping, nil, dbService, openedProperties, nil)
	routesMap["PingGetRoute"] = proto.NewRoute(http.MethodGet, "/ping", "application/json", controller.Ping, nil, dbService, openedProperties, nil)

	routesMap["VersionGetRoute"] = proto.NewRoute(http.MethodGet, "/version", "application/json", controller.Version, middlewaresMap, dbService, secureProperties, nil)
	routesMap["ConfigGetRoute"] = proto.NewRoute(http.MethodGet, "/api/v1/config", "application/json", controller.Config, middlewaresMap, dbService, secureProperties, nil)
	routesMap["StartPostRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/start", "application/json", controller.Start, middlewaresMap, dbService, secureProperties, nil)
	routesMap["StopPostRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/stop", "application/json", controller.Stop, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
