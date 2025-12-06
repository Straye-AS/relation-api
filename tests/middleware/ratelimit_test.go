package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func createTestRateLimiter(cfg *config.RateLimitConfig) *middleware.RateLimiter {
	logger := zap.NewNop()
	return middleware.NewRateLimiter(cfg, logger)
}

func TestRateLimiter_Disabled(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           false,
		RequestsPerMinute: 5,
	}

	rl := createTestRateLimiter(cfg)
	handlerCalled := 0

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests - should all pass since rate limiting is disabled
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 100, handlerCalled)
}

func TestRateLimiter_WhitelistedIP(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 2, // Very low limit
		WhitelistIPs:      []string{"127.0.0.1"},
		WhitelistPaths:    []string{},
	}

	rl := createTestRateLimiter(cfg)
	handlerCalled := 0

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests from whitelisted IP - should all pass
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 50, handlerCalled)
}

func TestRateLimiter_WhitelistedPath(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 2, // Very low limit
		WhitelistIPs:      []string{},
		WhitelistPaths:    []string{"/health"},
	}

	rl := createTestRateLimiter(cfg)
	handlerCalled := 0

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// Make many requests to whitelisted path - should all pass
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 50, handlerCalled)
}

func TestRateLimiter_WhitelistedPathPrefix(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 2, // Very low limit
		WhitelistIPs:      []string{},
		WhitelistPaths:    []string{"/health/*"},
	}

	rl := createTestRateLimiter(cfg)
	handlerCalled := 0

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// Make requests to whitelisted path prefix
	paths := []string{"/health/db", "/health/ready", "/health/live"}
	for i := 0; i < 30; i++ {
		req := httptest.NewRequest(http.MethodGet, paths[i%len(paths)], nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 30, handlerCalled)
}

func TestRateLimiter_LimitExceeded(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 5,
		BurstSize:         5,
		WhitelistIPs:      []string{},
		WhitelistPaths:    []string{},
	}

	rl := createTestRateLimiter(cfg)

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	okCount := 0
	rateLimitedCount := 0

	// Make more requests than the limit
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			okCount++
		} else if w.Code == http.StatusTooManyRequests {
			rateLimitedCount++
			// Check response headers
			assert.NotEmpty(t, w.Header().Get("Retry-After"))
		}
	}

	// Some requests should succeed, some should be rate limited
	assert.Greater(t, okCount, 0, "Some requests should succeed")
	assert.Greater(t, rateLimitedCount, 0, "Some requests should be rate limited")
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 3,
		BurstSize:         3,
		WhitelistIPs:      []string{},
		WhitelistPaths:    []string{},
	}

	rl := createTestRateLimiter(cfg)

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Each IP should have its own independent limit
	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}

	for _, ip := range ips {
		okCount := 0
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = ip
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code == http.StatusOK {
				okCount++
			}
		}
		// Each IP should get at least some successful requests
		assert.Greater(t, okCount, 0, "IP %s should get successful requests", ip)
	}
}

func TestRateLimiter_XForwardedFor(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 3,
		BurstSize:         3,
		WhitelistIPs:      []string{"10.0.0.1"},
		WhitelistPaths:    []string{},
	}

	rl := createTestRateLimiter(cfg)
	handlerCalled := 0

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// Request with whitelisted IP in X-Forwarded-For
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"          // Proxy IP
		req.Header.Set("X-Forwarded-For", "10.0.0.1") // Real client IP (whitelisted)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 20, handlerCalled)
}

func TestRateLimiter_XRealIP(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 3,
		BurstSize:         3,
		WhitelistIPs:      []string{"10.0.0.2"},
		WhitelistPaths:    []string{},
	}

	rl := createTestRateLimiter(cfg)
	handlerCalled := 0

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled++
		w.WriteHeader(http.StatusOK)
	}))

	// Request with whitelisted IP in X-Real-IP
	for i := 0; i < 20; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"    // Proxy IP
		req.Header.Set("X-Real-IP", "10.0.0.2") // Real client IP (whitelisted)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	assert.Equal(t, 20, handlerCalled)
}

func TestRateLimiter_AuthenticatedUserLimit(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:               true,
		RequestsPerMinute:     2,  // Low limit for unauthenticated
		RequestsPerMinuteAuth: 10, // Higher limit for authenticated
		BurstSize:             2,
		WhitelistIPs:          []string{},
		WhitelistPaths:        []string{},
	}

	rl := createTestRateLimiter(cfg)

	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create authenticated context
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleMarket},
	}

	okCount := 0

	// Make authenticated requests - should have higher limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req = req.WithContext(auth.WithUserContext(context.Background(), userCtx))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code == http.StatusOK {
			okCount++
		}
	}

	// Should get more successful requests than the unauthenticated limit
	assert.Greater(t, okCount, 2, "Authenticated users should have higher limit")
}

func TestRateLimiter_RateLimitResponseFormat(t *testing.T) {
	cfg := &config.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 1,
		BurstSize:         1,
		WhitelistIPs:      []string{},
		WhitelistPaths:    []string{},
	}

	rl := createTestRateLimiter(cfg)

	handler := rl.LimitByIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the limit
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.200:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Wait a tiny bit and make another request to trigger rate limit
	time.Sleep(10 * time.Millisecond)

	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.200:12345"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code == http.StatusTooManyRequests {
		// Check response format
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("Retry-After"))
		assert.Contains(t, w.Body.String(), "rate limit exceeded")
	}
}
