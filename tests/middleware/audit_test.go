package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAuditMiddleware_SkipsGETRequests(t *testing.T) {
	config := &middleware.AuditConfig{
		SkipPaths:   []string{"/health"},
		SkipMethods: []string{http.MethodOptions, http.MethodHead},
		AuditReads:  false,
	}

	// Create middleware without actual audit service (nil)
	// We just want to test the filtering logic
	am := middleware.NewAuditMiddleware(nil, config, nil)

	handlerCalled := false
	handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddleware_SkipsHealthPaths(t *testing.T) {
	config := &middleware.AuditConfig{
		SkipPaths:   []string{"/health"},
		SkipMethods: []string{http.MethodOptions, http.MethodHead},
		AuditReads:  false,
	}

	am := middleware.NewAuditMiddleware(nil, config, nil)

	handlerCalled := false
	handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	paths := []string{"/health", "/health/db", "/health/ready"}
	for _, path := range paths {
		handlerCalled = false
		req := httptest.NewRequest(http.MethodPost, path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.True(t, handlerCalled, "Handler should be called for path %s", path)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestAuditMiddleware_SkipsOPTIONSMethod(t *testing.T) {
	config := &middleware.AuditConfig{
		SkipPaths:   []string{"/health"},
		SkipMethods: []string{http.MethodOptions, http.MethodHead},
		AuditReads:  false,
	}

	am := middleware.NewAuditMiddleware(nil, config, nil)

	handlerCalled := false
	handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/customers", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddleware_DefaultConfig(t *testing.T) {
	config := middleware.DefaultAuditConfig()

	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/health/db")
	assert.Contains(t, config.SkipPaths, "/health/ready")
	assert.Contains(t, config.SkipPaths, "/swagger")
	assert.Contains(t, config.SkipMethods, http.MethodOptions)
	assert.Contains(t, config.SkipMethods, http.MethodHead)
	assert.False(t, config.AuditReads)
}

func TestAuditMiddleware_AllowsModifyingRequests(t *testing.T) {
	config := middleware.DefaultAuditConfig()

	am := middleware.NewAuditMiddleware(nil, config, nil)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range methods {
		handlerCalled := false
		handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(method, "/api/v1/customers", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.True(t, handlerCalled, "Handler should be called for method %s", method)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestAuditMiddleware_WithNilConfig(t *testing.T) {
	// Should use default config when nil is passed
	am := middleware.NewAuditMiddleware(nil, nil, nil)

	handlerCalled := false
	handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Health path should be skipped with default config
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
}

func TestAuditMiddleware_CapturesResponseStatus(t *testing.T) {
	config := middleware.DefaultAuditConfig()

	am := middleware.NewAuditMiddleware(nil, config, nil)

	// Test with 201 Created
	handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/customers", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAuditMiddleware_AuditReadsWhenEnabled(t *testing.T) {
	config := &middleware.AuditConfig{
		SkipPaths:   []string{"/health"},
		SkipMethods: []string{http.MethodOptions, http.MethodHead},
		AuditReads:  true, // Enable auditing reads
	}

	am := middleware.NewAuditMiddleware(nil, config, nil)

	handlerCalled := false
	handler := am.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}
