package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/straye-as/relation-api/internal/domain"
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
// @Description Returns dashboard metrics with configurable time range. All metrics exclude draft and expired offers.
// @Description
// @Description **IMPORTANT: Aggregation Logic (Avoids Double-Counting)**
// @Description When a project has multiple offers, only the highest value offer per phase is counted.
// @Description Orphan offers (without project) are included at full value.
// @Description Example: Project A has offers 23M and 25M in "sent" phase - totalValue shows 25M, not 48M.
// @Description
// @Description **Time Range Options:**
// @Description - `rolling12months` (default): Uses a rolling 12-month window from the current date
// @Description - `allTime`: Calculates metrics without any date filter
// @Description - Custom range: Use `fromDate` and `toDate` parameters (YYYY-MM-DD format)
// @Description
// @Description **Custom Date Range:**
// @Description When `fromDate` and/or `toDate` are provided, they override the `timeRange` parameter.
// @Description - `fromDate`: Start of range at 00:00:00 local time
// @Description - `toDate`: End of range at 23:59:59 local time
// @Description - Both parameters are optional; can be used together or individually
// @Description
// @Description **Offer Metrics (Pipeline Phase):**
// @Description - `totalOfferCount`: Count of offers excluding drafts and expired
// @Description - `totalProjectCount`: Count of unique projects with offers (excludes orphan offers)
// @Description - `offerReserve`: Total value of active offers - best per project (avoids double-counting)
// @Description - `weightedOfferReserve`: Sum of (value * probability/100) for active offers
// @Description - `averageProbability`: Average probability of active offers
// @Description
// @Description **Pipeline Data:**
// @Description - Returns phases: in_progress, sent, order, completed, lost with counts and values
// @Description - `count`: Total offer count in phase
// @Description - `projectCount`: Unique projects in phase (excludes orphan offers)
// @Description - `totalValue`: Sum of best offer value per project (avoids double-counting)
// @Description - Excludes draft and expired offers
// @Description
// @Description **Win Rate Metrics:**
// @Description - `winRate`: won_count / (won_count + lost_count) - returns 0-1 scale (e.g., 0.5 = 50%)
// @Description - `economicWinRate`: won_value / (won_value + lost_value) - value-based win rate
// @Description - Also includes `wonCount`, `lostCount`, `wonValue`, `lostValue` for transparency
// @Description
// @Description **Order Metrics (Execution Phase - from offers):**
// @Description - `activeOrderCount`: Count of offers in order phase (active execution)
// @Description - `completedOrderCount`: Count of offers in completed phase
// @Description - `orderValue`: Total value of offers in order phase
// @Description - `orderReserve`: Sum of (value - invoiced) for order phase offers
// @Description - `totalInvoiced`: Sum of invoiced for order phase offers
// @Description - `totalSpent`: Sum of spent for order phase offers
// @Description - `averageOrderProgress`: Average completion percentage for order phase offers
// @Description - `healthDistribution`: Count of order phase offers by health status (onTrack, atRisk, delayed, overBudget)
// @Description
// @Description **Financial Summary:**
// @Description - `totalValue`: orderReserve + totalInvoiced
// @Description
// @Description **Recent Lists (limit 5 each):**
// @Description - `recentOffers`: Offers in in_progress phase (Siste tilbud), sorted by update recency
// @Description - `recentOrders`: Offers in order phase (Siste ordre), sorted by update recency
// @Description
// @Description **Top Customers:** Ranked by won offer count (order + completed phases) with total won value
// @Tags Dashboard
// @Produce json
// @Param timeRange query string false "Time range for metrics (ignored if fromDate/toDate provided)" Enums(rolling12months, allTime) default(rolling12months)
// @Param fromDate query string false "Start date for custom range (YYYY-MM-DD format, 00:00:00)"
// @Param toDate query string false "End date for custom range (YYYY-MM-DD format, 23:59:59)"
// @Success 200 {object} domain.DashboardMetrics
// @Failure 400 {object} domain.APIError "Invalid timeRange value or date format"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /dashboard/metrics [get]
func (h *DashboardHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse date range parameters
	fromDateParam := r.URL.Query().Get("fromDate")
	toDateParam := r.URL.Query().Get("toDate")

	var dateRange *domain.DateRangeFilter

	// Parse custom date range if provided
	if fromDateParam != "" || toDateParam != "" {
		dateRange = &domain.DateRangeFilter{}

		if fromDateParam != "" {
			// Parse as YYYY-MM-DD and set to start of day (00:00:00)
			parsedDate, err := time.Parse("2006-01-02", fromDateParam)
			if err != nil {
				respondWithError(w, http.StatusBadRequest,
					fmt.Sprintf("Invalid fromDate format: '%s'. Expected YYYY-MM-DD (e.g., 2024-01-01)", fromDateParam))
				return
			}
			// Set to start of day
			startOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 0, 0, 0, 0, parsedDate.Location())
			dateRange.FromDate = &startOfDay
		}

		if toDateParam != "" {
			// Parse as YYYY-MM-DD and set to end of day (23:59:59)
			parsedDate, err := time.Parse("2006-01-02", toDateParam)
			if err != nil {
				respondWithError(w, http.StatusBadRequest,
					fmt.Sprintf("Invalid toDate format: '%s'. Expected YYYY-MM-DD (e.g., 2024-12-31)", toDateParam))
				return
			}
			// Set to end of day
			endOfDay := time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 23, 59, 59, 999999999, parsedDate.Location())
			dateRange.ToDate = &endOfDay
		}
	}

	// Parse timeRange query parameter (only used if no custom date range)
	timeRangeParam := r.URL.Query().Get("timeRange")
	var timeRange domain.TimeRange

	if dateRange != nil {
		// Custom date range overrides timeRange parameter
		timeRange = domain.TimeRangeCustom
	} else if timeRangeParam == "" {
		// Default to rolling 12 months
		timeRange = domain.TimeRangeRolling12Months
	} else {
		timeRange = domain.TimeRange(timeRangeParam)
		if !timeRange.IsValid() || timeRange == domain.TimeRangeCustom {
			respondWithError(w, http.StatusBadRequest,
				fmt.Sprintf("Invalid timeRange value: '%s'. Must be one of: rolling12months, allTime", timeRangeParam))
			return
		}
	}

	metrics, err := h.dashboardService.GetMetrics(r.Context(), timeRange, dateRange)
	if err != nil {
		h.logger.Error("failed to get dashboard metrics", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve dashboard metrics")
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
