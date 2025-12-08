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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestProjectService is an integration test suite for ProjectService
// Requires a running PostgreSQL database with migrations applied

func setupProjectTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=relation_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	return db
}

func setupProjectTestService(t *testing.T, db *gorm.DB) (*service.ProjectService, *projectTestFixtures) {
	log := zap.NewNop()

	projectRepo := repository.NewProjectRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	dimensionRepo := repository.NewBudgetDimensionRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	svc := service.NewProjectServiceWithDeps(
		projectRepo,
		offerRepo,
		customerRepo,
		dimensionRepo,
		activityRepo,
		log,
		db,
	)

	fixtures := &projectTestFixtures{
		db:            db,
		customerRepo:  customerRepo,
		offerRepo:     offerRepo,
		projectRepo:   projectRepo,
		dimensionRepo: dimensionRepo,
	}

	return svc, fixtures
}

type projectTestFixtures struct {
	db            *gorm.DB
	customerRepo  *repository.CustomerRepository
	offerRepo     *repository.OfferRepository
	projectRepo   *repository.ProjectRepository
	dimensionRepo *repository.BudgetDimensionRepository
}

func (f *projectTestFixtures) createTestCustomer(t *testing.T, ctx context.Context, name string) *domain.Customer {
	customer := &domain.Customer{
		Name:    name,
		Email:   name + "@test.com",
		Phone:   "12345678",
		Country: "Norway",
		Status:  domain.CustomerStatusActive,
		Tier:    domain.CustomerTierBronze,
	}
	err := f.customerRepo.Create(ctx, customer)
	require.NoError(t, err)
	return customer
}

func (f *projectTestFixtures) createTestOffer(t *testing.T, ctx context.Context, title string, phase domain.OfferPhase, customerID uuid.UUID) *domain.Offer {
	offer := &domain.Offer{
		Title:             title,
		CustomerID:        customerID,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             phase,
		Probability:       50,
		Value:             100000,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: "test-user-id",
		Description:       "Test offer description",
	}
	err := f.offerRepo.Create(ctx, offer)
	require.NoError(t, err)
	return offer
}

func (f *projectTestFixtures) createTestBudgetDimension(t *testing.T, ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, name string, cost, revenue float64, order int) *domain.BudgetDimension {
	dimension := &domain.BudgetDimension{
		ParentType:   parentType,
		ParentID:     parentID,
		CustomName:   name,
		Cost:         cost,
		Revenue:      revenue,
		DisplayOrder: order,
	}
	err := f.dimensionRepo.Create(ctx, dimension)
	require.NoError(t, err)
	return dimension
}

func (f *projectTestFixtures) createTestProject(t *testing.T, ctx context.Context, name string, status domain.ProjectStatus) (*domain.Project, *domain.Customer) {
	customer := f.createTestCustomer(t, ctx, "Customer for "+name)

	health := domain.ProjectHealthOnTrack
	project := &domain.Project{
		Name:              name,
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Status:            status,
		StartDate:         time.Now(),
		Budget:            100000,
		Spent:             0,
		ManagerID:         "test-manager-id",
		HasDetailedBudget: false,
		Health:            &health,
	}
	err := f.projectRepo.Create(ctx, project)
	require.NoError(t, err)

	return project, customer
}

func (f *projectTestFixtures) cleanup(t *testing.T) {
	f.db.Exec("DELETE FROM activities WHERE target_type = 'Project' OR target_type = 'Customer'")
	f.db.Exec("DELETE FROM budget_dimensions WHERE parent_type = 'project' OR parent_type = 'offer'")
	f.db.Exec("DELETE FROM projects WHERE name LIKE 'Test%'")
	f.db.Exec("DELETE FROM offers WHERE title LIKE 'Test%'")
	f.db.Exec("DELETE FROM customers WHERE name LIKE 'Customer for%' OR name LIKE 'Test%'")
}

func createProjectTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		DisplayName: "Test User",
		Email:       "test@straye.no",
		Roles:       []domain.UserRoleType{domain.RoleManager},
		CompanyID:   domain.CompanyStalbygg,
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func createProjectTestContextWithManagerID(managerID string) context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.MustParse(managerID),
		DisplayName: "Test Manager",
		Email:       "manager@straye.no",
		Roles:       []domain.UserRoleType{domain.RoleProjectManager},
		CompanyID:   domain.CompanyStalbygg,
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// ============================================================================
// CRUD Tests
// ============================================================================

func TestProjectService_Create(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("create project successfully", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Create")

		req := &domain.CreateProjectRequest{
			Name:       "Test Project Create",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusPlanning,
			StartDate:  time.Now(),
			Budget:     150000,
			ManagerID:  "test-manager",
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Project Create", result.Name)
		assert.Equal(t, customer.ID, result.CustomerID)
		assert.Equal(t, customer.Name, result.CustomerName)
		assert.Equal(t, domain.ProjectStatusPlanning, result.Status)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOnTrack, *result.Health)
	})

	t.Run("create project with default health", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Health")

		req := &domain.CreateProjectRequest{
			Name:       "Test Project Default Health",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "test-manager",
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOnTrack, *result.Health)
	})

	t.Run("create fails with non-existent customer", func(t *testing.T) {
		req := &domain.CreateProjectRequest{
			Name:       "Test Project Bad Customer",
			CustomerID: uuid.New(),
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusPlanning,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "test-manager",
		}

		result, err := svc.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrCustomerNotFound)
	})
}

