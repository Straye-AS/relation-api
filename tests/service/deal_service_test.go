package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func setupDealServiceTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createDealServiceTestCustomer(t *testing.T, db *gorm.DB) *domain.Customer {
	return testutil.CreateTestCustomer(t, db, "Test Customer")
}

func createDealService(t *testing.T, db *gorm.DB) *service.DealService {
	logger := zap.NewNop()
	dealRepo := repository.NewDealRepository(db)
	historyRepo := repository.NewDealStageHistoryRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	return service.NewDealService(dealRepo, historyRepo, customerRepo, projectRepo, activityRepo, offerRepo, notificationRepo, logger)
}

func createTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin}, // SuperAdmin bypasses company filter
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func TestDealService_Create(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	t.Run("create with minimal fields", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "New Deal",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Value:      100000,
		}

		deal, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, deal)
		assert.Equal(t, "New Deal", deal.Title)
		assert.Equal(t, domain.DealStageLead, deal.Stage)
		assert.Equal(t, 10, deal.Probability) // Default for lead
		assert.Equal(t, "NOK", deal.Currency) // Default currency
	})

	t.Run("create with all fields", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		expectedClose := time.Now().AddDate(0, 1, 0)
		req := &domain.CreateDealRequest{
			Title:             "Full Deal",
			Description:       "A complete deal with all fields",
			CustomerID:        customer.ID,
			CompanyID:         domain.CompanyStalbygg,
			Stage:             domain.DealStageQualified,
			Probability:       30,
			Value:             500000,
			Currency:          "USD",
			ExpectedCloseDate: &expectedClose,
			OwnerID:           userCtx.UserID.String(),
			Source:            "Referral",
			Notes:             "High priority deal",
		}

		deal, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, deal)
		assert.Equal(t, "Full Deal", deal.Title)
		assert.Equal(t, domain.DealStageQualified, deal.Stage)
		assert.Equal(t, 30, deal.Probability)
		assert.Equal(t, "USD", deal.Currency)
		assert.Equal(t, "Referral", deal.Source)
	})

	t.Run("create with invalid customer", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Invalid Deal",
			CustomerID: uuid.New(), // Non-existent
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
		}

		deal, err := svc.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, deal)
		assert.Contains(t, err.Error(), "customer not found")
	})
}

func TestDealService_GetByID(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create a deal first
	userCtx, _ := auth.FromContext(ctx)
	req := &domain.CreateDealRequest{
		Title:      "Test Deal",
		CustomerID: customer.ID,
		CompanyID:  domain.CompanyStalbygg,
		OwnerID:    userCtx.UserID.String(),
		Value:      100000,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("get existing deal", func(t *testing.T) {
		deal, err := svc.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, deal)
		assert.Equal(t, created.ID, deal.ID)
		assert.Equal(t, "Test Deal", deal.Title)
	})

	t.Run("get non-existent deal", func(t *testing.T) {
		deal, err := svc.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, deal)
	})
}

func TestDealService_Update(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create a deal first
	userCtx, _ := auth.FromContext(ctx)
	req := &domain.CreateDealRequest{
		Title:      "Original Title",
		CustomerID: customer.ID,
		CompanyID:  domain.CompanyStalbygg,
		OwnerID:    userCtx.UserID.String(),
		Value:      100000,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("update title and value", func(t *testing.T) {
		updateReq := &domain.UpdateDealRequest{
			Title: "Updated Title",
			Value: 200000,
		}

		deal, err := svc.Update(ctx, created.ID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, deal)
		assert.Equal(t, "Updated Title", deal.Title)
		assert.Equal(t, float64(200000), deal.Value)
	})

	t.Run("update stage", func(t *testing.T) {
		updateReq := &domain.UpdateDealRequest{
			Title: "Updated Title",
			Stage: domain.DealStageQualified,
		}

		deal, err := svc.Update(ctx, created.ID, updateReq)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageQualified, deal.Stage)
		assert.Equal(t, 25, deal.Probability) // Default for qualified
	})

	t.Run("update by non-owner without permission", func(t *testing.T) {
		// Create context with different user
		otherUserCtx := &auth.UserContext{
			UserID:      uuid.New(),
			DisplayName: "Other User",
			Email:       "other@example.com",
			Roles:       []domain.UserRoleType{domain.RoleViewer}, // Not manager
		}
		otherCtx := auth.WithUserContext(context.Background(), otherUserCtx)

		updateReq := &domain.UpdateDealRequest{
			Title: "Unauthorized Update",
		}

		deal, err := svc.Update(otherCtx, created.ID, updateReq)
		assert.Error(t, err)
		assert.Nil(t, deal)
	})
}

