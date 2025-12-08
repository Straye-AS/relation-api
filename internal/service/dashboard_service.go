package service

import (
	"context"
	"fmt"

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

func (s *DashboardService) GetMetrics(ctx context.Context) (*domain.DashboardMetrics, error) {
	// Get offer statistics
	offerStats, err := s.offerRepo.GetOfferStats(ctx)
	if err != nil {
		s.logger.Warn("failed to get offer stats", zap.Error(err))
		offerStats = &repository.OfferStats{ByPhase: make(map[domain.OfferPhase]int)}
	}

	// Get win rate
	winRate, err := s.offerRepo.GetWinRate(ctx)
	if err != nil {
		s.logger.Warn("failed to get win rate", zap.Error(err))
	}

	// Get recent offers
	recentOffers, err := s.offerRepo.GetRecentOffers(ctx, 5)
	if err != nil {
		s.logger.Warn("failed to get recent offers", zap.Error(err))
	}
	recentOfferDTOs := make([]domain.OfferDTO, len(recentOffers))
	for i, o := range recentOffers {
		recentOfferDTOs[i] = mapper.ToOfferDTO(&o)
	}

	// Get active projects
	activeProjects, err := s.projectRepo.GetActiveProjects(ctx, 5)
	if err != nil {
		s.logger.Warn("failed to get active projects", zap.Error(err))
	}
	activeProjectDTOs := make([]domain.ProjectDTO, len(activeProjects))
	for i, p := range activeProjects {
		activeProjectDTOs[i] = mapper.ToProjectDTO(&p)
	}

	// Get recent projects
	recentProjects, err := s.projectRepo.GetRecentProjects(ctx, 5)
	if err != nil {
		s.logger.Warn("failed to get recent projects", zap.Error(err))
	}
	recentProjectDTOs := make([]domain.ProjectDTO, len(recentProjects))
	for i, p := range recentProjects {
		recentProjectDTOs[i] = mapper.ToProjectDTO(&p)
	}

	// Get top customers
	topCustomers, err := s.customerRepo.GetTopCustomers(ctx, 5)
	if err != nil {
		s.logger.Warn("failed to get top customers", zap.Error(err))
	}
	topCustomerDTOs := make([]domain.CustomerDTO, len(topCustomers))
	for i, c := range topCustomers {
		topCustomerDTOs[i] = mapper.ToCustomerDTO(&c, 0, 0)
	}

	// Get recent activities
	recentActivities, err := s.activityRepo.GetRecentActivities(ctx, 10)
	if err != nil {
		s.logger.Warn("failed to get recent activities", zap.Error(err))
	}
	recentActivityDTOs := make([]domain.ActivityDTO, len(recentActivities))
	for i, a := range recentActivities {
		recentActivityDTOs[i] = mapper.ToActivityDTO(&a)
	}

	// Build pipeline phase data
	pipeline := []domain.PipelinePhaseData{}
	for phase, count := range offerStats.ByPhase {
		pipeline = append(pipeline, domain.PipelinePhaseData{
			Phase: phase,
			Count: count,
		})
	}

	return &domain.DashboardMetrics{
		TotalOffers:           int(offerStats.TotalOffers),
		ActiveOffers:          int(offerStats.ActiveOffers),
		WonOffers:             int(offerStats.WonOffers),
		LostOffers:            int(offerStats.LostOffers),
		TotalValue:            offerStats.TotalValue,
		WeightedValue:         offerStats.WeightedValue,
		AverageProbability:    offerStats.AvgProbability,
		OffersByPhase:         offerStats.ByPhase,
		Pipeline:              pipeline,
		OfferReserve:          offerStats.TotalValue,
		WinRate:               winRate,
		RevenueForecast30Days: offerStats.WeightedValue * 0.3, // Simple estimate
		RevenueForecast90Days: offerStats.WeightedValue * 0.7, // Simple estimate
		TopDisciplines:        []domain.DisciplineStats{},
		ActiveProjects:        activeProjectDTOs,
		TopCustomers:          topCustomerDTOs,
		TeamPerformance:       []domain.TeamMemberStats{},
		RecentOffers:          recentOfferDTOs,
		RecentProjects:        recentProjectDTOs,
		RecentActivities:      recentActivityDTOs,
	}, nil
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
