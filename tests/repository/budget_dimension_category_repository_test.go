package repository_test

import (
	"context"
	"testing"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupBudgetDimensionCategoryTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		cleanupBudgetDimensionCategoryTestData(t, db)
	})
	return db
}

// cleanupBudgetDimensionCategoryTestData removes test categories (prefixed with test_)
func cleanupBudgetDimensionCategoryTestData(t *testing.T, db *gorm.DB) {
	// First delete any budget dimensions referencing test categories
	err := db.Exec("DELETE FROM budget_dimensions WHERE category_id LIKE 'test_%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test budget dimensions: %v", err)
	}

	// Then delete test categories
	err = db.Exec("DELETE FROM budget_dimension_categories WHERE id LIKE 'test_%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test budget dimension categories: %v", err)
	}
}

func createTestCategory(t *testing.T, db *gorm.DB, id, name string, displayOrder int, isActive bool) *domain.BudgetDimensionCategory {
	category := &domain.BudgetDimensionCategory{
		ID:           id,
		Name:         name,
		Description:  "Test description for " + name,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
	}
	err := db.Create(category).Error
	require.NoError(t, err)
	return category
}

func TestBudgetDimensionCategoryRepository_Create(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	category := &domain.BudgetDimensionCategory{
		ID:           "test_create_cat",
		Name:         "Test Create Category",
		Description:  "A test category for creation",
		DisplayOrder: 100,
		IsActive:     true,
	}

	err := repo.Create(context.Background(), category)
	assert.NoError(t, err)

	// Verify it was created
	found, err := repo.GetByID(context.Background(), category.ID)
	assert.NoError(t, err)
	assert.Equal(t, category.Name, found.Name)
	assert.Equal(t, category.Description, found.Description)
	assert.Equal(t, category.DisplayOrder, found.DisplayOrder)
	assert.True(t, found.IsActive)
}

func TestBudgetDimensionCategoryRepository_GetByID(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	t.Run("found", func(t *testing.T) {
		category := createTestCategory(t, db, "test_getbyid_found", "Test GetByID Found", 1, true)

		found, err := repo.GetByID(context.Background(), category.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, category.ID, found.ID)
		assert.Equal(t, category.Name, found.Name)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), "test_nonexistent_id")
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestBudgetDimensionCategoryRepository_GetByName(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	// Create a test category
	createTestCategory(t, db, "test_getbyname", "Test GetByName Category", 1, true)

	t.Run("exact match", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "Test GetByName Category")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname", found.ID)
	})

	t.Run("case insensitive - lowercase", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "test getbyname category")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname", found.ID)
	})

	t.Run("case insensitive - uppercase", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "TEST GETBYNAME CATEGORY")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname", found.ID)
	})

	t.Run("case insensitive - mixed case", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "TeSt GeTbYnAmE cAtEgOrY")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname", found.ID)
	})

	t.Run("with whitespace", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "  Test GetByName Category  ")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname", found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "Nonexistent Category Name")
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestBudgetDimensionCategoryRepository_Update(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	category := createTestCategory(t, db, "test_update", "Original Name", 1, true)

	// Update the category
	category.Name = "Updated Name"
	category.Description = "Updated description"
	category.DisplayOrder = 50
	category.IsActive = false

	err := repo.Update(context.Background(), category)
	assert.NoError(t, err)

	// Verify the update
	found, err := repo.GetByID(context.Background(), category.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "Updated description", found.Description)
	assert.Equal(t, 50, found.DisplayOrder)
	assert.False(t, found.IsActive)
}

func TestBudgetDimensionCategoryRepository_Delete(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	category := createTestCategory(t, db, "test_delete", "Delete Me", 1, true)

	// Delete the category
	err := repo.Delete(context.Background(), category.ID)
	assert.NoError(t, err)

	// Verify it was deleted
	found, err := repo.GetByID(context.Background(), category.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBudgetDimensionCategoryRepository_List(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	// Create test categories with different display orders and active states
	createTestCategory(t, db, "test_list_c", "Category C", 3, true)
	createTestCategory(t, db, "test_list_a", "Category A", 1, true)
	createTestCategory(t, db, "test_list_b", "Category B", 2, false) // Inactive
	createTestCategory(t, db, "test_list_d", "Category D", 4, true)

	t.Run("list all", func(t *testing.T) {
		categories, err := repo.List(context.Background(), false)
		assert.NoError(t, err)

		// Find our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_a" || c.ID == "test_list_b" || c.ID == "test_list_c" || c.ID == "test_list_d" {
				testCats = append(testCats, c)
			}
		}

		assert.Len(t, testCats, 4)
	})

	t.Run("list active only", func(t *testing.T) {
		categories, err := repo.List(context.Background(), true)
		assert.NoError(t, err)

		// Find our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_a" || c.ID == "test_list_b" || c.ID == "test_list_c" || c.ID == "test_list_d" {
				testCats = append(testCats, c)
			}
		}

		assert.Len(t, testCats, 3) // Should not include the inactive one
		for _, c := range testCats {
			assert.NotEqual(t, "test_list_b", c.ID)
		}
	})

	t.Run("ordered by display_order", func(t *testing.T) {
		categories, err := repo.List(context.Background(), false)
		assert.NoError(t, err)

		// Filter to only our test categories and verify order
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_a" || c.ID == "test_list_b" || c.ID == "test_list_c" || c.ID == "test_list_d" {
				testCats = append(testCats, c)
			}
		}

		// Categories should be ordered by display_order
		assert.Equal(t, "test_list_a", testCats[0].ID) // display_order 1
		assert.Equal(t, "test_list_b", testCats[1].ID) // display_order 2
		assert.Equal(t, "test_list_c", testCats[2].ID) // display_order 3
		assert.Equal(t, "test_list_d", testCats[3].ID) // display_order 4
	})
}

