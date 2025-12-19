package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OfferFilters defines filter options for offer listing
type OfferFilters struct {
	CustomerID *uuid.UUID
	ProjectID  *uuid.UUID
	Phase      *domain.OfferPhase
	Status     *domain.OfferStatus
}

// offerSortableFields maps API field names to database column names for offers
// Only fields in this map can be used for sorting (whitelist approach)
var offerSortableFields = map[string]string{
	"createdAt":    "created_at",
	"updatedAt":    "updated_at",
	"title":        "title",
	"value":        "value",
	"probability":  "probability",
	"phase":        "phase",
	"status":       "status",
	"dueDate":      "due_date",
	"customerName": "customer_name",
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

// GetByIDWithBudgetItems retrieves an offer by ID with all related data including budget items
func (r *OfferRepository) GetByIDWithBudgetItems(ctx context.Context, id uuid.UUID) (*domain.Offer, []domain.BudgetItem, error) {
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

	// Fetch budget items separately (polymorphic relationship)
	var items []domain.BudgetItem
	err = r.db.WithContext(ctx).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentOffer, id).
		Order("display_order ASC, created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load budget items: %w", err)
	}

	return &offer, items, nil
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
	return r.ListWithFilters(ctx, page, pageSize, filters, DefaultSortConfig())
}

// ListWithFilters returns a paginated list of offers with filter and sort options
func (r *OfferRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *OfferFilters, sort SortConfig) ([]domain.Offer, int64, error) {
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

	// Build order clause from sort config
	orderClause := BuildOrderClause(sort, offerSortableFields, "updated_at")

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&offers).Error

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
		Where(`LOWER(title) LIKE ? OR
			LOWER(offer_number) LIKE ? OR
			LOWER(external_reference) LIKE ? OR
			LOWER(customer_name) LIKE ? OR
			LOWER(description) LIKE ? OR
			LOWER(location) LIKE ? OR
			LOWER(responsible_user_name) LIKE ?`,
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("updated_at DESC").Limit(limit).Find(&offers).Error
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

// CalculateTotalsFromBudgetItems calculates and updates the offer's Value field
// by summing the expected_revenue from all budget items linked to this offer
// Applies company filter for multi-tenant isolation on the update
func (r *OfferRepository) CalculateTotalsFromBudgetItems(ctx context.Context, offerID uuid.UUID) (float64, error) {
	var totalRevenue float64

	// Calculate total revenue from budget items
	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentOffer, offerID).
		Select("COALESCE(SUM(expected_revenue), 0)").
		Scan(&totalRevenue).Error

	if err != nil {
		return 0, fmt.Errorf("failed to calculate totals from budget items: %w", err)
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

// GetBudgetItemsCount returns the number of budget items for an offer
// Applies company filter via subquery for multi-tenant isolation
func (r *OfferRepository) GetBudgetItemsCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64

	// Build subquery to filter offers by company
	offerSubquery := r.db.WithContext(ctx).Model(&domain.Offer{}).Select("id").Where("id = ?", offerID)
	offerSubquery = ApplyCompanyFilter(ctx, offerSubquery)

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id IN (?)", domain.BudgetParentOffer, offerSubquery).
		Count(&count).Error
	return int(count), err
}

// GetBudgetSummary returns aggregated budget totals for an offer
// Applies company filter via subquery for multi-tenant isolation
func (r *OfferRepository) GetBudgetSummary(ctx context.Context, offerID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost    float64
		TotalRevenue float64
		TotalProfit  float64
		ItemCount    int
	}

	// Build subquery to filter offers by company
	offerSubquery := r.db.WithContext(ctx).Model(&domain.Offer{}).Select("id").Where("id = ?", offerID)
	offerSubquery = ApplyCompanyFilter(ctx, offerSubquery)

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id IN (?)", domain.BudgetParentOffer, offerSubquery).
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

	// Won offers (order + completed phases)
	wonQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("phase IN ?", []domain.OfferPhase{
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
	})
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
		Order("updated_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}

// GetWinRate calculates the win rate (won / (won + lost) * 100)
// Won includes both order and completed phases
func (r *OfferRepository) GetWinRate(ctx context.Context) (float64, error) {
	var won, lost int64

	wonQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("phase IN ?", []domain.OfferPhase{
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
	})
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

// GetDashboardOfferStats returns offer statistics for the dashboard
// If since is nil, no date filter is applied (all time)
// Excludes drafts and expired offers from all calculations
func (r *OfferRepository) GetDashboardOfferStats(ctx context.Context, since *time.Time) (*DashboardOfferStats, error) {
	stats := &DashboardOfferStats{}

	// Valid phases for counting (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
		domain.OfferPhaseLost,
	}

	// Active phases for reserve calculation
	activePhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
	}

	// Total offer count (excluding drafts and expired)
	countQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", validPhases)
	if since != nil {
		countQuery = countQuery.Where("created_at >= ?", *since)
	}
	countQuery = ApplyCompanyFilter(ctx, countQuery)
	var totalCount int64
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count offers: %w", err)
	}
	stats.TotalOfferCount = int(totalCount)

	// Offer reserve (sum of value for active offers: in_progress, sent)
	reserveQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", activePhases)
	if since != nil {
		reserveQuery = reserveQuery.Where("created_at >= ?", *since)
	}
	reserveQuery = ApplyCompanyFilter(ctx, reserveQuery)
	if err := reserveQuery.Select("COALESCE(SUM(value), 0)").Scan(&stats.OfferReserve).Error; err != nil {
		return nil, fmt.Errorf("failed to sum offer reserve: %w", err)
	}

	// Weighted offer reserve (sum of value * probability / 100 for active offers)
	weightedQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", activePhases)
	if since != nil {
		weightedQuery = weightedQuery.Where("created_at >= ?", *since)
	}
	weightedQuery = ApplyCompanyFilter(ctx, weightedQuery)
	if err := weightedQuery.Select("COALESCE(SUM(value * probability / 100), 0)").Scan(&stats.WeightedOfferReserve).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate weighted reserve: %w", err)
	}

	// Average probability of active offers
	avgProbQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", activePhases)
	if since != nil {
		avgProbQuery = avgProbQuery.Where("created_at >= ?", *since)
	}
	avgProbQuery = ApplyCompanyFilter(ctx, avgProbQuery)
	if err := avgProbQuery.Select("COALESCE(AVG(probability), 0)").Scan(&stats.AverageProbability).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate avg probability: %w", err)
	}

	return stats, nil
}

