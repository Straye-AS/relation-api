package auth_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestUserContext_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		roles    []domain.UserRoleType
		role     domain.UserRoleType
		expected bool
	}{
		{
			name:     "has role",
			roles:    []domain.UserRoleType{domain.RoleManager, domain.RoleSales},
			role:     domain.RoleManager,
			expected: true,
		},
		{
			name:     "does not have role",
			roles:    []domain.UserRoleType{domain.RoleSales},
			role:     domain.RoleManager,
			expected: false,
		},
		{
			name:     "empty roles",
			roles:    []domain.UserRoleType{},
			role:     domain.RoleManager,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{Roles: tt.roles}
			assert.Equal(t, tt.expected, userCtx.HasRole(tt.role))
		})
	}
}

func TestUserContext_HasAnyRole(t *testing.T) {
	tests := []struct {
		name     string
		roles    []domain.UserRoleType
		check    []domain.UserRoleType
		expected bool
	}{
		{
			name:     "has one of the roles",
			roles:    []domain.UserRoleType{domain.RoleSales},
			check:    []domain.UserRoleType{domain.RoleManager, domain.RoleSales},
			expected: true,
		},
		{
			name:     "has none of the roles",
			roles:    []domain.UserRoleType{domain.RoleViewer},
			check:    []domain.UserRoleType{domain.RoleManager, domain.RoleSales},
			expected: false,
		},
		{
			name:     "empty check list",
			roles:    []domain.UserRoleType{domain.RoleSales},
			check:    []domain.UserRoleType{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{Roles: tt.roles}
			assert.Equal(t, tt.expected, userCtx.HasAnyRole(tt.check...))
		})
	}
}

func TestUserContext_IsSuperAdmin(t *testing.T) {
	tests := []struct {
		name     string
		roles    []domain.UserRoleType
		expected bool
	}{
		{
			name:     "is super admin",
			roles:    []domain.UserRoleType{domain.RoleSuperAdmin},
			expected: true,
		},
		{
			name:     "is not super admin",
			roles:    []domain.UserRoleType{domain.RoleCompanyAdmin},
			expected: false,
		},
		{
			name:     "has multiple roles including super admin",
			roles:    []domain.UserRoleType{domain.RoleManager, domain.RoleSuperAdmin},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{Roles: tt.roles}
			assert.Equal(t, tt.expected, userCtx.IsSuperAdmin())
		})
	}
}

func TestUserContext_IsCompanyAdmin(t *testing.T) {
	tests := []struct {
		name     string
		roles    []domain.UserRoleType
		expected bool
	}{
		{
			name:     "is company admin",
			roles:    []domain.UserRoleType{domain.RoleCompanyAdmin},
			expected: true,
		},
		{
			name:     "is super admin (also counts as admin)",
			roles:    []domain.UserRoleType{domain.RoleSuperAdmin},
			expected: true,
		},
		{
			name:     "is not admin",
			roles:    []domain.UserRoleType{domain.RoleManager},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{Roles: tt.roles}
			assert.Equal(t, tt.expected, userCtx.IsCompanyAdmin())
		})
	}
}

func TestUserContext_IsGruppenUser(t *testing.T) {
	tests := []struct {
		name      string
		companyID domain.CompanyID
		roles     []domain.UserRoleType
		expected  bool
	}{
		{
			name:      "belongs to gruppen",
			companyID: domain.CompanyGruppen,
			roles:     []domain.UserRoleType{domain.RoleSales},
			expected:  true,
		},
		{
			name:      "super admin from other company",
			companyID: domain.CompanyStalbygg,
			roles:     []domain.UserRoleType{domain.RoleSuperAdmin},
			expected:  true,
		},
		{
			name:      "regular user from other company",
			companyID: domain.CompanyStalbygg,
			roles:     []domain.UserRoleType{domain.RoleSales},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{
				CompanyID: tt.companyID,
				Roles:     tt.roles,
			}
			assert.Equal(t, tt.expected, userCtx.IsGruppenUser())
		})
	}
}

func TestUserContext_CanAccessCompany(t *testing.T) {
	tests := []struct {
		name          string
		userCompanyID domain.CompanyID
		roles         []domain.UserRoleType
		targetCompany domain.CompanyID
		expected      bool
	}{
		{
			name:          "super admin can access any company",
			userCompanyID: domain.CompanyStalbygg,
			roles:         []domain.UserRoleType{domain.RoleSuperAdmin},
			targetCompany: domain.CompanyHybridbygg,
			expected:      true,
		},
		{
			name:          "gruppen user can access any company",
			userCompanyID: domain.CompanyGruppen,
			roles:         []domain.UserRoleType{domain.RoleSales},
			targetCompany: domain.CompanyIndustri,
			expected:      true,
		},
		{
			name:          "user can access own company",
			userCompanyID: domain.CompanyStalbygg,
			roles:         []domain.UserRoleType{domain.RoleSales},
			targetCompany: domain.CompanyStalbygg,
			expected:      true,
		},
		{
			name:          "user cannot access other company",
			userCompanyID: domain.CompanyStalbygg,
			roles:         []domain.UserRoleType{domain.RoleSales},
			targetCompany: domain.CompanyHybridbygg,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{
				CompanyID: tt.userCompanyID,
				Roles:     tt.roles,
			}
			assert.Equal(t, tt.expected, userCtx.CanAccessCompany(tt.targetCompany))
		})
	}
}

