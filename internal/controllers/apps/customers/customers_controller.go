// Package customers provides the controller for managing customer-related operations.
package customers

import (
	"encoding/json"
	"net/http"

	fscm "github.com/rafa-mori/gdbase/factory/models"
	t "github.com/rafa-mori/gobe/internal/proto/types"
	"gorm.io/gorm"
)

type CustomerController struct {
	customerService fscm.ClientService
	APIWrapper      *t.APIWrapper[fscm.ClientModel]
}

func NewCustomerController(db *gorm.DB) *CustomerController {
	return &CustomerController{
		customerService: fscm.NewClientService(fscm.NewClientRepo(db)),
		APIWrapper:      t.NewApiWrapper[fscm.ClientModel](),
	}
}

func (cc *CustomerController) GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := cc.customerService.ListClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(customers)
}

func (cc *CustomerController) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	customer, err := cc.customerService.GetClientByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(customer)
}

func (cc *CustomerController) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customerRequest fscm.ClientModel
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

func (cc *CustomerController) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	var customerRequest fscm.ClientModel
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

func (cc *CustomerController) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := cc.customerService.DeleteClient(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