func TestBudgetDimensionCategoryRepository_GetUsageCount(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	// Create a test category
	category := createTestCategory(t, db, "test_usage_count", "Usage Count Category", 1, true)

	t.Run("no usage", func(t *testing.T) {
		count, err := repo.GetUsageCount(context.Background(), category.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("with usage", func(t *testing.T) {
		// Create a test customer and offer first
		customer := testutil.CreateTestCustomer(t, db, "Usage Test Customer")

		offer := &domain.Offer{
			Title:             "Test Offer for Usage",
			CustomerID:        customer.ID,
			CustomerName:      customer.Name,
			CompanyID:         domain.CompanyStalbygg,
			Phase:             domain.OfferPhaseDraft,
			Status:            domain.OfferStatusActive,
			ResponsibleUserID: "test-user",
		}
		err := db.Create(offer).Error
		require.NoError(t, err)

		// Create budget dimensions using this category
		categoryID := category.ID
		for i := 0; i < 3; i++ {
			dimension := &domain.BudgetDimension{
				ParentType:   domain.BudgetParentOffer,
				ParentID:     offer.ID,
				CategoryID:   &categoryID,
				Cost:         1000,
				Revenue:      1500,
				DisplayOrder: i,
			}
			err := db.Create(dimension).Error
			require.NoError(t, err)
		}

		count, err := repo.GetUsageCount(context.Background(), category.ID)
		assert.NoError(t, err)
		assert.Equal(t, 3, count)
	})
}

func TestBudgetDimensionCategoryRepository_ListWithUsageCounts(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	// Create test categories
	cat1 := createTestCategory(t, db, "test_usage_list_1", "Usage List Category 1", 1, true)
	cat2 := createTestCategory(t, db, "test_usage_list_2", "Usage List Category 2", 2, true)
	createTestCategory(t, db, "test_usage_list_3", "Usage List Category 3 (Inactive)", 3, false)

	// Create a test customer and offer
	customer := testutil.CreateTestCustomer(t, db, "Usage List Test Customer")

	offer := &domain.Offer{
		Title:             "Test Offer for Usage List",
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             domain.OfferPhaseDraft,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: "test-user",
	}
	err := db.Create(offer).Error
	require.NoError(t, err)

	// Create budget dimensions - 2 for cat1, 1 for cat2
	cat1ID := cat1.ID
	cat2ID := cat2.ID

	for i := 0; i < 2; i++ {
		dimension := &domain.BudgetDimension{
			ParentType:   domain.BudgetParentOffer,
			ParentID:     offer.ID,
			CategoryID:   &cat1ID,
			Cost:         1000,
			Revenue:      1500,
			DisplayOrder: i,
		}
		err := db.Create(dimension).Error
		require.NoError(t, err)
	}

	dimension := &domain.BudgetDimension{
		ParentType:   domain.BudgetParentOffer,
		ParentID:     offer.ID,
		CategoryID:   &cat2ID,
		Cost:         2000,
		Revenue:      2500,
		DisplayOrder: 2,
	}
	err = db.Create(dimension).Error
	require.NoError(t, err)

	t.Run("list all with usage counts", func(t *testing.T) {
		results, err := repo.ListWithUsageCounts(context.Background(), false)
		assert.NoError(t, err)

		// Find our test categories
		var cat1Result, cat2Result, cat3Result *repository.CategoryWithUsage
		for i := range results {
			switch results[i].ID {
			case "test_usage_list_1":
				cat1Result = &results[i]
			case "test_usage_list_2":
				cat2Result = &results[i]
			case "test_usage_list_3":
				cat3Result = &results[i]
			}
		}

		require.NotNil(t, cat1Result)
		require.NotNil(t, cat2Result)
		require.NotNil(t, cat3Result)

		assert.Equal(t, 2, cat1Result.UsageCount)
		assert.Equal(t, 1, cat2Result.UsageCount)
		assert.Equal(t, 0, cat3Result.UsageCount)
	})

	t.Run("list active only with usage counts", func(t *testing.T) {
		results, err := repo.ListWithUsageCounts(context.Background(), true)
		assert.NoError(t, err)

		// Find our test categories
		foundInactive := false
		for _, r := range results {
			if r.ID == "test_usage_list_3" {
				foundInactive = true
			}
		}
		assert.False(t, foundInactive, "Inactive category should not be in active-only list")
	})
}

func TestBudgetDimensionCategoryRepository_Count(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	// Get initial count (may include seed data)
	initialAll, err := repo.Count(context.Background(), false)
	require.NoError(t, err)
	initialActive, err := repo.Count(context.Background(), true)
	require.NoError(t, err)

	// Create test categories
	createTestCategory(t, db, "test_count_1", "Count Category 1", 1, true)
	createTestCategory(t, db, "test_count_2", "Count Category 2", 2, true)
	createTestCategory(t, db, "test_count_3", "Count Category 3 (Inactive)", 3, false)

	t.Run("count all", func(t *testing.T) {
		count, err := repo.Count(context.Background(), false)
		assert.NoError(t, err)
		assert.Equal(t, initialAll+3, count)
	})

	t.Run("count active only", func(t *testing.T) {
		count, err := repo.Count(context.Background(), true)
		assert.NoError(t, err)
		assert.Equal(t, initialActive+2, count) // 2 active categories added
	})
}