func TestUserContext_GetCompanyFilter(t *testing.T) {
	tests := []struct {
		name          string
		companyID     domain.CompanyID
		roles         []domain.UserRoleType
		expectNil     bool
		expectedValue domain.CompanyID
	}{
		{
			name:      "super admin - no filter",
			companyID: domain.CompanyStalbygg,
			roles:     []domain.UserRoleType{domain.RoleSuperAdmin},
			expectNil: true,
		},
		{
			name:      "gruppen user - no filter",
			companyID: domain.CompanyGruppen,
			roles:     []domain.UserRoleType{domain.RoleSales},
			expectNil: true,
		},
		{
			name:          "regular user - filtered by company",
			companyID:     domain.CompanyStalbygg,
			roles:         []domain.UserRoleType{domain.RoleSales},
			expectNil:     false,
			expectedValue: domain.CompanyStalbygg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{
				CompanyID: tt.companyID,
				Roles:     tt.roles,
			}
			filter := userCtx.GetCompanyFilter()
			if tt.expectNil {
				assert.Nil(t, filter)
			} else {
				assert.NotNil(t, filter)
				assert.Equal(t, tt.expectedValue, *filter)
			}
		})
	}
}

func TestUserContext_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		roles      []domain.UserRoleType
		permission domain.PermissionType
		expected   bool
	}{
		{
			name:       "super admin has all permissions",
			roles:      []domain.UserRoleType{domain.RoleSuperAdmin},
			permission: domain.PermissionSystemAdmin,
			expected:   true,
		},
		{
			name:       "sales can read customers",
			roles:      []domain.UserRoleType{domain.RoleSales},
			permission: domain.PermissionCustomersRead,
			expected:   true,
		},
		{
			name:       "sales cannot delete customers",
			roles:      []domain.UserRoleType{domain.RoleSales},
			permission: domain.PermissionCustomersDelete,
			expected:   false,
		},
		{
			name:       "viewer can only read",
			roles:      []domain.UserRoleType{domain.RoleViewer},
			permission: domain.PermissionCustomersWrite,
			expected:   false,
		},
		{
			name:       "company admin can manage users",
			roles:      []domain.UserRoleType{domain.RoleCompanyAdmin},
			permission: domain.PermissionUsersManageRoles,
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{Roles: tt.roles}
			assert.Equal(t, tt.expected, userCtx.HasPermission(tt.permission))
		})
	}
}

func TestUserContext_GetDisplayNameInitials(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		expected    string
	}{
		{
			name:        "two names",
			displayName: "John Doe",
			expected:    "JD",
		},
		{
			name:        "three names",
			displayName: "John Middle Doe",
			expected:    "JMD",
		},
		{
			name:        "single name",
			displayName: "John",
			expected:    "J",
		},
		{
			name:        "empty name",
			displayName: "",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userCtx := &auth.UserContext{DisplayName: tt.displayName}
			assert.Equal(t, tt.expected, userCtx.GetDisplayNameInitials())
		})
	}
}

func TestUserContext_RolesAsStrings(t *testing.T) {
	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleManager, domain.RoleSales},
	}

	result := userCtx.RolesAsStrings()

	assert.Equal(t, []string{"manager", "sales"}, result)
}

func TestWithUserContext_and_FromContext(t *testing.T) {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSales},
		CompanyID:   domain.CompanyStalbygg,
	}

	ctx := auth.WithUserContext(context.Background(), userCtx)

	retrieved, ok := auth.FromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, userCtx.UserID, retrieved.UserID)
	assert.Equal(t, userCtx.DisplayName, retrieved.DisplayName)
	assert.Equal(t, userCtx.Email, retrieved.Email)
	assert.Equal(t, userCtx.CompanyID, retrieved.CompanyID)
}

func TestFromContext_Missing(t *testing.T) {
	ctx := context.Background()

	_, ok := auth.FromContext(ctx)
	assert.False(t, ok)
}

func TestMustFromContext_Panics(t *testing.T) {
	ctx := context.Background()

	assert.Panics(t, func() {
		auth.MustFromContext(ctx)
	})
}

func TestMustFromContext_Success(t *testing.T) {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
	}

	ctx := auth.WithUserContext(context.Background(), userCtx)

	assert.NotPanics(t, func() {
		retrieved := auth.MustFromContext(ctx)
		assert.Equal(t, userCtx.UserID, retrieved.UserID)
	})
}
