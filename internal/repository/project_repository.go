package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProjectFilters defines filter options for project listing
type ProjectFilters struct {
	CustomerID *uuid.UUID
	Phase      *domain.ProjectPhase
	Health     *domain.ProjectHealth
	ManagerID  *string
}

// projectSortableFields maps API field names to database column names for projects
// Only fields in this map can be used for sorting (whitelist approach)
var projectSortableFields = map[string]string{
	"createdAt":    "created_at",
	"updatedAt":    "updated_at",
	"name":         "name",
	"phase":        "phase",
	"health":       "health",
	"budget":       "budget",
	"spent":        "spent",
	"startDate":    "start_date",
	"endDate":      "end_date",
	"customerName": "customer_name",
	"wonAt":        "won_at",
}

// ProjectBudgetMetrics holds calculated budget metrics for a project
type ProjectBudgetMetrics struct {
	Value         float64
	Cost          float64
	MarginPercent float64
	Spent         float64
	Remaining     float64
	PercentUsed   float64
}

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	// Omit associations to avoid GORM trying to validate related records
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(project).Error
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	query := r.db.WithContext(ctx).Preload("Customer").Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Project{}, "id = ?", id).Error
}

// List returns a paginated list of projects with optional filters
// Deprecated: Use ListWithFilters for new code
func (r *ProjectRepository) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID, phase *domain.ProjectPhase) ([]domain.Project, int64, error) {
	filters := &ProjectFilters{
		CustomerID: customerID,
		Phase:      phase,
	}
	return r.ListWithFilters(ctx, page, pageSize, filters, DefaultSortConfig())
}

// ListWithFilters returns a paginated list of projects with filter and sort options
func (r *ProjectRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *ProjectFilters, sort SortConfig) ([]domain.Project, int64, error) {
	var projects []domain.Project
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

	query := r.db.WithContext(ctx).Model(&domain.Project{}).Preload("Customer")

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	// Apply filters
	if filters != nil {
		if filters.CustomerID != nil {
			query = query.Where("customer_id = ?", *filters.CustomerID)
		}

		if filters.Phase != nil {
			query = query.Where("phase = ?", *filters.Phase)
		}

		if filters.Health != nil {
			query = query.Where("health = ?", *filters.Health)
		}

		if filters.ManagerID != nil {
			query = query.Where("manager_id = ?", *filters.ManagerID)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build order clause from sort config
	orderClause := BuildOrderClause(sort, projectSortableFields, "created_at")

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&projects).Error

	return projects, total, err
}

func (r *ProjectRepository) CountActive(ctx context.Context) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("phase IN (?)", []domain.ProjectPhase{domain.ProjectPhaseWorking, domain.ProjectPhaseActive})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return int(count), err
}

func (r *ProjectRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).Preload("Customer").
		Where(`LOWER(name) LIKE ? OR
			LOWER(summary) LIKE ? OR
			LOWER(project_number) LIKE ? OR
			LOWER(customer_name) LIKE ? OR
			LOWER(description) LIKE ?`,
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("created_at DESC").Limit(limit).Find(&projects).Error
	return projects, err
}

// GetByIDWithRelations retrieves a project by ID with all related data
// Preloads: Customer, Offer, Deal, BudgetItems, Activities
func (r *ProjectRepository) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*domain.Project, []domain.BudgetItem, []domain.Activity, error) {
	var project domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Offer").
		Preload("Deal").
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&project).Error
	if err != nil {
		return nil, nil, nil, err
	}

	// Fetch budget items separately (polymorphic relationship)
	var items []domain.BudgetItem
	err = r.db.WithContext(ctx).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentProject, id).
		Order("display_order ASC, created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load budget items: %w", err)
	}

	// Fetch activities separately (polymorphic relationship)
	var activities []domain.Activity
	err = r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", domain.ActivityTargetProject, id).
		Order("occurred_at DESC").
		Limit(50). // Limit activities to most recent 50
		Find(&activities).Error
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load activities: %w", err)
	}

	return &project, items, activities, nil
}

