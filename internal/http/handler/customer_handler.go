package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
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

// List godoc
// @Summary List customers
// @Description Get paginated list of customers with optional filters
// @Tags Customers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param search query string false "Search by name or organization number"
// @Param city query string false "Filter by city"
// @Param country query string false "Filter by country"
// @Param status query string false "Filter by status" Enums(active, inactive, lead, churned)
// @Param tier query string false "Filter by tier" Enums(bronze, silver, gold, platinum)
// @Param industry query string false "Filter by industry" Enums(construction, manufacturing, retail, logistics, agriculture, energy, public_sector, real_estate, other)
// @Param sortBy query string false "Sort option" Enums(name_asc, name_desc, created_desc, created_asc, city_asc, city_desc)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.CustomerDTO}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
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

	// Build filters
	filters := &repository.CustomerFilters{
		Search:  r.URL.Query().Get("search"),
		City:    r.URL.Query().Get("city"),
		Country: r.URL.Query().Get("country"),
	}

	// Parse optional status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.CustomerStatus(status)
		filters.Status = &s
	}

	// Parse optional tier filter
	if tier := r.URL.Query().Get("tier"); tier != "" {
		t := domain.CustomerTier(tier)
		filters.Tier = &t
	}

	// Parse optional industry filter
	if industry := r.URL.Query().Get("industry"); industry != "" {
		i := domain.CustomerIndustry(industry)
		filters.Industry = &i
	}

	// Parse sort option
	sortBy := repository.CustomerSortByCreatedDesc
	if s := r.URL.Query().Get("sortBy"); s != "" {
		sortBy = repository.CustomerSortOption(s)
	}

	result, err := h.customerService.ListWithFilters(r.Context(), page, pageSize, filters, sortBy)
	if err != nil {
		h.logger.Error("failed to list customers", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list customers",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// FuzzySearch godoc
// @Summary Fuzzy search for best matching customer
// @Description Find the single best matching customer for a query using fuzzy matching (handles typos, abbreviations, partial matches). Use q=all to get all customers. Returns minimal customer data (id and name only). Also supports email domain matching (e.g., 'hauk@straye.no' matches 'Straye').
// @Tags Customers
// @Accept json
// @Produce json
// @Param q query string true "Search query (e.g., 'AF', 'NTN', 'Veidikke', 'all' for all customers, or email like 'user@company.no')"
// @Success 200 {object} domain.FuzzyCustomerSearchResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/search [get]
func (h *CustomerHandler) FuzzySearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Query parameter 'q' is required",
		})
		return
	}

	result, err := h.customerService.FuzzySearchBestMatch(r.Context(), query)
	if err != nil {
		// Check for validation errors (400) vs internal errors (500)
		if strings.Contains(err.Error(), "query too long") {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		h.logger.Error("failed to fuzzy search customer", zap.Error(err), zap.String("query", query))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to search for customer",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary Get customer by ID
// @Description Get a customer with full details including stats, contacts, active deals, and projects
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 200 {object} domain.CustomerWithDetailsDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	customer, err := h.customerService.GetByIDWithDetails(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to get customer", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get customer",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// Create godoc
// @Summary Create customer
// @Description Create a new customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body domain.CreateCustomerRequest true "Customer data"
// @Success 201 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Duplicate organization number"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers [post]
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	customer, err := h.customerService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateOrgNumber) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: "A customer with this organization number already exists",
			})
			return
		}
		h.logger.Error("failed to create customer", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create customer",
		})
		return
	}

	w.Header().Set("Location", "/api/v1/customers/"+customer.ID.String())
	respondJSON(w, http.StatusCreated, customer)
}

// Update godoc
// @Summary Update customer
// @Description Update an existing customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerRequest true "Customer data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Duplicate organization number"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id} [put]
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	customer, err := h.customerService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		if errors.Is(err, service.ErrDuplicateOrgNumber) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: "A customer with this organization number already exists",
			})
			return
		}
		h.logger.Error("failed to update customer", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// Delete godoc
// @Summary Delete customer
// @Description Soft delete a customer. Cannot delete customers with active projects, deals, or offers.
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Customer has active dependencies"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id} [delete]
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	if err := h.customerService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		if errors.Is(err, service.ErrCustomerHasActiveDependencies) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
			})
			return
		}
		h.logger.Error("failed to delete customer", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete customer",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListContacts godoc
// @Summary List customer contacts
// @Description Get all contacts associated with a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Success 200 {array} domain.ContactDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/contacts [get]
func (h *CustomerHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	contacts, err := h.contactService.ListByCustomer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to list contacts", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list contacts",
		})
		return
	}

	respondJSON(w, http.StatusOK, contacts)
}

// CreateContact godoc
// @Summary Create contact for customer
// @Description Create a new contact and associate it with the specified customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.CreateContactRequest true "Contact data"
// @Success 201 {object} domain.ContactDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/contacts [post]
func (h *CustomerHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
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
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create contact",
		})
		return
	}

	respondJSON(w, http.StatusCreated, contact)
}
