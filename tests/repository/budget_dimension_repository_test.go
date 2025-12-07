package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupBudgetDimensionTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	// Clean up any leftover test data from previous runs BEFORE the test
	cleanupBudgetDimensionTestData(t, db)
	// Also clean up after the test
	t.Cleanup(func() {
		cleanupBudgetDimensionTestData(t, db)
	})
	return db
}

// cleanupBudgetDimensionTestData removes test data
func cleanupBudgetDimensionTestData(t *testing.T, db *gorm.DB) {
	// Delete budget dimensions first (referential integrity)
	err := db.Exec("DELETE FROM budget_dimensions WHERE custom_name LIKE 'Test%' OR custom_name LIKE 'test%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test budget dimensions: %v", err)
	}

	// Delete test offers
	err = db.Exec("DELETE FROM offers WHERE title LIKE 'Test%' OR title LIKE 'test%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test offers: %v", err)
	}

	// Delete test projects
	err = db.Exec("DELETE FROM projects WHERE name LIKE 'Test%' OR name LIKE 'test%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test projects: %v", err)
	}

	// Delete test customers
	err = db.Exec("DELETE FROM customers WHERE name LIKE 'Test%' OR name LIKE 'test%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test customers: %v", err)
	}
}

// createTestOffer creates a test offer and returns it
func createTestOffer(t *testing.T, db *gorm.DB, title string) *domain.Offer {
	customer := testutil.CreateTestCustomer(t, db, "Test Customer for "+title)

	offer := &domain.Offer{
		Title:             title,
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             domain.OfferPhaseDraft,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: "test-user",
	}
	err := db.Create(offer).Error
	require.NoError(t, err)
	return offer
}

// createTestProject creates a test project and returns it
func createTestProject(t *testing.T, db *gorm.DB, name string) *domain.Project {
	customer := testutil.CreateTestCustomer(t, db, "Test Customer for "+name)

	project := &domain.Project{
		Name:        name,
		CustomerID:  customer.ID,
		CompanyID:   domain.CompanyStalbygg,
		Status:      domain.ProjectStatusActive,
		StartDate:   time.Now(),
		ManagerID:   "test-manager",
		ManagerName: "Test Manager",
	}
	err := db.Create(project).Error
	require.NoError(t, err)
	return project
}

// createTestDimension creates a budget dimension using raw SQL to avoid GORM issues
func createTestDimension(t *testing.T, db *gorm.DB, parentType domain.BudgetParentType, parentID uuid.UUID, customName string, cost, revenue float64, displayOrder int) *domain.BudgetDimension {
	id := uuid.New()
	err := db.Exec(`INSERT INTO budget_dimensions (id, parent_type, parent_id, custom_name, cost, revenue, display_order)
                    VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, parentType, parentID, customName, cost, revenue, displayOrder).Error
	require.NoError(t, err)

	// Fetch the created dimension to get the computed margin_percent
	var dimension domain.BudgetDimension
	err = db.Preload("Category").First(&dimension, "id = ?", id).Error
	require.NoError(t, err)

	return &dimension
}

// createTestDimensionWithCategory creates a budget dimension with a category reference
func createTestDimensionWithCategory(t *testing.T, db *gorm.DB, parentType domain.BudgetParentType, parentID uuid.UUID, categoryID string, cost, revenue float64, displayOrder int) *domain.BudgetDimension {
	id := uuid.New()
	err := db.Exec(`INSERT INTO budget_dimensions (id, parent_type, parent_id, category_id, cost, revenue, display_order)
                    VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, parentType, parentID, categoryID, cost, revenue, displayOrder).Error
	require.NoError(t, err)

	// Fetch the created dimension to get the computed margin_percent
	var dimension domain.BudgetDimension
	err = db.Preload("Category").First(&dimension, "id = ?", id).Error
	require.NoError(t, err)

	return &dimension
}

// createTestDimensionWithMarginOverride creates a dimension with margin override
func createTestDimensionWithMarginOverride(t *testing.T, db *gorm.DB, parentType domain.BudgetParentType, parentID uuid.UUID, customName string, cost float64, targetMarginPercent float64, displayOrder int) *domain.BudgetDimension {
	id := uuid.New()
	err := db.Exec(`INSERT INTO budget_dimensions (id, parent_type, parent_id, custom_name, cost, target_margin_percent, margin_override, display_order)
                    VALUES (?, ?, ?, ?, ?, ?, true, ?)`,
		id, parentType, parentID, customName, cost, targetMarginPercent, displayOrder).Error
	require.NoError(t, err)

	// Fetch the created dimension to get the computed revenue and margin_percent
	var dimension domain.BudgetDimension
	err = db.Preload("Category").First(&dimension, "id = ?", id).Error
	require.NoError(t, err)

	return &dimension
}

