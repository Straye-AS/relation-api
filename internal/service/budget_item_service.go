package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

// BudgetItemService handles business logic for budget items
type BudgetItemService struct {
	budgetItemRepo *repository.BudgetItemRepository
	offerRepo      *repository.OfferRepository
	projectRepo    *repository.ProjectRepository
	logger         *zap.Logger
}

// NewBudgetItemService creates a new BudgetItemService instance
func NewBudgetItemService(
	budgetItemRepo *repository.BudgetItemRepository,
	offerRepo *repository.OfferRepository,
	projectRepo *repository.ProjectRepository,
	logger *zap.Logger,
) *BudgetItemService {
	return &BudgetItemService{
		budgetItemRepo: budgetItemRepo,
		offerRepo:      offerRepo,
		projectRepo:    projectRepo,
		logger:         logger,
	}
}

// Create creates a new budget item
func (s *BudgetItemService) Create(ctx context.Context, req domain.CreateBudgetItemRequest) (*domain.BudgetItemDTO, error) {
	// Validate parent exists
	if err := s.validateParent(ctx, req.ParentType, req.ParentID); err != nil {
		return nil, err
	}

	// Get next display order if not specified
	displayOrder := req.DisplayOrder
	if displayOrder == 0 {
		maxOrder, err := s.budgetItemRepo.GetMaxDisplayOrder(ctx, req.ParentType, req.ParentID)
		if err != nil {
			s.logger.Error("Failed to get max display order", zap.Error(err))
			return nil, fmt.Errorf("failed to get max display order: %w", err)
		}
		displayOrder = maxOrder + 1
	}

	item := &domain.BudgetItem{
		ParentType:     req.ParentType,
		ParentID:       req.ParentID,
		Name:           req.Name,
		ExpectedCost:   req.ExpectedCost,
		ExpectedMargin: req.ExpectedMargin,
		Quantity:       req.Quantity,
		PricePerItem:   req.PricePerItem,
		Description:    req.Description,
		DisplayOrder:   displayOrder,
	}

	if err := s.budgetItemRepo.Create(ctx, item); err != nil {
		s.logger.Error("Failed to create budget item", zap.Error(err))
		return nil, fmt.Errorf("failed to create budget item: %w", err)
	}

	// Refresh to get computed fields (expected_revenue, expected_profit)
	item, err := s.budgetItemRepo.GetByID(ctx, item.ID)
	if err != nil {
		s.logger.Error("Failed to refresh budget item", zap.Error(err))
		return nil, fmt.Errorf("failed to refresh budget item: %w", err)
	}

	// Update parent totals
	if err := s.updateParentTotals(ctx, req.ParentType, req.ParentID); err != nil {
		s.logger.Warn("Failed to update parent totals after create", zap.Error(err))
	}

	dto := mapper.ToBudgetItemDTO(item)
	return &dto, nil
}

// GetByID retrieves a budget item by ID
func (s *BudgetItemService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetItemDTO, error) {
	item, err := s.budgetItemRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get budget item", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to get budget item: %w", err)
	}

	dto := mapper.ToBudgetItemDTO(item)
	return &dto, nil
}

// Update updates an existing budget item
func (s *BudgetItemService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateBudgetItemRequest) (*domain.BudgetItemDTO, error) {
	item, err := s.budgetItemRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get budget item for update", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to get budget item: %w", err)
	}

	// Update fields
	item.Name = req.Name
	item.ExpectedCost = req.ExpectedCost
	item.ExpectedMargin = req.ExpectedMargin
	item.Quantity = req.Quantity
	item.PricePerItem = req.PricePerItem
	item.Description = req.Description
	item.DisplayOrder = req.DisplayOrder

	if err := s.budgetItemRepo.Update(ctx, item); err != nil {
		s.logger.Error("Failed to update budget item", zap.Error(err))
		return nil, fmt.Errorf("failed to update budget item: %w", err)
	}

	// Refresh to get computed fields
	item, err = s.budgetItemRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to refresh budget item", zap.Error(err))
		return nil, fmt.Errorf("failed to refresh budget item: %w", err)
	}

	// Update parent totals
	if err := s.updateParentTotals(ctx, item.ParentType, item.ParentID); err != nil {
		s.logger.Warn("Failed to update parent totals after update", zap.Error(err))
	}

	dto := mapper.ToBudgetItemDTO(item)
	return &dto, nil
}

