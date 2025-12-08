package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type DealHandler struct {
	dealService *service.DealService
	logger      *zap.Logger
}

func NewDealHandler(dealService *service.DealService, logger *zap.Logger) *DealHandler {
	return &DealHandler{
		dealService: dealService,
		logger:      logger,
	}
}

// @Summary List deals
// @Description List deals with optional filters
// @Tags Deals
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param stage query string false "Filter by stage (lead, qualified, proposal, negotiation, won, lost)"
// @Param ownerId query string false "Filter by owner ID"
// @Param customerId query string false "Filter by customer ID"
// @Param companyId query string false "Filter by company ID"
// @Param source query string false "Filter by source"
// @Param minValue query number false "Minimum value"
// @Param maxValue query number false "Maximum value"
// @Param createdAfter query string false "Created after date (YYYY-MM-DD)"
// @Param createdBefore query string false "Created before date (YYYY-MM-DD)"
// @Param closeAfter query string false "Expected close after date (YYYY-MM-DD)"
// @Param closeBefore query string false "Expected close before date (YYYY-MM-DD)"
// @Param sort query string false "Sort by (created_desc, created_asc, value_desc, value_asc, probability_desc, probability_asc, close_date_desc, close_date_asc, weighted_desc, weighted_asc)"
// @Success 200 {object} domain.PaginatedResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals [get]
func (h *DealHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}

	filters := &repository.DealFilters{}

	// Stage filter
	if s := r.URL.Query().Get("stage"); s != "" {
		stage := domain.DealStage(s)
		filters.Stage = &stage
	}

	// Owner filter
	if o := r.URL.Query().Get("ownerId"); o != "" {
		filters.OwnerID = &o
	}

	// Customer filter
	if cid := r.URL.Query().Get("customerId"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			filters.CustomerID = &id
		}
	}

	// Company filter
	if compID := r.URL.Query().Get("companyId"); compID != "" {
		companyID := domain.CompanyID(compID)
		filters.CompanyID = &companyID
	}

	// Source filter
	if src := r.URL.Query().Get("source"); src != "" {
		filters.Source = &src
	}

	// Value range filters
	if minVal := r.URL.Query().Get("minValue"); minVal != "" {
		if v, err := strconv.ParseFloat(minVal, 64); err == nil {
			filters.MinValue = &v
		}
	}
	if maxVal := r.URL.Query().Get("maxValue"); maxVal != "" {
		if v, err := strconv.ParseFloat(maxVal, 64); err == nil {
			filters.MaxValue = &v
		}
	}

	// Date range filters
	if ca := r.URL.Query().Get("createdAfter"); ca != "" {
		if t, err := time.Parse("2006-01-02", ca); err == nil {
			filters.CreatedAfter = &t
		}
	}
	if cb := r.URL.Query().Get("createdBefore"); cb != "" {
		if t, err := time.Parse("2006-01-02", cb); err == nil {
			filters.CreatedBefore = &t
		}
	}
	if cla := r.URL.Query().Get("closeAfter"); cla != "" {
		if t, err := time.Parse("2006-01-02", cla); err == nil {
			filters.CloseAfter = &t
		}
	}
	if clb := r.URL.Query().Get("closeBefore"); clb != "" {
		if t, err := time.Parse("2006-01-02", clb); err == nil {
			filters.CloseBefore = &t
		}
	}

	// Search query
	if q := r.URL.Query().Get("q"); q != "" {
		filters.SearchQuery = &q
	}

	// Sort option
	sortBy := repository.DealSortByCreatedDesc
	if s := r.URL.Query().Get("sort"); s != "" {
		sortBy = repository.DealSortOption(s)
	}

	result, err := h.dealService.List(r.Context(), page, pageSize, filters, sortBy)
	if err != nil {
		h.logger.Error("failed to list deals", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to list deals")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// @Summary Create deal
// @Description Create a new deal in the sales pipeline
// @Tags Deals
// @Accept json
// @Produce json
// @Param request body domain.CreateDealRequest true "Deal data"
// @Success 201 {object} domain.DealDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals [post]
func (h *DealHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	deal, err := h.dealService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			respondWithError(w, http.StatusBadRequest, "Customer not found")
			return
		}
		h.logger.Error("failed to create deal", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to create deal")
		return
	}

	w.Header().Set("Location", "/api/v1/deals/"+deal.ID.String())
	respondJSON(w, http.StatusCreated, deal)
}

// @Summary Get deal
// @Description Get a deal by ID with full details including stage history
// @Tags Deals
// @Produce json
// @Param id path string true "Deal ID"
// @Success 200 {object} DealWithHistoryResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id} [get]
func (h *DealHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	deal, err := h.dealService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to get deal", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusInternalServerError, "Failed to get deal")
		return
	}

	// Get stage history
	history, _ := h.dealService.GetStageHistory(r.Context(), id)

	response := DealWithHistoryResponse{
		Deal:         deal,
		StageHistory: history,
	}

	respondJSON(w, http.StatusOK, response)
}

// DealWithHistoryResponse wraps deal with its stage history
type DealWithHistoryResponse struct {
	Deal         *domain.DealDTO              `json:"deal"`
	StageHistory []domain.DealStageHistoryDTO `json:"stageHistory"`
}

// @Summary Update deal
// @Description Update an existing deal
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param request body domain.UpdateDealRequest true "Deal data"
// @Success 200 {object} domain.DealDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id} [put]
func (h *DealHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	var req domain.UpdateDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	deal, err := h.dealService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			respondWithError(w, http.StatusForbidden, "Insufficient permissions to update this deal")
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to update deal", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusInternalServerError, "Failed to update deal")
		return
	}

	respondJSON(w, http.StatusOK, deal)
}

