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
