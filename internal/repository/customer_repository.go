package repository

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *CustomerRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	var customer domain.Customer
	query := r.db.WithContext(ctx).Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) Update(ctx context.Context, customer *domain.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Customer{}, "id = ?", id).Error
}

func (r *CustomerRepository) List(ctx context.Context, page, pageSize int, search string) ([]domain.Customer, int64, error) {
	var customers []domain.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Customer{})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&customers).Error

	return customers, total, err
}

func (r *CustomerRepository) GetContactsCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Contact{}).Where("customer_id = ?", customerID).Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) GetProjectsCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Project{}).Where("customer_id = ?", customerID).Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) GetOffersCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Offer{}).Where("customer_id = ?", customerID).Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) Count(ctx context.Context) (int, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Customer{})
	query = ApplyCompanyFilter(ctx, query)
	err := query.Count(&count).Error
	return int(count), err
}

func (r *CustomerRepository) Search(ctx context.Context, searchQuery string, limit int) ([]domain.Customer, error) {
	var customers []domain.Customer
	searchPattern := "%" + strings.ToLower(searchQuery) + "%"
	query := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ? OR LOWER(org_number) LIKE ?", searchPattern, searchPattern)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Limit(limit).Find(&customers).Error
	return customers, err
}
