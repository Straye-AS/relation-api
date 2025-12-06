package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// ContactFilters contains all filter options for listing contacts
type ContactFilters struct {
	SearchQuery       *string                   // Search in first_name, last_name, email
	Title             *string                   // Filter by job title
	Department        *string                   // Filter by department
	EntityType        *domain.ContactEntityType // Filter by related entity type
	EntityID          *uuid.UUID                // Filter by specific entity
	PrimaryCustomerID *uuid.UUID                // Filter by primary customer
	IsActive          *bool                     // Filter by active status
}

// ContactSortOption represents available sort options for contacts
type ContactSortOption string

const (
	ContactSortByNameAsc     ContactSortOption = "name_asc"
	ContactSortByNameDesc    ContactSortOption = "name_desc"
	ContactSortByEmailAsc    ContactSortOption = "email_asc"
	ContactSortByEmailDesc   ContactSortOption = "email_desc"
	ContactSortByCreatedAsc  ContactSortOption = "created_asc"
	ContactSortByCreatedDesc ContactSortOption = "created_desc"
	ContactSortByUpdatedDesc ContactSortOption = "updated_desc"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// Create creates a new contact
func (r *ContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Create(contact).Error
}

// GetByID retrieves a contact by ID with preloaded relationships
func (r *ContactRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Contact, error) {
	var contact domain.Contact
	err := r.db.WithContext(ctx).
		Preload("Relationships").
		Preload("PrimaryCustomer").
		First(&contact, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// GetByEmail retrieves a contact by email address (for uniqueness checks)
func (r *ContactRepository) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	var contact domain.Contact
	err := r.db.WithContext(ctx).
		Preload("Relationships").
		Where("LOWER(email) = LOWER(?)", email).
		First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// Update updates an existing contact
func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

// Delete soft-deletes a contact by setting is_active to false
func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Contact{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

// HardDelete permanently deletes a contact (use with caution)
func (r *ContactRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Contact{}, "id = ?", id).Error
}

// List retrieves contacts with filtering, sorting, and pagination
func (r *ContactRepository) List(ctx context.Context, page, pageSize int, filters *ContactFilters, sortBy ContactSortOption) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	// Start building the query
	query := r.db.WithContext(ctx).Model(&domain.Contact{})

	// Apply filters
	query = r.applyFilters(query, filters)

	// Get total count before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting contacts: %w", err)
	}

	// Apply sorting
	query = r.applySorting(query, sortBy)

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.
		Preload("Relationships").
		Offset(offset).
		Limit(pageSize).
		Find(&contacts).Error

	if err != nil {
		return nil, 0, fmt.Errorf("listing contacts: %w", err)
	}

	return contacts, total, nil
}

// Search performs a search on contacts by name or email
func (r *ContactRepository) Search(ctx context.Context, query string, limit int) ([]domain.Contact, error) {
	var contacts []domain.Contact
	searchPattern := "%" + strings.ToLower(query) + "%"

	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name || ' ' || last_name) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Preload("Relationships").
		Order("last_name, first_name").
		Limit(limit).
		Find(&contacts).Error

	if err != nil {
		return nil, fmt.Errorf("searching contacts: %w", err)
	}

	return contacts, nil
}

// GetContactsForEntity returns all contacts related to a specific entity
func (r *ContactRepository) GetContactsForEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) ([]domain.Contact, error) {
	var contacts []domain.Contact

	err := r.db.WithContext(ctx).
		Joins("JOIN contact_relationships cr ON cr.contact_id = contacts.id").
		Where("cr.entity_type = ? AND cr.entity_id = ?", entityType, entityID).
		Where("contacts.is_active = ?", true).
		Preload("Relationships").
		Order("cr.is_primary DESC, contacts.last_name, contacts.first_name").
		Find(&contacts).Error

	if err != nil {
		return nil, fmt.Errorf("getting contacts for entity: %w", err)
	}

	return contacts, nil
}