func TestProjectService_GetByID(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get project by id successfully", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Get Project", domain.ProjectStatusActive)

		result, err := svc.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, project.ID, result.ID)
		assert.Equal(t, project.Name, result.Name)
	})

	t.Run("get non-existent project", func(t *testing.T) {
		result, err := svc.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

func TestProjectService_Update(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	t.Run("update project as manager", func(t *testing.T) {
		// Create context with matching manager ID
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Update Project", domain.ProjectStatusPlanning)

		req := &domain.UpdateProjectRequest{
			Name:      "Test Updated Project Name",
			CompanyID: domain.CompanyStalbygg,
			Status:    domain.ProjectStatusActive,
			StartDate: time.Now(),
			Budget:    200000,
			Spent:     50000,
			ManagerID: "test-manager-id",
		}

		result, err := svc.Update(ctx, project.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Updated Project Name", result.Name)
		assert.Equal(t, domain.ProjectStatusActive, result.Status)
		assert.Equal(t, float64(200000), result.Budget)
	})

	t.Run("update project as admin", func(t *testing.T) {
		ctx := createProjectTestContext() // Manager role allows update
		project, _ := fixtures.createTestProject(t, ctx, "Test Update Admin", domain.ProjectStatusPlanning)

		req := &domain.UpdateProjectRequest{
			Name:      "Test Updated By Admin",
			CompanyID: domain.CompanyStalbygg,
			Status:    domain.ProjectStatusActive,
			StartDate: time.Now(),
			Budget:    150000,
			ManagerID: "different-manager",
		}

		result, err := svc.Update(ctx, project.ID, req)
		require.NoError(t, err)
		assert.Equal(t, "Test Updated By Admin", result.Name)
	})

	t.Run("update project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		req := &domain.UpdateProjectRequest{
			Name:      "Test Non-Existent",
			CompanyID: domain.CompanyStalbygg,
			Status:    domain.ProjectStatusActive,
			StartDate: time.Now(),
			Budget:    100000,
			ManagerID: "test-manager",
		}

		result, err := svc.Update(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

func TestProjectService_Delete(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	t.Run("delete project as manager", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Delete Project", domain.ProjectStatusPlanning)

		err := svc.Delete(ctx, project.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = svc.GetByID(ctx, project.ID)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})

	t.Run("delete non-existent project", func(t *testing.T) {
		ctx := createProjectTestContext()
		err := svc.Delete(ctx, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

func TestProjectService_List(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("list projects with pagination", func(t *testing.T) {
		// Create multiple projects
		for i := 0; i < 5; i++ {
			fixtures.createTestProject(t, ctx, "Test List Project", domain.ProjectStatusActive)
		}

		result, err := svc.List(ctx, 1, 3, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.LessOrEqual(t, len(result.Data.([]domain.ProjectDTO)), 3)
		assert.GreaterOrEqual(t, result.Total, int64(5))
	})

	t.Run("list projects filtered by status", func(t *testing.T) {
		fixtures.createTestProject(t, ctx, "Test Active Project Filter", domain.ProjectStatusActive)
		fixtures.createTestProject(t, ctx, "Test Planning Project Filter", domain.ProjectStatusPlanning)

		activeStatus := domain.ProjectStatusActive
		result, err := svc.List(ctx, 1, 100, nil, &activeStatus)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// All returned projects should be active
		projects := result.Data.([]domain.ProjectDTO)
		for _, p := range projects {
			assert.Equal(t, domain.ProjectStatusActive, p.Status)
		}
	})
}

// ============================================================================
// Budget Inheritance Tests
// ============================================================================

func TestProjectService_InheritBudgetFromOffer(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("inherit budget dimensions from won offer", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Inherit Customer")
		offer := fixtures.createTestOffer(t, ctx, "Test Inherit Offer", domain.OfferPhaseWon, customer.ID)

		// Create budget dimensions on the offer
		fixtures.createTestBudgetDimension(t, ctx, domain.BudgetParentOffer, offer.ID, "Steel Structure", 50000, 75000, 0)
		fixtures.createTestBudgetDimension(t, ctx, domain.BudgetParentOffer, offer.ID, "Assembly", 30000, 45000, 1)

		// Create project without inherited budget
		project, _ := fixtures.createTestProject(t, ctx, "Test Inherit Project", domain.ProjectStatusPlanning)
		assert.False(t, project.HasDetailedBudget)

		// Inherit budget
		resp, err := svc.InheritBudgetFromOffer(ctx, project.ID, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, resp.DimensionsCount)

		// Verify dimensions were cloned
		projectDims, err := fixtures.dimensionRepo.GetByParent(ctx, domain.BudgetParentProject, project.ID)
		require.NoError(t, err)
		assert.Len(t, projectDims, 2)

		// Verify dimension values
		assert.Equal(t, "Steel Structure", projectDims[0].CustomName)
		assert.Equal(t, float64(50000), projectDims[0].Cost)
		assert.Equal(t, float64(75000), projectDims[0].Revenue)

		assert.Equal(t, "Assembly", projectDims[1].CustomName)
		assert.Equal(t, float64(30000), projectDims[1].Cost)
		assert.Equal(t, float64(45000), projectDims[1].Revenue)

		// Verify project was updated
		updatedProject, err := svc.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.True(t, updatedProject.HasDetailedBudget)
		assert.Equal(t, offer.Value, updatedProject.Budget)
	})

	t.Run("cannot inherit from non-won offer", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Non-Won Customer")
		offer := fixtures.createTestOffer(t, ctx, "Test Non-Won Offer", domain.OfferPhaseDraft, customer.ID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Non-Won Project", domain.ProjectStatusPlanning)

		_, err := svc.InheritBudgetFromOffer(ctx, project.ID, offer.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrOfferNotWon)
	})

	t.Run("project not found", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test NotFound Customer")
		offer := fixtures.createTestOffer(t, ctx, "Test NotFound Offer", domain.OfferPhaseWon, customer.ID)

		_, err := svc.InheritBudgetFromOffer(ctx, uuid.New(), offer.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})

	t.Run("offer not found", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Offer NotFound Project", domain.ProjectStatusPlanning)

		_, err := svc.InheritBudgetFromOffer(ctx, project.ID, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestProjectService_GetBudgetSummary(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get budget summary for project with dimensions", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Budget Summary", domain.ProjectStatusActive)

		// Add budget dimensions
		fixtures.createTestBudgetDimension(t, ctx, domain.BudgetParentProject, project.ID, "Dim 1", 10000, 15000, 0)
		fixtures.createTestBudgetDimension(t, ctx, domain.BudgetParentProject, project.ID, "Dim 2", 20000, 30000, 1)

		summary, err := svc.GetBudgetSummary(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, domain.BudgetParentProject, summary.ParentType)
		assert.Equal(t, project.ID, summary.ParentID)
		assert.Equal(t, 2, summary.DimensionCount)
		assert.Equal(t, float64(30000), summary.TotalCost)
		assert.Equal(t, float64(45000), summary.TotalRevenue)
		assert.Equal(t, float64(15000), summary.TotalProfit)
	})

	t.Run("budget summary for project not found", func(t *testing.T) {
		summary, err := svc.GetBudgetSummary(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

// ============================================================================
// Status Lifecycle Tests
// ============================================================================

func TestProjectService_UpdateStatus(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	t.Run("valid status transition planning to active", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Status Planning Active", domain.ProjectStatusPlanning)

		result, err := svc.UpdateStatus(ctx, project.ID, domain.ProjectStatusActive)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectStatusActive, result.Status)
	})

	t.Run("valid status transition active to completed", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Status Active Complete", domain.ProjectStatusActive)

		result, err := svc.UpdateStatus(ctx, project.ID, domain.ProjectStatusCompleted)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectStatusCompleted, result.Status)
		// Completion percent should be auto-set to 100
		assert.NotNil(t, result.CompletionPercent)
		assert.Equal(t, float64(100), *result.CompletionPercent)
	})

	t.Run("valid status transition active to on_hold", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Status Active Hold", domain.ProjectStatusActive)

		result, err := svc.UpdateStatus(ctx, project.ID, domain.ProjectStatusOnHold)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectStatusOnHold, result.Status)
	})

	t.Run("invalid status transition from completed", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Status Completed Invalid", domain.ProjectStatusCompleted)

		result, err := svc.UpdateStatus(ctx, project.ID, domain.ProjectStatusActive)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidStatusTransition)
	})

	t.Run("invalid status transition from cancelled", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Status Cancelled Invalid", domain.ProjectStatusCancelled)

		result, err := svc.UpdateStatus(ctx, project.ID, domain.ProjectStatusActive)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidStatusTransition)
	})

	t.Run("same status is valid", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Status Same", domain.ProjectStatusActive)

		result, err := svc.UpdateStatus(ctx, project.ID, domain.ProjectStatusActive)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectStatusActive, result.Status)
	})

	t.Run("project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		result, err := svc.UpdateStatus(ctx, uuid.New(), domain.ProjectStatusActive)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

// ============================================================================
// Completion Percent Tests
// ============================================================================

func TestProjectService_UpdateCompletionPercent(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	t.Run("update completion percent successfully", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion Update", domain.ProjectStatusActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, 50.0)
		require.NoError(t, err)
		assert.NotNil(t, result.CompletionPercent)
		assert.Equal(t, float64(50), *result.CompletionPercent)
	})

	t.Run("setting 100% auto-completes active project", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion 100", domain.ProjectStatusActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, 100.0)
		require.NoError(t, err)
		assert.Equal(t, float64(100), *result.CompletionPercent)
		assert.Equal(t, domain.ProjectStatusCompleted, result.Status)
	})

	t.Run("reject negative percent", func(t *testing.T) {
		ctx := createProjectTestContext()
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion Negative", domain.ProjectStatusActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, -10.0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidCompletionPercent)
	})

	t.Run("reject percent over 100", func(t *testing.T) {
		ctx := createProjectTestContext()
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion Over100", domain.ProjectStatusActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, 150.0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidCompletionPercent)
	})

	t.Run("project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		result, err := svc.UpdateCompletionPercent(ctx, uuid.New(), 50.0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

// ============================================================================
// Health Calculation Tests
// ============================================================================

func TestProjectService_RecalculateHealth(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("health on_track when under budget", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Health OnTrack", domain.ProjectStatusActive)
		// Budget: 100000, Spent: 0 -> 0% variance -> on_track
		fixtures.db.Model(project).Updates(map[string]interface{}{"spent": 50000})

		result, err := svc.RecalculateHealth(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOnTrack, *result.Health)
	})

	t.Run("health at_risk when approaching budget", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Health AtRisk", domain.ProjectStatusActive)
		// Budget: 100000, Spent: 115000 -> 115% variance -> at_risk (110-120%)
		fixtures.db.Model(project).Updates(map[string]interface{}{"spent": 115000})

		result, err := svc.RecalculateHealth(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthAtRisk, *result.Health)
	})

	t.Run("health over_budget when exceeding threshold", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Health OverBudget", domain.ProjectStatusActive)
		// Budget: 100000, Spent: 125000 -> 125% variance -> over_budget (>120%)
		fixtures.db.Model(project).Updates(map[string]interface{}{"spent": 125000})

		result, err := svc.RecalculateHealth(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOverBudget, *result.Health)
	})

	t.Run("project not found", func(t *testing.T) {
		result, err := svc.RecalculateHealth(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

// ============================================================================
// Query Method Tests
// ============================================================================

func TestProjectService_GetByManager(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get projects by manager", func(t *testing.T) {
		// Create projects with specific manager
		customer := fixtures.createTestCustomer(t, ctx, "Test Manager Customer")
		project := &domain.Project{
			Name:       "Test Manager Project 1",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "specific-manager-id",
		}
		fixtures.projectRepo.Create(ctx, project)

		project2 := &domain.Project{
			Name:       "Test Manager Project 2",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "specific-manager-id",
		}
		fixtures.projectRepo.Create(ctx, project2)

		results, err := svc.GetByManager(ctx, "specific-manager-id")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		for _, p := range results {
			assert.Equal(t, "specific-manager-id", p.ManagerID)
		}
	})
}

func TestProjectService_GetByHealth(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get projects by health status", func(t *testing.T) {
		// Create projects with different health statuses
		atRiskHealth := domain.ProjectHealthAtRisk
		customer := fixtures.createTestCustomer(t, ctx, "Test Health Customer")
		project := &domain.Project{
			Name:       "Test AtRisk Project",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "test-manager",
			Health:     &atRiskHealth,
		}
		fixtures.projectRepo.Create(ctx, project)

		results, err := svc.GetByHealth(ctx, domain.ProjectHealthAtRisk)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		for _, p := range results {
			assert.NotNil(t, p.Health)
			assert.Equal(t, domain.ProjectHealthAtRisk, *p.Health)
		}
	})
}

