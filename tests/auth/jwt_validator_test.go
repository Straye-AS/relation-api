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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testKeyPair holds RSA keys for testing
type testKeyPair struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	kid        string
}

// generateTestKeyPair generates an RSA key pair for testing
func generateTestKeyPair(t *testing.T) *testKeyPair {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return &testKeyPair{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		kid:        "test-key-id-123",
	}
}

// createMockJWKSServer creates a mock JWKS endpoint server
func createMockJWKSServer(t *testing.T, keyPair *testKeyPair) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert public key to JWKS format
		n := base64.RawURLEncoding.EncodeToString(keyPair.publicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(keyPair.publicKey.E)).Bytes())

		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "RSA",
					"use": "sig",
					"kid": keyPair.kid,
					"n":   n,
					"e":   e,
					"alg": "RS256",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	}))
}

// createTestToken creates a signed JWT token for testing
func createTestToken(t *testing.T, keyPair *testKeyPair, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyPair.kid

	tokenString, err := token.SignedString(keyPair.privateKey)
	require.NoError(t, err)

	return tokenString
}

func TestJWTValidator_ValidateToken_ValidToken(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":   "test-client-id",
		"iss":   "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"nbf":   time.Now().Add(-time.Minute).Unix(),
		"iat":   time.Now().Unix(),
		"oid":   "12345678-1234-1234-1234-123456789012",
		"name":  "Test User",
		"email": "test@example.com",
		"roles": []interface{}{"admin", "user"},
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	require.NoError(t, err)
	assert.NotNil(t, userCtx)
	assert.Equal(t, "Test User", userCtx.DisplayName)
	assert.Equal(t, "test@example.com", userCtx.Email)
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", userCtx.UserID.String())
	assert.Len(t, userCtx.Roles, 2)
}

func TestJWTValidator_ValidateToken_ExpiredToken(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":  "test-client-id",
		"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":  time.Now().Add(-time.Hour).Unix(), // Expired
		"nbf":  time.Now().Add(-2 * time.Hour).Unix(),
		"iat":  time.Now().Add(-2 * time.Hour).Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, userCtx)
	assert.ErrorIs(t, err, auth.ErrExpiredToken)
}

func TestJWTValidator_ValidateToken_InvalidAudience(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":  "wrong-client-id", // Wrong audience
		"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"nbf":  time.Now().Add(-time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, userCtx)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestJWTValidator_ValidateToken_InvalidIssuer(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":  "test-client-id",
		"iss":  "https://login.microsoftonline.com/wrong-tenant-id/v2.0", // Wrong issuer
		"exp":  time.Now().Add(time.Hour).Unix(),
		"nbf":  time.Now().Add(-time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, userCtx)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestJWTValidator_ValidateToken_MissingRequiredScope(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "api.access", // Required scope
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":  "test-client-id",
		"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"nbf":  time.Now().Add(-time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
		"scp":  "other.scope", // Missing required scope
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, userCtx)
	assert.ErrorIs(t, err, auth.ErrInvalidScope)
}

func TestJWTValidator_ValidateToken_WithRequiredScope(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "api.access",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":  "test-client-id",
		"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"nbf":  time.Now().Add(-time.Minute).Unix(),
		"iat":  time.Now().Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
		"scp":  "api.access api.write", // Has required scope
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	require.NoError(t, err)
	assert.NotNil(t, userCtx)
	assert.Equal(t, "Test User", userCtx.DisplayName)
}

