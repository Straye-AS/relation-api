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

func setupDealTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createTestCustomer(t *testing.T, db *gorm.DB) *domain.Customer {
	return testutil.CreateTestCustomer(t, db, "Test Customer")
}

func TestDealRepository_Create(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deal := &domain.Deal{
		Title:        "Test Deal",
		Description:  "A test deal",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Probability:  10,
		Value:        100000,
		Currency:     "NOK",
		OwnerID:      "user-123",
		OwnerName:    "Test User",
		Source:       "Website",
	}

	err := repo.Create(context.Background(), deal)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, deal.ID)
}

func TestDealRepository_GetByID(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deal := &domain.Deal{
		Title:        "Test Deal",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Probability:  10,
		Value:        100000,
		Currency:     "NOK",
		OwnerID:      "user-123",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), deal.ID)
	assert.NoError(t, err)
	assert.Equal(t, deal.Title, found.Title)
	assert.Equal(t, deal.CustomerID, found.CustomerID)
	assert.Equal(t, deal.Stage, found.Stage)
	assert.Equal(t, deal.Value, found.Value)
}

func TestDealRepository_Update(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deal := &domain.Deal{
		Title:       "Original Title",
		CustomerID:  customer.ID,
		CompanyID:   domain.CompanyStalbygg,
		Stage:       domain.DealStageLead,
		Probability: 10,
		Value:       100000,
		Currency:    "NOK",
		OwnerID:     "user-123",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	deal.Title = "Updated Title"
	deal.Value = 200000
	deal.Stage = domain.DealStageQualified

	err = repo.Update(context.Background(), deal)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), deal.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", found.Title)
	assert.Equal(t, float64(200000), found.Value)
	assert.Equal(t, domain.DealStageQualified, found.Stage)
}

func TestDealRepository_Delete(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deal := &domain.Deal{
		Title:      "Test Deal",
		CustomerID: customer.ID,
		CompanyID:  domain.CompanyStalbygg,
		Stage:      domain.DealStageLead,
		OwnerID:    "user-123",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	err = repo.Delete(context.Background(), deal.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), deal.ID)
	assert.Error(t, err)
}

