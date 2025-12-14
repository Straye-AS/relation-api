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

// Project-specific service errors
var (
	// ErrInvalidStatusTransition is returned when trying to make an invalid status transition
	ErrInvalidStatusTransition = errors.New("invalid project status transition")

	// ErrInvalidCompletionPercent is returned when completion percent is out of range
	ErrInvalidCompletionPercent = errors.New("completion percent must be between 0 and 100")

	// ErrOfferNotWon is returned when trying to inherit from an offer that is not won
	ErrOfferNotWon = errors.New("can only inherit budget from won offers")

	// ErrProjectNotManager is returned when user is not the project manager
	ErrProjectNotManager = errors.New("user is not the project manager")

	// ErrCannotReopenProject is returned when a project cannot be reopened
	ErrCannotReopenProject = errors.New("project cannot be reopened from its current state")

	// ErrWorkingPhaseRequiresStartDate is returned when transitioning to working without a start date
	ErrWorkingPhaseRequiresStartDate = errors.New("working phase requires a start date")

	// ErrInvalidPhaseTransition is returned when trying to make an invalid phase transition
	ErrInvalidPhaseTransition = errors.New("invalid project phase transition")

	// ErrTilbudPhaseCannotHaveWinningOffer is returned when a project in tilbud phase has a winning offer
	ErrTilbudPhaseCannotHaveWinningOffer = errors.New("project in tilbud phase cannot have a winning offer")
)

// ProjectService handles business logic for projects
type ProjectService struct {
	projectRepo      *repository.ProjectRepository
	offerRepo        *repository.OfferRepository
	customerRepo     *repository.CustomerRepository
	budgetItemRepo   *repository.BudgetItemRepository
	activityRepo     *repository.ActivityRepository
	userRepo         *repository.UserRepository
	companyService   *CompanyService
	numberSeqService *NumberSequenceService
	logger           *zap.Logger
	db               *gorm.DB
}

// NewProjectService creates a new ProjectService with basic dependencies
func NewProjectService(
	projectRepo *repository.ProjectRepository,
	customerRepo *repository.CustomerRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *ProjectService {
	return &ProjectService{
		projectRepo:  projectRepo,
		customerRepo: customerRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

// NewProjectServiceWithDeps creates a ProjectService with all dependencies for full feature support
func NewProjectServiceWithDeps(
	projectRepo *repository.ProjectRepository,
	offerRepo *repository.OfferRepository,
	customerRepo *repository.CustomerRepository,
	budgetItemRepo *repository.BudgetItemRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	companyService *CompanyService,
	numberSeqService *NumberSequenceService,
	logger *zap.Logger,
	db *gorm.DB,
) *ProjectService {
	return &ProjectService{
		projectRepo:      projectRepo,
		offerRepo:        offerRepo,
		customerRepo:     customerRepo,
		budgetItemRepo:   budgetItemRepo,
		activityRepo:     activityRepo,
		userRepo:         userRepo,
		companyService:   companyService,
		numberSeqService: numberSeqService,
		logger:           logger,
		db:               db,
	}
}

// Create creates a new project with activity logging
func (s *ProjectService) Create(ctx context.Context, req *domain.CreateProjectRequest) (*domain.ProjectDTO, error) {
	// Verify customer exists and get name for denormalized field
	customer, err := s.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}

	// Set default health if not provided
	health := req.Health
	if health == nil {
		defaultHealth := domain.ProjectHealthOnTrack
		health = &defaultHealth
	}

	// Auto-assign manager from company default if not provided
	managerID := req.ManagerID
	if (managerID == nil || *managerID == "") && req.CompanyID != "" && s.companyService != nil {
		defaultManager := s.companyService.GetDefaultProjectResponsible(ctx, req.CompanyID)
		if defaultManager != nil && *defaultManager != "" {
			managerID = defaultManager
			s.logger.Debug("auto-assigned manager from company default",
				zap.String("companyID", string(req.CompanyID)),
				zap.String("managerID", *managerID))
		}
	}

	// Auto-generate project number if not provided and company ID is valid
	projectNumber := req.ProjectNumber
	if projectNumber == "" && req.CompanyID != "" && s.numberSeqService != nil {
		if domain.IsValidCompanyID(string(req.CompanyID)) {
			generatedNumber, err := s.numberSeqService.GenerateProjectNumber(ctx, req.CompanyID)
			if err != nil {
				s.logger.Error("failed to generate project number",
					zap.Error(err),
					zap.String("companyID", string(req.CompanyID)))
				// Don't fail project creation, just log the error
				// Project can still be created without a number
			} else {
				projectNumber = generatedNumber
				s.logger.Info("auto-generated project number",
					zap.String("projectNumber", projectNumber),
					zap.String("companyID", string(req.CompanyID)))
			}
		}
	}

	// Set default phase if not provided
	phase := req.Phase
	if phase == "" {
		phase = domain.ProjectPhaseTilbud
	}

	project := &domain.Project{
		CustomerID:              req.CustomerID,
		CustomerName:            customer.Name,
		Name:                    req.Name,
		ProjectNumber:           projectNumber,
		Summary:                 req.Summary,
		Description:             req.Description,
		Value:                   req.Value,
		Cost:                    req.Cost,
		Spent:                   req.Spent,
		Phase:                   phase,
		EndDate:                 req.EndDate,
		CompanyID:               req.CompanyID,
		ManagerID:               managerID,
		TeamMembers:             req.TeamMembers,
		OfferID:                 req.OfferID,
		DealID:                  req.DealID,
		HasDetailedBudget:       req.HasDetailedBudget,
		Health:                  health,
		CompletionPercent:       req.CompletionPercent,
		EstimatedCompletionDate: req.EstimatedCompletionDate,
		CalculatedOfferValue:    req.Value, // Initialize with value
	}

	// Set StartDate if provided
	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}

	// Set user tracking fields on creation
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.CreatedByID = userCtx.UserID.String()
		project.CreatedByName = userCtx.DisplayName
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Reload with customer relation
	project, err = s.projectRepo.GetByID(ctx, project.ID)
	if err != nil {
		s.logger.Warn("failed to reload project after create", zap.Error(err))
	}

	// Log activity with project number if available
	activityBody := fmt.Sprintf("Project '%s' was created for customer %s", project.Name, project.CustomerName)
	if project.ProjectNumber != "" {
		activityBody = fmt.Sprintf("Project '%s' (%s) was created for customer %s", project.Name, project.ProjectNumber, project.CustomerName)
	}
	s.logActivity(ctx, project.ID, "Project created", activityBody)

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// GetByID retrieves a project by ID
func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// GetByIDWithRelations retrieves a project with all related data
func (s *ProjectService) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, []domain.BudgetItemDTO, []domain.ActivityDTO, error) {
	project, budgetItems, activities, err := s.projectRepo.GetByIDWithRelations(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrProjectNotFound
		}
		return nil, nil, nil, fmt.Errorf("failed to get project with relations: %w", err)
	}

	projectDTO := mapper.ToProjectDTO(project)

	itemDTOs := make([]domain.BudgetItemDTO, len(budgetItems))
	for i, item := range budgetItems {
		itemDTOs[i] = mapper.ToBudgetItemDTO(&item)
	}

	activityDTOs := make([]domain.ActivityDTO, len(activities))
	for i, act := range activities {
		activityDTOs[i] = mapper.ToActivityDTO(&act)
	}

	return &projectDTO, itemDTOs, activityDTOs, nil
}

