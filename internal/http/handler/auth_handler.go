package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// UserRepository interface for dependency injection
type UserRepository interface {
	Upsert(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	ListByCompanyAccess(ctx context.Context, companyFilter *domain.CompanyID) ([]domain.User, error)
}

// PermissionServiceInterface for dependency injection
type PermissionServiceInterface interface {
	GetEffectivePermissions(ctx context.Context, userCtx *auth.UserContext) ([]domain.PermissionType, error)
}

// GraphClientInterface for dependency injection
type GraphClientInterface interface {
	GetUserProfile(ctx context.Context, accessToken string) (*auth.GraphUserProfile, error)
}

type AuthHandler struct {
	userRepo          UserRepository
	permissionService PermissionServiceInterface
	graphClient       GraphClientInterface
	logger            *zap.Logger
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	permissionService *service.PermissionService,
	graphClient *auth.GraphClient,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:          userRepo,
		permissionService: permissionService,
		graphClient:       graphClient,
		logger:            logger,
	}
}

// NewAuthHandlerWithMocks creates an auth handler with mock dependencies for testing
func NewAuthHandlerWithMocks(
	userRepo UserRepository,
	permissionService PermissionServiceInterface,
	graphClient GraphClientInterface,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:          userRepo,
		permissionService: permissionService,
		graphClient:       graphClient,
		logger:            logger,
	}
}

// getCompanyDisplayName returns the display name for a company ID
func getCompanyDisplayName(companyID domain.CompanyID) string {
	names := map[domain.CompanyID]string{
		domain.CompanyGruppen:    "Straye Gruppen",
		domain.CompanyStalbygg:   "Stålbygg",
		domain.CompanyHybridbygg: "Hybridbygg",
		domain.CompanyIndustri:   "Industri",
		domain.CompanyTak:        "Tak",
		domain.CompanyMontasje:   "Montasje",
	}
	if name, ok := names[companyID]; ok {
		return name
	}
	return string(companyID)
}

// Me godoc
// @Summary Get current authenticated user
// @Description Returns the current authenticated user with roles, company, and permissions info
// @Tags Auth
// @Produce json
// @Success 200 {object} domain.AuthUserDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	// Fetch additional profile data from Microsoft Graph API (if configured)
	var graphProfile *auth.GraphUserProfile
	if h.graphClient != nil && userCtx.AccessToken != "" {
		profile, err := h.graphClient.GetUserProfile(r.Context(), userCtx.AccessToken)
		if err != nil {
			h.logger.Warn("failed to fetch user profile from Graph API",
				zap.String("user_id", userCtx.UserID.String()),
				zap.Error(err))
		} else {
			graphProfile = profile
		}
	}

	// Upsert user in database with all available info from token and Graph API
	now := time.Now()
	user := &domain.User{
		ID:            userCtx.UserID.String(),
		DisplayName:   userCtx.DisplayName,
		Email:         userCtx.Email,
		Roles:         userCtx.RolesAsStrings(),
		AzureADRoles:  pq.StringArray(userCtx.AzureADRoles),
		LastIPAddress: userCtx.IPAddress,
		LastLoginAt:   &now,
		Department:    userCtx.Department,
	}

	// Enrich with Graph API data if available
	if graphProfile != nil {
		if graphProfile.Department != "" {
			user.Department = graphProfile.Department
		}
		if graphProfile.GivenName != "" {
			user.FirstName = graphProfile.GivenName
		}
		if graphProfile.Surname != "" {
			user.LastName = graphProfile.Surname
		}
	}

	if err := h.userRepo.Upsert(r.Context(), user); err != nil {
		h.logger.Warn("failed to upsert user", zap.Error(err))
	}

	// Fetch user from database to get persisted roles and company
	dbUser, err := h.userRepo.GetByEmail(r.Context(), userCtx.Email)
	if err != nil {
		h.logger.Warn("failed to fetch user from database", zap.Error(err))
	}

	// Use database roles and company if available, otherwise fallback to JWT
	roles := userCtx.RolesAsStrings()
	companyID := userCtx.CompanyID
	isSuperAdmin := userCtx.IsSuperAdmin()
	isCompanyAdmin := userCtx.IsCompanyAdmin()

	if dbUser != nil {
		// Override with database values if they exist
		if len(dbUser.Roles) > 0 {
			roles = dbUser.Roles
			// Check if super_admin or company_admin is in database roles
			for _, role := range dbUser.Roles {
				if role == string(domain.RoleSuperAdmin) {
					isSuperAdmin = true
				}
				if role == string(domain.RoleCompanyAdmin) {
					isCompanyAdmin = true
				}
			}
		}
		if dbUser.CompanyID != nil && *dbUser.CompanyID != "" {
			companyID = *dbUser.CompanyID
		}
	}

	// Build company info if present
	var company *domain.CompanyDTO
	if companyID != "" {
		company = &domain.CompanyDTO{
			ID:   string(companyID),
			Name: getCompanyDisplayName(companyID),
		}
	}

	dto := domain.AuthUserDTO{
		ID:             userCtx.UserID.String(),
		Name:           userCtx.DisplayName,
		Email:          userCtx.Email,
		Roles:          roles,
		Company:        company,
		Initials:       userCtx.GetDisplayNameInitials(),
		IsSuperAdmin:   isSuperAdmin,
		IsCompanyAdmin: isCompanyAdmin,
	}

	respondJSON(w, http.StatusOK, dto)
}

