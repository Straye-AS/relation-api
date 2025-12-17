package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

// RoleService handles role management operations
type RoleService struct {
	userRoleRepo *repository.UserRoleRepository
	activityRepo *repository.ActivityRepository
	logger       *zap.Logger
}

// NewRoleService creates a new role service
func NewRoleService(
	userRoleRepo *repository.UserRoleRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *RoleService {
	return &RoleService{
		userRoleRepo: userRoleRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

// GetUserRoles returns all active roles for a user
func (s *RoleService) GetUserRoles(ctx context.Context, userID string) ([]domain.UserRole, error) {
	return s.userRoleRepo.GetByUserID(ctx, userID)
}

// GetUserRoleTypes returns just the role types for a user
func (s *RoleService) GetUserRoleTypes(ctx context.Context, userID string) ([]domain.UserRoleType, error) {
	return s.userRoleRepo.GetRoleTypes(ctx, userID)
}

// HasRole checks if a user has a specific role
func (s *RoleService) HasRole(ctx context.Context, userID string, role domain.UserRoleType) (bool, error) {
	return s.userRoleRepo.HasRole(ctx, userID, role)
}

// AssignRoleInput contains the input for assigning a role
type AssignRoleInput struct {
	UserID    string
	Role      domain.UserRoleType
	CompanyID *domain.CompanyID
	ExpiresAt *time.Time
}

// AssignRole assigns a role to a user
func (s *RoleService) AssignRole(ctx context.Context, input AssignRoleInput) (*domain.UserRole, error) {
	// Validate role
	if !isValidRole(input.Role) {
		return nil, ErrInvalidRole
	}

	// Check if role is already assigned
	hasRole, err := s.userRoleRepo.HasRole(ctx, input.UserID, input.Role)
	if err != nil {
		return nil, err
	}
	if hasRole {
		return nil, ErrRoleAlreadyAssigned
	}

	// Get the granting user from context
	grantedBy := "system"
	if userCtx, ok := auth.FromContext(ctx); ok {
		grantedBy = userCtx.UserID.String()
	}

	// Assign the role
	role, err := s.userRoleRepo.AssignRole(ctx, input.UserID, input.Role, input.CompanyID, grantedBy, input.ExpiresAt)
	if err != nil {
		return nil, err
	}

	// Log the activity
	s.logRoleChange(ctx, input.UserID, input.Role, "tildelt", grantedBy)

	return role, nil
}

// RemoveRole removes a role from a user
func (s *RoleService) RemoveRole(ctx context.Context, userID string, role domain.UserRoleType, companyID *domain.CompanyID) error {
	// Check if this is the last super admin
	if role == domain.RoleSuperAdmin {
		isLast, err := s.isLastSuperAdmin(ctx, userID)
		if err != nil {
			return err
		}
		if isLast {
			return ErrCannotRemoveLastAdmin
		}
	}

	// Get the granting user from context
	removedBy := "system"
	if userCtx, ok := auth.FromContext(ctx); ok {
		removedBy = userCtx.UserID.String()
	}

	err := s.userRoleRepo.RemoveRole(ctx, userID, role, companyID)
	if err != nil {
		return err
	}

	// Log the activity
	s.logRoleChange(ctx, userID, role, "fjernet", removedBy)

	return nil
}

// RemoveRoleByID removes a specific role assignment by ID
func (s *RoleService) RemoveRoleByID(ctx context.Context, roleID uuid.UUID) error {
	// Get the role first to check if it's the last super admin
	role, err := s.userRoleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return ErrRoleNotFound
	}

	if role.Role == domain.RoleSuperAdmin {
		isLast, err := s.isLastSuperAdmin(ctx, role.UserID)
		if err != nil {
			return err
		}
		if isLast {
			return ErrCannotRemoveLastAdmin
		}
	}

	// Get the granting user from context
	removedBy := "system"
	if userCtx, ok := auth.FromContext(ctx); ok {
		removedBy = userCtx.UserID.String()
	}

	err = s.userRoleRepo.RemoveRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	// Log the activity
	s.logRoleChange(ctx, role.UserID, role.Role, "fjernet", removedBy)

	return nil
}

// ReplaceUserRoles removes all existing roles and assigns new ones
func (s *RoleService) ReplaceUserRoles(ctx context.Context, userID string, roles []domain.UserRoleType, companyID *domain.CompanyID) error {
	// Validate all roles first
	for _, role := range roles {
		if !isValidRole(role) {
			return ErrInvalidRole
		}
	}

	// Get the granting user from context
	grantedBy := "system"
	if userCtx, ok := auth.FromContext(ctx); ok {
		grantedBy = userCtx.UserID.String()
	}

	// Remove all existing roles
	err := s.userRoleRepo.RemoveAllRoles(ctx, userID)
	if err != nil {
		return err
	}

	// Assign new roles
	for _, role := range roles {
		_, err := s.userRoleRepo.AssignRole(ctx, userID, role, companyID, grantedBy, nil)
		if err != nil {
			return err
		}
	}

	s.logger.Info("replaced user roles",
		zap.String("user_id", userID),
		zap.Any("roles", roles),
		zap.String("granted_by", grantedBy))

	return nil
}

// ListAllRoles returns all active role assignments (paginated)
func (s *RoleService) ListAllRoles(ctx context.Context, page, pageSize int) ([]domain.UserRole, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.userRoleRepo.ListAll(ctx, pageSize, offset)
}

// CleanupExpiredRoles deactivates expired roles
func (s *RoleService) CleanupExpiredRoles(ctx context.Context) (int64, error) {
	count, err := s.userRoleRepo.DeactivateExpiredRoles(ctx)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		s.logger.Info("deactivated expired roles", zap.Int64("count", count))
	}
	return count, nil
}

// isLastSuperAdmin checks if a user is the last active super admin
func (s *RoleService) isLastSuperAdmin(ctx context.Context, userID string) (bool, error) {
	// This would need to count all super admins
	// For now, we'll assume there are multiple super admins
	// A full implementation would query the database
	return false, nil
}

// logRoleChange logs a role change activity
func (s *RoleService) logRoleChange(ctx context.Context, userID string, role domain.UserRoleType, action string, changedBy string) {
	if s.activityRepo == nil {
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Warn("could not parse user ID for activity log", zap.String("user_id", userID))
		return
	}

	activity := &domain.Activity{
		TargetType:   domain.ActivityTargetType("User"),
		TargetID:     userUUID,
		TargetName:   string(role),
		Title:        "Rolle " + action,
		Body:         "Rollen '" + string(role) + "' ble " + action + " av " + changedBy,
		ActivityType: domain.ActivityTypeSystem,
		Status:       domain.ActivityStatusCompleted,
		CreatorID:    changedBy,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log role change activity", zap.Error(err))
	}
}

// isValidRole checks if a role type is valid
func isValidRole(role domain.UserRoleType) bool {
	validRoles := []domain.UserRoleType{
		domain.RoleSuperAdmin,
		domain.RoleCompanyAdmin,
		domain.RoleManager,
		domain.RoleMarket,
		domain.RoleProjectManager,
		domain.RoleProjectLeader,
		domain.RoleViewer,
		domain.RoleAPIService,
	}

	for _, r := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

// GetValidRoles returns all valid role types
func GetValidRoles() []domain.UserRoleType {
	return []domain.UserRoleType{
		domain.RoleSuperAdmin,
		domain.RoleCompanyAdmin,
		domain.RoleManager,
		domain.RoleMarket,
		domain.RoleProjectManager,
		domain.RoleProjectLeader,
		domain.RoleViewer,
		domain.RoleAPIService,
	}
}