// GetDashboardPipelineStats returns pipeline statistics by phase for the dashboard
// If since is nil, no date filter is applied (all time)
// Includes only: in_progress, sent, order, completed, lost (excludes draft and expired)
func (r *OfferRepository) GetDashboardPipelineStats(ctx context.Context, since *time.Time) ([]DashboardPipelineStats, error) {
	// Valid phases for pipeline (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
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
		Where("phase IN ?", validPhases)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	query = query.Group("phase")
	query = ApplyCompanyFilter(ctx, query)

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get pipeline stats: %w", err)
	}

	// Convert to DashboardPipelineStats
	pipelineStats := make([]DashboardPipelineStats, len(results))
	for i, r := range results {
		pipelineStats[i] = DashboardPipelineStats(r)
	}

	return pipelineStats, nil
}

// GetDashboardWinRateStats returns win rate statistics for the dashboard
// If since is nil, no date filter is applied (all time)
// Won includes both order and completed phases
func (r *OfferRepository) GetDashboardWinRateStats(ctx context.Context, fromDate, toDate *time.Time) (*DashboardWinRateStats, error) {
	stats := &DashboardWinRateStats{}

	// Get won count and value (order + completed phases)
	wonQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
		})
	if fromDate != nil {
		wonQuery = wonQuery.Where("created_at >= ?", *fromDate)
	}
	if toDate != nil {
		wonQuery = wonQuery.Where("created_at <= ?", *toDate)
	}
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
		Where("phase = ?", domain.OfferPhaseLost)
	if fromDate != nil {
		lostQuery = lostQuery.Where("created_at >= ?", *fromDate)
	}
	if toDate != nil {
		lostQuery = lostQuery.Where("created_at <= ?", *toDate)
	}
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
// If since is nil, no date filter is applied (all time)
// Excludes drafts from the results
func (r *OfferRepository) GetRecentOffersInWindow(ctx context.Context, since *time.Time, limit int) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("phase != ?", domain.OfferPhaseDraft)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	query = query.Order("updated_at DESC").Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}

// GetRecentOffersByPhase returns the most recent offers in a specific phase within the time window
// fromDate and toDate filter by updated_at for recency
// Sorted by updated_at DESC for recency
func (r *OfferRepository) GetRecentOffersByPhase(ctx context.Context, phase domain.OfferPhase, fromDate, toDate *time.Time, limit int) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("phase = ?", phase)
	if fromDate != nil {
		query = query.Where("updated_at >= ?", *fromDate)
	}
	if toDate != nil {
		query = query.Where("updated_at <= ?", *toDate)
	}
	query = query.Order("updated_at DESC").Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}

// ============================================================================
// Inquiry (Draft Offer) Methods
// ============================================================================

// ListInquiries returns a paginated list of offers in draft phase (inquiries)
func (r *OfferRepository) ListInquiries(ctx context.Context, page, pageSize int, customerID *uuid.UUID) ([]domain.Offer, int64, error) {
	var offers []domain.Offer
	var total int64

	// Validate and normalize pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	query := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Preload("Customer").
		Where("phase = ?", domain.OfferPhaseDraft)

	query = ApplyCompanyFilter(ctx, query)

	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&offers).Error

	return offers, total, err
}

