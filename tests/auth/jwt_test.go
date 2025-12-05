package auth_test

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestJWTValidator_ExtractScopes(t *testing.T) {
	tests := []struct {
		name     string
		claims   jwt.MapClaims
		expected []string
	}{
		{
			name: "single scope from scp",
			claims: jwt.MapClaims{
				"scp": "api.access",
			},
			expected: []string{"api.access"},
		},
		{
			name: "multiple scopes from scp",
			claims: jwt.MapClaims{
				"scp": "api.access api.write",
			},
			expected: []string{"api.access", "api.write"},
		},
		{
			name: "scopes from scope field",
			claims: jwt.MapClaims{
				"scope": "openid profile email",
			},
			expected: []string{"openid", "profile", "email"},
		},
		{
			name:     "no scopes",
			claims:   jwt.MapClaims{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopes := auth.ExtractScopes(tt.claims)
			assert.ElementsMatch(t, tt.expected, scopes)
		})
	}
}

func TestJWTValidator_ExtractRoles(t *testing.T) {
	tests := []struct {
		name     string
		claims   jwt.MapClaims
		expected []domain.UserRoleType
	}{
		{
			name: "array of roles",
			claims: jwt.MapClaims{
				"roles": []interface{}{"admin", "user"},
			},
			expected: []domain.UserRoleType{"admin", "user"},
		},
		{
			name: "single role as string",
			claims: jwt.MapClaims{
				"role": "admin",
			},
			expected: []domain.UserRoleType{"admin"},
		},
		{
			name:     "no roles",
			claims:   jwt.MapClaims{},
			expected: []domain.UserRoleType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles := auth.ExtractRoles(tt.claims)
			assert.ElementsMatch(t, tt.expected, roles)
		})
	}
}

func TestHasRequiredScope(t *testing.T) {
	tests := []struct {
		name        string
		tokenScopes []string
		required    string
		expected    bool
	}{
		{
			name:        "has required scope",
			tokenScopes: []string{"api.access", "api.write"},
			required:    "api.access",
			expected:    true,
		},
		{
			name:        "missing required scope",
			tokenScopes: []string{"api.write"},
			required:    "api.access",
			expected:    false,
		},
		{
			name:        "case insensitive match",
			tokenScopes: []string{"API.ACCESS"},
			required:    "api.access",
			expected:    true,
		},
		{
			name:        "multiple required scopes - has one",
			tokenScopes: []string{"api.write"},
			required:    "api.access,api.write",
			expected:    true,
		},
		{
			name:        "no required scopes",
			tokenScopes: []string{"api.access"},
			required:    "",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := auth.HasRequiredScope(tt.tokenScopes, tt.required)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewJWTValidator(t *testing.T) {
	cfg := &config.AzureAdConfig{
		TenantId:       "test-tenant",
		ClientId:       "test-client",
		InstanceUrl:    "https://login.test.com/",
		RequiredScopes: "api.access",
	}

	validator := auth.NewJWTValidator(cfg)

	assert.NotNil(t, validator)
}
