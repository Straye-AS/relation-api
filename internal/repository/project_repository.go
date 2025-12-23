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
// Projects are now cross-company, so no CompanyID filter is needed.
type ProjectFilters struct {
	CustomerID *uuid.UUID
	Phase      *domain.ProjectPhase
}

// projectSortableFields maps API field names to database column names for projects
// Only fields in this map can be used for sorting (whitelist approach)
var projectSortableFields = map[string]string{
	"createdAt":    "created_at",
	"updatedAt":    "updated_at",
	"name":         "name",
	"phase":        "phase",
	"startDate":    "start_date",
	"endDate":      "end_date",
	"customerName": "customer_name",
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
	// Note: No company filter - projects are cross-company
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
// Projects are cross-company, so no company filter is applied.
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
	// Note: No company filter - projects are cross-company

	// Apply filters
	if filters != nil {
		if filters.CustomerID != nil {
			query = query.Where("customer_id = ?", *filters.CustomerID)
		}

		if filters.Phase != nil {
			query = query.Where("phase = ?", *filters.Phase)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build order clause from sort config
	orderClause := BuildOrderClause(sort, projectSortableFields, "updated_at")

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&projects).Error

	return projects, total, err
}

func (r *ProjectRepository) CountActive(ctx context.Context) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("phase IN (?)", []domain.ProjectPhase{domain.ProjectPhaseWorking, domain.ProjectPhaseTilbud, domain.ProjectPhaseOnHold})
	// Note: No company filter - projects are cross-company
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
	// Note: No company filter - projects are cross-company
	err := query.Order("updated_at DESC").Limit(limit).Find(&projects).Error
	return projects, err
}

// GetByIDWithRelations retrieves a project by ID with all related data
// Preloads: Customer, Offer, Deal
// Also returns BudgetItems and Activities which are polymorphic relationships
func (r *ProjectRepository) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*domain.Project, []domain.BudgetItem, []domain.Activity, error) {
	var project domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Deal").
		Where("id = ?", id)
	// Note: No company filter - projects are cross-company
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

// GetActiveProjects returns active projects with optional limit
func (r *ProjectRepository) GetActiveProjects(ctx context.Context, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Where("phase IN (?)", []domain.ProjectPhase{domain.ProjectPhaseWorking, domain.ProjectPhaseTilbud, domain.ProjectPhaseOnHold}).
		Order("updated_at DESC").
		Limit(limit)
	// Note: No company filter - projects are cross-company
	err := query.Find(&projects).Error
	return projects, err
}

// GetRecentProjects returns the most recently updated projects
func (r *ProjectRepository) GetRecentProjects(ctx context.Context, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Order("updated_at DESC").
		Limit(limit)
	// Note: No company filter - projects are cross-company
	err := query.Find(&projects).Error
	return projects, err
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
	query = query.Order("updated_at DESC").Limit(limit)
	// Note: No company filter - projects are cross-company
	err := query.Find(&projects).Error
	return projects, err
}

// ============================================================================
// Phase Management Methods
// ============================================================================

// UpdatePhase updates the project phase
func (r *ProjectRepository) UpdatePhase(ctx context.Context, projectID uuid.UUID, phase domain.ProjectPhase) error {
	updates := map[string]interface{}{
		"phase": phase,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	// Note: No company filter - projects are cross-company
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update project phase: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetByPhase returns all projects in a specific phase
func (r *ProjectRepository) GetByPhase(ctx context.Context, phase domain.ProjectPhase) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("phase = ?", phase)
	// Note: No company filter - projects are cross-company
	err := query.Order("updated_at DESC").Find(&projects).Error
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
	// Note: No company filter - projects are cross-company
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
	// Note: No company filter - projects are cross-company
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to cancel project: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// UpdateCustomer updates the project's customer information
// This is used when offers are linked/unlinked to update the denormalized customer fields.
// Pass nil for customerID to clear the customer association.
func (r *ProjectRepository) UpdateCustomer(ctx context.Context, projectID uuid.UUID, customerID *uuid.UUID, customerName string) error {
	updates := map[string]interface{}{
		"customer_id":   customerID,
		"customer_name": customerName,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	// Note: No company filter - projects are cross-company
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update project customer: %w", result.Error)
	}

	return nil
}

// UpdateLocation updates the project's location field
// This is used when offers are linked/unlinked to infer location from offers.
// Pass empty string to clear the location.
func (r *ProjectRepository) UpdateLocation(ctx context.Context, projectID uuid.UUID, location string) error {
	updates := map[string]interface{}{
		"location": location,
	}

	query := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("id = ?", projectID)
	// Note: No company filter - projects are cross-company
	result := query.Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update project location: %w", result.Error)
	}

	return nil
}

// GetByCustomerID returns all projects for a specific customer
func (r *ProjectRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]domain.Project, error) {
	var projects []domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("customer_id = ?", customerID)
	// Note: No company filter - projects are cross-company
	err := query.Order("updated_at DESC").Find(&projects).Error
	return projects, err
}

// ExistsByProjectNumber checks if a project with the given project number exists
func (r *ProjectRepository) ExistsByProjectNumber(ctx context.Context, projectNumber string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("project_number = ?", projectNumber).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check project number existence: %w", err)
	}
	return count > 0, nil
}

// GetByProjectNumber retrieves a project by its project number
func (r *ProjectRepository) GetByProjectNumber(ctx context.Context, projectNumber string) (*domain.Project, error) {
	var project domain.Project
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Where("project_number = ?", projectNumber)
	// Note: No company filter - projects are cross-company
	err := query.First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetHighestProjectNumber finds the highest project number for a given year
// Assumes format PROJECT-YYYY-NNN
func (r *ProjectRepository) GetHighestProjectNumber(ctx context.Context, year int) (string, error) {
	var projectNumber string
	prefix := fmt.Sprintf("PROJECT-%d-%%", year)

	err := r.db.WithContext(ctx).
		Model(&domain.Project{}).
		Where("project_number LIKE ?", prefix).
		Order("length(project_number) DESC, project_number DESC").
		Limit(1).
		Pluck("project_number", &projectNumber).Error

	if err != nil {
		return "", err
	}
	return projectNumber, nil
}
