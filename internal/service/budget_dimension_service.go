package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Service-level errors for budget dimensions
var (
	ErrBudgetDimensionNotFound  = errors.New("budget dimension not found")
	ErrInvalidParentType        = errors.New("invalid parent type")
	ErrParentNotFound           = errors.New("parent entity not found")
	ErrInvalidCost              = errors.New("cost must be greater than 0")
	ErrInvalidRevenue           = errors.New("revenue must be greater than or equal to 0")
	ErrInvalidTargetMargin      = errors.New("target margin percent must be between 0 and 100")
	ErrInvalidCategory          = errors.New("category not found or inactive")
	ErrMissingName              = errors.New("either categoryId or customName must be provided")
	ErrSourceDimensionsNotFound = errors.New("source has no budget dimensions to clone")
	ErrReorderCountMismatch     = errors.New("ordered IDs count does not match existing dimensions count")
)

// BudgetDimensionService handles business logic for budget dimensions
type BudgetDimensionService struct {
	dimensionRepo *repository.BudgetDimensionRepository
	categoryRepo  *repository.BudgetDimensionCategoryRepository
	offerRepo     *repository.OfferRepository
	projectRepo   *repository.ProjectRepository
	activityRepo  *repository.ActivityRepository
	logger        *zap.Logger
}

// NewBudgetDimensionService creates a new BudgetDimensionService instance
func NewBudgetDimensionService(
	dimensionRepo *repository.BudgetDimensionRepository,
	categoryRepo *repository.BudgetDimensionCategoryRepository,
	offerRepo *repository.OfferRepository,
	projectRepo *repository.ProjectRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *BudgetDimensionService {
	return &BudgetDimensionService{
		dimensionRepo: dimensionRepo,
		categoryRepo:  categoryRepo,
		offerRepo:     offerRepo,
		projectRepo:   projectRepo,
		activityRepo:  activityRepo,
		logger:        logger,
	}
}

// Create creates a new budget dimension
func (s *BudgetDimensionService) Create(ctx context.Context, req *domain.CreateBudgetDimensionRequest) (*domain.BudgetDimensionDTO, error) {
	// Validate parent exists
	if err := s.validateParentExists(ctx, req.ParentType, req.ParentID); err != nil {
		return nil, err
	}

	// Validate name source
	if req.CategoryID == nil && req.CustomName == "" {
		return nil, ErrMissingName
	}

	// Validate category if provided
	if req.CategoryID != nil {
		if err := s.validateCategory(ctx, *req.CategoryID); err != nil {
			return nil, err
		}
	}

	// Validate cost/revenue/margin
	if err := s.validateFinancials(req.Cost, req.Revenue, req.TargetMarginPercent, req.MarginOverride); err != nil {
		return nil, err
	}

	// Set display order if not provided
	displayOrder := req.DisplayOrder
	if displayOrder == 0 {
		maxOrder, err := s.dimensionRepo.GetMaxDisplayOrder(ctx, req.ParentType, req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get max display order: %w", err)
		}
		displayOrder = maxOrder + 1
	}

	// Create dimension
	dimension := &domain.BudgetDimension{
		ParentType:          req.ParentType,
		ParentID:            req.ParentID,
		CategoryID:          req.CategoryID,
		CustomName:          req.CustomName,
		Cost:                req.Cost,
		Revenue:             req.Revenue,
		TargetMarginPercent: req.TargetMarginPercent,
		MarginOverride:      req.MarginOverride,
		Description:         req.Description,
		Quantity:            req.Quantity,
		Unit:                req.Unit,
		DisplayOrder:        displayOrder,
	}

	if err := s.dimensionRepo.Create(ctx, dimension); err != nil {
		return nil, fmt.Errorf("failed to create budget dimension: %w", err)
	}

	// Reload with category preloaded
	dimension, err := s.dimensionRepo.GetByID(ctx, dimension.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload dimension: %w", err)
	}

	// Update parent totals
	if err := s.updateParentTotals(ctx, req.ParentType, req.ParentID); err != nil {
		s.logger.Warn("failed to update parent totals", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, s.getActivityTargetType(req.ParentType), req.ParentID,
		"Budget dimension added",
		fmt.Sprintf("Added budget line: %s (Cost: %.2f, Revenue: %.2f)", dimension.GetName(), dimension.Cost, dimension.Revenue))

	dto := mapper.ToBudgetDimensionDTO(dimension)
	return &dto, nil
}

// GetByID retrieves a budget dimension by ID
func (s *BudgetDimensionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetDimensionDTO, error) {
	dimension, err := s.dimensionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBudgetDimensionNotFound
		}
		return nil, fmt.Errorf("failed to get budget dimension: %w", err)
	}

	dto := mapper.ToBudgetDimensionDTO(dimension)
	return &dto, nil
}

