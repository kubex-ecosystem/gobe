package app

import (
	"context"
	"net/http"

	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"

	"github.com/gin-gonic/gin"
	customers_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/app/customers"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
)

type CustomerRoutes struct {
	ar.IRouter

	GetCustomersRoute   ar.IRoute
	GetCustomerRoute    ar.IRoute
	CreateCustomerRoute ar.IRoute
	UpdateCustomerRoute ar.IRoute
	DeleteCustomerRoute ar.IRoute

	GetCustomerOrdersRoute   ar.IRoute
	GetCustomerOrderRoute    ar.IRoute
	CreateCustomerOrderRoute ar.IRoute
	UpdateCustomerOrderRoute ar.IRoute
	DeleteCustomerOrderRoute ar.IRoute
}

func NewCustomerRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		gl.Log("error", "Router is nil for CustomerRoute")
		return nil
	}
	rtl := *rtr
	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for CustomerRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB(context.Background(), gdbasez.DefaultDBName)
	bridge := gdbasez.NewBridge(dbGorm)
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	customerController := customers_controller.NewCustomerController(bridge)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routes := map[string]ar.IRoute{
		"GetAllCustomers": proto.NewRoute(http.MethodGet, "/api/v1/customers", "application/json", gin.WrapF(customerController.GetAllCustomers), nil, dbService, secureProperties, nil),
		"GetCustomerByID": proto.NewRoute(http.MethodGet, "/api/v1/customers/:id", "application/json", gin.WrapF(customerController.GetCustomerByID), nil, dbService, secureProperties, nil),
		"CreateCustomer":  proto.NewRoute(http.MethodPost, "/api/v1/customers", "application/json", gin.WrapF(customerController.CreateCustomer), nil, dbService, secureProperties, nil),
		"UpdateCustomer":  proto.NewRoute(http.MethodPut, "/api/v1/customers/:id", "application/json", gin.WrapF(customerController.UpdateCustomer), nil, dbService, secureProperties, nil),
		"DeleteCustomer":  proto.NewRoute(http.MethodDelete, "/api/v1/customers/:id", "application/json", gin.WrapF(customerController.DeleteCustomer), nil, dbService, secureProperties, nil),
	}
	return routes
}

func (a *CustomerRoutes) DummyPlaceHolder(_ chan interface{}) gin.HandlerFunc {
	if a == nil {
		return nil
	}
	return func(c *gin.Context) {
		gl.Log("info", "Sending Dummy PlaceHolder context to data Channel")
		c.JSON(http.StatusOK, gin.H{"message": "Dummy PlaceHolder"})
	}
}
