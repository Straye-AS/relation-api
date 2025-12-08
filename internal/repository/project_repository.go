package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ProjectFilters defines filter options for project listing
type ProjectFilters struct {
	CustomerID *uuid.UUID
	Status     *domain.ProjectStatus
	Health     *domain.ProjectHealth
	ManagerID  *string
}

// ProjectBudgetMetrics holds calculated budget metrics for a project
type ProjectBudgetMetrics struct {
	Budget      float64
	Spent       float64
	Remaining   float64
	PercentUsed float64
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
func (r *ProjectRepository) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID, status *domain.ProjectStatus) ([]domain.Project, int64, error) {
	filters := &ProjectFilters{
		CustomerID: customerID,
		Status:     status,
	}
	return r.ListWithFilters(ctx, page, pageSize, filters)
}

// ListWithFilters returns a paginated list of projects with filter options including health and managerID
func (r *ProjectRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *ProjectFilters) ([]domain.Project, int64, error) {
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

		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
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

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&projects).Error

	return projects, total, err
}

func (r *ProjectRepository) CountActive(ctx context.Context) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("status = ?", domain.ProjectStatusActive)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return int(count), err
}

func (r *ProjectRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).Preload("Customer").
		Where("LOWER(name) LIKE ? OR LOWER(summary) LIKE ?", searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Find(&projects).Error
	return projects, err
}

// GetByIDWithRelations retrieves a project by ID with all related data
// Preloads: Customer, Offer, Deal, BudgetDimensions, Activities
func (r *ProjectRepository) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*domain.Project, []domain.BudgetDimension, []domain.Activity, error) {
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

	// Fetch budget dimensions separately (polymorphic relationship)
	var dimensions []domain.BudgetDimension
	err = r.db.WithContext(ctx).
		Preload("Category").
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentProject, id).
		Order("display_order ASC, created_at ASC").
		Find(&dimensions).Error
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load budget dimensions: %w", err)
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

	return &project, dimensions, activities, nil
}

// CalculateBudgetMetrics calculates budget metrics for a project
// Returns: budget, spent, remaining, percent_used
// If hasDetailedBudget is true, spent is calculated from BudgetDimensions cost sum
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
		Budget: project.Budget,
	}

	// Calculate spent based on whether project uses detailed budget
	if project.HasDetailedBudget {
		// Sum cost from budget dimensions
		var totalCost float64
		err = r.db.WithContext(ctx).
			Model(&domain.BudgetDimension{}).
			Where("parent_type = ? AND parent_id = ?", domain.BudgetParentProject, projectID).
			Select("COALESCE(SUM(cost), 0)").
			Scan(&totalCost).Error
		if err != nil {
			return nil, fmt.Errorf("failed to calculate budget dimension costs: %w", err)
		}
		metrics.Spent = totalCost
	} else {
		// Use the project's Spent field directly
		metrics.Spent = project.Spent
	}

	// Calculate remaining and percent used
	metrics.Remaining = metrics.Budget - metrics.Spent
	if metrics.Budget > 0 {
		metrics.PercentUsed = (metrics.Spent / metrics.Budget) * 100
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

	// If budget is 0, default to on_track to avoid division issues
	if metrics.Budget <= 0 {
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

// GetBudgetSummary returns aggregated budget totals for a project's budget dimensions
func (r *ProjectRepository) GetBudgetSummary(ctx context.Context, projectID uuid.UUID) (*domain.BudgetSummary, error) {
	var result struct {
		TotalCost      float64
		TotalRevenue   float64
		DimensionCount int
	}

	err := r.db.WithContext(ctx).
		Model(&domain.BudgetDimension{}).
		Where("parent_type = ? AND parent_id = ?", domain.BudgetParentProject, projectID).
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
		Where("status = ?", domain.ProjectStatusActive). // Only count active projects
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
