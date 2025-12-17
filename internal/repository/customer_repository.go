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
// Deprecated: Use SortConfig instead for new code
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

// customerSortableFields maps API field names to database column names for customers
// Only fields in this map can be used for sorting (whitelist approach)
var customerSortableFields = map[string]string{
	"createdAt": "created_at",
	"updatedAt": "updated_at",
	"name":      "name",
	"city":      "city",
	"country":   "country",
	"status":    "status",
	"tier":      "tier",
	"industry":  "industry",
	"orgNumber": "org_number",
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
	return r.ListWithSortConfig(ctx, page, pageSize, filters, DefaultSortConfig())
}

// ListWithFilters returns a paginated list of customers with filter and sort options
// Deprecated: Use ListWithSortConfig for new code
func (r *CustomerRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *CustomerFilters, sortBy CustomerSortOption) ([]domain.Customer, int64, error) {
	// Convert legacy sort option to SortConfig
	sort := r.convertLegacySortOption(sortBy)
	return r.ListWithSortConfig(ctx, page, pageSize, filters, sort)
}

// ListWithSortConfig returns a paginated list of customers with filter and sort options using SortConfig
func (r *CustomerRepository) ListWithSortConfig(ctx context.Context, page, pageSize int, filters *CustomerFilters, sort SortConfig) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
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

	// Build order clause from sort config
	orderClause := BuildOrderClause(sort, customerSortableFields, "updated_at")

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&customers).Error

	return customers, total, err
}

