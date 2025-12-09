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

// CustomerSortOption defines sorting options for customer listing
type CustomerSortOption string

const (
	CustomerSortByNameAsc     CustomerSortOption = "name_asc"
	CustomerSortByNameDesc    CustomerSortOption = "name_desc"
	CustomerSortByCreatedDesc CustomerSortOption = "created_desc"
	CustomerSortByCreatedAsc  CustomerSortOption = "created_asc"
	CustomerSortByCityAsc     CustomerSortOption = "city_asc"
	CustomerSortByCityDesc    CustomerSortOption = "city_desc"
)

// CustomerFilters defines filter options for customer listing
type CustomerFilters struct {
	Search   string
	City     string
	Country  string
	Status   *domain.CustomerStatus
	Tier     *domain.CustomerTier
	Industry *domain.CustomerIndustry
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
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&customer).Error
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
	filters := &CustomerFilters{Search: search}
	return r.ListWithFilters(ctx, page, pageSize, filters, CustomerSortByCreatedDesc)
}

// ListWithFilters returns a paginated list of customers with filter and sort options
func (r *CustomerRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *CustomerFilters, sortBy CustomerSortOption) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Customer{})

	// Exclude inactive customers by default unless explicitly filtering by status
	if filters == nil || filters.Status == nil {
		query = query.Where("status != ?", domain.CustomerStatusInactive)
	}

	// Apply filters
	if filters != nil {
		if filters.Search != "" {
			searchPattern := "%" + strings.ToLower(filters.Search) + "%"
			query = query.Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern)
		}
		if filters.City != "" {
			query = query.Where("LOWER(city) = LOWER(?)", filters.City)
		}
		if filters.Country != "" {
			query = query.Where("LOWER(country) = LOWER(?)", filters.Country)
		}
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.Tier != nil {
			query = query.Where("tier = ?", *filters.Tier)
		}
		if filters.Industry != nil {
			query = query.Where("industry = ?", *filters.Industry)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := r.getSortClause(sortBy)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&customers).Error

	return customers, total, err
}

// getSortClause returns the SQL ORDER BY clause for the given sort option
func (r *CustomerRepository) getSortClause(sortBy CustomerSortOption) string {
	switch sortBy {
	case CustomerSortByNameAsc:
		return "name ASC"
	case CustomerSortByNameDesc:
		return "name DESC"
	case CustomerSortByCreatedAsc:
		return "created_at ASC"
	case CustomerSortByCityAsc:
		return "city ASC"
	case CustomerSortByCityDesc:
		return "city DESC"
	case CustomerSortByCreatedDesc:
		fallthrough
	default:
		return "created_at DESC"
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
	err := r.db.WithContext(ctx).Model(&domain.Customer{}).
		Where("status != ?", domain.CustomerStatusInactive).
		Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Customer, error) {
	var customers []domain.Customer
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	err := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern).
		Where("status != ?", domain.CustomerStatusInactive).
		Limit(limit).Find(&customers).Error
	return customers, err
}

// GetByOrgNumber finds a customer by organization number
func (r *CustomerRepository) GetByOrgNumber(ctx context.Context, orgNumber string) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).Where("org_number = ?", orgNumber).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// CustomerStats holds aggregated statistics for a customer
type CustomerStats struct {
	TotalValue     float64 `json:"totalValue"`
	ActiveOffers   int     `json:"activeOffers"`
	ActiveDeals    int     `json:"activeDeals"`
	ActiveProjects int     `json:"activeProjects"`
	TotalContacts  int     `json:"totalContacts"`
}

