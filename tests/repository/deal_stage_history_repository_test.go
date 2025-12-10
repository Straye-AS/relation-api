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

func setupStageHistoryTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupCleanTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createHistoryTestDeal(t *testing.T, db *gorm.DB) *domain.Deal {
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

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
		OwnerName:    "Test User",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	return deal
}

func TestDealStageHistoryRepository_Create(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	fromStage := domain.DealStageLead
	history := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage,
		ToStage:       domain.DealStageQualified,
		ChangedByID:   "user-123",
		ChangedByName: "Test User",
		Notes:         "Moved to qualified stage",
		ChangedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), history)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, history.ID)
}

func TestDealStageHistoryRepository_GetByDealID(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	// Create multiple history entries
	stages := []domain.DealStage{domain.DealStageLead, domain.DealStageQualified, domain.DealStageProposal}
	for i := 0; i < len(stages)-1; i++ {
		fromStage := stages[i]
		history := &domain.DealStageHistory{
			DealID:        deal.ID,
			FromStage:     &fromStage,
			ToStage:       stages[i+1],
			ChangedByID:   "user-123",
			ChangedByName: "Test User",
			ChangedAt:     time.Now().Add(time.Duration(i) * time.Hour),
		}
		err := db.Create(history).Error
		require.NoError(t, err)
	}

	histories, err := repo.GetByDealID(context.Background(), deal.ID)
	assert.NoError(t, err)
	assert.Len(t, histories, 2)

	// Should be ordered by changed_at DESC (most recent first)
	assert.Equal(t, domain.DealStageProposal, histories[0].ToStage)
	assert.Equal(t, domain.DealStageQualified, histories[1].ToStage)
}

func TestDealStageHistoryRepository_GetLatestByDealID(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	// Create multiple history entries
	fromStage1 := domain.DealStageLead
	history1 := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage1,
		ToStage:       domain.DealStageQualified,
		ChangedByID:   "user-123",
		ChangedByName: "Test User",
		ChangedAt:     time.Now().Add(-1 * time.Hour),
	}
	err := db.Create(history1).Error
	require.NoError(t, err)

	fromStage2 := domain.DealStageQualified
	history2 := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage2,
		ToStage:       domain.DealStageProposal,
		ChangedByID:   "user-123",
		ChangedByName: "Test User",
		ChangedAt:     time.Now(),
	}
	err = db.Create(history2).Error
	require.NoError(t, err)

	latest, err := repo.GetLatestByDealID(context.Background(), deal.ID)
	assert.NoError(t, err)
	assert.Equal(t, domain.DealStageProposal, latest.ToStage)
}

func TestDealStageHistoryRepository_GetByUserID(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	// Create entries by different users
	fromStage := domain.DealStageLead
	history1 := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage,
		ToStage:       domain.DealStageQualified,
		ChangedByID:   "user-123",
		ChangedByName: "User One",
		ChangedAt:     time.Now(),
	}
	err := db.Create(history1).Error
	require.NoError(t, err)

	fromStage2 := domain.DealStageQualified
	history2 := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage2,
		ToStage:       domain.DealStageProposal,
		ChangedByID:   "user-456",
		ChangedByName: "User Two",
		ChangedAt:     time.Now(),
	}
	err = db.Create(history2).Error
	require.NoError(t, err)

	histories, err := repo.GetByUserID(context.Background(), "user-123", 10)
	assert.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.Equal(t, "user-123", histories[0].ChangedByID)
}

func TestDealStageHistoryRepository_GetTransitionsToStage(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	fromStage := domain.DealStageLead
	history1 := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage,
		ToStage:       domain.DealStageQualified,
		ChangedByID:   "user-123",
		ChangedByName: "Test User",
		ChangedAt:     now,
	}
	err := db.Create(history1).Error
	require.NoError(t, err)

	fromStage2 := domain.DealStageQualified
	history2 := &domain.DealStageHistory{
		DealID:        deal.ID,
		FromStage:     &fromStage2,
		ToStage:       domain.DealStageWon,
		ChangedByID:   "user-123",
		ChangedByName: "Test User",
		ChangedAt:     now,
	}
	err = db.Create(history2).Error
	require.NoError(t, err)

	histories, err := repo.GetTransitionsToStage(context.Background(), domain.DealStageWon, yesterday, tomorrow)
	assert.NoError(t, err)
	assert.Len(t, histories, 1)
	assert.Equal(t, domain.DealStageWon, histories[0].ToStage)
}

