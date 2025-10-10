// Package customers provides the controller for managing customer-related operations.
package customers

import (
	"context"
	"encoding/json"
	"net/http"

	mdl "github.com/kubex-ecosystem/gdbase/factory/models"
	gdbasez "github.com/kubex-ecosystem/gobe/internal/bridges/gdbasez"
	gl "github.com/kubex-ecosystem/gobe/internal/module/kbx"

	t "github.com/kubex-ecosystem/gobe/internal/contracts/types"
)

type CustomerController struct {
	customerService mdl.ClientService
	APIWrapper      *t.APIWrapper[mdl.ClientModel]
}

type (
	// ErrorResponse padroniza respostas de erro dos endpoints de clientes.
	ErrorResponse = t.ErrorResponse
)

func NewCustomerController(bridge *gdbasez.Bridge) *CustomerController {
	clientRepo := mdl.NewClientRepo(context.Background(), bridge.DBService().(*gdbasez.DBServiceImpl))
	if clientRepo == nil {
		gl.Log("error", "Failed to create ClientRepo")
		return nil
	}
	clientService := mdl.NewClientService(clientRepo)
	return &CustomerController{
		customerService: clientService,
		APIWrapper:      t.NewAPIWrapper[mdl.ClientModel](),
	}
}

// GetAllCustomers retorna todos os clientes cadastrados.
//
// @Summary     Listar clientes
// @Description Recupera a coleção de clientes registrados. [Em desenvolvimento]
// @Tags        customers beta
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} ClientModel
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/customers [get]
func (cc *CustomerController) GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := cc.customerService.ListClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customers)
}

// GetCustomerByID retorna um cliente pelo ID informado.
//
// @Summary     Obter cliente
// @Description Busca um cliente específico utilizando o identificador no caminho. [Em desenvolvimento]
// @Tags        customers beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cliente"
// @Success     200 {object} ClientModel
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Router      /api/v1/customers/{id} [get]
func (cc *CustomerController) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	customer, err := cc.customerService.GetClientByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

// CreateCustomer adiciona um novo cliente.
//
// @Summary     Criar cliente
// @Description Persiste um novo cliente com os dados enviados no corpo. [Em desenvolvimento]
// @Tags        customers beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       payload body ClientModel true "Dados do cliente"
// @Success     200 {object} ClientModel
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/customers [post]
func (cc *CustomerController) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customerRequest gdbasez.ClientModel
	if err := json.NewDecoder(r.Body).Decode(&customerRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdCustomer, err := cc.customerService.CreateClient(&customerRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(createdCustomer)
}

// UpdateCustomer atualiza os dados de um cliente existente.
//
// @Summary     Atualizar cliente
// @Description Atualiza um cliente identificado pelo ID. [Em desenvolvimento]
// @Tags        customers beta
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id      path string           true "ID do cliente"
// @Param       payload body ClientModel true "Dados atualizados"
// @Success     200 {object} ClientModel
// @Failure     400 {object} ErrorResponse
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/customers/{id} [put]
func (cc *CustomerController) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	var customerRequest gdbasez.ClientModel
	if err := json.NewDecoder(r.Body).Decode(&customerRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedCustomer, err := cc.customerService.UpdateClient(&customerRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(updatedCustomer)
}

// DeleteCustomer remove um cliente.
//
// @Summary     Remover cliente
// @Description Exclui o cliente identificado pelo ID informado. [Em desenvolvimento]
// @Tags        customers beta
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID do cliente"
// @Success     204 {string} string "Cliente removido"
// @Failure     401 {object} ErrorResponse
// @Failure     404 {object} ErrorResponse
// @Failure     500 {object} ErrorResponse
// @Router      /api/v1/customers/{id} [delete]
func (cc *CustomerController) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := cc.customerService.DeleteClient(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