func TestJWTValidator_ValidateToken_InvalidTokenFormat(t *testing.T) {
	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    "https://login.test.com/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	userCtx, err := validator.ValidateToken("not-a-valid-jwt-token")

	assert.Error(t, err)
	assert.Nil(t, userCtx)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestJWTValidator_ValidateToken_MissingKid(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":  "test-client-id",
		"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
	}

	// Create token without kid in header
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	// Don't set kid
	tokenString, err := token.SignedString(keyPair.privateKey)
	require.NoError(t, err)

	userCtx, err := validator.ValidateToken(tokenString)

	assert.Error(t, err)
	assert.Nil(t, userCtx)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestJWTValidator_ValidateToken_ExtractsEmailFromAlternativeClaims(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	tests := []struct {
		name          string
		claims        jwt.MapClaims
		expectedEmail string
	}{
		{
			name: "email from upn claim",
			claims: jwt.MapClaims{
				"aud":  "test-client-id",
				"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
				"exp":  time.Now().Add(time.Hour).Unix(),
				"oid":  "12345678-1234-1234-1234-123456789012",
				"name": "Test User",
				"upn":  "user@company.com",
			},
			expectedEmail: "user@company.com",
		},
		{
			name: "email from unique_name claim",
			claims: jwt.MapClaims{
				"aud":         "test-client-id",
				"iss":         "https://login.microsoftonline.com/test-tenant-id/v2.0",
				"exp":         time.Now().Add(time.Hour).Unix(),
				"oid":         "12345678-1234-1234-1234-123456789012",
				"name":        "Test User",
				"unique_name": "unique@company.com",
			},
			expectedEmail: "unique@company.com",
		},
		{
			name: "email claim takes precedence",
			claims: jwt.MapClaims{
				"aud":         "test-client-id",
				"iss":         "https://login.microsoftonline.com/test-tenant-id/v2.0",
				"exp":         time.Now().Add(time.Hour).Unix(),
				"oid":         "12345678-1234-1234-1234-123456789012",
				"name":        "Test User",
				"email":       "primary@company.com",
				"upn":         "secondary@company.com",
				"unique_name": "tertiary@company.com",
			},
			expectedEmail: "primary@company.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := auth.NewJWTValidator(cfg)
			tokenString := createTestToken(t, keyPair, tt.claims)

			userCtx, err := validator.ValidateToken(tokenString)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedEmail, userCtx.Email)
		})
	}
}

func TestJWTValidator_ValidateToken_GeneratesUserIDFromEmailWhenMissing(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	claims := jwt.MapClaims{
		"aud":   "test-client-id",
		"iss":   "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"name":  "Test User",
		"email": "test@example.com",
		// No oid or sub claim
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	require.NoError(t, err)
	assert.NotEmpty(t, userCtx.UserID)
	// Should be deterministic based on email
	assert.NotEqual(t, "00000000-0000-0000-0000-000000000000", userCtx.UserID.String())
}

func TestJWTValidator_ValidateToken_ExtractsRolesFromVariousClaims(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	tests := []struct {
		name          string
		claims        jwt.MapClaims
		expectedRoles int
	}{
		{
			name: "roles as array of interfaces",
			claims: jwt.MapClaims{
				"aud":   "test-client-id",
				"iss":   "https://login.microsoftonline.com/test-tenant-id/v2.0",
				"exp":   time.Now().Add(time.Hour).Unix(),
				"oid":   "12345678-1234-1234-1234-123456789012",
				"name":  "Test User",
				"roles": []interface{}{"admin", "user", "viewer"},
			},
			expectedRoles: 3,
		},
		{
			name: "role as single string",
			claims: jwt.MapClaims{
				"aud":  "test-client-id",
				"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
				"exp":  time.Now().Add(time.Hour).Unix(),
				"oid":  "12345678-1234-1234-1234-123456789012",
				"name": "Test User",
				"role": "admin",
			},
			expectedRoles: 1,
		},
		{
			name: "no roles",
			claims: jwt.MapClaims{
				"aud":  "test-client-id",
				"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
				"exp":  time.Now().Add(time.Hour).Unix(),
				"oid":  "12345678-1234-1234-1234-123456789012",
				"name": "Test User",
			},
			expectedRoles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := auth.NewJWTValidator(cfg)
			tokenString := createTestToken(t, keyPair, tt.claims)

			userCtx, err := validator.ValidateToken(tokenString)

			require.NoError(t, err)
			assert.Len(t, userCtx.Roles, tt.expectedRoles)
		})
	}
}

func TestJWTValidator_ValidateToken_AudienceWithAPIPrefix(t *testing.T) {
	keyPair := generateTestKeyPair(t)
	server := createMockJWKSServer(t, keyPair)
	defer server.Close()

	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant-id",
		ClientId:       "test-client-id",
		InstanceUrl:    server.URL + "/",
		RequiredScopes: "",
	}

	validator := auth.NewJWTValidator(cfg)

	// Azure AD sometimes returns audience with api:// prefix
	claims := jwt.MapClaims{
		"aud":  "api://test-client-id",
		"iss":  "https://login.microsoftonline.com/test-tenant-id/v2.0",
		"exp":  time.Now().Add(time.Hour).Unix(),
		"oid":  "12345678-1234-1234-1234-123456789012",
		"name": "Test User",
	}

	tokenString := createTestToken(t, keyPair, claims)

	userCtx, err := validator.ValidateToken(tokenString)

	require.NoError(t, err)
	assert.NotNil(t, userCtx)
}