// UpdateField updates a single field on an offer
// Returns error if offer not found or user lacks access
func (r *OfferRepository) UpdateField(ctx context.Context, id uuid.UUID, field string, value interface{}) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Update(field, value)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer %s: %w", field, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateFields updates multiple fields on an offer
func (r *OfferRepository) UpdateFields(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ============================================================================
// Offer Number Generation Methods
// ============================================================================

// GenerateOfferNumber generates the next unique offer number for a company
// Format: {PREFIX}-{YEAR}-{SEQUENCE} e.g., "STB-2024-001"
// Uses SELECT FOR UPDATE to ensure thread-safe sequence generation
func (r *OfferRepository) GenerateOfferNumber(ctx context.Context, companyID domain.CompanyID) (string, error) {
	year := time.Now().Year()
	prefix := domain.GetCompanyPrefix(companyID)

	var seq domain.OfferNumberSequence
	var nextSeq int

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Try to get existing sequence with row lock
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("company_id = ? AND year = ?", companyID, year).
			First(&seq)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new sequence for this company/year
			seq = domain.OfferNumberSequence{
				CompanyID:    companyID,
				Year:         year,
				LastSequence: 1,
			}
			if err := tx.Create(&seq).Error; err != nil {
				return fmt.Errorf("failed to create offer number sequence: %w", err)
			}
			nextSeq = 1
		} else if result.Error != nil {
			return fmt.Errorf("failed to get offer number sequence: %w", result.Error)
		} else {
			// Increment existing sequence
			nextSeq = seq.LastSequence + 1
			if err := tx.Model(&seq).Update("last_sequence", nextSeq).Error; err != nil {
				return fmt.Errorf("failed to update offer number sequence: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Format: PREFIX-YYYY-NNN (zero-padded to 3 digits)
	offerNumber := fmt.Sprintf("%s-%d-%03d", prefix, year, nextSeq)
	return offerNumber, nil
}

// SetOfferNumber sets the offer number for an offer
func (r *OfferRepository) SetOfferNumber(ctx context.Context, id uuid.UUID, offerNumber string) error {
	return r.UpdateField(ctx, id, "offer_number", offerNumber)
}

// LinkToProject sets the project_id and project_name for an offer
func (r *OfferRepository) LinkToProject(ctx context.Context, offerID uuid.UUID, projectID uuid.UUID) error {
	// Fetch project name for denormalized field
	var projectName string
	err := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID).
		Pluck("name", &projectName).Error
	if err != nil {
		return fmt.Errorf("failed to get project name: %w", err)
	}

	return r.UpdateFields(ctx, offerID, map[string]interface{}{
		"project_id":   projectID,
		"project_name": projectName,
	})
}

// UnlinkFromProject removes the project link from an offer
func (r *OfferRepository) UnlinkFromProject(ctx context.Context, offerID uuid.UUID) error {
	return r.UpdateFields(ctx, offerID, map[string]interface{}{
		"project_id":   nil,
		"project_name": "",
	})
}

// OfferNumberExists checks if an offer number already exists, excluding the given offer ID
func (r *OfferRepository) OfferNumberExists(ctx context.Context, offerNumber string, excludeOfferID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("offer_number = ? AND id != ?", offerNumber, excludeOfferID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check offer number existence: %w", err)
	}
	return count > 0, nil
}

// SetExternalReference sets the external reference for an offer
func (r *OfferRepository) SetExternalReference(ctx context.Context, id uuid.UUID, externalReference string) error {
	return r.UpdateField(ctx, id, "external_reference", externalReference)
}

// ExternalReferenceExists checks if an external reference already exists within a company, excluding the given offer ID
func (r *OfferRepository) ExternalReferenceExists(ctx context.Context, externalReference string, companyID domain.CompanyID, excludeOfferID uuid.UUID) (bool, error) {
	if externalReference == "" {
		return false, nil // Empty references are allowed to not be unique
	}
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("external_reference = ? AND company_id = ? AND id != ?", externalReference, companyID, excludeOfferID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check external reference existence: %w", err)
	}
	return count > 0, nil
}

// ============================================================================
// Project-Offer Relationship Methods (Offer Folder Model)
// ============================================================================

// ListByProject returns all offers linked to a specific project
func (r *OfferRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Items").
		Where("project_id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("updated_at DESC").Find(&offers).Error
	return offers, err
}

// GetHighestOfferValueForProject returns the highest offer value among all offers in a project
// Only considers offers that are not in terminal states (order, completed, lost, expired)
func (r *OfferRepository) GetHighestOfferValueForProject(ctx context.Context, projectID uuid.UUID) (float64, error) {
	var maxValue float64
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("COALESCE(MAX(value), 0)").
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Scan(&maxValue).Error
	return maxValue, err
}

// ExpireSiblingOffers marks all other offers in the same project as expired
// This is called when one offer becomes an order - the others become expired (NOT lost)
// Returns the IDs of the expired offers
func (r *OfferRepository) ExpireSiblingOffers(ctx context.Context, projectID uuid.UUID, winningOfferID uuid.UUID) ([]uuid.UUID, error) {
	// First get the IDs of offers that will be expired
	var offerIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("id").
		Where("project_id = ?", projectID).
		Where("id != ?", winningOfferID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		}).
		Pluck("id", &offerIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sibling offer IDs: %w", err)
	}

	if len(offerIDs) == 0 {
		return nil, nil
	}

	// Update the phase to expired for sibling offers
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id IN ?", offerIDs).
		Update("phase", domain.OfferPhaseExpired)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to expire sibling offers: %w", result.Error)
	}

	return offerIDs, nil
}

// GetExpiredSiblingOffers returns offers that were expired by selecting another offer
func (r *OfferRepository) GetExpiredSiblingOffers(ctx context.Context, projectID uuid.UUID, winningOfferID uuid.UUID) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_id = ?", projectID).
		Where("id != ?", winningOfferID).
		Where("phase = ?", domain.OfferPhaseExpired)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&offers).Error
	return offers, err
}

