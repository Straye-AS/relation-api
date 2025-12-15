package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func setupOfferHandlerTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
}

func createOfferHandler(t *testing.T, db *gorm.DB) *handler.OfferHandler {
	logger := zap.NewNop()
	offerRepo := repository.NewOfferRepository(db)
	offerItemRepo := repository.NewOfferItemRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	fileRepo := repository.NewFileRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	numberSequenceRepo := repository.NewNumberSequenceRepository(db)

	companyService := service.NewCompanyService(logger)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, logger)

	offerService := service.NewOfferService(
		offerRepo,
		offerItemRepo,
		customerRepo,
		projectRepo,
		budgetItemRepo,
		fileRepo,
		activityRepo,
		companyService,
		numberSequenceService,
		logger,
		db,
	)

	return handler.NewOfferHandler(offerService, logger)
}

func createOfferTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// testOfferCounter is used to generate unique offer numbers for tests
var testOfferCounter int64

func createTestOffer(t *testing.T, db *gorm.DB, customer *domain.Customer, phase domain.OfferPhase) *domain.Offer {
	testOfferCounter++

	offer := &domain.Offer{
		Title:        "Test Offer",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Phase:        phase,
		Status:       domain.OfferStatusActive,
		Probability:  50,
		Value:        100000,
	}

	// For non-draft offers, we need a unique offer number due to the unique constraint
	// idx_offers_company_offer_number ON offers(company_id, offer_number) WHERE offer_number IS NOT NULL
	if phase != domain.OfferPhaseDraft {
		offer.OfferNumber = fmt.Sprintf("TEST-%d-%d", time.Now().UnixNano(), testOfferCounter)
	}

	err := db.Create(offer).Error
	require.NoError(t, err)
	return offer
}

// withChiContext adds Chi route context with the given URL parameters
func withChiContext(ctx context.Context, params map[string]string) context.Context {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

// TestOfferHandler_List tests the List endpoint
func TestOfferHandler_List(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	// Create test offers
	for i := 0; i < 5; i++ {
		createTestOffer(t, db, customer, domain.OfferPhaseDraft)
	}

	t.Run("list all offers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, int64(5)) // At least 5 offers created in this test
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
	})

	t.Run("list with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers?page=1&pageSize=2", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, result.Total, int64(5)) // At least 5 offers
		assert.Equal(t, 2, result.PageSize)
		assert.GreaterOrEqual(t, result.TotalPages, 3) // At least 3 pages with 2 per page
	})

	t.Run("list with customer filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers?customerId="+customer.ID.String(), nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
	})

	t.Run("list with phase filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers?phase=draft", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
	})
}