// Delete removes a budget item
func (s *BudgetItemService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get item first to know parent for totals update
	item, err := s.budgetItemRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get budget item for delete", zap.String("id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to get budget item: %w", err)
	}

	if err := s.budgetItemRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete budget item", zap.Error(err))
		return fmt.Errorf("failed to delete budget item: %w", err)
	}

	// Update parent totals
	if err := s.updateParentTotals(ctx, item.ParentType, item.ParentID); err != nil {
		s.logger.Warn("Failed to update parent totals after delete", zap.Error(err))
	}

	return nil
}

// ListByOffer returns all budget items for an offer
func (s *BudgetItemService) ListByOffer(ctx context.Context, offerID uuid.UUID) ([]domain.BudgetItemDTO, error) {
	return s.listByParent(ctx, domain.BudgetParentOffer, offerID)
}

// ListByProject returns all budget items for a project
func (s *BudgetItemService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.BudgetItemDTO, error) {
	return s.listByParent(ctx, domain.BudgetParentProject, projectID)
}

// listByParent returns all budget items for a parent entity
func (s *BudgetItemService) listByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) ([]domain.BudgetItemDTO, error) {
	items, err := s.budgetItemRepo.ListByParent(ctx, parentType, parentID)
	if err != nil {
		s.logger.Error("Failed to list budget items",
			zap.String("parentType", string(parentType)),
			zap.String("parentId", parentID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to list budget items: %w", err)
	}

	dtos := make([]domain.BudgetItemDTO, len(items))
	for i, item := range items {
		dtos[i] = mapper.ToBudgetItemDTO(&item)
	}

	return dtos, nil
}

// GetOfferBudgetWithDimensions returns offer budget summary and items
func (s *BudgetItemService) GetOfferBudgetWithDimensions(ctx context.Context, offerID uuid.UUID) (*domain.BudgetSummaryDTO, []domain.BudgetItemDTO, error) {
	// Verify offer exists
	_, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, nil, fmt.Errorf("offer not found: %w", err)
	}

	return s.getBudgetWithItems(ctx, domain.BudgetParentOffer, offerID)
}

// GetProjectBudgetWithItems returns project budget summary and items
func (s *BudgetItemService) GetProjectBudgetWithItems(ctx context.Context, projectID uuid.UUID) (*domain.BudgetSummaryDTO, []domain.BudgetItemDTO, error) {
	// Verify project exists
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("project not found: %w", err)
	}

	return s.getBudgetWithItems(ctx, domain.BudgetParentProject, projectID)
}

// getBudgetWithItems returns budget summary and items for a parent
func (s *BudgetItemService) getBudgetWithItems(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (*domain.BudgetSummaryDTO, []domain.BudgetItemDTO, error) {
	items, err := s.budgetItemRepo.ListByParent(ctx, parentType, parentID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list budget items: %w", err)
	}

	summary := mapper.ToBudgetSummaryDTO(parentType, parentID, items)

	dtos := make([]domain.BudgetItemDTO, len(items))
	for i, item := range items {
		dtos[i] = mapper.ToBudgetItemDTO(&item)
	}

	return &summary, dtos, nil
}

// AddToOffer adds a budget item to an offer
func (s *BudgetItemService) AddToOffer(ctx context.Context, offerID uuid.UUID, req domain.AddOfferBudgetItemRequest) (*domain.BudgetItemDTO, error) {
	// Verify offer exists
	_, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("offer not found: %w", err)
	}

	// Convert to CreateBudgetItemRequest
	createReq := domain.CreateBudgetItemRequest{
		ParentType:     domain.BudgetParentOffer,
		ParentID:       offerID,
		Name:           req.Name,
		ExpectedCost:   req.ExpectedCost,
		ExpectedMargin: req.ExpectedMargin,
		Quantity:       req.Quantity,
		PricePerItem:   req.PricePerItem,
		Description:    req.Description,
		DisplayOrder:   req.DisplayOrder,
	}

	return s.Create(ctx, createReq)
}

