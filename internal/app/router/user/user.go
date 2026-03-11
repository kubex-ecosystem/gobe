// Package user provides the user-related routes for the application.
package user

import (
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/logz"

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
	// rtl := *rtr

	// dbService := rtl.GetDatabaseService()
	// if dbService == nil {
	// 	gl.Log("error", "Database service is nil for AuthRoute")
	// 	return nil
	// }
	// dbGorm, err := dbService.GetDB()
	// if err != nil {
	// 	gl.Log("error", "Failed to get DB from service", err)
	// 	return nil
	// }
	// userController := users.NewUserController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	// middlewaresMap := rtl.GetMiddlewares()

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["LoginRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/sign-in", "application/json", nil, nil, nil, nil, nil)
	routesMap["LogoutRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/sign-out", "application/json", nil, nil, nil, nil, nil)
	routesMap["RefreshRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/check", "application/json", nil, nil, nil, nil, nil)
	routesMap["RegisterRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/sign-up", "application/json", nil, nil, nil, nil, nil)

	return routesMap
}

func NewUserRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for UserRoute")
		return nil
	}
	// rtl := *rtr

	// dbService := rtl.GetDatabaseService()
	// if dbService == nil {
	// 	gl.Log("error", "Database service is nil for UserRoute")
	// 	return nil
	// }
	// dbGorm, err := dbService.GetDB()
	// if err != nil {
	// 	gl.Log("error", "Failed to get DB from service", err)
	// 	return nil
	// }
	// userController := users.NewUserController(dbGorm)

	routesMap := make(map[string]ar.IRoute)

	// middlewaresMap := rtl.GetMiddlewares()

	// secureProperties := make(map[string]bool)
	// secureProperties["secure"] = true
	// secureProperties["validateAndSanitize"] = false
	// secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetAllUsers"] = proto.NewRoute(http.MethodGet, "/users", "application/json", nil, nil, nil, nil, nil)
	routesMap["GetUserByID"] = proto.NewRoute(http.MethodGet, "/users/:id", "application/json", nil, nil, nil, nil, nil)
	routesMap["UpdateUser"] = proto.NewRoute(http.MethodPut, "/users/:id", "application/json", nil, nil, nil, nil, nil)
	routesMap["DeleteUser"] = proto.NewRoute(http.MethodDelete, "/users/:id", "application/json", nil, nil, nil, nil, nil)

	return routesMap
}
