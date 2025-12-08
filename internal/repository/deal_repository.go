package repository

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DealFilters contains all filter options for listing deals
type DealFilters struct {
	Stage         *domain.DealStage
	OwnerID       *string
	CustomerID    *uuid.UUID
	CompanyID     *domain.CompanyID
	Source        *string
	MinValue      *float64
	MaxValue      *float64
	MinProb       *int
	MaxProb       *int
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	CloseAfter    *time.Time
	CloseBefore   *time.Time
	IsWon         *bool
	IsLost        *bool
	SearchQuery   *string
}

// DealSortOption represents available sort options
type DealSortOption string

const (
	DealSortByCreatedDesc     DealSortOption = "created_desc"
	DealSortByCreatedAsc      DealSortOption = "created_asc"
	DealSortByValueDesc       DealSortOption = "value_desc"
	DealSortByValueAsc        DealSortOption = "value_asc"
	DealSortByProbabilityDesc DealSortOption = "probability_desc"
	DealSortByProbabilityAsc  DealSortOption = "probability_asc"
	DealSortByCloseDateDesc   DealSortOption = "close_date_desc"
	DealSortByCloseDateAsc    DealSortOption = "close_date_asc"
	DealSortByWeightedDesc    DealSortOption = "weighted_desc"
	DealSortByWeightedAsc     DealSortOption = "weighted_asc"
)

type DealRepository struct {
	db *gorm.DB
}

func NewDealRepository(db *gorm.DB) *DealRepository {
	return &DealRepository{db: db}
}

func (r *DealRepository) Create(ctx context.Context, deal *domain.Deal) error {
	// Omit associations to avoid GORM trying to validate related records
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(deal).Error
}

func (r *DealRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Deal, error) {
	var deal domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&deal).Error
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (r *DealRepository) Update(ctx context.Context, deal *domain.Deal) error {
	return r.db.WithContext(ctx).Save(deal).Error
}

func (r *DealRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Deal{}, "id = ?", id).Error
}

func (r *DealRepository) List(ctx context.Context, page, pageSize int, filters *DealFilters, sortBy DealSortOption) ([]domain.Deal, int64, error) {
	var deals []domain.Deal
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Deal{}).Preload("Customer").Preload("Company")

	// Apply multi-tenant company filter from context
	query = ApplyCompanyFilter(ctx, query)

	// Apply additional filters
	query = r.applyFilters(query, filters)

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = r.applySorting(query, sortBy)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&deals).Error

	return deals, total, err
}

// GetByStage returns all deals in a specific stage for pipeline views
func (r *DealRepository) GetByStage(ctx context.Context, stage domain.DealStage) ([]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("stage = ?", stage)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("value DESC").Find(&deals).Error
	return deals, err
}

// GetPipelineOverview returns deals grouped by stage for full pipeline view
func (r *DealRepository) GetPipelineOverview(ctx context.Context) (map[domain.DealStage][]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("stage, weighted_value DESC").Find(&deals).Error
	if err != nil {
		return nil, err
	}

	// Group by stage
	pipeline := make(map[domain.DealStage][]domain.Deal)
	for _, deal := range deals {
		pipeline[deal.Stage] = append(pipeline[deal.Stage], deal)
	}
	return pipeline, nil
}

// GetByOwner returns all deals owned by a specific user
func (r *DealRepository) GetByOwner(ctx context.Context, ownerID string) ([]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("owner_id = ?", ownerID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("created_at DESC").Find(&deals).Error
	return deals, err
}

// GetByCustomer returns all deals for a specific customer
func (r *DealRepository) GetByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("customer_id = ?", customerID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("created_at DESC").Find(&deals).Error
	return deals, err
}

// GetTotalPipelineValue returns the total value of all open deals
func (r *DealRepository) GetTotalPipelineValue(ctx context.Context) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Select("COALESCE(SUM(value), 0)").Scan(&total).Error
	return total, err
}

// GetWeightedPipelineValue returns the weighted (probability-adjusted) pipeline value
func (r *DealRepository) GetWeightedPipelineValue(ctx context.Context) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Select("COALESCE(SUM(weighted_value), 0)").Scan(&total).Error
	return total, err
}