func TestBudgetDimensionRepository_Create(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	t.Run("create with custom name", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Create Custom")

		dimension := &domain.BudgetDimension{
			ParentType:   domain.BudgetParentOffer,
			ParentID:     offer.ID,
			CustomName:   "Test Custom Dimension",
			Cost:         1000.00,
			Revenue:      1500.00,
			DisplayOrder: 0,
		}

		err := repo.Create(context.Background(), dimension)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, dimension.ID)

		// Verify it was created
		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Test Custom Dimension", found.CustomName)
		assert.Equal(t, 1000.00, found.Cost)
		assert.Equal(t, 1500.00, found.Revenue)
	})

	t.Run("create with category", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Create Category")

		// Use an existing category from the seed data
		categoryID := "steel_structure"
		dimension := &domain.BudgetDimension{
			ParentType:   domain.BudgetParentOffer,
			ParentID:     offer.ID,
			CategoryID:   &categoryID,
			Cost:         2000.00,
			Revenue:      3000.00,
			DisplayOrder: 0,
		}

		err := repo.Create(context.Background(), dimension)
		assert.NoError(t, err)

		// Verify category is preloaded on GetByID
		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found.CategoryID)
		assert.Equal(t, categoryID, *found.CategoryID)
		assert.NotNil(t, found.Category)
	})

	t.Run("create with margin override", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Margin Override")

		targetMargin := 25.0
		dimension := &domain.BudgetDimension{
			ParentType:          domain.BudgetParentOffer,
			ParentID:            offer.ID,
			CustomName:          "Test Margin Override",
			Cost:                1000.00,
			TargetMarginPercent: &targetMargin,
			MarginOverride:      true,
			DisplayOrder:        0,
		}

		err := repo.Create(context.Background(), dimension)
		assert.NoError(t, err)

		// Verify revenue was calculated by DB trigger
		// Revenue = Cost / (1 - margin/100) = 1000 / 0.75 = 1333.33...
		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.InDelta(t, 1333.33, found.Revenue, 0.01)
		assert.InDelta(t, 25.0, found.MarginPercent, 0.01)
	})

	t.Run("create for project parent", func(t *testing.T) {
		project := createTestProject(t, db, "Test Project Budget")

		dimension := &domain.BudgetDimension{
			ParentType:   domain.BudgetParentProject,
			ParentID:     project.ID,
			CustomName:   "Test Project Dimension",
			Cost:         5000.00,
			Revenue:      7500.00,
			DisplayOrder: 0,
		}

		err := repo.Create(context.Background(), dimension)
		assert.NoError(t, err)

		// Verify it was created with project parent
		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.BudgetParentProject, found.ParentType)
		assert.Equal(t, project.ID, found.ParentID)
	})
}

func TestBudgetDimensionRepository_GetByID(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	t.Run("found", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer GetByID")
		dimension := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test GetByID Dimension", 1000, 1500, 0)

		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, dimension.ID, found.ID)
		assert.Equal(t, "Test GetByID Dimension", found.CustomName)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("with category preloaded", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer GetByID Category")
		dimension := createTestDimensionWithCategory(t, db, domain.BudgetParentOffer, offer.ID, "steel_structure", 2000, 3000, 0)

		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found.Category)
		assert.Equal(t, "steel_structure", found.Category.ID)
	})
}

func TestBudgetDimensionRepository_Update(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Update")
	dimension := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Update Original", 1000, 1500, 0)

	t.Run("update cost and revenue", func(t *testing.T) {
		dimension.Cost = 2000.00
		dimension.Revenue = 3000.00
		dimension.CustomName = "Test Update Modified"

		err := repo.Update(context.Background(), dimension)
		assert.NoError(t, err)

		// Verify the update
		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2000.00, found.Cost)
		assert.Equal(t, 3000.00, found.Revenue)
		assert.Equal(t, "Test Update Modified", found.CustomName)
	})

	t.Run("update triggers margin recalculation", func(t *testing.T) {
		// Update to values where margin can be easily calculated
		dimension.Cost = 800.00
		dimension.Revenue = 1000.00 // Should give 20% margin

		err := repo.Update(context.Background(), dimension)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.InDelta(t, 20.0, found.MarginPercent, 0.01)
	})
}

