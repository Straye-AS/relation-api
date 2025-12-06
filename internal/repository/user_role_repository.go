package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRoleRepository handles database operations for user roles
type UserRoleRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *gorm.DB, logger *zap.Logger) *UserRoleRepository {
	return &UserRoleRepository{
		db:     db,
		logger: logger,
	}
}

// GetByUserID returns all active roles for a user
func (r *UserRoleRepository) GetByUserID(ctx context.Context, userID string) ([]domain.UserRole, error) {
	var roles []domain.UserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&roles).Error
	if err != nil {
		r.logger.Error("failed to get user roles", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	return roles, nil
}

// GetByUserIDAndCompany returns roles for a user in a specific company
func (r *UserRoleRepository) GetByUserIDAndCompany(ctx context.Context, userID string, companyID domain.CompanyID) ([]domain.UserRole, error) {
	var roles []domain.UserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true", userID).
		Where("company_id = ? OR company_id IS NULL", companyID). // Global roles + company-specific
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&roles).Error
	if err != nil {
		r.logger.Error("failed to get user roles by company",
			zap.String("user_id", userID),
			zap.String("company_id", string(companyID)),
			zap.Error(err))
		return nil, err
	}
	return roles, nil
}

// GetRoleTypes returns just the role types for a user (for quick permission checks)
func (r *UserRoleRepository) GetRoleTypes(ctx context.Context, userID string) ([]domain.UserRoleType, error) {
	var roleTypes []domain.UserRoleType
	err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("user_id = ? AND is_active = true", userID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Pluck("role", &roleTypes).Error
	if err != nil {
		r.logger.Error("failed to get user role types", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}
	return roleTypes, nil
}

// Create creates a new user role assignment
func (r *UserRoleRepository) Create(ctx context.Context, role *domain.UserRole) error {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		r.logger.Error("failed to create user role",
			zap.String("user_id", role.UserID),
			zap.String("role", string(role.Role)),
			zap.Error(err))
		return err
	}
	return nil
}

// Delete removes a user role assignment
func (r *UserRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.UserRole{}, "id = ?", id).Error; err != nil {
		r.logger.Error("failed to delete user role", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

// Deactivate soft-deletes a role by setting is_active to false
func (r *UserRoleRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("id = ?", id).
		Update("is_active", false).Error; err != nil {
		r.logger.Error("failed to deactivate user role", zap.String("id", id.String()), zap.Error(err))
		return err
	}
	return nil
}

// HasRole checks if a user has a specific role
func (r *UserRoleRepository) HasRole(ctx context.Context, userID string, role domain.UserRoleType) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("user_id = ? AND role = ? AND is_active = true", userID, role).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	if err != nil {
		r.logger.Error("failed to check user role",
			zap.String("user_id", userID),
			zap.String("role", string(role)),
			zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// HasRoleInCompany checks if a user has a specific role in a specific company
func (r *UserRoleRepository) HasRoleInCompany(ctx context.Context, userID string, role domain.UserRoleType, companyID domain.CompanyID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("user_id = ? AND role = ? AND company_id = ? AND is_active = true", userID, role, companyID).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Count(&count).Error
	if err != nil {
		r.logger.Error("failed to check user role in company",
			zap.String("user_id", userID),
			zap.String("role", string(role)),
			zap.String("company_id", string(companyID)),
			zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

// AssignRole assigns a role to a user
func (r *UserRoleRepository) AssignRole(ctx context.Context, userID string, role domain.UserRoleType, companyID *domain.CompanyID, grantedBy string, expiresAt *time.Time) (*domain.UserRole, error) {
	userRole := &domain.UserRole{
		ID:        uuid.New(),
		UserID:    userID,
		Role:      role,
		CompanyID: companyID,
		GrantedBy: grantedBy,
		GrantedAt: time.Now(),
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	err := r.db.WithContext(ctx).Create(userRole).Error
	if err != nil {
		r.logger.Error("failed to assign role",
			zap.String("user_id", userID),
			zap.String("role", string(role)),
			zap.Error(err))
		return nil, err
	}
	return userRole, nil
}

// RemoveRole deactivates a role assignment (soft delete)
func (r *UserRoleRepository) RemoveRole(ctx context.Context, userID string, role domain.UserRoleType, companyID *domain.CompanyID) error {
	query := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("user_id = ? AND role = ? AND is_active = true", userID, role)

	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	} else {
		query = query.Where("company_id IS NULL")
	}

	if err := query.Update("is_active", false).Error; err != nil {
		r.logger.Error("failed to remove role",
			zap.String("user_id", userID),
			zap.String("role", string(role)),
			zap.Error(err))
		return err
	}
	return nil
}

// RemoveRoleByID deactivates a specific role assignment by ID
func (r *UserRoleRepository) RemoveRoleByID(ctx context.Context, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("id = ?", roleID).
		Update("is_active", false).Error
}

// RemoveAllRoles deactivates all roles for a user
func (r *UserRoleRepository) RemoveAllRoles(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("user_id = ? AND is_active = true", userID).
		Update("is_active", false).Error
}

// GetExpiredRoles returns roles that have expired but are still marked active
func (r *UserRoleRepository) GetExpiredRoles(ctx context.Context) ([]domain.UserRole, error) {
	var roles []domain.UserRole
	err := r.db.WithContext(ctx).
		Where("is_active = true AND expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// DeactivateExpiredRoles marks expired roles as inactive
func (r *UserRoleRepository) DeactivateExpiredRoles(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("is_active = true AND expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Update("is_active", false)
	return result.RowsAffected, result.Error
}

// GetByID returns a role assignment by ID
func (r *UserRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserRole, error) {
	var role domain.UserRole
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// ListAll returns all active role assignments (for admin purposes)
func (r *UserRoleRepository) ListAll(ctx context.Context, limit, offset int) ([]domain.UserRole, int64, error) {
	var roles []domain.UserRole
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.UserRole{}).Where("is_active = true")

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.
		Preload("Company").
		Order("granted_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

// UpdateRole updates a role assignment
func (r *UserRoleRepository) UpdateRole(ctx context.Context, role *domain.UserRole) error {
	return r.db.WithContext(ctx).Save(role).Error
}
