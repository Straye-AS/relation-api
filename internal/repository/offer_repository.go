package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

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

// Update saves an offer after verifying company access
// Returns error if offer not found or user lacks access
// Applies company filter for multi-tenant isolation
func (r *OfferRepository) Update(ctx context.Context, offer *domain.Offer) error {
	// First verify the offer exists and belongs to the user's company
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", offer.ID)
	query = ApplyCompanyFilter(ctx, query)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	return r.db.WithContext(ctx).Save(offer).Error
}

func (r *OfferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.db.WithContext(ctx).Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Delete(&domain.Offer{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// GetItemsCount returns the number of items for an offer
// Applies company filter via subquery for multi-tenant isolation
func (r *OfferRepository) GetItemsCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64

	// Build subquery to filter offers by company
	offerSubquery := r.db.WithContext(ctx).Model(&domain.Offer{}).Select("id").Where("id = ?", offerID)
	offerSubquery = ApplyCompanyFilter(ctx, offerSubquery)

	err := r.db.WithContext(ctx).Model(&domain.OfferItem{}).
		Where("offer_id IN (?)", offerSubquery).
		Count(&count).Error
	return int(count), err
}

// GetFilesCount returns the number of files for an offer
// Applies company filter via subquery for multi-tenant isolation
func (r *OfferRepository) GetFilesCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64

	// Build subquery to filter offers by company
	offerSubquery := r.db.WithContext(ctx).Model(&domain.Offer{}).Select("id").Where("id = ?", offerID)
	offerSubquery = ApplyCompanyFilter(ctx, offerSubquery)

	err := r.db.WithContext(ctx).Model(&domain.File{}).
		Where("offer_id IN (?)", offerSubquery).
		Count(&count).Error
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
// Applies company filter for multi-tenant isolation
func (r *OfferRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OfferStatus) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Update("status", status)

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
// Applies company filter for multi-tenant isolation
func (r *OfferRepository) UpdatePhase(ctx context.Context, id uuid.UUID, phase domain.OfferPhase) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Update("phase", phase)

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
// Applies company filter for multi-tenant isolation on the update
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

	// Update the offer's Value field with company filter
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", offerID)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Update("value", totalRevenue)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to update offer value: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	return totalRevenue, nil
}

// GetBudgetDimensionsCount returns the number of budget dimensions for an offer
// Applies company filter via subquery for multi-tenant isolation
func (r *OfferRepository) GetBudgetDimensionsCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64

	// Build subquery to filter offers by company
	offerSubquery := r.db.WithContext(ctx).Model(&domain.Offer{}).Select("id").Where("id = ?", offerID)
	offerSubquery = ApplyCompanyFilter(ctx, offerSubquery)

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id IN (?)", domain.BudgetParentOffer, offerSubquery).
		Count(&count).Error
	return int(count), err
}

// GetBudgetSummary returns aggregated budget totals for an offer
// Applies company filter via subquery for multi-tenant isolation
func (r *OfferRepository) GetBudgetSummary(ctx context.Context, offerID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost      float64
		TotalRevenue   float64
		DimensionCount int
	}

	// Build subquery to filter offers by company
	offerSubquery := r.db.WithContext(ctx).Model(&domain.Offer{}).Select("id").Where("id = ?", offerID)
	offerSubquery = ApplyCompanyFilter(ctx, offerSubquery)

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id IN (?)", domain.BudgetParentOffer, offerSubquery).
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

// PhaseStats holds stats for a single phase
type PhaseStats struct {
	Count         int
	TotalValue    float64
	WeightedValue float64
}

// OfferStats holds aggregated offer statistics for dashboard
type OfferStats struct {
	TotalOffers    int64
	ActiveOffers   int64
	WonOffers      int64
	LostOffers     int64
	TotalValue     float64
	WeightedValue  float64
	ByPhase        map[domain.OfferPhase]int
	ByPhaseStats   map[domain.OfferPhase]PhaseStats
	AvgProbability float64
}

