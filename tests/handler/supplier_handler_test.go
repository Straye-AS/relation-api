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

func setupSupplierHandlerTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
}

func createSupplierHandler(t *testing.T, db *gorm.DB) *handler.SupplierHandler {
	logger := zap.NewNop()
	supplierRepo := repository.NewSupplierRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	supplierService := service.NewSupplierService(supplierRepo, activityRepo, logger)

	return handler.NewSupplierHandler(supplierService, logger)
}

func createSupplierTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// TestSupplierHandler_List tests the List endpoint
func TestSupplierHandler_List(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create test suppliers
	suppliers := []struct {
		name     string
		city     string
		country  string
		category string
		status   domain.SupplierStatus
	}{
		{"Alpha Steel AS", "Oslo", "Norway", "steel", domain.SupplierStatusActive},
		{"Beta Windows AS", "Bergen", "Norway", "windows", domain.SupplierStatusActive},
		{"Gamma Concrete Ltd", "Stockholm", "Sweden", "concrete", domain.SupplierStatusInactive},
		{"Delta Materials AS", "Oslo", "Norway", "materials", domain.SupplierStatusActive},
		{"Epsilon Supply GmbH", "Munich", "Germany", "steel", domain.SupplierStatusPending},
	}
	for _, s := range suppliers {
		supplier := testutil.CreateTestSupplierWithCategory(t, db, s.name, s.category)
		db.Model(&domain.Supplier{}).Where("id = ?", supplier.ID).Updates(map[string]interface{}{
			"city":    s.city,
			"country": s.country,
			"status":  s.status,
		})
	}

	t.Run("list all suppliers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
	})

	t.Run("list with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?page=1&pageSize=2", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 2, result.PageSize)
		assert.Equal(t, 3, result.TotalPages)
	})

	t.Run("list with search filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?search=Alpha", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
	})

	t.Run("list with city filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?city=Oslo", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
	})

	t.Run("list with country filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?country=Norway", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?status=active", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total) // Alpha, Beta, Delta
	})

	t.Run("list with category filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?category=steel", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Total) // Alpha, Epsilon
	})

	t.Run("list with sort by name asc", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers?sortBy=name&sortOrder=asc", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)

		// Check first supplier is Alpha (alphabetically first)
		data := result.Data.([]interface{})
		firstSupplier := data[0].(map[string]interface{})
		assert.Equal(t, "Alpha Steel AS", firstSupplier["name"])
	})
}

// TestSupplierHandler_GetByID tests the GetByID endpoint
func TestSupplierHandler_GetByID(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create a test supplier
	supplier := testutil.CreateTestSupplier(t, db, "Test Supplier")

	t.Run("get existing supplier", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers/"+supplier.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierWithDetailsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, supplier.ID, result.ID)
		assert.Equal(t, "Test Supplier", result.Name)
		assert.NotNil(t, result.Stats)
	})

	t.Run("get non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/suppliers/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/suppliers/invalid-id", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-id")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Bad Request", errResp.Error)
	})
}

