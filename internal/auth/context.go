package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
)

// UserContext holds authenticated user information
type UserContext struct {
	UserID      uuid.UUID
	DisplayName string
	Email       string
	Roles       []domain.UserRoleType
	CompanyID   domain.CompanyID
}

type contextKey string

const userContextKey contextKey = "userContext"

// WithUserContext adds user context to the context
func WithUserContext(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// FromContext extracts user context from the context
func FromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(userContextKey).(*UserContext)
	return user, ok
}

// MustFromContext extracts user context or panics
func MustFromContext(ctx context.Context) *UserContext {
	user, ok := FromContext(ctx)
	if !ok {
		panic("user context not found in context")
	}
	return user
}

// HasRole checks if user has a specific role
func (u *UserContext) HasRole(role domain.UserRoleType) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (u *UserContext) HasAnyRole(roles ...domain.UserRoleType) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// IsSuperAdmin checks if user is a super admin (has access to all companies)
func (u *UserContext) IsSuperAdmin() bool {
	return u.HasRole(domain.RoleSuperAdmin)
}

// IsCompanyAdmin checks if user is an admin for their company
func (u *UserContext) IsCompanyAdmin() bool {
	return u.HasAnyRole(domain.RoleSuperAdmin, domain.RoleCompanyAdmin)
}

// IsGruppenUser checks if user belongs to Straye Gruppen (parent company)
// Gruppen users can typically access data across all subsidiary companies
func (u *UserContext) IsGruppenUser() bool {
	return u.CompanyID == domain.CompanyGruppen || u.IsSuperAdmin()
}

// CanAccessCompany checks if user can access data for a specific company
func (u *UserContext) CanAccessCompany(companyID domain.CompanyID) bool {
	// Super admins and Gruppen users can access all companies
	if u.IsSuperAdmin() || u.IsGruppenUser() {
		return true
	}
	// Otherwise, user can only access their own company
	return u.CompanyID == companyID
}

// GetCompanyFilter returns the company ID to filter queries by
// Returns nil for super admins and Gruppen users (no filtering needed)
func (u *UserContext) GetCompanyFilter() *domain.CompanyID {
	if u.IsSuperAdmin() || u.IsGruppenUser() {
		return nil
	}
	return &u.CompanyID
}

// HasPermission checks if user has a specific permission based on their roles
// This is a convenience method - full permission checking is in the permission service
func (u *UserContext) HasPermission(permission domain.PermissionType) bool {
	// Super admins have all permissions
	if u.IsSuperAdmin() {
		return true
	}

	// Check each role's default permissions
	for _, role := range u.Roles {
		if hasRolePermission(role, permission) {
			return true
		}
	}
	return false
}

// hasRolePermission checks if a role has a specific permission by default
func hasRolePermission(role domain.UserRoleType, permission domain.PermissionType) bool {
	// Define default permissions per role
	rolePermissions := map[domain.UserRoleType][]domain.PermissionType{
		domain.RoleSuperAdmin: {
			// Super admin has all permissions - handled above
		},
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

	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range perms {
		if p == permission {
			return true
		}
	}
	return false
}

// GetDisplayNameInitials returns initials from the display name (e.g., "John Doe" -> "JD")
func (u *UserContext) GetDisplayNameInitials() string {
	if u.DisplayName == "" {
		return ""
	}
	parts := strings.Fields(u.DisplayName)
	initials := ""
	for _, part := range parts {
		if len(part) > 0 {
			initials += strings.ToUpper(string(part[0]))
		}
	}
	return initials
}

// RolesAsStrings returns roles as a slice of strings
func (u *UserContext) RolesAsStrings() []string {
	result := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		result[i] = string(role)
	}
	return result
}
