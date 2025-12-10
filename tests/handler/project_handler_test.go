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

// Counter for generating unique offer numbers in tests
var testProjectOfferCounter int64

func setupProjectHandlerTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupCleanTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createProjectHandler(t *testing.T, db *gorm.DB) *handler.ProjectHandler {
	logger := zap.NewNop()
	projectRepo := repository.NewProjectRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	offerItemRepo := repository.NewOfferItemRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	fileRepo := repository.NewFileRepository(db)
	numberSequenceRepo := repository.NewNumberSequenceRepository(db)

	companyService := service.NewCompanyService(logger)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, logger)

	projectService := service.NewProjectService(projectRepo, customerRepo, activityRepo, logger)
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

	return handler.NewProjectHandler(projectService, offerService, logger)
}

// createProjectHandlerWithDeps creates a handler with all dependencies for full feature support
// This is needed for tests that require offer/budget dimension support
func createProjectHandlerWithDeps(t *testing.T, db *gorm.DB) *handler.ProjectHandler {
	logger := zap.NewNop()
	projectRepo := repository.NewProjectRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	offerItemRepo := repository.NewOfferItemRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	fileRepo := repository.NewFileRepository(db)
	numberSequenceRepo := repository.NewNumberSequenceRepository(db)

	companyService := service.NewCompanyService(logger)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, logger)

	projectService := service.NewProjectServiceWithDeps(
		projectRepo,
		offerRepo,
		customerRepo,
		budgetItemRepo,
		activityRepo,
		companyService,
		numberSequenceService,
		logger,
		db,
	)

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

	return handler.NewProjectHandler(projectService, offerService, logger)
}

func createProjectTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func createTestProject(t *testing.T, db *gorm.DB, customer *domain.Customer, name string, status domain.ProjectStatus, managerID string) *domain.Project {
	project := &domain.Project{
		Name:         name,
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Status:       status,
		Phase:        domain.ProjectPhaseActive, // Set to active to allow budget updates in tests
		StartDate:    time.Now(),
		Budget:       100000,
		ManagerID:    managerID,
	}
	err := db.Create(project).Error
	require.NoError(t, err)
	return project
}

// TestProjectHandler_List tests the List endpoint with various filters
func TestProjectHandler_List(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create test customers
	customer1 := testutil.CreateTestCustomer(t, db, "Customer One")
	customer2 := testutil.CreateTestCustomer(t, db, "Customer Two")

	// Create test projects with various statuses and health values
	projects := []struct {
		customer  *domain.Customer
		name      string
		status    domain.ProjectStatus
		managerID string
	}{
		{customer1, "Alpha Project", domain.ProjectStatusActive, userCtx.UserID.String()},
		{customer1, "Beta Project", domain.ProjectStatusPlanning, userCtx.UserID.String()},
		{customer1, "Gamma Project", domain.ProjectStatusCompleted, "other-manager"},
		{customer2, "Delta Project", domain.ProjectStatusActive, userCtx.UserID.String()},
		{customer2, "Epsilon Project", domain.ProjectStatusOnHold, "other-manager"},
	}

	for _, p := range projects {
		createTestProject(t, db, p.customer, p.name, p.status, p.managerID)
	}

	t.Run("list all projects", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
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
		req := httptest.NewRequest(http.MethodGet, "/projects?page=1&pageSize=2", nil)
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

	t.Run("list with status filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects?status=active", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Total)
	})

	t.Run("list with customer filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects?customerId="+customer1.ID.String(), nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list with manager filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects?managerId="+userCtx.UserID.String(), nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total)
	})

	t.Run("list respects max page size", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects?pageSize=500", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 200, result.PageSize) // Capped at 200
	})
}

// TestProjectHandler_GetByID tests the GetByID endpoint
func TestProjectHandler_GetByID(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	project := createTestProject(t, db, customer, "Test Project", domain.ProjectStatusActive, userCtx.UserID.String())

	t.Run("get existing project", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.ProjectWithDetailsDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, project.ID, result.ID)
		assert.Equal(t, "Test Project", result.Name)
		assert.NotNil(t, result.BudgetSummary)
	})

	t.Run("get non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/projects/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("get with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/invalid-id", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-id")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestProjectHandler_Create tests the Create endpoint
