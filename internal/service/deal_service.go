package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Stage transition rules: defines valid transitions between deal stages
var validStageTransitions = map[domain.DealStage][]domain.DealStage{
	domain.DealStageLead:        {domain.DealStageQualified, domain.DealStageLost},
	domain.DealStageQualified:   {domain.DealStageProposal, domain.DealStageLead, domain.DealStageLost},
	domain.DealStageProposal:    {domain.DealStageNegotiation, domain.DealStageQualified, domain.DealStageLost},
	domain.DealStageNegotiation: {domain.DealStageWon, domain.DealStageProposal, domain.DealStageLost},
	domain.DealStageWon:         {},                     // Terminal state
	domain.DealStageLost:        {domain.DealStageLead}, // Can reopen as new lead
}

// Default probabilities by stage
var stageProbabilities = map[domain.DealStage]int{
	domain.DealStageLead:        10,
	domain.DealStageQualified:   25,
	domain.DealStageProposal:    50,
	domain.DealStageNegotiation: 75,
	domain.DealStageWon:         100,
	domain.DealStageLost:        0,
}

type DealService struct {
	dealRepo         *repository.DealRepository
	historyRepo      *repository.DealStageHistoryRepository
	customerRepo     *repository.CustomerRepository
	projectRepo      *repository.ProjectRepository
	activityRepo     *repository.ActivityRepository
	offerRepo        *repository.OfferRepository
	budgetItemRepo   *repository.BudgetItemRepository
	notificationRepo *repository.NotificationRepository
	logger           *zap.Logger
	db               *gorm.DB
}

func NewDealService(
	dealRepo *repository.DealRepository,
	historyRepo *repository.DealStageHistoryRepository,
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	activityRepo *repository.ActivityRepository,
	offerRepo *repository.OfferRepository,
	budgetItemRepo *repository.BudgetItemRepository,
	notificationRepo *repository.NotificationRepository,
	logger *zap.Logger,
	db *gorm.DB,
) *DealService {
	return &DealService{
		dealRepo:         dealRepo,
		historyRepo:      historyRepo,
		customerRepo:     customerRepo,
		projectRepo:      projectRepo,
		activityRepo:     activityRepo,
		offerRepo:        offerRepo,
		budgetItemRepo:   budgetItemRepo,
		notificationRepo: notificationRepo,
		logger:           logger,
		db:               db,
	}
}

