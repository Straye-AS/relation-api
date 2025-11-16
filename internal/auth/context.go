package auth

import (
	"context"

	"github.com/google/uuid"
)

// UserContext holds authenticated user information
type UserContext struct {
	UserID      uuid.UUID
	DisplayName string
	Email       string
	Roles       []string
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
func (u *UserContext) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func (u *UserContext) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

