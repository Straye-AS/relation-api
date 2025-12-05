package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type CustomerHandler struct {
	customerService *service.CustomerService
	contactService  *service.ContactService
	logger          *zap.Logger
}

func NewCustomerHandler(customerService *service.CustomerService, contactService *service.ContactService, logger *zap.Logger) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
		contactService:  contactService,
		logger:          logger,
	}
}

// @Summary List customers
// @Tags Customers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param search query string false "Search query"
// @Success 200 {object} domain.PaginatedResponse
// @Router /customers [get]
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	search := r.URL.Query().Get("search")

	result, err := h.customerService.List(r.Context(), page, pageSize, search)
	if err != nil {
		h.logger.Error("failed to list customers", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// @Summary Create customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body domain.CreateCustomerRequest true "Customer data"
// @Success 201 {object} domain.CustomerDTO
// @Router /customers [post]
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	customer, err := h.customerService.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create customer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/customers/"+customer.ID.String())
	respondJSON(w, http.StatusCreated, customer)
}

// @Summary Get customer
// @Tags Customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} domain.CustomerDTO
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	customer, err := h.customerService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// @Summary Update customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body domain.UpdateCustomerRequest true "Customer data"
// @Success 200 {object} domain.CustomerDTO
// @Router /customers/{id} [put]
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var req domain.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	customer, err := h.customerService.Update(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update customer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// @Summary Delete customer
// @Tags Customers
// @Param id path string true "Customer ID"
// @Success 204
// @Router /customers/{id} [delete]
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	if err := h.customerService.Delete(r.Context(), id); err != nil {
		h.logger.Error("failed to delete customer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary List customer contacts
// @Tags Customers
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {array} domain.ContactDTO
// @Router /customers/{id}/contacts [get]
func (h *CustomerHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	contacts, err := h.contactService.ListByCustomer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to list contacts", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, contacts)
}

// @Summary Create contact for customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body domain.CreateContactRequest true "Contact data"
// @Success 201 {object} domain.ContactDTO
// @Router /customers/{id}/contacts [post]
func (h *CustomerHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	var req domain.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set the primary customer ID from the URL
	req.PrimaryCustomerID = &customerID

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	contact, err := h.contactService.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create contact", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, contact)
}
