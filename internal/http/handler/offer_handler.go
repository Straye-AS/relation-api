package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/datawarehouse"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type OfferHandler struct {
	offerService *service.OfferService
	dwClient     *datawarehouse.Client
	logger       *zap.Logger
}

func NewOfferHandler(offerService *service.OfferService, dwClient *datawarehouse.Client, logger *zap.Logger) *OfferHandler {
	return &OfferHandler{
		offerService: offerService,
		dwClient:     dwClient,
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
// @Param sortBy query string false "Sort field" Enums(createdAt, updatedAt, title, value, probability, phase, status, dueDate, customerName)
// @Param sortOrder query string false "Sort order" Enums(asc, desc) default(desc)
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

	// Parse sort configuration
	sort := repository.DefaultSortConfig()
	if sortBy := r.URL.Query().Get("sortBy"); sortBy != "" {
		sort.Field = sortBy
	}
	if sortOrder := r.URL.Query().Get("sortOrder"); sortOrder != "" {
		sort.Order = repository.ParseSortOrder(sortOrder)
	}

	result, err := h.offerService.ListWithSort(r.Context(), page, pageSize, customerID, projectID, phase, sort)
	if err != nil {
		h.logger.Error("failed to list offers", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to list offers")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// @Summary Create offer
// @Description Creates a new offer. Defaults to in_progress phase (use /inquiries for draft phase).
// @Description Supports three scenarios for customer/project association:
// @Description - customerId only: Creates offer with that customer, auto-creates project
// @Description - projectId only: Inherits customer from the existing project
// @Description - Both: Uses provided customer, links to specified project
// @Tags Offers
// @Accept json
// @Produce json
// @Param request body domain.CreateOfferRequest true "Offer data"
// @Success 201 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Validation error (missing customerId/projectId, project has no customer, etc.)"
// @Failure 404 {object} domain.ErrorResponse "Customer or project not found"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers [post]
func (h *OfferHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.Create(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create offer", zap.Error(err))
		h.handleOfferError(w, err)
		return
	}

	w.Header().Set("Location", "/api/v1/offers/"+offer.ID.String())
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	offer, err := h.offerService.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.AdvanceOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.Advance(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to advance offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	items, err := h.offerService.GetItems(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer items", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.CreateOfferItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	item, err := h.offerService.AddItem(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to add offer item", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	files, err := h.offerService.GetFiles(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer files", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
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
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	activities, err := h.offerService.GetActivities(r.Context(), id, limit)
	if err != nil {
		h.logger.Error("failed to get offer activities", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
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

// Win godoc
// @Summary Win an offer within a project
// @Description Transitions an offer to won phase within a project context. This also transitions the project from tilbud to active phase, expires sibling offers, and sets the winning offer's number on the project.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.WinOfferRequest true "Win options"
// @Success 200 {object} domain.WinOfferResponse "Won offer with project and expired sibling offers"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, offer not in project, or project not in tilbud phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/win [post]
func (h *OfferHandler) Win(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.WinOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is allowed
		req = domain.WinOfferRequest{}
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	response, err := h.offerService.WinOffer(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to win offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
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

	// Default IncludeBudget to true when cloning.
	// Using *bool allows distinguishing between "not specified" (nil -> default true)
	// and "explicitly set to false" (user wants to exclude budget items).
	if req.IncludeBudget == nil {
		includeBudget := true
		req.IncludeBudget = &includeBudget
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

// GetWithBudgetItems godoc
// @Summary Get offer with budget details
// @Description Get an offer including budget items and summary calculations
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
func (h *OfferHandler) GetWithBudgetItems(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	offer, err := h.offerService.GetByIDWithBudgetItems(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get offer with budget items", zap.Error(err), zap.String("offer_id", id.String()))
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
	case errors.Is(err, service.ErrOfferAlreadyWon):
		respondWithError(w, http.StatusBadRequest, "Offer is already won")
	case errors.Is(err, service.ErrOfferNotInProject):
		respondWithError(w, http.StatusBadRequest, "Offer must be linked to a project to be won through this endpoint")
	case errors.Is(err, service.ErrProjectNotInTilbudPhase):
		respondWithError(w, http.StatusBadRequest, "Project must be in tilbud phase to win an offer")
	case errors.Is(err, service.ErrOfferCannotClone):
		respondWithError(w, http.StatusBadRequest, "Cannot clone this offer")
	case errors.Is(err, service.ErrOfferMissingResponsible):
		respondWithError(w, http.StatusBadRequest, "Offer must have a responsible user or company with default responsible user to advance to in_progress")
	case errors.Is(err, service.ErrInvalidCompanyID):
		respondWithError(w, http.StatusBadRequest, "Invalid company ID")
	case errors.Is(err, service.ErrOfferNumberGenerationFailed):
		respondWithError(w, http.StatusInternalServerError, "Failed to generate offer number")
	case errors.Is(err, service.ErrProjectCreationFailed):
		respondWithError(w, http.StatusInternalServerError, "Failed to create project from offer")
	case errors.Is(err, service.ErrCustomerNotFound):
		respondWithError(w, http.StatusBadRequest, "Customer not found")
	case errors.Is(err, service.ErrProjectNotFound):
		respondWithError(w, http.StatusBadRequest, "Project not found")
	case errors.Is(err, service.ErrProjectNotInOfferPhase):
		respondWithError(w, http.StatusBadRequest, "Project must be in tilbud (offer) phase to link offers")
	case errors.Is(err, service.ErrMissingCustomerOrProject):
		respondWithError(w, http.StatusBadRequest, "Either customerId or projectId must be provided")
	case errors.Is(err, service.ErrProjectHasNoCustomer):
		respondWithError(w, http.StatusBadRequest, "Project has no customer to inherit from")
	case errors.Is(err, service.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	case errors.Is(err, service.ErrOfferNumberConflict):
		respondWithError(w, http.StatusConflict, "Offer number already exists")
	case errors.Is(err, service.ErrExternalReferenceConflict):
		respondWithError(w, http.StatusConflict, "External reference already exists for this company")
	// Order phase errors
	case errors.Is(err, service.ErrOfferNotInSentPhase):
		respondWithError(w, http.StatusBadRequest, "Offer must be in sent phase")
	case errors.Is(err, service.ErrOfferNotInOrderPhase):
		respondWithError(w, http.StatusBadRequest, "Offer must be in order phase")
	case errors.Is(err, service.ErrOfferAlreadyInOrder):
		respondWithError(w, http.StatusBadRequest, "Offer is already in order phase")
	case errors.Is(err, service.ErrOfferAlreadyCompleted):
		respondWithError(w, http.StatusBadRequest, "Offer is already completed")
	default:
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// ============================================================================
// Individual Property Update Endpoints
// ============================================================================

// UpdateProbability godoc
// @Summary Update offer probability
// @Description Updates only the probability field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferProbabilityRequest true "Probability data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/probability [put]
func (h *OfferHandler) UpdateProbability(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferProbabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateProbability(r.Context(), id, req.Probability)
	if err != nil {
		h.logger.Error("failed to update offer probability", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateTitle godoc
// @Summary Update offer title
// @Description Updates only the title field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferTitleRequest true "Title data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/title [put]
func (h *OfferHandler) UpdateTitle(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferTitleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateTitle(r.Context(), id, req.Title)
	if err != nil {
		h.logger.Error("failed to update offer title", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateResponsible godoc
// @Summary Update offer responsible user
// @Description Updates only the responsible user field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferResponsibleRequest true "Responsible user data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/responsible [put]
func (h *OfferHandler) UpdateResponsible(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferResponsibleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateResponsible(r.Context(), id, req.ResponsibleUserID)
	if err != nil {
		h.logger.Error("failed to update offer responsible", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateCustomer godoc
// @Summary Update offer customer
// @Description Updates only the customer field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferCustomerRequest true "Customer data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID, offer closed, or customer not found"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/customer [put]
func (h *OfferHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateCustomer(r.Context(), id, req.CustomerID)
	if err != nil {
		h.logger.Error("failed to update offer customer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateValue godoc
// @Summary Update offer value
// @Description Updates only the value field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferValueRequest true "Value data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/value [put]
func (h *OfferHandler) UpdateValue(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateValue(r.Context(), id, req.Value)
	if err != nil {
		h.logger.Error("failed to update offer value", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateCost godoc
// @Summary Update offer cost
// @Description Updates only the cost field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferCostRequest true "Cost data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/cost [put]
func (h *OfferHandler) UpdateCost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferCostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateCost(r.Context(), id, req.Cost)
	if err != nil {
		h.logger.Error("failed to update offer cost", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateDueDate godoc
// @Summary Update offer due date
// @Description Updates only the due date field of an offer (can be null to clear)
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferDueDateRequest true "Due date data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/due-date [put]
func (h *OfferHandler) UpdateDueDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferDueDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	offer, err := h.offerService.UpdateDueDate(r.Context(), id, req.DueDate)
	if err != nil {
		h.logger.Error("failed to update offer due date", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateExpirationDate godoc
// @Summary Update offer expiration date
// @Description Updates the expiration date of a sent offer. Used to extend the validity period for the customer to accept.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferExpirationDateRequest true "Expiration date data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID, offer not sent, or invalid date"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/expiration-date [put]
func (h *OfferHandler) UpdateExpirationDate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferExpirationDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	offer, err := h.offerService.UpdateExpirationDate(r.Context(), id, req.ExpirationDate)
	if err != nil {
		h.logger.Error("failed to update offer expiration date", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateDescription godoc
// @Summary Update offer description
// @Description Updates only the description field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferDescriptionRequest true "Description data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/description [put]
func (h *OfferHandler) UpdateDescription(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateDescription(r.Context(), id, req.Description)
	if err != nil {
		h.logger.Error("failed to update offer description", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateNotes godoc
// @Summary Update offer notes
// @Description Updates only the notes field of an offer
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferNotesRequest true "Notes data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/notes [put]
func (h *OfferHandler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateNotes(r.Context(), id, req.Notes)
	if err != nil {
		h.logger.Error("failed to update offer notes", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// LinkToProject godoc
// @Summary Link offer to project
// @Description Links an offer to an existing project
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferProjectRequest true "Project data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID, offer closed, or project not found"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/project [put]
func (h *OfferHandler) LinkToProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.LinkToProject(r.Context(), id, req.ProjectID)
	if err != nil {
		h.logger.Error("failed to link offer to project", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UnlinkFromProject godoc
// @Summary Unlink offer from project
// @Description Removes the project link from an offer
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/project [delete]
func (h *OfferHandler) UnlinkFromProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	offer, err := h.offerService.UnlinkFromProject(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to unlink offer from project", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateCustomerHasWonProject godoc
// @Summary Update customer has won project flag
// @Description Updates the flag indicating whether the customer has won their project
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferCustomerHasWonProjectRequest true "Customer has won project data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or offer closed"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/customer-has-won-project [put]
func (h *OfferHandler) UpdateCustomerHasWonProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferCustomerHasWonProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	offer, err := h.offerService.UpdateCustomerHasWonProject(r.Context(), id, req.CustomerHasWonProject)
	if err != nil {
		h.logger.Error("failed to update customer has won project", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateOfferNumber godoc
// @Summary Update offer number
// @Description Updates the internal offer number (e.g., "TK-2025-001"). Returns 409 if number already exists.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferNumberRequest true "Offer number data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or request"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 409 {object} domain.ErrorResponse "Offer number already exists"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/offer-number [put]
func (h *OfferHandler) UpdateOfferNumber(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferNumberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateOfferNumber(r.Context(), id, req.OfferNumber)
	if err != nil {
		h.logger.Error("failed to update offer number", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateExternalReference godoc
// @Summary Update external reference
// @Description Updates the external/customer reference number. Returns 409 if reference already exists for this company.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferExternalReferenceRequest true "External reference data"
// @Success 200 {object} domain.OfferDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid ID or request"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 409 {object} domain.ErrorResponse "External reference already exists for this company"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/external-reference [put]
func (h *OfferHandler) UpdateExternalReference(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferExternalReferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateExternalReference(r.Context(), id, req.ExternalReference)
	if err != nil {
		h.logger.Error("failed to update external reference", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// ============================================================================
// Order Phase Lifecycle Endpoints
// ============================================================================

// AcceptOrder godoc
// @Summary Accept order
// @Description Transitions a won offer to order phase, indicating work is beginning. This is used when a customer accepts a sent offer and work should start.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.AcceptOrderRequest true "Accept order options"
// @Success 200 {object} domain.AcceptOrderResponse "Offer transitioned to order phase"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, or offer not in valid phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/accept-order [post]
func (h *OfferHandler) AcceptOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.AcceptOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is allowed
		req = domain.AcceptOrderRequest{}
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	response, err := h.offerService.AcceptOrder(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to accept order", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}

// UpdateOfferHealth godoc
// @Summary Update offer health status
// @Description Updates the health status (on_track, at_risk, delayed, over_budget) and optionally the completion percentage of an offer in order phase. Used to track execution progress.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferHealthRequest true "Health data: health (enum: on_track|at_risk|delayed|over_budget), completionPercent (optional: 0-100)"
// @Success 200 {object} domain.OfferDTO "Updated offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, or offer not in order phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/health [put]
func (h *OfferHandler) UpdateOfferHealth(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferHealthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateOfferHealth(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update offer health", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateOfferSpent godoc
// @Summary Update offer spent amount
// @Description Updates the spent amount of an offer in order phase. Used to track actual costs incurred during execution.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferSpentRequest true "Spent amount data"
// @Success 200 {object} domain.OfferDTO "Updated offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, or offer not in order phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/spent [put]
func (h *OfferHandler) UpdateOfferSpent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferSpentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateOfferSpent(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update offer spent", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// UpdateOfferInvoiced godoc
// @Summary Update offer invoiced amount
// @Description Updates the invoiced amount of an offer in order phase. Used to track how much has been invoiced to the customer.
// @Tags Offers
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.UpdateOfferInvoicedRequest true "Invoiced amount data"
// @Success 200 {object} domain.OfferDTO "Updated offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID, request body, or offer not in order phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/invoiced [put]
func (h *OfferHandler) UpdateOfferInvoiced(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	var req domain.UpdateOfferInvoicedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	offer, err := h.offerService.UpdateOfferInvoiced(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update offer invoiced", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// CompleteOffer godoc
// @Summary Complete an offer
// @Description Transitions an offer from order phase to completed phase. Indicates work is finished.
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferDTO "Completed offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID or offer not in order phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/complete [post]
func (h *OfferHandler) CompleteOffer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	offer, err := h.offerService.CompleteOffer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to complete offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// ReopenOffer godoc
// @Summary Reopen a completed offer
// @Description Transitions an offer from completed phase back to order phase. Allows additional work on a finished order.
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferDTO "Reopened offer"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID or offer not in completed phase"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/reopen [post]
func (h *OfferHandler) ReopenOffer(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	offer, err := h.offerService.ReopenOffer(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to reopen offer", zap.Error(err), zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, offer)
}

// GetNextNumber godoc
// @Summary Get next offer number preview
// @Description Returns a preview of what the next offer number would be for a company without consuming the sequence. Useful for UI display.
// @Tags Offers
// @Produce json
// @Param companyId query string true "Company ID" Enums(gruppen, stalbygg, hybridbygg, industri, tak, montasje)
// @Success 200 {object} domain.NextOfferNumberResponse
// @Failure 400 {object} domain.ErrorResponse "Invalid or missing company ID"
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/next-number [get]
func (h *OfferHandler) GetNextNumber(w http.ResponseWriter, r *http.Request) {
	companyIDStr := r.URL.Query().Get("companyId")
	if companyIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "companyId query parameter is required")
		return
	}

	if !domain.IsValidCompanyID(companyIDStr) {
		respondWithError(w, http.StatusBadRequest, "Invalid companyId: must be one of gruppen, stalbygg, hybridbygg, industri, tak, montasje")
		return
	}

	companyID := domain.CompanyID(companyIDStr)

	result, err := h.offerService.GetNextOfferNumber(r.Context(), companyID)
	if err != nil {
		h.logger.Error("failed to get next offer number",
			zap.Error(err),
			zap.String("company_id", companyIDStr))

		if errors.Is(err, service.ErrInvalidCompanyID) {
			respondWithError(w, http.StatusBadRequest, "Invalid company ID")
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to get next offer number")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ============================================================================
// Data Warehouse Sync Endpoints (POC)
// ============================================================================

// GetExternalSync godoc
// @Summary Sync data warehouse financials for an offer
// @Description Syncs financial data from the data warehouse for the given offer and persists it.
// @Description The offer must have an external_reference to be matched in the data warehouse.
// @Description This endpoint queries the DW, persists the financial data to the offer, and returns the result.
// @Tags Offers
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} domain.OfferExternalSyncResponse "Data warehouse financials with sync status"
// @Failure 400 {object} domain.ErrorResponse "Invalid offer ID"
// @Failure 404 {object} domain.ErrorResponse "Offer not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/external-sync [get]
func (h *OfferHandler) GetExternalSync(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	// Call the service to sync from data warehouse
	response, err := h.offerService.SyncFromDataWarehouse(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to sync offer from data warehouse",
			zap.Error(err),
			zap.String("offer_id", id.String()))
		h.handleOfferError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}
