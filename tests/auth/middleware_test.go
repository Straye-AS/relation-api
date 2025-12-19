package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func createTestConfig(jwksURL, apiKey string) *config.Config {
	return &config.Config{
		AzureAd: config.AzureAdConfig{
			TenantId:       "test-tenant-id",
			ClientId:       "test-client-id",
			InstanceUrl:    jwksURL + "/",
			RequiredScopes: "",
		},
		ApiKey: config.ApiKeyConfig{
			Value: apiKey,
		},
	}
}

func createTestMiddleware(t *testing.T, jwksURL, apiKey string) *auth.Middleware {
	cfg := createTestConfig(jwksURL, apiKey)
	logger := zap.NewNop()
	return auth.NewMiddleware(cfg, nil, logger) // nil userLookup for tests
}

func TestMiddleware_Authenticate_WithAPIKey(t *testing.T) {
	apiKey := "test-api-key-12345"
	middleware := createTestMiddleware(t, "http://localhost", apiKey)

	handlerCalled := false
	var capturedUserCtx *auth.UserContext

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedUserCtx, _ = auth.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	req.Header.Set("x-api-key", apiKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedUserCtx)
	assert.Equal(t, "System", capturedUserCtx.DisplayName)
	assert.Equal(t, "system@straye.io", capturedUserCtx.Email)
	assert.True(t, capturedUserCtx.HasRole(domain.RoleSuperAdmin))
	assert.True(t, capturedUserCtx.HasRole(domain.RoleAPIService))
}

func TestMiddleware_Authenticate_WithInvalidAPIKey(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "correct-key")

	handlerCalled := false
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	req.Header.Set("x-api-key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_Authenticate_WithJWT(t *testing.T) {
	// Generate test keys
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	kid := "test-key-id"

	// Create mock JWKS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())
		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{"kty": "RSA", "use": "sig", "kid": kid, "n": n, "e": e, "alg": "RS256"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	middleware := createTestMiddleware(t, server.URL, "")

	// Create valid token
	claims := jwt.MapClaims{
		"aud":   "test-client-id",
		"iss":   "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"oid":   "12345678-1234-1234-1234-123456789012",
		"name":  "Test User",
		"email": "test@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	handlerCalled := false
	var capturedUserCtx *auth.UserContext

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedUserCtx, _ = auth.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedUserCtx)
	assert.Equal(t, "Test User", capturedUserCtx.DisplayName)
	assert.Equal(t, "test@example.com", capturedUserCtx.Email)
}

func TestMiddleware_Authenticate_MissingAuth(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMiddleware_Authenticate_InvalidBearerFormat(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	tests := []struct {
		name       string
		authHeader string
	}{
		{"no bearer prefix", "some-token"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"empty bearer", "Bearer "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
			}))

			req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.False(t, handlerCalled)
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

func TestMiddleware_OptionalAuthenticate_WithValidAPIKey(t *testing.T) {
	apiKey := "test-api-key"
	middleware := createTestMiddleware(t, "http://localhost", apiKey)

	handlerCalled := false
	var capturedUserCtx *auth.UserContext

	handler := middleware.OptionalAuthenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedUserCtx, _ = auth.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)
	req.Header.Set("x-api-key", apiKey)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedUserCtx)
	assert.Equal(t, "System", capturedUserCtx.DisplayName)
}

func TestMiddleware_OptionalAuthenticate_WithInvalidAPIKey(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "correct-key")

	handlerCalled := false
	var hasUserCtx bool

	handler := middleware.OptionalAuthenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		_, hasUserCtx = auth.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)
	req.Header.Set("x-api-key", "wrong-key")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Handler should still be called, but without user context
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, hasUserCtx)
}

func TestMiddleware_OptionalAuthenticate_NoAuth(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	var hasUserCtx bool

	handler := middleware.OptionalAuthenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		_, hasUserCtx = auth.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)
	// No auth headers
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, hasUserCtx)
}

func TestMiddleware_RequireRole_HasRole(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequireRole(domain.RoleManager, domain.RoleMarket)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Create context with user who has the required role
	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleMarket},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireRole_MissingRole(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequireRole(domain.RoleSuperAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	// Create context with user who does NOT have the required role
	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleViewer},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireRole_NoUserContext(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequireRole(domain.RoleViewer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	// No user context
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequireAdmin_IsAdmin(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleCompanyAdmin},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequireAdmin_NotAdmin(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleViewer},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_RequirePermission_HasPermission(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequirePermission(domain.PermissionCustomersRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Viewer role has customers:read permission
	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleViewer},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMiddleware_RequirePermission_MissingPermission(t *testing.T) {
	middleware := createTestMiddleware(t, "http://localhost", "test-key")

	handlerCalled := false
	handler := middleware.RequirePermission(domain.PermissionCustomersDelete)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	// Viewer role does NOT have customers:delete permission
	userCtx := &auth.UserContext{
		Roles: []domain.UserRoleType{domain.RoleViewer},
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/customers/123", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestMiddleware_APIKeyPriority(t *testing.T) {
	// Generate test keys for JWT
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	kid := "test-key-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())
		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{"kty": "RSA", "use": "sig", "kid": kid, "n": n, "e": e, "alg": "RS256"},
			},
		}
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	defer server.Close()

	apiKey := "test-api-key"
	middleware := createTestMiddleware(t, server.URL, apiKey)

	// Create valid JWT
	claims := jwt.MapClaims{
		"aud":   "test-client-id",
		"iss":   "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"oid":   "12345678-1234-1234-1234-123456789012",
		"name":  "JWT User",
		"email": "jwt@example.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	var capturedUserCtx *auth.UserContext

	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserCtx, _ = auth.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Send request with BOTH API key and JWT - API key should take priority
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedUserCtx)
	// Should be System user (from API key), not JWT User
	assert.Equal(t, "System", capturedUserCtx.DisplayName)
}
