package service

import (
	"context"
	"fmt"
	"time"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

type DashboardService struct {
	customerRepo     *repository.CustomerRepository
	projectRepo      *repository.ProjectRepository
	offerRepo        *repository.OfferRepository
	activityRepo     *repository.ActivityRepository
	notificationRepo *repository.NotificationRepository
	logger           *zap.Logger
}

func NewDashboardService(
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	offerRepo *repository.OfferRepository,
	activityRepo *repository.ActivityRepository,
	notificationRepo *repository.NotificationRepository,
	logger *zap.Logger,
) *DashboardService {
	return &DashboardService{
		customerRepo:     customerRepo,
		projectRepo:      projectRepo,
		offerRepo:        offerRepo,
		activityRepo:     activityRepo,
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

// GetMetrics returns dashboard metrics with configurable time range
// timeRange can be "rolling12months" (default), "allTime", or "custom" (when dateRange is provided)
// All metrics exclude draft and expired offers from calculations
//
// IMPORTANT: Pipeline and offer metrics use aggregation to avoid double-counting.
// When a project has multiple offers, only the highest value offer per phase is counted.
// Orphan offers (without project) are included at full value.
func (s *DashboardService) GetMetrics(ctx context.Context, timeRange domain.TimeRange, dateRange *domain.DateRangeFilter) (*domain.DashboardMetrics, error) {
	// Default to rolling 12 months if not specified or invalid
	if timeRange == "" {
		timeRange = domain.TimeRangeRolling12Months
	}

	// Calculate date filter based on time range or custom date range
	var since *time.Time
	var fromDate, toDate *time.Time

	if timeRange == domain.TimeRangeCustom && dateRange != nil {
		// Use custom date range
		fromDate = dateRange.FromDate
		toDate = dateRange.ToDate
		// For backward compatibility with existing pipeline stats, use fromDate as "since"
		since = fromDate
	} else if timeRange == domain.TimeRangeRolling12Months {
		t := time.Now().AddDate(-1, 0, 0)
		since = &t
		fromDate = &t
	}
	// For allTime, since/fromDate/toDate remain nil (no date filter)

	const recentLimit = 5

	metrics := &domain.DashboardMetrics{
		TimeRange:        timeRange,
		Pipeline:         []domain.PipelinePhaseData{},
		RecentOffers:     []domain.OfferDTO{},
		RecentOrders:     []domain.OfferDTO{},
		RecentActivities: []domain.ActivityDTO{},
		TopCustomers:     []domain.TopCustomerDTO{},
	}

	// Set date range in response for custom ranges
	if timeRange == domain.TimeRangeCustom && dateRange != nil {
		if dateRange.FromDate != nil {
			fromStr := dateRange.FromDate.UTC().Format(time.RFC3339)
			metrics.FromDate = &fromStr
		}
		if dateRange.ToDate != nil {
			toStr := dateRange.ToDate.UTC().Format(time.RFC3339)
			metrics.ToDate = &toStr
		}
	}

	// Get aggregated offer statistics (using aggregation to avoid double-counting)
	// This replaces the old GetDashboardOfferStats method
	aggregatedStats, err := s.offerRepo.GetAggregatedOfferStats(ctx, fromDate, toDate)
	if err != nil {
		s.logger.Warn("failed to get aggregated offer stats", zap.Error(err))
	} else {
		metrics.TotalOfferCount = aggregatedStats.TotalOfferCount
		metrics.TotalProjectCount = aggregatedStats.TotalProjectCount
		metrics.OfferReserve = aggregatedStats.OfferReserve
		metrics.WeightedOfferReserve = aggregatedStats.WeightedOfferReserve
		metrics.AverageProbability = aggregatedStats.AverageProbability
	}

	// Get aggregated pipeline statistics by phase (using aggregation to avoid double-counting)
	// This replaces the old GetDashboardPipelineStats method
	aggregatedPipelineStats, err := s.offerRepo.GetAggregatedPipelineStats(ctx, fromDate, toDate)
	if err != nil {
		s.logger.Warn("failed to get aggregated pipeline stats", zap.Error(err))
	} else {
		for _, ps := range aggregatedPipelineStats {
			metrics.Pipeline = append(metrics.Pipeline, domain.PipelinePhaseData{
				Phase:         ps.Phase,
				Count:         ps.OfferCount,
				ProjectCount:  ps.ProjectCount,
				TotalValue:    ps.TotalValue,
				WeightedValue: ps.WeightedValue,
			})
		}
	}

	// Get win rate statistics
	winRateStats, err := s.offerRepo.GetDashboardWinRateStats(ctx, fromDate, toDate)
	if err != nil {
		s.logger.Warn("failed to get dashboard win rate stats", zap.Error(err))
	} else {
		metrics.WinRateMetrics = domain.WinRateMetrics{
			WonCount:        winRateStats.WonCount,
			LostCount:       winRateStats.LostCount,
			WonValue:        winRateStats.WonValue,
			LostValue:       winRateStats.LostValue,
			WinRate:         winRateStats.WinRate,
			EconomicWinRate: winRateStats.EconomicWinRate,
		}
	}

	// Order Metrics (from offers in order/completed phases)
	// Economic tracking has moved from projects to offers
	orderStats, err := s.offerRepo.GetOrderStats(ctx, nil, fromDate, toDate)
	if err != nil {
		s.logger.Warn("failed to get order stats", zap.Error(err))
	} else {
		metrics.ActiveOrderCount = int(orderStats.TotalOrders)
		metrics.CompletedOrderCount = int(orderStats.CompletedOrders)
		metrics.OrderValue = orderStats.TotalValue
		metrics.OrderReserve = orderStats.OrderReserve
		metrics.TotalInvoiced = orderStats.TotalInvoiced
		metrics.TotalSpent = orderStats.TotalSpent
		metrics.AverageOrderProgress = orderStats.AvgCompletion
		metrics.TotalValue = orderStats.OrderReserve + orderStats.TotalInvoiced

		// Map health distribution from offer repository stats
		metrics.HealthDistribution = domain.HealthDistribution{
			OnTrack:    orderStats.ByHealth[domain.OfferHealthOnTrack],
			AtRisk:     orderStats.ByHealth[domain.OfferHealthAtRisk],
			Delayed:    orderStats.ByHealth[domain.OfferHealthDelayed],
			OverBudget: orderStats.ByHealth[domain.OfferHealthOverBudget],
		}
	}

	// Get recent offers (in_progress phase only - "Siste tilbud")
	// Uses the new GetRecentOffersByPhase method with date range support
	recentOffers, err := s.offerRepo.GetRecentOffersByPhase(ctx, domain.OfferPhaseInProgress, fromDate, toDate, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get recent offers", zap.Error(err))
	} else {
		for _, o := range recentOffers {
			metrics.RecentOffers = append(metrics.RecentOffers, mapper.ToOfferDTO(&o))
		}
	}

	// Get recent orders (order phase - "Siste ordre")
	// Uses the new GetRecentOffersByPhase method with date range support
	recentOrders, err := s.offerRepo.GetRecentOffersByPhase(ctx, domain.OfferPhaseOrder, fromDate, toDate, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get recent orders", zap.Error(err))
	} else {
		for _, o := range recentOrders {
			metrics.RecentOrders = append(metrics.RecentOrders, mapper.ToOfferDTO(&o))
		}
	}

	// Get recent activities
	recentActivities, err := s.activityRepo.GetRecentActivitiesInWindow(ctx, since, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get recent activities", zap.Error(err))
	} else {
		for _, a := range recentActivities {
			metrics.RecentActivities = append(metrics.RecentActivities, mapper.ToActivityDTO(&a))
		}
	}

	// Get top customers (ranked by won offer count and value - order + completed phases)
	// Uses the same date filter as other metrics
	topCustomers, err := s.customerRepo.GetTopCustomersWithWonStats(ctx, fromDate, toDate, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get top customers with won stats", zap.Error(err))
	} else {
		for _, c := range topCustomers {
			metrics.TopCustomers = append(metrics.TopCustomers, domain.TopCustomerDTO{
				ID:            c.CustomerID,
				Name:          c.CustomerName,
				OrgNumber:     c.OrgNumber,
				WonOfferCount: c.WonOfferCount,
				WonOfferValue: c.WonOfferValue,
			})
		}
	}

	return metrics, nil
}

func (s *DashboardService) Search(ctx context.Context, query string) (*domain.SearchResults, error) {
	limit := 10

	customers, err := s.customerRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	projects, err := s.projectRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	offers, err := s.offerRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search offers: %w", err)
	}

	// Convert to DTOs
	customerDTOs := make([]domain.CustomerDTO, len(customers))
	for i, c := range customers {
		// TODO: Calculate actual values
		totalValueActive := 0.0
		totalValueWon := 0.0
		activeOffers := 0
		customerDTOs[i] = mapper.ToCustomerDTO(&c, totalValueActive, totalValueWon, activeOffers)
	}

	projectDTOs := make([]domain.ProjectDTO, len(projects))
	for i, p := range projects {
		projectDTOs[i] = mapper.ToProjectDTO(&p)
	}

	offerDTOs := make([]domain.OfferDTO, len(offers))
	for i, o := range offers {
		offerDTOs[i] = mapper.ToOfferDTO(&o)
	}

	total := len(customers) + len(projects) + len(offers)

	return &domain.SearchResults{
		Customers: customerDTOs,
		Projects:  projectDTOs,
		Offers:    offerDTOs,
		Contacts:  []domain.ContactDTO{}, // TODO: Add contact search
		Total:     total,
	}, nil
}
