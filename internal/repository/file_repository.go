package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	// Omit associations to avoid GORM trying to validate/create related records
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(file).Error
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.File, error) {
	var file domain.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// ListByOffer returns all files attached to an offer, filtered by company access
// If companyFilter is nil, all files are returned (for gruppen users)
// If companyFilter is provided, only files matching that company OR "gruppen" are returned
func (r *FileRepository) ListByOffer(ctx context.Context, offerID uuid.UUID, companyFilter *domain.CompanyID) ([]domain.File, error) {
	var files []domain.File
	query := r.db.WithContext(ctx).Where("offer_id = ?", offerID)
	query = r.applyCompanyFilter(query, companyFilter)
	err := query.Order("created_at DESC").Find(&files).Error
	return files, err
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.File{}, "id = ?", id).Error
}

// ListByCustomer returns all files attached to a customer, filtered by company access
// If companyFilter is nil, all files are returned (for gruppen users)
// If companyFilter is provided, only files matching that company OR "gruppen" are returned
func (r *FileRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID, companyFilter *domain.CompanyID) ([]domain.File, error) {
	var files []domain.File
	query := r.db.WithContext(ctx).Where("customer_id = ?", customerID)
	query = r.applyCompanyFilter(query, companyFilter)
	err := query.Order("created_at DESC").Find(&files).Error
	return files, err
}

// ListByProject returns all files attached to a project, filtered by company access
// If companyFilter is nil, all files are returned (for gruppen users)
// If companyFilter is provided, only files matching that company OR "gruppen" are returned
func (r *FileRepository) ListByProject(ctx context.Context, projectID uuid.UUID, companyFilter *domain.CompanyID) ([]domain.File, error) {
	var files []domain.File
	query := r.db.WithContext(ctx).Where("project_id = ?", projectID)
	query = r.applyCompanyFilter(query, companyFilter)
	err := query.Order("created_at DESC").Find(&files).Error
	return files, err
}

// ListBySupplier returns all files attached to a supplier, filtered by company access
// If companyFilter is nil, all files are returned (for gruppen users)
// If companyFilter is provided, only files matching that company OR "gruppen" are returned
func (r *FileRepository) ListBySupplier(ctx context.Context, supplierID uuid.UUID, companyFilter *domain.CompanyID) ([]domain.File, error) {
	var files []domain.File
	query := r.db.WithContext(ctx).Where("supplier_id = ?", supplierID)
	query = r.applyCompanyFilter(query, companyFilter)
	err := query.Order("created_at DESC").Find(&files).Error
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

// ListByOfferSupplier returns all files attached to an offer-supplier relationship, filtered by company access
// If companyFilter is nil, all files are returned (for gruppen users)
// If companyFilter is provided, only files matching that company OR "gruppen" are returned
func (r *FileRepository) ListByOfferSupplier(ctx context.Context, offerSupplierID uuid.UUID, companyFilter *domain.CompanyID) ([]domain.File, error) {
	var files []domain.File
	query := r.db.WithContext(ctx).Where("offer_supplier_id = ?", offerSupplierID)
	query = r.applyCompanyFilter(query, companyFilter)
	err := query.Order("created_at DESC").Find(&files).Error
	return files, err
}

// CountByOfferSupplier returns the count of files attached to an offer-supplier relationship
func (r *FileRepository) CountByOfferSupplier(ctx context.Context, offerSupplierID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.File{}).
		Where("offer_supplier_id = ?", offerSupplierID).
		Count(&count).Error
	return count, err
}

// applyCompanyFilter adds company-based filtering to a query.
// If companyFilter is nil, no filtering is applied (user can see all files).
// If companyFilter is provided, only files matching that company OR "gruppen" are returned.
// This enables multi-tenant filtering where gruppen files are always visible.
func (r *FileRepository) applyCompanyFilter(query *gorm.DB, companyFilter *domain.CompanyID) *gorm.DB {
	if companyFilter == nil {
		// No filter - user can see all companies (gruppen user or super admin)
		return query
	}
	// Filter: show files from user's company OR gruppen (parent company files visible to all)
	return query.Where("company_id = ? OR company_id = ?", *companyFilter, domain.CompanyGruppen)
}
