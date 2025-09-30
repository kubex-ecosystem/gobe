// Package app contains the application routes for products.
package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	products_controller "github.com/kubex-ecosystem/gobe/internal/app/controllers/app/products"
	proto "github.com/kubex-ecosystem/gobe/internal/app/router/types"
	ar "github.com/kubex-ecosystem/gobe/internal/contracts/interfaces"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"
	l "github.com/kubex-ecosystem/logz"
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

	routesMap["GetProductsRoute"] = proto.NewRoute(http.MethodGet, "/api/v1/products", "application/json", gin.WrapF(productController.GetAllProducts), middlewaresMap, dbService, secureProperties, nil)
	routesMap["GetProductRoute"] = proto.NewRoute(http.MethodGet, "/api/v1/products/:id", "application/json", gin.WrapF(productController.GetProductByID), middlewaresMap, dbService, secureProperties, nil)
	routesMap["CreateProductRoute"] = proto.NewRoute(http.MethodPost, "/api/v1/products", "application/json", gin.WrapF(productController.CreateProduct), middlewaresMap, dbService, secureProperties, nil)
	routesMap["UpdateProductRoute"] = proto.NewRoute(http.MethodPut, "/api/v1/products/:id", "application/json", gin.WrapF(productController.UpdateProduct), middlewaresMap, dbService, secureProperties, nil)
	routesMap["DeleteProductRoute"] = proto.NewRoute(http.MethodDelete, "/api/v1/products/:id", "application/json", gin.WrapF(productController.DeleteProduct), middlewaresMap, dbService, secureProperties, nil)

	return routesMap
}
