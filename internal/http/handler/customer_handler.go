package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	offerService    *service.OfferService
	projectService  *service.ProjectService
	logger          *zap.Logger
}

func NewCustomerHandler(
	customerService *service.CustomerService,
	contactService *service.ContactService,
	offerService *service.OfferService,
	projectService *service.ProjectService,
	logger *zap.Logger,
) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
		contactService:  contactService,
		offerService:    offerService,
		projectService:  projectService,
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
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, name, city, country, status, tier, industry, orgNumber)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
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

	// Parse sort configuration
	sort := repository.DefaultSortConfig()
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		sort.Order = repository.ParseSortOrder(sortOrder)
	}

	result, err := h.customerService.ListWithSort(r.Context(), page, pageSize, filters, sort)
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

// UpdateStatus godoc
// @Summary Update customer status
// @Description Update only the status of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerStatusRequest true "Status data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/status [put]
func (h *CustomerHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerStatusRequest
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

	customer, err := h.customerService.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer status", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer status",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateTier godoc
// @Summary Update customer tier
// @Description Update only the tier of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerTierRequest true "Tier data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/tier [put]
func (h *CustomerHandler) UpdateTier(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerTierRequest
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

	customer, err := h.customerService.UpdateTier(r.Context(), id, req.Tier)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer tier", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer tier",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateIndustry godoc
// @Summary Update customer industry
// @Description Update only the industry of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerIndustryRequest true "Industry data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/industry [put]
func (h *CustomerHandler) UpdateIndustry(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerIndustryRequest
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

	customer, err := h.customerService.UpdateIndustry(r.Context(), id, req.Industry)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer industry", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer industry",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateNotes godoc
// @Summary Update customer notes
// @Description Update only the notes of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerNotesRequest true "Notes data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/notes [put]
func (h *CustomerHandler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	customer, err := h.customerService.UpdateNotes(r.Context(), id, req.Notes)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer notes", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer notes",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateCompany godoc
// @Summary Update customer company assignment
// @Description Assign or unassign a customer to/from a company
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerCompanyRequest true "Company data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/company [put]
func (h *CustomerHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerCompanyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	customer, err := h.customerService.UpdateCompanyID(r.Context(), id, req.CompanyID)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer company", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer company",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateCustomerClass godoc
// @Summary Update customer class
// @Description Update only the customer class
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerClassRequest true "Customer class data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/customer-class [put]
func (h *CustomerHandler) UpdateCustomerClass(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerClassRequest
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

	customer, err := h.customerService.UpdateCustomerClass(r.Context(), id, req.CustomerClass)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer class", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer class",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateCreditLimit godoc
// @Summary Update customer credit limit
// @Description Update only the credit limit of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerCreditLimitRequest true "Credit limit data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/credit-limit [put]
func (h *CustomerHandler) UpdateCreditLimit(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerCreditLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	customer, err := h.customerService.UpdateCreditLimit(r.Context(), id, req.CreditLimit)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer credit limit", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer credit limit",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateIsInternal godoc
// @Summary Update customer internal flag
// @Description Update the internal flag of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerIsInternalRequest true "Internal flag data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/is-internal [put]
func (h *CustomerHandler) UpdateIsInternal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerIsInternalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	customer, err := h.customerService.UpdateIsInternal(r.Context(), id, req.IsInternal)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer internal flag", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer internal flag",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateAddress godoc
// @Summary Update customer address
// @Description Update the address fields of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerAddressRequest true "Address data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/address [put]
func (h *CustomerHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerAddressRequest
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

	customer, err := h.customerService.UpdateAddress(r.Context(), id, req.Address, req.City, req.PostalCode, req.Country)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer address", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer address",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateContactInfo godoc
// @Summary Update customer contact information
// @Description Update the contact person, email, and phone of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerContactInfoRequest true "Contact info data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/contact-info [put]
func (h *CustomerHandler) UpdateContactInfo(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerContactInfoRequest
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

	customer, err := h.customerService.UpdateContactInfo(r.Context(), id, req.ContactPerson, req.ContactEmail, req.ContactPhone)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		if errors.Is(err, service.ErrInvalidEmailFormat) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrInvalidPhoneFormat) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		h.logger.Error("failed to update customer contact info", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer contact info",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdatePostalCode godoc
// @Summary Update customer postal code
// @Description Update only the postal code of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerPostalCodeRequest true "Postal code data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/postal-code [put]
func (h *CustomerHandler) UpdatePostalCode(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerPostalCodeRequest
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

	customer, err := h.customerService.UpdatePostalCode(r.Context(), id, req.PostalCode)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer postal code", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer postal code",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// UpdateCity godoc
// @Summary Update customer city
// @Description Update only the city of a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param request body domain.UpdateCustomerCityRequest true "City data"
// @Success 200 {object} domain.CustomerDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/city [put]
func (h *CustomerHandler) UpdateCity(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	var req domain.UpdateCustomerCityRequest
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

	customer, err := h.customerService.UpdateCity(r.Context(), id, req.City)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Error("failed to update customer city", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update customer city",
		})
		return
	}

	respondJSON(w, http.StatusOK, customer)
}

// ListOffers godoc
// @Summary List offers for a customer
// @Description Get paginated list of offers associated with a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param phase query string false "Filter by phase" Enums(draft, in_progress, sent, won, lost, expired)
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, title, value, probability, phase, status, dueDate, customerName)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.OfferDTO}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/offers [get]
func (h *CustomerHandler) ListOffers(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}

	// Parse optional phase filter
	var phase *domain.OfferPhase
	if phaseStr := r.URL.Query().Get("phase"); phaseStr != "" {
		p := domain.OfferPhase(phaseStr)
		phase = &p
	}

	// Parse sort configuration
	sort := repository.DefaultSortConfig()
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		sort.Order = repository.ParseSortOrder(sortOrder)
	}

	result, err := h.offerService.ListWithSort(r.Context(), page, pageSize, &customerID, nil, phase, sort)
	if err != nil {
		h.logger.Error("failed to list offers for customer", zap.String("customerID", customerID.String()), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list offers",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ListProjects godoc
// @Summary List projects for a customer
// @Description Get paginated list of projects associated with a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param status query string false "Filter by status" Enums(active, completed, cancelled, on_hold)
// @Param phase query string false "Filter by phase" Enums(tilbud, active, working, completed, cancelled)
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, name, status, phase, health, budget, spent, startDate, endDate, customerName, wonAt)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.ProjectDTO}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /customers/{id}/projects [get]
func (h *CustomerHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid customer ID format",
		})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}

	// Build filters
	filters := &repository.ProjectFilters{
		CustomerID: &customerID,
	}

	// Parse optional status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ProjectStatus(status)
		filters.Status = &s
	}

	// Parse optional phase filter
	if phase := r.URL.Query().Get("phase"); phase != "" {
		p := domain.ProjectPhase(phase)
		filters.Phase = &p
	}

	// Parse sort configuration
	sort := repository.DefaultSortConfig()
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		sort.Order = repository.ParseSortOrder(sortOrder)
	}

	result, err := h.projectService.ListWithSort(r.Context(), page, pageSize, filters, sort)
	if err != nil {
		h.logger.Error("failed to list projects for customer", zap.String("customerID", customerID.String()), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list projects",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}
