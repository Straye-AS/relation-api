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

func setupBudgetDimensionHandlerTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createBudgetDimensionHandler(t *testing.T, db *gorm.DB) *handler.BudgetDimensionHandler {
	logger := zap.NewNop()
	budgetDimensionRepo := repository.NewBudgetDimensionRepository(db)
	budgetDimensionCategoryRepo := repository.NewBudgetDimensionCategoryRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	budgetDimensionService := service.NewBudgetDimensionService(
		budgetDimensionRepo,
		budgetDimensionCategoryRepo,
		offerRepo,
		projectRepo,
		activityRepo,
		logger,
	)

	return handler.NewBudgetDimensionHandler(budgetDimensionService, logger)
}

func createBudgetDimensionTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func createTestOfferForBudget(t *testing.T, db *gorm.DB, customer *domain.Customer) *domain.Offer {
	offer := &domain.Offer{
		Title:        "Test Offer for Budget",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Phase:        domain.OfferPhaseDraft,
		Status:       domain.OfferStatusActive,
		Probability:  50,
		Value:        0,
	}
	err := db.Create(offer).Error
	require.NoError(t, err)
	return offer
}

func createTestBudgetDimension(t *testing.T, db *gorm.DB, offer *domain.Offer, customName string, cost, revenue float64) *domain.BudgetDimension {
	dim := &domain.BudgetDimension{
		ParentType:   domain.BudgetParentOffer,
		ParentID:     offer.ID,
		CustomName:   customName,
		Cost:         cost,
		Revenue:      revenue,
		DisplayOrder: 0,
	}
	err := db.Create(dim).Error
	require.NoError(t, err)
	return dim
}

