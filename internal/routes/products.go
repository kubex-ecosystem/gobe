package routes

import (
	"net/http"

	"github.com/girafa-morigin"
	products_controller "github.com/rafa-mori/gobe/internal/controllers/products"
	ar "github.com/rafa-mori/gobe/internal/interfaces"
	gl "github.com/rafa-mori/gobe/logger"
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
	dbGorm, err := dbService.GetDB()
	if err != nil {
		gl.Log("error", "Failed to get DB from service", err)
		return nil
	}
	productController := products_controller.NewProductController(dbGorm)

	routesMap := make(map[string]ar.IRoute)
	middlewaresMap := make(map[string]any)

	routesMap["GetProductsRoute"] = NewRoute(http.MethodGet, "/products", "application/json", gin.WrapF(productController.GetAllProducts), middlewaresMap, dbService)
	routesMap["GetProductRoute"] = NewRoute(http.MethodGet, "/products/:id", "application/json", gin.WrapF(productController.GetProductByID), middlewaresMap, dbService)
	routesMap["CreateProductRoute"] = NewRoute(http.MethodPost, "/products", "application/json", gin.WrapF(productController.CreateProduct), middlewaresMap, dbService)
	routesMap["UpdateProductRoute"] = NewRoute(http.MethodPut, "/products/:id", "application/json", gin.WrapF(productController.UpdateProduct), middlewaresMap, dbService)
	routesMap["DeleteProductRoute"] = NewRoute(http.MethodDelete, "/products/:id", "application/json", gin.WrapF(productController.DeleteProduct), middlewaresMap, dbService)

	return routesMap
}