// GetByIDWithDetails retrieves a project with full details including budget summary
func (s *ProjectService) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.ProjectWithDetailsDTO, error) {
	project, _, activities, err := s.projectRepo.GetByIDWithRelations(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project with relations: %w", err)
	}

	// Get budget summary
	budgetSummary, err := s.GetBudgetSummary(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get budget summary", zap.Error(err))
	}

	dto := mapper.ToProjectWithDetailsDTO(project, budgetSummary, activities, project.Offer, project.Deal)
	return &dto, nil
}

// UpdateHealth updates project health with optional completion percent
func (s *ProjectService) UpdateHealthAndCompletion(ctx context.Context, id uuid.UUID, req *domain.UpdateProjectHealthRequest) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check permissions - must be manager or admin
	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	changes := []string{}

	// Handle health update
	if req.Health != nil {
		oldHealth := "unknown"
		if project.Health != nil {
			oldHealth = string(*project.Health)
		}
		project.Health = req.Health
		changes = append(changes, fmt.Sprintf("health: %s -> %s", oldHealth, *req.Health))
	}

	// Handle completion percent update
	if req.CompletionPercent != nil {
		if *req.CompletionPercent < 0 || *req.CompletionPercent > 100 {
			return nil, ErrInvalidCompletionPercent
		}
		oldPercent := float64(0)
		if project.CompletionPercent != nil {
			oldPercent = *project.CompletionPercent
		}
		project.CompletionPercent = req.CompletionPercent
		changes = append(changes, fmt.Sprintf("completion: %.0f%% -> %.0f%%", oldPercent, *req.CompletionPercent))

		// Auto-transition to completed phase when reaching 100%
		if *req.CompletionPercent == 100 && project.Phase.IsActivePhase() {
			project.Phase = domain.ProjectPhaseCompleted
			changes = append(changes, "phase: auto-completed")
		}
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Recalculate health if not explicitly set
	if req.Health == nil {
		if err := s.projectRepo.UpdateHealth(ctx, id); err != nil {
			s.logger.Warn("failed to recalculate project health", zap.Error(err))
		}
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after update", zap.String("project_id", id.String()), zap.Error(err))
	}

	// Log activity
	if len(changes) > 0 {
		s.logActivity(ctx, project.ID, "Project updated",
			fmt.Sprintf("Project '%s' updated: %s", project.Name, changes))
	}

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// Update updates an existing project with permission check
func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateProjectRequest) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check permissions - must be manager or admin
	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	// Validate economics changes during tilbud phase
	// During tilbud phase, Value/Cost/Spent are read-only (they mirror offer values)
	if project.Phase == domain.ProjectPhaseTilbud {
		if req.Value != project.Value || req.Cost != project.Cost || req.Spent != project.Spent {
			return nil, ErrProjectEconomicsNotEditable
		}
	}

	// Track changes for activity logging
	changes := s.trackChanges(project, req)

	// Track if name changed for denormalized field update
	oldName := project.Name

	// Update fields
	project.Name = req.Name
	project.ProjectNumber = req.ProjectNumber
	project.Summary = req.Summary
	project.Description = req.Description

	// Only update economic fields if project is in editable phase
	if project.Phase.IsEditablePhase() {
		project.Value = req.Value
		project.Cost = req.Cost
		project.Spent = req.Spent
	}

	if req.StartDate != nil {
		project.StartDate = *req.StartDate
	}
	project.EndDate = req.EndDate
	project.CompanyID = req.CompanyID
	project.ManagerID = req.ManagerID
	project.TeamMembers = req.TeamMembers

	// Optional fields
	if req.DealID != nil {
		project.DealID = req.DealID
	}
	if req.HasDetailedBudget != nil {
		project.HasDetailedBudget = *req.HasDetailedBudget
	}
	if req.Health != nil {
		project.Health = req.Health
	}
	if req.CompletionPercent != nil {
		if *req.CompletionPercent < 0 || *req.CompletionPercent > 100 {
			return nil, ErrInvalidCompletionPercent
		}
		project.CompletionPercent = req.CompletionPercent
	}
	if req.EstimatedCompletionDate != nil {
		project.EstimatedCompletionDate = req.EstimatedCompletionDate
	}

	// Set updated by fields (never modify created by)
	if userCtx, ok := auth.FromContext(ctx); ok {
		project.UpdatedByID = userCtx.UserID.String()
		project.UpdatedByName = userCtx.DisplayName
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Update project_name on linked offers if name changed
	if oldName != project.Name && s.offerRepo != nil {
		if err := s.offerRepo.UpdateProjectNameByProjectID(ctx, id, project.Name); err != nil {
			s.logger.Warn("failed to update project_name on linked offers",
				zap.String("project_id", id.String()),
				zap.Error(err))
		}
	}

	// Recalculate health if budget changed
	if changes != "" {
		if err := s.projectRepo.UpdateHealth(ctx, id); err != nil {
			s.logger.Warn("failed to recalculate project health", zap.Error(err))
		}
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after update", zap.String("project_id", id.String()), zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Project '%s' was updated", project.Name)
	if changes != "" {
		activityBody = fmt.Sprintf("Project '%s' was updated: %s", project.Name, changes)
	}
	s.logActivity(ctx, project.ID, "Project updated", activityBody)

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// Delete removes a project with permission check
func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProjectNotFound
		}
		return fmt.Errorf("failed to get project: %w", err)
	}

	// Check permissions - must be manager or admin
	if err := s.checkProjectPermission(ctx, project); err != nil {
		return err
	}

	projectName := project.Name
	customerID := project.CustomerID

	if err := s.projectRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Log activity on customer since project is deleted
	s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, customerID,
		"Project deleted", fmt.Sprintf("Project '%s' was deleted", projectName))

	return nil
}

