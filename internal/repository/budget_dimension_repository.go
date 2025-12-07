package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// MaxPageSize is the maximum allowed page size for paginated queries
const MaxPageSize = 200

// BudgetDimensionRepository handles database operations for budget dimensions
type BudgetDimensionRepository struct {
	db *gorm.DB
}

// NewBudgetDimensionRepository creates a new repository instance
func NewBudgetDimensionRepository(db *gorm.DB) *BudgetDimensionRepository {
	return &BudgetDimensionRepository{db: db}
}

// Create creates a new budget dimension
// The DB trigger handles margin_override calculation automatically
func (r *BudgetDimensionRepository) Create(ctx context.Context, dimension *domain.BudgetDimension) error {
	return r.db.WithContext(ctx).Create(dimension).Error
}

// GetByID retrieves a budget dimension by its ID with Category preloaded
func (r *BudgetDimensionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetDimension, error) {
	var dimension domain.BudgetDimension
	err := r.db.WithContext(ctx).
		Preload("Category").
		First(&dimension, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dimension, nil
}

// Update updates an existing budget dimension
// The DB trigger handles margin_override recalculation automatically
func (r *BudgetDimensionRepository) Update(ctx context.Context, dimension *domain.BudgetDimension) error {
	return r.db.WithContext(ctx).Save(dimension).Error
}

// Delete removes a budget dimension by ID
func (r *BudgetDimensionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.BudgetDimension{}, "id = ?", id).Error
}

// GetByParent retrieves all budget dimensions for a parent entity, ordered by display_order
func (r *BudgetDimensionRepository) GetByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) ([]domain.BudgetDimension, error) {
	var dimensions []domain.BudgetDimension
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Order("display_order ASC, created_at ASC").
		Find(&dimensions).Error
	return dimensions, err
}

// GetByParentPaginated retrieves budget dimensions for a parent entity with pagination
// Page must be >= 1, pageSize is capped at MaxPageSize (200)
func (r *BudgetDimensionRepository) GetByParentPaginated(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, page, pageSize int) ([]domain.BudgetDimension, int64, error) {
	var dimensions []domain.BudgetDimension
	var total int64

	// Validate and normalize pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20 // Default page size
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	query := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Order("display_order ASC, created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&dimensions).Error

	return dimensions, total, err
}

// DeleteByParent removes all budget dimensions for a parent entity
func (r *BudgetDimensionRepository) DeleteByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Delete(&domain.BudgetDimension{}).Error
}

// GetBudgetSummary returns aggregated budget totals for a parent entity
func (r *BudgetDimensionRepository) GetBudgetSummary(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost      float64
		TotalRevenue   float64
		DimensionCount int
	}

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Select("COALESCE(SUM(cost), 0) as total_cost, COALESCE(SUM(revenue), 0) as total_revenue, COUNT(*) as dimension_count").
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	summary := &domain.BudgetSummary{
		TotalCost:      result.TotalCost,
		TotalRevenue:   result.TotalRevenue,
		TotalMargin:    result.TotalRevenue - result.TotalCost,
		DimensionCount: result.DimensionCount,
	}

	// Calculate margin percent: ((Revenue - Cost) / Revenue) * 100, 0 if revenue=0
	if result.TotalRevenue > 0 {
		summary.MarginPercent = ((result.TotalRevenue - result.TotalCost) / result.TotalRevenue) * 100
	}

	return summary, nil
}

// GetTotalCost returns the total cost for a parent entity
func (r *BudgetDimensionRepository) GetTotalCost(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Select("COALESCE(SUM(cost), 0)").
		Scan(&total).Error
	return total, err
}

// GetTotalRevenue returns the total revenue for a parent entity
func (r *BudgetDimensionRepository) GetTotalRevenue(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Select("COALESCE(SUM(revenue), 0)").
		Scan(&total).Error
	return total, err
}

// ReorderDimensions updates the display_order of multiple dimensions in a transaction
// The orderedIDs slice determines the new order (index 0 = display_order 0, etc.)
func (r *BudgetDimensionRepository) ReorderDimensions(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, orderedIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range orderedIDs {
			result := tx.Model(&domain.BudgetDimension{}).
				Where("id = ? AND parent_type = ? AND parent_id = ?", id, parentType, parentID).
				Update("display_order", i)

			if result.Error != nil {
				return fmt.Errorf("failed to update dimension %s: %w", id, result.Error)
			}

			if result.RowsAffected == 0 {
				return fmt.Errorf("dimension %s not found for parent %s/%s", id, parentType, parentID)
			}
		}
		return nil
	})
}

// GetMaxDisplayOrder returns the highest display_order for a parent entity
func (r *BudgetDimensionRepository) GetMaxDisplayOrder(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (int, error) {
	var maxOrder *int
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Select("MAX(display_order)").
		Scan(&maxOrder).Error

	if err != nil {
		return 0, err
	}

	if maxOrder == nil {
		return -1, nil // Return -1 to indicate no dimensions exist
	}

	return *maxOrder, nil
}

// Count returns the number of budget dimensions for a parent entity
func (r *BudgetDimensionRepository) Count(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", parentType, parentID).
		Count(&count).Error
	return int(count), err
}
