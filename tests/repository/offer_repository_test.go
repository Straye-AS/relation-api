package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupOfferTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	// Clean up any leftover test data from previous runs BEFORE the test
	cleanupOfferTestData(t, db)
	// Also clean up after the test
	t.Cleanup(func() {
		cleanupOfferTestData(t, db)
	})
	return db
}

// cleanupOfferTestData removes test data
func cleanupOfferTestData(t *testing.T, db *gorm.DB) {
	// Delete in order to respect foreign key constraints
	err := db.Exec("DELETE FROM budget_dimensions WHERE parent_id IN (SELECT id FROM offers WHERE title LIKE 'Test%' OR title LIKE 'test%')").Error
	if err != nil {
		t.Logf("Note: Could not clean test budget dimensions: %v", err)
	}

	err = db.Exec("DELETE FROM files WHERE offer_id IN (SELECT id FROM offers WHERE title LIKE 'Test%' OR title LIKE 'test%')").Error
	if err != nil {
		t.Logf("Note: Could not clean test files: %v", err)
	}

	err = db.Exec("DELETE FROM offer_items WHERE offer_id IN (SELECT id FROM offers WHERE title LIKE 'Test%' OR title LIKE 'test%')").Error
	if err != nil {
		t.Logf("Note: Could not clean test offer items: %v", err)
	}

	err = db.Exec("DELETE FROM offers WHERE title LIKE 'Test%' OR title LIKE 'test%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test offers: %v", err)
	}

	err = db.Exec("DELETE FROM customers WHERE name LIKE 'Test%' OR name LIKE 'test%'").Error
	if err != nil {
		t.Logf("Note: Could not clean test customers: %v", err)
	}
}

// createOfferTestOffer creates a test offer and returns it
func createOfferTestOffer(t *testing.T, db *gorm.DB, title string, phase domain.OfferPhase, status domain.OfferStatus) *domain.Offer {
	customer := testutil.CreateTestCustomer(t, db, "Test Customer for "+title)

	offer := &domain.Offer{
		Title:             title,
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             phase,
		Status:            status,
		Value:             0,
		ResponsibleUserID: "test-user",
	}
	err := db.Create(offer).Error
	require.NoError(t, err)
	return offer
}

// createOfferTestFile creates a test file linked to an offer
func createOfferTestFile(t *testing.T, db *gorm.DB, offerID uuid.UUID, filename string) *domain.File {
	file := &domain.File{
		Filename:    filename,
		ContentType: "application/pdf",
		Size:        1024,
		StoragePath: "/test/path/" + filename,
		OfferID:     &offerID,
	}
	err := db.Create(file).Error
	require.NoError(t, err)
	return file
}