// SetOfferNumberWithSuffix updates an offer's number by adding a suffix
// This is used when an offer wins to mark it with "_P" suffix
func (r *OfferRepository) SetOfferNumberWithSuffix(ctx context.Context, id uuid.UUID, suffix string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id).
		Update("offer_number", gorm.Expr("offer_number || ?", suffix)).Error
}

// CountOffersByProject returns the count of offers for a project
func (r *OfferRepository) CountOffersByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return count, err
}

// CountActiveOffersByProject returns the count of active offers (not order/completed/lost/expired) for a project
func (r *OfferRepository) CountActiveOffersByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return count, err
}

// GetActiveOffersByProject returns all active offers (not order/completed/lost/expired) for a project
func (r *OfferRepository) GetActiveOffersByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Offer, error) {
	var offers []domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("value DESC").Find(&offers).Error
	return offers, err
}

// GetBestActiveOfferForProject returns the highest value active offer for a project
// Returns nil if no active offers exist
func (r *OfferRepository) GetBestActiveOfferForProject(ctx context.Context, projectID uuid.UUID) (*domain.Offer, error) {
	var offer domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseOrder,
			domain.OfferPhaseCompleted,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		}).
		Order("value DESC").
		Limit(1)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&offer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No active offers found
		}
		return nil, err
	}
	return &offer, nil
}

// GetDistinctCustomerIDsForActiveOffers returns the distinct customer IDs from active offers
// (in_progress or sent phase) for a project. Used to determine if project customer can be inferred.
// Returns empty slice if no active offers exist.
func (r *OfferRepository) GetDistinctCustomerIDsForActiveOffers(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var customerIDs []uuid.UUID
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("DISTINCT customer_id").
		Where("project_id = ?", projectID).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Pluck("customer_id", &customerIDs).Error
	if err != nil {
		return nil, err
	}
	return customerIDs, nil
}

// UpdateProjectNameByProjectID updates the project_name for all offers linked to a project
func (r *OfferRepository) UpdateProjectNameByProjectID(ctx context.Context, projectID uuid.UUID, projectName string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID).
		Update("project_name", projectName)
	if result.Error != nil {
		return fmt.Errorf("failed to update project_name for offers: %w", result.Error)
	}
	return nil
}

// GetBestOfferForProject returns the "best" offer for a project based on priority:
// 1. Completed offer (if any exists)
// 2. Order offer (if any exists)
// 3. Sent offer with highest value
// 4. In_progress offer with highest value
// 5. Draft offer with highest value
// Returns nil if no offers exist for the project.
func (r *OfferRepository) GetBestOfferForProject(ctx context.Context, projectID uuid.UUID) (*domain.Offer, error) {
	var offer domain.Offer

	// Priority 1: Completed offer
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseCompleted).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query completed offers: %w", err)
	}

	// Priority 2: Order offer (in execution)
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseOrder).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query order offers: %w", err)
	}

	// Priority 3: Sent offer with highest value
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseSent).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query sent offers: %w", err)
	}

	// Priority 4: In_progress offer with highest value
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseInProgress).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query in_progress offers: %w", err)
	}

	// Priority 5: Draft offer with highest value
	err = r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("phase = ?", domain.OfferPhaseDraft).
		Order("value DESC").
		First(&offer).Error
	if err == nil {
		return &offer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to query draft offers: %w", err)
	}

	// No offers found
	return nil, nil
}

// ============================================================================
// Dashboard Aggregation Methods (Uses Pre-computed View)
// ============================================================================

// AggregatedPipelineStats holds aggregated pipeline statistics from the view
// This avoids double-counting by using MAX(value) per project per phase
type AggregatedPipelineStats struct {
	Phase         domain.OfferPhase
	ProjectCount  int     // Unique projects in this phase
	OfferCount    int     // Total offers (including all offers per project)
	TotalValue    float64 // Sum of best values per project (no double-counting)
	WeightedValue float64 // Weighted by probability
}

// GetAggregatedPipelineStats returns pipeline statistics using the aggregation view
// This solves the double-counting problem by using MAX(value) per project per phase
// If since is nil, no date filter is applied (queries the pre-computed view)
// Note: The view only includes valid phases (excludes draft and expired)
func (r *OfferRepository) GetAggregatedPipelineStats(ctx context.Context, fromDate, toDate *time.Time) ([]AggregatedPipelineStats, error) {
	// If we have a date filter, we need to fall back to raw query
	// because the view doesn't have date-based aggregation
	if fromDate != nil || toDate != nil {
		return r.getAggregatedPipelineStatsWithDateFilter(ctx, fromDate, toDate)
	}

	// Query the pre-computed view
	type viewResult struct {
		Phase         string  `gorm:"column:phase"`
		ProjectCount  int     `gorm:"column:project_count"`
		OfferCount    int     `gorm:"column:offer_count"`
		TotalValue    float64 `gorm:"column:total_value"`
		WeightedValue float64 `gorm:"column:weighted_value"`
	}

	var results []viewResult
	query := r.db.WithContext(ctx).
		Table("dashboard_metrics_aggregation").
		Select("phase, project_count, offer_count, total_value, weighted_value")
	query = ApplyCompanyFilterWithColumn(ctx, query, "company_id")

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get aggregated pipeline stats: %w", err)
	}

	// Convert to our stats type
	stats := make([]AggregatedPipelineStats, len(results))
	for i, r := range results {
		stats[i] = AggregatedPipelineStats{
			Phase:         domain.OfferPhase(r.Phase),
			ProjectCount:  r.ProjectCount,
			OfferCount:    r.OfferCount,
			TotalValue:    r.TotalValue,
			WeightedValue: r.WeightedValue,
		}
	}

	return stats, nil
}

