package handler

// This file contains phase transition and lifecycle handlers for the OfferHandler.
// Includes:
// - Phase transitions (Send, Accept, Reject, Win)
// - Order lifecycle (AcceptOrder, Complete, Reopen)
// - Cloning
// - Data warehouse sync

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

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

// TriggerBulkDWSync godoc
// @Summary Trigger bulk data warehouse sync (Admin only)
// @Description Manually triggers a sync of ALL offers with external_reference from the data warehouse (regardless of phase).
// @Description This is an admin-only endpoint for forcing a full sync outside of the scheduled cron job.
// @Tags Offers
// @Produce json
// @Success 200 {object} map[string]interface{} "Sync results with synced/failed counts"
// @Failure 403 {object} domain.ErrorResponse "Forbidden - requires super admin"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /offers/admin/trigger-dw-sync [post]
func (h *OfferHandler) TriggerBulkDWSync(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Only super admins can trigger bulk sync
	if !userCtx.IsSuperAdmin() {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires super admin role")
		return
	}

	h.logger.Info("admin triggered bulk DW sync",
		zap.String("user_id", userCtx.UserID.String()),
		zap.String("user_email", userCtx.Email))

	// Call the service to sync all offers
	synced, failed, err := h.offerService.SyncAllOffersFromDataWarehouse(r.Context())
	if err != nil {
		h.logger.Error("bulk DW sync failed",
			zap.Error(err),
			zap.String("triggered_by", userCtx.Email))
		respondWithError(w, http.StatusInternalServerError, "Failed to run bulk sync: "+err.Error())
		return
	}

	h.logger.Info("bulk DW sync completed",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.String("triggered_by", userCtx.Email))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Bulk data warehouse sync completed",
		"synced":  synced,
		"failed":  failed,
		"total":   synced + failed,
	})
}