// GetPipelineStats returns aggregated statistics for the sales pipeline
func (r *DealRepository) GetPipelineStats(ctx context.Context) (*PipelineStats, error) {
	stats := &PipelineStats{
		ByStage: make(map[domain.DealStage]StageStats),
	}

	// Get stats per stage
	type stageResult struct {
		Stage         domain.DealStage
		Count         int64
		TotalValue    float64
		WeightedValue float64
	}

	var results []stageResult
	query := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Select("stage, COUNT(*) as count, COALESCE(SUM(value), 0) as total_value, COALESCE(SUM(weighted_value), 0) as weighted_value").
		Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost}).
		Group("stage")
	query = ApplyCompanyFilter(ctx, query)
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	for _, r := range results {
		stats.ByStage[r.Stage] = StageStats{
			Count:         r.Count,
			TotalValue:    r.TotalValue,
			WeightedValue: r.WeightedValue,
		}
		stats.TotalCount += r.Count
		stats.TotalValue += r.TotalValue
		stats.WeightedValue += r.WeightedValue
	}

	return stats, nil
}

// PipelineStats holds aggregated pipeline statistics
type PipelineStats struct {
	TotalCount    int64
	TotalValue    float64
	WeightedValue float64
	ByStage       map[domain.DealStage]StageStats
}

// StageStats holds statistics for a single stage
type StageStats struct {
	Count         int64
	TotalValue    float64
	WeightedValue float64
}

// GetWonDeals returns deals that have been won within a date range
func (r *DealRepository) GetWonDeals(ctx context.Context, from, to time.Time) ([]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("stage = ?", domain.DealStageWon).
		Where("actual_close_date >= ? AND actual_close_date <= ?", from, to)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("actual_close_date DESC").Find(&deals).Error
	return deals, err
}

// GetLostDeals returns deals that have been lost within a date range
func (r *DealRepository) GetLostDeals(ctx context.Context, from, to time.Time) ([]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("stage = ?", domain.DealStageLost).
		Where("actual_close_date >= ? AND actual_close_date <= ?", from, to)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("actual_close_date DESC").Find(&deals).Error
	return deals, err
}

// GetDealsClosingBetween returns deals expected to close within a date range
func (r *DealRepository) GetDealsClosingBetween(ctx context.Context, from, to time.Time) ([]domain.Deal, error) {
	var deals []domain.Deal
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost}).
		Where("expected_close_date >= ? AND expected_close_date <= ?", from, to)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("expected_close_date ASC").Find(&deals).Error
	return deals, err
}

// Search performs full-text search on deals
func (r *DealRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Deal, error) {
	var deals []domain.Deal
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Company").
		Where("LOWER(title) LIKE ? OR LOWER(customer_name) LIKE ? OR LOWER(description) LIKE ?",
			searchPattern, searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Order("created_at DESC").Find(&deals).Error
	return deals, err
}

// UpdateStage updates only the stage field (used with stage history tracking)
func (r *DealRepository) UpdateStage(ctx context.Context, id uuid.UUID, stage domain.DealStage) error {
	updates := map[string]interface{}{
		"stage":      stage,
		"updated_at": time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.Deal{}).Where("id = ?", id).Updates(updates).Error
}

// MarkAsWon marks a deal as won with the close date
func (r *DealRepository) MarkAsWon(ctx context.Context, id uuid.UUID, closeDate time.Time) error {
	updates := map[string]interface{}{
		"stage":             domain.DealStageWon,
		"actual_close_date": closeDate,
		"probability":       100,
		"updated_at":        time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.Deal{}).Where("id = ?", id).Updates(updates).Error
}

// MarkAsLost marks a deal as lost with the close date, reason category, and notes
func (r *DealRepository) MarkAsLost(ctx context.Context, id uuid.UUID, closeDate time.Time, reasonCategory domain.LossReasonCategory, notes string) error {
	updates := map[string]interface{}{
		"stage":                domain.DealStageLost,
		"actual_close_date":    closeDate,
		"probability":          0,
		"loss_reason_category": reasonCategory,
		"lost_reason":          notes,
		"updated_at":           time.Now(),
	}
	return r.db.WithContext(ctx).Model(&domain.Deal{}).Where("id = ?", id).Updates(updates).Error
}

// GetForecast returns forecast data for a time period
func (r *DealRepository) GetForecast(ctx context.Context, months int) ([]ForecastPeriod, error) {
	now := time.Now()
	var periods []ForecastPeriod

	for i := 0; i < months; i++ {
		periodStart := time.Date(now.Year(), now.Month()+time.Month(i), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, -1)

		var result struct {
			TotalValue    float64
			WeightedValue float64
			DealCount     int64
		}

		query := r.db.WithContext(ctx).Model(&domain.Deal{}).
			Select("COALESCE(SUM(value), 0) as total_value, COALESCE(SUM(weighted_value), 0) as weighted_value, COUNT(*) as deal_count").
			Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost}).
			Where("expected_close_date >= ? AND expected_close_date <= ?", periodStart, periodEnd)
		query = ApplyCompanyFilter(ctx, query)
		if err := query.Scan(&result).Error; err != nil {
			return nil, err
		}

		periods = append(periods, ForecastPeriod{
			PeriodStart:   periodStart,
			PeriodEnd:     periodEnd,
			TotalValue:    result.TotalValue,
			WeightedValue: result.WeightedValue,
			DealCount:     result.DealCount,
		})
	}

	return periods, nil
}

