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
// @Tags Dashboard
// @Produce json
// @Success 200 {object} domain.DashboardMetrics
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