// Update updates an existing budget dimension
func (s *BudgetDimensionService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateBudgetDimensionRequest) (*domain.BudgetDimensionDTO, error) {
	// Get existing dimension
	dimension, err := s.dimensionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBudgetDimensionNotFound
		}
		return nil, fmt.Errorf("failed to get budget dimension: %w", err)
	}

	// Validate name source
	if req.CategoryID == nil && req.CustomName == "" {
		return nil, ErrMissingName
	}

	// Validate category if provided
	if req.CategoryID != nil {
		if err := s.validateCategory(ctx, *req.CategoryID); err != nil {
			return nil, err
		}
	}

	// Validate financials
	if err := s.validateFinancials(req.Cost, req.Revenue, req.TargetMarginPercent, req.MarginOverride); err != nil {
		return nil, err
	}

	// Update fields
	dimension.CategoryID = req.CategoryID
	dimension.CustomName = req.CustomName
	dimension.Cost = req.Cost
	dimension.Revenue = req.Revenue
	dimension.TargetMarginPercent = req.TargetMarginPercent
	dimension.MarginOverride = req.MarginOverride
	dimension.Description = req.Description
	dimension.Quantity = req.Quantity
	dimension.Unit = req.Unit
	dimension.DisplayOrder = req.DisplayOrder

	if err := s.dimensionRepo.Update(ctx, dimension); err != nil {
		return nil, fmt.Errorf("failed to update budget dimension: %w", err)
	}

	// Reload to get computed fields
	dimension, err = s.dimensionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload dimension: %w", err)
	}

	// Update parent totals
	if err := s.updateParentTotals(ctx, dimension.ParentType, dimension.ParentID); err != nil {
		s.logger.Warn("failed to update parent totals", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, s.getActivityTargetType(dimension.ParentType), dimension.ParentID,
		"Budget dimension updated",
		fmt.Sprintf("Updated budget line: %s (Cost: %.2f, Revenue: %.2f)", dimension.GetName(), dimension.Cost, dimension.Revenue))

	dto := mapper.ToBudgetDimensionDTO(dimension)
	return &dto, nil
}

// Delete removes a budget dimension
func (s *BudgetDimensionService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get dimension for activity logging
	dimension, err := s.dimensionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBudgetDimensionNotFound
		}
		return fmt.Errorf("failed to get budget dimension: %w", err)
	}

	parentType := dimension.ParentType
	parentID := dimension.ParentID
	name := dimension.GetName()

	// Delete
	if err := s.dimensionRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete budget dimension: %w", err)
	}

	// Update parent totals
	if err := s.updateParentTotals(ctx, parentType, parentID); err != nil {
		s.logger.Warn("failed to update parent totals", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, s.getActivityTargetType(parentType), parentID,
		"Budget dimension deleted",
		fmt.Sprintf("Deleted budget line: %s", name))

	return nil
}

// ListByParent retrieves all budget dimensions for a parent entity
func (s *BudgetDimensionService) ListByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) ([]domain.BudgetDimensionDTO, error) {
	dimensions, err := s.dimensionRepo.GetByParent(ctx, parentType, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list budget dimensions: %w", err)
	}

	dtos := make([]domain.BudgetDimensionDTO, len(dimensions))
	for i, dim := range dimensions {
		dtos[i] = mapper.ToBudgetDimensionDTO(&dim)
	}

	return dtos, nil
}

