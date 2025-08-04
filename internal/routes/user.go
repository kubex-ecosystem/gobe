package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rafa-mori/gobe/internal/controllers/users"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"

	"net/http"
)

type AuthRoutes struct {
	ar.IRouter
}

func NewAuthRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("Router is nil for AuthRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for AuthRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	userController := users.NewUserController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := rtl.GetMiddlewares()

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["LoginRoute"] = NewRoute(http.MethodPost, "/api/v1/sign-in", "application/json", userController.AuthenticateUser, nil, dbService, nil)
	routesMap["LogoutRoute"] = NewRoute(http.MethodPost, "/api/v1/sign-out", "application/json", userController.Logout, middlewaresMap, dbService, secureProperties)
	routesMap["RefreshRoute"] = NewRoute(http.MethodPost, "/api/v1/check", "application/json", userController.RefreshToken, middlewaresMap, dbService, secureProperties)
	routesMap["RegisterRoute"] = NewRoute(http.MethodPost, "/api/v1/sign-up", "application/json", userController.CreateUser, nil, dbService, nil)

	return routesMap
}

func NewUserRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for UserRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for UserRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	userController := users.NewUserController(dbGorm)

	routesMap := make(map[string]ar.IRoute)

	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllUsers"] = NewRoute(http.MethodGet, "/users", "application/json", userController.GetAllUsers, middlewaresMap, dbService, secureProperties)
	routesMap["GetUserByID"] = NewRoute(http.MethodGet, "/users/:id", "application/json", userController.GetUserByID, middlewaresMap, dbService, secureProperties)
	routesMap["UpdateUser"] = NewRoute(http.MethodPut, "/users/:id", "application/json", userController.UpdateUser, middlewaresMap, dbService, secureProperties)
	routesMap["DeleteUser"] = NewRoute(http.MethodDelete, "/users/:id", "application/json", userController.DeleteUser, middlewaresMap, dbService, secureProperties)

	return routesMap
}
