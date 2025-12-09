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

// GetMetrics returns dashboard metrics using a rolling 12-month window
// All metrics exclude draft and expired offers from calculations
func (s *DashboardService) GetMetrics(ctx context.Context) (*domain.DashboardMetrics, error) {
	// Calculate 12-month window cutoff
	since := time.Now().AddDate(-1, 0, 0)
	const recentLimit = 5

	metrics := &domain.DashboardMetrics{
		Pipeline:         []domain.PipelinePhaseData{},
		RecentOffers:     []domain.OfferDTO{},
		RecentProjects:   []domain.ProjectDTO{},
		RecentActivities: []domain.ActivityDTO{},
		TopCustomers:     []domain.TopCustomerDTO{},
	}

	// Get offer statistics (12-month window, excluding drafts and expired)
	offerStats, err := s.offerRepo.GetDashboardOfferStats(ctx, since)
	if err != nil {
		s.logger.Warn("failed to get dashboard offer stats", zap.Error(err))
	} else {
		metrics.TotalOfferCount = offerStats.TotalOfferCount
		metrics.OfferReserve = offerStats.OfferReserve
		metrics.WeightedOfferReserve = offerStats.WeightedOfferReserve
		metrics.AverageProbability = offerStats.AverageProbability
	}

	// Get pipeline statistics by phase
	pipelineStats, err := s.offerRepo.GetDashboardPipelineStats(ctx, since)
	if err != nil {
		s.logger.Warn("failed to get dashboard pipeline stats", zap.Error(err))
	} else {
		for _, ps := range pipelineStats {
			metrics.Pipeline = append(metrics.Pipeline, domain.PipelinePhaseData{
				Phase:         ps.Phase,
				Count:         ps.Count,
				TotalValue:    ps.TotalValue,
				WeightedValue: ps.WeightedValue,
			})
		}
	}

	// Get win rate statistics (12-month window)
	winRateStats, err := s.offerRepo.GetDashboardWinRateStats(ctx, since)
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

	// Get project statistics (order reserve and total invoiced)
	projectStats, err := s.projectRepo.GetDashboardProjectStats(ctx, since)
	if err != nil {
		s.logger.Warn("failed to get dashboard project stats", zap.Error(err))
	} else {
		metrics.OrderReserve = projectStats.OrderReserve
		metrics.TotalInvoiced = projectStats.TotalInvoiced
		metrics.TotalValue = projectStats.OrderReserve + projectStats.TotalInvoiced
	}

	// Get recent offers (12-month window, excluding drafts)
	recentOffers, err := s.offerRepo.GetRecentOffersInWindow(ctx, since, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get recent offers", zap.Error(err))
	} else {
		for _, o := range recentOffers {
			metrics.RecentOffers = append(metrics.RecentOffers, mapper.ToOfferDTO(&o))
		}
	}

	// Get recent projects (12-month window)
	recentProjects, err := s.projectRepo.GetRecentProjectsInWindow(ctx, since, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get recent projects", zap.Error(err))
	} else {
		for _, p := range recentProjects {
			metrics.RecentProjects = append(metrics.RecentProjects, mapper.ToProjectDTO(&p))
		}
	}

	// Get recent activities (12-month window)
	recentActivities, err := s.activityRepo.GetRecentActivitiesInWindow(ctx, since, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get recent activities", zap.Error(err))
	} else {
		for _, a := range recentActivities {
			metrics.RecentActivities = append(metrics.RecentActivities, mapper.ToActivityDTO(&a))
		}
	}

	// Get top customers (12-month window, ranked by offer count)
	topCustomers, err := s.customerRepo.GetTopCustomersWithOfferStats(ctx, since, recentLimit)
	if err != nil {
		s.logger.Warn("failed to get top customers with offer stats", zap.Error(err))
	} else {
		for _, c := range topCustomers {
			metrics.TopCustomers = append(metrics.TopCustomers, domain.TopCustomerDTO{
				ID:            c.CustomerID,
				Name:          c.CustomerName,
				OrgNumber:     c.OrgNumber,
				OfferCount:    c.OfferCount,
				EconomicValue: c.EconomicValue,
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
		totalValue := 0.0
		activeOffers := 0
		customerDTOs[i] = mapper.ToCustomerDTO(&c, totalValue, activeOffers)
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