func (s *DealService) Create(ctx context.Context, req *domain.CreateDealRequest) (*domain.DealDTO, error) {
	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Set defaults
	stage := req.Stage
	if stage == "" {
		stage = domain.DealStageLead
	}

	probability := req.Probability
	if probability == 0 {
		probability = stageProbabilities[stage]
	}

	currency := req.Currency
	if currency == "" {
		currency = "NOK"
	}

	// Get owner name from context if available
	ownerName := ""
	var creatorName string
	if userCtx, ok := auth.FromContext(ctx); ok {
		creatorName = userCtx.DisplayName
		if req.OwnerID == userCtx.UserID.String() {
			ownerName = userCtx.DisplayName
		}
	}

	deal := &domain.Deal{
		Title:             req.Title,
		Description:       req.Description,
		CustomerID:        req.CustomerID,
		CustomerName:      customer.Name,
		CompanyID:         req.CompanyID,
		Stage:             stage,
		Probability:       probability,
		Value:             req.Value,
		Currency:          currency,
		ExpectedCloseDate: req.ExpectedCloseDate,
		OwnerID:           req.OwnerID,
		OwnerName:         ownerName,
		Source:            req.Source,
		Notes:             req.Notes,
		OfferID:           req.OfferID,
	}

	if err := s.dealRepo.Create(ctx, deal); err != nil {
		return nil, fmt.Errorf("failed to create deal: %w", err)
	}

	// Record initial stage history
	if err := s.historyRepo.RecordTransition(ctx, deal.ID, nil, stage, req.OwnerID, ownerName, "Deal created"); err != nil {
		s.logger.Warn("failed to record initial stage history", zap.Error(err))
	}

	// Create activity
	if creatorName != "" {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Deal created",
			Body:        fmt.Sprintf("Deal '%s' was created with value %s %.2f", deal.Title, deal.Currency, deal.Value),
			CreatorName: creatorName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	// Reload with relations
	deal, err = s.dealRepo.GetByID(ctx, deal.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload deal: %w", err)
	}

	dto := mapper.ToDealDTO(deal)
	return &dto, nil
}

func (s *DealService) GetByID(ctx context.Context, id uuid.UUID) (*domain.DealDTO, error) {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	dto := mapper.ToDealDTO(deal)
	return &dto, nil
}

func (s *DealService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateDealRequest) (*domain.DealDTO, error) {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Check permission (owner or manager can modify)
	if userCtx, ok := auth.FromContext(ctx); ok {
		if deal.OwnerID != userCtx.UserID.String() && !userCtx.HasAnyRole(domain.RoleManager, domain.RoleCompanyAdmin, domain.RoleSuperAdmin) {
			return nil, ErrForbidden
		}
	}

	oldStage := deal.Stage

	// Update fields
	deal.Title = req.Title
	deal.Description = req.Description
	if req.Stage != "" {
		deal.Stage = req.Stage
	}
	if req.Probability > 0 || req.Stage != "" {
		if req.Probability > 0 {
			deal.Probability = req.Probability
		} else if req.Stage != "" {
			deal.Probability = stageProbabilities[req.Stage]
		}
	}
	deal.Value = req.Value
	if req.Currency != "" {
		deal.Currency = req.Currency
	}
	deal.ExpectedCloseDate = req.ExpectedCloseDate
	deal.ActualCloseDate = req.ActualCloseDate
	if req.OwnerID != "" {
		deal.OwnerID = req.OwnerID
	}
	deal.Source = req.Source
	deal.Notes = req.Notes
	deal.LostReason = req.LostReason

	if err := s.dealRepo.Update(ctx, deal); err != nil {
		return nil, fmt.Errorf("failed to update deal: %w", err)
	}

	// Record stage change if different
	if req.Stage != "" && req.Stage != oldStage {
		var changedByID, changedByName string
		if userCtx, ok := auth.FromContext(ctx); ok {
			changedByID = userCtx.UserID.String()
			changedByName = userCtx.DisplayName
		}
		s.historyRepo.RecordTransition(ctx, deal.ID, &oldStage, req.Stage, changedByID, changedByName, "Stage updated via deal update")
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Deal updated",
			Body:        fmt.Sprintf("Deal '%s' was updated", deal.Title),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToDealDTO(deal)
	return &dto, nil
}

func (s *DealService) Delete(ctx context.Context, id uuid.UUID) error {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("deal not found: %w", err)
	}

	// Check permission
	if userCtx, ok := auth.FromContext(ctx); ok {
		if deal.OwnerID != userCtx.UserID.String() && !userCtx.HasAnyRole(domain.RoleManager, domain.RoleCompanyAdmin, domain.RoleSuperAdmin) {
			return ErrForbidden
		}
	}

	// Delete stage history first
	if err := s.historyRepo.DeleteByDealID(ctx, id); err != nil {
		s.logger.Warn("failed to delete deal stage history", zap.Error(err))
	}

	if err := s.dealRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete deal: %w", err)
	}

	return nil
}

