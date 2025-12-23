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

// ListByOffer returns all files attached to an offer
func (r *FileRepository) ListByOffer(ctx context.Context, offerID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("offer_id = ?", offerID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.File{}, "id = ?", id).Error
}

// ListByCustomer returns all files attached to a customer
func (r *FileRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListByProject returns all files attached to a project
func (r *FileRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// ListBySupplier returns all files attached to a supplier
func (r *FileRepository) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.File, error) {
	var files []domain.File
	err := r.db.WithContext(ctx).
		Where("supplier_id = ?", supplierID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// CountByOffer returns the count of files attached to an offer
func (r *FileRepository) CountByOffer(ctx context.Context, offerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("offer_id = ?", offerID).
		Count(&count).Error
	return count, err
}

// CountByCustomer returns the count of files attached to a customer
func (r *FileRepository) CountByCustomer(ctx context.Context, customerID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("customer_id = ?", customerID).
		Count(&count).Error
	return count, err
}

// CountByProject returns the count of files attached to a project
func (r *FileRepository) CountByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return count, err
}

// CountBySupplier returns the count of files attached to a supplier
func (r *FileRepository) CountBySupplier(ctx context.Context, supplierID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("supplier_id = ?", supplierID).
		Count(&count).Error
	return count, err
}