// TestOfferHandler_Create tests the Create endpoint
func TestOfferHandler_Create(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("create valid offer", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		reqBody := domain.CreateOfferRequest{
			Title:             "New Offer",
			CustomerID:        &customer.ID,
			CompanyID:         domain.CompanyStalbygg,
			Phase:             domain.OfferPhaseDraft,
			Status:            domain.OfferStatusActive,
			ResponsibleUserID: userCtx.UserID.String(),
			Items: []domain.CreateOfferItemRequest{
				{
					Discipline: "Steel Work",
					Cost:       50000,
					Revenue:    75000,
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var offer domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &offer)
		assert.NoError(t, err)
		assert.Equal(t, "New Offer", offer.Title)
		assert.NotEmpty(t, rr.Header().Get("Location"))
	})

	t.Run("create with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/offers", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestOfferHandler_GetByID tests the GetByID endpoint
func TestOfferHandler_GetByID(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

	t.Run("get existing offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String(), nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferWithItemsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, offer.ID, result.ID)
	})

	t.Run("get non-existent offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+uuid.New().String(), nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/invalid-id", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": "invalid-id"}))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestOfferHandler_Update tests the Update endpoint
func TestOfferHandler_Update(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

	t.Run("update existing offer", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		reqBody := domain.UpdateOfferRequest{
			Title:             "Updated Offer",
			Phase:             domain.OfferPhaseInProgress,
			Status:            domain.OfferStatusActive,
			Probability:       75,
			ResponsibleUserID: userCtx.UserID.String(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String(), bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Offer", result.Title)
	})

	t.Run("update with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String(), bytes.NewReader([]byte("invalid")))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("update non-existent offer", func(t *testing.T) {
		userCtx, _ := auth.FromContext(ctx)
		reqBody := domain.UpdateOfferRequest{
			Title:             "Updated Offer",
			Phase:             domain.OfferPhaseInProgress,
			Status:            domain.OfferStatusActive,
			ResponsibleUserID: userCtx.UserID.String(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+uuid.New().String(), bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestOfferHandler_Delete tests the Delete endpoint
func TestOfferHandler_Delete(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("delete existing offer", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		req := httptest.NewRequest(http.MethodDelete, "/offers/"+offer.ID.String(), nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify offer is deleted
		var count int64
		db.Model(&domain.Offer{}).Where("id = ?", offer.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete non-existent offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/offers/"+uuid.New().String(), nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/offers/invalid-id", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": "invalid-id"}))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestOfferHandler_Send tests the Send endpoint
func TestOfferHandler_Send(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("send draft offer", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/send", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
	})

	t.Run("send in_progress offer", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseInProgress)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/send", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
	})

	t.Run("send already sent offer fails", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/send", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("send won offer fails", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseWon)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/send", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("send non-existent offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/offers/"+uuid.New().String()+"/send", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestOfferHandler_Accept tests the Accept endpoint
func TestOfferHandler_Accept(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("accept sent offer without project creation", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		reqBody := domain.AcceptOfferRequest{
			CreateProject: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/accept", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Accept(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.AcceptOfferResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseWon, result.Offer.Phase)
		assert.Nil(t, result.Project)
	})

	t.Run("accept sent offer with project creation", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		reqBody := domain.AcceptOfferRequest{
			CreateProject: true,
			ProjectName:   "New Project from Offer",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/accept", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Accept(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.AcceptOfferResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseWon, result.Offer.Phase)
		assert.NotNil(t, result.Project)
		assert.Equal(t, "New Project from Offer", result.Project.Name)
	})

	t.Run("accept draft offer fails", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		reqBody := domain.AcceptOfferRequest{
			CreateProject: false,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/accept", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Accept(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("accept with invalid body", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/accept", bytes.NewReader([]byte("invalid")))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Accept(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestOfferHandler_Reject tests the Reject endpoint
func TestOfferHandler_Reject(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("reject sent offer", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		reqBody := domain.RejectOfferRequest{
			Reason: "Customer chose competitor",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/reject", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Reject(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseLost, result.Phase)
	})

	t.Run("reject sent offer without reason", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		reqBody := domain.RejectOfferRequest{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/reject", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Reject(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseLost, result.Phase)
	})

	t.Run("reject draft offer fails", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		reqBody := domain.RejectOfferRequest{
			Reason: "Not applicable",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/reject", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Reject(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestOfferHandler_Clone tests the Clone endpoint
func TestOfferHandler_Clone(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("clone offer with new title", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseSent)

		includeBudget := true
		reqBody := domain.CloneOfferRequest{
			NewTitle:      "Cloned Offer",
			IncludeBudget: &includeBudget,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/clone", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Clone(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Cloned Offer", result.Title)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase) // Cloned offers start as draft
		assert.NotEqual(t, offer.ID, result.ID)
		assert.NotEmpty(t, rr.Header().Get("Location"))
	})

	t.Run("clone offer without title uses default", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		includeBudget := true
		reqBody := domain.CloneOfferRequest{
			IncludeBudget: &includeBudget,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/clone", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Clone(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Contains(t, result.Title, "Copy of")
	})

	t.Run("clone non-existent offer", func(t *testing.T) {
		reqBody := domain.CloneOfferRequest{
			NewTitle: "Cloned Offer",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+uuid.New().String()+"/clone", bytes.NewReader(body))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Clone(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("clone with invalid body", func(t *testing.T) {
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/clone", bytes.NewReader([]byte("invalid")))
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Clone(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestOfferHandler_GetBudgetSummary tests the GetBudgetSummary endpoint
func TestOfferHandler_GetBudgetSummary(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

	t.Run("get budget summary for existing offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String()+"/budget", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.GetBudgetSummary(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.BudgetSummaryDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, offer.ID, result.ParentID)
	})

	t.Run("get budget summary for non-existent offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+uuid.New().String()+"/budget", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))

		rr := httptest.NewRecorder()
		h.GetBudgetSummary(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestOfferHandler_GetWithBudgetItems tests the GetWithBudgetItems endpoint
func TestOfferHandler_GetWithBudgetItems(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

	t.Run("get offer detail for existing offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String()+"/detail", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.GetWithBudgetItems(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.OfferDetailDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, offer.ID, result.ID)
	})

	t.Run("get offer detail for non-existent offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+uuid.New().String()+"/detail", nil)
		req = req.WithContext(withChiContext(ctx, map[string]string{"id": uuid.New().String()}))

		rr := httptest.NewRecorder()
		h.GetWithBudgetItems(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestOfferHandler_Lifecycle tests the full offer lifecycle flow
func TestOfferHandler_Lifecycle(t *testing.T) {
	db := setupOfferHandlerTestDB(t)
	h := createOfferHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createOfferTestContext()

	t.Run("full lifecycle: draft -> send -> accept with project", func(t *testing.T) {
		// 1. Create offer in draft phase
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		// 2. Send the offer
		sendReq := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/send", nil)
		sendReq = sendReq.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, sendReq)
		assert.Equal(t, http.StatusOK, rr.Code)

		var sentOffer domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &sentOffer)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseSent, sentOffer.Phase)

		// 3. Accept the offer with project creation
		acceptBody := domain.AcceptOfferRequest{
			CreateProject: true,
			ProjectName:   "Project from Lifecycle Test",
		}
		bodyBytes, _ := json.Marshal(acceptBody)

		acceptReq := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/accept", bytes.NewReader(bodyBytes))
		acceptReq = acceptReq.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		acceptReq.Header.Set("Content-Type", "application/json")

		rr = httptest.NewRecorder()
		h.Accept(rr, acceptReq)
		assert.Equal(t, http.StatusOK, rr.Code)

		var acceptResponse domain.AcceptOfferResponse
		err = json.Unmarshal(rr.Body.Bytes(), &acceptResponse)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseWon, acceptResponse.Offer.Phase)
		assert.NotNil(t, acceptResponse.Project)
		assert.Equal(t, "Project from Lifecycle Test", acceptResponse.Project.Name)
	})

	t.Run("full lifecycle: draft -> send -> reject", func(t *testing.T) {
		// 1. Create offer in draft phase
		offer := createTestOffer(t, db, customer, domain.OfferPhaseDraft)

		// 2. Send the offer
		sendReq := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/send", nil)
		sendReq = sendReq.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.Send(rr, sendReq)
		assert.Equal(t, http.StatusOK, rr.Code)

		// 3. Reject the offer
		rejectBody := domain.RejectOfferRequest{
			Reason: "Customer found a better price",
		}
		bodyBytes, _ := json.Marshal(rejectBody)

		rejectReq := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/reject", bytes.NewReader(bodyBytes))
		rejectReq = rejectReq.WithContext(withChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		rejectReq.Header.Set("Content-Type", "application/json")

		rr = httptest.NewRecorder()
		h.Reject(rr, rejectReq)
		assert.Equal(t, http.StatusOK, rr.Code)

		var rejectedOffer domain.OfferDTO
		err := json.Unmarshal(rr.Body.Bytes(), &rejectedOffer)
		assert.NoError(t, err)
		assert.Equal(t, domain.OfferPhaseLost, rejectedOffer.Phase)
	})
}
