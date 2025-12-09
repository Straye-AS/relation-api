package repository

import (
	"context"

	"github.com/straye-as/relation-api/internal/auth"
	"gorm.io/gorm"
)

// MaxPageSize is the maximum allowed page size for paginated queries
const MaxPageSize = 200

// ApplyCompanyFilter applies the multi-tenant company filter to a GORM query
// This should be called on queries that need to be filtered by company_id
// If no filter is set (user has access to all companies), the query is returned unchanged
func ApplyCompanyFilter(ctx context.Context, query *gorm.DB) *gorm.DB {
	companyID := auth.GetEffectiveCompanyFilter(ctx)
	if companyID != nil {
		return query.Where("company_id = ?", *companyID)
	}
	return query
}

// ApplyCompanyFilterWithColumn applies the company filter using a specific column name
// Use this when the company_id column has a different name or needs table qualification
func ApplyCompanyFilterWithColumn(ctx context.Context, query *gorm.DB, columnName string) *gorm.DB {
	companyID := auth.GetEffectiveCompanyFilter(ctx)
	if companyID != nil {
		return query.Where(columnName+" = ?", *companyID)
	}
	return query
}

// ApplyCompanyFilterWithAlias applies the company filter using a table alias
// Use this when joining multiple tables and you need to specify which table's company_id to filter on
func ApplyCompanyFilterWithAlias(ctx context.Context, query *gorm.DB, tableAlias string) *gorm.DB {
	companyID := auth.GetEffectiveCompanyFilter(ctx)
	if companyID != nil {
		return query.Where(tableAlias+".company_id = ?", *companyID)
	}
	return query
}

// MustHaveCompanyAccess checks if the user has access to a specific company's data
// Returns true if access is allowed, false otherwise
// This is useful for single-record operations where you need to verify access
func MustHaveCompanyAccess(ctx context.Context, recordCompanyID string) bool {
	// Get the effective filter
	companyID := auth.GetEffectiveCompanyFilter(ctx)

	// If no filter, user has access to all
	if companyID == nil {
		return true
	}

	// Check if the record's company matches the filter
	return string(*companyID) == recordCompanyID
}
