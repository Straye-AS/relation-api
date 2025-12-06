package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type ContactRepository struct {
	db *gorm.DB
}

func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Create(contact).Error
}

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

func (r *ContactRepository) List(ctx context.Context, page, pageSize int) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	offset := (page - 1) * pageSize

	// Get total count
	if err := r.db.WithContext(ctx).Model(&domain.Contact{}).
		Where("is_active = ?", true).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Preload("Relationships").
		Where("is_active = ?", true).
		Order("last_name, first_name").
		Offset(offset).
		Limit(pageSize).
		Find(&contacts).Error

	return contacts, total, err
}

// ListByEntity returns contacts related to a specific entity via contact_relationships
func (r *ContactRepository) ListByEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) ([]domain.Contact, error) {
	var contacts []domain.Contact

	err := r.db.WithContext(ctx).
		Joins("JOIN contact_relationships cr ON cr.contact_id = contacts.id").
		Where("cr.entity_type = ? AND cr.entity_id = ?", entityType, entityID).
		Preload("Relationships").
		Order("contacts.last_name, contacts.first_name").
		Find(&contacts).Error

	return contacts, err
}

// ListByPrimaryCustomer returns contacts with a specific primary customer
func (r *ContactRepository) ListByPrimaryCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Contact, error) {
	var contacts []domain.Contact
	err := r.db.WithContext(ctx).
		Preload("Relationships").
		Where("primary_customer_id = ? AND is_active = ?", customerID, true).
		Order("last_name, first_name").
		Find(&contacts).Error
	return contacts, err
}

func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Contact{}, "id = ?", id).Error
}

// GetByEmail finds a contact by email address
func (r *ContactRepository) GetByEmail(ctx context.Context, email string) (*domain.Contact, error) {
	var contact domain.Contact
	err := r.db.WithContext(ctx).
		Where("LOWER(email) = LOWER(?)", email).
		First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

// Search searches contacts by name or email
func (r *ContactRepository) Search(ctx context.Context, query string, limit int) ([]domain.Contact, error) {
	var contacts []domain.Contact
	searchPattern := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
			searchPattern, searchPattern, searchPattern).
		Order("last_name, first_name").
		Limit(limit).
		Find(&contacts).Error

	return contacts, err
}

// ContactRelationship methods

func (r *ContactRepository) AddRelationship(ctx context.Context, rel *domain.ContactRelationship) error {
	return r.db.WithContext(ctx).Create(rel).Error
}

func (r *ContactRepository) RemoveRelationship(ctx context.Context, contactID uuid.UUID, entityType domain.ContactEntityType, entityID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
		Delete(&domain.ContactRelationship{}).Error
}

func (r *ContactRepository) GetRelationships(ctx context.Context, contactID uuid.UUID) ([]domain.ContactRelationship, error) {
	var relationships []domain.ContactRelationship
	err := r.db.WithContext(ctx).
		Where("contact_id = ?", contactID).
		Find(&relationships).Error
	return relationships, err
}

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

// ContactFilters holds filters for listing contacts
type ContactFilters struct {
	Search      string
	Title       string
	ContactType *domain.ContactType
	EntityType  *domain.ContactEntityType
	EntityID    *uuid.UUID
}

// ContactSortOption defines sort options for contacts
type ContactSortOption string

const (
	ContactSortByNameAsc     ContactSortOption = "name_asc"
	ContactSortByNameDesc    ContactSortOption = "name_desc"
	ContactSortByEmailAsc    ContactSortOption = "email_asc"
	ContactSortByCreatedDesc ContactSortOption = "created_desc"
)

// ListWithFilters returns contacts with filters and pagination
func (r *ContactRepository) ListWithFilters(ctx context.Context, page, pageSize int, filters *ContactFilters, sortBy ContactSortOption) ([]domain.Contact, int64, error) {
	var contacts []domain.Contact
	var total int64

	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&domain.Contact{}).Where("is_active = ?", true)

	// Apply filters
	if filters != nil {
		if filters.Search != "" {
			searchPattern := "%" + filters.Search + "%"
			query = query.Where(
				"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?",
				searchPattern, searchPattern, searchPattern,
			)
		}

		if filters.Title != "" {
			query = query.Where("title ILIKE ?", "%"+filters.Title+"%")
		}

		if filters.ContactType != nil {
			query = query.Where("contact_type = ?", *filters.ContactType)
		}

		if filters.EntityType != nil && filters.EntityID != nil {
			query = query.Joins("JOIN contact_relationships cr ON cr.contact_id = contacts.id").
				Where("cr.entity_type = ? AND cr.entity_id = ?", *filters.EntityType, *filters.EntityID)
		}
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch sortBy {
	case ContactSortByNameAsc:
		query = query.Order("last_name ASC, first_name ASC")
	case ContactSortByNameDesc:
		query = query.Order("last_name DESC, first_name DESC")
	case ContactSortByEmailAsc:
		query = query.Order("email ASC")
	case ContactSortByCreatedDesc:
		query = query.Order("created_at DESC")
	default:
		query = query.Order("last_name ASC, first_name ASC")
	}

	// Get paginated results
	err := query.Preload("Relationships").
		Preload("PrimaryCustomer").
		Offset(offset).
		Limit(pageSize).
		Find(&contacts).Error

	return contacts, total, err
}

// GetRelationshipByID returns a relationship by its ID
func (r *ContactRepository) GetRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*domain.ContactRelationship, error) {
	var rel domain.ContactRelationship
	err := r.db.WithContext(ctx).First(&rel, "id = ?", relationshipID).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

// RemoveRelationshipByID removes a relationship by its ID
func (r *ContactRepository) RemoveRelationshipByID(ctx context.Context, relationshipID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ContactRelationship{}, "id = ?", relationshipID).Error
}

// CheckRelationshipExists checks if a relationship already exists
func (r *ContactRepository) CheckRelationshipExists(ctx context.Context, contactID uuid.UUID, entityType domain.ContactEntityType, entityID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ContactRelationship{}).
		Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
		Count(&count).Error
	return count > 0, err
}
