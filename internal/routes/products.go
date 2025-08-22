package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	products_controller "github.com/rafa-mori/gobe/internal/controllers/apps/products"
	gl "github.com/rafa-mori/gobe/internal/module/logger"
	ar "github.com/rafa-mori/gobe/internal/proto/interfaces"
	l "github.com/rafa-mori/logz"
)

type ProductRoutes struct {
	ar.IRouter
}

func NewProductRoutes(rtr *ar.IRouter) map[string]ar.IRoute {
	if rtr == nil {
		l.ErrorCtx("Router is nil for ProductRoute", nil)
		return nil
	}
	rtl := *rtr

	dbService := rtl.GetDatabaseService()
	if dbService == nil {
		gl.Log("error", "Database service is nil for ProductRoute")
		return nil
	}
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	productController := products_controller.NewProductController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]gin.HandlerFunc)

	secureProperties := make(map[string]bool)
	secureProperties["secure"] = true
	secureProperties["validateAndSanitize"] = false
	secureProperties["validateAndSanitizeBody"] = false

	routesMap["GetProductsRoute"] = NewRoute(http.MethodGet, "/api/v1/products", "application/json", gin.WrapF(productController.GetAllProducts), middlewaresMap, dbService, secureProperties)
	routesMap["GetProductRoute"] = NewRoute(http.MethodGet, "/api/v1/products/:id", "application/json", gin.WrapF(productController.GetProductByID), middlewaresMap, dbService, secureProperties)
	routesMap["CreateProductRoute"] = NewRoute(http.MethodPost, "/api/v1/products", "application/json", gin.WrapF(productController.CreateProduct), middlewaresMap, dbService, secureProperties)
	routesMap["UpdateProductRoute"] = NewRoute(http.MethodPut, "/api/v1/products/:id", "application/json", gin.WrapF(productController.UpdateProduct), middlewaresMap, dbService, secureProperties)
	routesMap["DeleteProductRoute"] = NewRoute(http.MethodDelete, "/api/v1/products/:id", "application/json", gin.WrapF(productController.DeleteProduct), middlewaresMap, dbService, secureProperties)

	return routesMap
}