// UpdateOfferDimension updates a budget item belonging to an offer
func (s *BudgetItemService) UpdateOfferDimension(ctx context.Context, offerID, itemID uuid.UUID, req domain.UpdateBudgetItemRequest) (*domain.BudgetItemDTO, error) {
	// Verify item belongs to this offer
	item, err := s.budgetItemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("budget item not found: %w", err)
	}
	if item.ParentType != domain.BudgetParentOffer || item.ParentID != offerID {
		return nil, fmt.Errorf("budget item does not belong to this offer")
	}

	return s.Update(ctx, itemID, req)
}

// DeleteOfferDimension deletes a budget item from an offer
func (s *BudgetItemService) DeleteOfferDimension(ctx context.Context, offerID, itemID uuid.UUID) error {
	// Verify item belongs to this offer
	item, err := s.budgetItemRepo.GetByID(ctx, itemID)
	if err != nil {
		return fmt.Errorf("budget item not found: %w", err)
	}
	if item.ParentType != domain.BudgetParentOffer || item.ParentID != offerID {
		return fmt.Errorf("budget item does not belong to this offer")
	}

	return s.Delete(ctx, itemID)
}

// ReorderOfferDimensions reorders budget items for an offer
func (s *BudgetItemService) ReorderOfferDimensions(ctx context.Context, offerID uuid.UUID, req domain.ReorderBudgetItemsRequest) error {
	// Verify all items belong to this offer
	for _, id := range req.OrderedIDs {
		item, err := s.budgetItemRepo.GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("budget item %s not found: %w", id, err)
		}
		if item.ParentType != domain.BudgetParentOffer || item.ParentID != offerID {
			return fmt.Errorf("budget item %s does not belong to this offer", id)
		}
	}

	return s.budgetItemRepo.UpdateDisplayOrders(ctx, req.OrderedIDs)
}

// validateParent checks that the parent entity exists
func (s *BudgetItemService) validateParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	switch parentType {
	case domain.BudgetParentOffer:
		_, err := s.offerRepo.GetByID(ctx, parentID)
		if err != nil {
			return fmt.Errorf("offer not found: %w", err)
		}
	case domain.BudgetParentProject:
		_, err := s.projectRepo.GetByID(ctx, parentID)
		if err != nil {
			return fmt.Errorf("project not found: %w", err)
		}
	default:
		return fmt.Errorf("invalid parent type: %s", parentType)
	}
	return nil
}

// updateParentTotals recalculates and updates the parent's totals
func (s *BudgetItemService) updateParentTotals(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	summary, err := s.budgetItemRepo.GetSummaryByParent(ctx, parentType, parentID)
	if err != nil {
		return fmt.Errorf("failed to get budget summary: %w", err)
	}

	switch parentType {
	case domain.BudgetParentOffer:
		// Update offer value based on budget items
		if _, err := s.offerRepo.CalculateTotalsFromBudgetItems(ctx, parentID); err != nil {
			return fmt.Errorf("failed to update offer totals: %w", err)
		}
	case domain.BudgetParentProject:
		// Update project value based on budget items
		project, err := s.projectRepo.GetByID(ctx, parentID)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}
		project.Value = summary.TotalRevenue
		project.Cost = summary.TotalCost
		project.HasDetailedBudget = summary.ItemCount > 0
		if err := s.projectRepo.Update(ctx, project); err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}
	}

	return nil
}