// @Summary Delete deal
// @Description Delete a deal
// @Tags Deals
// @Param id path string true "Deal ID"
// @Success 204 "No Content"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id} [delete]
func (h *DealHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	if err := h.dealService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			respondWithError(w, http.StatusForbidden, "Insufficient permissions to delete this deal")
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to delete deal", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusInternalServerError, "Failed to delete deal")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Advance deal stage
// @Description Advance a deal to the next stage in the pipeline
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param request body domain.UpdateDealStageRequest true "Stage data"
// @Success 200 {object} domain.DealDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id}/advance [post]
func (h *DealHandler) AdvanceStage(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	var req domain.UpdateDealStageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	deal, err := h.dealService.AdvanceStage(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to advance deal stage", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusBadRequest, "Failed to advance deal stage: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, deal)
}

// @Summary Win deal
// @Description Mark a deal as won and optionally create a project
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param createProject query bool false "Create a project from the deal" default(true)
// @Success 200 {object} WinDealResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id}/win [post]
func (h *DealHandler) WinDeal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	createProject := true
	if cp := r.URL.Query().Get("createProject"); cp != "" {
		createProject, _ = strconv.ParseBool(cp)
	}

	deal, project, err := h.dealService.WinDeal(r.Context(), id, createProject)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to win deal", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusBadRequest, "Failed to win deal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, WinDealResponse{
		Deal:    deal,
		Project: project,
	})
}

// WinDealResponse wraps the response for winning a deal
type WinDealResponse struct {
	Deal    *domain.DealDTO    `json:"deal"`
	Project *domain.ProjectDTO `json:"project,omitempty"`
}

// @Summary Lose deal
// @Description Mark a deal as lost with a categorized reason and detailed notes
// @Tags Deals
// @Accept json
// @Produce json
// @Param id path string true "Deal ID"
// @Param request body domain.LoseDealRequest true "Loss reason category and notes"
// @Success 200 {object} domain.DealDTO
// @Failure 400 {object} domain.ErrorResponse "Invalid request - reason must be one of: price, timing, competitor, requirements, other. Notes must be 10-500 characters."
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id}/lose [post]
func (h *DealHandler) LoseDeal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	var req domain.LoseDealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body: malformed JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	deal, err := h.dealService.LoseDeal(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to lose deal", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusBadRequest, "Failed to mark deal as lost: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, deal)
}

// @Summary Reopen deal
// @Description Reopen a lost deal as a new lead
// @Tags Deals
// @Produce json
// @Param id path string true "Deal ID"
// @Success 200 {object} domain.DealDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id}/reopen [post]
func (h *DealHandler) ReopenDeal(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	deal, err := h.dealService.ReopenDeal(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to reopen deal", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusBadRequest, "Failed to reopen deal: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, deal)
}

// @Summary Get deal stage history
// @Description Get the stage history for a deal
// @Tags Deals
// @Produce json
// @Param id path string true "Deal ID"
// @Success 200 {array} domain.DealStageHistoryDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id}/history [get]
func (h *DealHandler) GetStageHistory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	history, err := h.dealService.GetStageHistory(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to get stage history", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusInternalServerError, "Failed to get stage history")
		return
	}

	respondJSON(w, http.StatusOK, history)
}

// @Summary Get deal activities
// @Description Get activities for a deal
// @Tags Deals
// @Produce json
// @Param id path string true "Deal ID"
// @Param limit query int false "Limit results" default(50)
// @Success 200 {array} domain.ActivityDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/{id}/activities [get]
func (h *DealHandler) GetActivities(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid deal ID: must be a valid UUID")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	activities, err := h.dealService.GetActivities(r.Context(), id, limit)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Deal not found")
			return
		}
		h.logger.Error("failed to get deal activities", zap.Error(err), zap.String("deal_id", id.String()))
		respondWithError(w, http.StatusInternalServerError, "Failed to get deal activities")
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

// @Summary Get pipeline overview
// @Description Get all deals grouped by stage for pipeline view
// @Tags Deals
// @Produce json
// @Success 200 {object} map[string][]domain.DealDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/pipeline [get]
func (h *DealHandler) GetPipelineOverview(w http.ResponseWriter, r *http.Request) {
	pipeline, err := h.dealService.GetPipelineOverview(r.Context())
	if err != nil {
		h.logger.Error("failed to get pipeline overview", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get pipeline overview")
		return
	}

	respondJSON(w, http.StatusOK, pipeline)
}

// @Summary Get pipeline statistics
// @Description Get aggregated statistics for the sales pipeline
// @Tags Deals
// @Produce json
// @Success 200 {object} repository.PipelineStats
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/stats [get]
func (h *DealHandler) GetPipelineStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.dealService.GetPipelineStats(r.Context())
	if err != nil {
		h.logger.Error("failed to get pipeline stats", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get pipeline statistics")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// @Summary Get pipeline forecast
// @Description Get forecast data for upcoming months
// @Tags Deals
// @Produce json
// @Param months query int false "Number of months to forecast" default(3)
// @Success 200 {array} repository.ForecastPeriod
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /deals/forecast [get]
func (h *DealHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	months, _ := strconv.Atoi(r.URL.Query().Get("months"))
	if months < 1 {
		months = 3
	}
	if months > 12 {
		months = 12
	}

	forecast, err := h.dealService.GetForecast(r.Context(), months)
	if err != nil {
		h.logger.Error("failed to get forecast", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to get forecast")
		return
	}

	respondJSON(w, http.StatusOK, forecast)
}
