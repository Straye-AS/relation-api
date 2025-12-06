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

func setupCustomerHandlerTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createCustomerHandler(t *testing.T, db *gorm.DB) *handler.CustomerHandler {
	logger := zap.NewNop()
	customerRepo := repository.NewCustomerRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	contactRepo := repository.NewContactRepository(db)

	customerService := service.NewCustomerService(customerRepo, activityRepo, logger)
	contactService := service.NewContactService(contactRepo, customerRepo, activityRepo, logger)

	return handler.NewCustomerHandler(customerService, contactService, logger)
}

func createCustomerTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// TestCustomerHandler_List tests the List endpoint
func TestCustomerHandler_List(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	// Create test customers
	customers := []struct {
		name    string
		city    string
		country string
	}{
		{"Alpha Corp", "Oslo", "Norway"},
		{"Beta Inc", "Bergen", "Norway"},
		{"Gamma Ltd", "Stockholm", "Sweden"},
		{"Delta AS", "Oslo", "Norway"},
		{"Epsilon GmbH", "Munich", "Germany"},
	}
	for _, c := range customers {
		testutil.CreateTestCustomer(t, db, c.name)
		// Update with city and country
		db.Model(&domain.Customer{}).Where("name = ?", c.name).Updates(map[string]interface{}{
			"city":    c.city,
			"country": c.country,
		})
	}

	t.Run("list all customers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers", nil)
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
		req := httptest.NewRequest(http.MethodGet, "/customers?page=1&pageSize=2", nil)
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
		req := httptest.NewRequest(http.MethodGet, "/customers?search=Alpha", nil)
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
		req := httptest.NewRequest(http.MethodGet, "/customers?city=Oslo", nil)
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
		req := httptest.NewRequest(http.MethodGet, "/customers?country=Norway", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list with sort by name asc", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers?sortBy=name_asc", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)

		// Check first customer is Alpha (alphabetically first)
		data := result.Data.([]interface{})
		firstCustomer := data[0].(map[string]interface{})
		assert.Equal(t, "Alpha Corp", firstCustomer["name"])
	})
}