// getAggregatedPipelineStatsWithDateFilter computes aggregated stats with a date filter
// This is needed because the view doesn't support date filtering
func (r *OfferRepository) getAggregatedPipelineStatsWithDateFilter(ctx context.Context, fromDate, toDate *time.Time) ([]AggregatedPipelineStats, error) {
	// Valid phases (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
		domain.OfferPhaseLost,
	}

	type aggregatedResult struct {
		Phase         string  `gorm:"column:phase"`
		ProjectCount  int     `gorm:"column:project_count"`
		OfferCount    int     `gorm:"column:offer_count"`
		TotalValue    float64 `gorm:"column:total_value"`
		WeightedValue float64 `gorm:"column:weighted_value"`
	}

	// Apply company filter
	companyID := auth.GetEffectiveCompanyFilter(ctx)

	// Build date filter conditions
	dateConditions := []string{}
	dateArgs := []interface{}{}
	argIndex := 1

	if fromDate != nil {
		dateConditions = append(dateConditions, fmt.Sprintf("o.created_at >= $%d", argIndex))
		dateArgs = append(dateArgs, *fromDate)
		argIndex++
	}
	if toDate != nil {
		dateConditions = append(dateConditions, fmt.Sprintf("o.created_at <= $%d", argIndex))
		dateArgs = append(dateArgs, *toDate)
		argIndex++
	}

	dateFilter := "TRUE"
	if len(dateConditions) > 0 {
		dateFilter = strings.Join(dateConditions, " AND ")
	}

	// Build inner date filter for subqueries (same conditions but for o2/o3)
	innerDateFilter := strings.ReplaceAll(dateFilter, "o.", "o2.")
	innerDateFilter3 := strings.ReplaceAll(dateFilter, "o.", "o3.")

	var results []aggregatedResult
	var query *gorm.DB

	// Phase placeholders start after date args
	phaseStartIdx := argIndex
	allArgs := append(dateArgs, validPhases[0], validPhases[1], validPhases[2], validPhases[3], validPhases[4])

	if companyID != nil {
		companyIdx := phaseStartIdx + 5
		allArgs = append(allArgs, *companyID)
		query = r.db.WithContext(ctx).Raw(fmt.Sprintf(`
			WITH
			project_best_offers AS (
				SELECT
					o.company_id,
					o.phase,
					o.project_id,
					MAX(o.value) AS best_value,
					(
						SELECT o2.probability
						FROM offers o2
						WHERE o2.project_id = o.project_id
						  AND o2.phase = o.phase
						  AND o2.company_id = o.company_id
						  AND o2.value = (SELECT MAX(o3.value) FROM offers o3 WHERE o3.project_id = o.project_id AND o3.phase = o.phase AND o3.company_id = o.company_id AND %s)
						  AND %s
						LIMIT 1
					) AS best_probability,
					COUNT(*) AS offer_count
				FROM offers o
				WHERE o.project_id IS NOT NULL
				  AND o.phase IN ($%d, $%d, $%d, $%d, $%d)
				  AND %s
				  AND o.company_id = $%d
				GROUP BY o.company_id, o.phase, o.project_id
			),
			orphan_offers AS (
				SELECT
					o.company_id,
					o.phase,
					o.value,
					o.probability,
					1 AS offer_count
				FROM offers o
				WHERE o.project_id IS NULL
				  AND o.phase IN ($%d, $%d, $%d, $%d, $%d)
				  AND %s
				  AND o.company_id = $%d
			),
			combined_metrics AS (
				SELECT
					company_id,
					phase,
					project_id,
					best_value AS value,
					best_probability AS probability,
					offer_count,
					1 AS project_count
				FROM project_best_offers

				UNION ALL

				SELECT
					company_id,
					phase,
					NULL AS project_id,
					value,
					probability,
					offer_count,
					0 AS project_count
				FROM orphan_offers
			)
			SELECT
				phase,
				COALESCE(SUM(project_count), 0) AS project_count,
				COALESCE(SUM(offer_count), 0) AS offer_count,
				COALESCE(SUM(value), 0) AS total_value,
				COALESCE(SUM(value * COALESCE(probability, 0) / 100.0), 0) AS weighted_value
			FROM combined_metrics
			GROUP BY phase
		`, innerDateFilter3, innerDateFilter,
			phaseStartIdx, phaseStartIdx+1, phaseStartIdx+2, phaseStartIdx+3, phaseStartIdx+4, dateFilter, companyIdx,
			phaseStartIdx, phaseStartIdx+1, phaseStartIdx+2, phaseStartIdx+3, phaseStartIdx+4, dateFilter, companyIdx), allArgs...)
	} else {
		query = r.db.WithContext(ctx).Raw(fmt.Sprintf(`
			WITH
			project_best_offers AS (
				SELECT
					o.company_id,
					o.phase,
					o.project_id,
					MAX(o.value) AS best_value,
					(
						SELECT o2.probability
						FROM offers o2
						WHERE o2.project_id = o.project_id
						  AND o2.phase = o.phase
						  AND o2.company_id = o.company_id
						  AND o2.value = (SELECT MAX(o3.value) FROM offers o3 WHERE o3.project_id = o.project_id AND o3.phase = o.phase AND o3.company_id = o.company_id AND %s)
						  AND %s
						LIMIT 1
					) AS best_probability,
					COUNT(*) AS offer_count
				FROM offers o
				WHERE o.project_id IS NOT NULL
				  AND o.phase IN ($%d, $%d, $%d, $%d, $%d)
				  AND %s
				GROUP BY o.company_id, o.phase, o.project_id
			),
			orphan_offers AS (
				SELECT
					o.company_id,
					o.phase,
					o.value,
					o.probability,
					1 AS offer_count
				FROM offers o
				WHERE o.project_id IS NULL
				  AND o.phase IN ($%d, $%d, $%d, $%d, $%d)
				  AND %s
			),
			combined_metrics AS (
				SELECT
					company_id,
					phase,
					project_id,
					best_value AS value,
					best_probability AS probability,
					offer_count,
					1 AS project_count
				FROM project_best_offers

				UNION ALL

				SELECT
					company_id,
					phase,
					NULL AS project_id,
					value,
					probability,
					offer_count,
					0 AS project_count
				FROM orphan_offers
			)
			SELECT
				phase,
				COALESCE(SUM(project_count), 0) AS project_count,
				COALESCE(SUM(offer_count), 0) AS offer_count,
				COALESCE(SUM(value), 0) AS total_value,
				COALESCE(SUM(value * COALESCE(probability, 0) / 100.0), 0) AS weighted_value
			FROM combined_metrics
			GROUP BY phase
		`, innerDateFilter3, innerDateFilter,
			phaseStartIdx, phaseStartIdx+1, phaseStartIdx+2, phaseStartIdx+3, phaseStartIdx+4, dateFilter,
			phaseStartIdx, phaseStartIdx+1, phaseStartIdx+2, phaseStartIdx+3, phaseStartIdx+4, dateFilter), allArgs...)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get aggregated pipeline stats with date filter: %w", err)
	}

	// Convert to our stats type
	stats := make([]AggregatedPipelineStats, len(results))
	for i, r := range results {
		stats[i] = AggregatedPipelineStats{
			Phase:         domain.OfferPhase(r.Phase),
			ProjectCount:  r.ProjectCount,
			OfferCount:    r.OfferCount,
			TotalValue:    r.TotalValue,
			WeightedValue: r.WeightedValue,
		}
	}

	return stats, nil
}