func TestBudgetDimensionRepository_Delete(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Delete")
	dimension := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Delete Dimension", 1000, 1500, 0)

	err := repo.Delete(context.Background(), dimension.ID)
	assert.NoError(t, err)

	// Verify it was deleted
	found, err := repo.GetByID(context.Background(), dimension.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBudgetDimensionRepository_GetByParent(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	t.Run("for offer", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer GetByParent")

		// Create multiple dimensions with different display orders
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Dim C", 3000, 4500, 2)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Dim A", 1000, 1500, 0)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Dim B", 2000, 3000, 1)

		dimensions, err := repo.GetByParent(context.Background(), domain.BudgetParentOffer, offer.ID)
		assert.NoError(t, err)
		assert.Len(t, dimensions, 3)

		// Verify ordering by display_order
		assert.Equal(t, "Test Dim A", dimensions[0].CustomName)
		assert.Equal(t, "Test Dim B", dimensions[1].CustomName)
		assert.Equal(t, "Test Dim C", dimensions[2].CustomName)
	})

	t.Run("for project", func(t *testing.T) {
		project := createTestProject(t, db, "Test Project GetByParent")

		createTestDimension(t, db, domain.BudgetParentProject, project.ID, "Test Project Dim 1", 5000, 7500, 0)
		createTestDimension(t, db, domain.BudgetParentProject, project.ID, "Test Project Dim 2", 3000, 4500, 1)

		dimensions, err := repo.GetByParent(context.Background(), domain.BudgetParentProject, project.ID)
		assert.NoError(t, err)
		assert.Len(t, dimensions, 2)
		assert.Equal(t, domain.BudgetParentProject, dimensions[0].ParentType)
	})

	t.Run("empty result", func(t *testing.T) {
		dimensions, err := repo.GetByParent(context.Background(), domain.BudgetParentOffer, uuid.New())
		assert.NoError(t, err)
		assert.Len(t, dimensions, 0)
	})
}

func TestBudgetDimensionRepository_GetByParentPaginated(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Paginated")

	// Create 5 dimensions
	for i := 0; i < 5; i++ {
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Paginated Dim", float64(1000*(i+1)), float64(1500*(i+1)), i)
	}

	t.Run("first page", func(t *testing.T) {
		dimensions, total, err := repo.GetByParentPaginated(context.Background(), domain.BudgetParentOffer, offer.ID, 1, 2)
		assert.NoError(t, err)
		assert.Len(t, dimensions, 2)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, 0, dimensions[0].DisplayOrder)
		assert.Equal(t, 1, dimensions[1].DisplayOrder)
	})

	t.Run("second page", func(t *testing.T) {
		dimensions, total, err := repo.GetByParentPaginated(context.Background(), domain.BudgetParentOffer, offer.ID, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, dimensions, 2)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, 2, dimensions[0].DisplayOrder)
		assert.Equal(t, 3, dimensions[1].DisplayOrder)
	})

	t.Run("last page partial", func(t *testing.T) {
		dimensions, total, err := repo.GetByParentPaginated(context.Background(), domain.BudgetParentOffer, offer.ID, 3, 2)
		assert.NoError(t, err)
		assert.Len(t, dimensions, 1)
		assert.Equal(t, int64(5), total)
		assert.Equal(t, 4, dimensions[0].DisplayOrder)
	})
}

func TestBudgetDimensionRepository_DeleteByParent(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer DeleteByParent")

	// Create multiple dimensions
	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Delete Parent 1", 1000, 1500, 0)
	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Delete Parent 2", 2000, 3000, 1)
	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Delete Parent 3", 3000, 4500, 2)

	// Verify they exist
	dimensions, err := repo.GetByParent(context.Background(), domain.BudgetParentOffer, offer.ID)
	assert.NoError(t, err)
	assert.Len(t, dimensions, 3)

	// Delete all
	err = repo.DeleteByParent(context.Background(), domain.BudgetParentOffer, offer.ID)
	assert.NoError(t, err)

	// Verify they're gone
	dimensions, err = repo.GetByParent(context.Background(), domain.BudgetParentOffer, offer.ID)
	assert.NoError(t, err)
	assert.Len(t, dimensions, 0)
}

func TestBudgetDimensionRepository_MarginCalculations(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	t.Run("manual revenue sets margin correctly", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Manual Revenue")

		// Cost=800, Revenue=1000 should give margin = ((1000-800)/1000)*100 = 20%
		dimension := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Manual Revenue", 800, 1000, 0)

		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.InDelta(t, 20.0, found.MarginPercent, 0.01)
	})

	t.Run("margin override calculates revenue from cost", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Margin Override Calc")

		// Cost=1000, Target margin=20%
		// Revenue = Cost / (1 - margin/100) = 1000 / 0.80 = 1250
		dimension := createTestDimensionWithMarginOverride(t, db, domain.BudgetParentOffer, offer.ID, "Test Override Calc", 1000, 20.0, 0)

		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.InDelta(t, 1250.0, found.Revenue, 0.01)
		assert.InDelta(t, 20.0, found.MarginPercent, 0.01)
	})

	t.Run("zero revenue gives zero margin", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Zero Revenue")

		dimension := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Zero Revenue", 1000, 0, 0)

		found, err := repo.GetByID(context.Background(), dimension.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, found.MarginPercent)
	})
}