// List returns a paginated list of projects with optional filters and default sorting
func (s *ProjectService) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID, phase *domain.ProjectPhase) (*domain.PaginatedResponse, error) {
	filters := &repository.ProjectFilters{
		CustomerID: customerID,
		Phase:      phase,
	}
	return s.ListWithSort(ctx, page, pageSize, filters, repository.DefaultSortConfig())
}

// ListWithFilters returns a paginated list of projects with filter options and default sorting
func (s *ProjectService) ListWithFilters(ctx context.Context, page, pageSize int, filters *repository.ProjectFilters) (*domain.PaginatedResponse, error) {
	return s.ListWithSort(ctx, page, pageSize, filters, repository.DefaultSortConfig())
}

// ListWithSort returns a paginated list of projects with filter and sort options
func (s *ProjectService) ListWithSort(ctx context.Context, page, pageSize int, filters *repository.ProjectFilters, sort repository.SortConfig) (*domain.PaginatedResponse, error) {
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

	projects, total, err := s.projectRepo.ListWithFilters(ctx, page, pageSize, filters, sort)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	dtos := make([]domain.ProjectDTO, len(projects))
	for i, project := range projects {
		dtos[i] = mapper.ToProjectDTO(&project)
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

// ============================================================================
// Budget Inheritance Methods
// ============================================================================

// InheritBudgetFromOffer clones budget items from an offer to the project.
// This is typically called when a project is created from a won offer.
// Returns the updated project and the count of items cloned.
func (s *ProjectService) InheritBudgetFromOffer(ctx context.Context, projectID, offerID uuid.UUID) (*domain.InheritBudgetResponse, error) {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check permissions - must be manager or admin
	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	// Verify offer exists and is won
	if s.offerRepo == nil {
		return nil, fmt.Errorf("offer repository not available")
	}
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if offer.Phase != domain.OfferPhaseWon {
		return nil, ErrOfferNotWon
	}

	// Get budget items from offer
	if s.budgetItemRepo == nil {
		return nil, fmt.Errorf("budget item repository not available")
	}
	items, err := s.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer budget items: %w", err)
	}

	itemsCount := len(items)

	if itemsCount == 0 {
		s.logger.Info("no budget items to inherit from offer",
			zap.String("projectID", projectID.String()),
			zap.String("offerID", offerID.String()))

		// Still update project to link to offer even if no items
		project.OfferID = &offerID
		project.Value = offer.Value
		project.Cost = offer.Cost
		if err := s.projectRepo.Update(ctx, project); err != nil {
			return nil, fmt.Errorf("failed to update project: %w", err)
		}

		// Reload project
		project, err = s.projectRepo.GetByID(ctx, projectID)
		if err != nil {
			s.logger.Warn("failed to reload project after budget inheritance", zap.Error(err))
		}
		dto := mapper.ToProjectDTO(project)
		return &domain.InheritBudgetResponse{
			Project:    &dto,
			ItemsCount: 0,
		}, nil
	}

	// Use transaction for atomicity
	if s.db == nil {
		return nil, fmt.Errorf("database connection not available for transaction")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Clone each budget item
		for _, item := range items {
			cloned := domain.BudgetItem{
				ParentType:     domain.BudgetParentProject,
				ParentID:       projectID,
				Name:           item.Name,
				ExpectedCost:   item.ExpectedCost,
				ExpectedMargin: item.ExpectedMargin,
				Quantity:       item.Quantity,
				PricePerItem:   item.PricePerItem,
				Description:    item.Description,
				DisplayOrder:   item.DisplayOrder,
			}
			if err := tx.Create(&cloned).Error; err != nil {
				return fmt.Errorf("failed to clone budget item: %w", err)
			}
		}

		// Update project to reflect detailed budget and link to offer
		project.HasDetailedBudget = true
		project.OfferID = &offerID
		project.Value = offer.Value
		project.Cost = offer.Cost
		if err := tx.Save(project).Error; err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Log activity
	s.logActivity(ctx, projectID, "Budget inherited from offer",
		fmt.Sprintf("Budget items (%d items) inherited from offer '%s'", itemsCount, offer.Title))

	// Reload project to get updated data
	project, err = s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		s.logger.Warn("failed to reload project after budget inheritance", zap.Error(err))
	}
	dto := mapper.ToProjectDTO(project)

	return &domain.InheritBudgetResponse{
		Project:    &dto,
		ItemsCount: itemsCount,
	}, nil
}

// GetBudgetSummary returns aggregated budget totals for a project
func (s *ProjectService) GetBudgetSummary(ctx context.Context, id uuid.UUID) (*domain.BudgetSummaryDTO, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	summary, err := s.projectRepo.GetBudgetSummary(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
	}

	dto := &domain.BudgetSummaryDTO{
		ParentType:    domain.BudgetParentProject,
		ParentID:      id,
		ItemCount:     summary.ItemCount,
		TotalCost:     summary.TotalCost,
		TotalRevenue:  summary.TotalRevenue,
		TotalProfit:   summary.TotalProfit,
		MarginPercent: summary.MarginPercent,
	}

	return dto, nil
}

// GetBudget returns budget information for a project
func (s *ProjectService) GetBudget(ctx context.Context, id uuid.UUID) (*domain.ProjectBudgetDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	dto := mapper.ToProjectBudgetDTO(project)
	return &dto, nil
}

// GetBudgetMetrics calculates detailed budget metrics for a project
func (s *ProjectService) GetBudgetMetrics(ctx context.Context, id uuid.UUID) (*repository.ProjectBudgetMetrics, error) {
	metrics, err := s.projectRepo.CalculateBudgetMetrics(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to calculate budget metrics: %w", err)
	}
	return metrics, nil
}

// ============================================================================
// Lifecycle Methods
// ============================================================================

// UpdateCompletionPercent updates the project completion percentage with validation
func (s *ProjectService) UpdateCompletionPercent(ctx context.Context, id uuid.UUID, percent float64) (*domain.ProjectDTO, error) {
	// Validate range
	if percent < 0 || percent > 100 {
		return nil, ErrInvalidCompletionPercent
	}

	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check permissions
	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	oldPercent := float64(0)
	if project.CompletionPercent != nil {
		oldPercent = *project.CompletionPercent
	}

	project.CompletionPercent = &percent

	// Auto-transition to completed phase when reaching 100%
	if percent == 100 && project.Phase.IsActivePhase() {
		project.Phase = domain.ProjectPhaseCompleted
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update completion percent: %w", err)
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after completion update", zap.String("project_id", id.String()), zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, project.ID, "Project progress updated",
		fmt.Sprintf("Project '%s' completion updated from %.1f%% to %.1f%%", project.Name, oldPercent, percent))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// RecalculateHealth recalculates and updates the project health status
func (s *ProjectService) RecalculateHealth(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, error) {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	oldHealth := project.Health

	// Recalculate health
	if err := s.projectRepo.UpdateHealth(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to recalculate health: %w", err)
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after health recalculation", zap.String("project_id", id.String()), zap.Error(err))
	}

	// Log activity if health changed
	if oldHealth == nil || *oldHealth != *project.Health {
		oldHealthStr := "unknown"
		if oldHealth != nil {
			oldHealthStr = string(*oldHealth)
		}
		s.logActivity(ctx, project.ID, "Project health updated",
			fmt.Sprintf("Project '%s' health changed from %s to %s", project.Name, oldHealthStr, *project.Health))
	}

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// ============================================================================
// Activity Methods
// ============================================================================

// GetActivities returns activities for a project
func (s *ProjectService) GetActivities(ctx context.Context, id uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	activities, err := s.activityRepo.ListByTarget(ctx, domain.ActivityTargetProject, id, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}

// ============================================================================
// Query Methods
// ============================================================================

// GetByManager returns all projects managed by a specific user
func (s *ProjectService) GetByManager(ctx context.Context, managerID string) ([]domain.ProjectDTO, error) {
	projects, err := s.projectRepo.GetByManager(ctx, managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects by manager: %w", err)
	}

	dtos := make([]domain.ProjectDTO, len(projects))
	for i, project := range projects {
		dtos[i] = mapper.ToProjectDTO(&project)
	}

	return dtos, nil
}

// GetByHealth returns all projects with a specific health status
func (s *ProjectService) GetByHealth(ctx context.Context, health domain.ProjectHealth) ([]domain.ProjectDTO, error) {
	projects, err := s.projectRepo.GetByHealth(ctx, health)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects by health: %w", err)
	}

	dtos := make([]domain.ProjectDTO, len(projects))
	for i, project := range projects {
		dtos[i] = mapper.ToProjectDTO(&project)
	}

	return dtos, nil
}

// GetHealthSummary returns project counts grouped by health status
func (s *ProjectService) GetHealthSummary(ctx context.Context) (map[domain.ProjectHealth]int64, error) {
	return s.projectRepo.CountByHealth(ctx)
}

// ============================================================================
// Helper Methods
// ============================================================================

// checkProjectPermission verifies the user has permission to modify the project
// Users must be the project manager or have admin role
func (s *ProjectService) checkProjectPermission(ctx context.Context, project *domain.Project) error {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	// Check if user is admin
	for _, role := range userCtx.Roles {
		if role == domain.RoleSuperAdmin || role == domain.RoleCompanyAdmin || role == domain.RoleManager {
			return nil
		}
	}

	// Check if user is the project manager
	if project.ManagerID != nil && *project.ManagerID == userCtx.UserID.String() {
		return nil
	}

	return ErrProjectNotManager
}

// trackChanges creates a summary of changes between the project and update request
func (s *ProjectService) trackChanges(project *domain.Project, req *domain.UpdateProjectRequest) string {
	var changes []string

	if project.Name != req.Name {
		changes = append(changes, fmt.Sprintf("name: '%s' -> '%s'", project.Name, req.Name))
	}
	if project.Value != req.Value {
		changes = append(changes, fmt.Sprintf("value: %.2f -> %.2f", project.Value, req.Value))
	}
	if project.Cost != req.Cost {
		changes = append(changes, fmt.Sprintf("cost: %.2f -> %.2f", project.Cost, req.Cost))
	}
	if project.Spent != req.Spent {
		changes = append(changes, fmt.Sprintf("spent: %.2f -> %.2f", project.Spent, req.Spent))
	}
	// Compare ManagerID pointers
	oldManager := ""
	newManager := ""
	if project.ManagerID != nil {
		oldManager = *project.ManagerID
	}
	if req.ManagerID != nil {
		newManager = *req.ManagerID
	}
	if oldManager != newManager {
		changes = append(changes, fmt.Sprintf("manager: %s -> %s", oldManager, newManager))
	}

	if len(changes) == 0 {
		return ""
	}
	if len(changes) > 3 {
		return fmt.Sprintf("%d fields updated", len(changes))
	}
	result := ""
	for i, c := range changes {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

// logActivity creates an activity log entry for a project
func (s *ProjectService) logActivity(ctx context.Context, projectID uuid.UUID, title, body string) {
	s.logActivityOnTarget(ctx, domain.ActivityTargetProject, projectID, title, body)
}

// logActivityOnTarget creates an activity log entry for any target
func (s *ProjectService) logActivityOnTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		s.logger.Warn("no user context for activity logging")
		return
	}

	activity := &domain.Activity{
		TargetType:  targetType,
		TargetID:    targetID,
		Title:       title,
		Body:        body,
		OccurredAt:  time.Now(),
		CreatorName: userCtx.DisplayName,
		CreatorID:   userCtx.UserID.String(),
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log activity", zap.Error(err))
	}
}

// ============================================================================
// Individual Property Update Methods
// ============================================================================

// UpdateName updates only the project name
func (s *ProjectService) UpdateName(ctx context.Context, id uuid.UUID, name string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	oldName := project.Name
	project.Name = name

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project name: %w", err)
	}

	// Update project_name on linked offers
	if oldName != name && s.offerRepo != nil {
		if err := s.offerRepo.UpdateProjectNameByProjectID(ctx, id, name); err != nil {
			s.logger.Warn("failed to update project_name on linked offers",
				zap.String("project_id", id.String()),
				zap.Error(err))
		}
	}

	s.logActivity(ctx, project.ID, "Project name updated", fmt.Sprintf("Project name changed from '%s' to '%s'", oldName, name))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateDescription updates only the project description and summary
func (s *ProjectService) UpdateDescription(ctx context.Context, id uuid.UUID, summary, description string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	project.Summary = summary
	project.Description = description

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project description: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project description updated", "Project description was updated")

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdatePhase updates only the project phase
// Validates phase transitions and enforces business rules:
// - Working phase requires StartDate
// - Project cannot be tilbud if it has a WinningOfferID
func (s *ProjectService) UpdatePhase(ctx context.Context, id uuid.UUID, phase domain.ProjectPhase) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	// Validate phase transition
	if !project.Phase.CanTransitionTo(phase) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidPhaseTransition, project.Phase, phase)
	}

	// Working phase requires StartDate
	if phase == domain.ProjectPhaseWorking && project.StartDate.IsZero() {
		return nil, ErrWorkingPhaseRequiresStartDate
	}

	// Project cannot be tilbud if it has a WinningOfferID
	if phase == domain.ProjectPhaseTilbud && project.WinningOfferID != nil {
		return nil, ErrTilbudPhaseCannotHaveWinningOffer
	}

	oldPhase := project.Phase
	project.Phase = phase

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project phase: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project phase updated", fmt.Sprintf("Project phase changed from '%s' to '%s'", oldPhase, phase))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// ReopenProject reopens a completed or cancelled project
// Business rules:
// - Completed projects can be reopened to working (winning offer reverts to sent)
// - Cancelled projects can be reopened to tilbud (WinningOfferID cleared, winning offer reverts to sent if exists)
// - Working phase requires StartDate (auto-set to now if not present)
func (s *ProjectService) ReopenProject(ctx context.Context, id uuid.UUID, req *domain.ReopenProjectRequest) (*domain.ReopenProjectResponse, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	// Validate source phase - only completed or cancelled can be reopened
	if !project.Phase.IsClosedPhase() {
		return nil, fmt.Errorf("%w: project is in %s phase, only completed or cancelled projects can be reopened",
			ErrCannotReopenProject, project.Phase)
	}

	// Validate target phase transition
	if !project.Phase.CanTransitionTo(req.TargetPhase) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s",
			ErrInvalidPhaseTransition, project.Phase, req.TargetPhase)
	}

	previousPhase := project.Phase
	response := &domain.ReopenProjectResponse{
		PreviousPhase: string(previousPhase),
	}

	// Handle winning offer if exists
	var revertedOffer *domain.Offer
	if project.WinningOfferID != nil {
		// Revert the winning offer to sent phase
		offer, err := s.offerRepo.GetByID(ctx, *project.WinningOfferID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("failed to get winning offer: %w", err)
			}
			// Offer not found - just clear the ID
			s.logger.Warn("winning offer not found during reopen, clearing reference",
				zap.String("projectID", id.String()),
				zap.String("winningOfferID", project.WinningOfferID.String()))
		} else if offer.Phase == domain.OfferPhaseWon {
			// Revert to sent
			offer.Phase = domain.OfferPhaseSent
			if err := s.offerRepo.Update(ctx, offer); err != nil {
				return nil, fmt.Errorf("failed to revert winning offer to sent: %w", err)
			}
			revertedOffer = offer
			s.logger.Info("reverted winning offer to sent",
				zap.String("projectID", id.String()),
				zap.String("offerID", offer.ID.String()))

			// Log activity on offer
			s.logActivityOnTarget(ctx, domain.ActivityTargetOffer, offer.ID,
				"Offer reverted to sent",
				fmt.Sprintf("Offer '%s' was reverted from won to sent (project '%s' reopened)", offer.Title, project.Name))
		}

		// Clear WinningOfferID if going to tilbud or working
		if req.TargetPhase == domain.ProjectPhaseTilbud || req.TargetPhase == domain.ProjectPhaseWorking {
			project.WinningOfferID = nil
			project.InheritedOfferNumber = ""
			project.WonAt = nil
			response.ClearedOfferID = true
			response.ClearedOfferValue = true
		}
	}

	// Set StartDate if going to working and not already set
	if req.TargetPhase == domain.ProjectPhaseWorking && project.StartDate.IsZero() {
		project.StartDate = time.Now()
		s.logger.Info("auto-set StartDate for working phase",
			zap.String("projectID", id.String()),
			zap.Time("startDate", project.StartDate))
	}

	// Update phase
	project.Phase = req.TargetPhase

	// Save project
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Reload project
	project, err = s.projectRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload project after reopen", zap.Error(err))
	}

	// Log activity
	activityBody := fmt.Sprintf("Project reopened from '%s' to '%s'", previousPhase, req.TargetPhase)
	if req.Notes != "" {
		activityBody = fmt.Sprintf("%s. Notes: %s", activityBody, req.Notes)
	}
	if response.ClearedOfferID {
		activityBody = fmt.Sprintf("%s. Winning offer reference cleared.", activityBody)
	}
	if revertedOffer != nil {
		activityBody = fmt.Sprintf("%s. Winning offer '%s' reverted to sent.", activityBody, revertedOffer.Title)
	}
	s.logActivity(ctx, project.ID, "Project reopened", activityBody)

	// Build response
	projectDTO := mapper.ToProjectDTO(project)
	response.Project = &projectDTO

	if revertedOffer != nil {
		offerDTO := mapper.ToOfferDTO(revertedOffer)
		response.RevertedOffer = &offerDTO
	}

	return response, nil
}

