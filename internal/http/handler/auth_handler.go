package handler

import (
	"net/http"

	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	logger   *zap.Logger
}

func NewAuthHandler(userRepo *repository.UserRepository, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// @Summary Get current user
// @Tags Auth
// @Produce json
// @Success 200 {object} domain.UserDTO
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustFromContext(r.Context())

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

	dto := domain.UserDTO{
		ID:    userCtx.UserID.String(),
		Name:  userCtx.DisplayName,
		Email: userCtx.Email,
		Roles: userCtx.RolesAsStrings(),
	}

	respondJSON(w, http.StatusOK, dto)
}

// @Summary List users
// @Tags Users
// @Produce json
// @Success 200 {array} domain.UserDTO
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