// ForecastPeriod represents forecast data for a time period
type ForecastPeriod struct {
	PeriodStart   time.Time
	PeriodEnd     time.Time
	TotalValue    float64
	WeightedValue float64
	DealCount     int64
}

// WithTransaction executes operations within a transaction
func (r *DealRepository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// applyFilters applies all filter criteria to the query
func (r *DealRepository) applyFilters(query *gorm.DB, filters *DealFilters) *gorm.DB {
	if filters == nil {
		return query
	}

	if filters.Stage != nil {
		query = query.Where("stage = ?", *filters.Stage)
	}

	if filters.OwnerID != nil {
		query = query.Where("owner_id = ?", *filters.OwnerID)
	}

	if filters.CustomerID != nil {
		query = query.Where("customer_id = ?", *filters.CustomerID)
	}

	if filters.CompanyID != nil {
		query = query.Where("company_id = ?", *filters.CompanyID)
	}

	if filters.Source != nil {
		query = query.Where("source = ?", *filters.Source)
	}

	if filters.MinValue != nil {
		query = query.Where("value >= ?", *filters.MinValue)
	}

	if filters.MaxValue != nil {
		query = query.Where("value <= ?", *filters.MaxValue)
	}

	if filters.MinProb != nil {
		query = query.Where("probability >= ?", *filters.MinProb)
	}

	if filters.MaxProb != nil {
		query = query.Where("probability <= ?", *filters.MaxProb)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	if filters.CloseAfter != nil {
		query = query.Where("expected_close_date >= ?", *filters.CloseAfter)
	}

	if filters.CloseBefore != nil {
		query = query.Where("expected_close_date <= ?", *filters.CloseBefore)
	}

	if filters.IsWon != nil && *filters.IsWon {
		query = query.Where("stage = ?", domain.DealStageWon)
	}

	if filters.IsLost != nil && *filters.IsLost {
		query = query.Where("stage = ?", domain.DealStageLost)
	}

	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		searchPattern := "%" + strings.ToLower(*filters.SearchQuery) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(customer_name) LIKE ?", searchPattern, searchPattern)
	}

	return query
}

// applySorting applies the sorting option to the query
func (r *DealRepository) applySorting(query *gorm.DB, sortBy DealSortOption) *gorm.DB {
	switch sortBy {
	case DealSortByCreatedAsc:
		return query.Order("created_at ASC")
	case DealSortByValueDesc:
		return query.Order("value DESC")
	case DealSortByValueAsc:
		return query.Order("value ASC")
	case DealSortByProbabilityDesc:
		return query.Order("probability DESC")
	case DealSortByProbabilityAsc:
		return query.Order("probability ASC")
	case DealSortByCloseDateDesc:
		return query.Order("expected_close_date DESC NULLS LAST")
	case DealSortByCloseDateAsc:
		return query.Order("expected_close_date ASC NULLS LAST")
	case DealSortByWeightedDesc:
		return query.Order("weighted_value DESC")
	case DealSortByWeightedAsc:
		return query.Order("weighted_value ASC")
	default: // DealSortByCreatedDesc
		return query.Order("created_at DESC")
	}
}

// PipelineSummaryFromView represents a row from v_sales_pipeline_summary view
type PipelineSummaryFromView struct {
	CompanyID          domain.CompanyID
	CompanyName        string
	Stage              domain.DealStage
	DealCount          int64
	TotalValue         float64
	TotalWeightedValue float64
	AvgProbability     float64
	AvgDealValue       float64
	EarliestCloseDate  *time.Time
	LatestCloseDate    *time.Time
	OverdueCount       int64
}

// GetPipelineSummaryFromView queries the v_sales_pipeline_summary view for analytics
func (r *DealRepository) GetPipelineSummaryFromView(ctx context.Context, companyID *domain.CompanyID) ([]PipelineSummaryFromView, error) {
	var results []PipelineSummaryFromView

	query := r.db.WithContext(ctx).
		Table("v_sales_pipeline_summary").
		Select("company_id, company_name, stage, deal_count, total_value, total_weighted_value, avg_probability, avg_deal_value, earliest_close_date, latest_close_date, overdue_count")

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	}

	// Apply multi-tenant company filter from context
	query = ApplyCompanyFilterWithColumn(ctx, query, "company_id")

	err := query.Find(&results).Error
	return results, err
}

