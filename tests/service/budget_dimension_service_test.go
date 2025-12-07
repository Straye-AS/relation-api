package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestBudgetDimensionService is an integration test suite for BudgetDimensionService
// Requires a running PostgreSQL database with migrations applied

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=relation_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	return db
}

func setupTestService(t *testing.T, db *gorm.DB) (*service.BudgetDimensionService, *testFixtures) {
	logger := zap.NewNop()

	dimensionRepo := repository.NewBudgetDimensionRepository(db)
	categoryRepo := repository.NewBudgetDimensionCategoryRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	svc := service.NewBudgetDimensionService(
		dimensionRepo,
		categoryRepo,
		offerRepo,
		projectRepo,
		activityRepo,
		logger,
	)

	fixtures := &testFixtures{
		db:          db,
		offerRepo:   offerRepo,
		projectRepo: projectRepo,
	}

	return svc, fixtures
}

type testFixtures struct {
	db          *gorm.DB
	offerRepo   *repository.OfferRepository
	projectRepo *repository.ProjectRepository
}

func (f *testFixtures) createTestOffer(t *testing.T, ctx context.Context) *domain.Offer {
	// First create a customer
	customer := &domain.Customer{
		Name:    "Test Customer",
		Email:   "test@example.com",
		Phone:   "12345678",
		Country: "Norway",
		Status:  domain.CustomerStatusActive,
		Tier:    domain.CustomerTierBronze,
	}
	err := f.db.WithContext(ctx).Create(customer).Error
	require.NoError(t, err)

	offer := &domain.Offer{
		Title:             "Test Offer",
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             domain.OfferPhaseDraft,
		Probability:       50,
		Value:             0,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: "test-user",
	}
	err = f.offerRepo.Create(ctx, offer)
	require.NoError(t, err)

	return offer
}

func (f *testFixtures) createTestProject(t *testing.T, ctx context.Context) *domain.Project {
	// First create a customer
	customer := &domain.Customer{
		Name:    "Test Project Customer",
		Email:   "test-project@example.com",
		Phone:   "12345678",
		Country: "Norway",
		Status:  domain.CustomerStatusActive,
		Tier:    domain.CustomerTierBronze,
	}
	err := f.db.WithContext(ctx).Create(customer).Error
	require.NoError(t, err)

	project := &domain.Project{
		Name:         "Test Project",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Status:       domain.ProjectStatusActive,
		StartDate:    time.Now(),
		Budget:       0,
		ManagerID:    "test-manager",
	}
	err = f.projectRepo.Create(ctx, project)
	require.NoError(t, err)

	return project
}

func (f *testFixtures) createTestCategory(t *testing.T, ctx context.Context) *domain.BudgetDimensionCategory {
	category := &domain.BudgetDimensionCategory{
		ID:           "test-category-" + uuid.New().String()[:8],
		Name:         "Test Category",
		Description:  "A test category",
		DisplayOrder: 0,
		IsActive:     true,
	}
	err := f.db.WithContext(ctx).Create(category).Error
	require.NoError(t, err)
	return category
}

func (f *testFixtures) cleanup(t *testing.T) {
	// Clean up in reverse order of dependencies
	f.db.Exec("DELETE FROM activities WHERE target_type IN ('Offer', 'Project')")
	f.db.Exec("DELETE FROM budget_dimensions")
	f.db.Exec("DELETE FROM budget_dimension_categories WHERE id LIKE 'test-category-%'")
	f.db.Exec("DELETE FROM offer_items")
	f.db.Exec("DELETE FROM offers")
	f.db.Exec("DELETE FROM projects")
	f.db.Exec("DELETE FROM customers WHERE email LIKE 'test%@example.com'")
}

func contextWithUser(ctx context.Context) context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		Email:       "test@example.com",
		DisplayName: "Test User",
		CompanyID:   domain.CompanyStalbygg,
	}
	return auth.WithUserContext(ctx, userCtx)
}

// =============================================================================
// Create Tests
// =============================================================================

func TestCreate_Success(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Steel Structure",
		Cost:       100000,
		Revenue:    150000,
	}

	dto, err := svc.Create(ctx, req)

	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, dto.ID)
	assert.Equal(t, domain.BudgetParentOffer, dto.ParentType)
	assert.Equal(t, offer.ID, dto.ParentID)
	assert.Equal(t, "Steel Structure", dto.Name)
	assert.Equal(t, 100000.0, dto.Cost)
	assert.Equal(t, 150000.0, dto.Revenue)
	assert.Equal(t, 0, dto.DisplayOrder) // First item gets display order 0
}