func TestDealService_Delete(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create a deal first
	userCtx, _ := auth.FromContext(ctx)
	req := &domain.CreateDealRequest{
		Title:      "Deal to Delete",
		CustomerID: customer.ID,
		CompanyID:  domain.CompanyStalbygg,
		OwnerID:    userCtx.UserID.String(),
		Value:      100000,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("delete existing deal", func(t *testing.T) {
		err := svc.Delete(ctx, created.ID)
		assert.NoError(t, err)

		// Verify it's gone
		_, err = svc.GetByID(ctx, created.ID)
		assert.Error(t, err)
	})

	t.Run("delete non-existent deal", func(t *testing.T) {
		err := svc.Delete(ctx, uuid.New())
		assert.Error(t, err)
	})
}

func TestDealService_AdvanceStage(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create a deal in lead stage
	userCtx, _ := auth.FromContext(ctx)
	req := &domain.CreateDealRequest{
		Title:      "Stage Test Deal",
		CustomerID: customer.ID,
		CompanyID:  domain.CompanyStalbygg,
		OwnerID:    userCtx.UserID.String(),
		Value:      100000,
		Stage:      domain.DealStageLead,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("valid transition lead to qualified", func(t *testing.T) {
		stageReq := &domain.UpdateDealStageRequest{
			Stage: domain.DealStageQualified,
			Notes: "Customer qualified after initial meeting",
		}

		deal, err := svc.AdvanceStage(ctx, created.ID, stageReq)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageQualified, deal.Stage)
		assert.Equal(t, 25, deal.Probability)
	})

	t.Run("valid transition qualified to proposal", func(t *testing.T) {
		stageReq := &domain.UpdateDealStageRequest{
			Stage: domain.DealStageProposal,
		}

		deal, err := svc.AdvanceStage(ctx, created.ID, stageReq)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageProposal, deal.Stage)
		assert.Equal(t, 50, deal.Probability)
	})

	t.Run("invalid transition proposal to won", func(t *testing.T) {
		stageReq := &domain.UpdateDealStageRequest{
			Stage: domain.DealStageWon,
		}

		deal, err := svc.AdvanceStage(ctx, created.ID, stageReq)
		assert.Error(t, err)
		assert.Nil(t, deal)
		assert.Contains(t, err.Error(), "invalid stage transition")
	})

	t.Run("valid transition to lost", func(t *testing.T) {
		stageReq := &domain.UpdateDealStageRequest{
			Stage: domain.DealStageLost,
			Notes: "Budget constraints",
		}

		deal, err := svc.AdvanceStage(ctx, created.ID, stageReq)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageLost, deal.Stage)
		assert.Equal(t, 0, deal.Probability)
	})
}