// AggregatedOfferStats holds aggregated offer statistics avoiding double-counting
type AggregatedOfferStats struct {
	TotalOfferCount      int     // Count of offers excluding drafts and expired
	TotalProjectCount    int     // Count of unique projects
	OfferReserve         float64 // Total value of active offers (best per project)
	WeightedOfferReserve float64 // Weighted by probability
	AverageProbability   float64 // Average probability of active offers
}

// GetAggregatedOfferStats returns offer statistics using aggregation logic
// This avoids double-counting by using MAX(value) per project
func (r *OfferRepository) GetAggregatedOfferStats(ctx context.Context, fromDate, toDate *time.Time) (*AggregatedOfferStats, error) {
	stats := &AggregatedOfferStats{}

	// Get aggregated pipeline stats
	pipelineStats, err := r.GetAggregatedPipelineStats(ctx, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregated pipeline stats: %w", err)
	}

	// Aggregate across all phases
	for _, ps := range pipelineStats {
		stats.TotalOfferCount += ps.OfferCount
		stats.TotalProjectCount += ps.ProjectCount

		// Only active phases contribute to reserve
		if ps.Phase == domain.OfferPhaseInProgress || ps.Phase == domain.OfferPhaseSent {
			stats.OfferReserve += ps.TotalValue
			stats.WeightedOfferReserve += ps.WeightedValue
		}
	}

	// Calculate average probability from raw offers for active phases
	// (view doesn't provide this directly)
	activePhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
	}
	avgProbQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", activePhases)
	if fromDate != nil {
		avgProbQuery = avgProbQuery.Where("created_at >= ?", *fromDate)
	}
	if toDate != nil {
		avgProbQuery = avgProbQuery.Where("created_at <= ?", *toDate)
	}
	avgProbQuery = ApplyCompanyFilter(ctx, avgProbQuery)
	if err := avgProbQuery.Select("COALESCE(AVG(probability), 0)").Scan(&stats.AverageProbability).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate avg probability: %w", err)
	}

	return stats, nil
}

// ============================================================================
// Order Phase Execution Methods
// ============================================================================