func TestCreate_WithCategory(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)
	category := fixtures.createTestCategory(t, ctx)

	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CategoryID: &category.ID,
		Cost:       50000,
		Revenue:    75000,
	}

	dto, err := svc.Create(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, dto.CategoryID)
	assert.Equal(t, category.ID, *dto.CategoryID)
	assert.Equal(t, category.Name, dto.Name) // Name should come from category
	assert.NotNil(t, dto.Category)
	assert.Equal(t, category.Name, dto.Category.Name)
}

func TestCreate_WithMarginOverride(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	targetMargin := 20.0
	req := &domain.CreateBudgetDimensionRequest{
		ParentType:          domain.BudgetParentOffer,
		ParentID:            offer.ID,
		CustomName:          "Margin Override Item",
		Cost:                100000,
		Revenue:             0, // Will be calculated from margin
		TargetMarginPercent: &targetMargin,
		MarginOverride:      true,
	}

	dto, err := svc.Create(ctx, req)

	require.NoError(t, err)
	assert.True(t, dto.MarginOverride)
	assert.NotNil(t, dto.TargetMarginPercent)
	assert.Equal(t, 20.0, *dto.TargetMarginPercent)
	// Revenue should be calculated by DB trigger: Revenue = Cost / (1 - TargetMargin/100)
	// For 20% margin and 100000 cost: Revenue = 100000 / 0.8 = 125000
	assert.InDelta(t, 125000.0, dto.Revenue, 1.0) // Allow small delta for floating point
}

func TestCreate_InvalidParent(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	nonExistentID := uuid.New()

	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   nonExistentID,
		CustomName: "Test Item",
		Cost:       100000,
		Revenue:    150000,
	}

	dto, err := svc.Create(ctx, req)

	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrParentNotFound))
}

func TestCreate_InvalidCategory(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	nonExistentCategory := "non-existent-category"
	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CategoryID: &nonExistentCategory,
		Cost:       100000,
		Revenue:    150000,
	}

	dto, err := svc.Create(ctx, req)

	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidCategory))
}

func TestCreate_MissingName(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		// No CategoryID and no CustomName
		Cost:    100000,
		Revenue: 150000,
	}

	dto, err := svc.Create(ctx, req)

	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrMissingName))
}

func TestCreate_InvalidCost(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	testCases := []struct {
		name string
		cost float64
	}{
		{"zero cost", 0},
		{"negative cost", -100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &domain.CreateBudgetDimensionRequest{
				ParentType: domain.BudgetParentOffer,
				ParentID:   offer.ID,
				CustomName: "Test Item",
				Cost:       tc.cost,
				Revenue:    150000,
			}

			dto, err := svc.Create(ctx, req)

			assert.Nil(t, dto)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidCost))
		})
	}
}

func TestCreate_InvalidMargin(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	testCases := []struct {
		name         string
		targetMargin *float64
	}{
		{"margin too high", floatPtr(100)},
		{"margin too high 2", floatPtr(150)},
		{"negative margin", floatPtr(-10)},
		{"nil margin with override", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &domain.CreateBudgetDimensionRequest{
				ParentType:          domain.BudgetParentOffer,
				ParentID:            offer.ID,
				CustomName:          "Test Item",
				Cost:                100000,
				Revenue:             0,
				TargetMarginPercent: tc.targetMargin,
				MarginOverride:      true,
			}

			dto, err := svc.Create(ctx, req)

			assert.Nil(t, dto)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidTargetMargin))
		})
	}
}

func TestCreate_AutoDisplayOrder(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create first item
	req1 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "First Item",
		Cost:       100000,
		Revenue:    150000,
	}
	dto1, err := svc.Create(ctx, req1)
	require.NoError(t, err)

	// Create second item
	req2 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Second Item",
		Cost:       50000,
		Revenue:    75000,
	}
	dto2, err := svc.Create(ctx, req2)
	require.NoError(t, err)

	// Create third item
	req3 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Third Item",
		Cost:       25000,
		Revenue:    40000,
	}
	dto3, err := svc.Create(ctx, req3)
	require.NoError(t, err)

	assert.Equal(t, 0, dto1.DisplayOrder)
	assert.Equal(t, 1, dto2.DisplayOrder)
	assert.Equal(t, 2, dto3.DisplayOrder)
}

// =============================================================================
// GetByID Tests
// =============================================================================

