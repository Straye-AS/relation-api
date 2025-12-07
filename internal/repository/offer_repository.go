package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// OfferFilters defines filter options for offer listing
type OfferFilters struct {
	CustomerID *uuid.UUID
	ProjectID  *uuid.UUID
	Phase      *domain.OfferPhase
	Status     *domain.OfferStatus
}

type OfferRepository struct {
	db *gorm.DB
}

func NewOfferRepository(db *gorm.DB) *OfferRepository {
	return &OfferRepository{db: db}
}

func (r *OfferRepository) Create(ctx context.Context, offer *domain.Offer) error {
	return r.db.WithContext(ctx).Create(offer).Error
}

func (r *OfferRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	var offer domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Items").
		Preload("Files").
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

// GetByIDWithBudgetDimensions retrieves an offer by ID with all related data including budget dimensions
func (r *OfferRepository) GetByIDWithBudgetDimensions(ctx context.Context, id uuid.UUID) (*domain.Offer, []domain.BudgetDimension, error) {
	var offer domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Items").
		Preload("Files").
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&offer).Error
	if err != nil {
		return nil, nil, err
	}

	// Fetch budget dimensions separately (polymorphic relationship)
	var dimensions []domain.BudgetDimension
	err = r.db.WithContext(ctx).
		Preload("Category").
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentOffer, id).
		Order("display_order ASC, created_at ASC").
		Find(&dimensions).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load budget dimensions: %w", err)
	}

	return &offer, dimensions, nil
}

func (r *OfferRepository) Update(ctx context.Context, offer *domain.Offer) error {
	return r.db.WithContext(ctx).Save(offer).Error
}

func (r *OfferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Offer{}, "id = ?", id).Error
}

// List returns a paginated list of offers with optional filters
// Deprecated: Use ListWithFilters for new code
func (r *OfferRepository) List(ctx context.Context, page, pageSize int, customerID, projectID *uuid.UUID, phase *domain.OfferPhase) ([]domain.Offer, int64, error) {
	filters := &OfferFilters{
		CustomerID: customerID,
		ProjectID:  projectID,
		Phase:      phase,
	}
	return r.ListWithFilters(ctx, page, pageSize, filters)
}

// ListWithFilters returns a paginated list of offers with filter options including status
func (r *OfferRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *OfferFilters) ([]domain.Offer, int64, error) {
	var offers []domain.Offer
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

	query := r.db.WithContext(ctx).Model(&domain.Offer{}).Preload("Customer")

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	// Apply filters
	if filters != nil {
		if filters.CustomerID != nil {
			query = query.Where("customer_id = ?", *filters.CustomerID)
		}

		if filters.ProjectID != nil {
			query = query.Where("project_id = ?", *filters.ProjectID)
		}

		if filters.Phase != nil {
			query = query.Where("phase = ?", *filters.Phase)
		}

		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&offers).Error

	return offers, total, err
}

func (r *OfferRepository) GetItemsCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.OfferItem{}).Where("offer_id = ?", offerID).Count(&count).Error
	return int(count), err
}

func (r *OfferRepository) GetFilesCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.File{}).Where("offer_id = ?", offerID).Count(&count).Error
	return int(count), err
}

func (r *OfferRepository) GetTotalPipelineValue(ctx context.Context) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseSent,
			domain.OfferPhaseInProgress,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Select("COALESCE(SUM(value), 0)").
		Scan(&total).Error
	return total, err
}

func (r *OfferRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Offer, error) {
	var offers []domain.Offer
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("LOWER(title) LIKE ?", searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Find(&offers).Error
	return offers, err
}

// UpdateStatus updates only the status field of an offer
// Returns error if offer not found or update fails
func (r *OfferRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OfferStatus) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update offer status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// UpdatePhase updates only the phase field of an offer
// Returns error if offer not found or update fails
func (r *OfferRepository) UpdatePhase(ctx context.Context, id uuid.UUID, phase domain.OfferPhase) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id).
		Update("phase", phase)

	if result.Error != nil {
		return fmt.Errorf("failed to update offer phase: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// CalculateTotalsFromDimensions calculates and updates the offer's Value field
// by summing the revenue from all budget dimensions linked to this offer
func (r *OfferRepository) CalculateTotalsFromDimensions(ctx context.Context, offerID uuid.UUID) (float64, error) {
	var totalRevenue float64

	// Calculate total revenue from budget dimensions
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentOffer, offerID).
		Select("COALESCE(SUM(revenue), 0)").
		Scan(&totalRevenue).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate totals from dimensions: %w", err)
	}

	// Update the offer's Value field
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", offerID).
		Update("value", totalRevenue)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to update offer value: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	return totalRevenue, nil
}

// GetBudgetDimensionsCount returns the number of budget dimensions for an offer
func (r *OfferRepository) GetBudgetDimensionsCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentOffer, offerID).
		Count(&count).Error
	return int(count), err
}

// GetBudgetSummary returns aggregated budget totals for an offer
func (r *OfferRepository) GetBudgetSummary(ctx context.Context, offerID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost      float64
		TotalRevenue   float64
		DimensionCount int
	}

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentOffer, offerID).
		Select("COALESCE(SUM(cost), 0) as total_cost, COALESCE(SUM(revenue), 0) as total_revenue, COUNT(*) as dimension_count").
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get budget summary: %w", err)
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
