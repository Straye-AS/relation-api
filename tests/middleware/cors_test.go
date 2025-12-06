package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCORS_DevelopmentAllowsAllOrigins(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Location"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "development", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// In development with no configured origins, should allow all
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_ExplicitOrigins(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"https://app.example.com", "https://admin.example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "production", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test allowed origin
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"https://app.example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "production", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test disallowed origin
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://malicious.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should not have CORS headers for disallowed origin
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardOrigin(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "development", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "http://any-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightRequest(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"https://app.example.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Location", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           600,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "production", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "600", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORS_ActualRequest(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{"https://app.example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Location"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handlerCalled := false
	handler := middleware.CORS(cfg, "production", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled, "Handler should be called for actual request")
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Expose-Headers"), "Location")
}

func TestCORS_ProductionNoConfiguredOrigins(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "production", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// In production with no configured origins, should deny all
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_LocalEnvironment(t *testing.T) {
	cfg := &config.CORSConfig{
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	logger := zap.NewNop()
	handler := middleware.CORS(cfg, "local", logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Local environment should allow all like development
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}