// RevenueForecastResult represents forecast data for a time period
type RevenueForecastResult struct {
	DealCount     int64
	TotalValue    float64
	WeightedValue float64
}

// GetRevenueForecastByDays returns forecast based on expected_close_date within given days
func (r *DealRepository) GetRevenueForecastByDays(ctx context.Context, days int, companyID *domain.CompanyID, ownerID *string) (*RevenueForecastResult, error) {
	result := &RevenueForecastResult{}

	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	query := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Select("COUNT(*) as deal_count, COALESCE(SUM(value), 0) as total_value, COALESCE(SUM(weighted_value), 0) as weighted_value").
		Where("stage NOT IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost}).
		Where("expected_close_date IS NOT NULL").
		Where("expected_close_date >= ? AND expected_close_date <= ?", now, endDate)

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	}

	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}

	// Apply multi-tenant company filter from context
	query = ApplyCompanyFilter(ctx, query)

	err := query.Scan(result).Error
	return result, err
}

// WinRateResult holds win/loss analysis data
type WinRateResult struct {
	TotalClosed      int64
	TotalWon         int64
	TotalLost        int64
	WinRate          float64
	WonValue         float64
	LostValue        float64
	AvgWonDealValue  float64
	AvgLostDealValue float64
	AvgDaysToClose   float64
}