func TestGetByID_Found(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)
	category := fixtures.createTestCategory(t, ctx)

	// Create a dimension with category
	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CategoryID: &category.ID,
		Cost:       100000,
		Revenue:    150000,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	// Get by ID
	dto, err := svc.GetByID(ctx, created.ID)

	require.NoError(t, err)
	assert.Equal(t, created.ID, dto.ID)
	assert.NotNil(t, dto.Category)
	assert.Equal(t, category.Name, dto.Category.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())

	dto, err := svc.GetByID(ctx, uuid.New())

	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrBudgetDimensionNotFound))
}

// =============================================================================
// Update Tests
// =============================================================================

func TestUpdate_Success(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create a dimension
	createReq := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Original Name",
		Cost:       100000,
		Revenue:    150000,
	}
	created, err := svc.Create(ctx, createReq)
	require.NoError(t, err)

	// Update it
	updateReq := &domain.UpdateBudgetDimensionRequest{
		CustomName:   "Updated Name",
		Cost:         200000,
		Revenue:      300000,
		Description:  "Updated description",
		DisplayOrder: 5,
	}

	dto, err := svc.Update(ctx, created.ID, updateReq)

	require.NoError(t, err)
	assert.Equal(t, "Updated Name", dto.CustomName)
	assert.Equal(t, "Updated Name", dto.Name)
	assert.Equal(t, 200000.0, dto.Cost)
	assert.Equal(t, 300000.0, dto.Revenue)
	assert.Equal(t, "Updated description", dto.Description)
	assert.Equal(t, 5, dto.DisplayOrder)
}

func TestUpdate_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())

	updateReq := &domain.UpdateBudgetDimensionRequest{
		CustomName: "Updated Name",
		Cost:       200000,
		Revenue:    300000,
	}

	dto, err := svc.Update(ctx, uuid.New(), updateReq)

	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrBudgetDimensionNotFound))
}

func TestUpdate_InvalidCost(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create a dimension
	createReq := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Test Item",
		Cost:       100000,
		Revenue:    150000,
	}
	created, err := svc.Create(ctx, createReq)
	require.NoError(t, err)

	// Try to update with invalid cost
	updateReq := &domain.UpdateBudgetDimensionRequest{
		CustomName: "Test Item",
		Cost:       0, // Invalid
		Revenue:    150000,
	}

	dto, err := svc.Update(ctx, created.ID, updateReq)

	assert.Nil(t, dto)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrInvalidCost))
}

func TestUpdate_UpdatesParentTotals(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create a dimension
	createReq := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Test Item",
		Cost:       100000,
		Revenue:    150000,
	}
	created, err := svc.Create(ctx, createReq)
	require.NoError(t, err)

	// Update with new values
	updateReq := &domain.UpdateBudgetDimensionRequest{
		CustomName: "Test Item",
		Cost:       200000,
		Revenue:    350000,
	}

	_, err = svc.Update(ctx, created.ID, updateReq)
	require.NoError(t, err)

	// Verify parent offer was updated
	updatedOffer, err := fixtures.offerRepo.GetByID(ctx, offer.ID)
	require.NoError(t, err)
	assert.Equal(t, 350000.0, updatedOffer.Value)
}

// =============================================================================
// Delete Tests
// =============================================================================

func TestDelete_Success(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create a dimension
	createReq := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Test Item",
		Cost:       100000,
		Revenue:    150000,
	}
	created, err := svc.Create(ctx, createReq)
	require.NoError(t, err)

	// Delete it
	err = svc.Delete(ctx, created.ID)

	require.NoError(t, err)

	// Verify it's gone
	_, err = svc.GetByID(ctx, created.ID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrBudgetDimensionNotFound))
}

func TestDelete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())

	err := svc.Delete(ctx, uuid.New())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrBudgetDimensionNotFound))
}

func TestDelete_UpdatesParentTotals(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create two dimensions
	req1 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Item 1",
		Cost:       100000,
		Revenue:    150000,
	}
	created1, err := svc.Create(ctx, req1)
	require.NoError(t, err)

	req2 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Item 2",
		Cost:       50000,
		Revenue:    75000,
	}
	_, err = svc.Create(ctx, req2)
	require.NoError(t, err)

	// Delete first item
	err = svc.Delete(ctx, created1.ID)
	require.NoError(t, err)

	// Verify parent offer value updated (should be just item 2)
	updatedOffer, err := fixtures.offerRepo.GetByID(ctx, offer.ID)
	require.NoError(t, err)
	assert.Equal(t, 75000.0, updatedOffer.Value)
}

