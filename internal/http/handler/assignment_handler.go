package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// AssignmentHandler handles HTTP requests for assignments (ERP work orders)
type AssignmentHandler struct {
	assignmentService *service.AssignmentService
	logger            *zap.Logger
}

// NewAssignmentHandler creates a new assignment handler
func NewAssignmentHandler(assignmentService *service.AssignmentService, logger *zap.Logger) *AssignmentHandler {
	return &AssignmentHandler{
		assignmentService: assignmentService,
		logger:            logger,
	}
}

// SyncAssignments godoc
// @Summary Sync assignments for an offer
// @Description Fetches assignments (work orders) from the ERP datawarehouse and syncs them to the local database.
// @Description The offer must have an external_reference that matches the project Code in the ERP system.
// @Tags Assignments
// @Accept json
// @Produce json
// @Param id path string true "Offer ID" format(uuid)
// @Success 200 {object} domain.AssignmentSyncResultDTO
// @Failure 400 {object} domain.APIError "Invalid offer ID"
// @Failure 404 {object} domain.APIError "Offer not found"
// @Failure 500 {object} domain.APIError "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/assignments/sync [post]
func (h *AssignmentHandler) SyncAssignments(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	result, err := h.assignmentService.SyncAssignmentsForOffer(r.Context(), offerID)
	if err != nil {
		h.handleAssignmentError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ListAssignments godoc
// @Summary List assignments for an offer
// @Description Returns all synced assignments (work orders) for a specific offer
// @Tags Assignments
// @Accept json
// @Produce json
// @Param id path string true "Offer ID" format(uuid)
// @Success 200 {array} domain.AssignmentDTO
// @Failure 400 {object} domain.APIError "Invalid offer ID"
// @Failure 404 {object} domain.APIError "Offer not found"
// @Failure 500 {object} domain.APIError "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/assignments [get]
func (h *AssignmentHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	assignments, err := h.assignmentService.GetAssignmentsForOffer(r.Context(), offerID)
	if err != nil {
		h.handleAssignmentError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, assignments)
}

// GetAssignmentSummary godoc
// @Summary Get assignment summary for an offer
// @Description Returns aggregated assignment data with comparison to the offer value.
// @Description Includes total FixedPriceAmount, count, and difference calculations.
// @Tags Assignments
// @Accept json
// @Produce json
// @Param id path string true "Offer ID" format(uuid)
// @Success 200 {object} domain.AssignmentSummaryDTO
// @Failure 400 {object} domain.APIError "Invalid offer ID"
// @Failure 404 {object} domain.APIError "Offer not found"
// @Failure 500 {object} domain.APIError "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/{id}/assignments/summary [get]
func (h *AssignmentHandler) GetAssignmentSummary(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID: must be a valid UUID")
		return
	}

	summary, err := h.assignmentService.GetAssignmentSummaryForOffer(r.Context(), offerID)
	if err != nil {
		h.handleAssignmentError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

// handleAssignmentError maps service errors to HTTP responses
func (h *AssignmentHandler) handleAssignmentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrOfferNotFound):
		respondWithError(w, http.StatusNotFound, "Offer not found")
	case errors.Is(err, service.ErrOfferNoExternalReference):
		respondWithError(w, http.StatusBadRequest, "Offer has no external reference for assignment sync")
	case errors.Is(err, service.ErrDWClientNotAvailable):
		respondWithError(w, http.StatusServiceUnavailable, "Data warehouse not available")
	default:
		h.logger.Error("assignment handler error", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Internal server error")
	}
}