// GetCustomerStats returns aggregated statistics for a customer
func (r *CustomerRepository) GetCustomerStats(ctx context.Context, customerID uuid.UUID) (*CustomerStats, error) {
	stats := &CustomerStats{}

	// Get active offers count and total value
	var offerStats struct {
		Count      int64
		TotalValue float64
	}
	err := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Select("COUNT(*) as count, COALESCE(SUM(value), 0) as total_value").
		Where("customer_id = ? AND status = ?", customerID, domain.OfferStatusActive).
		Scan(&offerStats).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveOffers = int(offerStats.Count)
	stats.TotalValue = offerStats.TotalValue

	// Get active deals count
	var dealsCount int64
	err = r.db.WithContext(ctx).Model(&domain.Deal{}).
		Where("customer_id = ? AND stage NOT IN (?, ?)", customerID, domain.DealStageWon, domain.DealStageLost).
		Count(&dealsCount).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveDeals = int(dealsCount)

	// Get active projects count
	var projectsCount int64
	err = r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ? AND status IN (?, ?)", customerID, domain.ProjectStatusPlanning, domain.ProjectStatusActive).
		Count(&projectsCount).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveProjects = int(projectsCount)

	// Get total contacts count
	var contactsCount int64
	err = r.db.WithContext(ctx).Model(&domain.Contact{}).
		Where("primary_customer_id = ?", customerID).
		Count(&contactsCount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalContacts = int(contactsCount)

	return stats, nil
}

// GetCustomerWithRelations returns a customer with preloaded contacts
func (r *CustomerRepository) GetCustomerWithRelations(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	var customer domain.Customer
	err := r.db.WithContext(ctx).
		Preload("Contacts", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Limit(10)
		}).
		Where("id = ?", id).
		First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// HasActiveRelations checks if a customer has active projects, offers, or deals
func (r *CustomerRepository) HasActiveRelations(ctx context.Context, customerID uuid.UUID) (bool, string, error) {
	// Check for active projects
	var projectCount int64
	err := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ? AND status IN (?, ?)", customerID, domain.ProjectStatusPlanning, domain.ProjectStatusActive).
		Count(&projectCount).Error
	if err != nil {
		return false, "", err
	}
	if projectCount > 0 {
		return true, "customer has active projects", nil
	}

	// Check for active deals
	var dealCount int64
	err = r.db.WithContext(ctx).Model(&domain.Deal{}).
		Where("customer_id = ? AND stage NOT IN (?, ?)", customerID, domain.DealStageWon, domain.DealStageLost).
		Count(&dealCount).Error
	if err != nil {
		return false, "", err
	}
	if dealCount > 0 {
		return true, "customer has active deals", nil
	}

	// Check for active offers
	var offerCount int64
	err = r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("customer_id = ? AND status = ?", customerID, domain.OfferStatusActive).
		Count(&offerCount).Error
	if err != nil {
		return false, "", err
	}
	if offerCount > 0 {
		return true, "customer has active offers", nil
	}

	return false, "", nil
}

// GetTopCustomers returns customers with the most active offers/projects
func (r *CustomerRepository) GetTopCustomers(ctx context.Context, limit int) ([]domain.Customer, error) {
	var customers []domain.Customer
	query := r.db.WithContext(ctx).
		Where("status = ?", domain.CustomerStatusActive).
		Order("updated_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&customers).Error
	return customers, err
}

// TopCustomerWithStats holds customer data with offer statistics
type TopCustomerWithStats struct {
	CustomerID    uuid.UUID
	CustomerName  string
	OrgNumber     string
	OfferCount    int
	EconomicValue float64
}

// GetTopCustomersWithOfferStats returns top customers ranked by offer count within a time window
// Excludes draft and expired offers from the counts
func (r *CustomerRepository) GetTopCustomersWithOfferStats(ctx context.Context, since time.Time, limit int) ([]TopCustomerWithStats, error) {
	// Valid phases for counting (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseWon,
		domain.OfferPhaseLost,
	}

	var results []TopCustomerWithStats

	// Build subquery to get offer stats per customer
	// Then join with customers to get customer info
	query := r.db.WithContext(ctx).
		Table("offers").
		Select("customers.id as customer_id, customers.name as customer_name, customers.org_number, COUNT(offers.id) as offer_count, COALESCE(SUM(offers.value), 0) as economic_value").
		Joins("JOIN customers ON customers.id = offers.customer_id").
		Where("offers.created_at >= ?", since).
		Where("offers.phase IN ?", validPhases).
		Where("customers.status != ?", domain.CustomerStatusInactive).
		Group("customers.id, customers.name, customers.org_number").
		Order("offer_count DESC, economic_value DESC").
		Limit(limit)

	// Apply company filter on offers table
	query = ApplyCompanyFilterWithAlias(ctx, query, "offers")

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get top customers with offer stats: %w", err)
	}

	return results, nil
}