// =============================================================================
// List Tests
// =============================================================================

func TestListByParent_Success(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create dimensions
	for i := 0; i < 3; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType: domain.BudgetParentOffer,
			ParentID:   offer.ID,
			CustomName: "Item",
			Cost:       float64((i + 1) * 10000),
			Revenue:    float64((i + 1) * 15000),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// List
	dtos, err := svc.ListByParent(ctx, domain.BudgetParentOffer, offer.ID)

	require.NoError(t, err)
	assert.Len(t, dtos, 3)
	// Verify ordered by display_order
	assert.Equal(t, 0, dtos[0].DisplayOrder)
	assert.Equal(t, 1, dtos[1].DisplayOrder)
	assert.Equal(t, 2, dtos[2].DisplayOrder)
}

func TestListByParent_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	dtos, err := svc.ListByParent(ctx, domain.BudgetParentOffer, offer.ID)

	require.NoError(t, err)
	assert.Len(t, dtos, 0)
}

func TestListByParentPaginated_FirstPage(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create 5 dimensions
	for i := 0; i < 5; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType: domain.BudgetParentOffer,
			ParentID:   offer.ID,
			CustomName: "Item",
			Cost:       float64((i + 1) * 10000),
			Revenue:    float64((i + 1) * 15000),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// Get first page with 2 items
	response, err := svc.ListByParentPaginated(ctx, domain.BudgetParentOffer, offer.ID, 1, 2)

	require.NoError(t, err)
	assert.Equal(t, int64(5), response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 2, response.PageSize)
	assert.Equal(t, 3, response.TotalPages)
	assert.Len(t, response.Data, 2)
}

func TestListByParentPaginated_Clamping(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Request with page size > 200
	response, err := svc.ListByParentPaginated(ctx, domain.BudgetParentOffer, offer.ID, 1, 500)

	require.NoError(t, err)
	assert.Equal(t, 200, response.PageSize) // Should be clamped to 200
}

// =============================================================================
// Summary Tests (CalculateTotals)
// =============================================================================

func TestGetSummary_WithDimensions(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create dimensions
	req1 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Item 1",
		Cost:       100000,
		Revenue:    150000,
	}
	_, err := svc.Create(ctx, req1)
	require.NoError(t, err)

	req2 := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Item 2",
		Cost:       50000,
		Revenue:    80000,
	}
	_, err = svc.Create(ctx, req2)
	require.NoError(t, err)

	// Get summary
	summary, err := svc.GetSummary(ctx, domain.BudgetParentOffer, offer.ID)

	require.NoError(t, err)
	assert.Equal(t, 2, summary.DimensionCount)
	assert.Equal(t, 150000.0, summary.TotalCost)                // 100000 + 50000
	assert.Equal(t, 230000.0, summary.TotalRevenue)             // 150000 + 80000
	assert.Equal(t, 80000.0, summary.TotalProfit)               // 230000 - 150000
	assert.InDelta(t, 34.78, summary.OverallMarginPercent, 0.1) // (80000 / 230000) * 100
}

func TestGetSummary_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Get summary without any dimensions
	summary, err := svc.GetSummary(ctx, domain.BudgetParentOffer, offer.ID)

	require.NoError(t, err)
	assert.Equal(t, 0, summary.DimensionCount)
	assert.Equal(t, 0.0, summary.TotalCost)
	assert.Equal(t, 0.0, summary.TotalRevenue)
	assert.Equal(t, 0.0, summary.TotalProfit)
	assert.Equal(t, 0.0, summary.OverallMarginPercent)
}

// =============================================================================
// Reorder Tests
// =============================================================================

func TestReorderDimensions_Success(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create 3 dimensions
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType: domain.BudgetParentOffer,
			ParentID:   offer.ID,
			CustomName: "Item",
			Cost:       float64((i + 1) * 10000),
			Revenue:    float64((i + 1) * 15000),
		}
		dto, err := svc.Create(ctx, req)
		require.NoError(t, err)
		ids = append(ids, dto.ID)
	}

	// Reverse the order
	reordered := []uuid.UUID{ids[2], ids[1], ids[0]}
	err := svc.ReorderDimensions(ctx, domain.BudgetParentOffer, offer.ID, reordered)

	require.NoError(t, err)

	// Verify new order
	dtos, err := svc.ListByParent(ctx, domain.BudgetParentOffer, offer.ID)
	require.NoError(t, err)

	assert.Equal(t, ids[2], dtos[0].ID)
	assert.Equal(t, ids[1], dtos[1].ID)
	assert.Equal(t, ids[0], dtos[2].ID)
}