// UpdateOfferHealth updates the health status and optional completion percent for offers in order phase
func (r *OfferRepository) UpdateOfferHealth(ctx context.Context, tx *gorm.DB, id uuid.UUID, health domain.OfferHealth, completionPercent *float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	updates := map[string]interface{}{
		"health": health,
	}
	if completionPercent != nil {
		updates["completion_percent"] = *completionPercent
	}

	query := db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer health: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateOfferSpent updates the spent amount for offers in order phase
func (r *OfferRepository) UpdateOfferSpent(ctx context.Context, tx *gorm.DB, id uuid.UUID, spent float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	query := db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Update("spent", spent)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer spent: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateOfferInvoiced updates the invoiced amount for offers in order phase
func (r *OfferRepository) UpdateOfferInvoiced(ctx context.Context, tx *gorm.DB, id uuid.UUID, invoiced float64) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	query := db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Update("invoiced", invoiced)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer invoiced: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateOfferManager updates the manager for offers
func (r *OfferRepository) UpdateOfferManager(ctx context.Context, tx *gorm.DB, id uuid.UUID, managerID *string, managerName string) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	query := db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Updates(map[string]interface{}{
		"manager_id":   managerID,
		"manager_name": managerName,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update offer manager: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateOfferTeamMembers updates team members for offers
func (r *OfferRepository) UpdateOfferTeamMembers(ctx context.Context, tx *gorm.DB, id uuid.UUID, teamMembers []string) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	query := db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Update("team_members", pq.StringArray(teamMembers))
	if result.Error != nil {
		return fmt.Errorf("failed to update offer team members: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateOfferDates updates start/end dates for offers in order phase
func (r *OfferRepository) UpdateOfferDates(ctx context.Context, tx *gorm.DB, id uuid.UUID, startDate, endDate, estimatedCompletionDate *time.Time) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	updates := map[string]interface{}{}
	if startDate != nil {
		updates["start_date"] = *startDate
	}
	if endDate != nil {
		updates["end_date"] = *endDate
	}
	if estimatedCompletionDate != nil {
		updates["estimated_completion_date"] = *estimatedCompletionDate
	}

	if len(updates) == 0 {
		return nil // No updates needed
	}

	query := db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)

	result := query.Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update offer dates: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ============================================================================
// Order Dashboard Statistics Methods
// ============================================================================

// OrderStats holds statistics for offers in order/completed phases
type OrderStats struct {
	TotalOrders     int64                      // Count of offers in order phase
	CompletedOrders int64                      // Count of offers in completed phase
	TotalValue      float64                    // Sum of value for order phase offers
	TotalSpent      float64                    // Sum of spent for order phase offers
	TotalInvoiced   float64                    // Sum of invoiced for order phase offers
	OrderReserve    float64                    // Sum of (value - invoiced) for order phase offers
	AvgCompletion   float64                    // Average completion_percent for order phase offers
	ByHealth        map[domain.OfferHealth]int // Count by health status
}

// GetOrderStats returns statistics for offers in order/completed phases
func (r *OfferRepository) GetOrderStats(ctx context.Context, companyID *domain.CompanyID, fromDate, toDate *time.Time) (*OrderStats, error) {
	stats := &OrderStats{
		ByHealth: make(map[domain.OfferHealth]int),
	}

	// Build base query
	baseQuery := r.db.WithContext(ctx).Model(&domain.Offer{})
	if companyID != nil {
		baseQuery = baseQuery.Where("company_id = ?", *companyID)
	} else {
		baseQuery = ApplyCompanyFilter(ctx, baseQuery)
	}
	// Apply date filters
	if fromDate != nil {
		baseQuery = baseQuery.Where("created_at >= ?", *fromDate)
	}
	if toDate != nil {
		baseQuery = baseQuery.Where("created_at <= ?", *toDate)
	}

	// Count offers in order phase
	orderQuery := baseQuery.Session(&gorm.Session{}).Where("phase = ?", domain.OfferPhaseOrder)
	if err := orderQuery.Count(&stats.TotalOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count order phase offers: %w", err)
	}

	// Count completed offers
	completedQuery := baseQuery.Session(&gorm.Session{}).Where("phase = ?", domain.OfferPhaseCompleted)
	if err := completedQuery.Count(&stats.CompletedOrders).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed offers: %w", err)
	}

	// Aggregate values for order phase offers
	var aggregates struct {
		TotalValue    float64
		TotalSpent    float64
		TotalInvoiced float64
		OrderReserve  float64
		AvgCompletion float64
	}
	aggregateQuery := baseQuery.Session(&gorm.Session{}).
		Where("phase = ?", domain.OfferPhaseOrder).
		Select(`
			COALESCE(SUM(value), 0) as total_value,
			COALESCE(SUM(spent), 0) as total_spent,
			COALESCE(SUM(invoiced), 0) as total_invoiced,
			COALESCE(SUM(value - invoiced), 0) as order_reserve,
			COALESCE(AVG(completion_percent), 0) as avg_completion
		`)
	if err := aggregateQuery.Scan(&aggregates).Error; err != nil {
		return nil, fmt.Errorf("failed to aggregate order stats: %w", err)
	}

	stats.TotalValue = aggregates.TotalValue
	stats.TotalSpent = aggregates.TotalSpent
	stats.TotalInvoiced = aggregates.TotalInvoiced
	stats.OrderReserve = aggregates.OrderReserve
	stats.AvgCompletion = aggregates.AvgCompletion

	// Count by health status
	healthCounts, err := r.CountByHealthStatus(ctx, companyID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to count by health: %w", err)
	}
	stats.ByHealth = healthCounts

	return stats, nil
}

// CountByHealthStatus returns count of offers grouped by health status
// Only counts offers in order phase (active execution)
func (r *OfferRepository) CountByHealthStatus(ctx context.Context, companyID *domain.CompanyID, fromDate, toDate *time.Time) (map[domain.OfferHealth]int, error) {
	type healthCount struct {
		Health domain.OfferHealth
		Count  int
	}

	var results []healthCount

	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("health, COUNT(*) as count").
		Where("phase = ?", domain.OfferPhaseOrder).
		Where("health IS NOT NULL").
		Group("health")

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	} else {
		query = ApplyCompanyFilter(ctx, query)
	}
	// Apply date filters
	if fromDate != nil {
		query = query.Where("created_at >= ?", *fromDate)
	}
	if toDate != nil {
		query = query.Where("created_at <= ?", *toDate)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count by health status: %w", err)
	}

	counts := make(map[domain.OfferHealth]int)
	for _, r := range results {
		counts[r.Health] = r.Count
	}

	return counts, nil
}

// ListOrderPhaseOffers returns offers in order phase with pagination
func (r *OfferRepository) ListOrderPhaseOffers(ctx context.Context, page, pageSize int, companyID *domain.CompanyID, health *domain.OfferHealth) ([]domain.Offer, int64, error) {
	var offers []domain.Offer
	var total int64

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Preload("Customer").
		Where("phase = ?", domain.OfferPhaseOrder)

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	} else {
		query = ApplyCompanyFilter(ctx, query)
	}

	if health != nil {
		query = query.Where("health = ?", *health)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count order phase offers: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.
		Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&offers).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list order phase offers: %w", err)
	}

	return offers, total, nil
}

// ============================================================================
// Project Offer Count Methods
// ============================================================================

// CountOffersByProjectIDs returns the count of offers for each project ID
// Returns a map where key is project ID and value is offer count
func (r *OfferRepository) CountOffersByProjectIDs(ctx context.Context, projectIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(projectIDs) == 0 {
		return make(map[uuid.UUID]int), nil
	}

	type projectCount struct {
		ProjectID uuid.UUID `gorm:"column:project_id"`
		Count     int       `gorm:"column:count"`
	}

	var results []projectCount

	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("project_id, COUNT(*) as count").
		Where("project_id IN ?", projectIDs).
		Group("project_id")

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count offers by project IDs: %w", err)
	}

	// Convert to map
	counts := make(map[uuid.UUID]int)
	for _, r := range results {
		counts[r.ProjectID] = r.Count
	}

	return counts, nil
}