func (s *DealService) List(ctx context.Context, page, pageSize int, filters *repository.DealFilters, sortBy repository.DealSortOption) (*domain.PaginatedResponse, error) {
	// Clamp page size
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page < 1 {
		page = 1
	}

	deals, total, err := s.dealRepo.List(ctx, page, pageSize, filters, sortBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list deals: %w", err)
	}

	dtos := make([]domain.DealDTO, len(deals))
	for i, deal := range deals {
		dtos[i] = mapper.ToDealDTO(&deal)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &domain.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// AdvanceStage moves a deal to the next stage with validation
func (s *DealService) AdvanceStage(ctx context.Context, id uuid.UUID, req *domain.UpdateDealStageRequest) (*domain.DealDTO, error) {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Validate stage transition
	if !s.isValidTransition(deal.Stage, req.Stage) {
		return nil, fmt.Errorf("invalid stage transition from %s to %s", deal.Stage, req.Stage)
	}

	oldStage := deal.Stage
	deal.Stage = req.Stage
	deal.Probability = stageProbabilities[req.Stage]

	if err := s.dealRepo.Update(ctx, deal); err != nil {
		return nil, fmt.Errorf("failed to update deal stage: %w", err)
	}

	// Record stage history
	var changedByID, changedByName string
	if userCtx, ok := auth.FromContext(ctx); ok {
		changedByID = userCtx.UserID.String()
		changedByName = userCtx.DisplayName
	}
	s.historyRepo.RecordTransition(ctx, deal.ID, &oldStage, req.Stage, changedByID, changedByName, req.Notes)

	// Create activity
	if changedByName != "" {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Deal stage changed",
			Body:        fmt.Sprintf("Deal '%s' moved from %s to %s", deal.Title, oldStage, req.Stage),
			CreatorName: changedByName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToDealDTO(deal)
	return &dto, nil
}

// WinDeal marks a deal as won and optionally creates a project
func (s *DealService) WinDeal(ctx context.Context, id uuid.UUID, createProject bool) (*domain.DealDTO, *domain.ProjectDTO, error) {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Can only win from negotiation stage
	if deal.Stage != domain.DealStageNegotiation {
		return nil, nil, fmt.Errorf("deal must be in negotiation stage to be won (current: %s)", deal.Stage)
	}

	oldStage := deal.Stage
	closeDate := time.Now()

	if err := s.dealRepo.MarkAsWon(ctx, id, closeDate); err != nil {
		return nil, nil, fmt.Errorf("failed to mark deal as won: %w", err)
	}

	// Record stage history
	var changedByID, changedByName string
	if userCtx, ok := auth.FromContext(ctx); ok {
		changedByID = userCtx.UserID.String()
		changedByName = userCtx.DisplayName
	}
	s.historyRepo.RecordTransition(ctx, deal.ID, &oldStage, domain.DealStageWon, changedByID, changedByName, "Deal won")

	// Create activity
	if changedByName != "" {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Deal won!",
			Body:        fmt.Sprintf("Deal '%s' was won with value %s %.2f", deal.Title, deal.Currency, deal.Value),
			CreatorName: changedByName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	// Reload deal
	deal, _ = s.dealRepo.GetByID(ctx, id)
	dealDTO := mapper.ToDealDTO(deal)

	// Create project if requested
	var projectDTO *domain.ProjectDTO
	if createProject {
		// Determine budget: inherit from offer if linked, otherwise use deal value
		budget := deal.Value
		var linkedOfferID *uuid.UUID

		if deal.OfferID != nil && s.offerRepo != nil {
			offer, err := s.offerRepo.GetByID(ctx, *deal.OfferID)
			if err == nil && offer != nil {
				// Use offer value as budget (offer value typically more accurate)
				budget = offer.Value
				linkedOfferID = deal.OfferID
				s.logger.Info("inherited budget from linked offer",
					zap.String("deal_id", deal.ID.String()),
					zap.String("offer_id", deal.OfferID.String()),
					zap.Float64("budget", budget))
			}
		}

		managerID := deal.OwnerID
		project := &domain.Project{
			Name:         deal.Title,
			Description:  deal.Description,
			CustomerID:   deal.CustomerID,
			CustomerName: deal.CustomerName,
			CompanyID:    deal.CompanyID,
			Status:       domain.ProjectStatusPlanning,
			StartDate:    closeDate,
			Value:        budget,
			ManagerID:    &managerID,
			ManagerName:  deal.OwnerName,
			DealID:       &deal.ID,
			OfferID:      linkedOfferID,
		}

		if err := s.projectRepo.Create(ctx, project); err != nil {
			s.logger.Error("failed to create project from deal", zap.Error(err))
			return nil, nil, fmt.Errorf("failed to create project: %w", err)
		} else {
			dto := mapper.ToProjectDTO(project)
			projectDTO = &dto

			// Create activity for project creation
			if changedByName != "" {
				budgetSource := "deal value"
				if linkedOfferID != nil {
					budgetSource = "linked offer"
				}
				activity := &domain.Activity{
					TargetType:  domain.ActivityTargetProject,
					TargetID:    project.ID,
					Title:       "Project created from deal",
					Body:        fmt.Sprintf("Project '%s' created from won deal with budget %.2f (from %s)", project.Name, budget, budgetSource),
					CreatorName: changedByName,
				}
				s.activityRepo.Create(ctx, activity)
			}
		}
	}

	// Send notifications to stakeholders about the deal win
	s.notifyDealWon(ctx, deal, changedByID)

	return &dealDTO, projectDTO, nil
}

// notifyDealWon sends notifications to relevant stakeholders when a deal is won
func (s *DealService) notifyDealWon(ctx context.Context, deal *domain.Deal, winnerID string) {
	if s.notificationRepo == nil {
		return
	}

	// Parse winner ID to UUID for notifications
	winnerUUID, err := uuid.Parse(winnerID)
	if err != nil {
		s.logger.Warn("invalid winner ID for notifications", zap.String("winner_id", winnerID))
		return
	}

	// Notify the deal owner (if not the winner)
	if deal.OwnerID != "" && deal.OwnerID != winnerID {
		ownerUUID, err := uuid.Parse(deal.OwnerID)
		if err == nil {
			notification := &domain.Notification{
				UserID:     ownerUUID,
				Type:       "deal_won",
				Title:      "Deal Won!",
				Message:    fmt.Sprintf("Congratulations! Deal '%s' has been won with value %s %.2f", deal.Title, deal.Currency, deal.Value),
				EntityID:   &deal.ID,
				EntityType: "deal",
			}
			if err := s.notificationRepo.Create(ctx, notification); err != nil {
				s.logger.Warn("failed to create notification for deal owner", zap.Error(err))
			}
		}
	}

	// Notify the winner themselves (confirmation notification)
	notification := &domain.Notification{
		UserID:     winnerUUID,
		Type:       "deal_won_confirmation",
		Title:      "Deal Closed Successfully",
		Message:    fmt.Sprintf("You've successfully closed deal '%s' with value %s %.2f", deal.Title, deal.Currency, deal.Value),
		EntityID:   &deal.ID,
		EntityType: "deal",
	}
	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		s.logger.Warn("failed to create confirmation notification", zap.Error(err))
	}

	s.logger.Info("deal win notifications sent",
		zap.String("deal_id", deal.ID.String()),
		zap.String("deal_title", deal.Title),
		zap.Float64("value", deal.Value))
}

// LoseDeal marks a deal as lost with a categorized reason and detailed notes
func (s *DealService) LoseDeal(ctx context.Context, id uuid.UUID, req *domain.LoseDealRequest) (*domain.DealDTO, error) {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Can't lose an already won deal
	if deal.Stage == domain.DealStageWon {
		return nil, fmt.Errorf("cannot mark a won deal as lost")
	}

	// Can't lose an already lost deal
	if deal.Stage == domain.DealStageLost {
		return nil, fmt.Errorf("deal is already marked as lost")
	}

	oldStage := deal.Stage
	closeDate := time.Now()

	if err := s.dealRepo.MarkAsLost(ctx, id, closeDate, req.Reason, req.Notes); err != nil {
		return nil, fmt.Errorf("failed to mark deal as lost: %w", err)
	}

	// Record stage history with both category and notes
	historyNotes := fmt.Sprintf("[%s] %s", req.Reason, req.Notes)
	var changedByID, changedByName string
	if userCtx, ok := auth.FromContext(ctx); ok {
		changedByID = userCtx.UserID.String()
		changedByName = userCtx.DisplayName
	}
	s.historyRepo.RecordTransition(ctx, deal.ID, &oldStage, domain.DealStageLost, changedByID, changedByName, historyNotes)

	// Create activity with category context
	if changedByName != "" {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Deal lost",
			Body:        fmt.Sprintf("Deal '%s' was lost. Category: %s. Details: %s", deal.Title, req.Reason, req.Notes),
			CreatorName: changedByName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	// Reload deal
	deal, _ = s.dealRepo.GetByID(ctx, id)
	dto := mapper.ToDealDTO(deal)
	return &dto, nil
}

// ReopenDeal reopens a lost deal as a new lead
func (s *DealService) ReopenDeal(ctx context.Context, id uuid.UUID) (*domain.DealDTO, error) {
	deal, err := s.dealRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	if deal.Stage != domain.DealStageLost {
		return nil, fmt.Errorf("only lost deals can be reopened")
	}

	oldStage := deal.Stage
	deal.Stage = domain.DealStageLead
	deal.Probability = stageProbabilities[domain.DealStageLead]
	deal.ActualCloseDate = nil
	deal.LostReason = ""
	deal.LossReasonCategory = nil

	if err := s.dealRepo.Update(ctx, deal); err != nil {
		return nil, fmt.Errorf("failed to reopen deal: %w", err)
	}

	// Record stage history
	var changedByID, changedByName string
	if userCtx, ok := auth.FromContext(ctx); ok {
		changedByID = userCtx.UserID.String()
		changedByName = userCtx.DisplayName
	}
	s.historyRepo.RecordTransition(ctx, deal.ID, &oldStage, domain.DealStageLead, changedByID, changedByName, "Deal reopened")

	// Create activity
	if changedByName != "" {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Deal reopened",
			Body:        fmt.Sprintf("Deal '%s' was reopened as a new lead", deal.Title),
			CreatorName: changedByName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToDealDTO(deal)
	return &dto, nil
}

// GetStageHistory returns the stage history for a deal
func (s *DealService) GetStageHistory(ctx context.Context, dealID uuid.UUID) ([]domain.DealStageHistoryDTO, error) {
	history, err := s.historyRepo.GetByDealID(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage history: %w", err)
	}

	dtos := make([]domain.DealStageHistoryDTO, len(history))
	for i, h := range history {
		dtos[i] = mapper.ToDealStageHistoryDTO(&h)
	}

	return dtos, nil
}

// GetPipelineOverview returns deals grouped by stage
func (s *DealService) GetPipelineOverview(ctx context.Context) (map[string][]domain.DealDTO, error) {
	pipeline, err := s.dealRepo.GetPipelineOverview(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline overview: %w", err)
	}

	result := make(map[string][]domain.DealDTO)
	for stage, deals := range pipeline {
		dtos := make([]domain.DealDTO, len(deals))
		for i, deal := range deals {
			dtos[i] = mapper.ToDealDTO(&deal)
		}
		result[string(stage)] = dtos
	}

	return result, nil
}

// GetPipelineStats returns aggregated statistics
func (s *DealService) GetPipelineStats(ctx context.Context) (*repository.PipelineStats, error) {
	return s.dealRepo.GetPipelineStats(ctx)
}

// GetForecast returns pipeline forecast for upcoming months
func (s *DealService) GetForecast(ctx context.Context, months int) ([]repository.ForecastPeriod, error) {
	if months < 1 {
		months = 3
	}
	if months > 12 {
		months = 12
	}
	return s.dealRepo.GetForecast(ctx, months)
}

// GetActivities returns activities for a deal
func (s *DealService) GetActivities(ctx context.Context, id uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	activities, err := s.activityRepo.ListByTarget(ctx, domain.ActivityTargetDeal, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}

// isValidTransition checks if a stage transition is allowed
func (s *DealService) isValidTransition(from, to domain.DealStage) bool {
	validNextStages, ok := validStageTransitions[from]
	if !ok {
		return false
	}

	for _, validStage := range validNextStages {
		if validStage == to {
			return true
		}
	}
	return false
}

// GetPipelineAnalytics returns comprehensive pipeline analytics with forecasts and conversion rates
func (s *DealService) GetPipelineAnalytics(ctx context.Context, filters *domain.PipelineAnalyticsFilters) (*domain.PipelineAnalyticsDTO, error) {
	analytics := &domain.PipelineAnalyticsDTO{
		Summary:         make([]domain.StageSummaryDTO, 0),
		ConversionRates: make([]domain.ConversionRateDTO, 0),
		GeneratedAt:     time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Extract filter values
	var companyID *domain.CompanyID
	var ownerID *string
	var dateFrom, dateTo *time.Time

	if filters != nil {
		companyID = filters.CompanyID
		ownerID = filters.OwnerID
		dateFrom = filters.DateFrom
		dateTo = filters.DateTo
	}

	// Get pipeline summary from the database view
	summaryRows, err := s.dealRepo.GetPipelineSummaryFromView(ctx, companyID)
	if err != nil {
		s.logger.Error("failed to get pipeline summary from view", zap.Error(err))
		// Continue with empty summary rather than failing entirely
	} else {
		for _, row := range summaryRows {
			analytics.Summary = append(analytics.Summary, domain.StageSummaryDTO{
				Stage:          string(row.Stage),
				DealCount:      row.DealCount,
				TotalValue:     row.TotalValue,
				WeightedValue:  row.TotalWeightedValue,
				AvgProbability: row.AvgProbability,
				AvgDealValue:   row.AvgDealValue,
				OverdueCount:   row.OverdueCount,
			})
		}
	}

	// Get 30-day forecast
	forecast30, err := s.dealRepo.GetRevenueForecastByDays(ctx, 30, companyID, ownerID)
	if err != nil {
		s.logger.Warn("failed to get 30-day forecast", zap.Error(err))
	} else {
		analytics.Forecast30Days = domain.RevenueForecastDTO{
			Period:        "30d",
			DealCount:     forecast30.DealCount,
			TotalValue:    forecast30.TotalValue,
			WeightedValue: forecast30.WeightedValue,
		}
	}

	// Get 90-day forecast
	forecast90, err := s.dealRepo.GetRevenueForecastByDays(ctx, 90, companyID, ownerID)
	if err != nil {
		s.logger.Warn("failed to get 90-day forecast", zap.Error(err))
	} else {
		analytics.Forecast90Days = domain.RevenueForecastDTO{
			Period:        "90d",
			DealCount:     forecast90.DealCount,
			TotalValue:    forecast90.TotalValue,
			WeightedValue: forecast90.WeightedValue,
		}
	}

	// Get conversion rates
	conversionRates, err := s.dealRepo.GetConversionRates(ctx, companyID)
	if err != nil {
		s.logger.Warn("failed to get conversion rates", zap.Error(err))
	} else {
		for _, rate := range conversionRates {
			analytics.ConversionRates = append(analytics.ConversionRates, domain.ConversionRateDTO{
				FromStage:      string(rate.FromStage),
				ToStage:        string(rate.ToStage),
				ConversionRate: rate.ConversionRate,
				DealsConverted: rate.DealsConverted,
				TotalDeals:     rate.TotalDeals,
			})
		}
	}

	// Get win rate analysis
	winRate, err := s.dealRepo.GetWinRateAnalysis(ctx, companyID, ownerID, dateFrom, dateTo)
	if err != nil {
		s.logger.Warn("failed to get win rate analysis", zap.Error(err))
	} else {
		analytics.WinRateAnalysis = domain.WinRateAnalysisDTO{
			TotalClosed:      winRate.TotalClosed,
			TotalWon:         winRate.TotalWon,
			TotalLost:        winRate.TotalLost,
			WinRate:          winRate.WinRate,
			WonValue:         winRate.WonValue,
			LostValue:        winRate.LostValue,
			AvgWonDealValue:  winRate.AvgWonDealValue,
			AvgLostDealValue: winRate.AvgLostDealValue,
			AvgDaysToClose:   winRate.AvgDaysToClose,
		}
	}

	return analytics, nil
}

// CreateOfferFromDeal creates an offer linked to a deal, advancing the deal to proposal stage
func (s *DealService) CreateOfferFromDeal(ctx context.Context, dealID uuid.UUID, req *domain.CreateOfferFromDealRequest) (*domain.CreateOfferFromDealResponse, error) {
	// Get deal
	deal, err := s.dealRepo.GetByID(ctx, dealID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDealNotFound
		}
		return nil, fmt.Errorf("failed to get deal: %w", err)
	}

	// Validate deal stage - only lead or qualified can create an offer
	if deal.Stage != domain.DealStageLead && deal.Stage != domain.DealStageQualified {
		return nil, ErrDealInvalidStageForOffer
	}

	// Ensure deal doesn't already have a linked offer
	if deal.OfferID != nil {
		return nil, ErrDealAlreadyHasOffer
	}

	// Determine offer title
	offerTitle := req.Title
	if offerTitle == "" {
		offerTitle = deal.Title
	}

	// Get user context
	var creatorName, ownerID string
	if userCtx, ok := auth.FromContext(ctx); ok {
		creatorName = userCtx.DisplayName
		ownerID = userCtx.UserID.String()
	}

	// Use deal owner as responsible user if no user context
	responsibleUserID := ownerID
	responsibleUserName := creatorName
	if responsibleUserID == "" {
		responsibleUserID = deal.OwnerID
		responsibleUserName = deal.OwnerName
	}

	var createdOffer *domain.Offer
	oldStage := deal.Stage

	// Use transaction for atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create the offer
		createdOffer = &domain.Offer{
			Title:               offerTitle,
			CustomerID:          deal.CustomerID,
			CustomerName:        deal.CustomerName,
			CompanyID:           deal.CompanyID,
			Phase:               domain.OfferPhaseDraft,
			Probability:         stageProbabilities[domain.DealStageProposal], // 50%
			Value:               deal.Value,
			Status:              domain.OfferStatusActive,
			ResponsibleUserID:   responsibleUserID,
			ResponsibleUserName: responsibleUserName,
			Description:         deal.Description,
			Notes:               deal.Notes,
		}

		if err := tx.Create(createdOffer).Error; err != nil {
			return fmt.Errorf("failed to create offer: %w", err)
		}

		// If template offer provided, copy budget items
		if req.TemplateOfferID != nil && s.budgetItemRepo != nil {
			items, err := s.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, *req.TemplateOfferID)
			if err == nil && len(items) > 0 {
				s.logger.Info("copying budget items from template offer",
					zap.String("template_offer_id", req.TemplateOfferID.String()),
					zap.String("new_offer_id", createdOffer.ID.String()),
					zap.Int("item_count", len(items)))

				totalRevenue := 0.0
				for _, item := range items {
					cloned := domain.BudgetItem{
						ParentType:     domain.BudgetParentOffer,
						ParentID:       createdOffer.ID,
						Name:           item.Name,
						ExpectedCost:   item.ExpectedCost,
						ExpectedMargin: item.ExpectedMargin,
						Quantity:       item.Quantity,
						PricePerItem:   item.PricePerItem,
						Description:    item.Description,
						DisplayOrder:   item.DisplayOrder,
					}
					if err := tx.Create(&cloned).Error; err != nil {
						s.logger.Warn("failed to clone budget item",
							zap.Error(err),
							zap.String("item_id", item.ID.String()))
					}
					totalRevenue += item.ExpectedRevenue
				}

				// Update offer value from items if template was used
				if totalRevenue > 0 {
					createdOffer.Value = totalRevenue
					if err := tx.Save(createdOffer).Error; err != nil {
						s.logger.Warn("failed to update offer value from items", zap.Error(err))
					}
				}
			}
		}

		// Update deal: link to offer and advance to proposal stage
		deal.OfferID = &createdOffer.ID
		deal.Stage = domain.DealStageProposal
		deal.Probability = stageProbabilities[domain.DealStageProposal]

		if err := tx.Save(deal).Error; err != nil {
			return fmt.Errorf("failed to update deal: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Record stage history (outside transaction for non-critical operation)
	if creatorName != "" {
		s.historyRepo.RecordTransition(ctx, deal.ID, &oldStage, domain.DealStageProposal, ownerID, creatorName, "Offer created from deal")
	}

	// Log activity on deal
	if creatorName != "" {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetDeal,
			TargetID:    deal.ID,
			Title:       "Offer created from deal",
			Body:        fmt.Sprintf("Deal '%s' advanced to proposal stage with new offer '%s'", deal.Title, createdOffer.Title),
			CreatorName: creatorName,
		}
		s.activityRepo.Create(ctx, activity)

		// Log activity on offer
		offerActivity := &domain.Activity{
			TargetType:  domain.ActivityTargetOffer,
			TargetID:    createdOffer.ID,
			Title:       "Offer created from deal",
			Body:        fmt.Sprintf("Offer '%s' was created from deal '%s'", createdOffer.Title, deal.Title),
			CreatorName: creatorName,
		}
		s.activityRepo.Create(ctx, offerActivity)
	}

	// Reload offer with relations
	if s.offerRepo != nil {
		reloadedOffer, err := s.offerRepo.GetByID(ctx, createdOffer.ID)
		if err == nil {
			createdOffer = reloadedOffer
		}
	}

	// Reload deal
	deal, _ = s.dealRepo.GetByID(ctx, dealID)

	// Build response
	offerDTO := mapper.ToOfferDTO(createdOffer)
	dealDTO := mapper.ToDealDTO(deal)

	return &domain.CreateOfferFromDealResponse{
		Offer: &offerDTO,
		Deal:  &dealDTO,
	}, nil
}