// TestCustomerHandler_GetByID tests the GetByID endpoint
func TestCustomerHandler_GetByID(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	// Create a test customer
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")

	t.Run("get existing customer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/"+customer.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.CustomerWithDetailsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, customer.ID, result.ID)
		assert.Equal(t, "Test Customer", result.Name)
		assert.NotNil(t, result.Stats)
	})

	t.Run("get non-existent customer", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/customers/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/invalid-id", nil)
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

// TestCustomerHandler_Create tests the Create endpoint
func TestCustomerHandler_Create(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	t.Run("create valid customer", func(t *testing.T) {
		reqBody := domain.CreateCustomerRequest{
			Name:       "New Customer AS",
			OrgNumber:  "123456789",
			Email:      "contact@newcustomer.no",
			Phone:      "+47 12345678",
			Address:    "Test Street 1",
			City:       "Oslo",
			PostalCode: "0150",
			Country:    "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Location"))

		var result domain.CustomerDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "New Customer AS", result.Name)
		assert.Equal(t, "123456789", result.OrgNumber)
		assert.NotEqual(t, uuid.Nil, result.ID)
	})

	t.Run("create with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with missing required fields", func(t *testing.T) {
		reqBody := domain.CreateCustomerRequest{
			Name: "Incomplete Customer",
			// Missing required fields: OrgNumber, Email, Phone, Country
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with duplicate org number", func(t *testing.T) {
		// Create first customer
		reqBody := domain.CreateCustomerRequest{
			Name:      "First Customer",
			OrgNumber: "999888777",
			Email:     "first@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)

		// Try to create second customer with same org number
		reqBody2 := domain.CreateCustomerRequest{
			Name:      "Second Customer",
			OrgNumber: "999888777", // Same org number
			Email:     "second@example.com",
			Phone:     "87654321",
			Country:   "Norway",
		}
		body2, _ := json.Marshal(reqBody2)

		req2 := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewReader(body2))
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
}

// TestCustomerHandler_Update tests the Update endpoint
func TestCustomerHandler_Update(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	// Create a test customer
	customer := testutil.CreateTestCustomer(t, db, "Original Customer")

	t.Run("update customer successfully", func(t *testing.T) {
		reqBody := domain.UpdateCustomerRequest{
			Name:      "Updated Customer Name",
			OrgNumber: customer.OrgNumber,
			Email:     "updated@example.com",
			Phone:     "99887766",
			Country:   "Sweden",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/customers/"+customer.ID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.CustomerDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Customer Name", result.Name)
		assert.Equal(t, "Sweden", result.Country)
	})

	t.Run("update non-existent customer", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateCustomerRequest{
			Name:      "Updated Name",
			OrgNumber: "111222333",
			Email:     "test@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/customers/"+nonExistentID.String(), bytes.NewReader(body))
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
		reqBody := domain.UpdateCustomerRequest{
			Name: "Test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/customers/invalid", bytes.NewReader(body))
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
		reqBody := domain.UpdateCustomerRequest{
			// Missing required fields
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/customers/"+customer.ID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestCustomerHandler_Delete tests the Delete endpoint
func TestCustomerHandler_Delete(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	t.Run("delete customer successfully", func(t *testing.T) {
		// Create a customer to delete
		customer := testutil.CreateTestCustomer(t, db, "Customer To Delete")

		req := httptest.NewRequest(http.MethodDelete, "/customers/"+customer.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify customer is deleted
		var found domain.Customer
		err := db.Where("id = ?", customer.ID).First(&found).Error
		assert.Error(t, err) // Should not be found
	})

	t.Run("delete non-existent customer", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/customers/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/customers/invalid", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("delete customer with active projects returns conflict", func(t *testing.T) {
		// Create a customer with an active project
		customer := testutil.CreateTestCustomer(t, db, "Customer With Project")

		// Create an active project for this customer
		userCtx, _ := auth.FromContext(ctx)
		project := &domain.Project{
			Name:         "Active Project",
			CustomerID:   customer.ID,
			CustomerName: customer.Name,
			CompanyID:    domain.CompanyStalbygg,
			Status:       domain.ProjectStatusActive,
			ManagerID:    userCtx.UserID.String(),
			Budget:       100000,
		}
		err := db.Create(project).Error
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodDelete, "/customers/"+customer.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)

		var errResp domain.ErrorResponse
		err = json.Unmarshal(rr.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, "Conflict", errResp.Error)
	})
}

// TestCustomerHandler_ListContacts tests the ListContacts endpoint
func TestCustomerHandler_ListContacts(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	// Create a test customer
	customer := testutil.CreateTestCustomer(t, db, "Customer With Contacts")

	// Create some contacts for the customer
	contacts := []domain.Contact{
		{
			FirstName:         "John",
			LastName:          "Doe",
			Email:             "john@example.com",
			PrimaryCustomerID: &customer.ID,
			IsActive:          true,
		},
		{
			FirstName:         "Jane",
			LastName:          "Smith",
			Email:             "jane@example.com",
			PrimaryCustomerID: &customer.ID,
			IsActive:          true,
		},
	}
	for _, c := range contacts {
		err := db.Create(&c).Error
		require.NoError(t, err)
	}

	t.Run("list customer contacts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/"+customer.ID.String()+"/contacts", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.ListContacts(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []domain.ContactDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("list contacts with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/customers/invalid/contacts", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.ListContacts(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestCustomerHandler_CreateContact tests the CreateContact endpoint
func TestCustomerHandler_CreateContact(t *testing.T) {
	db := setupCustomerHandlerTestDB(t)
	h := createCustomerHandler(t, db)
	ctx := createCustomerTestContext()

	// Create a test customer
	customer := testutil.CreateTestCustomer(t, db, "Customer For Contact")

	t.Run("create contact successfully", func(t *testing.T) {
		reqBody := domain.CreateContactRequest{
			FirstName: "New",
			LastName:  "Contact",
			Email:     "new.contact@example.com",
			Phone:     "12345678",
			Title:     "Manager",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/contacts", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.CreateContact(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result domain.ContactDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "New", result.FirstName)
		assert.Equal(t, "Contact", result.LastName)
		assert.Equal(t, customer.ID, *result.PrimaryCustomerID)
	})

	t.Run("create contact with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/contacts", bytes.NewReader([]byte("invalid")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.CreateContact(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create contact with missing required fields", func(t *testing.T) {
		reqBody := domain.CreateContactRequest{
			FirstName: "", // Required but empty
			LastName:  "", // Required but empty
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/customers/"+customer.ID.String()+"/contacts", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", customer.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.CreateContact(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
