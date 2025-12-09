package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/http/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCompanyFilterMiddleware_GruppenUser_NoFilter(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Gruppen user
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/customers", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.Nil(t, capturedFilter.CompanyID, "Gruppen user without filter should have nil CompanyID")
}

func TestCompanyFilterMiddleware_GruppenUser_WithFilter(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Gruppen user
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/customers?company_id=stalbygg", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.NotNil(t, capturedFilter.CompanyID)
	assert.Equal(t, domain.CompanyStalbygg, *capturedFilter.CompanyID)
	assert.True(t, capturedFilter.RequestedByGruppenUser)
}

func TestCompanyFilterMiddleware_SubsidiaryUser_AutoFilter(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Stalbygg user
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyStalbygg,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/customers", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.NotNil(t, capturedFilter.CompanyID)
	assert.Equal(t, domain.CompanyStalbygg, *capturedFilter.CompanyID)
	assert.False(t, capturedFilter.RequestedByGruppenUser)
}

func TestCompanyFilterMiddleware_SubsidiaryUser_DeniedAccessToOtherCompany(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Stalbygg user
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyStalbygg,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Try to access hybridbygg data
	req := httptest.NewRequest("GET", "/api/v1/customers?company_id=hybridbygg", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCompanyFilterMiddleware_SubsidiaryUser_CanAccessOwnCompany(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Stalbygg user
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyStalbygg,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Explicitly request own company
	req := httptest.NewRequest("GET", "/api/v1/customers?company_id=stalbygg", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.NotNil(t, capturedFilter.CompanyID)
	assert.Equal(t, domain.CompanyStalbygg, *capturedFilter.CompanyID)
}

func TestCompanyFilterMiddleware_InvalidCompanyID(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Gruppen user (who would normally have access)
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Try to access invalid company
	req := httptest.NewRequest("GET", "/api/v1/customers?company_id=invalid_company", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCompanyFilterMiddleware_SuperAdmin_CanAccessAnyCompany(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Super Admin from Stalbygg
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyStalbygg,
		Roles:     []domain.UserRoleType{domain.RoleSuperAdmin},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Super admin requesting another company's data
	req := httptest.NewRequest("GET", "/api/v1/customers?company_id=hybridbygg", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.NotNil(t, capturedFilter.CompanyID)
	assert.Equal(t, domain.CompanyHybridbygg, *capturedFilter.CompanyID)
}

func TestCompanyFilterMiddleware_NoUserContext(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	handlerCalled := false
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	// Request without user context
	req := httptest.NewRequest("GET", "/api/v1/customers", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled, "Handler should be called even without user context")
}

func TestCompanyFilterMiddleware_APIServiceUser_WithCompanyHeader(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Simulate an API service user (SuperAdmin + APIService roles) with X-Company-ID set to stalbygg
	// This is the scenario when using API key authentication with X-Company-ID header
	userCtx := &auth.UserContext{
		UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		CompanyID: domain.CompanyStalbygg, // Set via X-Company-ID header in auth middleware
		Roles:     []domain.UserRoleType{domain.RoleSuperAdmin, domain.RoleAPIService},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Request WITHOUT company_id query param - should still filter by stalbygg
	req := httptest.NewRequest("GET", "/api/v1/dashboard/metrics", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.NotNil(t, capturedFilter.CompanyID, "API service user with specific company should have company filter applied")
	assert.Equal(t, domain.CompanyStalbygg, *capturedFilter.CompanyID)
}

func TestCompanyFilterMiddleware_APIServiceUser_WithGruppenHeader(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// API service user with X-Company-ID set to gruppen (or not set, defaulting to gruppen)
	userCtx := &auth.UserContext{
		UserID:    uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		CompanyID: domain.CompanyGruppen, // Default when X-Company-ID not provided
		Roles:     []domain.UserRoleType{domain.RoleSuperAdmin, domain.RoleAPIService},
	}

	var capturedFilter *auth.CompanyFilter
	handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedFilter, _ = auth.CompanyFilterFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/dashboard/metrics", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotNil(t, capturedFilter)
	assert.Nil(t, capturedFilter.CompanyID, "API service user with gruppen company should see all data")
}

func TestCompanyFilterMiddleware_AllCompaniesValid(t *testing.T) {
	logger := zap.NewNop()
	m := middleware.NewCompanyFilterMiddleware(logger)

	// Create a Gruppen user
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}

	validCompanies := []domain.CompanyID{
		domain.CompanyAll,
		domain.CompanyGruppen,
		domain.CompanyStalbygg,
		domain.CompanyHybridbygg,
		domain.CompanyIndustri,
		domain.CompanyTak,
		domain.CompanyMontasje,
	}

	for _, companyID := range validCompanies {
		t.Run(string(companyID), func(t *testing.T) {
			handler := m.Filter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/api/v1/customers?company_id="+string(companyID), nil)
			req = req.WithContext(auth.WithUserContext(req.Context(), userCtx))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "Company %s should be valid", companyID)
		})
	}
}
