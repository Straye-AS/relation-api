package service

import (
	"context"

	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

// PermissionService handles permission checking with database overrides
type PermissionService struct {
	userRoleRepo       *repository.UserRoleRepository
	userPermissionRepo *repository.UserPermissionRepository
	logger             *zap.Logger
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	userRoleRepo *repository.UserRoleRepository,
	userPermissionRepo *repository.UserPermissionRepository,
	logger *zap.Logger,
) *PermissionService {
	return &PermissionService{
		userRoleRepo:       userRoleRepo,
		userPermissionRepo: userPermissionRepo,
		logger:             logger,
	}
}

// CheckPermission checks if a user has a specific permission
// This considers: 1) Super admin status, 2) Permission overrides, 3) Role-based permissions
func (s *PermissionService) CheckPermission(ctx context.Context, userCtx *auth.UserContext, permission domain.PermissionType) (bool, error) {
	// Super admins have all permissions
	if userCtx.IsSuperAdmin() {
		return true, nil
	}

	// Check for permission override in database
	var companyID *domain.CompanyID
	if userCtx.CompanyID != "" {
		companyID = &userCtx.CompanyID
	}

	override, err := s.userPermissionRepo.GetPermissionOverride(ctx, userCtx.UserID.String(), permission, companyID)
	if err != nil {
		s.logger.Error("failed to check permission override",
			zap.String("user_id", userCtx.UserID.String()),
			zap.String("permission", string(permission)),
			zap.Error(err))
		// Fall back to role-based check on error
		return userCtx.HasPermission(permission), nil
	}

	// If there's an override, use it
	if override != nil {
		return override.IsGranted, nil
	}

	// Fall back to role-based permissions
	return userCtx.HasPermission(permission), nil
}

// CheckPermissionForCompany checks permission for a specific company context
func (s *PermissionService) CheckPermissionForCompany(ctx context.Context, userCtx *auth.UserContext, permission domain.PermissionType, companyID domain.CompanyID) (bool, error) {
	// Super admins have all permissions
	if userCtx.IsSuperAdmin() {
		return true, nil
	}

	// User must be able to access the company first
	if !userCtx.CanAccessCompany(companyID) {
		return false, nil
	}

	// Check for permission override
	override, err := s.userPermissionRepo.GetPermissionOverride(ctx, userCtx.UserID.String(), permission, &companyID)
	if err != nil {
		s.logger.Error("failed to check permission override for company",
			zap.String("user_id", userCtx.UserID.String()),
			zap.String("permission", string(permission)),
			zap.String("company_id", string(companyID)),
			zap.Error(err))
		return userCtx.HasPermission(permission), nil
	}

	if override != nil {
		return override.IsGranted, nil
	}

	return userCtx.HasPermission(permission), nil
}

// GetEffectivePermissions returns all permissions a user has (from roles + overrides)
func (s *PermissionService) GetEffectivePermissions(ctx context.Context, userCtx *auth.UserContext) ([]domain.PermissionType, error) {
	// Super admins have all permissions
	if userCtx.IsSuperAdmin() {
		return getAllPermissions(), nil
	}

	// Start with role-based permissions
	permissions := make(map[domain.PermissionType]bool)
	for _, perm := range getRolePermissions(userCtx.Roles) {
		permissions[perm] = true
	}

	// Apply overrides from database
	var companyID *domain.CompanyID
	if userCtx.CompanyID != "" {
		companyID = &userCtx.CompanyID
	}

	overrides, err := s.userPermissionRepo.GetByUserIDAndCompany(ctx, userCtx.UserID.String(), *companyID)
	if err != nil {
		s.logger.Error("failed to get permission overrides",
			zap.String("user_id", userCtx.UserID.String()),
			zap.Error(err))
		// Return role-based permissions on error
		return mapKeysToSlice(permissions), nil
	}

	for _, override := range overrides {
		if override.IsGranted {
			permissions[override.Permission] = true
		} else {
			delete(permissions, override.Permission)
		}
	}

	return mapKeysToSlice(permissions), nil
}

// GetUserRoles returns all active roles for a user from the database
func (s *PermissionService) GetUserRoles(ctx context.Context, userID string) ([]domain.UserRoleType, error) {
	return s.userRoleRepo.GetRoleTypes(ctx, userID)
}

// CanPerformAction is a convenience method for common permission checks
func (s *PermissionService) CanPerformAction(ctx context.Context, userCtx *auth.UserContext, action string, resource string) (bool, error) {
	permission := domain.PermissionType(resource + ":" + action)
	return s.CheckPermission(ctx, userCtx, permission)
}

// RequirePermission returns an error if user doesn't have the permission
func (s *PermissionService) RequirePermission(ctx context.Context, userCtx *auth.UserContext, permission domain.PermissionType) error {
	hasPermission, err := s.CheckPermission(ctx, userCtx, permission)
	if err != nil {
		return err
	}
	if !hasPermission {
		return ErrPermissionDenied
	}
	return nil
}

// getRolePermissions returns all permissions for a set of roles
func getRolePermissions(roles []domain.UserRoleType) []domain.PermissionType {
	permMap := make(map[domain.PermissionType]bool)

	rolePermissions := map[domain.UserRoleType][]domain.PermissionType{
		domain.RoleCompanyAdmin: {
			domain.PermissionCustomersRead, domain.PermissionCustomersWrite, domain.PermissionCustomersDelete,
			domain.PermissionContactsRead, domain.PermissionContactsWrite, domain.PermissionContactsDelete,
			domain.PermissionDealsRead, domain.PermissionDealsWrite, domain.PermissionDealsDelete,
			domain.PermissionOffersRead, domain.PermissionOffersWrite, domain.PermissionOffersDelete, domain.PermissionOffersApprove,
			domain.PermissionProjectsRead, domain.PermissionProjectsWrite, domain.PermissionProjectsDelete,
			domain.PermissionBudgetsRead, domain.PermissionBudgetsWrite,
			domain.PermissionActivitiesRead, domain.PermissionActivitiesWrite, domain.PermissionActivitiesDelete,
			domain.PermissionUsersRead, domain.PermissionUsersWrite, domain.PermissionUsersManageRoles,
			domain.PermissionReportsView, domain.PermissionReportsExport,
			domain.PermissionSystemAuditLogs,
		},
		domain.RoleManager: {
			domain.PermissionCustomersRead, domain.PermissionCustomersWrite,
			domain.PermissionContactsRead, domain.PermissionContactsWrite,
			domain.PermissionDealsRead, domain.PermissionDealsWrite,
			domain.PermissionOffersRead, domain.PermissionOffersWrite, domain.PermissionOffersApprove,
			domain.PermissionProjectsRead, domain.PermissionProjectsWrite,
			domain.PermissionBudgetsRead, domain.PermissionBudgetsWrite,
			domain.PermissionActivitiesRead, domain.PermissionActivitiesWrite,
			domain.PermissionUsersRead,
			domain.PermissionReportsView, domain.PermissionReportsExport,
		},
		domain.RoleSales: {
			domain.PermissionCustomersRead, domain.PermissionCustomersWrite,
			domain.PermissionContactsRead, domain.PermissionContactsWrite,
			domain.PermissionDealsRead, domain.PermissionDealsWrite,
			domain.PermissionOffersRead, domain.PermissionOffersWrite,
			domain.PermissionProjectsRead,
			domain.PermissionBudgetsRead,
			domain.PermissionActivitiesRead, domain.PermissionActivitiesWrite,
			domain.PermissionReportsView,
		},
		domain.RoleProjectManager: {
			domain.PermissionCustomersRead,
			domain.PermissionContactsRead, domain.PermissionContactsWrite,
			domain.PermissionDealsRead,
			domain.PermissionOffersRead,
			domain.PermissionProjectsRead, domain.PermissionProjectsWrite,
			domain.PermissionBudgetsRead, domain.PermissionBudgetsWrite,
			domain.PermissionActivitiesRead, domain.PermissionActivitiesWrite,
			domain.PermissionReportsView,
		},
		domain.RoleViewer: {
			domain.PermissionCustomersRead,
			domain.PermissionContactsRead,
			domain.PermissionDealsRead,
			domain.PermissionOffersRead,
			domain.PermissionProjectsRead,
			domain.PermissionBudgetsRead,
			domain.PermissionActivitiesRead,
			domain.PermissionReportsView,
		},
		domain.RoleAPIService: {
			domain.PermissionCustomersRead, domain.PermissionCustomersWrite,
			domain.PermissionContactsRead, domain.PermissionContactsWrite,
			domain.PermissionDealsRead, domain.PermissionDealsWrite,
			domain.PermissionOffersRead, domain.PermissionOffersWrite,
			domain.PermissionProjectsRead, domain.PermissionProjectsWrite,
			domain.PermissionBudgetsRead, domain.PermissionBudgetsWrite,
			domain.PermissionActivitiesRead, domain.PermissionActivitiesWrite,
		},
	}

	for _, role := range roles {
		if perms, ok := rolePermissions[role]; ok {
			for _, perm := range perms {
				permMap[perm] = true
			}
		}
	}

	return mapKeysToSlice(permMap)
}

// getAllPermissions returns all defined permissions
func getAllPermissions() []domain.PermissionType {
	return []domain.PermissionType{
		domain.PermissionCustomersRead, domain.PermissionCustomersWrite, domain.PermissionCustomersDelete,
		domain.PermissionContactsRead, domain.PermissionContactsWrite, domain.PermissionContactsDelete,
		domain.PermissionDealsRead, domain.PermissionDealsWrite, domain.PermissionDealsDelete,
		domain.PermissionOffersRead, domain.PermissionOffersWrite, domain.PermissionOffersDelete, domain.PermissionOffersApprove,
		domain.PermissionProjectsRead, domain.PermissionProjectsWrite, domain.PermissionProjectsDelete,
		domain.PermissionBudgetsRead, domain.PermissionBudgetsWrite,
		domain.PermissionActivitiesRead, domain.PermissionActivitiesWrite, domain.PermissionActivitiesDelete,
		domain.PermissionUsersRead, domain.PermissionUsersWrite, domain.PermissionUsersManageRoles,
		domain.PermissionCompaniesRead, domain.PermissionCompaniesWrite,
		domain.PermissionReportsView, domain.PermissionReportsExport,
		domain.PermissionSystemAdmin, domain.PermissionSystemAuditLogs,
	}
}

func mapKeysToSlice(m map[domain.PermissionType]bool) []domain.PermissionType {
	result := make([]domain.PermissionType, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
