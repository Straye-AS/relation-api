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
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	numberSequenceRepo := repository.NewNumberSequenceRepository(db)
	userRepo := repository.NewUserRepository(db)

	companyService := service.NewCompanyService(log)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, log)

	svc := service.NewProjectServiceWithDeps(
		projectRepo,
		offerRepo,
		customerRepo,
		budgetItemRepo,
		activityRepo,
		userRepo,
		companyService,
		numberSequenceService,
		log,
		db,
	)

	fixtures := &projectTestFixtures{
		db:             db,
		customerRepo:   customerRepo,
		offerRepo:      offerRepo,
		projectRepo:    projectRepo,
		budgetItemRepo: budgetItemRepo,
	}

	return svc, fixtures
}

type projectTestFixtures struct {
	db             *gorm.DB
	customerRepo   *repository.CustomerRepository
	offerRepo      *repository.OfferRepository
	projectRepo    *repository.ProjectRepository
	budgetItemRepo *repository.BudgetItemRepository
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

func (f *projectTestFixtures) createTestBudgetItem(t *testing.T, ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, name string, expectedCost float64, marginPercent float64, order int) *domain.BudgetItem {
	item := &domain.BudgetItem{
		ParentType:     parentType,
		ParentID:       parentID,
		Name:           name,
		ExpectedCost:   expectedCost,
		ExpectedMargin: marginPercent,
		DisplayOrder:   order,
	}
	err := f.budgetItemRepo.Create(ctx, item)
	require.NoError(t, err)
	return item
}

func (f *projectTestFixtures) createTestProject(t *testing.T, ctx context.Context, name string, phase domain.ProjectPhase) (*domain.Project, *domain.Customer) {
	customer := f.createTestCustomer(t, ctx, "Customer for "+name)

	health := domain.ProjectHealthOnTrack
	managerID := "test-manager-id"
	startDate := time.Now()
	project := &domain.Project{
		Name:              name,
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             phase,
		StartDate:         startDate,
		Value:             100000,
		Cost:              80000,
		Spent:             0,
		ManagerID:         &managerID,
		HasDetailedBudget: false,
		Health:            &health,
	}
	err := f.projectRepo.Create(ctx, project)
	require.NoError(t, err)

	return project, customer
}

func (f *projectTestFixtures) cleanup(t *testing.T) {
	f.db.Exec("DELETE FROM activities WHERE target_type = 'Project' OR target_type = 'Customer'")
	f.db.Exec("DELETE FROM budget_items WHERE parent_type = 'project' OR parent_type = 'offer'")
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

		startDate := time.Now()
		managerID := "test-manager"
		req := &domain.CreateProjectRequest{
			Name:       "Test Project Create",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseTilbud,
			StartDate:  &startDate,
			Value:      150000,
			Cost:       120000,
			ManagerID:  &managerID,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Project Create", result.Name)
		assert.Equal(t, customer.ID, result.CustomerID)
		assert.Equal(t, customer.Name, result.CustomerName)
		assert.Equal(t, domain.ProjectPhaseTilbud, result.Phase)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOnTrack, *result.Health)
	})

	t.Run("create project with default health", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Health")

		startDate := time.Now()
		managerID := "test-manager"
		req := &domain.CreateProjectRequest{
			Name:       "Test Project Default Health",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseActive,
			StartDate:  &startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOnTrack, *result.Health)
	})

	t.Run("create fails with non-existent customer", func(t *testing.T) {
		startDate := time.Now()
		managerID := "test-manager"
		req := &domain.CreateProjectRequest{
			Name:       "Test Project Bad Customer",
			CustomerID: uuid.New(),
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseTilbud,
			StartDate:  &startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
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
		project, _ := fixtures.createTestProject(t, ctx, "Test Get Project", domain.ProjectPhaseActive)

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
		project, _ := fixtures.createTestProject(t, ctx, "Test Update Project", domain.ProjectPhaseTilbud)

		startDate := time.Now()
		managerID := "test-manager-id"
		req := &domain.UpdateProjectRequest{
			Name:      "Test Updated Project Name",
			CompanyID: domain.CompanyStalbygg,
			StartDate: &startDate,
			Value:     200000,
			Cost:      160000,
			Spent:     50000,
			ManagerID: &managerID,
		}

		result, err := svc.Update(ctx, project.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Updated Project Name", result.Name)
		assert.Equal(t, domain.ProjectPhaseActive, result.Phase)
		assert.Equal(t, float64(200000), result.Value)
	})

	t.Run("update project as admin", func(t *testing.T) {
		ctx := createProjectTestContext() // Manager role allows update
		project, _ := fixtures.createTestProject(t, ctx, "Test Update Admin", domain.ProjectPhaseTilbud)

		startDate := time.Now()
		managerID := "different-manager"
		req := &domain.UpdateProjectRequest{
			Name:      "Test Updated By Admin",
			CompanyID: domain.CompanyStalbygg,
			StartDate: &startDate,
			Value:     150000,
			Cost:      120000,
			ManagerID: &managerID,
		}

		result, err := svc.Update(ctx, project.ID, req)
		require.NoError(t, err)
		assert.Equal(t, "Test Updated By Admin", result.Name)
	})

	t.Run("update project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		startDate := time.Now()
		managerID := "test-manager"
		req := &domain.UpdateProjectRequest{
			Name:      "Test Non-Existent",
			CompanyID: domain.CompanyStalbygg,
			StartDate: &startDate,
			Value:     100000,
			Cost:      80000,
			ManagerID: &managerID,
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
		project, _ := fixtures.createTestProject(t, ctx, "Test Delete Project", domain.ProjectPhaseTilbud)

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
			fixtures.createTestProject(t, ctx, "Test List Project", domain.ProjectPhaseActive)
		}

		result, err := svc.List(ctx, 1, 3, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.LessOrEqual(t, len(result.Data.([]domain.ProjectDTO)), 3)
		assert.GreaterOrEqual(t, result.Total, int64(5))
	})

	t.Run("list projects filtered by status", func(t *testing.T) {
		fixtures.createTestProject(t, ctx, "Test Active Project Filter", domain.ProjectPhaseActive)
		fixtures.createTestProject(t, ctx, "Test Planning Project Filter", domain.ProjectPhaseTilbud)

		activePhase := domain.ProjectPhaseActive
		result, err := svc.List(ctx, 1, 100, nil, &activePhase)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// All returned projects should be active
		projects := result.Data.([]domain.ProjectDTO)
		for _, p := range projects {
			assert.Equal(t, domain.ProjectPhaseActive, p.Phase)
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

	t.Run("inherit budget items from won offer", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Inherit Customer")
		offer := fixtures.createTestOffer(t, ctx, "Test Inherit Offer", domain.OfferPhaseWon, customer.ID)

		// Create budget items on the offer
		fixtures.createTestBudgetItem(t, ctx, domain.BudgetParentOffer, offer.ID, "Steel Structure", 50000, 50, 0) // Revenue=75000
		fixtures.createTestBudgetItem(t, ctx, domain.BudgetParentOffer, offer.ID, "Assembly", 30000, 50, 1)        // Revenue=45000

		// Create project without inherited budget
		project, _ := fixtures.createTestProject(t, ctx, "Test Inherit Project", domain.ProjectPhaseTilbud)
		assert.False(t, project.HasDetailedBudget)

		// Inherit budget
		resp, err := svc.InheritBudgetFromOffer(ctx, project.ID, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 2, resp.ItemsCount)

		// Verify items were cloned
		projectItems, err := fixtures.budgetItemRepo.ListByParent(ctx, domain.BudgetParentProject, project.ID)
		require.NoError(t, err)
		assert.Len(t, projectItems, 2)

		// Verify item values
		assert.Equal(t, "Steel Structure", projectItems[0].Name)
		assert.Equal(t, float64(50000), projectItems[0].ExpectedCost)
		assert.Equal(t, float64(50), projectItems[0].ExpectedMargin)

		assert.Equal(t, "Assembly", projectItems[1].Name)
		assert.Equal(t, float64(30000), projectItems[1].ExpectedCost)
		assert.Equal(t, float64(50), projectItems[1].ExpectedMargin)

		// Verify project was updated
		updatedProject, err := svc.GetByID(ctx, project.ID)
		require.NoError(t, err)
		assert.True(t, updatedProject.HasDetailedBudget)
		assert.Equal(t, offer.Value, updatedProject.Value)
	})

	t.Run("cannot inherit from non-won offer", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Test Non-Won Customer")
		offer := fixtures.createTestOffer(t, ctx, "Test Non-Won Offer", domain.OfferPhaseDraft, customer.ID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Non-Won Project", domain.ProjectPhaseTilbud)

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
		project, _ := fixtures.createTestProject(t, ctx, "Test Offer NotFound Project", domain.ProjectPhaseTilbud)

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

	t.Run("get budget summary for project with budget items", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Budget Summary", domain.ProjectPhaseActive)

		// Add budget items
		fixtures.createTestBudgetItem(t, ctx, domain.BudgetParentProject, project.ID, "Item 1", 10000, 50, 0) // Revenue=15000
		fixtures.createTestBudgetItem(t, ctx, domain.BudgetParentProject, project.ID, "Item 2", 20000, 50, 1) // Revenue=30000

		summary, err := svc.GetBudgetSummary(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, domain.BudgetParentProject, summary.ParentType)
		assert.Equal(t, project.ID, summary.ParentID)
		assert.Equal(t, 2, summary.ItemCount)
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
// Phase Lifecycle Tests
// ============================================================================

func TestProjectService_UpdatePhase(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	t.Run("valid phase transition tilbud to active", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Tilbud Active", domain.ProjectPhaseTilbud)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseActive)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseActive, result.Phase)
	})

	t.Run("valid phase transition active to completed", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Active Complete", domain.ProjectPhaseActive)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseCompleted)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseCompleted, result.Phase)
	})

	t.Run("valid phase transition tilbud to working", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Tilbud Working", domain.ProjectPhaseTilbud)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseWorking)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseWorking, result.Phase)
	})

	t.Run("invalid phase transition from completed", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Completed Invalid", domain.ProjectPhaseCompleted)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseActive)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidPhaseTransition)
	})

	t.Run("invalid phase transition from cancelled", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Cancelled Invalid", domain.ProjectPhaseCancelled)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseActive)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidPhaseTransition)
	})

	t.Run("same phase is valid", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Same", domain.ProjectPhaseActive)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseActive)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseActive, result.Phase)
	})

	t.Run("project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		result, err := svc.UpdatePhase(ctx, uuid.New(), domain.ProjectPhaseActive)
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
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion Update", domain.ProjectPhaseActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, 50.0)
		require.NoError(t, err)
		assert.NotNil(t, result.CompletionPercent)
		assert.Equal(t, float64(50), *result.CompletionPercent)
	})

	t.Run("setting 100% auto-completes active project", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID("test-manager-id")
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion 100", domain.ProjectPhaseActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, 100.0)
		require.NoError(t, err)
		assert.Equal(t, float64(100), *result.CompletionPercent)
		assert.Equal(t, domain.ProjectPhaseCompleted, result.Phase)
	})

	t.Run("reject negative percent", func(t *testing.T) {
		ctx := createProjectTestContext()
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion Negative", domain.ProjectPhaseActive)

		result, err := svc.UpdateCompletionPercent(ctx, project.ID, -10.0)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidCompletionPercent)
	})

	t.Run("reject percent over 100", func(t *testing.T) {
		ctx := createProjectTestContext()
		project, _ := fixtures.createTestProject(t, ctx, "Test Completion Over100", domain.ProjectPhaseActive)

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
		project, _ := fixtures.createTestProject(t, ctx, "Test Health OnTrack", domain.ProjectPhaseActive)
		// Budget: 100000, Spent: 0 -> 0% variance -> on_track
		fixtures.db.Model(project).Updates(map[string]interface{}{"spent": 50000})

		result, err := svc.RecalculateHealth(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthOnTrack, *result.Health)
	})

	t.Run("health at_risk when approaching budget", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Health AtRisk", domain.ProjectPhaseActive)
		// Budget: 100000, Spent: 115000 -> 115% variance -> at_risk (110-120%)
		fixtures.db.Model(project).Updates(map[string]interface{}{"spent": 115000})

		result, err := svc.RecalculateHealth(ctx, project.ID)
		require.NoError(t, err)
		assert.NotNil(t, result.Health)
		assert.Equal(t, domain.ProjectHealthAtRisk, *result.Health)
	})

	t.Run("health over_budget when exceeding threshold", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Health OverBudget", domain.ProjectPhaseActive)
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
		startDate := time.Now()
		managerID := "specific-manager-id"
		project := &domain.Project{
			Name:       "Test Manager Project 1",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseActive,
			StartDate:  startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
		}
		fixtures.projectRepo.Create(ctx, project)

		project2 := &domain.Project{
			Name:       "Test Manager Project 2",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseActive,
			StartDate:  startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
		}
		fixtures.projectRepo.Create(ctx, project2)

		results, err := svc.GetByManager(ctx, "specific-manager-id")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		for _, p := range results {
			assert.NotNil(t, p.ManagerID)
			assert.Equal(t, "specific-manager-id", *p.ManagerID)
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
		startDate := time.Now()
		managerID := "test-manager"
		project := &domain.Project{
			Name:       "Test AtRisk Project",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseActive,
			StartDate:  startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
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
		startDate := time.Now()
		managerID := "test-manager"

		fixtures.db.Create(&domain.Project{
			Name:       "Test Summary OnTrack",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseActive,
			StartDate:  startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
			Health:     &onTrack,
		})

		fixtures.db.Create(&domain.Project{
			Name:       "Test Summary AtRisk",
			CustomerID: customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.ProjectPhaseActive,
			StartDate:  startDate,
			Value:      100000,
			Cost:       80000,
			ManagerID:  &managerID,
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
		project, _ := fixtures.createTestProject(t, ctx, "Test Activity Project", domain.ProjectPhaseActive)

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
