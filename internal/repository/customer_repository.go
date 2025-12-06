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

// CustomerFilters contains all filter options for listing customers
type CustomerFilters struct {
	CompanyID     *domain.CompanyID
	City          *string
	Country       *string
	SearchQuery   *string // Search in name and org_number
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// CustomerSortOption represents available sort options
type CustomerSortOption string

const (
	CustomerSortByNameAsc     CustomerSortOption = "name_asc"
	CustomerSortByNameDesc    CustomerSortOption = "name_desc"
	CustomerSortByCreatedDesc CustomerSortOption = "created_desc"
	CustomerSortByCreatedAsc  CustomerSortOption = "created_asc"
	CustomerSortByUpdatedDesc CustomerSortOption = "updated_desc"
	CustomerSortByUpdatedAsc  CustomerSortOption = "updated_asc"
	CustomerSortByCityAsc     CustomerSortOption = "city_asc"
	CustomerSortByCityDesc    CustomerSortOption = "city_desc"
)

// CustomerStats holds aggregated statistics for a customer
type CustomerStats struct {
	ActiveDealsCount    int64
	TotalDealsCount     int64
	TotalDealValue      float64
	WonDealsValue       float64 // Lifetime value
	ActiveProjectsCount int64
}

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	var customer domain.Customer
	query := r.db.WithContext(ctx).Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Customer{}, "id = ?", id).Error
}

// GetByOrgNumber retrieves a customer by their organization number
// Returns nil, nil if no customer is found (not an error condition)
// Used for uniqueness validation when creating/updating customers
func (r *CustomerRepository) GetByOrgNumber(ctx context.Context, orgNumber string) (*domain.Customer, error) {
	var customer domain.Customer
	query := r.db.WithContext(ctx).Where("org_number = ?", orgNumber)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&customer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get customer by org number: %w", err)
	}
	return &customer, nil
}

// List retrieves customers with optional filtering and sorting
// If filters is nil, returns all customers
// If sortBy is empty, defaults to created_at DESC
func (r *CustomerRepository) List(ctx context.Context, page, pageSize int, filters *CustomerFilters, sortBy CustomerSortOption) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Customer{})

	// Apply multi-tenant company filter from context
	query = ApplyCompanyFilter(ctx, query)

	// Apply additional filters
	query = r.applyFilters(query, filters)

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	// Apply sorting
	query = r.applySorting(query, sortBy)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Find(&customers).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}

	return customers, total, nil
}

// applyFilters applies all filter criteria to the query
func (r *CustomerRepository) applyFilters(query *gorm.DB, filters *CustomerFilters) *gorm.DB {
	if filters == nil {
		return query
	}

	if filters.CompanyID != nil {
		query = query.Where("company_id = ?", *filters.CompanyID)
	}

	if filters.City != nil && *filters.City != "" {
		query = query.Where("LOWER(city) = LOWER(?)", *filters.City)
	}

	if filters.Country != nil && *filters.Country != "" {
		query = query.Where("LOWER(country) = LOWER(?)", *filters.Country)
	}

	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		searchPattern := "%" + strings.ToLower(*filters.SearchQuery) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern)
	}

	if filters.CreatedAfter != nil {
		query = query.Where("created_at >= ?", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		query = query.Where("created_at <= ?", *filters.CreatedBefore)
	}

	return query
}

// applySorting applies the sorting option to the query
func (r *CustomerRepository) applySorting(query *gorm.DB, sortBy CustomerSortOption) *gorm.DB {
	switch sortBy {
	case CustomerSortByNameAsc:
		return query.Order("name ASC")
	case CustomerSortByNameDesc:
		return query.Order("name DESC")
	case CustomerSortByCreatedAsc:
		return query.Order("created_at ASC")
	case CustomerSortByUpdatedDesc:
		return query.Order("updated_at DESC")
	case CustomerSortByUpdatedAsc:
		return query.Order("updated_at ASC")
	case CustomerSortByCityAsc:
		return query.Order("city ASC NULLS LAST")
	case CustomerSortByCityDesc:
		return query.Order("city DESC NULLS LAST")
	default: // CustomerSortByCreatedDesc
		return query.Order("created_at DESC")
	}
}

func (r *CustomerRepository) GetContactsCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Contact{}).Where("customer_id = ?", customerID).Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) GetProjectsCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Project{}).Where("customer_id = ?", customerID).Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) GetOffersCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("customer_id = ?", customerID).Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) Count(ctx context.Context) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Customer{})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Customer, error) {
	var customers []domain.Customer
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Find(&customers).Error
	return customers, err
}

// GetCustomerStats retrieves aggregated statistics for a customer
// This includes deal counts, values, and project counts
func (r *CustomerRepository) GetCustomerStats(ctx context.Context, customerID uuid.UUID) (*CustomerStats, error) {
	stats := &CustomerStats{}

	// Get deal statistics in a single query
	type dealResult struct {
		TotalCount  int64
		ActiveCount int64
		TotalValue  float64
		WonValue    float64
	}
	var dealStats dealResult

	dealQuery := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Select(`
			COUNT(*) as total_count,
			COUNT(CASE WHEN stage NOT IN ('won', 'lost') THEN 1 END) as active_count,
			COALESCE(SUM(value), 0) as total_value,
			COALESCE(SUM(CASE WHEN stage = 'won' THEN value ELSE 0 END), 0) as won_value
		`).
		Where("customer_id = ?", customerID)
	dealQuery = ApplyCompanyFilter(ctx, dealQuery)

	if err := dealQuery.Scan(&dealStats).Error; err != nil {
		return nil, fmt.Errorf("get deal stats: %w", err)
	}

	stats.TotalDealsCount = dealStats.TotalCount
	stats.ActiveDealsCount = dealStats.ActiveCount
	stats.TotalDealValue = dealStats.TotalValue
	stats.WonDealsValue = dealStats.WonValue

	// Get active projects count
	projectQuery := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ?", customerID).
		Where("status IN ?", []domain.ProjectStatus{domain.ProjectStatusPlanning, domain.ProjectStatusActive, domain.ProjectStatusOnHold})
	projectQuery = ApplyCompanyFilterWithColumn(ctx, projectQuery, "company_id")

	if err := projectQuery.Count(&stats.ActiveProjectsCount).Error; err != nil {
		return nil, fmt.Errorf("get active projects count: %w", err)
	}

	return stats, nil
}

// HasActiveDealsOrProjects checks if a customer has any active deals or projects
// that would prevent deletion
func (r *CustomerRepository) HasActiveDealsOrProjects(ctx context.Context, customerID uuid.UUID) (bool, string, error) {
	// Check for active deals
	var activeDealsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Where("customer_id = ? AND stage NOT IN ('won', 'lost')", customerID).
		Count(&activeDealsCount).Error; err != nil {
		return false, "", err
	}
	if activeDealsCount > 0 {
		return true, "customer has active deals", nil
	}

	// Check for active projects
	var activeProjectsCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ? AND status IN ('planning', 'active', 'on_hold')", customerID).
		Count(&activeProjectsCount).Error; err != nil {
		return false, "", err
	}
	if activeProjectsCount > 0 {
		return true, "customer has active projects", nil
	}

	return false, "", nil
}
