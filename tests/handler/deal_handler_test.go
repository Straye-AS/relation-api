package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func setupDealHandlerTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createDealHandlerTestCustomer(t *testing.T, db *gorm.DB) *domain.Customer {
	return testutil.CreateTestCustomer(t, db, "Test Customer")
}

func createDealHandler(t *testing.T, db *gorm.DB) *handler.DealHandler {
	logger := zap.NewNop()
	dealRepo := repository.NewDealRepository(db)
	historyRepo := repository.NewDealStageHistoryRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	dealService := service.NewDealService(dealRepo, historyRepo, customerRepo, projectRepo, activityRepo, offerRepo, budgetItemRepo, notificationRepo, logger, db)

	return handler.NewDealHandler(dealService, logger)
}

func createDealTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin}, // SuperAdmin bypasses company filter
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func TestDealHandler_Create(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	t.Run("create valid deal", func(t *testing.T) {
		reqBody := domain.CreateDealRequest{
			Title:      "New Deal",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			OwnerID:    userCtx.UserID.String(),
			Value:      100000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var deal domain.DealDTO
		err := json.Unmarshal(rr.Body.Bytes(), &deal)
		assert.NoError(t, err)
		assert.Equal(t, "New Deal", deal.Title)
		assert.NotEmpty(t, rr.Header().Get("Location"))
	})

	t.Run("create with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/deals", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with missing required fields", func(t *testing.T) {
		reqBody := domain.CreateDealRequest{
			Title: "Missing Fields",
			// Missing CustomerID and CompanyID
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDealHandler_GetByID(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a deal first
	deal := &domain.Deal{
		Title:        "Test Deal",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Probability:  10,
		Value:        100000,
		Currency:     "NOK",
		OwnerID:      userCtx.UserID.String(),
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	t.Run("get existing deal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/"+deal.ID.String(), nil)
		req = req.WithContext(ctx)

		// Set chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response handler.DealWithHistoryResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, deal.ID, response.Deal.ID)
	})

	t.Run("get non-existent deal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/"+uuid.New().String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/invalid-id", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-id")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDealHandler_Update(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a deal
	deal := &domain.Deal{
		Title:        "Original Deal",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Probability:  10,
		Value:        100000,
		Currency:     "NOK",
		OwnerID:      userCtx.UserID.String(),
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	t.Run("update deal", func(t *testing.T) {
		reqBody := domain.UpdateDealRequest{
			Title: "Updated Deal",
			Value: 200000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/deals/"+deal.ID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var updated domain.DealDTO
		err := json.Unmarshal(rr.Body.Bytes(), &updated)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Deal", updated.Title)
		assert.Equal(t, float64(200000), updated.Value)
	})

	t.Run("update with invalid ID", func(t *testing.T) {
		reqBody := domain.UpdateDealRequest{Title: "Test"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/deals/invalid", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDealHandler_Delete(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a deal
	deal := &domain.Deal{
		Title:        "Deal to Delete",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		OwnerID:      userCtx.UserID.String(),
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	t.Run("delete deal", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/deals/"+deal.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify deleted
		var found domain.Deal
		err := db.Where("id = ?", deal.ID).First(&found).Error
		assert.Error(t, err)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/deals/invalid", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDealHandler_List(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create some deals
	for i := 0; i < 5; i++ {
		err := db.Create(&domain.Deal{
			Title:        "Deal " + string(rune('A'+i)),
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Stage:        domain.DealStageLead,
			Value:        float64((i + 1) * 100000),
			Currency:     "NOK",
			OwnerID:      userCtx.UserID.String(),
		}).Error
		require.NoError(t, err)
	}

	t.Run("list all", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
	})

	t.Run("list with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals?page=1&pageSize=2", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 2, result.PageSize)
	})

	t.Run("list with stage filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals?stage=lead", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
	})

	t.Run("list with value filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals?minValue=300000", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total) // Deals with value >= 300000
	})
}

func TestDealHandler_AdvanceStage(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a deal in lead stage
	deal := &domain.Deal{
		Title:        "Stage Test Deal",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Probability:  10,
		Currency:     "NOK",
		OwnerID:      userCtx.UserID.String(),
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	t.Run("advance to qualified", func(t *testing.T) {
		reqBody := domain.UpdateDealStageRequest{
			Stage: domain.DealStageQualified,
			Notes: "Customer qualified",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/advance", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.AdvanceStage(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var updated domain.DealDTO
		err := json.Unmarshal(rr.Body.Bytes(), &updated)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageQualified, updated.Stage)
		assert.Equal(t, 25, updated.Probability)
	})

	t.Run("invalid stage transition", func(t *testing.T) {
		reqBody := domain.UpdateDealStageRequest{
			Stage: domain.DealStageWon, // Invalid: qualified -> won
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/advance", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.AdvanceStage(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestDealHandler_WinDeal(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a deal in negotiation stage
	deal := &domain.Deal{
		Title:        "Deal to Win",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageNegotiation,
		Probability:  75,
		Value:        500000,
		Currency:     "NOK",
		OwnerID:      userCtx.UserID.String(),
		OwnerName:    userCtx.DisplayName,
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	t.Run("win deal with project creation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/win?createProject=true", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.WinDeal(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response handler.WinDealResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageWon, response.Deal.Stage)
		assert.NotNil(t, response.Project)
		assert.Equal(t, deal.Title, response.Project.Name)
	})
}

func TestDealHandler_LoseDeal(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	t.Run("lose deal with valid reason and notes", func(t *testing.T) {
		// Create a fresh deal for this test
		deal := &domain.Deal{
			Title:        "Deal to Lose",
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Stage:        domain.DealStageProposal,
			Probability:  50,
			Currency:     "NOK",
			OwnerID:      userCtx.UserID.String(),
		}
		err := db.Create(deal).Error
		require.NoError(t, err)

		reqBody := domain.LoseDealRequest{
			Reason: domain.LossReasonCompetitor,
			Notes:  "Lost to competitor XYZ who offered a lower price",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/lose", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.LoseDeal(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var lostDeal domain.DealDTO
		err = json.Unmarshal(rr.Body.Bytes(), &lostDeal)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageLost, lostDeal.Stage)
		assert.Equal(t, "Lost to competitor XYZ who offered a lower price", lostDeal.LostReason)
		assert.NotNil(t, lostDeal.LossReasonCategory)
		assert.Equal(t, domain.LossReasonCompetitor, *lostDeal.LossReasonCategory)
	})

	t.Run("lose deal with missing reason category", func(t *testing.T) {
		deal := &domain.Deal{
			Title:        "Deal Missing Reason",
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Stage:        domain.DealStageProposal,
			Probability:  50,
			Currency:     "NOK",
			OwnerID:      userCtx.UserID.String(),
		}
		err := db.Create(deal).Error
		require.NoError(t, err)

		reqBody := map[string]string{
			"notes": "This has notes but no reason category",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/lose", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.LoseDeal(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("lose deal with invalid reason category", func(t *testing.T) {
		deal := &domain.Deal{
			Title:        "Deal Invalid Reason",
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Stage:        domain.DealStageProposal,
			Probability:  50,
			Currency:     "NOK",
			OwnerID:      userCtx.UserID.String(),
		}
		err := db.Create(deal).Error
		require.NoError(t, err)

		reqBody := map[string]string{
			"reason": "invalid_reason",
			"notes":  "This has an invalid reason category",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/lose", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.LoseDeal(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("lose deal with notes too short", func(t *testing.T) {
		deal := &domain.Deal{
			Title:        "Deal Short Notes",
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Stage:        domain.DealStageProposal,
			Probability:  50,
			Currency:     "NOK",
			OwnerID:      userCtx.UserID.String(),
		}
		err := db.Create(deal).Error
		require.NoError(t, err)

		reqBody := domain.LoseDealRequest{
			Reason: domain.LossReasonPrice,
			Notes:  "Too short", // Less than 10 characters
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/lose", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.LoseDeal(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("lose deal with all valid reason categories", func(t *testing.T) {
		categories := []domain.LossReasonCategory{
			domain.LossReasonPrice,
			domain.LossReasonTiming,
			domain.LossReasonCompetitor,
			domain.LossReasonRequirements,
			domain.LossReasonOther,
		}

		for _, category := range categories {
			deal := &domain.Deal{
				Title:        "Deal " + string(category),
				CustomerID:   customer.ID,
				CustomerName: customer.Name,
				CompanyID:    domain.CompanyStalbygg,
				Stage:        domain.DealStageProposal,
				Probability:  50,
				Currency:     "NOK",
				OwnerID:      userCtx.UserID.String(),
			}
			err := db.Create(deal).Error
			require.NoError(t, err)

			reqBody := domain.LoseDealRequest{
				Reason: category,
				Notes:  "Testing category: " + string(category) + " validation",
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/lose", bytes.NewReader(body))
			req = req.WithContext(ctx)
			req.Header.Set("Content-Type", "application/json")

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", deal.ID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			h.LoseDeal(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code, "Expected OK for category %s", category)

			var lostDeal domain.DealDTO
			err = json.Unmarshal(rr.Body.Bytes(), &lostDeal)
			assert.NoError(t, err)
			assert.Equal(t, domain.DealStageLost, lostDeal.Stage)
			assert.NotNil(t, lostDeal.LossReasonCategory)
			assert.Equal(t, category, *lostDeal.LossReasonCategory)
		}
	})
}

func TestDealHandler_ReopenDeal(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a lost deal with loss reason category
	lossCategory := domain.LossReasonPrice
	deal := &domain.Deal{
		Title:              "Lost Deal",
		CustomerID:         customer.ID,
		CustomerName:       customer.Name,
		CompanyID:          domain.CompanyStalbygg,
		Stage:              domain.DealStageLost,
		Probability:        0,
		LostReason:         "Previous loss due to price",
		LossReasonCategory: &lossCategory,
		Currency:           "NOK",
		OwnerID:            userCtx.UserID.String(),
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	t.Run("reopen lost deal clears loss reason", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/deals/"+deal.ID.String()+"/reopen", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", deal.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.ReopenDeal(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var reopened domain.DealDTO
		err := json.Unmarshal(rr.Body.Bytes(), &reopened)
		assert.NoError(t, err)
		assert.Equal(t, domain.DealStageLead, reopened.Stage)
		assert.Equal(t, 10, reopened.Probability)
		assert.Empty(t, reopened.LostReason)
		assert.Nil(t, reopened.LossReasonCategory)
	})
}

func TestDealHandler_GetStageHistory(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create a deal with history
	deal := &domain.Deal{
		Title:        "History Deal",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Currency:     "NOK",
		OwnerID:      userCtx.UserID.String(),
	}
	err := db.Create(deal).Error
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/deals/"+deal.ID.String()+"/history", nil)
	req = req.WithContext(ctx)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", deal.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.GetStageHistory(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDealHandler_GetPipelineOverview(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create deals in different stages
	stages := []domain.DealStage{domain.DealStageLead, domain.DealStageLead, domain.DealStageQualified}
	for i, stage := range stages {
		err := db.Create(&domain.Deal{
			Title:        "Deal " + string(rune('A'+i)),
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Stage:        stage,
			Currency:     "NOK",
			OwnerID:      userCtx.UserID.String(),
		}).Error
		require.NoError(t, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/deals/pipeline", nil)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetPipelineOverview(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var pipeline map[string][]domain.DealDTO
	err := json.Unmarshal(rr.Body.Bytes(), &pipeline)
	assert.NoError(t, err)
	assert.Len(t, pipeline["lead"], 2)
	assert.Len(t, pipeline["qualified"], 1)
}

func TestDealHandler_GetPipelineStats(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	ctx := createDealTestContext()

	req := httptest.NewRequest(http.MethodGet, "/deals/stats", nil)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.GetPipelineStats(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestDealHandler_GetForecast(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	ctx := createDealTestContext()

	t.Run("default months", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/forecast", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetForecast(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var forecast []repository.ForecastPeriod
		err := json.Unmarshal(rr.Body.Bytes(), &forecast)
		assert.NoError(t, err)
		assert.Len(t, forecast, 3)
	})

	t.Run("custom months", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/forecast?months=6", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetForecast(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var forecast []repository.ForecastPeriod
		err := json.Unmarshal(rr.Body.Bytes(), &forecast)
		assert.NoError(t, err)
		assert.Len(t, forecast, 6)
	})
}

func TestDealHandler_GetPipelineAnalytics(t *testing.T) {
	db := setupDealHandlerTestDB(t)
	h := createDealHandler(t, db)
	customer := createDealHandlerTestCustomer(t, db)
	ctx := createDealTestContext()
	userCtx, _ := auth.FromContext(ctx)

	t.Run("get analytics successfully with no deals", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/analytics", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		assert.NotEmpty(t, analytics.GeneratedAt)
		assert.NotNil(t, analytics.Summary)
		assert.NotNil(t, analytics.ConversionRates)
		// With no deals, win rate should have zeros
		assert.Equal(t, int64(0), analytics.WinRateAnalysis.TotalClosed)
		assert.Equal(t, float64(0), analytics.WinRateAnalysis.WinRate)
	})

	t.Run("get analytics with deals", func(t *testing.T) {
		// Create deals in different stages
		stages := []domain.DealStage{
			domain.DealStageLead,
			domain.DealStageQualified,
			domain.DealStageProposal,
		}
		for i, stage := range stages {
			err := db.Create(&domain.Deal{
				Title:        "Analytics Deal " + string(rune('A'+i)),
				CustomerID:   customer.ID,
				CustomerName: customer.Name,
				CompanyID:    domain.CompanyStalbygg,
				Stage:        stage,
				Probability:  10 + i*20,
				Value:        float64((i + 1) * 100000),
				Currency:     "NOK",
				OwnerID:      userCtx.UserID.String(),
			}).Error
			require.NoError(t, err)
		}

		req := httptest.NewRequest(http.MethodGet, "/deals/analytics", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		assert.NotEmpty(t, analytics.GeneratedAt)
		// The summary should include data from the view
		// Note: The view aggregates by stage, so we should see entries
	})

	t.Run("get analytics with company filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/analytics?companyId=stalbygg", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		assert.NotEmpty(t, analytics.GeneratedAt)
	})

	t.Run("get analytics with owner filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/analytics?ownerId="+userCtx.UserID.String(), nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		assert.NotEmpty(t, analytics.GeneratedAt)
	})

	t.Run("get analytics with date range filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/analytics?dateFrom=2024-01-01&dateTo=2025-12-31", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		assert.NotEmpty(t, analytics.GeneratedAt)
	})

	t.Run("analytics returns forecasts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/analytics", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		assert.Equal(t, "30d", analytics.Forecast30Days.Period)
		assert.Equal(t, "90d", analytics.Forecast90Days.Period)
	})

	t.Run("analytics returns conversion rates", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deals/analytics", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetPipelineAnalytics(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var analytics domain.PipelineAnalyticsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &analytics)
		assert.NoError(t, err)
		// Should have conversion rates for each stage transition
		assert.Len(t, analytics.ConversionRates, 4) // lead->qualified, qualified->proposal, proposal->negotiation, negotiation->won
	})
}
