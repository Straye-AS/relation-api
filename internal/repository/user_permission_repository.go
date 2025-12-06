package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserPermissionRepository handles database operations for permission overrides
type UserPermissionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserPermissionRepository creates a new user permission repository
func NewUserPermissionRepository(db *gorm.DB, logger *zap.Logger) *UserPermissionRepository {
	return &UserPermissionRepository{
		db:     db,
		logger: logger,
	}
}

// GetByUserID returns all active permission overrides for a user
func (r *UserPermissionRepository) GetByUserID(ctx context.Context, userID string) ([]domain.UserPermission, error) {
	var perms []domain.UserPermission
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// GetByUserIDAndCompany returns permission overrides for a user in a specific company
func (r *UserPermissionRepository) GetByUserIDAndCompany(ctx context.Context, userID string, companyID domain.CompanyID) ([]domain.UserPermission, error) {
	var perms []domain.UserPermission
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND (company_id = ? OR company_id IS NULL)", userID, companyID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("company_id NULLS LAST"). // Global overrides first, then company-specific
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// GetPermissionOverride checks for a specific permission override
func (r *UserPermissionRepository) GetPermissionOverride(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID) (*domain.UserPermission, error) {
	var perm domain.UserPermission

	query := r.db.WithContext(ctx).
		Where("user_id = ? AND permission = ?", userID, permission).
		Where("expires_at IS NULL OR expires_at > ?", time.Now())

	if companyID != nil {
		// Check for company-specific or global override
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID).
			Order("company_id DESC NULLS LAST") // Company-specific takes precedence
	} else {
		query = query.Where("company_id IS NULL")
	}

	err := query.First(&perm).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// GrantPermission creates a permission grant override
func (r *UserPermissionRepository) GrantPermission(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID, grantedBy string, reason string, expiresAt *time.Time) (*domain.UserPermission, error) {
	// First check if an override already exists
	existing, err := r.GetPermissionOverride(ctx, userID, permission, companyID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing != nil {
		// Update existing override
		existing.IsGranted = true
		existing.GrantedBy = grantedBy
		existing.GrantedAt = now
		existing.Reason = reason
		existing.ExpiresAt = expiresAt
		existing.UpdatedAt = now
		err = r.db.WithContext(ctx).Save(existing).Error
		return existing, err
	}

	// Create new override
	perm := &domain.UserPermission{
		ID:         uuid.New(),
		UserID:     userID,
		Permission: permission,
		CompanyID:  companyID,
		IsGranted:  true,
		GrantedBy:  grantedBy,
		GrantedAt:  now,
		ExpiresAt:  expiresAt,
		Reason:     reason,
	}

	err = r.db.WithContext(ctx).Create(perm).Error
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// DenyPermission creates a permission denial override
func (r *UserPermissionRepository) DenyPermission(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID, grantedBy string, reason string, expiresAt *time.Time) (*domain.UserPermission, error) {
	// First check if an override already exists
	existing, err := r.GetPermissionOverride(ctx, userID, permission, companyID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing != nil {
		// Update existing override
		existing.IsGranted = false
		existing.GrantedBy = grantedBy
		existing.GrantedAt = now
		existing.Reason = reason
		existing.ExpiresAt = expiresAt
		existing.UpdatedAt = now
		err = r.db.WithContext(ctx).Save(existing).Error
		return existing, err
	}

	// Create new override
	perm := &domain.UserPermission{
		ID:         uuid.New(),
		UserID:     userID,
		Permission: permission,
		CompanyID:  companyID,
		IsGranted:  false,
		GrantedBy:  grantedBy,
		GrantedAt:  now,
		ExpiresAt:  expiresAt,
		Reason:     reason,
	}

	err = r.db.WithContext(ctx).Create(perm).Error
	if err != nil {
		return nil, err
	}
	return perm, nil
}

// RemoveOverride removes a permission override (hard delete)
func (r *UserPermissionRepository) RemoveOverride(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID) error {
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND permission = ?", userID, permission)

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	} else {
		query = query.Where("company_id IS NULL")
	}

	return query.Delete(&domain.UserPermission{}).Error
}

// RemoveOverrideByID removes a specific permission override by ID
func (r *UserPermissionRepository) RemoveOverrideByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserPermission{}, "id = ?", id).Error
}

// RemoveAllOverrides removes all permission overrides for a user
func (r *UserPermissionRepository) RemoveAllOverrides(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.UserPermission{}).Error
}

// GetExpiredOverrides returns overrides that have expired
func (r *UserPermissionRepository) GetExpiredOverrides(ctx context.Context) ([]domain.UserPermission, error) {
	var perms []domain.UserPermission
	err := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// DeleteExpiredOverrides removes expired permission overrides
func (r *UserPermissionRepository) DeleteExpiredOverrides(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Delete(&domain.UserPermission{})
	return result.RowsAffected, result.Error
}

// GetByID returns a permission override by ID
func (r *UserPermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserPermission, error) {
	var perm domain.UserPermission
	err := r.db.WithContext(ctx).First(&perm, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// ListAll returns all permission overrides (for admin purposes)
func (r *UserPermissionRepository) ListAll(ctx context.Context, limit, offset int) ([]domain.UserPermission, int64, error) {
	var perms []domain.UserPermission
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.UserPermission{})

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Preload("Company").
		Order("granted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&perms).Error
	if err != nil {
		return nil, 0, err
	}

	return perms, total, nil
}

// ListByPermission returns all overrides for a specific permission
func (r *UserPermissionRepository) ListByPermission(ctx context.Context, permission domain.PermissionType) ([]domain.UserPermission, error) {
	var perms []domain.UserPermission
	err := r.db.WithContext(ctx).
		Where("permission = ?", permission).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// GetGrantedPermissions returns all granted permission overrides for a user
func (r *UserPermissionRepository) GetGrantedPermissions(ctx context.Context, userID string) ([]domain.PermissionType, error) {
	var perms []domain.PermissionType
	err := r.db.WithContext(ctx).
		Model(&domain.UserPermission{}).
		Where("user_id = ? AND is_granted = true", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Pluck("permission", &perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

// GetDeniedPermissions returns all denied permission overrides for a user
func (r *UserPermissionRepository) GetDeniedPermissions(ctx context.Context, userID string) ([]domain.PermissionType, error) {
	var perms []domain.PermissionType
	err := r.db.WithContext(ctx).
		Model(&domain.UserPermission{}).
		Where("user_id = ? AND is_granted = false", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Pluck("permission", &perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}