// UpdateManager updates only the project manager
func (s *ProjectService) UpdateManager(ctx context.Context, id uuid.UUID, managerID, managerName string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Permission check removed - anyone can update project manager

	// If managerName not provided, try to look it up from the user repository
	if managerName == "" && managerID != "" && s.userRepo != nil {
		user, err := s.userRepo.GetByStringID(ctx, managerID)
		if err == nil && user != nil {
			managerName = user.DisplayName
		} else {
			s.logger.Debug("could not look up manager name", zap.String("managerID", managerID), zap.Error(err))
		}
	}

	oldManager := project.ManagerName
	if oldManager == "" && project.ManagerID != nil {
		oldManager = *project.ManagerID
	}
	project.ManagerID = &managerID
	project.ManagerName = managerName

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project manager: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project manager updated", fmt.Sprintf("Project manager changed from '%s' to '%s'", oldManager, managerName))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateDates updates only the project start and end dates
func (s *ProjectService) UpdateDates(ctx context.Context, id uuid.UUID, startDate, endDate *time.Time) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	if startDate != nil {
		project.StartDate = *startDate
	}
	// EndDate is a pointer in the model, so we can assign directly
	project.EndDate = endDate

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project dates: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project dates updated", "Project start/end dates were updated")

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateBudget updates only the project value (only allowed in active phase)
// Note: This is named UpdateBudget for backwards compatibility but actually updates Value
func (s *ProjectService) UpdateBudget(ctx context.Context, id uuid.UUID, value float64) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	// Value is read-only during tilbud phase
	if project.Phase == domain.ProjectPhaseTilbud {
		return nil, ErrProjectEconomicsNotEditable
	}

	oldValue := project.Value
	project.Value = value

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project value: %w", err)
	}

	// Recalculate health after value change
	if err := s.projectRepo.UpdateHealth(ctx, id); err != nil {
		s.logger.Warn("failed to recalculate project health", zap.Error(err))
	}

	// Reload project to get updated health
	project, _ = s.projectRepo.GetByID(ctx, id)

	s.logActivity(ctx, project.ID, "Project value updated", fmt.Sprintf("Project value changed from %.2f to %.2f", oldValue, value))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateSpent updates only the project spent amount (only allowed in active phase)