func TestReorderDimensions_InvalidCount(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create 2 dimensions
	for i := 0; i < 2; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType: domain.BudgetParentOffer,
			ParentID:   offer.ID,
			CustomName: "Item",
			Cost:       float64((i + 1) * 10000),
			Revenue:    float64((i + 1) * 15000),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// Try to reorder with wrong number of IDs
	wrongIDs := []uuid.UUID{uuid.New()} // Only 1 instead of 2
	err := svc.ReorderDimensions(ctx, domain.BudgetParentOffer, offer.ID, wrongIDs)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrReorderCountMismatch))
}

// =============================================================================
// Clone Tests
// =============================================================================

func TestCloneDimensions_OfferToProject(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)
	project := fixtures.createTestProject(t, ctx)

	// Create dimensions on offer
	for i := 0; i < 3; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType: domain.BudgetParentOffer,
			ParentID:   offer.ID,
			CustomName: "Item",
			Cost:       float64((i + 1) * 10000),
			Revenue:    float64((i + 1) * 15000),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// Clone to project
	cloned, err := svc.CloneDimensions(ctx, domain.BudgetParentOffer, offer.ID, domain.BudgetParentProject, project.ID)

	require.NoError(t, err)
	assert.Len(t, cloned, 3)

	// Verify all cloned to project
	for _, dto := range cloned {
		assert.Equal(t, domain.BudgetParentProject, dto.ParentType)
		assert.Equal(t, project.ID, dto.ParentID)
	}
}

func TestCloneDimensions_PreservesOrder(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)
	project := fixtures.createTestProject(t, ctx)

	// Create dimensions with specific order
	for i := 0; i < 3; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType:   domain.BudgetParentOffer,
			ParentID:     offer.ID,
			CustomName:   "Item",
			Cost:         float64((i + 1) * 10000),
			Revenue:      float64((i + 1) * 15000),
			DisplayOrder: i * 10, // 0, 10, 20
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// Clone
	cloned, err := svc.CloneDimensions(ctx, domain.BudgetParentOffer, offer.ID, domain.BudgetParentProject, project.ID)

	require.NoError(t, err)

	// Verify order preserved
	assert.Equal(t, 0, cloned[0].DisplayOrder)
	assert.Equal(t, 10, cloned[1].DisplayOrder)
	assert.Equal(t, 20, cloned[2].DisplayOrder)
}

func TestCloneDimensions_PreservesFinancials(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)
	project := fixtures.createTestProject(t, ctx)

	// Create dimension with specific financials
	targetMargin := 25.0
	req := &domain.CreateBudgetDimensionRequest{
		ParentType:          domain.BudgetParentOffer,
		ParentID:            offer.ID,
		CustomName:          "Test Item",
		Cost:                100000,
		Revenue:             0, // Will be calculated
		TargetMarginPercent: &targetMargin,
		MarginOverride:      true,
	}
	original, err := svc.Create(ctx, req)
	require.NoError(t, err)

	// Clone
	cloned, err := svc.CloneDimensions(ctx, domain.BudgetParentOffer, offer.ID, domain.BudgetParentProject, project.ID)

	require.NoError(t, err)
	require.Len(t, cloned, 1)

	// Verify financials preserved
	assert.Equal(t, original.Cost, cloned[0].Cost)
	assert.InDelta(t, original.Revenue, cloned[0].Revenue, 1.0)
	assert.True(t, cloned[0].MarginOverride)
	assert.NotNil(t, cloned[0].TargetMarginPercent)
	assert.Equal(t, 25.0, *cloned[0].TargetMarginPercent)
}

func TestCloneDimensions_SourceEmpty(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)
	project := fixtures.createTestProject(t, ctx)

	// Try to clone from empty source
	cloned, err := svc.CloneDimensions(ctx, domain.BudgetParentOffer, offer.ID, domain.BudgetParentProject, project.ID)

	assert.Nil(t, cloned)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrSourceDimensionsNotFound))
}

func TestCloneDimensions_InvalidSource(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	project := fixtures.createTestProject(t, ctx)

	// Try to clone from non-existent source
	cloned, err := svc.CloneDimensions(ctx, domain.BudgetParentOffer, uuid.New(), domain.BudgetParentProject, project.ID)

	assert.Nil(t, cloned)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrParentNotFound))
}

