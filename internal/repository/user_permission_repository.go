package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserPermissionRepository handles user permission override data access
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
	var permissions []domain.UserPermission
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&permissions).Error
	if err != nil {
		r.logger.Error("failed to get user permissions", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	return permissions, nil
}

// GetByUserIDAndCompany returns permission overrides for a user in a specific company
func (r *UserPermissionRepository) GetByUserIDAndCompany(ctx context.Context, userID string, companyID domain.CompanyID) ([]domain.UserPermission, error) {
	var permissions []domain.UserPermission
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("company_id = ? OR company_id IS NULL", companyID). // Global overrides + company-specific
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&permissions).Error
	if err != nil {
		r.logger.Error("failed to get user permissions by company",
			zap.String("user_id", userID),
			zap.String("company_id", string(companyID)),
			zap.Error(err))
		return nil, err
	}
	return permissions, nil
}

// Create creates a new permission override
func (r *UserPermissionRepository) Create(ctx context.Context, permission *domain.UserPermission) error {
	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		r.logger.Error("failed to create user permission",
			zap.String("user_id", permission.UserID),
			zap.String("permission", string(permission.Permission)),
			zap.Error(err))
		return err
	}
	return nil
}

// Delete removes a permission override
func (r *UserPermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.UserPermission{}, "id = ?", id).Error; err != nil {
		r.logger.Error("failed to delete user permission", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

// GetPermissionOverride checks if a user has a specific permission override
// Returns: (permission exists, is_granted value, error)
func (r *UserPermissionRepository) GetPermissionOverride(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID) (*domain.UserPermission, error) {
	var perm domain.UserPermission
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND permission = ?", userID, permission).
		Where("expires_at IS NULL OR expires_at > ?", time.Now())

	if companyID != nil {
		// Check for company-specific override first, then global
		query = query.Where("company_id = ? OR company_id IS NULL", *companyID).
			Order("company_id DESC NULLS LAST") // Company-specific takes precedence
	} else {
		query = query.Where("company_id IS NULL")
	}

	err := query.First(&perm).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No override exists
		}
		r.logger.Error("failed to get permission override",
			zap.String("user_id", userID),
			zap.String("permission", string(permission)),
			zap.Error(err))
		return nil, err
	}
	return &perm, nil
}

// GrantPermission creates a permission grant override
func (r *UserPermissionRepository) GrantPermission(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID, grantedBy string, reason string) error {
	perm := &domain.UserPermission{
		UserID:     userID,
		Permission: permission,
		CompanyID:  companyID,
		IsGranted:  true,
		GrantedBy:  grantedBy,
		GrantedAt:  time.Now(),
		Reason:     reason,
	}
	return r.Create(ctx, perm)
}

// DenyPermission creates a permission deny override
func (r *UserPermissionRepository) DenyPermission(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID, grantedBy string, reason string) error {
	perm := &domain.UserPermission{
		UserID:     userID,
		Permission: permission,
		CompanyID:  companyID,
		IsGranted:  false,
		GrantedBy:  grantedBy,
		GrantedAt:  time.Now(),
		Reason:     reason,
	}
	return r.Create(ctx, perm)
}

// RevokePermission removes a permission override (restores default role-based permission)
func (r *UserPermissionRepository) RevokePermission(ctx context.Context, userID string, permission domain.PermissionType, companyID *domain.CompanyID) error {
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND permission = ?", userID, permission)

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	} else {
		query = query.Where("company_id IS NULL")
	}

	if err := query.Delete(&domain.UserPermission{}).Error; err != nil {
		r.logger.Error("failed to revoke permission",
			zap.String("user_id", userID),
			zap.String("permission", string(permission)),
			zap.Error(err))
		return err
	}
	return nil
}
