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
	notificationRepo *repository.NotificationRepository
	logger           *zap.Logger
}

func NewDashboardService(
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	offerRepo *repository.OfferRepository,
	notificationRepo *repository.NotificationRepository,
	logger *zap.Logger,
) *DashboardService {
	return &DashboardService{
		customerRepo:     customerRepo,
		projectRepo:      projectRepo,
		offerRepo:        offerRepo,
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

func (s *DashboardService) GetMetrics(ctx context.Context) (*domain.DashboardMetrics, error) {
	pipelineValue, err := s.offerRepo.GetTotalPipelineValue(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline value: %w", err)
	}

	// TODO: Implement comprehensive dashboard metrics
	// These would come from proper queries on offer repository
	return &domain.DashboardMetrics{
		TotalOffers:           0,
		ActiveOffers:          0,
		WonOffers:             0,
		LostOffers:            0,
		TotalValue:            pipelineValue,
		WeightedValue:         pipelineValue,
		AverageProbability:    0,
		OffersByPhase:         make(map[domain.OfferPhase]int),
		Pipeline:              []domain.PipelinePhaseData{},
		OfferReserve:          pipelineValue,
		WinRate:               0,
		RevenueForecast30Days: 0,
		RevenueForecast90Days: 0,
		TopDisciplines:        []domain.DisciplineStats{},
		ActiveProjects:        []domain.ProjectDTO{},
		TopCustomers:          []domain.CustomerDTO{},
		TeamPerformance:       []domain.TeamMemberStats{},
		RecentOffers:          []domain.OfferDTO{},
		RecentProjects:        []domain.ProjectDTO{},
		RecentActivities:      []domain.NotificationDTO{},
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
