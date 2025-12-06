package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserRoleRepository handles user role data access
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

// GetByUserIDAndCompany returns active roles for a user in a specific company
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

// HasRole checks if a user has a specific role (active and not expired)
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

// GetRoleTypes returns just the role types for a user (for use in UserContext)
func (r *UserRoleRepository) GetRoleTypes(ctx context.Context, userID string) ([]domain.UserRoleType, error) {
	roles, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleTypes := make([]domain.UserRoleType, len(roles))
	for i, role := range roles {
		roleTypes[i] = role.Role
	}
	return roleTypes, nil
}
