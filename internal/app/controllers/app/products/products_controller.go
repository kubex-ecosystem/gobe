// Package products provides the controller for managing products in the application.
package products

import (
	"encoding/json"
	"net/http"

	fscm "github.com/kubex-ecosystem/gdbase/factory/models"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
	"gorm.io/gorm"
)

type ProductController struct {
	productService fscm.ProductService
	APIWrapper     *t.APIWrapper[fscm.ProductModel]
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{
		productService: fscm.NewProductService(fscm.NewProductRepo(db)),
		APIWrapper:     t.NewAPIWrapper[fscm.ProductModel](),
	}
}

func (pc *ProductController) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := pc.productService.ListProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(products)
}

func (pc *ProductController) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(product)
}

func (pc *ProductController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var productRequest fscm.ProductModel
	if err := json.NewDecoder(r.Body).Decode(&productRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdProduct, err := pc.productService.CreateProduct(&productRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(createdProduct)
}

func (pc *ProductController) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var productRequest fscm.ProductModel
	if err := json.NewDecoder(r.Body).Decode(&productRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedProduct, err := pc.productService.UpdateProduct(&productRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedProduct)
}

func (pc *ProductController) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := pc.productService.DeleteProduct(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
