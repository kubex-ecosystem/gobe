// Package products provides the controller for managing products in the application.
package products

import (
	"encoding/json"
	"net/http"

	svc "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

type ProductController struct {
	productService svc.ProductService
	APIWrapper     *t.APIWrapper[svc.ProductModel]
}

type (
	// ErrorResponse padroniza a documentação de erros dos endpoints de produtos.
	ErrorResponse = t.ErrorResponse
)

func NewProductController(bridge *svc.Bridge) *ProductController {
	return &ProductController{
		productService: bridge.ProductService(),
		APIWrapper:     t.NewAPIWrapper[svc.ProductModel](),
	}
}

// GetAllProducts retorna todos os produtos disponíveis.
//
// @Summary     Listar produtos
// @Description Recupera a lista de produtos registrados na base. [Em desenvolvimento]
// @Tags        products beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} svc.ProductModel
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/products [get]
func (pc *ProductController) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := pc.productService.ListProducts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(products)
}

// GetProductByID retorna um produto pelo identificador informado.
//
// @Summary     Obter produto
// @Description Busca um produto específico pelo ID informado no caminho. [Em desenvolvimento]
// @Tags        products beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do produto"
// @Success     200 {object} svc.ProductModel
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/products/{id} [get]
func (pc *ProductController) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	product, err := pc.productService.GetProductByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(product)
}

// CreateProduct cria um novo produto na base.
//
// @Summary     Criar produto
// @Description Persiste um novo produto com os dados enviados no corpo. [Em desenvolvimento]
// @Tags        products beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body svc.ProductModel true "Dados do produto"
// @Success     200 {object} svc.ProductModel
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/products [post]
func (pc *ProductController) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var productRequest svc.ProductModel
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

// UpdateProduct atualiza um produto existente.
//
// @Summary     Atualizar produto
// @Description Atualiza os dados de um produto identificado por ID. [Em desenvolvimento]
// @Tags        products beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string            true "ID do produto"
// @Param       payload body svc.ProductModel true "Dados atualizados"
// @Success     200 {object} svc.ProductModel
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/products/{id} [put]
func (pc *ProductController) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var productRequest svc.ProductModel
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

// DeleteProduct remove um produto da base de dados.
//
// @Summary     Remover produto
// @Description Exclui um produto identificado pelo ID informado. [Em desenvolvimento]
// @Tags        products beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do produto"
// @Success     204 {string} string "Produto removido"
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/products/{id} [delete]
func (pc *ProductController) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := pc.productService.DeleteProduct(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