func (s *ProjectService) UpdateSpent(ctx context.Context, id uuid.UUID, spent float64) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	// Spent is read-only during tilbud phase
	if project.Phase == domain.ProjectPhaseTilbud {
		return nil, ErrProjectEconomicsNotEditable
	}

	oldSpent := project.Spent
	project.Spent = spent

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project spent: %w", err)
	}

	// Recalculate health after spent change
	if err := s.projectRepo.UpdateHealth(ctx, id); err != nil {
		s.logger.Warn("failed to recalculate project health", zap.Error(err))
	}

	// Reload project to get updated health
	project, _ = s.projectRepo.GetByID(ctx, id)

	s.logActivity(ctx, project.ID, "Project spent updated", fmt.Sprintf("Project spent changed from %.2f to %.2f", oldSpent, spent))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateTeamMembers updates only the project team members
func (s *ProjectService) UpdateTeamMembers(ctx context.Context, id uuid.UUID, teamMembers []string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	project.TeamMembers = teamMembers

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project team members: %w", err)
	}

	s.logActivity(ctx, project.ID, "Team members updated", fmt.Sprintf("Project team members updated (%d members)", len(teamMembers)))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateHealth updates only the project health
func (s *ProjectService) UpdateHealth(ctx context.Context, id uuid.UUID, health domain.ProjectHealth) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	oldHealth := "unknown"
	if project.Health != nil {
		oldHealth = string(*project.Health)
	}
	project.Health = &health

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project health: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project health updated", fmt.Sprintf("Project health changed from '%s' to '%s'", oldHealth, health))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateEstimatedCompletionDate updates only the estimated completion date
func (s *ProjectService) UpdateEstimatedCompletionDate(ctx context.Context, id uuid.UUID, estimatedDate *time.Time) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	project.EstimatedCompletionDate = estimatedDate

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update estimated completion date: %w", err)
	}

	activityMsg := "Estimated completion date cleared"
	if estimatedDate != nil {
		activityMsg = fmt.Sprintf("Estimated completion date set to %s", estimatedDate.Format("2006-01-02"))
	}
	s.logActivity(ctx, project.ID, "Estimated completion date updated", activityMsg)

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateProjectNumber updates only the project number
func (s *ProjectService) UpdateProjectNumber(ctx context.Context, id uuid.UUID, projectNumber string) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	oldNumber := project.ProjectNumber
	project.ProjectNumber = projectNumber

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project number: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project number updated", fmt.Sprintf("Project number changed from '%s' to '%s'", oldNumber, projectNumber))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// UpdateCompanyID updates only the project company assignment
func (s *ProjectService) UpdateCompanyID(ctx context.Context, id uuid.UUID, companyID domain.CompanyID) (*domain.ProjectDTO, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, err
	}

	oldCompany := project.CompanyID
	project.CompanyID = companyID

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project company: %w", err)
	}

	s.logActivity(ctx, project.ID, "Project company updated", fmt.Sprintf("Project company changed from '%s' to '%s'", oldCompany, companyID))

	dto := mapper.ToProjectDTO(project)
	return &dto, nil
}

