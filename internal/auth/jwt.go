package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidScope = errors.New("token missing required scope")
)

// JWTValidator validates JWT tokens from Azure AD
type JWTValidator struct {
	config     *config.AzureAdConfig
	publicKeys map[string]*rsa.PublicKey
	lastUpdate time.Time
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(cfg *config.AzureAdConfig) *JWTValidator {
	return &JWTValidator{
		config:     cfg,
		publicKeys: make(map[string]*rsa.PublicKey),
	}
}

// ValidateToken validates a JWT token and returns user context
func (v *JWTValidator) ValidateToken(tokenString string) (*UserContext, error) {
	// Parse token without validation first to get header
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Get key ID from header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: missing kid in header", ErrInvalidToken)
	}

	// Get or fetch public key
	publicKey, err := v.getPublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate token
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	// Validate audience
	if v.config.ClientId != "" {
		aud, _ := claims.GetAudience()
		validAud := false
		for _, a := range aud {
			if a == v.config.ClientId || strings.Contains(a, v.config.ClientId) {
				validAud = true
				break
			}
		}
		if !validAud {
			return nil, fmt.Errorf("%w: invalid audience", ErrInvalidToken)
		}
	}

	// Validate issuer
	iss, _ := claims.GetIssuer()
	if !strings.Contains(iss, v.config.TenantId) {
		return nil, fmt.Errorf("%w: invalid issuer", ErrInvalidToken)
	}

	// Validate scopes
	if v.config.RequiredScopes != "" {
		scopes := ExtractScopes(claims)
		if !HasRequiredScope(scopes, v.config.RequiredScopes) {
			return nil, ErrInvalidScope
		}
	}

	// Extract user information
	userCtx := &UserContext{
		DisplayName: extractString(claims, "name", "unique_name", "preferred_username"),
		Email:       extractString(claims, "email", "upn", "unique_name"),
		Roles:       ExtractRoles(claims),
	}

	// Extract user ID
	if oidStr := extractString(claims, "oid", "sub"); oidStr != "" {
		if uid, err := uuid.Parse(oidStr); err == nil {
			userCtx.UserID = uid
		}
	}

	// If no user ID, generate one from email
	if userCtx.UserID == uuid.Nil && userCtx.Email != "" {
		userCtx.UserID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(userCtx.Email))
	}

	return userCtx, nil
}

func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache
	if key, exists := v.publicKeys[kid]; exists && time.Since(v.lastUpdate) < 24*time.Hour {
		return key, nil
	}

	// Fetch keys from JWKS endpoint
	if err := v.refreshPublicKeys(); err != nil {
		return nil, err
	}

	key, exists := v.publicKeys[kid]
	if !exists {
		return nil, fmt.Errorf("public key not found for kid: %s", kid)
	}

	return key, nil
}

func (v *JWTValidator) refreshPublicKeys() error {
	jwksURL := fmt.Sprintf("%s%s/discovery/v2.0/keys", v.config.InstanceUrl, v.config.TenantId)

	resp, err := http.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
			Kty string `json:"kty"`
			Use string `json:"use"`
			Alg string `json:"alg"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}

		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}

		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}

		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}

		publicKey := &rsa.PublicKey{
			N: n,
			E: e,
		}

		newKeys[key.Kid] = publicKey
	}

	v.publicKeys = newKeys
	v.lastUpdate = time.Now()

	return nil
}

func extractString(claims jwt.MapClaims, keys ...string) string {
	for _, key := range keys {
		if val, ok := claims[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}
	return ""
}

// ExtractRoles extracts roles from JWT claims and returns them as UserRoleType
func ExtractRoles(claims jwt.MapClaims) []domain.UserRoleType {
	roles := []domain.UserRoleType{}

	// Try different claim names
	for _, key := range []string{"roles", "role"} {
		if val, ok := claims[key]; ok {
			switch v := val.(type) {
			case []interface{}:
				for _, r := range v {
					if str, ok := r.(string); ok {
						roles = append(roles, domain.UserRoleType(str))
					}
				}
			case []string:
				for _, str := range v {
					roles = append(roles, domain.UserRoleType(str))
				}
			case string:
				roles = append(roles, domain.UserRoleType(v))
			}
		}
	}

	return roles
}

// ExtractRolesAsStrings extracts roles from JWT claims as strings (for backward compatibility)
func ExtractRolesAsStrings(claims jwt.MapClaims) []string {
	roles := []string{}

	// Try different claim names
	for _, key := range []string{"roles", "role"} {
		if val, ok := claims[key]; ok {
			switch v := val.(type) {
			case []interface{}:
				for _, r := range v {
					if str, ok := r.(string); ok {
						roles = append(roles, str)
					}
				}
			case []string:
				roles = append(roles, v...)
			case string:
				roles = append(roles, v)
			}
		}
	}

	return roles
}

// ExtractScopes extracts scopes from JWT claims
func ExtractScopes(claims jwt.MapClaims) []string {
	scopes := []string{}

	if val, ok := claims["scp"]; ok {
		if str, ok := val.(string); ok {
			scopes = strings.Split(str, " ")
		}
	}

	if val, ok := claims["scope"]; ok {
		if str, ok := val.(string); ok {
			scopes = append(scopes, strings.Split(str, " ")...)
		}
	}

	return scopes
}

// HasRequiredScope checks if token has required scopes
func HasRequiredScope(tokenScopes []string, required string) bool {
	required = strings.TrimSpace(required)
	if required == "" {
		return true
	}

	requiredScopes := strings.Split(required, ",")
	for _, req := range requiredScopes {
		req = strings.TrimSpace(req)
		if req == "" {
			continue
		}
		for _, scope := range tokenScopes {
			if strings.EqualFold(scope, req) {
				return true
			}
		}
	}
	return false
}
