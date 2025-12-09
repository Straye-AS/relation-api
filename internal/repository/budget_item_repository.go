package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// BudgetItemRepository handles database operations for budget items
type BudgetItemRepository struct {
	db *gorm.DB
}

// NewBudgetItemRepository creates a new BudgetItemRepository instance
func NewBudgetItemRepository(db *gorm.DB) *BudgetItemRepository {
	return &BudgetItemRepository{db: db}
}

// Create inserts a new budget item into the database
func (r *BudgetItemRepository) Create(ctx context.Context, item *domain.BudgetItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetByID retrieves a budget item by its ID
func (r *BudgetItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetItem, error) {
	var item domain.BudgetItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update saves changes to an existing budget item
func (r *BudgetItemRepository) Update(ctx context.Context, item *domain.BudgetItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete removes a budget item from the database
func (r *BudgetItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.BudgetItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByParent returns all budget items for a specific parent (offer or project)
func (r *BudgetItemRepository) ListByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) ([]domain.BudgetItem, error) {
	var items []domain.BudgetItem
	err := r.db.WithContext(ctx).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Order("display_order ASC, created_at ASC").
		Find(&items).Error
	return items, err
}

// DeleteByParent removes all budget items for a specific parent
func (r *BudgetItemRepository) DeleteByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Delete(&domain.BudgetItem{}).Error
}

// CountByParent returns the count of budget items for a specific parent
func (r *BudgetItemRepository) CountByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Count(&count).Error
	return int(count), err
}

// GetMaxDisplayOrder returns the maximum display_order for a parent's budget items
func (r *BudgetItemRepository) GetMaxDisplayOrder(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (int, error) {
	var maxOrder *int
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Select("MAX(display_order)").
		Scan(&maxOrder).Error
	if err != nil {
		return 0, err
	}
	if maxOrder == nil {
		return 0, nil
	}
	return *maxOrder, nil
}

// UpdateDisplayOrders updates the display order of multiple budget items in a transaction
func (r *BudgetItemRepository) UpdateDisplayOrders(ctx context.Context, orderedIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range orderedIDs {
			result := tx.Model(&domain.BudgetItem{}).
				Where("id = ?", id).
				Update("display_order", i)
			if result.Error != nil {
				return fmt.Errorf("failed to update display order for item %s: %w", id, result.Error)
			}
		}
		return nil
	})
}

// GetSummaryByParent calculates aggregated budget totals for a parent entity
func (r *BudgetItemRepository) GetSummaryByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost    float64
		TotalRevenue float64
		TotalProfit  float64
		ItemCount    int
	}

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Select("COALESCE(SUM(expected_cost), 0) as total_cost, COALESCE(SUM(expected_revenue), 0) as total_revenue, COALESCE(SUM(expected_profit), 0) as total_profit, COUNT(*) as item_count").
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
	}

	summary := &domain.BudgetSummary{
		TotalCost:    result.TotalCost,
		TotalRevenue: result.TotalRevenue,
		TotalProfit:  result.TotalProfit,
		ItemCount:    result.ItemCount,
	}

	// Calculate margin percent: (Profit / Revenue) * 100, 0 if revenue=0
	if result.TotalRevenue > 0 {
		summary.MarginPercent = (result.TotalProfit / result.TotalRevenue) * 100
	}

	return summary, nil
}

// CloneItems copies all budget items from one parent to another
// Returns the number of items cloned
func (r *BudgetItemRepository) CloneItems(ctx context.Context, fromParentType domain.BudgetParentType, fromParentID uuid.UUID, toParentType domain.BudgetParentType, toParentID uuid.UUID) (int, error) {
	// Get source items
	sourceItems, err := r.ListByParent(ctx, fromParentType, fromParentID)
	if err != nil {
		return 0, fmt.Errorf("failed to get source budget items: %w", err)
	}

	if len(sourceItems) == 0 {
		return 0, nil
	}

	// Clone items within a transaction
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range sourceItems {
			newItem := domain.BudgetItem{
				ParentType:     toParentType,
				ParentID:       toParentID,
				Name:           item.Name,
				ExpectedCost:   item.ExpectedCost,
				ExpectedMargin: item.ExpectedMargin,
				Quantity:       item.Quantity,
				PricePerItem:   item.PricePerItem,
				Description:    item.Description,
				DisplayOrder:   item.DisplayOrder,
			}
			if err := tx.Create(&newItem).Error; err != nil {
				return fmt.Errorf("failed to clone budget item: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return len(sourceItems), nil
}
