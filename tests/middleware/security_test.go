package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders_DefaultConfig(t *testing.T) {
	cfg := &config.SecurityConfig{
		EnableHSTS:            false,
		ContentTypeNosniff:    true,
		FrameOptions:          "DENY",
		XSSProtection:         "1; mode=block",
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"), "HSTS should not be set when disabled")
}

func TestSecurityHeaders_HSTSEnabled(t *testing.T) {
	cfg := &config.SecurityConfig{
		EnableHSTS:            true,
		HSTSMaxAge:            31536000, // 1 year
		HSTSIncludeSubdomains: false,
		HSTSPreload:           false,
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "max-age=31536000", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_HSTSWithSubdomains(t *testing.T) {
	cfg := &config.SecurityConfig{
		EnableHSTS:            true,
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_HSTSWithPreload(t *testing.T) {
	cfg := &config.SecurityConfig{
		EnableHSTS:            true,
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "max-age=31536000; includeSubDomains; preload", w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_FrameOptionsSameOrigin(t *testing.T) {
	cfg := &config.SecurityConfig{
		FrameOptions: "SAMEORIGIN",
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_FrameOptionsDisabled(t *testing.T) {
	cfg := &config.SecurityConfig{
		FrameOptions: "",
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_ContentTypeNosniffDisabled(t *testing.T) {
	cfg := &config.SecurityConfig{
		ContentTypeNosniff: false,
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("X-Content-Type-Options"))
}

func TestSecurityHeaders_CustomCSP(t *testing.T) {
	cfg := &config.SecurityConfig{
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
	}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'", w.Header().Get("Content-Security-Policy"))
}

func TestSecurityHeaders_MinimalConfig(t *testing.T) {
	cfg := &config.SecurityConfig{}

	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// With minimal config, no security headers should be set
	assert.Empty(t, w.Header().Get("X-Content-Type-Options"))
	assert.Empty(t, w.Header().Get("X-Frame-Options"))
	assert.Empty(t, w.Header().Get("X-XSS-Protection"))
	assert.Empty(t, w.Header().Get("Content-Security-Policy"))
	assert.Empty(t, w.Header().Get("Referrer-Policy"))
	assert.Empty(t, w.Header().Get("Permissions-Policy"))
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_PassesThroughRequest(t *testing.T) {
	cfg := &config.SecurityConfig{
		ContentTypeNosniff: true,
	}

	handlerCalled := false
	handler := middleware.SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled, "Handler should be called")
	assert.Equal(t, "OK", w.Body.String())
}
