package routes

import (
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
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	userController := users.NewUserController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["LoginRoute"] = NewRoute(http.MethodPost, "/sign-in", "application/json", userController.AuthenticateUser, middlewaresMap, dbService)
	routesMap["LogoutRoute"] = NewRoute(http.MethodPost, "/sign-out", "application/json", userController.Logout, middlewaresMap, dbService)
	routesMap["RefreshRoute"] = NewRoute(http.MethodPost, "check", "application/json", userController.RefreshToken, middlewaresMap, dbService)
	routesMap["RegisterRoute"] = NewRoute(http.MethodPost, "/sign-up", "application/json", userController.CreateUser, middlewaresMap, dbService)

	return routesMap
}

func NewUserRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for UserRoute")
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	userController := users.NewUserController(dbGorm)

	routesMap := make(map[string]ar.IRoute)

	routesMap["GetAllUsers"] = NewRoute(http.MethodGet, "/users", "application/json", userController.GetAllUsers, nil, nil)
	routesMap["GetUserByID"] = NewRoute(http.MethodGet, "/users/:id", "application/json", userController.GetUserByID, nil, nil)
	routesMap["UpdateUser"] = NewRoute(http.MethodPut, "/users/:id", "application/json", userController.UpdateUser, nil, nil)
	routesMap["DeleteUser"] = NewRoute(http.MethodDelete, "/users/:id", "application/json", userController.DeleteUser, nil, nil)

	return routesMap
}