func TestDealService_WinDeal(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	createDealInNegotiation := func() *domain.DealDTO {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Deal to Win",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Value:      500000,
			Stage:      domain.DealStageNegotiation,
		}
		deal, err := svc.Create(ctx, req)
		require.NoError(t, err)
		// Manually update to negotiation stage since create validates transitions
		db.Model(&domain.Deal{}).Where("id = ?", deal.ID).Update("stage", domain.DealStageNegotiation)
		return deal
	}

	t.Run("win deal without project creation", func(t *testing.T) {
		deal := createDealInNegotiation()

		wonDeal, project, err := svc.WinDeal(ctx, deal.ID, false)
		assert.NoError(t, err)
		assert.NotNil(t, wonDeal)
		assert.Nil(t, project)
		assert.Equal(t, domain.DealStageWon, wonDeal.Stage)
		assert.Equal(t, 100, wonDeal.Probability)
		assert.NotNil(t, wonDeal.ActualCloseDate)
	})

	t.Run("win deal with project creation", func(t *testing.T) {
		deal := createDealInNegotiation()

		wonDeal, project, err := svc.WinDeal(ctx, deal.ID, true)
		require.NoError(t, err)
		require.NotNil(t, wonDeal)
		require.NotNil(t, project, "project should be created when createProject=true")
		assert.Equal(t, domain.DealStageWon, wonDeal.Stage)
		assert.Equal(t, deal.Title, project.Name)
		assert.Equal(t, domain.ProjectStatusPlanning, project.Status)
	})

	t.Run("win deal not in negotiation", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Lead Deal",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Stage:      domain.DealStageLead,
		}
		deal, _ := svc.Create(ctx, req)

		_, _, err := svc.WinDeal(ctx, deal.ID, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be in negotiation stage")
	})
}

func TestDealService_LoseDeal(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	t.Run("lose deal with reason category and notes", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Deal to Lose",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Stage:      domain.DealStageProposal,
		}
		deal, _ := svc.Create(ctx, req)

		loseReq := &domain.LoseDealRequest{
			Reason: domain.LossReasonCompetitor,
			Notes:  "Lost to competitor XYZ who offered lower price",
		}
		lostDeal, err := svc.LoseDeal(ctx, deal.ID, loseReq)
		assert.NoError(t, err)
		assert.NotNil(t, lostDeal)
		assert.Equal(t, domain.DealStageLost, lostDeal.Stage)
		assert.Equal(t, 0, lostDeal.Probability)
		assert.Equal(t, loseReq.Notes, lostDeal.LostReason)
		assert.NotNil(t, lostDeal.LossReasonCategory)
		assert.Equal(t, domain.LossReasonCompetitor, *lostDeal.LossReasonCategory)
		assert.NotNil(t, lostDeal.ActualCloseDate)
	})

	t.Run("cannot lose already won deal", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Won Deal",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Stage:      domain.DealStageNegotiation,
		}
		deal, _ := svc.Create(ctx, req)
		// Manually update to won
		db.Model(&domain.Deal{}).Where("id = ?", deal.ID).Update("stage", domain.DealStageWon)

		loseReq := &domain.LoseDealRequest{
			Reason: domain.LossReasonOther,
			Notes:  "Testing that won deals cannot be lost",
		}
		_, err := svc.LoseDeal(ctx, deal.ID, loseReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot mark a won deal as lost")
	})

	t.Run("cannot lose already lost deal", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Lost Deal",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
		}
		deal, _ := svc.Create(ctx, req)
		// Lose it first
		firstLossReq := &domain.LoseDealRequest{
			Reason: domain.LossReasonPrice,
			Notes:  "First loss due to price concerns from the client",
		}
		svc.LoseDeal(ctx, deal.ID, firstLossReq)

		secondLossReq := &domain.LoseDealRequest{
			Reason: domain.LossReasonTiming,
			Notes:  "Attempting second loss which should fail validation",
		}
		_, err := svc.LoseDeal(ctx, deal.ID, secondLossReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already marked as lost")
	})
}

func TestDealService_ReopenDeal(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	t.Run("reopen lost deal", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Deal to Reopen",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Stage:      domain.DealStageProposal,
		}
		deal, _ := svc.Create(ctx, req)

		// Lose it first
		loseReq := &domain.LoseDealRequest{
			Reason: domain.LossReasonTiming,
			Notes:  "Temporary loss due to timing issues with client",
		}
		_, err := svc.LoseDeal(ctx, deal.ID, loseReq)
		require.NoError(t, err)

		// Reopen
		reopened, err := svc.ReopenDeal(ctx, deal.ID)
		assert.NoError(t, err)
		assert.NotNil(t, reopened)
		assert.Equal(t, domain.DealStageLead, reopened.Stage)
		assert.Equal(t, 10, reopened.Probability)
		assert.Nil(t, reopened.ActualCloseDate)
		assert.Empty(t, reopened.LostReason)
		assert.Nil(t, reopened.LossReasonCategory)
	})

	t.Run("cannot reopen non-lost deal", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		req := &domain.CreateDealRequest{
			Title:      "Active Deal",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
		}
		deal, _ := svc.Create(ctx, req)

		_, err := svc.ReopenDeal(ctx, deal.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only lost deals can be reopened")
	})
}

