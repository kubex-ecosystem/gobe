// Package user provides the user-related routes for the application.
package user

import (
	"github.com/gin-gonic/gin"
	"github.com/kubex-ecosystem/gobe/internal/app/controllers/sys/federation/users"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

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

	routesMap["LoginRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/sign-in", "application/json", userController.AuthenticateUser, nil, dbService, nil, nil)
	routesMap["LogoutRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/sign-out", "application/json", userController.Logout, middlewaresMap, dbService, secureProperties, nil)
	routesMap["RefreshRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/check", "application/json", userController.RefreshToken, middlewaresMap, dbService, secureProperties, nil)
	routesMap["RegisterRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/sign-up", "application/json", userController.CreateUser, nil, dbService, nil, nil)

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

	routesMap["GetAllUsers"] = proto.NewRoute(http.MethodGet, "/users", "application/json", userController.GetAllUsers, middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetUserByID"] = proto.NewRoute(http.MethodGet, "/users/:id", "application/json", userController.GetUserByID, middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdateUser"] = proto.NewRoute(http.MethodPut, "/users/:id", "application/json", userController.UpdateUser, middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteUser"] = proto.NewRoute(http.MethodDelete, "/users/:id", "application/json", userController.DeleteUser, middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
