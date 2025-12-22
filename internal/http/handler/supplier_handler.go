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

// SupplierHandler handles HTTP requests for supplier operations
type SupplierHandler struct {
	supplierService *service.SupplierService
	logger          *zap.Logger
}

// NewSupplierHandler creates a new supplier handler instance
func NewSupplierHandler(
	supplierService *service.SupplierService,
	logger *zap.Logger,
) *SupplierHandler {
	return &SupplierHandler{
		supplierService: supplierService,
		logger:          logger,
	}
}

// List godoc
// @Summary List suppliers
// @Description Get paginated list of suppliers with optional filters
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param search query string false "Search by name or organization number"
// @Param city query string false "Filter by city"
// @Param country query string false "Filter by country"
// @Param status query string false "Filter by status" Enums(active, inactive, pending, blacklisted)
// @Param category query string false "Filter by category"
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, name, city, country, status, category, orgNumber)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.SupplierDTO}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers [get]
func (h *SupplierHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}

	// Build filters
	filters := &repository.SupplierFilters{
		Search:   r.URL.Query().Get("search"),
		City:     r.URL.Query().Get("city"),
		Country:  r.URL.Query().Get("country"),
		Category: r.URL.Query().Get("category"),
	}

	// Parse optional status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.SupplierStatus(status)
		filters.Status = &s
	}

	// Parse sort configuration
	sort := repository.DefaultSortConfig()
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		sort.Order = repository.ParseSortOrder(sortOrder)
	}

	result, err := h.supplierService.ListWithSort(r.Context(), page, pageSize, filters, sort)
	if err != nil {
		h.logger.Error("failed to list suppliers", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list suppliers",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetByID godoc
// @Summary Get supplier by ID
// @Description Get a supplier with full details including stats, contacts, and recent offers
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Success 200 {object} domain.SupplierWithDetailsDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id} [get]
func (h *SupplierHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	supplier, err := h.supplierService.GetByIDWithDetails(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to get supplier", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get supplier",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// Create godoc
// @Summary Create supplier
// @Description Create a new supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param request body domain.CreateSupplierRequest true "Supplier data"
// @Success 201 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Duplicate organization number"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers [post]
func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSupplierRequest
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

	supplier, err := h.supplierService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateSupplierOrgNumber) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: "A supplier with this organization number already exists",
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
		h.logger.Error("failed to create supplier", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create supplier",
		})
		return
	}

	w.Header().Set("Location", "/api/v1/suppliers/"+supplier.ID.String())
	respondJSON(w, http.StatusCreated, supplier)
}