// GetOfferStats returns aggregated offer statistics for the dashboard
func (r *OfferRepository) GetOfferStats(ctx context.Context) (*OfferStats, error) {
	stats := &OfferStats{
		ByPhase:      make(map[domain.OfferPhase]int),
		ByPhaseStats: make(map[domain.OfferPhase]PhaseStats),
	}

	// Build base query with company filter
	baseQuery := r.db.WithContext(ctx).Model(&domain.Offer{})
	baseQuery = ApplyCompanyFilter(ctx, baseQuery)

	// Total offers
	if err := baseQuery.Count(&stats.TotalOffers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total offers: %w", err)
	}

	// Active offers (draft, in_progress, sent)
	activeQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseDraft,
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
		})
	activeQuery = ApplyCompanyFilter(ctx, activeQuery)
	if err := activeQuery.Count(&stats.ActiveOffers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active offers: %w", err)
	}

	// Won offers
	wonQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("phase = ?", domain.OfferPhaseWon)
	wonQuery = ApplyCompanyFilter(ctx, wonQuery)
	if err := wonQuery.Count(&stats.WonOffers).Error; err != nil {
		return nil, fmt.Errorf("failed to count won offers: %w", err)
	}

	// Lost offers
	lostQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("phase = ?", domain.OfferPhaseLost)
	lostQuery = ApplyCompanyFilter(ctx, lostQuery)
	if err := lostQuery.Count(&stats.LostOffers).Error; err != nil {
		return nil, fmt.Errorf("failed to count lost offers: %w", err)
	}

	// Total value of active offers
	valueQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseDraft,
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
		})
	valueQuery = ApplyCompanyFilter(ctx, valueQuery)
	if err := valueQuery.Select("COALESCE(SUM(value), 0)").Scan(&stats.TotalValue).Error; err != nil {
		return nil, fmt.Errorf("failed to sum offer values: %w", err)
	}

	// Weighted value (value * probability / 100)
	weightedQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseDraft,
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
		})
	weightedQuery = ApplyCompanyFilter(ctx, weightedQuery)
	if err := weightedQuery.Select("COALESCE(SUM(value * probability / 100), 0)").Scan(&stats.WeightedValue).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate weighted value: %w", err)
	}

	// Count and values by phase
	type phaseStats struct {
		Phase         domain.OfferPhase
		Count         int
		TotalValue    float64
		WeightedValue float64
	}
	var phaseStatsList []phaseStats
	phaseQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Select("phase, COUNT(*) as count, COALESCE(SUM(value), 0) as total_value, COALESCE(SUM(value * probability / 100), 0) as weighted_value").
		Group("phase")
	phaseQuery = ApplyCompanyFilter(ctx, phaseQuery)
	if err := phaseQuery.Scan(&phaseStatsList).Error; err != nil {
		return nil, fmt.Errorf("failed to count offers by phase: %w", err)
	}
	for _, ps := range phaseStatsList {
		stats.ByPhase[ps.Phase] = ps.Count
		stats.ByPhaseStats[ps.Phase] = PhaseStats{
			Count:         ps.Count,
			TotalValue:    ps.TotalValue,
			WeightedValue: ps.WeightedValue,
		}
	}

	// Average probability of active offers
	avgProbQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseDraft,
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
		})
	avgProbQuery = ApplyCompanyFilter(ctx, avgProbQuery)
	if err := avgProbQuery.Select("COALESCE(AVG(probability), 0)").Scan(&stats.AvgProbability).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate avg probability: %w", err)
	}

	return stats, nil
}

// GetRecentOffers returns the most recent offers
func (r *OfferRepository) GetRecentOffers(ctx context.Context, limit int) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Order("created_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}

// GetWinRate calculates the win rate (won / (won + lost) * 100)
func (r *OfferRepository) GetWinRate(ctx context.Context) (float64, error) {
	var won, lost int64

	wonQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("phase = ?", domain.OfferPhaseWon)
	wonQuery = ApplyCompanyFilter(ctx, wonQuery)
	if err := wonQuery.Count(&won).Error; err != nil {
		return 0, err
	}

	lostQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("phase = ?", domain.OfferPhaseLost)
	lostQuery = ApplyCompanyFilter(ctx, lostQuery)
	if err := lostQuery.Count(&lost).Error; err != nil {
		return 0, err
	}

	total := won + lost
	if total == 0 {
		return 0, nil
	}

	return float64(won) / float64(total) * 100, nil
}

// DashboardOfferStats holds offer statistics for the dashboard with 12-month window
type DashboardOfferStats struct {
	TotalOfferCount      int     // Count of offers excluding drafts and expired
	OfferReserve         float64 // Total value of active offers (in_progress, sent)
	WeightedOfferReserve float64 // Sum of (value * probability/100) for active offers
	AverageProbability   float64 // Average probability of active offers
}

// DashboardPipelineStats holds pipeline phase statistics for the dashboard
type DashboardPipelineStats struct {
	Phase         domain.OfferPhase
	Count         int
	TotalValue    float64
	WeightedValue float64
}

// DashboardWinRateStats holds win/loss statistics for the dashboard
type DashboardWinRateStats struct {
	WonCount        int
	LostCount       int
	WonValue        float64
	LostValue       float64
	WinRate         float64 // won_count / (won_count + lost_count), 0-1 scale
	EconomicWinRate float64 // won_value / (won_value + lost_value), 0-1 scale
}

// GetDashboardOfferStats returns offer statistics for the dashboard within a 12-month window
// Excludes drafts and expired offers from all calculations
func (r *OfferRepository) GetDashboardOfferStats(ctx context.Context, since time.Time) (*DashboardOfferStats, error) {
	stats := &DashboardOfferStats{}

	// Valid phases for counting (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseWon,
		domain.OfferPhaseLost,
	}

	// Active phases for reserve calculation
	activePhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
	}

	// Total offer count (excluding drafts and expired)
	countQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("created_at >= ?", since).
		Where("phase IN ?", validPhases)
	countQuery = ApplyCompanyFilter(ctx, countQuery)
	var totalCount int64
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count offers: %w", err)
	}
	stats.TotalOfferCount = int(totalCount)

	// Offer reserve (sum of value for active offers: in_progress, sent)
	reserveQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("created_at >= ?", since).
		Where("phase IN ?", activePhases)
	reserveQuery = ApplyCompanyFilter(ctx, reserveQuery)
	if err := reserveQuery.Select("COALESCE(SUM(value), 0)").Scan(&stats.OfferReserve).Error; err != nil {
		return nil, fmt.Errorf("failed to sum offer reserve: %w", err)
	}

	// Weighted offer reserve (sum of value * probability / 100 for active offers)
	weightedQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("created_at >= ?", since).
		Where("phase IN ?", activePhases)
	weightedQuery = ApplyCompanyFilter(ctx, weightedQuery)
	if err := weightedQuery.Select("COALESCE(SUM(value * probability / 100), 0)").Scan(&stats.WeightedOfferReserve).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate weighted reserve: %w", err)
	}

	// Average probability of active offers
	avgProbQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("created_at >= ?", since).
		Where("phase IN ?", activePhases)
	avgProbQuery = ApplyCompanyFilter(ctx, avgProbQuery)
	if err := avgProbQuery.Select("COALESCE(AVG(probability), 0)").Scan(&stats.AverageProbability).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate avg probability: %w", err)
	}

	return stats, nil
}

