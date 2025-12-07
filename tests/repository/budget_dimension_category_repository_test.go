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
	// Clean up any leftover test data from previous runs BEFORE the test
	cleanupBudgetDimensionCategoryTestData(t, db)
	// Also clean up after the test
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

// createTestCategory creates a test category using raw SQL to handle boolean false values correctly
func createTestCategory(t *testing.T, db *gorm.DB, id, name string, companyID *domain.CompanyID, displayOrder int, isActive bool) *domain.BudgetDimensionCategory {
	description := "Test description for " + name

	var companyIDValue interface{} = nil
	if companyID != nil {
		companyIDValue = string(*companyID)
	}

	// Use raw SQL to ensure boolean false values are properly inserted
	// GORM's Create skips false/zero values even with Select, falling back to DB defaults
	err := db.Exec(`INSERT INTO budget_dimension_categories (id, company_id, name, description, display_order, is_active)
                    VALUES (?, ?, ?, ?, ?, ?)`,
		id, companyIDValue, name, description, displayOrder, isActive).Error
	require.NoError(t, err)

	// Return the category struct
	return &domain.BudgetDimensionCategory{
		ID:           id,
		CompanyID:    companyID,
		Name:         name,
		Description:  description,
		DisplayOrder: displayOrder,
		IsActive:     isActive,
	}
}

func TestBudgetDimensionCategoryRepository_Create(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	t.Run("create global category", func(t *testing.T) {
		category := &domain.BudgetDimensionCategory{
			ID:           "test_create_global",
			CompanyID:    nil, // Global category
			Name:         "Test Global Category",
			Description:  "A test global category",
			DisplayOrder: 100,
			IsActive:     true,
		}

		err := repo.Create(context.Background(), category)
		assert.NoError(t, err)

		// Verify it was created
		found, err := repo.GetByID(context.Background(), category.ID)
		assert.NoError(t, err)
		assert.Equal(t, category.Name, found.Name)
		assert.Nil(t, found.CompanyID)
	})

	t.Run("create company-specific category", func(t *testing.T) {
		companyID := domain.CompanyStalbygg
		category := &domain.BudgetDimensionCategory{
			ID:           "test_create_company",
			CompanyID:    &companyID,
			Name:         "Test Company Category",
			Description:  "A test company-specific category",
			DisplayOrder: 101,
			IsActive:     true,
		}

		err := repo.Create(context.Background(), category)
		assert.NoError(t, err)

		// Verify it was created
		found, err := repo.GetByID(context.Background(), category.ID)
		assert.NoError(t, err)
		assert.Equal(t, category.Name, found.Name)
		assert.NotNil(t, found.CompanyID)
		assert.Equal(t, domain.CompanyStalbygg, *found.CompanyID)
	})
}

