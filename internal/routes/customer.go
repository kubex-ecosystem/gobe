package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	customers_controller "github.com/rafa-mori/gobe/internal/controllers/customers"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
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
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	customerController := customers_controller.NewCustomerController(dbGorm)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routes := map[string]ar.IRoute{
		"GetAllCustomers": NewRoute(http.MethodGet, "/customers", "application/json", gin.WrapF(customerController.GetAllCustomers), nil, dbService, secureProperties),
		"GetCustomerByID": NewRoute(http.MethodGet, "/customers/:id", "application/json", gin.WrapF(customerController.GetCustomerByID), nil, dbService, secureProperties),
		"CreateCustomer":  NewRoute(http.MethodPost, "/customers", "application/json", gin.WrapF(customerController.CreateCustomer), nil, dbService, secureProperties),
		"UpdateCustomer":  NewRoute(http.MethodPut, "/customers/:id", "application/json", gin.WrapF(customerController.UpdateCustomer), nil, dbService, secureProperties),
		"DeleteCustomer":  NewRoute(http.MethodDelete, "/customers/:id", "application/json", gin.WrapF(customerController.DeleteCustomer), nil, dbService, secureProperties),
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