// GetPrimaryContactForEntity returns the primary contact for an entity
func (r *ContactRepository) GetPrimaryContactForEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) (*domain.Contact, error) {
	var contact domain.Contact

	err := r.db.WithContext(ctx).
		Joins("JOIN contact_relationships cr ON cr.contact_id = contacts.id").
		Where("cr.entity_type = ? AND cr.entity_id = ? AND cr.is_primary = ?", entityType, entityID, true).
		Where("contacts.is_active = ?", true).
		Preload("Relationships").
		First(&contact).Error

	if err != nil {
		return nil, err
	}

	return &contact, nil
}

// ListByPrimaryCustomer returns contacts with a specific primary customer
func (r *ContactRepository) ListByPrimaryCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Contact, error) {
	var contacts []domain.Contact
	err := r.db.WithContext(ctx).
		Preload("Relationships").
		Where("primary_customer_id = ? AND is_active = ?", customerID, true).
		Order("last_name, first_name").
		Find(&contacts).Error
	if err != nil {
		return nil, fmt.Errorf("listing contacts by primary customer: %w", err)
	}
	return contacts, err
}

// ListByEntity returns contacts related to a specific entity via contact_relationships
// Deprecated: Use GetContactsForEntity instead
func (r *ContactRepository) ListByEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) ([]domain.Contact, error) {
	return r.GetContactsForEntity(ctx, entityType, entityID)
}

// =============================================================================
// Contact Relationship Methods
// =============================================================================

// GetContactRelationships returns all relationships for a contact
func (r *ContactRepository) GetContactRelationships(ctx context.Context, contactID uuid.UUID) ([]domain.ContactRelationship, error) {
	var relationships []domain.ContactRelationship
	err := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Order("is_primary DESC, created_at").
		Find(&relationships).Error
	if err != nil {
		return nil, fmt.Errorf("getting contact relationships: %w", err)
	}
	return relationships, nil
}

// GetRelationships is an alias for GetContactRelationships
// Deprecated: Use GetContactRelationships instead
func (r *ContactRepository) GetRelationships(ctx context.Context, contactID uuid.UUID) ([]domain.ContactRelationship, error) {
	return r.GetContactRelationships(ctx, contactID)
}

