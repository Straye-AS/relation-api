package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func setupProjectHandlerTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
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

func createProjectTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func createTestProject(t *testing.T, db *gorm.DB, customer *domain.Customer, name string, phase domain.ProjectPhase, _ string) *domain.Project {
	startDate := time.Now()
	customerID := customer.ID
	project := &domain.Project{
		Name:         name,
		CustomerID:   &customerID,
		CustomerName: customer.Name,
		Phase:        phase,
		StartDate:    startDate,
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

	// Create test projects with various phases and health values
	projects := []struct {
		customer  *domain.Customer
		name      string
		phase     domain.ProjectPhase
		managerID string
	}{
		{customer1, "Alpha Project", domain.ProjectPhaseWorking, userCtx.UserID.String()},
		{customer1, "Beta Project", domain.ProjectPhaseTilbud, userCtx.UserID.String()},
		{customer1, "Gamma Project", domain.ProjectPhaseCompleted, "other-manager"},
		{customer2, "Delta Project", domain.ProjectPhaseWorking, userCtx.UserID.String()},
		{customer2, "Epsilon Project", domain.ProjectPhaseWorking, "other-manager"},
	}

	for _, p := range projects {
		createTestProject(t, db, p.customer, p.name, p.phase, p.managerID)
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

	t.Run("list with phase filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects?phase=working", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total) // Alpha, Delta, Epsilon are "working"
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

	// Note: managerId filter was removed as projects no longer have managers
	// Manager information is now tracked at the offer level

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
	project := createTestProject(t, db, customer, "Test Project", domain.ProjectPhaseWorking, userCtx.UserID.String())

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
		// Note: BudgetSummary may be nil for projects without linked offers with budget items
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

	customer := testutil.CreateTestCustomer(t, db, "New Customer")

	t.Run("create valid project", func(t *testing.T) {
		startDate := time.Now()
		customerID := customer.ID
		reqBody := domain.CreateProjectRequest{
			Name:       "New Project",
			CustomerID: &customerID,
			Phase:      domain.ProjectPhaseTilbud,
			StartDate:  &startDate,
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
		assert.NotNil(t, result.CustomerID)
		assert.Equal(t, customer.ID, *result.CustomerID)
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
		// Only Name is required for project creation now
		reqBody := domain.CreateProjectRequest{
			// Name is required but missing
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
		startDate := time.Now()
		nonExistentID := uuid.New()
		reqBody := domain.CreateProjectRequest{
			Name:       "Project Without Customer",
			CustomerID: &nonExistentID, // Non-existent customer
			Phase:      domain.ProjectPhaseTilbud,
			StartDate:  &startDate,
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
	project := createTestProject(t, db, customer, "Original Project", domain.ProjectPhaseWorking, userCtx.UserID.String())

	t.Run("update project successfully", func(t *testing.T) {
		startDate := time.Now()
		reqBody := domain.UpdateProjectRequest{
			Name:        "Updated Project Name",
			StartDate:   &startDate,
			Description: "Updated description",
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
		assert.Equal(t, "Updated description", result.Description)
	})

	t.Run("update non-existent project", func(t *testing.T) {
		nonExistentID := uuid.New()
		startDate := time.Now()
		reqBody := domain.UpdateProjectRequest{
			Name:      "Updated Name",
			StartDate: &startDate,
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
		project := createTestProject(t, db, customer, "Project To Delete", domain.ProjectPhaseTilbud, userCtx.UserID.String())

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

// TestProjectHandler_GetActivities tests the GetActivities endpoint
func TestProjectHandler_GetActivities(t *testing.T) {
	db := setupProjectHandlerTestDB(t)
	h := createProjectHandler(t, db)
	ctx := createProjectTestContext()
	userCtx, _ := auth.FromContext(ctx)

	customer := testutil.CreateTestCustomer(t, db, "Activity Customer")
	project := createTestProject(t, db, customer, "Activity Project", domain.ProjectPhaseWorking, userCtx.UserID.String())

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