func TestBudgetDimensionCategoryRepository_GetByID(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	t.Run("found", func(t *testing.T) {
		category := createTestCategory(t, db, "test_getbyid_found", "Test GetByID Found", nil, 1, true)

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

	// Create a global test category
	createTestCategory(t, db, "test_getbyname_global", "Test GetByName Category", nil, 1, true)

	// Create a company-specific test category
	companyID := domain.CompanyStalbygg
	createTestCategory(t, db, "test_getbyname_company", "Company Specific Category", &companyID, 2, true)

	t.Run("exact match - global", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "Test GetByName Category", nil)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname_global", found.ID)
	})

	t.Run("case insensitive - lowercase", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "test getbyname category", nil)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname_global", found.ID)
	})

	t.Run("with company filter - finds global", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "Test GetByName Category", &companyID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname_global", found.ID)
	})

	t.Run("with company filter - finds company specific", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "Company Specific Category", &companyID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "test_getbyname_company", found.ID)
	})

	t.Run("company category not found without company filter", func(t *testing.T) {
		// Searching without company filter should not find company-specific categories
		found, err := repo.GetByName(context.Background(), "Company Specific Category", nil)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByName(context.Background(), "Nonexistent Category Name", nil)
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestBudgetDimensionCategoryRepository_Update(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	category := createTestCategory(t, db, "test_update", "Original Name", nil, 1, true)

	// Update the category
	category.Name = "Updated Name"
	category.Description = "Updated description"
	category.DisplayOrder = 99

	err := repo.Update(context.Background(), category)
	assert.NoError(t, err)

	// Verify the update
	found, err := repo.GetByID(context.Background(), category.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "Updated description", found.Description)
	assert.Equal(t, 99, found.DisplayOrder)
}

func TestBudgetDimensionCategoryRepository_Delete(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	category := createTestCategory(t, db, "test_delete", "To Delete", nil, 1, true)

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

	// Create test categories
	companyID := domain.CompanyStalbygg
	createTestCategory(t, db, "test_list_global_a", "Global A", nil, 1, true)
	createTestCategory(t, db, "test_list_global_b", "Global B (Inactive)", nil, 2, false)
	createTestCategory(t, db, "test_list_company_a", "Company A", &companyID, 3, true)
	createTestCategory(t, db, "test_list_company_b", "Company B (Inactive)", &companyID, 4, false)

	t.Run("list global only - all", func(t *testing.T) {
		categories, err := repo.List(context.Background(), nil, false)
		assert.NoError(t, err)

		// Filter to our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_global_a" || c.ID == "test_list_global_b" {
				testCats = append(testCats, c)
			}
		}
		assert.Len(t, testCats, 2)
	})

	t.Run("list global only - active only", func(t *testing.T) {
		categories, err := repo.List(context.Background(), nil, true)
		assert.NoError(t, err)

		// Filter to our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_global_a" || c.ID == "test_list_global_b" {
				testCats = append(testCats, c)
			}
		}
		assert.Len(t, testCats, 1)
		assert.Equal(t, "test_list_global_a", testCats[0].ID)
	})

	t.Run("list with company - includes global and company", func(t *testing.T) {
		categories, err := repo.List(context.Background(), &companyID, false)
		assert.NoError(t, err)

		// Filter to our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_global_a" || c.ID == "test_list_global_b" ||
				c.ID == "test_list_company_a" || c.ID == "test_list_company_b" {
				testCats = append(testCats, c)
			}
		}
		assert.Len(t, testCats, 4) // All 4 test categories
	})

	t.Run("list with company - active only", func(t *testing.T) {
		categories, err := repo.List(context.Background(), &companyID, true)
		assert.NoError(t, err)

		// Filter to our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_global_a" || c.ID == "test_list_global_b" ||
				c.ID == "test_list_company_a" || c.ID == "test_list_company_b" {
				testCats = append(testCats, c)
			}
		}
		assert.Len(t, testCats, 2) // Only active ones
		for _, c := range testCats {
			assert.True(t, c.IsActive)
		}
	})

	t.Run("ordered by display_order", func(t *testing.T) {
		categories, err := repo.List(context.Background(), &companyID, false)
		assert.NoError(t, err)

		// Filter to our test categories
		var testCats []domain.BudgetDimensionCategory
		for _, c := range categories {
			if c.ID == "test_list_global_a" || c.ID == "test_list_global_b" ||
				c.ID == "test_list_company_a" || c.ID == "test_list_company_b" {
				testCats = append(testCats, c)
			}
		}

		// Verify ordering
		for i := 1; i < len(testCats); i++ {
			assert.LessOrEqual(t, testCats[i-1].DisplayOrder, testCats[i].DisplayOrder)
		}
	})
}

func TestBudgetDimensionCategoryRepository_ListByCompanyOnly(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	companyID := domain.CompanyStalbygg
	createTestCategory(t, db, "test_componly_global", "Global Cat", nil, 1, true)
	createTestCategory(t, db, "test_componly_company", "Company Cat", &companyID, 2, true)

	categories, err := repo.ListByCompanyOnly(context.Background(), companyID, false)
	assert.NoError(t, err)

	// Should only include company-specific, not global
	var testCats []domain.BudgetDimensionCategory
	for _, c := range categories {
		if c.ID == "test_componly_global" || c.ID == "test_componly_company" {
			testCats = append(testCats, c)
		}
	}
	assert.Len(t, testCats, 1)
	assert.Equal(t, "test_componly_company", testCats[0].ID)
}

