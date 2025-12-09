package handler

import (
	"net/http"

	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
	logger           *zap.Logger
}

func NewDashboardHandler(dashboardService *service.DashboardService, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		logger:           logger,
	}
}

// @Summary Get dashboard metrics
// @Description Returns dashboard metrics using a rolling 12-month window. All metrics exclude draft and expired offers.
// @Description
// @Description **Offer Metrics:**
// @Description - `totalOfferCount`: Count of offers excluding drafts and expired
// @Description - `offerReserve`: Total value of active offers (in_progress, sent)
// @Description - `weightedOfferReserve`: Sum of (value * probability/100) for active offers
// @Description - `averageProbability`: Average probability of active offers
// @Description
// @Description **Pipeline Data:**
// @Description - Returns phases: in_progress, sent, won, lost with counts and values
// @Description - Excludes draft and expired offers
// @Description
// @Description **Win Rate Metrics:**
// @Description - `winRate`: won_count / (won_count + lost_count) - returns 0-1 scale (e.g., 0.5 = 50%)
// @Description - `economicWinRate`: won_value / (won_value + lost_value) - value-based win rate
// @Description - Also includes `wonCount`, `lostCount`, `wonValue`, `lostValue` for transparency
// @Description
// @Description **Order Reserve:** Sum of (budget - spent) on active projects
// @Description
// @Description **Financial Summary:**
// @Description - `totalInvoiced`: Sum of spent on all projects in 12-month window
// @Description - `totalValue`: orderReserve + totalInvoiced
// @Description
// @Description **Recent Lists:** 12-month window, limit 10 each
// @Description **Top Customers:** Ranked by offer count (excluding drafts/expired), includes economicValue
// @Tags Dashboard
// @Produce json
// @Success 200 {object} domain.DashboardMetrics
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /dashboard/metrics [get]
func (h *DashboardHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.dashboardService.GetMetrics(r.Context())
	if err != nil {
		h.logger.Error("failed to get dashboard metrics", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, metrics)
}

// @Summary Global search
// @Tags Search
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} domain.SearchResult
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /search [get]
func (h *DashboardHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	result, err := h.dashboardService.Search(r.Context(), query)
	if err != nil {
		h.logger.Error("failed to search", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, result)
}