// Update godoc
// @Summary Update supplier
// @Description Update an existing supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierRequest true "Supplier data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Duplicate organization number"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id} [put]
func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierRequest
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

	supplier, err := h.supplierService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		if errors.Is(err, service.ErrDuplicateSupplierOrgNumber) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: "A supplier with this organization number already exists",
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
		h.logger.Error("failed to update supplier", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// Delete godoc
// @Summary Delete supplier
// @Description Soft delete a supplier. The supplier is hidden from lists but preserved for historical reference.
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Supplier has active relationships"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id} [delete]
func (h *SupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	if err := h.supplierService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		if errors.Is(err, service.ErrSupplierHasActiveRelations) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
			})
			return
		}
		h.logger.Error("failed to delete supplier", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete supplier",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateStatus godoc
// @Summary Update supplier status
// @Description Update only the status of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierStatusRequest true "Status data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/status [put]
func (h *SupplierHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierStatusRequest
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

	supplier, err := h.supplierService.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier status", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier status",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdateNotes godoc
// @Summary Update supplier notes
// @Description Update only the notes of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierNotesRequest true "Notes data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/notes [put]
func (h *SupplierHandler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	supplier, err := h.supplierService.UpdateNotes(r.Context(), id, req.Notes)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier notes", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier notes",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdateCategory godoc
// @Summary Update supplier category
// @Description Update only the category of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierCategoryRequest true "Category data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/category [put]
func (h *SupplierHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierCategoryRequest
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

	supplier, err := h.supplierService.UpdateCategory(r.Context(), id, req.Category)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier category", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier category",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdatePaymentTerms godoc
// @Summary Update supplier payment terms
// @Description Update only the payment terms of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierPaymentTermsRequest true "Payment terms data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/payment-terms [put]
func (h *SupplierHandler) UpdatePaymentTerms(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierPaymentTermsRequest
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

	supplier, err := h.supplierService.UpdatePaymentTerms(r.Context(), id, req.PaymentTerms)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier payment terms", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier payment terms",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdateEmail godoc
// @Summary Update supplier email
// @Description Update only the email of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierEmailRequest true "Email data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/email [put]
func (h *SupplierHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierEmailRequest
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

	supplier, err := h.supplierService.UpdateEmail(r.Context(), id, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
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
		h.logger.Error("failed to update supplier email", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier email",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdatePhone godoc
// @Summary Update supplier phone
// @Description Update only the phone of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierPhoneRequest true "Phone data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/phone [put]
func (h *SupplierHandler) UpdatePhone(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierPhoneRequest
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

	supplier, err := h.supplierService.UpdatePhone(r.Context(), id, req.Phone)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
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
		h.logger.Error("failed to update supplier phone", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier phone",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdateWebsite godoc
// @Summary Update supplier website
// @Description Update only the website of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierWebsiteRequest true "Website data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/website [put]
func (h *SupplierHandler) UpdateWebsite(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierWebsiteRequest
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

	supplier, err := h.supplierService.UpdateWebsite(r.Context(), id, req.Website)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier website", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier website",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdateAddress godoc
// @Summary Update supplier address
// @Description Update only the address of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierAddressRequest true "Address data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/address [put]
func (h *SupplierHandler) UpdateAddress(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierAddressRequest
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

	supplier, err := h.supplierService.UpdateAddress(r.Context(), id, req.Address)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier address", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier address",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdatePostalCode godoc
// @Summary Update supplier postal code
// @Description Update only the postal code of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierPostalCodeRequest true "Postal code data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/postal-code [put]
func (h *SupplierHandler) UpdatePostalCode(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierPostalCodeRequest
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

	supplier, err := h.supplierService.UpdatePostalCode(r.Context(), id, req.PostalCode)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier postal code", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier postal code",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// UpdateCity godoc
// @Summary Update supplier city
// @Description Update only the city of a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.UpdateSupplierCityRequest true "City data"
// @Success 200 {object} domain.SupplierDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/city [put]
func (h *SupplierHandler) UpdateCity(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.UpdateSupplierCityRequest
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

	supplier, err := h.supplierService.UpdateCity(r.Context(), id, req.City)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to update supplier city", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier city",
		})
		return
	}

	respondJSON(w, http.StatusOK, supplier)
}

// ============================================================================
// Supplier Offers Handler
// ============================================================================

// ListOffers godoc
// @Summary List offers for a supplier
// @Description Get paginated list of offers associated with a supplier via the offer_suppliers junction table
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param phase query string false "Filter by phase" Enums(draft, in_progress, sent, order, completed, won, lost, expired)
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, title, value, probability, phase, status, dueDate, customerName)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.OfferDTO}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/offers [get]
func (h *SupplierHandler) ListOffers(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
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

	result, err := h.supplierService.ListOffers(r.Context(), supplierID, page, pageSize, phase, sort)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to list offers for supplier", zap.String("supplierID", supplierID.String()), zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list offers",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ============================================================================
// Supplier Contact Handlers
// ============================================================================

// ListContacts godoc
// @Summary List contacts for a supplier
// @Description Get all contact persons for a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Success 200 {array} domain.SupplierContactDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/contacts [get]
func (h *SupplierHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	contacts, err := h.supplierService.ListContacts(r.Context(), supplierID)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		h.logger.Error("failed to list supplier contacts", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list supplier contacts",
		})
		return
	}

	respondJSON(w, http.StatusOK, contacts)
}

// GetContact godoc
// @Summary Get a supplier contact by ID
// @Description Get a specific contact person for a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param contactId path string true "Contact ID" format(uuid)
// @Success 200 {object} domain.SupplierContactDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/contacts/{contactId} [get]
func (h *SupplierHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	contactID, err := uuid.Parse(chi.URLParam(r, "contactId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid contact ID format",
		})
		return
	}

	contact, err := h.supplierService.GetContact(r.Context(), supplierID, contactID)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		if errors.Is(err, service.ErrSupplierContactNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Contact not found",
			})
			return
		}
		h.logger.Error("failed to get supplier contact", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get supplier contact",
		})
		return
	}

	respondJSON(w, http.StatusOK, contact)
}

// CreateContact godoc
// @Summary Create a contact for a supplier
// @Description Create a new contact person for a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param request body domain.CreateSupplierContactRequest true "Contact data"
// @Success 201 {object} domain.SupplierContactDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/contacts [post]
func (h *SupplierHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	var req domain.CreateSupplierContactRequest
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

	contact, err := h.supplierService.CreateContact(r.Context(), supplierID, &req)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
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
		h.logger.Error("failed to create supplier contact", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create supplier contact",
		})
		return
	}

	w.Header().Set("Location", "/api/v1/suppliers/"+supplierID.String()+"/contacts/"+contact.ID.String())
	respondJSON(w, http.StatusCreated, contact)
}

// UpdateContact godoc
// @Summary Update a supplier contact
// @Description Update a contact person for a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param contactId path string true "Contact ID" format(uuid)
// @Param request body domain.UpdateSupplierContactRequest true "Contact data"
// @Success 200 {object} domain.SupplierContactDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/contacts/{contactId} [put]
func (h *SupplierHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	contactID, err := uuid.Parse(chi.URLParam(r, "contactId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid contact ID format",
		})
		return
	}

	var req domain.UpdateSupplierContactRequest
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

	contact, err := h.supplierService.UpdateContact(r.Context(), supplierID, contactID, &req)
	if err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		if errors.Is(err, service.ErrSupplierContactNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Contact not found",
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
		h.logger.Error("failed to update supplier contact", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update supplier contact",
		})
		return
	}

	respondJSON(w, http.StatusOK, contact)
}

// DeleteContact godoc
// @Summary Delete a supplier contact
// @Description Delete a contact person for a supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID" format(uuid)
// @Param contactId path string true "Contact ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Contact is assigned to active offers"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /suppliers/{id}/contacts/{contactId} [delete]
func (h *SupplierHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	supplierID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid supplier ID format",
		})
		return
	}

	contactID, err := uuid.Parse(chi.URLParam(r, "contactId"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid contact ID format",
		})
		return
	}

	if err := h.supplierService.DeleteContact(r.Context(), supplierID, contactID); err != nil {
		if errors.Is(err, service.ErrSupplierNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Supplier not found",
			})
			return
		}
		if errors.Is(err, service.ErrSupplierContactNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Contact not found",
			})
			return
		}
		if errors.Is(err, service.ErrContactUsedInActiveOffers) {
			respondJSON(w, http.StatusConflict, domain.ErrorResponse{
				Error:   "Conflict",
				Message: "Contact is assigned to active offers and cannot be deleted",
			})
			return
		}
		h.logger.Error("failed to delete supplier contact", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete supplier contact",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