// ResyncFromBestOffer syncs project economics (value, cost) from the best connected offer.
// This is useful when:
// - Offer data has been updated and you want the project to reflect those changes
// - Project numbers have been changed and you want to reset from offer data
// Returns the updated project and the offer that was used for syncing.
func (s *ProjectService) ResyncFromBestOffer(ctx context.Context, id uuid.UUID) (*domain.ProjectDTO, *domain.Offer, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrProjectNotFound
		}
		return nil, nil, fmt.Errorf("failed to get project: %w", err)
	}

	if err := s.checkProjectPermission(ctx, project); err != nil {
		return nil, nil, err
	}

	// Get best offer for this project
	if s.offerRepo == nil {
		return nil, nil, fmt.Errorf("offer repository not available")
	}

	bestOffer, err := s.offerRepo.GetBestOfferForProject(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get best offer: %w", err)
	}

	if bestOffer == nil {
		return nil, nil, fmt.Errorf("no offers found for project")
	}

	// Store old values for activity logging
	oldValue := project.Value
	oldCost := project.Cost

	// Sync economics from offer
	project.Value = bestOffer.Value
	project.Cost = bestOffer.Cost
	project.CalculatedOfferValue = bestOffer.Value

	// If offer has project number, sync that too (for won offers)
	if bestOffer.Phase == domain.OfferPhaseWon && bestOffer.ExternalReference != "" {
		project.InheritedOfferNumber = bestOffer.ExternalReference
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Recalculate health after economic changes
	if err := s.projectRepo.UpdateHealth(ctx, id); err != nil {
		s.logger.Warn("failed to recalculate project health", zap.Error(err))
	}

	// Reload project
	project, _ = s.projectRepo.GetByID(ctx, id)

	s.logActivity(ctx, project.ID, "Project synced from offer",
		fmt.Sprintf("Project economics synced from offer '%s': value %.2f -> %.2f, cost %.2f -> %.2f",
			bestOffer.Title, oldValue, project.Value, oldCost, project.Cost))

	dto := mapper.ToProjectDTO(project)
	return &dto, bestOffer, nil
}