// AddRelationship creates a new relationship between a contact and an entity
func (r *ContactRepository) AddRelationship(ctx context.Context, rel *domain.ContactRelationship) error {
	if rel.ID == uuid.Nil {
		rel.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(rel).Error
}

// GetRelationshipByID retrieves a relationship by its ID
func (r *ContactRepository) GetRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*domain.ContactRelationship, error) {
	var rel domain.ContactRelationship
	err := r.db.WithContext(ctx).
		Preload("Contact").
		First(&rel, "id = ?", relationshipID).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// RemoveRelationshipByID removes a relationship by its ID
func (r *ContactRepository) RemoveRelationshipByID(ctx context.Context, relationshipID uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.ContactRelationship{}, "id = ?", relationshipID)
	if result.Error != nil {
		return fmt.Errorf("removing relationship: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// RemoveRelationship removes a specific relationship between a contact and an entity
func (r *ContactRepository) RemoveRelationship(ctx context.Context, contactID uuid.UUID, entityType domain.ContactEntityType, entityID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
		Delete(&domain.ContactRelationship{})
	if result.Error != nil {
		return fmt.Errorf("removing relationship: %w", result.Error)
	}
	return nil
}

// SetPrimaryContact sets a contact as the primary contact for an entity
// This unsets any existing primary contact for the same entity first
func (r *ContactRepository) SetPrimaryContact(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID, contactID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, unset any existing primary contacts for this entity
		if err := tx.Model(&domain.ContactRelationship{}).
			Where("entity_type = ? AND entity_id = ? AND is_primary = ?", entityType, entityID, true).
			Update("is_primary", false).Error; err != nil {
			return fmt.Errorf("unsetting existing primary contact: %w", err)
		}

		// Then set the new primary contact
		result := tx.Model(&domain.ContactRelationship{}).
			Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
			Update("is_primary", true)

		if result.Error != nil {
			return fmt.Errorf("setting primary contact: %w", result.Error)
		}

		// If no relationship exists, we need to create one
		if result.RowsAffected == 0 {
			rel := &domain.ContactRelationship{
				ID:         uuid.New(),
				ContactID:  contactID,
				EntityType: entityType,
				EntityID:   entityID,
				IsPrimary:  true,
			}
			if err := tx.Create(rel).Error; err != nil {
				return fmt.Errorf("creating primary relationship: %w", err)
			}
		}

		return nil
	})
}

// SetPrimaryRelationship sets a specific relationship as primary for a contact-entity type combination
// Deprecated: Use SetPrimaryContact instead for entity-centric primary contact management
func (r *ContactRepository) SetPrimaryRelationship(ctx context.Context, contactID uuid.UUID, entityType domain.ContactEntityType, entityID uuid.UUID) error {
	// First, unset any existing primary for this contact-entity type combination
	if err := r.db.WithContext(ctx).
		Model(&domain.ContactRelationship{}).
		Where("contact_id = ? AND entity_type = ?", contactID, entityType).
		Update("is_primary", false).Error; err != nil {
		return err
	}

	// Then set the new primary
	return r.db.WithContext(ctx).
		Model(&domain.ContactRelationship{}).
		Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
		Update("is_primary", true).Error
}

// GetRelationshipsForEntity returns all relationships for a specific entity
func (r *ContactRepository) GetRelationshipsForEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) ([]domain.ContactRelationship, error) {
	var relationships []domain.ContactRelationship
	err := r.db.WithContext(ctx).
		Preload("Contact").
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("is_primary DESC, created_at").
		Find(&relationships).Error
	if err != nil {
		return nil, fmt.Errorf("getting relationships for entity: %w", err)
	}
	return relationships, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// applyFilters applies the filter criteria to the query
func (r *ContactRepository) applyFilters(query *gorm.DB, filters *ContactFilters) *gorm.DB {
	if filters == nil {
		// Default to only active contacts
		return query.Where("is_active = ?", true)
	}

	// Active status filter (default to true)
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("is_active = ?", true)
	}

	// Search query (name or email)
	if filters.SearchQuery != nil && *filters.SearchQuery != "" {
		searchPattern := "%" + strings.ToLower(*filters.SearchQuery) + "%"
		query = query.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name || ' ' || last_name) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Title filter
	if filters.Title != nil && *filters.Title != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(*filters.Title)+"%")
	}

	// Department filter
	if filters.Department != nil && *filters.Department != "" {
		query = query.Where("LOWER(department) LIKE ?", "%"+strings.ToLower(*filters.Department)+"%")
	}

	// Primary customer filter
	if filters.PrimaryCustomerID != nil {
		query = query.Where("primary_customer_id = ?", *filters.PrimaryCustomerID)
	}

	// Entity relationship filter
	if filters.EntityType != nil && filters.EntityID != nil {
		query = query.Joins("JOIN contact_relationships cr ON cr.contact_id = contacts.id").
			Where("cr.entity_type = ? AND cr.entity_id = ?", *filters.EntityType, *filters.EntityID)
	} else if filters.EntityType != nil {
		// Filter by entity type only
		query = query.Joins("JOIN contact_relationships cr ON cr.contact_id = contacts.id").
			Where("cr.entity_type = ?", *filters.EntityType)
	}

	return query
}

// applySorting applies the sorting option to the query
func (r *ContactRepository) applySorting(query *gorm.DB, sortBy ContactSortOption) *gorm.DB {
	switch sortBy {
	case ContactSortByNameDesc:
		return query.Order("last_name DESC, first_name DESC")
	case ContactSortByEmailAsc:
		return query.Order("email ASC")
	case ContactSortByEmailDesc:
		return query.Order("email DESC")
	case ContactSortByCreatedAsc:
		return query.Order("created_at ASC")
	case ContactSortByCreatedDesc:
		return query.Order("created_at DESC")
	case ContactSortByUpdatedDesc:
		return query.Order("updated_at DESC")
	default: // ContactSortByNameAsc
		return query.Order("last_name ASC, first_name ASC")
	}
}

// WithTransaction executes operations within a transaction
func (r *ContactRepository) WithTransaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
