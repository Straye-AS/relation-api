package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupMinimalTestDB creates a minimal test database for tenant filter tests
func setupMinimalTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	return db
}

// SimpleModel is a minimal model for testing company filter
type SimpleModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key"`
	Name      string
	CompanyID string `gorm:"column:company_id"`
}

func TestApplyCompanyFilter_WithFilter(t *testing.T) {
	db := setupMinimalTestDB(t)
	_ = db.AutoMigrate(&SimpleModel{})

	// Create a context with company filter
	stalbygg := domain.CompanyStalbygg
	filter := &auth.CompanyFilter{
		CompanyID: &stalbygg,
	}
	ctx := auth.WithCompanyFilter(context.Background(), filter)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return repository.ApplyCompanyFilter(ctx, tx.Model(&SimpleModel{})).Find(&[]SimpleModel{})
	})

	assert.Contains(t, sql, "company_id", "Query should contain company_id filter")
}

func TestApplyCompanyFilter_WithoutFilter(t *testing.T) {
	db := setupMinimalTestDB(t)
	_ = db.AutoMigrate(&SimpleModel{})

	// Create a context without company filter (Gruppen user)
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}
	ctx := auth.WithUserContext(context.Background(), userCtx)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return repository.ApplyCompanyFilter(ctx, tx.Model(&SimpleModel{})).Find(&[]SimpleModel{})
	})

	// For Gruppen users, no company filter should be applied
	assert.NotContains(t, sql, "company_id =", "Query should not contain company_id filter for Gruppen users")
}

func TestApplyCompanyFilterWithColumn(t *testing.T) {
	db := setupMinimalTestDB(t)
	_ = db.AutoMigrate(&SimpleModel{})

	// Create a context with company filter
	stalbygg := domain.CompanyStalbygg
	filter := &auth.CompanyFilter{
		CompanyID: &stalbygg,
	}
	ctx := auth.WithCompanyFilter(context.Background(), filter)

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return repository.ApplyCompanyFilterWithColumn(ctx, tx.Model(&SimpleModel{}), "projects.company_id").Find(&[]SimpleModel{})
	})

	assert.Contains(t, sql, "projects.company_id", "Query should contain qualified column name")
}

func TestMustHaveCompanyAccess_WithFilter(t *testing.T) {
	tests := []struct {
		name            string
		filterCompanyID domain.CompanyID
		recordCompanyID string
		expected        bool
	}{
		{
			name:            "matching company",
			filterCompanyID: domain.CompanyStalbygg,
			recordCompanyID: string(domain.CompanyStalbygg),
			expected:        true,
		},
		{
			name:            "non-matching company",
			filterCompanyID: domain.CompanyStalbygg,
			recordCompanyID: string(domain.CompanyHybridbygg),
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &auth.CompanyFilter{
				CompanyID: &tt.filterCompanyID,
			}
			ctx := auth.WithCompanyFilter(context.Background(), filter)

			result := repository.MustHaveCompanyAccess(ctx, tt.recordCompanyID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMustHaveCompanyAccess_NoFilter(t *testing.T) {
	// Create a context without company filter (super admin or Gruppen user)
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}
	ctx := auth.WithUserContext(context.Background(), userCtx)

	// Without filter, user should have access to all records
	result := repository.MustHaveCompanyAccess(ctx, string(domain.CompanyStalbygg))
	assert.True(t, result, "User without filter should have access to all companies")

	result = repository.MustHaveCompanyAccess(ctx, string(domain.CompanyHybridbygg))
	assert.True(t, result, "User without filter should have access to all companies")
}

func TestGetEffectiveCompanyFilter_Priority(t *testing.T) {
	// Test that explicit company filter takes precedence over user context

	// Create user context with subsidiary company
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyStalbygg,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}
	ctx := auth.WithUserContext(context.Background(), userCtx)

	// Without explicit filter, user's company should be used
	filter := auth.GetEffectiveCompanyFilter(ctx)
	assert.NotNil(t, filter)
	assert.Equal(t, domain.CompanyStalbygg, *filter)

	// With explicit filter, it should take precedence
	hybridbygg := domain.CompanyHybridbygg
	explicitFilter := &auth.CompanyFilter{
		CompanyID: &hybridbygg,
	}
	ctx = auth.WithCompanyFilter(ctx, explicitFilter)

	filter = auth.GetEffectiveCompanyFilter(ctx)
	assert.NotNil(t, filter)
	assert.Equal(t, domain.CompanyHybridbygg, *filter)
}

func TestGetEffectiveCompanyFilter_GruppenUser(t *testing.T) {
	// Gruppen users should not have a filter by default
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyGruppen,
		Roles:     []domain.UserRoleType{domain.RoleMarket},
	}
	ctx := auth.WithUserContext(context.Background(), userCtx)

	filter := auth.GetEffectiveCompanyFilter(ctx)
	assert.Nil(t, filter, "Gruppen users should have no filter by default")
}

func TestGetEffectiveCompanyFilter_SuperAdmin(t *testing.T) {
	// Super admins should not have a filter by default
	userCtx := &auth.UserContext{
		UserID:    uuid.New(),
		CompanyID: domain.CompanyStalbygg,
		Roles:     []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	ctx := auth.WithUserContext(context.Background(), userCtx)

	filter := auth.GetEffectiveCompanyFilter(ctx)
	assert.Nil(t, filter, "Super admins should have no filter by default")
}
