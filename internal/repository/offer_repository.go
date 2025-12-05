package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type OfferRepository struct {
	db *gorm.DB
}

func NewOfferRepository(db *gorm.DB) *OfferRepository {
	return &OfferRepository{db: db}
}

func (r *OfferRepository) Create(ctx context.Context, offer *domain.Offer) error {
	return r.db.WithContext(ctx).Create(offer).Error
}

func (r *OfferRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Offer, error) {
	var offer domain.Offer
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Project").
		Preload("Items").
		Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *OfferRepository) Update(ctx context.Context, offer *domain.Offer) error {
	return r.db.WithContext(ctx).Save(offer).Error
}

func (r *OfferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Offer{}, "id = ?", id).Error
}

func (r *OfferRepository) List(ctx context.Context, page, pageSize int, customerID, projectID *uuid.UUID, phase *domain.OfferPhase) ([]domain.Offer, int64, error) {
	var offers []domain.Offer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Offer{}).Preload("Customer").Preload("Project")

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}

	if phase != nil {
		query = query.Where("phase = ?", *phase)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&offers).Error

	return offers, total, err
}

func (r *OfferRepository) GetItemsCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.OfferItem{}).Where("offer_id = ?", offerID).Count(&count).Error
	return int(count), err
}

func (r *OfferRepository) GetFilesCount(ctx context.Context, offerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.File{}).Where("offer_id = ?", offerID).Count(&count).Error
	return int(count), err
}

func (r *OfferRepository) GetTotalPipelineValue(ctx context.Context) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&domain.Offer{}).
		Where("phase IN ?", []domain.OfferPhase{
			domain.OfferPhaseSent,
			domain.OfferPhaseInProgress,
		})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Select("COALESCE(SUM(value), 0)").
		Scan(&total).Error
	return total, err
}

func (r *OfferRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Offer, error) {
	var offers []domain.Offer
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Project").
		Where("LOWER(title) LIKE ?", searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Find(&offers).Error
	return offers, err
}