// ProjectCustomerInfo holds customer information for offers in a project
type ProjectCustomerInfo struct {
	CustomerID   uuid.UUID
	CustomerName string
}

// GetUniqueCustomersForProject returns the unique customer IDs and names for all offers in a project
// If all offers have the same customer, returns a single entry
// If offers have different customers, returns multiple entries
// Returns empty slice if project has no offers
func (r *OfferRepository) GetUniqueCustomersForProject(ctx context.Context, projectID uuid.UUID) ([]ProjectCustomerInfo, error) {
	type customerResult struct {
		CustomerID   uuid.UUID `gorm:"column:customer_id"`
		CustomerName string    `gorm:"column:customer_name"`
	}

	var results []customerResult

	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Select("DISTINCT customer_id, customer_name").
		Where("project_id = ?", projectID)

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get unique customers for project: %w", err)
	}

	// Convert to return type
	customers := make([]ProjectCustomerInfo, len(results))
	for i, r := range results {
		customers[i] = ProjectCustomerInfo(r)
	}

	return customers, nil
}

// DWFinancials holds data warehouse financial data for updating an offer
type DWFinancials struct {
	TotalIncome   float64
	MaterialCosts float64
	EmployeeCosts float64
	OtherCosts    float64
	NetResult     float64
}

// UpdateDWFinancials updates the data warehouse synced financial fields on an offer.
// This is called after successfully querying the data warehouse for project financials.
func (r *OfferRepository) UpdateDWFinancials(ctx context.Context, id uuid.UUID, financials *DWFinancials) error {
	now := time.Now()
	updates := map[string]interface{}{
		"dw_total_income":   financials.TotalIncome,
		"dw_material_costs": financials.MaterialCosts,
		"dw_employee_costs": financials.EmployeeCosts,
		"dw_other_costs":    financials.OtherCosts,
		"dw_net_result":     financials.NetResult,
		"dw_last_synced_at": now,
		"updated_at":        now,
	}

	result := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update DW financials: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("offer not found: %s", id)
	}

	return nil
}

// GetOffersForDWSync returns all offers that have an external_reference set.
// These are offers that can be synced with the data warehouse.
// Returns offers regardless of phase since financial data is relevant for all phases.
func (r *OfferRepository) GetOffersForDWSync(ctx context.Context) ([]domain.Offer, error) {
	var offers []domain.Offer

	err := r.db.WithContext(ctx).
		Where("external_reference IS NOT NULL").
		Where("external_reference != ''").
		Find(&offers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get offers for DW sync: %w", err)
	}

	return offers, nil
}
