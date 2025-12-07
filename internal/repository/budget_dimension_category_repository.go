package repository

import (
	"context"
	"strings"

	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// CategoryWithUsage embeds BudgetDimensionCategory with usage count
type CategoryWithUsage struct {
	domain.BudgetDimensionCategory
	UsageCount int `json:"usageCount"`
}

// BudgetDimensionCategoryRepository handles database operations for budget dimension categories
type BudgetDimensionCategoryRepository struct {
	db *gorm.DB
}

// NewBudgetDimensionCategoryRepository creates a new repository instance
func NewBudgetDimensionCategoryRepository(db *gorm.DB) *BudgetDimensionCategoryRepository {
	return &BudgetDimensionCategoryRepository{db: db}
}

// Create creates a new budget dimension category
func (r *BudgetDimensionCategoryRepository) Create(ctx context.Context, category *domain.BudgetDimensionCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID retrieves a budget dimension category by its ID
func (r *BudgetDimensionCategoryRepository) GetByID(ctx context.Context, id string) (*domain.BudgetDimensionCategory, error) {
	var category domain.BudgetDimensionCategory
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByName retrieves a budget dimension category by name (case-insensitive)
// If companyID is provided, searches within that company's categories AND global categories
// If companyID is nil, only searches global categories
func (r *BudgetDimensionCategoryRepository) GetByName(ctx context.Context, name string, companyID *domain.CompanyID) (*domain.BudgetDimensionCategory, error) {
	var category domain.BudgetDimensionCategory
	query := r.db.WithContext(ctx).Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name))

	if companyID != nil {
		// Search company-specific OR global categories
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID)
	} else {
		// Only search global categories
		query = query.Where("company_id IS NULL")
	}

	err := query.First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Update updates an existing budget dimension category
func (r *BudgetDimensionCategoryRepository) Update(ctx context.Context, category *domain.BudgetDimensionCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete removes a budget dimension category by ID
func (r *BudgetDimensionCategoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.BudgetDimensionCategory{}, "id = ?", id).Error
}

// List returns budget dimension categories ordered by display_order
// If companyID is provided, returns both global categories AND company-specific categories
// If companyID is nil, returns only global categories
// If activeOnly is true, only returns categories where is_active = true
func (r *BudgetDimensionCategoryRepository) List(ctx context.Context, companyID *domain.CompanyID, activeOnly bool) ([]domain.BudgetDimensionCategory, error) {
	var categories []domain.BudgetDimensionCategory

	query := r.db.WithContext(ctx).Model(&domain.BudgetDimensionCategory{})

	if companyID != nil {
		// Return global categories AND company-specific categories
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID)
	} else {
		// Only return global categories
		query = query.Where("company_id IS NULL")
	}

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("display_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

// ListByCompanyOnly returns only company-specific categories (excludes global)
// Useful for managing a company's custom categories
func (r *BudgetDimensionCategoryRepository) ListByCompanyOnly(ctx context.Context, companyID domain.CompanyID, activeOnly bool) ([]domain.BudgetDimensionCategory, error) {
	var categories []domain.BudgetDimensionCategory

	query := r.db.WithContext(ctx).
		Model(&domain.BudgetDimensionCategory{}).
		Where("company_id = ?", companyID)

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("display_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

// ListGlobalOnly returns only global categories (company_id IS NULL)
func (r *BudgetDimensionCategoryRepository) ListGlobalOnly(ctx context.Context, activeOnly bool) ([]domain.BudgetDimensionCategory, error) {
	var categories []domain.BudgetDimensionCategory

	query := r.db.WithContext(ctx).
		Model(&domain.BudgetDimensionCategory{}).
		Where("company_id IS NULL")

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("display_order ASC, name ASC").Find(&categories).Error
	return categories, err
}

// GetUsageCount returns the number of BudgetDimensions using a specific category
func (r *BudgetDimensionCategoryRepository) GetUsageCount(ctx context.Context, categoryID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("category_id = ?", categoryID).
		Count(&count).Error
	return int(count), err
}

// ListWithUsageCounts returns categories with their usage counts
// If companyID is provided, returns both global categories AND company-specific categories
// If companyID is nil, returns only global categories
// If activeOnly is true, only returns categories where is_active = true
func (r *BudgetDimensionCategoryRepository) ListWithUsageCounts(ctx context.Context, companyID *domain.CompanyID, activeOnly bool) ([]CategoryWithUsage, error) {
	var results []CategoryWithUsage

	// Using a subquery to count usage for each category
	subQuery := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Select("category_id, COUNT(*) as usage_count").
		Where("category_id IS NOT NULL").
		Group("category_id")

	query := r.db.WithContext(ctx).
		Model(&domain.BudgetDimensionCategory{}).
		Select("budget_dimension_categories.*, COALESCE(usage.usage_count, 0) as usage_count").
		Joins("LEFT JOIN (?) as usage ON budget_dimension_categories.id = usage.category_id", subQuery)

	if companyID != nil {
		// Return global categories AND company-specific categories
		query = query.Where("budget_dimension_categories.company_id = ? OR budget_dimension_categories.company_id IS NULL", *companyID)
	} else {
		// Only return global categories
		query = query.Where("budget_dimension_categories.company_id IS NULL")
	}

	if activeOnly {
		query = query.Where("budget_dimension_categories.is_active = ?", true)
	}

	err := query.Order("budget_dimension_categories.display_order ASC, budget_dimension_categories.name ASC").
		Scan(&results).Error

	return results, err
}

// Count returns the total number of budget dimension categories
// If companyID is provided, counts both global categories AND company-specific categories
// If companyID is nil, counts only global categories
// If activeOnly is true, only counts categories where is_active = true
func (r *BudgetDimensionCategoryRepository) Count(ctx context.Context, companyID *domain.CompanyID, activeOnly bool) (int, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&domain.BudgetDimensionCategory{})

	if companyID != nil {
		// Count global categories AND company-specific categories
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID)
	} else {
		// Only count global categories
		query = query.Where("company_id IS NULL")
	}

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	err := query.Count(&count).Error
	return int(count), err
}