func TestDealService_GetStageHistory(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create a deal and advance through stages
	userCtx, _ := auth.FromContext(ctx)
	req := &domain.CreateDealRequest{
		Title:      "History Test Deal",
		CustomerID: customer.ID,
		CompanyID:  domain.CompanyStalbygg,
		OwnerID:    userCtx.UserID.String(),
	}
	deal, err := svc.Create(ctx, req)
	require.NoError(t, err)

	// Advance to qualified
	svc.AdvanceStage(ctx, deal.ID, &domain.UpdateDealStageRequest{Stage: domain.DealStageQualified})

	// Get history
	history, err := svc.GetStageHistory(ctx, deal.ID)
	assert.NoError(t, err)
	assert.Len(t, history, 2) // Initial creation + advance

	// History is ordered DESC, so most recent first
	assert.Equal(t, domain.DealStageQualified, history[0].ToStage)
	assert.Equal(t, domain.DealStageLead, history[1].ToStage)
}

func TestDealService_List(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create test deals
	userCtx, _ := auth.FromContext(ctx)
	for i := 0; i < 5; i++ {
		req := &domain.CreateDealRequest{
			Title:      "Deal " + string(rune('A'+i)),
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Value:      float64((i + 1) * 100000),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	t.Run("list all", func(t *testing.T) {
		result, err := svc.List(ctx, 1, 10, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Data, 5)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(ctx, 1, 2, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, 3, result.TotalPages)

		result, err = svc.List(ctx, 2, 2, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})

	t.Run("list with filter", func(t *testing.T) {
		stage := domain.DealStageLead
		filters := &repository.DealFilters{Stage: &stage}
		result, err := svc.List(ctx, 1, 10, filters, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total) // All are leads
	})

	t.Run("clamps page size", func(t *testing.T) {
		result, err := svc.List(ctx, 1, 500, nil, repository.DealSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, 200, result.PageSize)
	})
}

func TestDealService_GetPipelineOverview(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create deals in different stages
	userCtx, _ := auth.FromContext(ctx)
	stages := []domain.DealStage{domain.DealStageLead, domain.DealStageLead, domain.DealStageQualified}
	for i, stage := range stages {
		req := &domain.CreateDealRequest{
			Title:      "Deal " + string(rune('A'+i)),
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Stage:      stage,
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	pipeline, err := svc.GetPipelineOverview(ctx)
	assert.NoError(t, err)
	assert.Len(t, pipeline["lead"], 2)
	assert.Len(t, pipeline["qualified"], 1)
}

func TestDealService_GetForecast(t *testing.T) {
	db := setupDealServiceTestDB(t)
	svc := createDealService(t, db)
	customer := createDealServiceTestCustomer(t, db)
	ctx := createTestContext()

	// Create deals with expected close dates
	userCtx, _ := auth.FromContext(ctx)
	closeDate := time.Now().AddDate(0, 1, 0)
	req := &domain.CreateDealRequest{
		Title:             "Forecast Deal",
		CustomerID:        customer.ID,
		CompanyID:         domain.CompanyStalbygg,
		OwnerID:           userCtx.UserID.String(),
		Value:             100000,
		ExpectedCloseDate: &closeDate,
	}
	_, err := svc.Create(ctx, req)
	require.NoError(t, err)

	forecast, err := svc.GetForecast(ctx, 3)
	assert.NoError(t, err)
	assert.Len(t, forecast, 3)
}
