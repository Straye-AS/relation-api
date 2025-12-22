package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// SupplierFilters defines filter options for supplier listing
type SupplierFilters struct {
	Search    string
	City      string
	Country   string
	Status    *domain.SupplierStatus
	Category  string
	CompanyID *domain.CompanyID
}

// supplierSortableFields maps API field names to database column names for suppliers
// Only fields in this map can be used for sorting (whitelist approach)
var supplierSortableFields = map[string]string{
	"createdAt": "created_at",
	"updatedAt": "updated_at",
	"name":      "name",
	"city":      "city",
	"country":   "country",
	"status":    "status",
	"category":  "category",
	"orgNumber": "org_number",
	"companyId": "company_id",
}

// SupplierRepository handles supplier data access operations
type SupplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new supplier repository instance
func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// Create creates a new supplier in the database
func (r *SupplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

// GetByID retrieves a supplier by its ID
func (r *SupplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// GetByOrgNumber finds a supplier by organization number
func (r *SupplierRepository) GetByOrgNumber(ctx context.Context, orgNumber string) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).Where("org_number = ?", orgNumber).First(&supplier).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &supplier, nil
}

// Update updates an existing supplier in the database
func (r *SupplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

// Delete performs a soft delete on a supplier
func (r *SupplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Supplier{}, "id = ?", id).Error
}

// List returns a paginated list of suppliers with default filters
func (r *SupplierRepository) List(ctx context.Context, page, pageSize int, search string) ([]domain.Supplier, int64, error) {
	filters := &SupplierFilters{Search: search}
	return r.ListWithSortConfig(ctx, page, pageSize, filters, DefaultSortConfig())
}

// ListWithSortConfig returns a paginated list of suppliers with filter and sort options using SortConfig
func (r *SupplierRepository) ListWithSortConfig(ctx context.Context, page, pageSize int, filters *SupplierFilters, sort SortConfig) ([]domain.Supplier, int64, error) {
	var suppliers []domain.Supplier
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

	query := r.db.WithContext(ctx).Model(&domain.Supplier{})

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
		if filters.Category != "" {
			categoryPattern := "%" + strings.ToLower(filters.Category) + "%"
			query = query.Where("LOWER(category) LIKE ?", categoryPattern)
		}
		if filters.CompanyID != nil {
			query = query.Where("company_id = ?", *filters.CompanyID)
		}
	}

	// Note: Suppliers are global entities, no company filter applied

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build order clause from sort config
	orderClause := BuildOrderClause(sort, supplierSortableFields, "updated_at")

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order(orderClause).Find(&suppliers).Error

	return suppliers, total, err
}

// Count returns the total count of active suppliers
// Note: Suppliers are global entities, no company filter applied
func (r *SupplierRepository) Count(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Supplier{}).
		Where("status != ?", domain.SupplierStatusBlacklisted).
		Count(&count).Error
	return int(count), err
}

// Search searches for suppliers by name or org number
// Note: Suppliers are global entities, no company filter applied
func (r *SupplierRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Supplier, error) {
	var suppliers []domain.Supplier
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	err := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern).
		Where("status != ?", domain.SupplierStatusBlacklisted).
		Limit(limit).
		Find(&suppliers).Error
	return suppliers, err
}

// GetWithContacts retrieves a supplier with preloaded contacts
func (r *SupplierRepository) GetWithContacts(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).
		Preload("Contacts", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, last_name ASC, first_name ASC").Limit(20)
		}).
		Where("id = ?", id).
		First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// SupplierStats holds aggregated statistics for a supplier
type SupplierStats struct {
	TotalOffers     int `json:"totalOffers"`
	ActiveOffers    int `json:"activeOffers"`
	CompletedOffers int `json:"completedOffers"`
	TotalProjects   int `json:"totalProjects"`
}

// GetSupplierStats returns aggregated statistics for a supplier
func (r *SupplierRepository) GetSupplierStats(ctx context.Context, supplierID uuid.UUID) (*SupplierStats, error) {
	stats := &SupplierStats{}

	// Get total offers count (via offer_suppliers junction table)
	var totalOffers int64
	err := r.db.WithContext(ctx).Model(&domain.OfferSupplier{}).
		Where("supplier_id = ?", supplierID).
		Count(&totalOffers).Error
	if err != nil {
		return nil, err
	}
	stats.TotalOffers = int(totalOffers)

	// Get active offers count (offers in in_progress, sent, or order phases)
	var activeOffers int64
	err = r.db.WithContext(ctx).Model(&domain.OfferSupplier{}).
		Joins("JOIN offers ON offers.id = offer_suppliers.offer_id").
		Where("offer_suppliers.supplier_id = ?", supplierID).
		Where("offers.phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
			domain.OfferPhaseOrder,
		}).
		Count(&activeOffers).Error
	if err != nil {
		return nil, err
	}
	stats.ActiveOffers = int(activeOffers)

	// Get completed offers count (offers in completed phase)
	var completedOffers int64
	err = r.db.WithContext(ctx).Model(&domain.OfferSupplier{}).
		Joins("JOIN offers ON offers.id = offer_suppliers.offer_id").
		Where("offer_suppliers.supplier_id = ?", supplierID).
		Where("offers.phase = ?", domain.OfferPhaseCompleted).
		Count(&completedOffers).Error
	if err != nil {
		return nil, err
	}
	stats.CompletedOffers = int(completedOffers)

	// Get total unique projects count (via offers linked to projects)
	var totalProjects int64
	err = r.db.WithContext(ctx).Model(&domain.OfferSupplier{}).
		Select("COUNT(DISTINCT offers.project_id)").
		Joins("JOIN offers ON offers.id = offer_suppliers.offer_id").
		Where("offer_suppliers.supplier_id = ?", supplierID).
		Where("offers.project_id IS NOT NULL").
		Count(&totalProjects).Error
	if err != nil {
		return nil, err
	}
	stats.TotalProjects = int(totalProjects)

	return stats, nil
}