// ListByParentPaginated retrieves paginated budget dimensions for a parent entity
func (s *BudgetDimensionService) ListByParentPaginated(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, page, pageSize int) (*domain.PaginatedResponse, error) {
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

	dimensions, total, err := s.dimensionRepo.GetByParentPaginated(ctx, parentType, parentID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list budget dimensions: %w", err)
	}

	dtos := make([]domain.BudgetDimensionDTO, len(dimensions))
	for i, dim := range dimensions {
		dtos[i] = mapper.ToBudgetDimensionDTO(&dim)
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

// GetSummary calculates and returns budget totals for a parent entity (AC3: CalculateTotals)
func (s *BudgetDimensionService) GetSummary(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (*domain.BudgetSummaryDTO, error) {
	summary, err := s.dimensionRepo.GetBudgetSummary(ctx, parentType, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
	}

	dto := &domain.BudgetSummaryDTO{
		ParentType:           parentType,
		ParentID:             parentID,
		DimensionCount:       summary.DimensionCount,
		TotalCost:            summary.TotalCost,
		TotalRevenue:         summary.TotalRevenue,
		OverallMarginPercent: summary.MarginPercent,
		TotalProfit:          summary.TotalMargin,
	}

	return dto, nil
}

// ReorderDimensions updates the display order of dimensions (AC5)
func (s *BudgetDimensionService) ReorderDimensions(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, orderedIDs []uuid.UUID) error {
	// Validate parent exists
	if err := s.validateParentExists(ctx, parentType, parentID); err != nil {
		return err
	}

	// Validate count matches
	count, err := s.dimensionRepo.Count(ctx, parentType, parentID)
	if err != nil {
		return fmt.Errorf("failed to count dimensions: %w", err)
	}

	if len(orderedIDs) != count {
		return fmt.Errorf("%w: got %d, expected %d", ErrReorderCountMismatch, len(orderedIDs), count)
	}

	// Perform reorder
	if err := s.dimensionRepo.ReorderDimensions(ctx, parentType, parentID, orderedIDs); err != nil {
		return fmt.Errorf("failed to reorder dimensions: %w", err)
	}

	// Log activity
	s.logActivity(ctx, s.getActivityTargetType(parentType), parentID,
		"Budget dimensions reordered",
		fmt.Sprintf("Reordered %d budget lines", len(orderedIDs)))

	return nil
}

// CloneDimensions copies all budget dimensions from source to target (AC4)
// This is typically used when an offer is won and converted to a project
func (s *BudgetDimensionService) CloneDimensions(ctx context.Context, sourceType domain.BudgetParentType, sourceID uuid.UUID, targetType domain.BudgetParentType, targetID uuid.UUID) ([]domain.BudgetDimensionDTO, error) {
	// Validate source exists
	if err := s.validateParentExists(ctx, sourceType, sourceID); err != nil {
		return nil, fmt.Errorf("source validation failed: %w", err)
	}

	// Validate target exists
	if err := s.validateParentExists(ctx, targetType, targetID); err != nil {
		return nil, fmt.Errorf("target validation failed: %w", err)
	}

	// Get source dimensions
	sourceDimensions, err := s.dimensionRepo.GetByParent(ctx, sourceType, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source dimensions: %w", err)
	}

	if len(sourceDimensions) == 0 {
		return nil, ErrSourceDimensionsNotFound
	}

	// Clone each dimension
	clonedDTOs := make([]domain.BudgetDimensionDTO, 0, len(sourceDimensions))
	for _, src := range sourceDimensions {
		cloned := &domain.BudgetDimension{
			ParentType:          targetType,
			ParentID:            targetID,
			CategoryID:          src.CategoryID,
			CustomName:          src.CustomName,
			Cost:                src.Cost,
			Revenue:             src.Revenue,
			TargetMarginPercent: src.TargetMarginPercent,
			MarginOverride:      src.MarginOverride,
			Description:         src.Description,
			Quantity:            src.Quantity,
			Unit:                src.Unit,
			DisplayOrder:        src.DisplayOrder,
		}

		if err := s.dimensionRepo.Create(ctx, cloned); err != nil {
			return nil, fmt.Errorf("failed to clone dimension: %w", err)
		}

		// Reload to get computed fields and category
		cloned, err = s.dimensionRepo.GetByID(ctx, cloned.ID)
		if err != nil {
			s.logger.Warn("failed to reload cloned dimension", zap.Error(err))
			continue
		}

		clonedDTOs = append(clonedDTOs, mapper.ToBudgetDimensionDTO(cloned))
	}

	// Update target totals
	if err := s.updateParentTotals(ctx, targetType, targetID); err != nil {
		s.logger.Warn("failed to update target totals", zap.Error(err))
	}

	// Log activity on target
	s.logActivity(ctx, s.getActivityTargetType(targetType), targetID,
		"Budget dimensions cloned",
		fmt.Sprintf("Cloned %d budget lines from %s", len(clonedDTOs), sourceType))

	return clonedDTOs, nil
}

// DeleteByParent removes all budget dimensions for a parent entity
func (s *BudgetDimensionService) DeleteByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	// Get count for activity log
	count, err := s.dimensionRepo.Count(ctx, parentType, parentID)
	if err != nil {
		return fmt.Errorf("failed to count dimensions: %w", err)
	}

	if count == 0 {
		return nil // Nothing to delete
	}

	// Delete all
	if err := s.dimensionRepo.DeleteByParent(ctx, parentType, parentID); err != nil {
		return fmt.Errorf("failed to delete dimensions: %w", err)
	}

	// Update parent totals (will set to 0)
	if err := s.updateParentTotals(ctx, parentType, parentID); err != nil {
		s.logger.Warn("failed to update parent totals", zap.Error(err))
	}

	// Log activity
	s.logActivity(ctx, s.getActivityTargetType(parentType), parentID,
		"Budget dimensions cleared",
		fmt.Sprintf("Removed all %d budget lines", count))

	return nil
}

// validateFinancials validates cost, revenue, and margin settings (AC2: ValidateMarginOverride)
func (s *BudgetDimensionService) validateFinancials(cost, revenue float64, targetMargin *float64, marginOverride bool) error {
	// Cost validation - must be positive
	if cost <= 0 {
		return ErrInvalidCost
	}

	// When margin override is true, revenue is calculated from cost and target margin
	// So we don't validate revenue in that case
	if !marginOverride {
		// Revenue validation - must be non-negative
		if revenue < 0 {
			return ErrInvalidRevenue
		}
	}

	// Target margin validation when margin override is enabled
	if marginOverride {
		if targetMargin == nil {
			return fmt.Errorf("%w: target margin is required when margin override is enabled", ErrInvalidTargetMargin)
		}
		if *targetMargin < 0 || *targetMargin >= 100 {
			return ErrInvalidTargetMargin
		}
	}

	return nil
}

// validateCategory checks if category exists and is active
func (s *BudgetDimensionService) validateCategory(ctx context.Context, categoryID string) error {
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidCategory
		}
		return fmt.Errorf("failed to validate category: %w", err)
	}

	if !category.IsActive {
		return ErrInvalidCategory
	}

	return nil
}