func TestProjectService_GetHealthSummary(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get health summary", func(t *testing.T) {
		// Create projects with different health statuses
		customer := fixtures.createTestCustomer(t, ctx, "Test Summary Customer")

		onTrack := domain.ProjectHealthOnTrack
		atRisk := domain.ProjectHealthAtRisk

		fixtures.db.Create(&domain.Project{
			Name:       "Test Summary OnTrack",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "test-manager",
			Health:     &onTrack,
		})

		fixtures.db.Create(&domain.Project{
			Name:       "Test Summary AtRisk",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			Budget:     100000,
			ManagerID:  "test-manager",
			Health:     &atRisk,
		})

		summary, err := svc.GetHealthSummary(ctx)
		require.NoError(t, err)
		assert.NotNil(t, summary)
		// Just verify we get a map back - counts depend on test data
		assert.NotNil(t, summary)
	})
}

// ============================================================================
// Activity Tests
// ============================================================================

func TestProjectService_GetActivities(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get project activities", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Activity Project", domain.ProjectStatusActive)

		// The creation should have logged an activity
		activities, err := svc.GetActivities(ctx, project.ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(activities), 1)
	})

	t.Run("project not found", func(t *testing.T) {
		activities, err := svc.GetActivities(ctx, uuid.New(), 10)
		assert.Error(t, err)
		assert.Nil(t, activities)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}