// CalculateBudgetMetrics calculates budget metrics for a project
// Returns: budget, spent, remaining, percent_used
// If hasDetailedBudget is true, spent is calculated from BudgetItems expected_cost sum
// Otherwise, it uses the project's Spent field directly
func (r *ProjectRepository) CalculateBudgetMetrics(ctx context.Context, projectID uuid.UUID) (*ProjectBudgetMetrics, error) {
	// First get the project to get budget and check if it has detailed budget
	var project domain.Project
	query := r.db.WithContext(ctx).Where("id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&project).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	metrics := &ProjectBudgetMetrics{
		Value:         project.Value,
		Cost:          project.Cost,
		MarginPercent: project.MarginPercent,
	}

	// Calculate spent based on whether project uses detailed budget
	if project.HasDetailedBudget {
		// Sum expected_cost from budget items
		var totalCost float64
		err = r.db.WithContext(ctx).
			Model(&domain.BudgetItem{}).
			Where("parent_type = ? AND parent_id = ?", domain.BudgetParentProject, projectID).
			Select("COALESCE(SUM(expected_cost), 0)").
			Scan(&totalCost).Error
		if err != nil {
			return nil, fmt.Errorf("failed to calculate budget item costs: %w", err)
		}
		metrics.Spent = totalCost
	} else {
		// Use the project's Spent field directly
		metrics.Spent = project.Spent
	}

	// Calculate remaining and percent used
	metrics.Remaining = metrics.Value - metrics.Spent
	if metrics.Value > 0 {
		metrics.PercentUsed = (metrics.Spent / metrics.Value) * 100
	}

	return metrics, nil
}

// CalculateHealth calculates the health status of a project based on budget variance
// Health formula:
//   - on_track: variance < 10%
//   - at_risk: variance 10-20%
//   - over_budget: variance > 20%
//
// Variance = (actual_cost / budget) * 100
//
// Edge case: If budget is zero or negative, returns on_track to avoid
// division errors and assumes unbudgeted projects are healthy by default.
//
// Returns the calculated health status without updating the project
func (r *ProjectRepository) CalculateHealth(ctx context.Context, projectID uuid.UUID) (domain.ProjectHealth, error) {
	metrics, err := r.CalculateBudgetMetrics(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to calculate budget metrics: %w", err)
	}

	// If value is 0, default to on_track to avoid division issues
	if metrics.Value <= 0 {
		return domain.ProjectHealthOnTrack, nil
	}

	// Calculate variance as percentage of budget used
	variance := metrics.PercentUsed

	switch {
	case variance > 120:
		return domain.ProjectHealthOverBudget, nil
	case variance >= 110:
		return domain.ProjectHealthAtRisk, nil
	default:
		return domain.ProjectHealthOnTrack, nil
	}
}

