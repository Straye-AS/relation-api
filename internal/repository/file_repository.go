package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *FileRepository) ListByOffer(ctx context.Context, offerID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("offer_id = ?", offerID).
		Order("updated_at DESC").
		Find(&files).Error
	return files, err
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.File{}, "id = ?", id).Error
}