// validateParentExists verifies the parent entity (offer or project) exists
func (s *BudgetDimensionService) validateParentExists(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	switch parentType {
	case domain.BudgetParentOffer:
		_, err := s.offerRepo.GetByID(ctx, parentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: offer %s not found", ErrParentNotFound, parentID)
			}
			return fmt.Errorf("failed to validate offer: %w", err)
		}
	case domain.BudgetParentProject:
		_, err := s.projectRepo.GetByID(ctx, parentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: project %s not found", ErrParentNotFound, parentID)
			}
			return fmt.Errorf("failed to validate project: %w", err)
		}
	default:
		return ErrInvalidParentType
	}

	return nil
}

// updateParentTotals recalculates and updates the parent entity's budget totals
func (s *BudgetDimensionService) updateParentTotals(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	summary, err := s.dimensionRepo.GetBudgetSummary(ctx, parentType, parentID)
	if err != nil {
		return err
	}

	switch parentType {
	case domain.BudgetParentOffer:
		// Update offer.Value to match total revenue
		offer, err := s.offerRepo.GetByID(ctx, parentID)
		if err != nil {
			return err
		}
		offer.Value = summary.TotalRevenue
		return s.offerRepo.Update(ctx, offer)

	case domain.BudgetParentProject:
		// Update project.Budget to match total revenue (or cost, depending on business rules)
		project, err := s.projectRepo.GetByID(ctx, parentID)
		if err != nil {
			return err
		}
		project.Budget = summary.TotalRevenue
		project.HasDetailedBudget = summary.DimensionCount > 0
		return s.projectRepo.Update(ctx, project)
	}

	return nil
}

// getActivityTargetType maps budget parent type to activity target type
func (s *BudgetDimensionService) getActivityTargetType(parentType domain.BudgetParentType) domain.ActivityTargetType {
	switch parentType {
	case domain.BudgetParentOffer:
		return domain.ActivityTargetOffer
	case domain.BudgetParentProject:
		return domain.ActivityTargetProject
	default:
		return domain.ActivityTargetOffer
	}
}

// logActivity creates an activity log entry (AC6)
func (s *BudgetDimensionService) logActivity(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, title, body string) {
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
		CreatorName: userCtx.DisplayName,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log activity", zap.Error(err))
	}
}
