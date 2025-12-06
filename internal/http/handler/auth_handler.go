package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// UserRepository interface for dependency injection
type UserRepository interface {
	Upsert(ctx context.Context, user *domain.User) error
}

// PermissionServiceInterface for dependency injection
type PermissionServiceInterface interface {
	GetEffectivePermissions(ctx context.Context, userCtx *auth.UserContext) ([]domain.PermissionType, error)
}

type AuthHandler struct {
	userRepo          UserRepository
	permissionService PermissionServiceInterface
	logger            *zap.Logger
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	permissionService *service.PermissionService,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:          userRepo,
		permissionService: permissionService,
		logger:            logger,
	}
}

// NewAuthHandlerWithMocks creates an auth handler with mock dependencies for testing
func NewAuthHandlerWithMocks(
	userRepo UserRepository,
	permissionService PermissionServiceInterface,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		userRepo:          userRepo,
		permissionService: permissionService,
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

	// Upsert user in database
	user := &domain.User{
		ID:          userCtx.UserID.String(),
		DisplayName: userCtx.DisplayName,
		Email:       userCtx.Email,
		Roles:       userCtx.RolesAsStrings(),
	}

	if err := h.userRepo.Upsert(r.Context(), user); err != nil {
		h.logger.Warn("failed to upsert user", zap.Error(err))
	}

	// Build company info if present
	var company *domain.CompanyDTO
	if userCtx.CompanyID != "" {
		company = &domain.CompanyDTO{
			ID:   string(userCtx.CompanyID),
			Name: getCompanyDisplayName(userCtx.CompanyID),
		}
	}

	dto := domain.AuthUserDTO{
		ID:             userCtx.UserID.String(),
		Name:           userCtx.DisplayName,
		Email:          userCtx.Email,
		Roles:          userCtx.RolesAsStrings(),
		Company:        company,
		Initials:       userCtx.GetDisplayNameInitials(),
		IsSuperAdmin:   userCtx.IsSuperAdmin(),
		IsCompanyAdmin: userCtx.IsCompanyAdmin(),
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
// @Description Returns a list of users (currently only returns the authenticated user)
// @Tags Users
// @Produce json
// @Success 200 {array} domain.UserDTO
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /users [get]
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustFromContext(r.Context())

	dto := []domain.UserDTO{{
		ID:    userCtx.UserID.String(),
		Name:  userCtx.DisplayName,
		Email: userCtx.Email,
		Roles: userCtx.RolesAsStrings(),
	}}

	respondJSON(w, http.StatusOK, dto)
}