// UpdateHealth recalculates and updates the health status of a project
// This should be called after budget-related changes
func (r *ProjectRepository) UpdateHealth(ctx context.Context, projectID uuid.UUID) error {
	health, err := r.CalculateHealth(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to calculate health: %w", err)
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Update("health", health)

	if result.Error != nil {
		return fmt.Errorf("failed to update project health: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetBudgetSummary returns aggregated budget totals for a project's budget items
func (r *ProjectRepository) GetBudgetSummary(ctx context.Context, projectID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost    float64
		TotalRevenue float64
		TotalProfit  float64
		ItemCount    int
	}

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetItem{}).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentProject, projectID).
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

// GetByManager returns all projects managed by a specific user
func (r *ProjectRepository) GetByManager(ctx context.Context, managerID string) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("manager_id = ?", managerID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// GetByHealth returns all projects with a specific health status
func (r *ProjectRepository) GetByHealth(ctx context.Context, health domain.ProjectHealth) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("health = ?", health)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// CountByHealth returns the count of projects grouped by health status
func (r *ProjectRepository) CountByHealth(ctx context.Context) (map[domain.ProjectHealth]int64, error) {
	type healthCount struct {
		Health domain.ProjectHealth
		Count  int64
	}

	var results []healthCount
	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Select("health, COUNT(*) as count").
		Where("phase IN (?)", []domain.ProjectPhase{domain.ProjectPhaseWorking, domain.ProjectPhaseActive}). // Only count active projects
		Group("health")
	query = ApplyCompanyFilter(ctx, query)
	err := query.Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count projects by health: %w", err)
	}

	counts := make(map[domain.ProjectHealth]int64)
	for _, r := range results {
		counts[r.Health] = r.Count
	}

	return counts, nil
}

// GetActiveProjects returns active projects with optional limit
func (r *ProjectRepository) GetActiveProjects(ctx context.Context, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Where("phase IN (?)", []domain.ProjectPhase{domain.ProjectPhaseWorking, domain.ProjectPhaseActive}).
		Order("updated_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&projects).Error
	return projects, err
}

// GetRecentProjects returns the most recently updated projects
func (r *ProjectRepository) GetRecentProjects(ctx context.Context, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Order("updated_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&projects).Error
	return projects, err
}

// DashboardProjectStats holds project statistics for the dashboard
type DashboardProjectStats struct {
	OrderReserve  float64 // Sum of (budget - spent) on active projects
	TotalInvoiced float64 // Sum of "spent" on all projects in the time window
}

// GetDashboardProjectStats returns project statistics for the dashboard
// If since is nil, no date filter is applied (all time)
func (r *ProjectRepository) GetDashboardProjectStats(ctx context.Context, since *time.Time) (*DashboardProjectStats, error) {
	stats := &DashboardProjectStats{}

	// Order reserve: sum of (budget - spent) on active projects
	// Only count projects with phase "working" or "active"
	reserveQuery := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("phase IN (?)", []domain.ProjectPhase{domain.ProjectPhaseWorking, domain.ProjectPhaseActive})
	reserveQuery = ApplyCompanyFilter(ctx, reserveQuery)
	if err := reserveQuery.Select("COALESCE(SUM(budget - spent), 0)").Scan(&stats.OrderReserve).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate order reserve: %w", err)
	}

	// Total invoiced: sum of "spent" on all projects in the time window
	invoicedQuery := r.db.WithContext(ctx).Model(&domain.Project{})
	if since != nil {
		invoicedQuery = invoicedQuery.Where("created_at >= ?", *since)
	}
	invoicedQuery = ApplyCompanyFilter(ctx, invoicedQuery)
	if err := invoicedQuery.Select("COALESCE(SUM(spent), 0)").Scan(&stats.TotalInvoiced).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate total invoiced: %w", err)
	}

	return stats, nil
}

// GetRecentProjectsInWindow returns the most recently created projects within the time window
// If since is nil, no date filter is applied (all time)
func (r *ProjectRepository) GetRecentProjectsInWindow(ctx context.Context, since *time.Time, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer")
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	query = query.Order("created_at DESC").Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&projects).Error
	return projects, err
}

// ============================================================================
// Phase Management Methods
// ============================================================================

// UpdatePhase updates the project phase and related fields
func (r *ProjectRepository) UpdatePhase(ctx context.Context, projectID uuid.UUID, phase domain.ProjectPhase) error {
	updates := map[string]interface{}{
		"phase": phase,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update project phase: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// SetWinningOffer sets the winning offer for a project and transitions it to active phase.
// Also conditionally propagates the offer's customer, responsible user, description, and location
// to the project (only if those fields are not already set on the project).
func (r *ProjectRepository) SetWinningOffer(ctx context.Context, projectID uuid.UUID, offerID uuid.UUID, inheritedOfferNumber string, offerValue float64, offerCost float64, customerID uuid.UUID, customerName string, managerID string, managerName string, description string, location string, wonAt time.Time) error {
	// Start with fields that are always set
	updates := map[string]interface{}{
		"phase":                  domain.ProjectPhaseActive,
		"winning_offer_id":       offerID,
		"inherited_offer_number": inheritedOfferNumber,
		"calculated_offer_value": offerValue,
		"value":                  offerValue,   // Set value from winning offer
		"cost":                   offerCost,    // Set cost from winning offer (margin_percent auto-calculated by trigger)
		"customer_id":            customerID,   // Propagate customer from winning offer
		"customer_name":          customerName, // Propagate customer name from winning offer
		"won_at":                 wonAt,
	}

	// Fetch current project to check which fields are empty
	var project domain.Project
	if err := r.db.WithContext(ctx).Where("id = ?", projectID).First(&project).Error; err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	// Only inherit manager if not already set
	if project.ManagerID == nil || *project.ManagerID == "" {
		updates["manager_id"] = managerID
		updates["manager_name"] = managerName
	}

	// Only inherit description if not already set
	if project.Description == "" && description != "" {
		updates["description"] = description
	}

	// Only inherit location if not already set
	if project.Location == "" && location != "" {
		updates["location"] = location
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to set winning offer: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// UpdateCalculatedOfferValue updates the calculated offer value for projects in tilbud phase
// This should be called when offers in a project are modified to keep the value in sync
func (r *ProjectRepository) UpdateCalculatedOfferValue(ctx context.Context, projectID uuid.UUID, value float64) error {
	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ? AND phase = ?", projectID, domain.ProjectPhaseTilbud)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Updates(map[string]interface{}{
		"calculated_offer_value": value,
		"budget":                 value, // Keep budget in sync during tilbud phase
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update calculated offer value: %w", result.Error)
	}

	// Note: not checking RowsAffected because project might not be in tilbud phase
	return nil
}

// GetByPhase returns all projects in a specific phase
func (r *ProjectRepository) GetByPhase(ctx context.Context, phase domain.ProjectPhase) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("phase = ?", phase)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// CountByPhase returns the count of projects grouped by phase
func (r *ProjectRepository) CountByPhase(ctx context.Context) (map[domain.ProjectPhase]int64, error) {
	type phaseCount struct {
		Phase domain.ProjectPhase
		Count int64
	}

	var results []phaseCount
	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Select("phase, COUNT(*) as count").
		Group("phase")
	query = ApplyCompanyFilter(ctx, query)
	err := query.Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count projects by phase: %w", err)
	}

	counts := make(map[domain.ProjectPhase]int64)
	for _, r := range results {
		counts[r.Phase] = r.Count
	}

	return counts, nil
}

// CancelProject updates the project phase to cancelled
func (r *ProjectRepository) CancelProject(ctx context.Context, projectID uuid.UUID) error {
	updates := map[string]interface{}{
		"phase": domain.ProjectPhaseCancelled,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	query = ApplyCompanyFilter(ctx, query)
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to cancel project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// RecalculateBestOfferEconomics recalculates the project's CalculatedOfferValue and Budget
// based on the highest value active offer linked to the project.
// This should only be called for projects in the tilbud phase.
func (r *ProjectRepository) RecalculateBestOfferEconomics(ctx context.Context, projectID uuid.UUID) error {
	// Find the highest value active offer for this project
	var maxValue float64
	err := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Where("project_id = ?", projectID).
		Where("phase NOT IN ?", []domain.OfferPhase{
			domain.OfferPhaseWon,
			domain.OfferPhaseLost,
			domain.OfferPhaseExpired,
		}).
		Select("COALESCE(MAX(value), 0)").
		Scan(&maxValue).Error

	if err != nil {
		return fmt.Errorf("failed to calculate best offer value: %w", err)
	}

	// Update project with the calculated value
	updates := map[string]interface{}{
		"calculated_offer_value": maxValue,
		"budget":                 maxValue,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID).
		Where("phase = ?", domain.ProjectPhaseTilbud) // Only update tilbud phase projects
	query = ApplyCompanyFilter(ctx, query)
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update project economics: %w", result.Error)
	}

	return nil
}

// UpdateCustomerFromOffers updates the project's customer based on active offers.
// If all active offers (in_progress/sent) are to the same customer, set that customer.
// If active offers are to different customers, set CustomerID to NULL.
// This should only be called for projects in the tilbud phase.
func (r *ProjectRepository) UpdateCustomerFromOffers(ctx context.Context, projectID uuid.UUID, customerID *uuid.UUID, customerName string) error {
	updates := map[string]interface{}{
		"customer_id":   customerID,
		"customer_name": customerName,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID).
		Where("phase = ?", domain.ProjectPhaseTilbud) // Only update tilbud phase projects
	query = ApplyCompanyFilter(ctx, query)
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update project customer: %w", result.Error)
	}

	return nil
}
