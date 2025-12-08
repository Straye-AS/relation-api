package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
)

// Middleware handles authentication for HTTP requests
type Middleware struct {
	jwtValidator *JWTValidator
	apiKey       string
	logger       *zap.Logger
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(cfg *config.Config, logger *zap.Logger) *Middleware {
	return &Middleware{
		jwtValidator: NewJWTValidator(&cfg.AzureAd),
		apiKey:       cfg.ApiKey.Value,
		logger:       logger,
	}
}

// Authenticate is the main authentication middleware
func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		authHeader := r.Header.Get("Authorization")

		// Try API key first
		if apiKey := r.Header.Get("x-api-key"); apiKey != "" {
			if m.validateAPIKey(apiKey) {
				// Determine company ID from header, default to gruppen
				companyID := domain.CompanyGruppen
				if companyHeader := r.Header.Get("X-Company-ID"); companyHeader != "" {
					if domain.IsValidCompanyID(companyHeader) {
						companyID = domain.CompanyID(companyHeader)
					}
				}

				// Create system user context with API service role
				userCtx := &UserContext{
					UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					DisplayName: "System",
					Email:       "system@straye.io",
					Roles:       []domain.UserRoleType{domain.RoleSuperAdmin, domain.RoleAPIService},
					CompanyID:   companyID,
				}
				ctx := WithUserContext(r.Context(), userCtx)

				// Log successful API key authentication
				m.logger.Info("request authenticated",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.String("auth_type", "api_key"),
					zap.String("user_id", userCtx.UserID.String()),
					zap.String("user_email", userCtx.Email),
					zap.Duration("auth_duration", time.Since(start)),
				)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			m.logger.Warn("invalid API key attempt",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
			)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Try JWT Bearer token
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "Unauthorized: invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userCtx, err := m.jwtValidator.ValidateToken(token)
		if err != nil {
			m.logger.Warn("token validation failed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Error(err),
			)
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Store the access token for OBO flow (to call Graph API)
		userCtx.AccessToken = token

		// Log successful JWT authentication
		m.logger.Info("request authenticated",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("auth_type", "jwt"),
			zap.String("user_id", userCtx.UserID.String()),
			zap.String("user_email", userCtx.Email),
			zap.Strings("roles", userCtx.RolesAsStrings()),
			zap.Duration("auth_duration", time.Since(start)),
		)

		ctx := WithUserContext(r.Context(), userCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate is middleware that attempts authentication but allows unauthenticated requests
// Use this for public endpoints that may have enhanced functionality for authenticated users
func (m *Middleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try API key first
		if apiKey := r.Header.Get("x-api-key"); apiKey != "" {
			if m.validateAPIKey(apiKey) {
				// Determine company ID from header, default to gruppen
				companyID := domain.CompanyGruppen
				if companyHeader := r.Header.Get("X-Company-ID"); companyHeader != "" {
					if domain.IsValidCompanyID(companyHeader) {
						companyID = domain.CompanyID(companyHeader)
					}
				}

				userCtx := &UserContext{
					UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000000"),
					DisplayName: "System",
					Email:       "system@straye.io",
					Roles:       []domain.UserRoleType{domain.RoleSuperAdmin, domain.RoleAPIService},
					CompanyID:   companyID,
				}
				ctx := WithUserContext(r.Context(), userCtx)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			// Invalid API key - continue without auth rather than failing
			m.logger.Debug("optional auth: invalid API key, continuing unauthenticated",
				zap.String("path", r.URL.Path),
			)
		}

		// Try JWT Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				token := parts[1]
				userCtx, err := m.jwtValidator.ValidateToken(token)
				if err == nil {
					ctx := WithUserContext(r.Context(), userCtx)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				// Invalid token - continue without auth rather than failing
				m.logger.Debug("optional auth: token validation failed, continuing unauthenticated",
					zap.String("path", r.URL.Path),
					zap.Error(err),
				)
			}
		}

		// No auth or invalid auth - continue without user context
		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware ensures user has specific role
func (m *Middleware) RequireRole(roles ...domain.UserRoleType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := FromContext(r.Context())
			if !ok {
				http.Error(w, "Forbidden: no user context", http.StatusForbidden)
				return
			}

			if !userCtx.HasAnyRole(roles...) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware ensures user has admin role or valid API key
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := FromContext(r.Context())
		if !ok {
			http.Error(w, "Forbidden: no user context", http.StatusForbidden)
			return
		}

		if !userCtx.IsCompanyAdmin() {
			http.Error(w, "Forbidden: admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequirePermission middleware ensures user has specific permission
func (m *Middleware) RequirePermission(permission domain.PermissionType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := FromContext(r.Context())
			if !ok {
				http.Error(w, "Forbidden: no user context", http.StatusForbidden)
				return
			}

			if !userCtx.HasPermission(permission) {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) validateAPIKey(key string) bool {
	if m.apiKey == "" {
		return false
	}
	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(key), []byte(m.apiKey)) == 1
}