func TestProjectHandler_Create(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "New Customer")

	t.Run("create valid project", func(t *testing.T) {
		reqBody := domain.CreateProjectRequest{
			Name:       "New Project",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusPlanning,
			StartDate:  time.Now(),
			Budget:     150000,
			ManagerID:  userCtx.UserID.String(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Location"))

		var result domain.ProjectDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", result.Name)
		assert.Equal(t, customer.ID, result.CustomerID)
		assert.NotEqual(t, uuid.Nil, result.ID)
	})

	t.Run("create with invalid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with missing required fields", func(t *testing.T) {
		reqBody := domain.CreateProjectRequest{
			Name: "Incomplete Project",
			// Missing required fields
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("create with non-existent customer", func(t *testing.T) {
		reqBody := domain.CreateProjectRequest{
			Name:       "Project Without Customer",
			CustomerID: uuid.New(), // Non-existent customer
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusPlanning,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  userCtx.UserID.String(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.Create(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestProjectHandler_Update tests the Update endpoint
func TestProjectHandler_Update(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Update Customer")
	project := createTestProject(t, db, customer, "Original Project", domain.ProjectStatusPlanning, userCtx.UserID.String())

	t.Run("update project successfully", func(t *testing.T) {
		reqBody := domain.UpdateProjectRequest{
			Name:      "Updated Project Name",
			CompanyID: domain.CompanyStalbygg,
			Status:    domain.ProjectStatusActive,
			StartDate: time.Now(),
			Budget:    200000,
			ManagerID: userCtx.UserID.String(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ID.String(), bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Update(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.ProjectDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Project Name", result.Name)
		assert.Equal(t, 200000.0, result.Budget)
	})

	t.Run("update non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateProjectRequest{
			Name:      "Updated Name",
			CompanyID: domain.CompanyStalbygg,
			Status:    domain.ProjectStatusActive,
			StartDate: time.Now(),
			Budget:    100000,
			ManagerID: userCtx.UserID.String(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+nonExistentID.String(), bytes.NewReader(body))
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
		reqBody := domain.UpdateProjectRequest{
			Name: "Test",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/invalid", bytes.NewReader(body))
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

// TestProjectHandler_Delete tests the Delete endpoint
func TestProjectHandler_Delete(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Delete Customer")

	t.Run("delete project successfully", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Project To Delete", domain.ProjectStatusPlanning, userCtx.UserID.String())

		req := httptest.NewRequest(http.MethodDelete, "/projects/"+project.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify project is deleted
		var found domain.Project
		err := db.Where("id = ?", project.ID).First(&found).Error
		assert.Error(t, err)
	})

	t.Run("delete non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/projects/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("delete with invalid ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/projects/invalid", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.Delete(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// TestProjectHandler_UpdateStatus tests the UpdateStatus endpoint
func TestProjectHandler_UpdateStatus(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Status Customer")

	t.Run("update status successfully", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Status Project", domain.ProjectStatusPlanning, userCtx.UserID.String())

		reqBody := domain.UpdateProjectStatusRequest{
			Status: domain.ProjectStatusActive,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.ProjectDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.ProjectStatusActive, result.Status)
	})

	t.Run("update status with health override", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Health Project", domain.ProjectStatusPlanning, userCtx.UserID.String())

		health := domain.ProjectHealthAtRisk
		reqBody := domain.UpdateProjectStatusRequest{
			Status: domain.ProjectStatusActive,
			Health: &health,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.ProjectDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, domain.ProjectStatusActive, result.Status)
		assert.Equal(t, domain.ProjectHealthAtRisk, *result.Health)
	})

	t.Run("invalid status transition", func(t *testing.T) {
		// Create a completed project - cannot transition from completed
		project := createTestProject(t, db, customer, "Completed Project", domain.ProjectStatusCompleted, userCtx.UserID.String())

		reqBody := domain.UpdateProjectStatusRequest{
			Status: domain.ProjectStatusActive, // Invalid transition from completed
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("update status with completion percent", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Percent Project", domain.ProjectStatusActive, userCtx.UserID.String())

		percent := 75.0
		reqBody := domain.UpdateProjectStatusRequest{
			Status:            domain.ProjectStatusActive,
			CompletionPercent: &percent,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+project.ID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.ProjectDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 75.0, *result.CompletionPercent)
	})

	t.Run("update status non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		reqBody := domain.UpdateProjectStatusRequest{
			Status: domain.ProjectStatusActive,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/projects/"+nonExistentID.String()+"/status", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.UpdateStatus(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestProjectHandler_GetBudget tests the GetBudget endpoint
func TestProjectHandler_GetBudget(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Budget Customer")
	project := createTestProject(t, db, customer, "Budget Project", domain.ProjectStatusActive, userCtx.UserID.String())
	// Update spent amount
	db.Model(project).Update("spent", 25000)

	t.Run("get budget successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID.String()+"/budget", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetBudget(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.ProjectBudgetDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 100000.0, result.Budget)
		assert.Equal(t, 25000.0, result.Spent)
		assert.Equal(t, 75000.0, result.Remaining)
		assert.Equal(t, 25.0, result.PercentUsed)
	})

	t.Run("get budget non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/projects/"+nonExistentID.String()+"/budget", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetBudget(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// TestProjectHandler_GetActivities tests the GetActivities endpoint
func TestProjectHandler_GetActivities(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Activity Customer")
	project := createTestProject(t, db, customer, "Activity Project", domain.ProjectStatusActive, userCtx.UserID.String())

	// Create some test activities
	for i := 0; i < 3; i++ {
		activity := &domain.Activity{
			TargetType: domain.ActivityTargetProject,
			TargetID:   project.ID,
			Title:      "Test Activity",
			Body:       "Activity body",
			OccurredAt: time.Now(),
		}
		db.Create(activity)
	}

	t.Run("get activities successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID.String()+"/activities", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetActivities(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []domain.ActivityDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("get activities with limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+project.ID.String()+"/activities?limit=2", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetActivities(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []domain.ActivityDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("get activities non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/projects/"+nonExistentID.String()+"/activities", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetActivities(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// Helper to create a test offer for project tests with optional budget dimensions
func createTestOfferForProject(t *testing.T, db *gorm.DB, customer *domain.Customer, title string, phase domain.OfferPhase, value float64, userID string) *domain.Offer {
	testProjectOfferCounter++

	offer := &domain.Offer{
		Title:               title,
		CustomerID:          customer.ID,
		CustomerName:        customer.Name,
		CompanyID:           domain.CompanyStalbygg,
		Phase:               phase,
		Probability:         80,
		Value:               value,
		Status:              domain.OfferStatusActive,
		ResponsibleUserID:   userID,
		ResponsibleUserName: "Test User",
	}

	// For non-draft offers, generate unique offer number
	if phase != domain.OfferPhaseDraft {
		offer.OfferNumber = fmt.Sprintf("PROJ-TEST-%d-%d", time.Now().UnixNano(), testProjectOfferCounter)
	}

	err := db.Create(offer).Error
	require.NoError(t, err)
	return offer
}

// Helper to create budget items for an offer
func createTestBudgetItems(t *testing.T, db *gorm.DB, offerID uuid.UUID, count int) []domain.BudgetItem {
	items := make([]domain.BudgetItem, count)
	for i := 0; i < count; i++ {
		item := domain.BudgetItem{
			ParentType:     domain.BudgetParentOffer,
			ParentID:       offerID,
			Name:           fmt.Sprintf("Test Item %d", i+1),
			ExpectedCost:   float64(10000 * (i + 1)),
			ExpectedMargin: 50, // 50% margin
			DisplayOrder:   i,
		}
		err := db.Create(&item).Error
		require.NoError(t, err)
		items[i] = item
	}
	return items
}

// TestProjectHandler_InheritBudget tests the InheritBudget endpoint
func TestProjectHandler_InheritBudget(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandlerWithDeps(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Inherit Budget Customer")

	t.Run("inherit budget from won offer successfully", func(t *testing.T) {
		// Create a won offer with budget items
		offer := createTestOfferForProject(t, db, customer, "Won Offer", domain.OfferPhaseWon, 150000, userCtx.UserID.String())
		createTestBudgetItems(t, db, offer.ID, 3)

		// Create a project to inherit into
		project := createTestProject(t, db, customer, "Project for Inheritance", domain.ProjectStatusPlanning, userCtx.UserID.String())

		reqBody := domain.InheritBudgetRequest{
			OfferID: offer.ID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.InheritBudgetResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result.Project)
		assert.Equal(t, project.ID, result.Project.ID)
		assert.Equal(t, 3, result.ItemsCount)
		assert.Equal(t, 150000.0, result.Project.Budget)
		assert.True(t, result.Project.HasDetailedBudget)
		assert.NotNil(t, result.Project.OfferID)
		assert.Equal(t, offer.ID, *result.Project.OfferID)
	})

	t.Run("inherit budget from won offer without budget items", func(t *testing.T) {
		// Create a won offer without budget items
		offer := createTestOfferForProject(t, db, customer, "Won Offer No Items", domain.OfferPhaseWon, 100000, userCtx.UserID.String())

		// Create a project to inherit into
		project := createTestProject(t, db, customer, "Project No Items", domain.ProjectStatusPlanning, userCtx.UserID.String())

		reqBody := domain.InheritBudgetRequest{
			OfferID: offer.ID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.InheritBudgetResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.NotNil(t, result.Project)
		assert.Equal(t, 0, result.ItemsCount)
		assert.Equal(t, 100000.0, result.Project.Budget)
		assert.NotNil(t, result.Project.OfferID)
	})

	t.Run("inherit budget fails for non-won offer", func(t *testing.T) {
		// Create an offer in draft phase (not won)
		offer := createTestOfferForProject(t, db, customer, "Draft Offer", domain.OfferPhaseDraft, 50000, userCtx.UserID.String())

		project := createTestProject(t, db, customer, "Project Draft Offer", domain.ProjectStatusPlanning, userCtx.UserID.String())

		reqBody := domain.InheritBudgetRequest{
			OfferID: offer.ID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("inherit budget fails for non-existent offer", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Project No Offer", domain.ProjectStatusPlanning, userCtx.UserID.String())
		nonExistentOfferID := uuid.New()

		reqBody := domain.InheritBudgetRequest{
			OfferID: nonExistentOfferID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("inherit budget fails for non-existent project", func(t *testing.T) {
		offer := createTestOfferForProject(t, db, customer, "Won Offer 2", domain.OfferPhaseWon, 75000, userCtx.UserID.String())
		nonExistentProjectID := uuid.New()

		reqBody := domain.InheritBudgetRequest{
			OfferID: offer.ID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+nonExistentProjectID.String()+"/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentProjectID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("inherit budget with invalid project ID", func(t *testing.T) {
		reqBody := domain.InheritBudgetRequest{
			OfferID: uuid.New(),
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/invalid-id/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-id")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("inherit budget with invalid JSON body", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Project Invalid JSON", domain.ProjectStatusPlanning, userCtx.UserID.String())

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/inherit-budget", bytes.NewReader([]byte("invalid json")))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("inherit budget with missing offer ID", func(t *testing.T) {
		project := createTestProject(t, db, customer, "Project Missing Offer", domain.ProjectStatusPlanning, userCtx.UserID.String())

		// Empty request body - missing required offerId
		reqBody := map[string]interface{}{}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/projects/"+project.ID.String()+"/inherit-budget", bytes.NewReader(body))
		req = req.WithContext(ctx)
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", project.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.InheritBudget(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