// withBudgetChiContext adds Chi route context with the given URL parameters
func withBudgetChiContext(ctx context.Context, params map[string]string) context.Context {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

// TestBudgetDimensionHandler_ListOfferDimensions tests the ListOfferDimensions endpoint
func TestBudgetDimensionHandler_ListOfferDimensions(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)

	// Create test dimensions
	createTestBudgetDimension(t, db, offer, "Steel Work", 50000, 75000)
	createTestBudgetDimension(t, db, offer, "Concrete Work", 30000, 45000)
	createTestBudgetDimension(t, db, offer, "Electrical", 20000, 30000)

	t.Run("list all dimensions for offer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String()+"/budget/dimensions", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.ListOfferDimensions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
	})

	t.Run("list with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String()+"/budget/dimensions?page=1&pageSize=2", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.ListOfferDimensions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
		assert.Equal(t, 2, result.PageSize)
		assert.Equal(t, 2, result.TotalPages)
	})

	t.Run("list with invalid offer ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/invalid-id/budget/dimensions", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": "invalid-id"}))

		rr := httptest.NewRecorder()
		h.ListOfferDimensions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestBudgetDimensionHandler_AddToOffer tests the AddToOffer endpoint
func TestBudgetDimensionHandler_AddToOffer(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)

	t.Run("add dimension with custom name", func(t *testing.T) {
		reqBody := domain.AddOfferBudgetDimensionRequest{
			CustomName: "Steel Structure",
			Cost:       50000,
			Revenue:    75000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.BudgetDimensionDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Steel Structure", result.Name)
		assert.Equal(t, 50000.0, result.Cost)
		assert.Equal(t, 75000.0, result.Revenue)
		assert.Equal(t, offer.ID, result.ParentID)
		assert.NotEmpty(t, rr.Header().Get("Location"))
	})

	t.Run("add dimension with margin override", func(t *testing.T) {
		targetMargin := 25.0
		reqBody := domain.AddOfferBudgetDimensionRequest{
			CustomName:          "High Margin Work",
			Cost:                100000,
			TargetMarginPercent: &targetMargin,
			MarginOverride:      true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.BudgetDimensionDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "High Margin Work", result.Name)
		assert.True(t, result.MarginOverride)
		assert.NotNil(t, result.TargetMarginPercent)
		assert.Equal(t, 25.0, *result.TargetMarginPercent)
	})

	t.Run("add dimension without name fails", func(t *testing.T) {
		reqBody := domain.AddOfferBudgetDimensionRequest{
			Cost:    50000,
			Revenue: 75000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("add dimension with zero cost fails", func(t *testing.T) {
		reqBody := domain.AddOfferBudgetDimensionRequest{
			CustomName: "Zero Cost Work",
			Cost:       0,
			Revenue:    75000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("add dimension with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("add dimension to non-existent offer", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.AddOfferBudgetDimensionRequest{
			CustomName: "Steel Structure",
			Cost:       50000,
			Revenue:    75000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+nonExistentID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": nonExistentID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestBudgetDimensionHandler_UpdateOfferDimension tests the UpdateOfferDimension endpoint
func TestBudgetDimensionHandler_UpdateOfferDimension(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)
	dimension := createTestBudgetDimension(t, db, offer, "Original Name", 50000, 75000)

	t.Run("update dimension successfully", func(t *testing.T) {
		reqBody := domain.UpdateBudgetDimensionRequest{
			CustomName: "Updated Name",
			Cost:       60000,
			Revenue:    90000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/dimensions/"+dimension.ID.String(), bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": dimension.ID.String(),
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.UpdateOfferDimension(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.BudgetDimensionDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", result.Name)
		assert.Equal(t, 60000.0, result.Cost)
		assert.Equal(t, 90000.0, result.Revenue)
	})

	t.Run("update non-existent dimension", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateBudgetDimensionRequest{
			CustomName: "Updated Name",
			Cost:       60000,
			Revenue:    90000,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/dimensions/"+nonExistentID.String(), bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": nonExistentID.String(),
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.UpdateOfferDimension(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update dimension from different offer fails", func(t *testing.T) {
		// Create another offer
		otherOffer := createTestOfferForBudget(t, db, customer)
		otherDimension := createTestBudgetDimension(t, db, otherOffer, "Other Dimension", 10000, 15000)

		reqBody := domain.UpdateBudgetDimensionRequest{
			CustomName: "Updated Name",
			Cost:       60000,
			Revenue:    90000,
		}
		body, _ := json.Marshal(reqBody)

		// Try to update otherDimension via the first offer's route
		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/dimensions/"+otherDimension.ID.String(), bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": otherDimension.ID.String(),
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.UpdateOfferDimension(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/dimensions/"+dimension.ID.String(), bytes.NewReader([]byte("invalid")))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": dimension.ID.String(),
		}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.UpdateOfferDimension(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestBudgetDimensionHandler_DeleteOfferDimension tests the DeleteOfferDimension endpoint
func TestBudgetDimensionHandler_DeleteOfferDimension(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)

	t.Run("delete dimension successfully", func(t *testing.T) {
		dimension := createTestBudgetDimension(t, db, offer, "To Be Deleted", 50000, 75000)

		req := httptest.NewRequest(http.MethodDelete, "/offers/"+offer.ID.String()+"/budget/dimensions/"+dimension.ID.String(), nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": dimension.ID.String(),
		}))

		rr := httptest.NewRecorder()
		h.DeleteOfferDimension(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify dimension is deleted
		var count int64
		db.Model(&domain.BudgetDimension{}).Where("id = ?", dimension.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete non-existent dimension", func(t *testing.T) {
		nonExistentID := uuid.New()

		req := httptest.NewRequest(http.MethodDelete, "/offers/"+offer.ID.String()+"/budget/dimensions/"+nonExistentID.String(), nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": nonExistentID.String(),
		}))

		rr := httptest.NewRecorder()
		h.DeleteOfferDimension(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete dimension from different offer fails", func(t *testing.T) {
		otherOffer := createTestOfferForBudget(t, db, customer)
		otherDimension := createTestBudgetDimension(t, db, otherOffer, "Other Dimension", 10000, 15000)

		req := httptest.NewRequest(http.MethodDelete, "/offers/"+offer.ID.String()+"/budget/dimensions/"+otherDimension.ID.String(), nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": otherDimension.ID.String(),
		}))

		rr := httptest.NewRecorder()
		h.DeleteOfferDimension(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid dimension ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/offers/"+offer.ID.String()+"/budget/dimensions/invalid-id", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{
			"id":          offer.ID.String(),
			"dimensionId": "invalid-id",
		}))

		rr := httptest.NewRecorder()
		h.DeleteOfferDimension(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestBudgetDimensionHandler_ReorderOfferDimensions tests the ReorderOfferDimensions endpoint
func TestBudgetDimensionHandler_ReorderOfferDimensions(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)

	// Create dimensions in a specific order
	dim1 := createTestBudgetDimension(t, db, offer, "First", 10000, 15000)
	dim2 := createTestBudgetDimension(t, db, offer, "Second", 20000, 30000)
	dim3 := createTestBudgetDimension(t, db, offer, "Third", 30000, 45000)

	t.Run("reorder dimensions successfully", func(t *testing.T) {
		// Reverse the order
		reqBody := domain.ReorderDimensionsRequest{
			OrderedIDs: []uuid.UUID{dim3.ID, dim2.ID, dim1.ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/reorder", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ReorderOfferDimensions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []domain.BudgetDimensionDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		// Verify the new order
		assert.Equal(t, dim3.ID, result[0].ID)
		assert.Equal(t, dim2.ID, result[1].ID)
		assert.Equal(t, dim1.ID, result[2].ID)
	})

	t.Run("reorder with wrong count fails", func(t *testing.T) {
		// Only provide 2 IDs when there are 3 dimensions
		reqBody := domain.ReorderDimensionsRequest{
			OrderedIDs: []uuid.UUID{dim1.ID, dim2.ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/reorder", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ReorderOfferDimensions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("reorder with empty list fails validation", func(t *testing.T) {
		reqBody := domain.ReorderDimensionsRequest{
			OrderedIDs: []uuid.UUID{},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/reorder", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ReorderOfferDimensions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("reorder for non-existent offer fails", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.ReorderDimensionsRequest{
			OrderedIDs: []uuid.UUID{dim1.ID, dim2.ID, dim3.ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/offers/"+nonExistentID.String()+"/budget/reorder", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": nonExistentID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ReorderOfferDimensions(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("reorder with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/offers/"+offer.ID.String()+"/budget/reorder", bytes.NewReader([]byte("invalid")))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ReorderOfferDimensions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestBudgetDimensionHandler_GetOfferBudgetWithDimensions tests the GetOfferBudgetWithDimensions endpoint
func TestBudgetDimensionHandler_GetOfferBudgetWithDimensions(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)

	// Create test dimensions
	createTestBudgetDimension(t, db, offer, "Steel Work", 50000, 75000)
	createTestBudgetDimension(t, db, offer, "Concrete Work", 30000, 45000)

	t.Run("get budget with dimensions and summary", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/"+offer.ID.String()+"/budget", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))

		rr := httptest.NewRecorder()
		h.GetOfferBudgetWithDimensions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result handler.OfferBudgetResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)

		// Check dimensions
		assert.Len(t, result.Dimensions, 2)

		// Check summary
		assert.NotNil(t, result.Summary)
		assert.Equal(t, offer.ID, result.Summary.ParentID)
		assert.Equal(t, 2, result.Summary.DimensionCount)
		assert.Equal(t, 80000.0, result.Summary.TotalCost)     // 50000 + 30000
		assert.Equal(t, 120000.0, result.Summary.TotalRevenue) // 75000 + 45000
		assert.Equal(t, 40000.0, result.Summary.TotalProfit)   // 120000 - 80000
	})

	t.Run("get budget for offer with no dimensions", func(t *testing.T) {
		emptyOffer := createTestOfferForBudget(t, db, customer)

		req := httptest.NewRequest(http.MethodGet, "/offers/"+emptyOffer.ID.String()+"/budget", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": emptyOffer.ID.String()}))

		rr := httptest.NewRecorder()
		h.GetOfferBudgetWithDimensions(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result handler.OfferBudgetResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)

		assert.Len(t, result.Dimensions, 0)
		assert.NotNil(t, result.Summary)
		assert.Equal(t, 0, result.Summary.DimensionCount)
		assert.Equal(t, 0.0, result.Summary.TotalCost)
		assert.Equal(t, 0.0, result.Summary.TotalRevenue)
	})

	t.Run("get budget with invalid offer ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/offers/invalid-id/budget", nil)
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": "invalid-id"}))

		rr := httptest.NewRecorder()
		h.GetOfferBudgetWithDimensions(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestBudgetDimensionHandler_ErrorHandling tests error mapping for various scenarios
func TestBudgetDimensionHandler_ErrorHandling(t *testing.T) {
	db := setupBudgetDimensionHandlerTestDB(t)
	h := createBudgetDimensionHandler(t, db)
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	ctx := createBudgetDimensionTestContext()

	offer := createTestOfferForBudget(t, db, customer)

	t.Run("invalid target margin percent returns 400", func(t *testing.T) {
		// Margin override without target margin
		reqBody := domain.AddOfferBudgetDimensionRequest{
			CustomName:     "Invalid Margin",
			Cost:           50000,
			MarginOverride: true,
			// Missing TargetMarginPercent
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("margin override with out of range target margin returns 400", func(t *testing.T) {
		invalidMargin := 150.0 // > 100%
		reqBody := domain.AddOfferBudgetDimensionRequest{
			CustomName:          "Invalid Margin",
			Cost:                50000,
			TargetMarginPercent: &invalidMargin,
			MarginOverride:      true,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/offers/"+offer.ID.String()+"/budget/dimensions", bytes.NewReader(body))
		req = req.WithContext(withBudgetChiContext(ctx, map[string]string{"id": offer.ID.String()}))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.AddToOffer(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