// Permissions godoc
// @Summary Get current user's permissions
// @Description Returns the full list of permissions for the current authenticated user
// @Tags Auth
// @Produce json
// @Success 200 {object} domain.PermissionsResponseDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /auth/permissions [get]
func (h *AuthHandler) Permissions(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
		return
	}

	// Get effective permissions from service
	permissions, err := h.permissionService.GetEffectivePermissions(r.Context(), userCtx)
	if err != nil {
		h.logger.Error("failed to get effective permissions",
			zap.String("user_id", userCtx.UserID.String()),
			zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get permissions",
		})
		return
	}

	// Convert to DTOs
	permissionDTOs := make([]domain.PermissionDTO, 0, len(permissions))
	for _, perm := range permissions {
		// Split permission into resource:action
		parts := strings.SplitN(string(perm), ":", 2)
		resource := parts[0]
		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}

		permissionDTOs = append(permissionDTOs, domain.PermissionDTO{
			Resource: resource,
			Action:   action,
			Allowed:  true,
		})
	}

	dto := domain.PermissionsResponseDTO{
		Permissions:  permissionDTOs,
		Roles:        userCtx.RolesAsStrings(),
		IsSuperAdmin: userCtx.IsSuperAdmin(),
	}

	respondJSON(w, http.StatusOK, dto)
}

// ListUsers godoc
// @Summary List users
// @Description Returns a list of active users. Super admins and gruppen users see all users. Regular users see only users from their company.
// @Tags Users
// @Produce json
// @Success 200 {array} domain.UserDTO
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /users [get]
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustFromContext(r.Context())

	// Get company filter based on user's access level
	companyFilter := userCtx.GetCompanyFilter()

	users, err := h.userRepo.ListByCompanyAccess(r.Context(), companyFilter)
	if err != nil {
		h.logger.Error("failed to list users", zap.Error(err))
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	// Convert to DTOs
	dtos := make([]domain.UserDTO, 0, len(users))
	for _, user := range users {
		dtos = append(dtos, domain.UserDTO{
			ID:         user.ID,
			Name:       user.DisplayName,
			Email:      user.Email,
			Roles:      user.Roles,
			Department: user.Department,
			Avatar:     user.Avatar,
		})
	}

	respondJSON(w, http.StatusOK, dtos)
}