func TestBudgetDimensionCategoryRepository_ListGlobalOnly(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	companyID := domain.CompanyStalbygg
	createTestCategory(t, db, "test_globalonly_global", "Global Cat", nil, 1, true)
	createTestCategory(t, db, "test_globalonly_company", "Company Cat", &companyID, 2, true)

	categories, err := repo.ListGlobalOnly(context.Background(), false)
	assert.NoError(t, err)

	// Should only include global, not company-specific
	var testCats []domain.BudgetDimensionCategory
	for _, c := range categories {
		if c.ID == "test_globalonly_global" || c.ID == "test_globalonly_company" {
			testCats = append(testCats, c)
		}
	}
	assert.Len(t, testCats, 1)
	assert.Equal(t, "test_globalonly_global", testCats[0].ID)
}

func TestBudgetDimensionCategoryRepository_GetUsageCount(t *testing.T) {
	db := setupBudgetDimensionCategoryTestDB(t)
	repo := repository.NewBudgetDimensionCategoryRepository(db)

	// Create a test category
	category := createTestCategory(t, db, "test_usage_count", "Usage Count Category", nil, 1, true)

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
	companyID := domain.CompanyStalbygg
	cat1 := createTestCategory(t, db, "test_usage_list_1", "Usage List Category 1", nil, 1, true)
	cat2 := createTestCategory(t, db, "test_usage_list_2", "Usage List Category 2", nil, 2, true)
	createTestCategory(t, db, "test_usage_list_3", "Usage List Category 3 (Inactive)", nil, 3, false)
	createTestCategory(t, db, "test_usage_list_company", "Company Usage Category", &companyID, 4, true)

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

	t.Run("list global with usage counts", func(t *testing.T) {
		results, err := repo.ListWithUsageCounts(context.Background(), nil, false)
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

	t.Run("list with company includes company categories", func(t *testing.T) {
		results, err := repo.ListWithUsageCounts(context.Background(), &companyID, false)
		assert.NoError(t, err)

		// Should include company category
		foundCompanyCat := false
		for _, r := range results {
			if r.ID == "test_usage_list_company" {
				foundCompanyCat = true
			}
		}
		assert.True(t, foundCompanyCat, "Company category should be in results")
	})

	t.Run("list active only excludes inactive", func(t *testing.T) {
		results, err := repo.ListWithUsageCounts(context.Background(), nil, true)
		assert.NoError(t, err)

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

	companyID := domain.CompanyStalbygg

	// Get initial counts
	initialGlobalAll, err := repo.Count(context.Background(), nil, false)
	require.NoError(t, err)
	initialGlobalActive, err := repo.Count(context.Background(), nil, true)
	require.NoError(t, err)
	initialCompanyAll, err := repo.Count(context.Background(), &companyID, false)
	require.NoError(t, err)
	initialCompanyActive, err := repo.Count(context.Background(), &companyID, true)
	require.NoError(t, err)

	// Create test categories
	createTestCategory(t, db, "test_count_1", "Count Category 1", nil, 1, true)         // Global active
	createTestCategory(t, db, "test_count_2", "Count Category 2", nil, 2, false)        // Global inactive
	createTestCategory(t, db, "test_count_3", "Count Category 3", &companyID, 3, true)  // Company active
	createTestCategory(t, db, "test_count_4", "Count Category 4", &companyID, 4, false) // Company inactive

	t.Run("count global all", func(t *testing.T) {
		count, err := repo.Count(context.Background(), nil, false)
		assert.NoError(t, err)
		assert.Equal(t, initialGlobalAll+2, count) // 2 global categories added
	})

	t.Run("count global active only", func(t *testing.T) {
		count, err := repo.Count(context.Background(), nil, true)
		assert.NoError(t, err)
		assert.Equal(t, initialGlobalActive+1, count) // 1 active global category added
	})

	t.Run("count with company - includes both", func(t *testing.T) {
		count, err := repo.Count(context.Background(), &companyID, false)
		assert.NoError(t, err)
		assert.Equal(t, initialCompanyAll+4, count) // 2 global + 2 company = 4 added
	})

	t.Run("count with company - active only", func(t *testing.T) {
		count, err := repo.Count(context.Background(), &companyID, true)
		assert.NoError(t, err)
		assert.Equal(t, initialCompanyActive+2, count) // 1 active global + 1 active company = 2 added
	})
}
