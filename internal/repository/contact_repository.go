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

// SetPrimaryContactForEntity sets a contact as the primary for a specific entity,
// unsetting any existing primary contact for that entity
func (r *ContactRepository) SetPrimaryContactForEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID, contactID uuid.UUID) error {
	// First, unset any existing primary contact for this entity
	if err := r.db.WithContext(ctx).
		Model(&domain.ContactRelationship{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Update("is_primary", false).Error; err != nil {
		return err
	}

	// Then set the new primary
	return r.db.WithContext(ctx).
		Model(&domain.ContactRelationship{}).
		Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
		Update("is_primary", true).Error
}

// GetRelationshipByID returns a specific relationship by its ID
func (r *ContactRepository) GetRelationshipByID(ctx context.Context, relationshipID uuid.UUID) (*domain.ContactRelationship, error) {
	var relationship domain.ContactRelationship
	err := r.db.WithContext(ctx).
		Preload("Contact").
		Where("id = ?", relationshipID).
		First(&relationship).Error
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// GetRelationshipByContactAndEntity returns a specific relationship by contact and entity
func (r *ContactRepository) GetRelationshipByContactAndEntity(ctx context.Context, contactID uuid.UUID, entityType domain.ContactEntityType, entityID uuid.UUID) (*domain.ContactRelationship, error) {
	var relationship domain.ContactRelationship
	err := r.db.WithContext(ctx).
		Where("contact_id = ? AND entity_type = ? AND entity_id = ?", contactID, entityType, entityID).
		First(&relationship).Error
	if err != nil {
		return nil, err
	}
	return &relationship, nil
}

// DeleteRelationshipByID deletes a relationship by its ID
func (r *ContactRepository) DeleteRelationshipByID(ctx context.Context, relationshipID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ContactRelationship{}, "id = ?", relationshipID).Error
}

// ExistsByEmail checks if a contact with the given email already exists
func (r *ContactRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Contact{}).
		Where("LOWER(email) = LOWER(?)", email).
		Count(&count).Error
	return count > 0, err
}

// ExistsByEmailExcluding checks if a contact with the given email exists, excluding a specific contact ID
func (r *ContactRepository) ExistsByEmailExcluding(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Contact{}).
		Where("LOWER(email) = LOWER(?) AND id != ?", email, excludeID).
		Count(&count).Error
	return count > 0, err
}

// HasRelationships checks if a contact has any relationships
func (r *ContactRepository) HasRelationships(ctx context.Context, contactID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ContactRelationship{}).
		Where("contact_id = ?", contactID).
		Count(&count).Error
	return count > 0, err
}

// CountRelationships returns the number of relationships for a contact
func (r *ContactRepository) CountRelationships(ctx context.Context, contactID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.ContactRelationship{}).
		Where("contact_id = ?", contactID).
		Count(&count).Error
	return count, err
}