func TestDealRepository_List(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	// Create test deals
	deals := []*domain.Deal{
		{Title: "Deal 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, Value: 100000, Probability: 10, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Deal 2", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageQualified, Value: 200000, Probability: 25, OwnerID: "user-2", Currency: "NOK"},
		{Title: "Deal 3", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageProposal, Value: 300000, Probability: 50, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	t.Run("list all", func(t *testing.T) {
		result, total, err := repo.List(context.Background(), 1, 10, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("filter by stage", func(t *testing.T) {
		stage := domain.DealStageLead
		filters := &repository.DealFilters{Stage: &stage}
		result, total, err := repo.List(context.Background(), 1, 10, filters, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "Deal 1", result[0].Title)
	})

	t.Run("filter by owner", func(t *testing.T) {
		ownerID := "user-1"
		filters := &repository.DealFilters{OwnerID: &ownerID}
		result, total, err := repo.List(context.Background(), 1, 10, filters, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})

	t.Run("filter by value range", func(t *testing.T) {
		minVal := float64(150000)
		maxVal := float64(250000)
		filters := &repository.DealFilters{MinValue: &minVal, MaxValue: &maxVal}
		result, total, err := repo.List(context.Background(), 1, 10, filters, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "Deal 2", result[0].Title)
	})

	t.Run("sort by value desc", func(t *testing.T) {
		result, _, err := repo.List(context.Background(), 1, 10, nil, repository.DealSortByValueDesc)
		assert.NoError(t, err)
		assert.Equal(t, "Deal 3", result[0].Title)
		assert.Equal(t, "Deal 2", result[1].Title)
		assert.Equal(t, "Deal 1", result[2].Title)
	})

	t.Run("sort by probability desc", func(t *testing.T) {
		result, _, err := repo.List(context.Background(), 1, 10, nil, repository.DealSortByProbabilityDesc)
		assert.NoError(t, err)
		assert.Equal(t, 50, result[0].Probability)
		assert.Equal(t, 25, result[1].Probability)
		assert.Equal(t, 10, result[2].Probability)
	})

	t.Run("pagination", func(t *testing.T) {
		result, total, err := repo.List(context.Background(), 1, 2, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 2)

		result, total, err = repo.List(context.Background(), 2, 2, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 1)
	})
}

func TestDealRepository_GetByStage(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deals := []*domain.Deal{
		{Title: "Lead 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, Value: 100000, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Lead 2", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, Value: 200000, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Qualified 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageQualified, Value: 300000, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	result, err := repo.GetByStage(context.Background(), domain.DealStageLead)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	result, err = repo.GetByStage(context.Background(), domain.DealStageQualified)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestDealRepository_GetPipelineOverview(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deals := []*domain.Deal{
		{Title: "Lead 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Qualified 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageQualified, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Won 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageWon, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	pipeline, err := repo.GetPipelineOverview(context.Background())
	assert.NoError(t, err)

	// Won deals should be excluded from pipeline
	assert.Len(t, pipeline[domain.DealStageLead], 1)
	assert.Len(t, pipeline[domain.DealStageQualified], 1)
	assert.Len(t, pipeline[domain.DealStageWon], 0)
}

func TestDealRepository_GetTotalPipelineValue(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deals := []*domain.Deal{
		{Title: "Open 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, Value: 100000, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Open 2", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageQualified, Value: 200000, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Won", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageWon, Value: 500000, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Lost", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLost, Value: 300000, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	total, err := repo.GetTotalPipelineValue(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, float64(300000), total) // Only open deals: 100000 + 200000
}

func TestDealRepository_MarkAsWon(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deal := &domain.Deal{
		Title:       "Deal to Win",
		CustomerID:  customer.ID,
		CompanyID:   domain.CompanyStalbygg,
		Stage:       domain.DealStageNegotiation,
		Probability: 75,
		Value:       100000,
		OwnerID:     "user-1",
		Currency:    "NOK",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	closeDate := time.Now()
	err = repo.MarkAsWon(context.Background(), deal.ID, closeDate)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), deal.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.DealStageWon, found.Stage)
	assert.Equal(t, 100, found.Probability)
	assert.NotNil(t, found.ActualCloseDate)
}

func TestDealRepository_MarkAsLost(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deal := &domain.Deal{
		Title:       "Deal to Lose",
		CustomerID:  customer.ID,
		CompanyID:   domain.CompanyStalbygg,
		Stage:       domain.DealStageProposal,
		Probability: 50,
		Value:       100000,
		OwnerID:     "user-1",
		Currency:    "NOK",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	closeDate := time.Now()
	reasonCategory := domain.LossReasonCompetitor
	notes := "Lost to competitor who offered lower price"
	err = repo.MarkAsLost(context.Background(), deal.ID, closeDate, reasonCategory, notes)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), deal.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.DealStageLost, found.Stage)
	assert.Equal(t, 0, found.Probability)
	assert.Equal(t, notes, found.LostReason)
	assert.NotNil(t, found.LossReasonCategory)
	assert.Equal(t, reasonCategory, *found.LossReasonCategory)
	assert.NotNil(t, found.ActualCloseDate)
}

func TestDealRepository_Search(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deals := []*domain.Deal{
		{Title: "Steel Building Project", CustomerID: customer.ID, CustomerName: "ABC Corp", CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Warehouse Construction", CustomerID: customer.ID, CustomerName: "XYZ Inc", CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Office Renovation", CustomerID: customer.ID, CustomerName: "Steel Corp", CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	result, err := repo.Search(context.Background(), "steel", 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2) // "Steel Building Project" and "Steel Corp" customer
}

func TestDealRepository_GetByOwner(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	deals := []*domain.Deal{
		{Title: "Deal 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Deal 2", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-2", Currency: "NOK"},
		{Title: "Deal 3", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	result, err := repo.GetByOwner(context.Background(), "user-1")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestDealRepository_GetByCustomer(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer1 := createTestCustomer(t, db)

	customer2 := testutil.CreateTestCustomer(t, db, "Another Customer")

	deals := []*domain.Deal{
		{Title: "Deal 1", CustomerID: customer1.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Deal 2", CustomerID: customer2.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Deal 3", CustomerID: customer1.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	result, err := repo.GetByCustomer(context.Background(), customer1.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestDealRepository_GetForecast(t *testing.T) {
	db := setupDealTestDB(t)
	repo := repository.NewDealRepository(db)
	customer := createTestCustomer(t, db)

	// Create deals with expected close dates in different months
	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, time.UTC)
	nextMonth := thisMonth.AddDate(0, 1, 0)

	deals := []*domain.Deal{
		{Title: "Deal This Month", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, Value: 100000, Probability: 50, ExpectedCloseDate: &thisMonth, OwnerID: "user-1", Currency: "NOK"},
		{Title: "Deal Next Month", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageQualified, Value: 200000, Probability: 75, ExpectedCloseDate: &nextMonth, OwnerID: "user-1", Currency: "NOK"},
	}
	for _, d := range deals {
		err := db.Create(d).Error
		require.NoError(t, err)
	}

	forecast, err := repo.GetForecast(context.Background(), 3)
	assert.NoError(t, err)
	assert.Len(t, forecast, 3)

	// First period should have Deal This Month
	assert.Equal(t, int64(1), forecast[0].DealCount)
	assert.Equal(t, float64(100000), forecast[0].TotalValue)
}
