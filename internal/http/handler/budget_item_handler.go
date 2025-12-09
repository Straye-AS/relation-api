package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// BudgetItemHandler handles HTTP requests for budget items
type BudgetItemHandler struct {
	budgetItemService *service.BudgetItemService
	logger            *zap.Logger
}

// NewBudgetItemHandler creates a new BudgetItemHandler instance
func NewBudgetItemHandler(budgetItemService *service.BudgetItemService, logger *zap.Logger) *BudgetItemHandler {
	return &BudgetItemHandler{
		budgetItemService: budgetItemService,
		logger:            logger,
	}
}

// ListOfferDimensions returns all budget items for an offer
// @Summary List budget items for an offer
// @Description Get all budget items belonging to a specific offer
// @Tags offers,budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {array} domain.BudgetItemDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Router /offers/{id}/budget/dimensions [get]
// @Security BearerAuth
func (h *BudgetItemHandler) ListOfferDimensions(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	items, err := h.budgetItemService.ListByOffer(r.Context(), offerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Offer not found")
			return
		}
		h.logger.Error("Failed to list budget items", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to list budget items")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// GetOfferBudgetWithDimensions returns budget summary and items for an offer
// @Summary Get offer budget with items
// @Description Get budget summary and all budget items for an offer
// @Tags offers,budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Router /offers/{id}/budget [get]
// @Security BearerAuth
func (h *BudgetItemHandler) GetOfferBudgetWithDimensions(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	summary, items, err := h.budgetItemService.GetOfferBudgetWithDimensions(r.Context(), offerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Offer not found")
			return
		}
		h.logger.Error("Failed to get offer budget", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get offer budget")
		return
	}

	response := map[string]interface{}{
		"summary": summary,
		"items":   items,
	}

	respondJSON(w, http.StatusOK, response)
}

// AddToOffer adds a new budget item to an offer
// @Summary Add budget item to offer
// @Description Add a new budget item to a specific offer
// @Tags offers,budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.AddOfferBudgetItemRequest true "Budget item details"
// @Success 201 {object} domain.BudgetItemDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Router /offers/{id}/budget/dimensions [post]
// @Security BearerAuth
func (h *BudgetItemHandler) AddToOffer(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.AddOfferBudgetItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	item, err := h.budgetItemService.AddToOffer(r.Context(), offerID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Offer not found")
			return
		}
		h.logger.Error("Failed to add budget item", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to add budget item")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

// UpdateOfferDimension updates a budget item belonging to an offer
// @Summary Update offer budget item
// @Description Update a budget item belonging to a specific offer
// @Tags offers,budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param dimensionId path string true "Budget Item ID"
// @Param request body domain.UpdateBudgetItemRequest true "Updated budget item details"
// @Success 200 {object} domain.BudgetItemDTO
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Router /offers/{id}/budget/dimensions/{dimensionId} [put]
// @Security BearerAuth
func (h *BudgetItemHandler) UpdateOfferDimension(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "dimensionId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid budget item ID")
		return
	}

	var req domain.UpdateBudgetItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	item, err := h.budgetItemService.UpdateOfferDimension(r.Context(), offerID, itemID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Budget item not found")
			return
		}
		if strings.Contains(err.Error(), "does not belong") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to update budget item", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to update budget item")
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// DeleteOfferDimension removes a budget item from an offer
// @Summary Delete offer budget item
// @Description Delete a budget item from a specific offer
// @Tags offers,budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param dimensionId path string true "Budget Item ID"
// @Success 204 "No Content"
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Router /offers/{id}/budget/dimensions/{dimensionId} [delete]
// @Security BearerAuth
func (h *BudgetItemHandler) DeleteOfferDimension(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	itemID, err := uuid.Parse(chi.URLParam(r, "dimensionId"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid budget item ID")
		return
	}

	err = h.budgetItemService.DeleteOfferDimension(r.Context(), offerID, itemID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, "Budget item not found")
			return
		}
		if strings.Contains(err.Error(), "does not belong") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to delete budget item", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to delete budget item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReorderOfferDimensions reorders budget items for an offer
// @Summary Reorder offer budget items
// @Description Reorder budget items for a specific offer by providing ordered IDs
// @Tags offers,budget
// @Accept json
// @Produce json
// @Param id path string true "Offer ID"
// @Param request body domain.ReorderBudgetItemsRequest true "Ordered list of budget item IDs"
// @Success 204 "No Content"
// @Failure 400 {object} domain.APIError
// @Failure 404 {object} domain.APIError
// @Failure 500 {object} domain.APIError
// @Router /offers/{id}/budget/reorder [put]
// @Security BearerAuth
func (h *BudgetItemHandler) ReorderOfferDimensions(w http.ResponseWriter, r *http.Request) {
	offerID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid offer ID")
		return
	}

	var req domain.ReorderBudgetItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	err = h.budgetItemService.ReorderOfferDimensions(r.Context(), offerID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "does not belong") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.logger.Error("Failed to reorder budget items", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to reorder budget items")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
