package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

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

func (r *CustomerRepository) List(ctx context.Context, page, pageSize int, search string) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Customer{})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&customers).Error

	return customers, total, err
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

// GetByOrgNumber retrieves a customer by their organization number
func (r *CustomerRepository) GetByOrgNumber(ctx context.Context, orgNumber string) (*domain.Customer, error) {
	var customer domain.Customer
	query := r.db.WithContext(ctx).Where("org_number = ?", orgNumber)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// CustomerStats holds aggregated statistics for a customer
type CustomerStats struct {
	ActiveDealsCount    int64
	TotalDealsCount     int64
	TotalDealValue      float64
	WonDealsValue       float64
	ActiveProjectsCount int64
	TotalProjectsCount  int64
}

// GetCustomerStats retrieves aggregated statistics for a customer
func (r *CustomerRepository) GetCustomerStats(ctx context.Context, customerID uuid.UUID) (*CustomerStats, error) {
	stats := &CustomerStats{}

	// Get deal statistics
	dealStatsQuery := `
		SELECT
			COUNT(*) FILTER (WHERE stage NOT IN ('won', 'lost')) as active_deals_count,
			COUNT(*) as total_deals_count,
			COALESCE(SUM(value), 0) as total_deal_value,
			COALESCE(SUM(value) FILTER (WHERE stage = 'won'), 0) as won_deals_value
		FROM deals
		WHERE customer_id = ?
	`
	var dealStats struct {
		ActiveDealsCount int64
		TotalDealsCount  int64
		TotalDealValue   float64
		WonDealsValue    float64
	}
	if err := r.db.WithContext(ctx).Raw(dealStatsQuery, customerID).Scan(&dealStats).Error; err != nil {
		return nil, err
	}
	stats.ActiveDealsCount = dealStats.ActiveDealsCount
	stats.TotalDealsCount = dealStats.TotalDealsCount
	stats.TotalDealValue = dealStats.TotalDealValue
	stats.WonDealsValue = dealStats.WonDealsValue

	// Get project statistics
	projectStatsQuery := `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('planning', 'active')) as active_projects_count,
			COUNT(*) as total_projects_count
		FROM projects
		WHERE customer_id = ?
	`
	var projectStats struct {
		ActiveProjectsCount int64
		TotalProjectsCount  int64
	}
	if err := r.db.WithContext(ctx).Raw(projectStatsQuery, customerID).Scan(&projectStats).Error; err != nil {
		return nil, err
	}
	stats.ActiveProjectsCount = projectStats.ActiveProjectsCount
	stats.TotalProjectsCount = projectStats.TotalProjectsCount

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
