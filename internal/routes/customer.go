package routes

import (
	"net/http"

	l "github.com/faelmori/logz"
	"github.com/gin-gonic/gin"
	customers_controller "github.com/rafa-mori/gobe/internal/controllers/customers"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
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
		l.ErrorCtx("Router is nil", nil)
		return nil
	}
	rtl := *rtr
	dbService := rtl.GetDatabaseService()
	dbGorm, err := dbService.GetDB()
	if err != nil {
		l.ErrorCtx("Failed to get DB from service", map[string]interface{}{"error": err.Error()})
		return nil
	}
	customerController := customers_controller.NewCustomerController(dbGorm)

	routes := map[string]ar.IRoute{
		"GetAllCustomers": NewRoute(http.MethodGet, "/customers", "application/json", gin.WrapF(customerController.GetAllCustomers), nil, dbService),
		"GetCustomerByID": NewRoute(http.MethodGet, "/customers/:id", "application/json", gin.WrapF(customerController.GetCustomerByID), nil, dbService),
		"CreateCustomer":  NewRoute(http.MethodPost, "/customers", "application/json", gin.WrapF(customerController.CreateCustomer), nil, dbService),
		"UpdateCustomer":  NewRoute(http.MethodPut, "/customers/:id", "application/json", gin.WrapF(customerController.UpdateCustomer), nil, dbService),
		"DeleteCustomer":  NewRoute(http.MethodDelete, "/customers/:id", "application/json", gin.WrapF(customerController.DeleteCustomer), nil, dbService),
	}
	return routes
}

func (a *CustomerRoutes) DummyPlaceHolder(_ chan interface{}) gin.HandlerFunc {
	if a == nil {
		return nil
	}
	return func(c *gin.Context) {
		l.NoticeCtx("Sending Dummy PlaceHolder context to data Channel", nil)
		c.JSON(http.StatusOK, gin.H{"message": "Dummy PlaceHolder"})
	}
}
