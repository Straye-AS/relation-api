package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// InquiryHandler handles HTTP requests for inquiries (draft offers)
type InquiryHandler struct {
	inquiryService *service.InquiryService
	logger         *zap.Logger
}

// NewInquiryHandler creates a new InquiryHandler
func NewInquiryHandler(inquiryService *service.InquiryService, logger *zap.Logger) *InquiryHandler {
	return &InquiryHandler{
		inquiryService: inquiryService,
		logger:         logger,
	}
}

// List godoc
// @Summary List inquiries
// @Description Returns a paginated list of inquiries (offers in draft phase)
// @Tags Inquiries
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param customerId query string false "Filter by customer ID"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /inquiries [get]
func (h *InquiryHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	var customerID *uuid.UUID
	if cid := r.URL.Query().Get("customerId"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			customerID = &id
		}
	}

	result, err := h.inquiryService.List(r.Context(), page, pageSize, customerID)
	if err != nil {
		h.logger.Error("failed to list inquiries", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to list inquiries")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// Create godoc
// @Summary Create inquiry
// @Description Creates a new inquiry (offer in draft phase with minimal required fields)
// @Tags Inquiries
// @Accept json
// @Produce json
// @Param request body domain.CreateInquiryRequest true "Inquiry data"
// @Success 201 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse "Customer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /inquiries [post]
func (h *InquiryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	inquiry, err := h.inquiryService.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create inquiry", zap.Error(err))
		h.handleInquiryError(w, err)
		return
	}

	w.Header().Set("Location", "/api/v1/inquiries/"+inquiry.ID.String())
	respondJSON(w, http.StatusCreated, inquiry)
}

// GetByID godoc
// @Summary Get inquiry
// @Description Returns a specific inquiry by ID
// @Tags Inquiries
// @Produce json
// @Param id path string true "Inquiry ID"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID"
// @Failure 404 {object} domain.ErrorResponse "Inquiry not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /inquiries/{id} [get]
func (h *InquiryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid inquiry ID: must be a valid UUID")
		return
	}

	inquiry, err := h.inquiryService.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get inquiry", zap.Error(err), zap.String("inquiry_id", id.String()))
		h.handleInquiryError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, inquiry)
}

// Delete godoc
// @Summary Delete inquiry
// @Description Deletes an inquiry by ID
// @Tags Inquiries
// @Param id path string true "Inquiry ID"
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or not an inquiry"
// @Failure 404 {object} domain.ErrorResponse "Inquiry not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /inquiries/{id} [delete]
func (h *InquiryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid inquiry ID: must be a valid UUID")
		return
	}

	err = h.inquiryService.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete inquiry", zap.Error(err), zap.String("inquiry_id", id.String()))
		h.handleInquiryError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Convert godoc
// @Summary Convert inquiry to offer
// @Description Converts an inquiry to an offer (phase=in_progress), generating an offer number
// @Tags Inquiries
// @Accept json
// @Produce json
// @Param id path string true "Inquiry ID"
// @Param request body domain.ConvertInquiryRequest true "Conversion options"
// @Success 200 {object} domain.ConvertInquiryResponse
// @Failure 400 {object} domain.ErrorResponse "Invalid ID, not an inquiry, or missing conversion data"
// @Failure 404 {object} domain.ErrorResponse "Inquiry not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /inquiries/{id}/convert [post]
func (h *InquiryHandler) Convert(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid inquiry ID: must be a valid UUID")
		return
	}

	var req domain.ConvertInquiryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is valid - will use defaults
		if err.Error() != "EOF" {
			respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
			return
		}
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	result, err := h.inquiryService.Convert(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to convert inquiry", zap.Error(err), zap.String("inquiry_id", id.String()))
		h.handleInquiryError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// handleInquiryError maps service errors to HTTP status codes
func (h *InquiryHandler) handleInquiryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInquiryNotFound):
		respondWithError(w, http.StatusNotFound, "Inquiry not found")
	case errors.Is(err, service.ErrNotAnInquiry):
		respondWithError(w, http.StatusBadRequest, "Offer is not an inquiry (must be in draft phase)")
	case errors.Is(err, service.ErrInquiryMissingConversionData):
		respondWithError(w, http.StatusBadRequest, "Conversion requires responsibleUserId or companyId with default responsible user")
	case errors.Is(err, service.ErrCustomerNotFound):
		respondWithError(w, http.StatusNotFound, "Customer not found")
	case errors.Is(err, service.ErrInvalidCompanyID):
		respondWithError(w, http.StatusBadRequest, "Invalid company ID. Valid values: gruppen, stalbygg, hybridbygg, industri, tak, montasje")
	case errors.Is(err, service.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	default:
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