// GetRecentOfferSuppliers returns recent offer-supplier relationships for a supplier
func (r *SupplierRepository) GetRecentOfferSuppliers(ctx context.Context, supplierID uuid.UUID, limit int) ([]domain.OfferSupplier, error) {
	var offerSuppliers []domain.OfferSupplier
	err := r.db.WithContext(ctx).
		Where("supplier_id = ?", supplierID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&offerSuppliers).Error
	return offerSuppliers, err
}

// HasActiveRelations checks if a supplier has active offer relationships
func (r *SupplierRepository) HasActiveRelations(ctx context.Context, supplierID uuid.UUID) (bool, string, error) {
	// Check for active offer-supplier relationships
	var activeCount int64
	err := r.db.WithContext(ctx).Model(&domain.OfferSupplier{}).
		Joins("JOIN offers ON offers.id = offer_suppliers.offer_id").
		Where("offer_suppliers.supplier_id = ?", supplierID).
		Where("offer_suppliers.status = ?", domain.OfferSupplierStatusActive).
		Where("offers.phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
			domain.OfferPhaseOrder,
		}).
		Count(&activeCount).Error
	if err != nil {
		return false, "", err
	}

	if activeCount > 0 {
		return true, "supplier has active offer relationships", nil
	}

	return false, "", nil
}

// ============================================================================
// Supplier Contact Methods
// ============================================================================

// ListContacts returns all contacts for a supplier
func (r *SupplierRepository) ListContacts(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplierContact, error) {
	var contacts []domain.SupplierContact
	err := r.db.WithContext(ctx).
		Where("supplier_id = ?", supplierID).
		Order("is_primary DESC, last_name ASC, first_name ASC").
		Find(&contacts).Error
	return contacts, err
}

// GetContactByID retrieves a supplier contact by its ID
func (r *SupplierRepository) GetContactByID(ctx context.Context, contactID uuid.UUID) (*domain.SupplierContact, error) {
	var contact domain.SupplierContact
	err := r.db.WithContext(ctx).Where("id = ?", contactID).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// CreateContact creates a new supplier contact
func (r *SupplierRepository) CreateContact(ctx context.Context, contact *domain.SupplierContact) error {
	return r.db.WithContext(ctx).Create(contact).Error
}

// UpdateContact updates an existing supplier contact
func (r *SupplierRepository) UpdateContact(ctx context.Context, contact *domain.SupplierContact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

// DeleteContact deletes a supplier contact
func (r *SupplierRepository) DeleteContact(ctx context.Context, contactID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SupplierContact{}, "id = ?", contactID).Error
}

// ClearPrimaryContacts sets is_primary to false for all contacts of a supplier
func (r *SupplierRepository) ClearPrimaryContacts(ctx context.Context, supplierID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.SupplierContact{}).
		Where("supplier_id = ?", supplierID).
		Update("is_primary", false).Error
}

// IsContactUsedInOffers checks if a contact is currently assigned to any offer-supplier relationships
func (r *SupplierRepository) IsContactUsedInOffers(ctx context.Context, contactID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.OfferSupplier{}).
		Joins("JOIN offers ON offers.id = offer_suppliers.offer_id").
		Where("offer_suppliers.contact_id = ?", contactID).
		Where("offers.phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseInProgress,
			domain.OfferPhaseSent,
			domain.OfferPhaseOrder,
		}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ============================================================================
// Supplier Offers Methods
// ============================================================================

// supplierOfferSortableFields maps API field names to database column names for offer sorting
var supplierOfferSortableFields = map[string]string{
	"createdAt":    "offers.created_at",
	"updatedAt":    "offers.updated_at",
	"title":        "offers.title",
	"value":        "offers.value",
	"probability":  "offers.probability",
	"phase":        "offers.phase",
	"status":       "offers.status",
	"dueDate":      "offers.due_date",
	"customerName": "offers.customer_name",
}

// GetOffersBySupplier returns a paginated list of offers linked to a supplier via offer_suppliers junction table
func (r *SupplierRepository) GetOffersBySupplier(ctx context.Context, supplierID uuid.UUID, page, pageSize int, phase *domain.OfferPhase, sort SortConfig) ([]domain.Offer, int64, error) {
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

	// Base query: join offers with offer_suppliers to get offers for this supplier
	query := r.db.WithContext(ctx).
		Model(&domain.Offer{}).
		Preload("Customer").
		Joins("JOIN offer_suppliers ON offer_suppliers.offer_id = offers.id").
		Where("offer_suppliers.supplier_id = ?", supplierID)

	// Apply company filter on offers table
	query = ApplyCompanyFilterWithAlias(ctx, query, "offers")

	// Apply optional phase filter
	if phase != nil {
		query = query.Where("offers.phase = ?", *phase)
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build order clause from sort config
	orderClause := BuildOrderClause(sort, supplierOfferSortableFields, "offers.updated_at")

	// Apply pagination and ordering
	offset := (page - 1) * pageSize
	err := query.
		Offset(offset).
		Limit(pageSize).
		Order(orderClause).
		Find(&offers).Error

	return offers, total, err
}
