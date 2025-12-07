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

type OfferHandler struct {
	offerService *service.OfferService
	logger       *zap.Logger
}

func NewOfferHandler(offerService *service.OfferService, logger *zap.Logger) *OfferHandler {
	return &OfferHandler{
		offerService: offerService,
		logger:       logger,
	}
}

// @Summary List offers
// @Tags Offers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param customerId query string false "Filter by customer ID"
// @Param projectId query string false "Filter by project ID"
// @Param phase query string false "Filter by phase"
// @Success 200 {object} domain.PaginatedResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers [get]
func (h *OfferHandler) List(w http.ResponseWriter, r *http.Request) {
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

	var customerID, projectID *uuid.UUID
	if cid := r.URL.Query().Get("customerId"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			customerID = &id
		}
	}
	if pid := r.URL.Query().Get("projectId"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			projectID = &id
		}
	}

	var phase *domain.OfferPhase
	if p := r.URL.Query().Get("phase"); p != "" {
		ph := domain.OfferPhase(p)
		phase = &ph
	}

	result, err := h.offerService.List(r.Context(), page, pageSize, customerID, projectID, phase)
	if err != nil {
		h.logger.Error("failed to list offers", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// @Summary Create offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param request body domain.CreateOfferRequest true "Offer data"
// @Success 201 {object} domain.OfferDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers [post]
func (h *OfferHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create offer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/offers/"+offer.ID.String())
	respondJSON(w, http.StatusCreated, offer)
}

// @Summary Get offer
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferWithItemsDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id} [get]
func (h *OfferHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	offer, err := h.offerService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Offer not found", http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// @Summary Update offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferRequest true "Offer data"
// @Success 200 {object} domain.OfferDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id} [put]
func (h *OfferHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	var req domain.UpdateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.Update(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// @Summary Advance offer phase
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.AdvanceOfferRequest true "Phase data"
// @Success 200 {object} domain.OfferDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/advance [post]
func (h *OfferHandler) Advance(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	var req domain.AdvanceOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.Advance(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to advance offer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// @Summary Get offer items
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {array} domain.OfferItemDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/items [get]
func (h *OfferHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	items, err := h.offerService.GetItems(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer items", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// @Summary Add offer item
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.CreateOfferItemRequest true "Item data"
// @Success 201 {object} domain.OfferItemDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/items [post]
func (h *OfferHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	var req domain.CreateOfferItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	item, err := h.offerService.AddItem(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to add offer item", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// @Summary Get offer files
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {array} domain.FileDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/files [get]
func (h *OfferHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	files, err := h.offerService.GetFiles(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer files", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, files)
}

// @Summary Get offer activities
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Param limit query int false "Limit" default(50)
// @Success 200 {array} domain.ActivityDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/activities [get]
func (h *OfferHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}

	activities, err := h.offerService.GetActivities(r.Context(), id, limit)
	if err != nil {
		h.logger.Error("failed to get activities", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

// Delete godoc
// @Summary Delete offer
// @Description Delete an offer by ID. Only offers in draft or in_progress phase can be deleted.
// @Tags Offers
// @Param id path string true "Offer ID"
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id} [delete]
func (h *OfferHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	err = h.offerService.Delete(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to delete offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Send godoc
// @Summary Send offer to customer
// @Description Transitions an offer from draft or in_progress phase to sent phase.
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferDTO "Updated offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID or offer not in valid phase for sending"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/send [post]
func (h *OfferHandler) Send(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	offer, err := h.offerService.SendOffer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to send offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// Accept godoc
// @Summary Accept offer
// @Description Transitions an offer from sent phase to won phase. Optionally creates a project from the offer.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.AcceptOfferRequest true "Accept options"
// @Success 200 {object} domain.AcceptOfferResponse "Accepted offer and optional project"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, or offer not in sent phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/accept [post]
func (h *OfferHandler) Accept(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.AcceptOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	response, err := h.offerService.AcceptOffer(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to accept offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// Reject godoc
// @Summary Reject offer
// @Description Transitions an offer from sent phase to lost phase with an optional reason.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.RejectOfferRequest true "Rejection reason"
// @Success 200 {object} domain.OfferDTO "Rejected offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, or offer not in sent phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/reject [post]
func (h *OfferHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.RejectOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.RejectOffer(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to reject offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// Clone godoc
// @Summary Clone offer
// @Description Creates a copy of an existing offer. The cloned offer starts in draft phase.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Source Offer ID"
// @Param request body domain.CloneOfferRequest true "Clone options"
// @Success 201 {object} domain.OfferDTO "Cloned offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID or request body"
// @Failure 404 {object} domain.ErrorResponse "Source offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/clone [post]
func (h *OfferHandler) Clone(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.CloneOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Default to include dimensions when cloning (if not explicitly set in JSON)
	if req.IncludeDimensions == nil {
		includeDims := true
		req.IncludeDimensions = &includeDims
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.CloneOffer(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to clone offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	w.Header().Set("Location", "/offers/"+offer.ID.String())
	respondJSON(w, http.StatusCreated, offer)
}

// GetWithBudgetDimensions godoc
// @Summary Get offer with budget details
// @Description Get an offer including budget dimensions and summary calculations
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferDetailDTO "Offer with budget details"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/detail [get]
func (h *OfferHandler) GetWithBudgetDimensions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	offer, err := h.offerService.GetByIDWithBudgetDimensions(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer with budget dimensions", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// GetBudgetSummary retrieves budget summary for an offer
// NOTE: This endpoint is now deprecated in favor of BudgetDimensionHandler.GetOfferBudgetWithDimensions
// which provides both dimensions and summary. Kept for backwards compatibility.
func (h *OfferHandler) GetBudgetSummary(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	summary, err := h.offerService.GetBudgetSummary(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get budget summary", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

// RecalculateTotals godoc
// @Summary Recalculate offer totals
// @Description Recalculates the offer value from budget dimensions
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferDTO "Updated offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/recalculate [post]
func (h *OfferHandler) RecalculateTotals(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	offer, err := h.offerService.RecalculateTotals(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to recalculate totals", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// handleOfferError maps service errors to HTTP status codes
func (h *OfferHandler) handleOfferError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOfferNotFound):
		respondWithError(w, http.StatusNotFound, "Offer not found")
	case errors.Is(err, service.ErrOfferNotInDraftPhase):
		respondWithError(w, http.StatusBadRequest, "Offer must be in draft or in_progress phase to be sent")
	case errors.Is(err, service.ErrOfferNotSent):
		respondWithError(w, http.StatusBadRequest, "Offer must be in sent phase to accept or reject")
	case errors.Is(err, service.ErrOfferAlreadyClosed):
		respondWithError(w, http.StatusBadRequest, "Offer is already in a closed state (won/lost/expired)")
	case errors.Is(err, service.ErrOfferCannotClone):
		respondWithError(w, http.StatusBadRequest, "Cannot clone this offer")
	case errors.Is(err, service.ErrProjectCreationFailed):
		respondWithError(w, http.StatusInternalServerError, "Failed to create project from offer")
	case errors.Is(err, service.ErrCustomerNotFound):
		respondWithError(w, http.StatusBadRequest, "Customer not found")
	case errors.Is(err, service.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	default:
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}