// convertLegacySortOption converts legacy CustomerSortOption to SortConfig
func (r *CustomerRepository) convertLegacySortOption(sortBy CustomerSortOption) SortConfig {
	switch sortBy {
	case CustomerSortByNameAsc:
		return SortConfig{Field: "name", Order: SortOrderAsc}
	case CustomerSortByNameDesc:
		return SortConfig{Field: "name", Order: SortOrderDesc}
	case CustomerSortByCreatedAsc:
		return SortConfig{Field: "createdAt", Order: SortOrderAsc}
	case CustomerSortByCityAsc:
		return SortConfig{Field: "city", Order: SortOrderAsc}
	case CustomerSortByCityDesc:
		return SortConfig{Field: "city", Order: SortOrderDesc}
	case CustomerSortByCreatedDesc:
		fallthrough
	default:
		return DefaultSortConfig()
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
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &customer, nil
}

// CustomerStats holds aggregated statistics for a customer
type CustomerStats struct {
	TotalValueActive float64 `json:"totalValueActive"` // Value of offers in order phase (active orders)
	TotalValueWon    float64 `json:"totalValueWon"`    // Value of offers in order or completed phases
	WorkingOffers    int     `json:"workingOffers"`    // Count of offers in in_progress or sent phases
	ActiveOffers     int     `json:"activeOffers"`     // Count of offers in order phase (active orders)
	CompletedOffers  int     `json:"completedOffers"`  // Count of offers in completed phase
	TotalOffers      int     `json:"totalOffers"`
	ActiveDeals      int     `json:"activeDeals"`
	ActiveProjects   int     `json:"activeProjects"`
	TotalProjects    int     `json:"totalProjects"`
	TotalContacts    int     `json:"totalContacts"`
}

// GetCustomerStats returns aggregated statistics for a customer
// Respects the X-Company-Id header for filtering offers, projects, and deals
func (r *CustomerRepository) GetCustomerStats(ctx context.Context, customerID uuid.UUID) (*CustomerStats, error) {
	stats := &CustomerStats{}

	// Get working offers count (in_progress or sent phases)
	var workingOffersCount int64
	workingOfferQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("customer_id = ? AND phase IN ?", customerID, []domain.OfferPhase{domain.OfferPhaseInProgress, domain.OfferPhaseSent})
	workingOfferQuery = ApplyCompanyFilter(ctx, workingOfferQuery)
	err := workingOfferQuery.Count(&workingOffersCount).Error
	if err != nil {
		return nil, err
	}
	stats.WorkingOffers = int(workingOffersCount)

	// Get active offers count and value (order phase only - active orders)
	var activeOfferStats struct {
		Count      int64
		TotalValue float64
	}
	activeOfferQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Select("COUNT(*) as count, COALESCE(SUM(value), 0) as total_value").
		Where("customer_id = ? AND phase = ?", customerID, domain.OfferPhaseOrder)
	activeOfferQuery = ApplyCompanyFilter(ctx, activeOfferQuery)
	err = activeOfferQuery.Scan(&activeOfferStats).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveOffers = int(activeOfferStats.Count)
	stats.TotalValueActive = activeOfferStats.TotalValue

	// Get completed offers count (completed phase)
	var completedOffersCount int64
	completedOfferQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("customer_id = ? AND phase = ?", customerID, domain.OfferPhaseCompleted)
	completedOfferQuery = ApplyCompanyFilter(ctx, completedOfferQuery)
	err = completedOfferQuery.Count(&completedOffersCount).Error
	if err != nil {
		return nil, err
	}
	stats.CompletedOffers = int(completedOffersCount)

	// Get won offers total value (order or completed phases)
	var wonValue float64
	wonOfferQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Select("COALESCE(SUM(value), 0)").
		Where("customer_id = ? AND phase IN ?", customerID, []domain.OfferPhase{domain.OfferPhaseOrder, domain.OfferPhaseCompleted})
	wonOfferQuery = ApplyCompanyFilter(ctx, wonOfferQuery)
	err = wonOfferQuery.Scan(&wonValue).Error
	if err != nil {
		return nil, err
	}
	stats.TotalValueWon = wonValue

	// Get total offers count (all offers for this customer)
	var totalOffersCount int64
	totalOffersQuery := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("customer_id = ?", customerID)
	totalOffersQuery = ApplyCompanyFilter(ctx, totalOffersQuery)
	err = totalOffersQuery.Count(&totalOffersCount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalOffers = int(totalOffersCount)

	// Get active deals count
	var dealsCount int64
	dealsQuery := r.db.WithContext(ctx).Model(&domain.Deal{}).
		Where("customer_id = ? AND stage NOT IN (?, ?)", customerID, domain.DealStageWon, domain.DealStageLost)
	dealsQuery = ApplyCompanyFilter(ctx, dealsQuery)
	err = dealsQuery.Count(&dealsCount).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveDeals = int(dealsCount)

	// Get active projects count (tilbud, working, or on_hold phases)
	var activeProjectsCount int64
	activeProjectsQuery := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ? AND phase IN (?)", customerID, []domain.ProjectPhase{domain.ProjectPhaseTilbud, domain.ProjectPhaseWorking, domain.ProjectPhaseOnHold})
	activeProjectsQuery = ApplyCompanyFilter(ctx, activeProjectsQuery)
	err = activeProjectsQuery.Count(&activeProjectsCount).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveProjects = int(activeProjectsCount)

	// Get total projects count (all projects for this customer)
	var totalProjectsCount int64
	totalProjectsQuery := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ?", customerID)
	totalProjectsQuery = ApplyCompanyFilter(ctx, totalProjectsQuery)
	err = totalProjectsQuery.Count(&totalProjectsCount).Error
	if err != nil {
		return nil, err
	}
	stats.TotalProjects = int(totalProjectsCount)

	// Get total contacts count
	// Note: Contacts are linked to customers which are global entities,
	// but we filter by company_id if the contact has one
	var contactsCount int64
	contactsQuery := r.db.WithContext(ctx).Model(&domain.Contact{}).
		Where("primary_customer_id = ?", customerID)
	contactsQuery = ApplyCompanyFilter(ctx, contactsQuery)
	err = contactsQuery.Count(&contactsCount).Error
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
	// Check for active projects (tilbud, working, or on_hold phases)
	var projectCount int64
	err := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("customer_id = ? AND phase IN (?)", customerID, []domain.ProjectPhase{domain.ProjectPhaseTilbud, domain.ProjectPhaseWorking, domain.ProjectPhaseOnHold}).
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

// FuzzySearchResult holds a customer with a similarity score
type FuzzySearchResult struct {
	Customer   domain.Customer
	Similarity float64
}

// FuzzySearchBestMatch finds the single best matching customer for a query using multiple strategies:
// 1. Exact match (case-insensitive)
// 2. Prefix match (query is start of name)
// 3. Contains match (query is substring of name)
// 4. Trigram similarity for typo tolerance
// 5. Abbreviation matching (AF -> AF Gruppen)
// Returns the best match with a confidence score (0-1)
func (r *CustomerRepository) FuzzySearchBestMatch(ctx context.Context, query string) (*FuzzySearchResult, error) {
	if query == "" {
		return nil, nil
	}

	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)
	queryPattern := "%" + queryLower + "%"

	// Check if query looks like an email - extract domain for matching
	var emailDomain string
	if strings.Contains(query, "@") {
		parts := strings.Split(queryLower, "@")
		if len(parts) == 2 && parts[1] != "" {
			// Extract company name from domain (e.g., "afgruppen.no" -> "afgruppen")
			domainParts := strings.Split(parts[1], ".")
			if len(domainParts) > 0 {
				emailDomain = domainParts[0]
			}
		}
	}

	// Strategy 0: Email domain match (if query is an email)
	// Match "hauk@straye.no" to "Straye", "test@afgruppen.no" to "AF Gruppen"
	if emailDomain != "" {
		// First try exact email match on customer email field
		var emailMatch domain.Customer
		err := r.db.WithContext(ctx).
			Where("LOWER(email) = ? AND status != ?", queryLower, domain.CustomerStatusInactive).
			First(&emailMatch).Error
		if err == nil {
			return &FuzzySearchResult{Customer: emailMatch, Similarity: 1.0}, nil
		}

		// Try matching domain to customer name
		var domainMatch domain.Customer
		err = r.db.WithContext(ctx).
			Where("LOWER(name) LIKE ? AND status != ?", "%"+emailDomain+"%", domain.CustomerStatusInactive).
			Order("LENGTH(name) ASC").
			First(&domainMatch).Error
		if err == nil {
			return &FuzzySearchResult{Customer: domainMatch, Similarity: 0.9}, nil
		}

		// Try matching domain to customer email domain
		var emailDomainMatch domain.Customer
		err = r.db.WithContext(ctx).
			Where("LOWER(email) LIKE ? AND status != ?", "%@"+emailDomain+"%", domain.CustomerStatusInactive).
			First(&emailDomainMatch).Error
		if err == nil {
			return &FuzzySearchResult{Customer: emailDomainMatch, Similarity: 0.9}, nil
		}
	}

	// Strategy 1: Exact match (highest confidence)
	var exactMatch domain.Customer
	err := r.db.WithContext(ctx).
		Where("LOWER(name) = ? AND status != ?", queryLower, domain.CustomerStatusInactive).
		First(&exactMatch).Error
	if err == nil {
		return &FuzzySearchResult{Customer: exactMatch, Similarity: 1.0}, nil
	}

	// Strategy 2: Exact match on org_number
	var orgMatch domain.Customer
	err = r.db.WithContext(ctx).
		Where("LOWER(org_number) = ? AND status != ?", queryLower, domain.CustomerStatusInactive).
		First(&orgMatch).Error
	if err == nil {
		return &FuzzySearchResult{Customer: orgMatch, Similarity: 1.0}, nil
	}

	// Strategy 3: Prefix match (name starts with query) - very high confidence
	var prefixMatch domain.Customer
	err = r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? AND status != ?", queryLower+"%", domain.CustomerStatusInactive).
		Order("LENGTH(name) ASC"). // Prefer shorter names (more likely exact match)
		First(&prefixMatch).Error
	if err == nil {
		// Calculate confidence based on how much of the name the query covers
		similarity := float64(len(query)) / float64(len(prefixMatch.Name))
		if similarity > 1.0 {
			similarity = 1.0
		}
		// Boost prefix matches
		similarity = 0.8 + (similarity * 0.2)
		return &FuzzySearchResult{Customer: prefixMatch, Similarity: similarity}, nil
	}

	// Strategy 4: Abbreviation match - check if query letters match first letters of words
	// e.g., "AF" matches "AF Gruppen", "NTN" matches "NTNU"
	var abbrevMatches []domain.Customer
	err = r.db.WithContext(ctx).
		Where("status != ?", domain.CustomerStatusInactive).
		Where("UPPER(name) LIKE ?", strings.ToUpper(query)+"%").
		Order("LENGTH(name) ASC").
		Limit(5).
		Find(&abbrevMatches).Error
	if err == nil && len(abbrevMatches) > 0 {
		return &FuzzySearchResult{Customer: abbrevMatches[0], Similarity: 0.85}, nil
	}

	// Strategy 5: Contains match - query is substring
	var containsMatch domain.Customer
	err = r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? AND status != ?", queryPattern, domain.CustomerStatusInactive).
		Order("LENGTH(name) ASC").
		First(&containsMatch).Error
	if err == nil {
		similarity := float64(len(query)) / float64(len(containsMatch.Name))
		if similarity > 0.7 {
			similarity = 0.7
		}
		return &FuzzySearchResult{Customer: containsMatch, Similarity: 0.5 + similarity}, nil
	}

	// Strategy 6: Trigram similarity (handles typos like "Veidikke" -> "Veidekke")
	// This requires pg_trgm extension, fall back to Levenshtein-like matching if not available
	var trigramMatch struct {
		domain.Customer
		Similarity float64
	}
	err = r.db.WithContext(ctx).
		Raw(`
			SELECT c.*, similarity(LOWER(c.name), ?) as similarity
			FROM customers c
			WHERE c.status != ?
			AND similarity(LOWER(c.name), ?) > 0.2
			ORDER BY similarity DESC
			LIMIT 1
		`, queryLower, domain.CustomerStatusInactive, queryLower).
		Scan(&trigramMatch).Error

	if err == nil && trigramMatch.ID != [16]byte{} {
		return &FuzzySearchResult{Customer: trigramMatch.Customer, Similarity: trigramMatch.Similarity}, nil
	}

	// Strategy 7: Fallback - simple LIKE with wildcards between characters for typo tolerance
	// Build pattern like "%v%e%i%d%e%k%k%e%" for "veidekke"
	var wildcardPattern strings.Builder
	wildcardPattern.WriteString("%")
	for _, char := range queryLower {
		wildcardPattern.WriteRune(char)
		wildcardPattern.WriteString("%")
	}

	var fallbackMatch domain.Customer
	err = r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? AND status != ?", wildcardPattern.String(), domain.CustomerStatusInactive).
		Order("LENGTH(name) ASC").
		First(&fallbackMatch).Error
	if err == nil {
		return &FuzzySearchResult{Customer: fallbackMatch, Similarity: 0.3}, nil
	}

	return nil, nil // No match found
}

// GetAllMinimal returns all active customers with minimal fields (id and name)
// Limited to 1000 results to prevent unbounded queries
func (r *CustomerRepository) GetAllMinimal(ctx context.Context) ([]domain.Customer, error) {
	var customers []domain.Customer
	err := r.db.WithContext(ctx).
		Select("id, name").
		Where("status != ?", domain.CustomerStatusInactive).
		Order("name ASC").
		Limit(1000).
		Find(&customers).Error
	return customers, err
}

// GetTopCustomersWithOfferStats returns top customers ranked by offer count within a time window
// If since is nil, no date filter is applied (all time)
// Excludes draft and expired offers from the counts
func (r *CustomerRepository) GetTopCustomersWithOfferStats(ctx context.Context, since *time.Time, limit int) ([]TopCustomerWithStats, error) {
	// Valid phases for counting (excludes draft and expired)
	validPhases := []domain.OfferPhase{
		domain.OfferPhaseInProgress,
		domain.OfferPhaseSent,
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
		domain.OfferPhaseLost,
	}

	var results []TopCustomerWithStats

	// Build subquery to get offer stats per customer
	// Then join with customers to get customer info
	query := r.db.WithContext(ctx).
		Table("offers").
		Select("customers.id as customer_id, customers.name as customer_name, customers.org_number, COUNT(offers.id) as offer_count, COALESCE(SUM(offers.value), 0) as economic_value").
		Joins("JOIN customers ON customers.id = offers.customer_id").
		Where("offers.phase IN ?", validPhases).
		Where("customers.status != ?", domain.CustomerStatusInactive)
	if since != nil {
		query = query.Where("offers.created_at >= ?", *since)
	}
	query = query.Group("customers.id, customers.name, customers.org_number").
		Order("offer_count DESC, economic_value DESC").
		Limit(limit)

	// Apply company filter on offers table
	query = ApplyCompanyFilterWithAlias(ctx, query, "offers")

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get top customers with offer stats: %w", err)
	}

	return results, nil
}

// TopCustomerWithWonStats holds customer data with won offer statistics
type TopCustomerWithWonStats struct {
	CustomerID    uuid.UUID
	CustomerName  string
	OrgNumber     string
	WonOfferCount int
	WonOfferValue float64
}

// GetTopCustomersWithWonStats returns top customers ranked by won offer count and value within a time window
// Won offers are those in order or completed phases
// fromDate and toDate filter by offer created_at (consistent with other pipeline metrics)
func (r *CustomerRepository) GetTopCustomersWithWonStats(ctx context.Context, fromDate, toDate *time.Time, limit int) ([]TopCustomerWithWonStats, error) {
	// Won phases (order + completed)
	wonPhases := []domain.OfferPhase{
		domain.OfferPhaseOrder,
		domain.OfferPhaseCompleted,
	}

	var results []TopCustomerWithWonStats

	// Build query to get won offer stats per customer
	query := r.db.WithContext(ctx).
		Table("offers").
		Select("customers.id as customer_id, customers.name as customer_name, customers.org_number, COUNT(offers.id) as won_offer_count, COALESCE(SUM(offers.value), 0) as won_offer_value").
		Joins("JOIN customers ON customers.id = offers.customer_id").
		Where("offers.phase IN ?", wonPhases).
		Where("customers.status != ?", domain.CustomerStatusInactive)

	if fromDate != nil {
		query = query.Where("offers.created_at >= ?", *fromDate)
	}
	if toDate != nil {
		query = query.Where("offers.created_at <= ?", *toDate)
	}

	query = query.Group("customers.id, customers.name, customers.org_number").
		Order("won_offer_count DESC, won_offer_value DESC").
		Limit(limit)

	// Apply company filter on offers table
	query = ApplyCompanyFilterWithAlias(ctx, query, "offers")

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get top customers with won stats: %w", err)
	}

	return results, nil
}
