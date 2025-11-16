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
	err := r.db.WithContext(ctx).First(&contact, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *ContactRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.Contact, error) {
	var contacts []domain.Contact
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("name").
		Find(&contacts).Error
	return contacts, err
}

func (r *ContactRepository) Update(ctx context.Context, contact *domain.Contact) error {
	return r.db.WithContext(ctx).Save(contact).Error
}

func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Contact{}, "id = ?", id).Error
}
