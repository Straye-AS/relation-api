package middleware

import (
	"net/http"

	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
)

// CompanyFilterMiddleware handles multi-tenant data isolation
// It extracts the user's company context and optionally allows
// Gruppen users to filter by a specific subsidiary company
type CompanyFilterMiddleware struct {
	logger *zap.Logger
}

// NewCompanyFilterMiddleware creates a new company filter middleware
func NewCompanyFilterMiddleware(logger *zap.Logger) *CompanyFilterMiddleware {
	return &CompanyFilterMiddleware{
		logger: logger,
	}
}

// Filter is the middleware handler that sets the effective company filter in context
// - Gruppen users and super admins can optionally filter by ?company_id=<company>
// - Subsidiary users are always filtered to their own company
// - If no filter is specified, Gruppen users see all data, subsidiary users see their company's data
func (m *CompanyFilterMiddleware) Filter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := auth.FromContext(r.Context())
		if !ok {
			// No user context - let request proceed without company filter
			// Authentication middleware should have already rejected unauthenticated requests
			next.ServeHTTP(w, r)
			return
		}

		var filter *auth.CompanyFilter

		// Check for company_id query parameter
		requestedCompanyID := r.URL.Query().Get("company_id")

		if requestedCompanyID != "" {
			// User is requesting a specific company
			companyID := domain.CompanyID(requestedCompanyID)

			// Validate the company ID
			if !isValidCompanyID(companyID) {
				http.Error(w, "Invalid company_id parameter", http.StatusBadRequest)
				return
			}

			// Check if user can access the requested company
			if !userCtx.CanAccessCompany(companyID) {
				m.logger.Warn("user attempted to access unauthorized company",
					zap.String("user_id", userCtx.UserID.String()),
					zap.String("user_company", string(userCtx.CompanyID)),
					zap.String("requested_company", requestedCompanyID),
				)
				http.Error(w, "Access denied: you cannot access data for this company", http.StatusForbidden)
				return
			}

			// Gruppen user explicitly requesting a specific company
			filter = &auth.CompanyFilter{
				CompanyID:              &companyID,
				RequestedByGruppenUser: userCtx.IsGruppenUser(),
			}
		} else {
			// No specific company requested via query param
			// Use the user's CompanyID (which may have been set via X-Company-Id header)
			// If the company is set and is not gruppen/all, use it as filter
			if userCtx.CompanyID != "" && userCtx.CompanyID != domain.CompanyGruppen && userCtx.CompanyID != domain.CompanyAll {
				companyID := userCtx.CompanyID
				filter = &auth.CompanyFilter{
					CompanyID:              &companyID,
					RequestedByGruppenUser: false,
				}
			} else {
				// No specific company - show all data
				filter = &auth.CompanyFilter{
					CompanyID:              nil,
					RequestedByGruppenUser: false,
				}
			}
		}

		// Add company filter to context
		ctx := auth.WithCompanyFilter(r.Context(), filter)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isValidCompanyID checks if the provided company ID is a valid Straye company
func isValidCompanyID(companyID domain.CompanyID) bool {
	validCompanies := []domain.CompanyID{
		domain.CompanyAll,
		domain.CompanyGruppen,
		domain.CompanyStalbygg,
		domain.CompanyHybridbygg,
		domain.CompanyIndustri,
		domain.CompanyTak,
		domain.CompanyMontasje,
	}

	for _, valid := range validCompanies {
		if companyID == valid {
			return true
		}
	}
	return false
}
