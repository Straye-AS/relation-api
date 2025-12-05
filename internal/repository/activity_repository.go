package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *ActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error) {
	var activity domain.Activity
	query := r.db.WithContext(ctx).Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) List(ctx context.Context, page, pageSize int, targetType *domain.ActivityTargetType, targetID *uuid.UUID) ([]domain.Activity, int64, error) {
	var activities []domain.Activity
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Activity{})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if targetType != nil {
		query = query.Where("target_type = ?", *targetType)
	}

	if targetID != nil {
		query = query.Where("target_id = ?", *targetID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("occurred_at DESC").Find(&activities).Error

	return activities, total, err
}

func (r *ActivityRepository) ListByTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, limit int) ([]domain.Activity, error) {
	var activities []domain.Activity
	query := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("occurred_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}