func TestCloneDimensions_InvalidTarget(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create dimension on offer
	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentOffer,
		ParentID:   offer.ID,
		CustomName: "Test Item",
		Cost:       100000,
		Revenue:    150000,
	}
	_, err := svc.Create(ctx, req)
	require.NoError(t, err)

	// Try to clone to non-existent target
	cloned, err := svc.CloneDimensions(ctx, domain.BudgetParentOffer, offer.ID, domain.BudgetParentProject, uuid.New())

	assert.Nil(t, cloned)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, service.ErrParentNotFound))
}

// =============================================================================
// Validation Tests (ValidateMarginOverride)
// =============================================================================

func TestValidateMarginOverride_Valid(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	testCases := []struct {
		name         string
		targetMargin float64
	}{
		{"zero margin", 0},
		{"small margin", 5},
		{"medium margin", 25},
		{"high margin", 50},
		{"very high margin", 99},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &domain.CreateBudgetDimensionRequest{
				ParentType:          domain.BudgetParentOffer,
				ParentID:            offer.ID,
				CustomName:          tc.name,
				Cost:                100000,
				Revenue:             0,
				TargetMarginPercent: &tc.targetMargin,
				MarginOverride:      true,
			}

			dto, err := svc.Create(ctx, req)

			require.NoError(t, err)
			assert.NotNil(t, dto)
		})
	}
}

func TestValidateMarginOverride_Invalid(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	testCases := []struct {
		name         string
		targetMargin *float64
	}{
		{"nil margin", nil},
		{"margin exactly 100", floatPtr(100)},
		{"margin above 100", floatPtr(110)},
		{"negative margin", floatPtr(-5)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &domain.CreateBudgetDimensionRequest{
				ParentType:          domain.BudgetParentOffer,
				ParentID:            offer.ID,
				CustomName:          tc.name,
				Cost:                100000,
				Revenue:             0,
				TargetMarginPercent: tc.targetMargin,
				MarginOverride:      true,
			}

			dto, err := svc.Create(ctx, req)

			assert.Nil(t, dto)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, service.ErrInvalidTargetMargin))
		})
	}
}

// =============================================================================
// DeleteByParent Tests
// =============================================================================

func TestDeleteByParent_Success(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Create dimensions
	for i := 0; i < 3; i++ {
		req := &domain.CreateBudgetDimensionRequest{
			ParentType: domain.BudgetParentOffer,
			ParentID:   offer.ID,
			CustomName: "Item",
			Cost:       float64((i + 1) * 10000),
			Revenue:    float64((i + 1) * 15000),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// Delete all
	err := svc.DeleteByParent(ctx, domain.BudgetParentOffer, offer.ID)

	require.NoError(t, err)

	// Verify all gone
	dtos, err := svc.ListByParent(ctx, domain.BudgetParentOffer, offer.ID)
	require.NoError(t, err)
	assert.Len(t, dtos, 0)

	// Verify offer value reset
	updatedOffer, err := fixtures.offerRepo.GetByID(ctx, offer.ID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, updatedOffer.Value)
}

func TestDeleteByParent_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	offer := fixtures.createTestOffer(t, ctx)

	// Delete from empty parent should not error
	err := svc.DeleteByParent(ctx, domain.BudgetParentOffer, offer.ID)

	require.NoError(t, err)
}

// =============================================================================
// Project Parent Tests
// =============================================================================

func TestCreate_ProjectParent(t *testing.T) {
	db := setupTestDB(t)
	svc, fixtures := setupTestService(t, db)
	defer fixtures.cleanup(t)

	ctx := contextWithUser(context.Background())
	project := fixtures.createTestProject(t, ctx)

	req := &domain.CreateBudgetDimensionRequest{
		ParentType: domain.BudgetParentProject,
		ParentID:   project.ID,
		CustomName: "Project Budget Line",
		Cost:       100000,
		Revenue:    150000,
	}

	dto, err := svc.Create(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, domain.BudgetParentProject, dto.ParentType)
	assert.Equal(t, project.ID, dto.ParentID)

	// Verify project was updated
	updatedProject, err := fixtures.projectRepo.GetByID(ctx, project.ID)
	require.NoError(t, err)
	assert.Equal(t, 150000.0, updatedProject.Budget)
	assert.True(t, updatedProject.HasDetailedBudget)
}

// =============================================================================
// Helper Functions
// =============================================================================

func floatPtr(f float64) *float64 {
	return &f
}