func TestDealStageHistoryRepository_CountTransitionsByStage(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Create multiple transitions to different stages
	stages := []domain.DealStage{domain.DealStageQualified, domain.DealStageQualified, domain.DealStageWon}
	for i, toStage := range stages {
		fromStage := domain.DealStageLead
		history := &domain.DealStageHistory{
			DealID:        deal.ID,
			FromStage:     &fromStage,
			ToStage:       toStage,
			ChangedByID:   "user-123",
			ChangedByName: "Test User",
			ChangedAt:     now.Add(time.Duration(i) * time.Minute),
		}
		err := db.Create(history).Error
		require.NoError(t, err)
	}

	counts, err := repo.CountTransitionsByStage(context.Background(), yesterday, tomorrow)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), counts[domain.DealStageQualified])
	assert.Equal(t, int64(1), counts[domain.DealStageWon])
}

func TestDealStageHistoryRepository_RecordTransition(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	fromStage := domain.DealStageLead
	err := repo.RecordTransition(
		context.Background(),
		deal.ID,
		&fromStage,
		domain.DealStageQualified,
		"user-123",
		"Test User",
		"Qualified after initial meeting",
	)
	assert.NoError(t, err)

	// Verify the record was created
	var history domain.DealStageHistory
	err = db.Where("deal_id = ?", deal.ID).First(&history).Error
	assert.NoError(t, err)
	assert.Equal(t, domain.DealStageQualified, history.ToStage)
	assert.Equal(t, domain.DealStageLead, *history.FromStage)
	assert.Equal(t, "Qualified after initial meeting", history.Notes)
}

func TestDealStageHistoryRepository_RecordTransition_InitialState(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	// Record initial state with nil from stage
	err := repo.RecordTransition(
		context.Background(),
		deal.ID,
		nil,
		domain.DealStageLead,
		"user-123",
		"Test User",
		"Deal created",
	)
	assert.NoError(t, err)

	var history domain.DealStageHistory
	err = db.Where("deal_id = ?", deal.ID).First(&history).Error
	assert.NoError(t, err)
	assert.Nil(t, history.FromStage)
	assert.Equal(t, domain.DealStageLead, history.ToStage)
}

func TestDealStageHistoryRepository_DeleteByDealID(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	// Create some history
	fromStage := domain.DealStageLead
	for i := 0; i < 3; i++ {
		history := &domain.DealStageHistory{
			DealID:        deal.ID,
			FromStage:     &fromStage,
			ToStage:       domain.DealStageQualified,
			ChangedByID:   "user-123",
			ChangedByName: "Test User",
			ChangedAt:     time.Now(),
		}
		err := db.Create(history).Error
		require.NoError(t, err)
	}

	// Verify records exist
	var count int64
	db.Model(&domain.DealStageHistory{}).Where("deal_id = ?", deal.ID).Count(&count)
	assert.Equal(t, int64(3), count)

	// Delete all history for the deal
	err := repo.DeleteByDealID(context.Background(), deal.ID)
	assert.NoError(t, err)

	// Verify records are deleted
	db.Model(&domain.DealStageHistory{}).Where("deal_id = ?", deal.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDealStageHistoryRepository_GetAverageTimeInStage(t *testing.T) {
	db := setupStageHistoryTestDB(t)
	repo := repository.NewDealStageHistoryRepository(db)
	deal := createHistoryTestDeal(t, db)

	baseTime := time.Now()

	// Create a progression through stages
	stages := []struct {
		from      *domain.DealStage
		to        domain.DealStage
		changedAt time.Time
	}{
		{nil, domain.DealStageLead, baseTime},
		{ptr(domain.DealStageLead), domain.DealStageQualified, baseTime.Add(2 * time.Hour)},
		{ptr(domain.DealStageQualified), domain.DealStageProposal, baseTime.Add(5 * time.Hour)},
	}

	for _, s := range stages {
		history := &domain.DealStageHistory{
			DealID:        deal.ID,
			FromStage:     s.from,
			ToStage:       s.to,
			ChangedByID:   "user-123",
			ChangedByName: "Test User",
			ChangedAt:     s.changedAt,
		}
		err := db.Create(history).Error
		require.NoError(t, err)
	}

	averages, err := repo.GetAverageTimeInStage(context.Background())
	assert.NoError(t, err)

	// Lead stage: 2 hours, Qualified stage: 3 hours
	if avgLead, ok := averages[domain.DealStageLead]; ok {
		assert.True(t, avgLead >= 1*time.Hour && avgLead <= 3*time.Hour, "Expected ~2 hours for lead stage, got %v", avgLead)
	}
	if avgQualified, ok := averages[domain.DealStageQualified]; ok {
		assert.True(t, avgQualified >= 2*time.Hour && avgQualified <= 4*time.Hour, "Expected ~3 hours for qualified stage, got %v", avgQualified)
	}
}

// Helper function for creating pointer to DealStage
func ptr(s domain.DealStage) *domain.DealStage {
	return &s
}