// GetDashboardPipelineStats returns pipeline statistics by phase for the dashboard
// Includes only: in_progress, sent, won, lost (excludes draft and expired)
func (r *OfferRepository) GetDashboardPipelineStats(ctx context.Context, since time.Time) ([]DashboardPipelineStats, error) {
	// Valid phases for pipeline (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseWon,
		domain.OfferPhaseLost,
	}

	type phaseResult struct {
		Phase         domain.OfferPhase
		Count         int
		TotalValue    float64
		WeightedValue float64
	}
	var results []phaseResult

	query := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Select("phase, COUNT(*) as count, COALESCE(SUM(value), 0) as total_value, COALESCE(SUM(value * probability / 100), 0) as weighted_value").
		Where("created_at >= ?", since).
		Where("phase IN ?", validPhases).
		Group("phase")
	query = ApplyCompanyFilter(ctx, query)

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get pipeline stats: %w", err)
	}

	// Convert to DashboardPipelineStats
	pipelineStats := make([]DashboardPipelineStats, len(results))
	for i, r := range results {
		pipelineStats[i] = DashboardPipelineStats{
			Phase:         r.Phase,
			Count:         r.Count,
			TotalValue:    r.TotalValue,
			WeightedValue: r.WeightedValue,
		}
	}

	return pipelineStats, nil
}

// GetDashboardWinRateStats returns win rate statistics for the dashboard within a 12-month window
func (r *OfferRepository) GetDashboardWinRateStats(ctx context.Context, since time.Time) (*DashboardWinRateStats, error) {
	stats := &DashboardWinRateStats{}

	// Get won count and value
	wonQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("created_at >= ?", since).
		Where("phase = ?", domain.OfferPhaseWon)
	wonQuery = ApplyCompanyFilter(ctx, wonQuery)

	var wonResult struct {
		Count      int64
		TotalValue float64
	}
	if err := wonQuery.Select("COUNT(*) as count, COALESCE(SUM(value), 0) as total_value").Scan(&wonResult).Error; err != nil {
		return nil, fmt.Errorf("failed to get won stats: %w", err)
	}
	stats.WonCount = int(wonResult.Count)
	stats.WonValue = wonResult.TotalValue

	// Get lost count and value
	lostQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("created_at >= ?", since).
		Where("phase = ?", domain.OfferPhaseLost)
	lostQuery = ApplyCompanyFilter(ctx, lostQuery)

	var lostResult struct {
		Count      int64
		TotalValue float64
	}
	if err := lostQuery.Select("COUNT(*) as count, COALESCE(SUM(value), 0) as total_value").Scan(&lostResult).Error; err != nil {
		return nil, fmt.Errorf("failed to get lost stats: %w", err)
	}
	stats.LostCount = int(lostResult.Count)
	stats.LostValue = lostResult.TotalValue

	// Calculate win rate (count-based): won / (won + lost)
	totalCount := stats.WonCount + stats.LostCount
	if totalCount > 0 {
		stats.WinRate = float64(stats.WonCount) / float64(totalCount)
	}

	// Calculate economic win rate (value-based): won_value / (won_value + lost_value)
	totalValue := stats.WonValue + stats.LostValue
	if totalValue > 0 {
		stats.EconomicWinRate = stats.WonValue / totalValue
	}

	return stats, nil
}

// GetRecentOffersInWindow returns the most recent offers created within the time window
// Excludes drafts from the results
func (r *OfferRepository) GetRecentOffersInWindow(ctx context.Context, since time.Time, limit int) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("created_at >= ?", since).
		Where("phase != ?", domain.OfferPhaseDraft).
		Order("created_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}