// GetWinRateAnalysis returns win/loss statistics
func (r *DealRepository) GetWinRateAnalysis(ctx context.Context, companyID *domain.CompanyID, ownerID *string, dateFrom, dateTo *time.Time) (*WinRateResult, error) {
	result := &WinRateResult{}

	// Get won and lost counts and values
	type winLossRow struct {
		Stage      domain.DealStage
		Count      int64
		TotalValue float64
		AvgValue   float64
	}

	var rows []winLossRow
	query := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Select("stage, COUNT(*) as count, COALESCE(SUM(value), 0) as total_value, COALESCE(AVG(value), 0) as avg_value").
		Where("stage IN ?", []domain.DealStage{domain.DealStageWon, domain.DealStageLost}).
		Group("stage")

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	}

	if ownerID != nil {
		query = query.Where("owner_id = ?", *ownerID)
	}

	if dateFrom != nil {
		query = query.Where("actual_close_date >= ?", *dateFrom)
	}

	if dateTo != nil {
		query = query.Where("actual_close_date <= ?", *dateTo)
	}

	// Apply multi-tenant company filter from context
	query = ApplyCompanyFilter(ctx, query)

	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if row.Stage == domain.DealStageWon {
			result.TotalWon = row.Count
			result.WonValue = row.TotalValue
			result.AvgWonDealValue = row.AvgValue
		} else if row.Stage == domain.DealStageLost {
			result.TotalLost = row.Count
			result.LostValue = row.TotalValue
			result.AvgLostDealValue = row.AvgValue
		}
	}

	result.TotalClosed = result.TotalWon + result.TotalLost
	if result.TotalClosed > 0 {
		result.WinRate = (float64(result.TotalWon) / float64(result.TotalClosed)) * 100
	}

	// Calculate average days to close for won deals
	var avgDays struct {
		AvgDays float64
	}
	daysQuery := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Select("AVG(EXTRACT(EPOCH FROM (actual_close_date - created_at)) / 86400) as avg_days").
		Where("stage = ?", domain.DealStageWon).
		Where("actual_close_date IS NOT NULL")

	if companyID != nil {
		daysQuery = daysQuery.Where("company_id = ?", *companyID)
	}

	if ownerID != nil {
		daysQuery = daysQuery.Where("owner_id = ?", *ownerID)
	}

	if dateFrom != nil {
		daysQuery = daysQuery.Where("actual_close_date >= ?", *dateFrom)
	}

	if dateTo != nil {
		daysQuery = daysQuery.Where("actual_close_date <= ?", *dateTo)
	}

	daysQuery = ApplyCompanyFilter(ctx, daysQuery)

	if err := daysQuery.Scan(&avgDays).Error; err == nil {
		result.AvgDaysToClose = avgDays.AvgDays
	}

	return result, nil
}

// ConversionRateResult holds stage conversion data
type ConversionRateResult struct {
	FromStage      domain.DealStage
	ToStage        domain.DealStage
	DealsConverted int64
	TotalDeals     int64
	ConversionRate float64
}

// GetConversionRates calculates stage-to-stage conversion rates from deal_stage_history
func (r *DealRepository) GetConversionRates(ctx context.Context, companyID *domain.CompanyID) ([]ConversionRateResult, error) {
	// Define the key conversion stages to track
	conversions := []struct {
		from domain.DealStage
		to   domain.DealStage
	}{
		{domain.DealStageLead, domain.DealStageQualified},
		{domain.DealStageQualified, domain.DealStageProposal},
		{domain.DealStageProposal, domain.DealStageNegotiation},
		{domain.DealStageNegotiation, domain.DealStageWon},
	}

	var results []ConversionRateResult

	for _, conv := range conversions {
		// Count deals that transitioned FROM the source stage
		var totalFromStage int64
		fromQuery := r.db.WithContext(ctx).
			Table("deal_stage_history").
			Joins("JOIN deals ON deal_stage_history.deal_id = deals.id").
			Select("COUNT(DISTINCT deal_id)").
			Where("from_stage = ?", conv.from)

		// Apply multi-tenant filter from context
		fromQuery = ApplyCompanyFilterWithColumn(ctx, fromQuery, "deals.company_id")

		// Apply additional company filter if explicitly provided
		if companyID != nil {
			fromQuery = fromQuery.Where("deals.company_id = ?", *companyID)
		}

		fromQuery.Scan(&totalFromStage)

		// Count deals that reached the target stage
		var transitioned int64
		toQuery := r.db.WithContext(ctx).
			Table("deal_stage_history").
			Joins("JOIN deals ON deal_stage_history.deal_id = deals.id").
			Select("COUNT(DISTINCT deal_id)").
			Where("from_stage = ? AND to_stage = ?", conv.from, conv.to)

		// Apply multi-tenant filter from context
		toQuery = ApplyCompanyFilterWithColumn(ctx, toQuery, "deals.company_id")

		// Apply additional company filter if explicitly provided
		if companyID != nil {
			toQuery = toQuery.Where("deals.company_id = ?", *companyID)
		}

		toQuery.Scan(&transitioned)

		rate := 0.0
		if totalFromStage > 0 {
			rate = (float64(transitioned) / float64(totalFromStage)) * 100
		}

		results = append(results, ConversionRateResult{
			FromStage:      conv.from,
			ToStage:        conv.to,
			DealsConverted: transitioned,
			TotalDeals:     totalFromStage,
			ConversionRate: rate,
		})
	}

	return results, nil
}
