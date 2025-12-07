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

// BudgetDimensionHandler handles HTTP requests for budget dimension operations
type BudgetDimensionHandler struct {
	dimensionService *service.BudgetDimensionService
	logger           *zap.Logger
}

// NewBudgetDimensionHandler creates a new BudgetDimensionHandler instance
func NewBudgetDimensionHandler(dimensionService *service.BudgetDimensionService, logger *zap.Logger) *BudgetDimensionHandler {
	return &BudgetDimensionHandler{
		dimensionService: dimensionService,
		logger:           logger,
	}
}

// ListOfferDimensions godoc
// @Summary List offer budget dimensions
// @Description Get all budget dimensions for a specific offer with pagination
// @Tags Offer Budget
// @Produce json
// @Param id path string true "Offer ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} domain.PaginatedResponse "Paginated list of budget dimensions"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/budget/dimensions [get]
func (h *BudgetDimensionHandler) ListOfferDimensions(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
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
	if pageSize > 200 {
		pageSize = 200
	}

	result, err := h.dimensionService.ListByParentPaginated(r.Context(), domain.BudgetParentOffer, offerID, page, pageSize)
	if err != nil {
		h.logger.Error("failed to list offer budget dimensions",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// AddToOffer godoc
// @Summary Add budget dimension to offer
// @Description Add a new budget dimension to an offer. Either categoryId or customName must be provided.
// @Tags Offer Budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.AddOfferBudgetDimensionRequest true "Dimension data"
// @Success 201 {object} domain.BudgetDimensionDTO "Created budget dimension"
// @Failure 400 {object} domain.ErrorResponse "Invalid request"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/budget/dimensions [post]
func (h *BudgetDimensionHandler) AddToOffer(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.AddOfferBudgetDimensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Convert to CreateBudgetDimensionRequest with parent info
	createReq := &domain.CreateBudgetDimensionRequest{
		ParentType:          domain.BudgetParentOffer,
		ParentID:            offerID,
		CategoryID:          req.CategoryID,
		CustomName:          req.CustomName,
		Cost:                req.Cost,
		Revenue:             req.Revenue,
		TargetMarginPercent: req.TargetMarginPercent,
		MarginOverride:      req.MarginOverride,
		Description:         req.Description,
		Quantity:            req.Quantity,
		Unit:                req.Unit,
		DisplayOrder:        req.DisplayOrder,
	}

	dimension, err := h.dimensionService.Create(r.Context(), createReq)
	if err != nil {
		h.logger.Error("failed to add budget dimension to offer",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	w.Header().Set("Location", "/offers/"+offerID.String()+"/budget/dimensions/"+dimension.ID.String())
	respondJSON(w, http.StatusCreated, dimension)
}

// UpdateOfferDimension godoc
// @Summary Update offer budget dimension
// @Description Update an existing budget dimension for an offer
// @Tags Offer Budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param dimensionId path string true "Dimension ID"
// @Param request body domain.UpdateBudgetDimensionRequest true "Dimension data"
// @Success 200 {object} domain.BudgetDimensionDTO "Updated budget dimension"
// @Failure 400 {object} domain.ErrorResponse "Invalid request"
// @Failure 404 {object} domain.ErrorResponse "Dimension not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/budget/dimensions/{dimensionId} [put]
func (h *BudgetDimensionHandler) UpdateOfferDimension(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	dimensionID, err := uuid.Parse(chi.URLParam(r, "dimensionId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid dimension ID")
		return
	}

	var req domain.UpdateBudgetDimensionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Verify the dimension belongs to this offer
	existingDimension, err := h.dimensionService.GetByID(r.Context(), dimensionID)
	if err != nil {
		h.handleBudgetDimensionError(w, err)
		return
	}

	if existingDimension.ParentType != domain.BudgetParentOffer || existingDimension.ParentID != offerID {
		respondWithError(w, http.StatusNotFound, "Budget dimension not found for this offer")
		return
	}

	dimension, err := h.dimensionService.Update(r.Context(), dimensionID, &req)
	if err != nil {
		h.logger.Error("failed to update offer budget dimension",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
			zap.String("dimension_id", dimensionID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dimension)
}

// DeleteOfferDimension godoc
// @Summary Delete offer budget dimension
// @Description Delete a budget dimension from an offer
// @Tags Offer Budget
// @Param id path string true "Offer ID"
// @Param dimensionId path string true "Dimension ID"
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse "Invalid ID"
// @Failure 404 {object} domain.ErrorResponse "Dimension not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/budget/dimensions/{dimensionId} [delete]
func (h *BudgetDimensionHandler) DeleteOfferDimension(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	dimensionID, err := uuid.Parse(chi.URLParam(r, "dimensionId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid dimension ID")
		return
	}

	// Verify the dimension belongs to this offer
	existingDimension, err := h.dimensionService.GetByID(r.Context(), dimensionID)
	if err != nil {
		h.handleBudgetDimensionError(w, err)
		return
	}

	if existingDimension.ParentType != domain.BudgetParentOffer || existingDimension.ParentID != offerID {
		respondWithError(w, http.StatusNotFound, "Budget dimension not found for this offer")
		return
	}

	if err := h.dimensionService.Delete(r.Context(), dimensionID); err != nil {
		h.logger.Error("failed to delete offer budget dimension",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
			zap.String("dimension_id", dimensionID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReorderOfferDimensions godoc
// @Summary Reorder offer budget dimensions
// @Description Reorder the budget dimensions for an offer. All dimension IDs must be provided.
// @Tags Offer Budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.ReorderDimensionsRequest true "Ordered dimension IDs"
// @Success 200 {array} domain.BudgetDimensionDTO "Reordered budget dimensions"
// @Failure 400 {object} domain.ErrorResponse "Invalid request or count mismatch"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/budget/reorder [put]
func (h *BudgetDimensionHandler) ReorderOfferDimensions(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.ReorderDimensionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	if err := h.dimensionService.ReorderDimensions(r.Context(), domain.BudgetParentOffer, offerID, req.OrderedIDs); err != nil {
		h.logger.Error("failed to reorder offer budget dimensions",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	// Return the reordered dimensions
	dimensions, err := h.dimensionService.ListByParent(r.Context(), domain.BudgetParentOffer, offerID)
	if err != nil {
		h.logger.Error("failed to fetch reordered dimensions",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, dimensions)
}

// GetOfferBudgetWithDimensions godoc
// @Summary Get offer budget breakdown
// @Description Get the complete budget breakdown for an offer including all dimensions and summary
// @Tags Offer Budget
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} OfferBudgetResponse "Budget breakdown with dimensions and summary"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/budget [get]
func (h *BudgetDimensionHandler) GetOfferBudgetWithDimensions(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	// Get dimensions
	dimensions, err := h.dimensionService.ListByParent(r.Context(), domain.BudgetParentOffer, offerID)
	if err != nil {
		h.logger.Error("failed to get offer budget dimensions",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	// Get summary
	summary, err := h.dimensionService.GetSummary(r.Context(), domain.BudgetParentOffer, offerID)
	if err != nil {
		h.logger.Error("failed to get offer budget summary",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
		)
		h.handleBudgetDimensionError(w, err)
		return
	}

	response := OfferBudgetResponse{
		Dimensions: dimensions,
		Summary:    summary,
	}

	respondJSON(w, http.StatusOK, response)
}

// OfferBudgetResponse represents the complete budget breakdown for an offer
type OfferBudgetResponse struct {
	Dimensions []domain.BudgetDimensionDTO `json:"dimensions"`
	Summary    *domain.BudgetSummaryDTO    `json:"summary"`
}

// handleBudgetDimensionError maps service errors to HTTP status codes
func (h *BudgetDimensionHandler) handleBudgetDimensionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBudgetDimensionNotFound):
		respondWithError(w, http.StatusNotFound, "Budget dimension not found")
	case errors.Is(err, service.ErrInvalidParentType):
		respondWithError(w, http.StatusBadRequest, "Invalid parent type")
	case errors.Is(err, service.ErrParentNotFound):
		respondWithError(w, http.StatusNotFound, "Offer not found")
	case errors.Is(err, service.ErrInvalidCost):
		respondWithError(w, http.StatusBadRequest, "Cost must be greater than 0")
	case errors.Is(err, service.ErrInvalidRevenue):
		respondWithError(w, http.StatusBadRequest, "Revenue must be greater than or equal to 0")
	case errors.Is(err, service.ErrInvalidTargetMargin):
		respondWithError(w, http.StatusBadRequest, "Target margin must be between 0 and 100%")
	case errors.Is(err, service.ErrInvalidCategory):
		respondWithError(w, http.StatusBadRequest, "Category not found or inactive")
	case errors.Is(err, service.ErrMissingName):
		respondWithError(w, http.StatusBadRequest, "Either categoryId or customName must be provided")
	case errors.Is(err, service.ErrReorderCountMismatch):
		respondWithError(w, http.StatusBadRequest, "Number of dimension IDs does not match existing dimensions")
	default:
		h.logger.Error("budget dimension error", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