// TestSupplierHandler_Create tests the Create endpoint
func TestSupplierHandler_Create(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	t.Run("create valid supplier", func(t *testing.T) {
		reqBody := domain.CreateSupplierRequest{
			Name:         "New Supplier AS",
			OrgNumber:    "123456789",
			Email:        "contact@newsupplier.no",
			Phone:        "+47 12345678",
			Address:      "Supplier Street 1",
			City:         "Oslo",
			PostalCode:   "0150",
			Country:      "Norway",
			Category:     "steel",
			PaymentTerms: "Net 30",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Location"))

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "New Supplier AS", result.Name)
		assert.Equal(t, "123456789", result.OrgNumber)
		assert.Equal(t, domain.SupplierStatusActive, result.Status) // Default status
		assert.NotEqual(t, uuid.Nil, result.ID)
	})

	t.Run("create with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with missing required fields", func(t *testing.T) {
		reqBody := domain.CreateSupplierRequest{
			Name: "Incomplete Supplier",
			// Missing required field: Country
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with duplicate org number", func(t *testing.T) {
		// Create first supplier
		reqBody := domain.CreateSupplierRequest{
			Name:      "First Supplier",
			OrgNumber: "999888777",
			Country:   "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)

		// Try to create second supplier with same org number
		reqBody2 := domain.CreateSupplierRequest{
			Name:      "Second Supplier",
			OrgNumber: "999888777", // Same org number
			Country:   "Norway",
		}
		body2, _ := json.Marshal(reqBody2)

		req2 := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body2))
		req2 = req2.WithContext(ctx)
		req2.Header.Set("Content-Type", "application/json")

		rr2 := httptest.NewRecorder()
		h.Create(rr2, req2)

		assert.Equal(t, http.StatusConflict, rr2.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(rr2.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Conflict", errResp.Error)
	})

	t.Run("create with specific status", func(t *testing.T) {
		reqBody := domain.CreateSupplierRequest{
			Name:    "Pending Supplier",
			Country: "Norway",
			Status:  domain.SupplierStatusPending,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.SupplierStatusPending, result.Status)
	})
}

// TestSupplierHandler_Update tests the Update endpoint
func TestSupplierHandler_Update(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create a test supplier
	supplier := testutil.CreateTestSupplier(t, db, "Original Supplier")

	t.Run("update supplier successfully", func(t *testing.T) {
		reqBody := domain.UpdateSupplierRequest{
			Name:     "Updated Supplier Name",
			Country:  "Sweden",
			Category: "windows",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Supplier Name", result.Name)
		assert.Equal(t, "Sweden", result.Country)
		assert.Equal(t, "windows", result.Category)
	})

	t.Run("update non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateSupplierRequest{
			Name:    "Updated Name",
			Country: "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+nonExistentID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update with invalid ID", func(t *testing.T) {
		reqBody := domain.UpdateSupplierRequest{
			Name:    "Test",
			Country: "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/invalid", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("update with validation error", func(t *testing.T) {
		reqBody := domain.UpdateSupplierRequest{
			// Missing required fields: Name, Country
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestSupplierHandler_Delete tests the Delete endpoint
func TestSupplierHandler_Delete(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	t.Run("delete supplier successfully", func(t *testing.T) {
		// Create a supplier to delete
		supplier := testutil.CreateTestSupplier(t, db, "Supplier To Delete")

		req := httptest.NewRequest(http.MethodDelete, "/suppliers/"+supplier.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify supplier is soft deleted (query with Unscoped to find deleted records)
		var found domain.Supplier
		err := db.Unscoped().Where("id = ?", supplier.ID).First(&found).Error
		assert.NoError(t, err)
		assert.NotNil(t, found.DeletedAt)
	})

	t.Run("delete non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/suppliers/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/suppliers/invalid", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestSupplierHandler_UpdateStatus tests the UpdateStatus endpoint
func TestSupplierHandler_UpdateStatus(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create a test supplier
	supplier := testutil.CreateTestSupplier(t, db, "Status Test Supplier")

	t.Run("update status successfully", func(t *testing.T) {
		reqBody := domain.UpdateSupplierStatusRequest{
			Status: domain.SupplierStatusInactive,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.SupplierStatusInactive, result.Status)
	})

	t.Run("update status with invalid status value", func(t *testing.T) {
		body := []byte(`{"status": "invalid_status"}`)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("update status for non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateSupplierStatusRequest{
			Status: domain.SupplierStatusInactive,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+nonExistentID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("update to blacklisted status", func(t *testing.T) {
		reqBody := domain.UpdateSupplierStatusRequest{
			Status: domain.SupplierStatusBlacklisted,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.SupplierStatusBlacklisted, result.Status)
	})
}

// TestSupplierHandler_UpdateNotes tests the UpdateNotes endpoint
func TestSupplierHandler_UpdateNotes(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create a test supplier
	supplier := testutil.CreateTestSupplier(t, db, "Notes Test Supplier")

	t.Run("update notes successfully", func(t *testing.T) {
		reqBody := domain.UpdateSupplierNotesRequest{
			Notes: "This is a new note for the supplier.",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/notes", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateNotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "This is a new note for the supplier.", result.Notes)
	})

	t.Run("clear notes", func(t *testing.T) {
		reqBody := domain.UpdateSupplierNotesRequest{
			Notes: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/notes", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateNotes(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "", result.Notes)
	})

	t.Run("update notes for non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateSupplierNotesRequest{
			Notes: "Some notes",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+nonExistentID.String()+"/notes", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateNotes(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestSupplierHandler_UpdateCategory tests the UpdateCategory endpoint
func TestSupplierHandler_UpdateCategory(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create a test supplier
	supplier := testutil.CreateTestSupplier(t, db, "Category Test Supplier")

	t.Run("update category successfully", func(t *testing.T) {
		reqBody := domain.UpdateSupplierCategoryRequest{
			Category: "steel, windows",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/category", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "steel, windows", result.Category)
	})

	t.Run("clear category", func(t *testing.T) {
		reqBody := domain.UpdateSupplierCategoryRequest{
			Category: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/category", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "", result.Category)
	})

	t.Run("update category for non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateSupplierCategoryRequest{
			Category: "concrete",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+nonExistentID.String()+"/category", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateCategory(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestSupplierHandler_UpdatePaymentTerms tests the UpdatePaymentTerms endpoint
func TestSupplierHandler_UpdatePaymentTerms(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	// Create a test supplier
	supplier := testutil.CreateTestSupplier(t, db, "Payment Terms Test Supplier")

	t.Run("update payment terms successfully", func(t *testing.T) {
		reqBody := domain.UpdateSupplierPaymentTermsRequest{
			PaymentTerms: "Net 60",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/payment-terms", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdatePaymentTerms(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Net 60", result.PaymentTerms)
	})

	t.Run("clear payment terms", func(t *testing.T) {
		reqBody := domain.UpdateSupplierPaymentTermsRequest{
			PaymentTerms: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+supplier.ID.String()+"/payment-terms", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", supplier.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdatePaymentTerms(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.SupplierDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "", result.PaymentTerms)
	})

	t.Run("update payment terms for non-existent supplier", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateSupplierPaymentTermsRequest{
			PaymentTerms: "Net 30",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/suppliers/"+nonExistentID.String()+"/payment-terms", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdatePaymentTerms(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestSupplierHandler_Create_ValidationError tests validation errors during create
func TestSupplierHandler_Create_ValidationError(t *testing.T) {
	db := setupSupplierHandlerTestDB(t)
	h := createSupplierHandler(t, db)
	ctx := createSupplierTestContext()

	t.Run("create with missing name", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"country": "Norway",
			// Name is missing
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with invalid email format", func(t *testing.T) {
		reqBody := domain.CreateSupplierRequest{
			Name:    "Bad Email Supplier",
			Country: "Norway",
			Email:   "not-an-email",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with name exceeding max length", func(t *testing.T) {
		// Create a name that exceeds 200 characters
		longName := ""
		for i := 0; i < 250; i++ {
			longName += "x"
		}

		reqBody := domain.CreateSupplierRequest{
			Name:    longName,
			Country: "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/suppliers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