func TestBudgetDimensionRepository_GetBudgetSummary(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	t.Run("with multiple dimensions", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Summary")

		// Create dimensions with known values
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Summary 1", 1000, 1500, 0)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Summary 2", 2000, 3000, 1)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Summary 3", 3000, 4500, 2)
		// Totals: Cost=6000, Revenue=9000

		summary, err := repo.GetBudgetSummary(context.Background(), domain.BudgetParentOffer, offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 6000.0, summary.TotalCost)
		assert.Equal(t, 9000.0, summary.TotalRevenue)
		assert.Equal(t, 3000.0, summary.TotalMargin)
		assert.InDelta(t, 33.33, summary.MarginPercent, 0.01) // (9000-6000)/9000*100
		assert.Equal(t, 3, summary.DimensionCount)
	})

	t.Run("empty parent", func(t *testing.T) {
		summary, err := repo.GetBudgetSummary(context.Background(), domain.BudgetParentOffer, uuid.New())
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 0.0, summary.TotalCost)
		assert.Equal(t, 0.0, summary.TotalRevenue)
		assert.Equal(t, 0.0, summary.TotalMargin)
		assert.Equal(t, 0.0, summary.MarginPercent)
		assert.Equal(t, 0, summary.DimensionCount)
	})
}

func TestBudgetDimensionRepository_GetTotalCost(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Total Cost")

	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Total Cost 1", 1000, 1500, 0)
	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Total Cost 2", 2000, 3000, 1)

	total, err := repo.GetTotalCost(context.Background(), domain.BudgetParentOffer, offer.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3000.0, total)
}

func TestBudgetDimensionRepository_GetTotalRevenue(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Total Revenue")

	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Total Revenue 1", 1000, 1500, 0)
	createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Total Revenue 2", 2000, 3000, 1)

	total, err := repo.GetTotalRevenue(context.Background(), domain.BudgetParentOffer, offer.ID)
	assert.NoError(t, err)
	assert.Equal(t, 4500.0, total)
}

func TestBudgetDimensionRepository_ReorderDimensions(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Reorder")

	// Create dimensions in order A, B, C
	dimA := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Reorder A", 1000, 1500, 0)
	dimB := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Reorder B", 2000, 3000, 1)
	dimC := createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Reorder C", 3000, 4500, 2)

	t.Run("reorder to C, A, B", func(t *testing.T) {
		newOrder := []uuid.UUID{dimC.ID, dimA.ID, dimB.ID}

		err := repo.ReorderDimensions(context.Background(), domain.BudgetParentOffer, offer.ID, newOrder)
		assert.NoError(t, err)

		// Verify new order
		dimensions, err := repo.GetByParent(context.Background(), domain.BudgetParentOffer, offer.ID)
		assert.NoError(t, err)
		assert.Len(t, dimensions, 3)
		assert.Equal(t, dimC.ID, dimensions[0].ID)
		assert.Equal(t, dimA.ID, dimensions[1].ID)
		assert.Equal(t, dimB.ID, dimensions[2].ID)
	})

	t.Run("error on invalid dimension ID", func(t *testing.T) {
		invalidOrder := []uuid.UUID{uuid.New(), dimA.ID, dimB.ID}

		err := repo.ReorderDimensions(context.Background(), domain.BudgetParentOffer, offer.ID, invalidOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBudgetDimensionRepository_GetMaxDisplayOrder(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	t.Run("with dimensions", func(t *testing.T) {
		offer := createTestOffer(t, db, "Test Offer Max Order")

		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Max Order 1", 1000, 1500, 0)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Max Order 2", 2000, 3000, 5)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Max Order 3", 3000, 4500, 3)

		maxOrder, err := repo.GetMaxDisplayOrder(context.Background(), domain.BudgetParentOffer, offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 5, maxOrder)
	})

	t.Run("empty parent returns -1", func(t *testing.T) {
		maxOrder, err := repo.GetMaxDisplayOrder(context.Background(), domain.BudgetParentOffer, uuid.New())
		assert.NoError(t, err)
		assert.Equal(t, -1, maxOrder)
	})
}

func TestBudgetDimensionRepository_Count(t *testing.T) {
	db := setupBudgetDimensionTestDB(t)
	repo := repository.NewBudgetDimensionRepository(db)

	offer := createTestOffer(t, db, "Test Offer Count")

	t.Run("empty initially", func(t *testing.T) {
		count, err := repo.Count(context.Background(), domain.BudgetParentOffer, offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("after adding dimensions", func(t *testing.T) {
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Count 1", 1000, 1500, 0)
		createTestDimension(t, db, domain.BudgetParentOffer, offer.ID, "Test Count 2", 2000, 3000, 1)

		count, err := repo.Count(context.Background(), domain.BudgetParentOffer, offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