// createOfferTestBudgetItem creates a budget item for an offer
func createOfferTestBudgetItem(t *testing.T, db *gorm.DB, offerID uuid.UUID, name string, cost, revenue float64, displayOrder int) *domain.BudgetItem {
	id := uuid.New()
	// Calculate margin: margin = (revenue - cost) / revenue * 100 (when revenue > 0)
	margin := 0.0
	if revenue > 0 {
		margin = (revenue - cost) / revenue * 100
	}
	err := db.Exec(`INSERT INTO budget_items (id, parent_type, parent_id, name, expected_cost, expected_margin, display_order)
                    VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, domain.BudgetParentOffer, offerID, name, cost, margin, displayOrder).Error
	require.NoError(t, err)

	var item domain.BudgetItem
	err = db.First(&item, "id = ?", id).Error
	require.NoError(t, err)

	return &item
}

func TestOfferRepository_Create(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	customer := testutil.CreateTestCustomer(t, db, "Test Customer Create")

	t.Run("create offer successfully", func(t *testing.T) {
		offer := &domain.Offer{
			Title:             "Test Offer Create",
			CustomerID:        customer.ID,
			CustomerName:      customer.Name,
			CompanyID:         domain.CompanyStalbygg,
			Phase:             domain.OfferPhaseDraft,
			Status:            domain.OfferStatusActive,
			Value:             10000.00,
			ResponsibleUserID: "test-user",
			Description:       "Test description",
		}

		err := repo.Create(context.Background(), offer)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, offer.ID)

		// Verify it was created
		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Test Offer Create", found.Title)
		assert.Equal(t, domain.OfferPhaseDraft, found.Phase)
		assert.Equal(t, domain.OfferStatusActive, found.Status)
	})
}

func TestOfferRepository_GetByID(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("found with customer preloaded", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer GetByID", domain.OfferPhaseDraft, domain.OfferStatusActive)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, offer.ID, found.ID)
		assert.NotNil(t, found.Customer)
		assert.Equal(t, offer.CustomerID, found.Customer.ID)
	})

	t.Run("found with files preloaded", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer GetByID Files", domain.OfferPhaseDraft, domain.OfferStatusActive)
		createOfferTestFile(t, db, offer.ID, "test-file-1.pdf")
		createOfferTestFile(t, db, offer.ID, "test-file-2.pdf")

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Len(t, found.Files, 2)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestOfferRepository_GetByIDWithBudgetItems(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("returns offer with budget items", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer WithItems", domain.OfferPhaseDraft, domain.OfferStatusActive)

		// Create budget items
		createOfferTestBudgetItem(t, db, offer.ID, "Test Item 1", 1000, 1500, 0)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Item 2", 2000, 3000, 1)

		found, items, err := repo.GetByIDWithBudgetItems(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, offer.ID, found.ID)
		assert.Len(t, items, 2)

		// Verify ordering by display_order
		assert.Equal(t, "Test Item 1", items[0].Name)
		assert.Equal(t, "Test Item 2", items[1].Name)
	})

	t.Run("returns empty items for offer without any", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer NoItems", domain.OfferPhaseDraft, domain.OfferStatusActive)

		found, items, err := repo.GetByIDWithBudgetItems(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Len(t, items, 0)
	})

	t.Run("not found", func(t *testing.T) {
		found, items, err := repo.GetByIDWithBudgetItems(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Nil(t, items)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestOfferRepository_Update(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	offer := createOfferTestOffer(t, db, "Test Offer Update", domain.OfferPhaseDraft, domain.OfferStatusActive)

	t.Run("update offer fields", func(t *testing.T) {
		offer.Title = "Test Offer Updated"
		offer.Value = 25000.00
		offer.Description = "Updated description"

		err := repo.Update(context.Background(), offer)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Test Offer Updated", found.Title)
		assert.Equal(t, 25000.00, found.Value)
		assert.Equal(t, "Updated description", found.Description)
	})
}

func TestOfferRepository_Delete(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	offer := createOfferTestOffer(t, db, "Test Offer Delete", domain.OfferPhaseDraft, domain.OfferStatusActive)

	err := repo.Delete(context.Background(), offer.ID)
	assert.NoError(t, err)

	// Verify it was deleted
	found, err := repo.GetByID(context.Background(), offer.ID)
	assert.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestOfferRepository_List(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	// Create test offers with different phases and statuses
	createOfferTestOffer(t, db, "Test Offer List 1", domain.OfferPhaseDraft, domain.OfferStatusActive)
	createOfferTestOffer(t, db, "Test Offer List 2", domain.OfferPhaseSent, domain.OfferStatusActive)
	createOfferTestOffer(t, db, "Test Offer List 3", domain.OfferPhaseDraft, domain.OfferStatusInactive)

	t.Run("list all", func(t *testing.T) {
		offers, total, err := repo.List(context.Background(), 1, 10, nil, nil, nil)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(3))
		assert.GreaterOrEqual(t, len(offers), 3)
	})

	t.Run("filter by phase", func(t *testing.T) {
		phase := domain.OfferPhaseDraft
		offers, total, err := repo.List(context.Background(), 1, 10, nil, nil, &phase)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))

		for _, offer := range offers {
			assert.Equal(t, domain.OfferPhaseDraft, offer.Phase)
		}
	})
}

func TestOfferRepository_ListWithFilters(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	// Create test offers with different phases and statuses
	createOfferTestOffer(t, db, "Test Offer Filter 1", domain.OfferPhaseDraft, domain.OfferStatusActive)
	createOfferTestOffer(t, db, "Test Offer Filter 2", domain.OfferPhaseSent, domain.OfferStatusActive)
	createOfferTestOffer(t, db, "Test Offer Filter 3", domain.OfferPhaseDraft, domain.OfferStatusInactive)
	createOfferTestOffer(t, db, "Test Offer Filter 4", domain.OfferPhaseWon, domain.OfferStatusArchived)

	t.Run("filter by status active", func(t *testing.T) {
		status := domain.OfferStatusActive
		filters := &repository.OfferFilters{Status: &status}
		offers, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.DefaultSortConfig())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))

		for _, offer := range offers {
			assert.Equal(t, domain.OfferStatusActive, offer.Status)
		}
	})

	t.Run("filter by status inactive", func(t *testing.T) {
		status := domain.OfferStatusInactive
		filters := &repository.OfferFilters{Status: &status}
		offers, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.DefaultSortConfig())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))

		for _, offer := range offers {
			assert.Equal(t, domain.OfferStatusInactive, offer.Status)
		}
	})

	t.Run("filter by status archived", func(t *testing.T) {
		status := domain.OfferStatusArchived
		filters := &repository.OfferFilters{Status: &status}
		offers, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.DefaultSortConfig())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))

		for _, offer := range offers {
			assert.Equal(t, domain.OfferStatusArchived, offer.Status)
		}
	})

	t.Run("filter by phase and status", func(t *testing.T) {
		phase := domain.OfferPhaseDraft
		status := domain.OfferStatusActive
		filters := &repository.OfferFilters{Phase: &phase, Status: &status}
		offers, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.DefaultSortConfig())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))

		for _, offer := range offers {
			assert.Equal(t, domain.OfferPhaseDraft, offer.Phase)
			assert.Equal(t, domain.OfferStatusActive, offer.Status)
		}
	})

	t.Run("pagination with default for invalid values", func(t *testing.T) {
		offers, total, err := repo.ListWithFilters(context.Background(), 0, 0, nil, repository.DefaultSortConfig())
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(4))
		assert.LessOrEqual(t, len(offers), 20) // Default page size
	})

	t.Run("pagination respects max page size", func(t *testing.T) {
		// Create many offers
		for i := 0; i < 5; i++ {
			createOfferTestOffer(t, db, "Test Offer Pagination Extra", domain.OfferPhaseDraft, domain.OfferStatusActive)
		}

		// Request more than max page size
		offers, _, err := repo.ListWithFilters(context.Background(), 1, 500, nil, repository.DefaultSortConfig())
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(offers), 200) // Max page size
	})
}

func TestOfferRepository_UpdateStatus(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("update status to inactive", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Status Update 1", domain.OfferPhaseDraft, domain.OfferStatusActive)

		err := repo.UpdateStatus(context.Background(), offer.ID, domain.OfferStatusInactive)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferStatusInactive, found.Status)
		// Verify phase was not changed
		assert.Equal(t, domain.OfferPhaseDraft, found.Phase)
	})

	t.Run("update status to archived", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Status Update 2", domain.OfferPhaseWon, domain.OfferStatusActive)

		err := repo.UpdateStatus(context.Background(), offer.ID, domain.OfferStatusArchived)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferStatusArchived, found.Status)
	})

	t.Run("update status for non-existent offer", func(t *testing.T) {
		err := repo.UpdateStatus(context.Background(), uuid.New(), domain.OfferStatusInactive)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestOfferRepository_UpdatePhase(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("update phase to in_progress", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Phase Update 1", domain.OfferPhaseDraft, domain.OfferStatusActive)

		err := repo.UpdatePhase(context.Background(), offer.ID, domain.OfferPhaseInProgress)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseInProgress, found.Phase)
		// Verify status was not changed
		assert.Equal(t, domain.OfferStatusActive, found.Status)
	})

	t.Run("update phase to sent", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Phase Update 2", domain.OfferPhaseInProgress, domain.OfferStatusActive)

		err := repo.UpdatePhase(context.Background(), offer.ID, domain.OfferPhaseSent)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseSent, found.Phase)
	})

	t.Run("update phase to won", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Phase Update 3", domain.OfferPhaseSent, domain.OfferStatusActive)

		err := repo.UpdatePhase(context.Background(), offer.ID, domain.OfferPhaseWon)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseWon, found.Phase)
	})

	t.Run("update phase to lost", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Phase Update 4", domain.OfferPhaseSent, domain.OfferStatusActive)

		err := repo.UpdatePhase(context.Background(), offer.ID, domain.OfferPhaseLost)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseLost, found.Phase)
	})

	t.Run("update phase to expired", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Phase Update 5", domain.OfferPhaseSent, domain.OfferStatusActive)

		err := repo.UpdatePhase(context.Background(), offer.ID, domain.OfferPhaseExpired)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseExpired, found.Phase)
	})

	t.Run("update phase for non-existent offer", func(t *testing.T) {
		err := repo.UpdatePhase(context.Background(), uuid.New(), domain.OfferPhaseInProgress)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestOfferRepository_CalculateTotalsFromBudgetItems(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("calculate totals from multiple budget items", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Calculate 1", domain.OfferPhaseDraft, domain.OfferStatusActive)

		// Create budget items with known revenue values
		createOfferTestBudgetItem(t, db, offer.ID, "Test Calc Item 1", 1000, 1500, 0)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Calc Item 2", 2000, 3000, 1)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Calc Item 3", 3000, 4500, 2)
		// Total revenue: 1500 + 3000 + 4500 = 9000

		totalRevenue, err := repo.CalculateTotalsFromBudgetItems(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.InDelta(t, 9000.0, totalRevenue, 1.0) // Allow small delta due to computed fields

		// Verify the offer's Value was updated
		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.InDelta(t, 9000.0, found.Value, 1.0)
	})

	t.Run("calculate totals with no items returns zero", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Calculate 2", domain.OfferPhaseDraft, domain.OfferStatusActive)
		offer.Value = 5000 // Set initial value
		err := db.Save(offer).Error
		require.NoError(t, err)

		totalRevenue, err := repo.CalculateTotalsFromBudgetItems(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, totalRevenue)

		// Verify the offer's Value was updated to 0
		found, err := repo.GetByID(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, found.Value)
	})

	t.Run("calculate totals for non-existent offer", func(t *testing.T) {
		_, err := repo.CalculateTotalsFromBudgetItems(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestOfferRepository_GetBudgetItemsCount(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("count items for offer", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Item Count", domain.OfferPhaseDraft, domain.OfferStatusActive)

		createOfferTestBudgetItem(t, db, offer.ID, "Test Count Item 1", 1000, 1500, 0)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Count Item 2", 2000, 3000, 1)

		count, err := repo.GetBudgetItemsCount(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("count items for offer with none", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer No Items", domain.OfferPhaseDraft, domain.OfferStatusActive)

		count, err := repo.GetBudgetItemsCount(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestOfferRepository_GetBudgetSummary(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	t.Run("get budget summary with multiple items", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Summary", domain.OfferPhaseDraft, domain.OfferStatusActive)

		// Create items with known values (revenue computed from cost and margin)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Summary Item 1", 1000, 1500, 0)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Summary Item 2", 2000, 3000, 1)
		createOfferTestBudgetItem(t, db, offer.ID, "Test Summary Item 3", 3000, 4500, 2)
		// Totals: Cost=6000, Revenue=9000

		summary, err := repo.GetBudgetSummary(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.InDelta(t, 6000.0, summary.TotalCost, 1.0)
		assert.InDelta(t, 9000.0, summary.TotalRevenue, 1.0)
		assert.InDelta(t, 3000.0, summary.TotalProfit, 1.0)
		assert.InDelta(t, 33.33, summary.MarginPercent, 1.0) // (9000-6000)/9000*100
		assert.Equal(t, 3, summary.ItemCount)
	})

	t.Run("get budget summary for offer with no items", func(t *testing.T) {
		offer := createOfferTestOffer(t, db, "Test Offer Empty Summary", domain.OfferPhaseDraft, domain.OfferStatusActive)

		summary, err := repo.GetBudgetSummary(context.Background(), offer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 0.0, summary.TotalCost)
		assert.Equal(t, 0.0, summary.TotalRevenue)
		assert.Equal(t, 0.0, summary.TotalProfit)
		assert.Equal(t, 0.0, summary.MarginPercent)
		assert.Equal(t, 0, summary.ItemCount)
	})
}

func TestOfferRepository_GetItemsCount(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	offer := createOfferTestOffer(t, db, "Test Offer Items Count", domain.OfferPhaseDraft, domain.OfferStatusActive)

	// Create offer items
	item1 := &domain.OfferItem{
		OfferID:    offer.ID,
		Discipline: "Steel",
		Cost:       1000,
		Revenue:    1500,
		Margin:     50,
	}
	item2 := &domain.OfferItem{
		OfferID:    offer.ID,
		Discipline: "Roofing",
		Cost:       2000,
		Revenue:    3000,
		Margin:     50,
	}
	err := db.Create(item1).Error
	require.NoError(t, err)
	err = db.Create(item2).Error
	require.NoError(t, err)

	count, err := repo.GetItemsCount(context.Background(), offer.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestOfferRepository_GetFilesCount(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	offer := createOfferTestOffer(t, db, "Test Offer Files Count", domain.OfferPhaseDraft, domain.OfferStatusActive)

	createOfferTestFile(t, db, offer.ID, "file1.pdf")
	createOfferTestFile(t, db, offer.ID, "file2.pdf")
	createOfferTestFile(t, db, offer.ID, "file3.pdf")

	count, err := repo.GetFilesCount(context.Background(), offer.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestOfferRepository_GetTotalPipelineValue(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	// Create offers in different pipeline phases
	offer1 := createOfferTestOffer(t, db, "Test Pipeline 1", domain.OfferPhaseInProgress, domain.OfferStatusActive)
	offer1.Value = 10000
	err := db.Save(offer1).Error
	require.NoError(t, err)

	offer2 := createOfferTestOffer(t, db, "Test Pipeline 2", domain.OfferPhaseSent, domain.OfferStatusActive)
	offer2.Value = 20000
	err = db.Save(offer2).Error
	require.NoError(t, err)

	// These should NOT be included in pipeline value
	offer3 := createOfferTestOffer(t, db, "Test Pipeline 3", domain.OfferPhaseDraft, domain.OfferStatusActive)
	offer3.Value = 5000
	err = db.Save(offer3).Error
	require.NoError(t, err)

	offer4 := createOfferTestOffer(t, db, "Test Pipeline 4", domain.OfferPhaseWon, domain.OfferStatusActive)
	offer4.Value = 15000
	err = db.Save(offer4).Error
	require.NoError(t, err)

	total, err := repo.GetTotalPipelineValue(context.Background())
	assert.NoError(t, err)
	// Total should include in_progress (10000) and sent (20000) = 30000
	assert.GreaterOrEqual(t, total, 30000.0)
}

func TestOfferRepository_Search(t *testing.T) {
	db := setupOfferTestDB(t)
	repo := repository.NewOfferRepository(db)

	createOfferTestOffer(t, db, "Test Search Steel Project", domain.OfferPhaseDraft, domain.OfferStatusActive)
	createOfferTestOffer(t, db, "Test Search Roofing Work", domain.OfferPhaseDraft, domain.OfferStatusActive)
	createOfferTestOffer(t, db, "Test Search Hybrid Building", domain.OfferPhaseDraft, domain.OfferStatusActive)

	t.Run("search by title", func(t *testing.T) {
		offers, err := repo.Search(context.Background(), "steel", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(offers), 1)

		found := false
		for _, offer := range offers {
			if offer.Title == "Test Search Steel Project" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("search case insensitive", func(t *testing.T) {
		offers, err := repo.Search(context.Background(), "ROOFING", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(offers), 1)
	})

	t.Run("search with limit", func(t *testing.T) {
		offers, err := repo.Search(context.Background(), "Test Search", 2)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(offers), 2)
	})
}
