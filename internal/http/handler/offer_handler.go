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
		h.logger.Error("failed to update offer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
