package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type OfferItemRepository struct {
	db *gorm.DB
}

func NewOfferItemRepository(db *gorm.DB) *OfferItemRepository {
	return &OfferItemRepository{db: db}
}

func (r *OfferItemRepository) Create(ctx context.Context, item *domain.OfferItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *OfferItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OfferItem, error) {
	var item domain.OfferItem
	err := r.db.WithContext(ctx).First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OfferItemRepository) ListByOffer(ctx context.Context, offerID uuid.UUID) ([]domain.OfferItem, error) {
	var items []domain.OfferItem
	err := r.db.WithContext(ctx).
		Where("offer_id = ?", offerID).
		Order("created_at").
		Find(&items).Error
	return items, err
}

func (r *OfferItemRepository) Update(ctx context.Context, item *domain.OfferItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *OfferItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.OfferItem{}, "id = ?", id).Error
}
