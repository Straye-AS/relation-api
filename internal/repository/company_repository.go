package repository

import (
	"context"

	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// CompanyRepository handles database operations for companies
type CompanyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository creates a new CompanyRepository
func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

// GetByID retrieves a company by its ID
func (r *CompanyRepository) GetByID(ctx context.Context, id domain.CompanyID) (*domain.Company, error) {
	var company domain.Company
	err := r.db.WithContext(ctx).First(&company, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// List returns all active companies
func (r *CompanyRepository) List(ctx context.Context) ([]domain.Company, error) {
	var companies []domain.Company
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("name ASC").
		Find(&companies).Error
	if err != nil {
		return nil, err
	}
	return companies, nil
}

// ListAll returns all companies including inactive ones
func (r *CompanyRepository) ListAll(ctx context.Context) ([]domain.Company, error) {
	var companies []domain.Company
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&companies).Error
	if err != nil {
		return nil, err
	}
	return companies, nil
}

// Update updates a company's fields
func (r *CompanyRepository) Update(ctx context.Context, company *domain.Company) error {
	return r.db.WithContext(ctx).Save(company).Error
}

// UpdateDefaultResponsibleUsers updates only the default responsible user fields
func (r *CompanyRepository) UpdateDefaultResponsibleUsers(ctx context.Context, id domain.CompanyID, offerResponsibleID, projectResponsibleID *string) error {
	updates := map[string]interface{}{
		"default_offer_responsible_id":   offerResponsibleID,
		"default_project_responsible_id": projectResponsibleID,
	}
	return r.db.WithContext(ctx).
		Model(&domain.Company{}).
		Where("id = ?", id).
		Updates(updates).Error
}
