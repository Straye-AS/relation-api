package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	query := r.db.WithContext(ctx).Preload("Customer").Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Project{}, "id = ?", id).Error
}

func (r *ProjectRepository) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID, status *domain.ProjectStatus) ([]domain.Project, int64, error) {
	var projects []domain.Project
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Project{}).Preload("Customer")

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&projects).Error

	return projects, total, err
}

func (r *ProjectRepository) CountActive(ctx context.Context) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Project{}).
		Where("status = ?", domain.ProjectStatusActive)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return int(count), err
}

func (r *ProjectRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Project, error) {
	var projects []domain.Project
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).Preload("Customer").
		Where("LOWER(name) LIKE ? OR LOWER(summary) LIKE ?", searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Find(&projects).Error
	return projects, err
}
