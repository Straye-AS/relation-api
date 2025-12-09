package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockUserRepository implements a mock user repository for testing
type MockUserRepository struct{}

func (m *MockUserRepository) Upsert(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, nil
}

// MockPermissionService implements a mock permission service for testing
type MockPermissionService struct {
	permissions []domain.PermissionType
}

func (m *MockPermissionService) GetEffectivePermissions(ctx context.Context, userCtx *auth.UserContext) ([]domain.PermissionType, error) {
	if m.permissions != nil {
		return m.permissions, nil
	}
	// Return default permissions for market role
	return []domain.PermissionType{
		domain.PermissionCustomersRead,
		domain.PermissionCustomersWrite,
		domain.PermissionContactsRead,
		domain.PermissionContactsWrite,
		domain.PermissionDealsRead,
		domain.PermissionDealsWrite,
		domain.PermissionOffersRead,
		domain.PermissionOffersWrite,
		domain.PermissionProjectsRead,
		domain.PermissionActivitiesRead,
		domain.PermissionActivitiesWrite,
		domain.PermissionReportsView,
	}, nil
}

// MockGraphClient implements a mock Graph client for testing
type MockGraphClient struct{}

func (m *MockGraphClient) GetUserProfile(ctx context.Context, accessToken string) (*auth.GraphUserProfile, error) {
	return nil, nil // Return nil to simulate Graph API not configured
}

// createTestAuthHandlerWithMocks creates an auth handler with mock dependencies
func createTestAuthHandlerWithMocks(permissions []domain.PermissionType) *handler.AuthHandler {
	logger := zap.NewNop()
	mockUserRepo := &MockUserRepository{}
	mockPermService := &MockPermissionService{permissions: permissions}
	mockGraphClient := &MockGraphClient{}

	return handler.NewAuthHandlerWithMocks(mockUserRepo, mockPermService, mockGraphClient, logger)
}

func TestAuthHandler_Me_Success(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "John Doe",
		Email:       "john@example.com",
		Roles:       []domain.UserRoleType{domain.RoleMarket},
		CompanyID:   domain.CompanyStalbygg,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.Me(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.AuthUserDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, userCtx.UserID.String(), response.ID)
	assert.Equal(t, "John Doe", response.Name)
	assert.Equal(t, "john@example.com", response.Email)
	assert.Equal(t, []string{"market"}, response.Roles)
	assert.Equal(t, "JD", response.Initials)
	assert.False(t, response.IsSuperAdmin)
	assert.False(t, response.IsCompanyAdmin)
	assert.NotNil(t, response.Company)
	assert.Equal(t, "stalbygg", response.Company.ID)
	assert.Equal(t, "Stålbygg", response.Company.Name)
}

func TestAuthHandler_Me_Unauthorized(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()

	h.Me(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Me_SuperAdmin(t *testing.T) {
	// SuperAdmin gets all permissions
	allPerms := []domain.PermissionType{
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
	h := createTestAuthHandlerWithMocks(allPerms)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Super Admin",
		Email:       "admin@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
		CompanyID:   domain.CompanyGruppen,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.Me(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.AuthUserDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.IsSuperAdmin)
	assert.True(t, response.IsCompanyAdmin) // SuperAdmin also counts as CompanyAdmin
}

func TestAuthHandler_Me_CompanyAdmin(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Company Admin",
		Email:       "companyadmin@example.com",
		Roles:       []domain.UserRoleType{domain.RoleCompanyAdmin},
		CompanyID:   domain.CompanyStalbygg,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.Me(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.AuthUserDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.IsSuperAdmin)
	assert.True(t, response.IsCompanyAdmin)
}

func TestAuthHandler_Permissions_Success(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Market User",
		Email:       "market@example.com",
		Roles:       []domain.UserRoleType{domain.RoleMarket},
		CompanyID:   domain.CompanyStalbygg,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/permissions", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.Permissions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.PermissionsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, []string{"market"}, response.Roles)
	assert.False(t, response.IsSuperAdmin)
	assert.NotEmpty(t, response.Permissions)

	// Check that permissions are properly formatted
	hasCustomersRead := false
	for _, perm := range response.Permissions {
		if perm.Resource == "customers" && perm.Action == "read" && perm.Allowed {
			hasCustomersRead = true
			break
		}
	}
	assert.True(t, hasCustomersRead, "Market role should have customers:read permission")
}

func TestAuthHandler_Permissions_Unauthorized(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	req := httptest.NewRequest(http.MethodGet, "/auth/permissions", nil)
	w := httptest.NewRecorder()

	h.Permissions(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandler_Permissions_SuperAdmin(t *testing.T) {
	// SuperAdmin gets all permissions
	allPerms := []domain.PermissionType{
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
	h := createTestAuthHandlerWithMocks(allPerms)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Super Admin",
		Email:       "admin@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
		CompanyID:   domain.CompanyGruppen,
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/permissions", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.Permissions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.PermissionsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.IsSuperAdmin)
	// Super admin should have all permissions
	assert.True(t, len(response.Permissions) > 20, "Super admin should have many permissions")
}

func TestAuthHandler_ListUsers(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleViewer},
		CompanyID:   domain.CompanyStalbygg,
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.ListUsers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.UserDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, userCtx.UserID.String(), response[0].ID)
	assert.Equal(t, "Test User", response[0].Name)
}

func TestAuthHandler_Me_NoCompany(t *testing.T) {
	h := createTestAuthHandlerWithMocks(nil)

	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "User Without Company",
		Email:       "nocompany@example.com",
		Roles:       []domain.UserRoleType{domain.RoleViewer},
		CompanyID:   "", // No company
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
	w := httptest.NewRecorder()

	h.Me(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.AuthUserDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Nil(t, response.Company)
}

func TestAuthHandler_Me_AllCompanyDisplayNames(t *testing.T) {
	tests := []struct {
		companyID    domain.CompanyID
		expectedName string
	}{
		{domain.CompanyGruppen, "Straye Gruppen"},
		{domain.CompanyStalbygg, "Stålbygg"},
		{domain.CompanyHybridbygg, "Hybridbygg"},
		{domain.CompanyIndustri, "Industri"},
		{domain.CompanyTak, "Tak"},
		{domain.CompanyMontasje, "Montasje"},
	}

	for _, tt := range tests {
		t.Run(string(tt.companyID), func(t *testing.T) {
			h := createTestAuthHandlerWithMocks(nil)

			userCtx := &auth.UserContext{
				UserID:      uuid.New(),
				DisplayName: "Test User",
				Email:       "test@example.com",
				Roles:       []domain.UserRoleType{domain.RoleMarket},
				CompanyID:   tt.companyID,
			}

			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
			w := httptest.NewRecorder()

			h.Me(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response domain.AuthUserDTO
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.NotNil(t, response.Company)
			assert.Equal(t, string(tt.companyID), response.Company.ID)
			assert.Equal(t, tt.expectedName, response.Company.Name)
		})
	}
}
