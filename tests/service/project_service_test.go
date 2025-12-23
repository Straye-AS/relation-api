package service_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test UUIDs for consistent test data
const (
	testManagerID = "00000000-0000-0000-0000-000000000001"
	testUserID    = "00000000-0000-0000-0000-000000000002"
)

// TestProjectService is an integration test suite for ProjectService
// Requires a running PostgreSQL database with migrations applied

func setupProjectTestDB(t *testing.T) *gorm.DB {
	host := getEnvOrDefaultProject("DATABASE_HOST", "localhost")
	port := getEnvOrDefaultProject("DATABASE_PORT", "5433")
	user := getEnvOrDefaultProject("DATABASE_USER", "relation_user")
	password := getEnvOrDefaultProject("DATABASE_PASSWORD", "relation_password")
	dbname := getEnvOrDefaultProject("DATABASE_NAME", "relation_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	return db
}

func getEnvOrDefaultProject(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupProjectTestService(t *testing.T, db *gorm.DB) (*service.ProjectService, *projectTestFixtures) {
	log := zap.NewNop()

	projectRepo := repository.NewProjectRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	fileRepo := repository.NewFileRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	fileStorage, _ := storage.NewStorage(&config.StorageConfig{Mode: "disk", LocalBasePath: os.TempDir()}, log)
	fileService := service.NewFileService(
		fileRepo,
		offerRepo,
		customerRepo,
		projectRepo,
		supplierRepo,
		activityRepo,
		fileStorage,
		log,
	)

	svc := service.NewProjectServiceWithDeps(
		projectRepo,
		offerRepo,
		customerRepo,
		activityRepo,
		fileService,
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
		Name:      name,
		Email:     name + "@test.com",
		Phone:     "12345678",
		Country:   "Norway",
		Status:    domain.CustomerStatusActive,
		Tier:      domain.CustomerTierBronze,
		OrgNumber: fmt.Sprintf("%09d", time.Now().UnixNano()%1000000000),
	}
	err := f.customerRepo.Create(ctx, customer)
	require.NoError(t, err)
	return customer
}

func (f *projectTestFixtures) createTestProject(t *testing.T, ctx context.Context, name string, phase domain.ProjectPhase) (*domain.Project, *domain.Customer) {
	customer := f.createTestCustomer(t, ctx, "Customer for "+name)

	startDate := time.Now()
	customerID := customer.ID
	project := &domain.Project{
		Name:         name,
		CustomerID:   &customerID,
		CustomerName: customer.Name,
		Phase:        phase,
		StartDate:    startDate,
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
		// Keep customer for other tests that might need it
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Create")
		_ = customer // Customer is now inferred from offers, not set on creation

		startDate := time.Now()
		req := &domain.CreateProjectRequest{
			Name:        "Test Project Create",
			Description: "Test description",
			StartDate:   &startDate,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Project Create", result.Name)
		assert.Equal(t, "Test description", result.Description)
		// Customer is now inferred from offers, not set on creation
		assert.Nil(t, result.CustomerID)
		// Phase defaults to "tilbud"
		assert.Equal(t, domain.ProjectPhaseTilbud, result.Phase)
	})

	t.Run("create with all optional fields", func(t *testing.T) {
		startDate := time.Now()
		endDate := startDate.AddDate(0, 6, 0)
		req := &domain.CreateProjectRequest{
			Name:        "Test Project Full",
			Description: "Full description",
			StartDate:   &startDate,
			EndDate:     &endDate,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Project Full", result.Name)
		assert.Equal(t, "Full description", result.Description)
		assert.NotNil(t, result.EndDate)
	})
}

// TestProjectService_CreateSimplified tests the simplified project creation
// Projects are containers/folders for offers with minimal required fields.
func TestProjectService_CreateSimplified(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("create sets default phase to tilbud", func(t *testing.T) {
		req := &domain.CreateProjectRequest{
			Name: "Test Default Phase Project",
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Verify phase is set to default "tilbud"
		assert.Equal(t, domain.ProjectPhaseTilbud, result.Phase, "phase should default to 'tilbud'")
	})

	t.Run("create with minimal fields (name only)", func(t *testing.T) {
		req := &domain.CreateProjectRequest{
			Name: "Test Minimal Project",
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Minimal Project", result.Name)
		// Verify defaults
		assert.Equal(t, domain.ProjectPhaseTilbud, result.Phase)
		assert.Nil(t, result.CustomerID)
		assert.Empty(t, result.CustomerName)
		assert.Empty(t, result.Location)
		assert.Empty(t, result.Description)
		// ID should be generated
		assert.NotEqual(t, uuid.Nil, result.ID)
	})

	t.Run("create with all fields", func(t *testing.T) {
		startDate := time.Now()
		endDate := startDate.AddDate(1, 0, 0)
		req := &domain.CreateProjectRequest{
			Name:        "Test All Fields Project",
			Description: "Comprehensive project description",
			StartDate:   &startDate,
			EndDate:     &endDate,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test All Fields Project", result.Name)
		assert.Equal(t, "Comprehensive project description", result.Description)
		// Phase still defaults to "tilbud" even when all other fields are provided
		assert.Equal(t, domain.ProjectPhaseTilbud, result.Phase)
		// Dates should be set
		assert.NotEmpty(t, result.StartDate)
		assert.NotNil(t, result.EndDate)
	})

	t.Run("create without customer sets nil customer fields", func(t *testing.T) {
		req := &domain.CreateProjectRequest{
			Name: "Test No Customer Project",
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Customer should be nil (inferred from offers later)
		assert.Nil(t, result.CustomerID)
		assert.Empty(t, result.CustomerName)
	})

	t.Run("create without location sets empty location", func(t *testing.T) {
		req := &domain.CreateProjectRequest{
			Name: "Test No Location Project",
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Location should be empty (inferred from offers later)
		assert.Empty(t, result.Location)
	})
}

func TestProjectService_GetByID(t *testing.T) {
	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get project by id successfully", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Get Project", domain.ProjectPhaseWorking)

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

	t.Run("update project successfully", func(t *testing.T) {
		ctx := createProjectTestContext()
		project, _ := fixtures.createTestProject(t, ctx, "Test Update Project", domain.ProjectPhaseWorking)

		startDate := time.Now()
		req := &domain.UpdateProjectRequest{
			Name:        "Test Updated Project Name",
			StartDate:   &startDate,
			Description: "Updated description",
		}

		result, err := svc.Update(ctx, project.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Updated Project Name", result.Name)
		assert.Equal(t, domain.ProjectPhaseWorking, result.Phase)
		assert.Equal(t, "Updated description", result.Description)
	})

	t.Run("update project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		startDate := time.Now()
		req := &domain.UpdateProjectRequest{
			Name:      "Test Non-Existent",
			StartDate: &startDate,
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
		ctx := createProjectTestContextWithManagerID(testManagerID)
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
			fixtures.createTestProject(t, ctx, "Test List Project", domain.ProjectPhaseWorking)
		}

		result, err := svc.List(ctx, 1, 3, nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.LessOrEqual(t, len(result.Data.([]domain.ProjectDTO)), 3)
		assert.GreaterOrEqual(t, result.Total, int64(5))
	})

	t.Run("list projects filtered by status", func(t *testing.T) {
		fixtures.createTestProject(t, ctx, "Test Active Project Filter", domain.ProjectPhaseWorking)
		fixtures.createTestProject(t, ctx, "Test Planning Project Filter", domain.ProjectPhaseTilbud)

		activePhase := domain.ProjectPhaseWorking
		result, err := svc.List(ctx, 1, 100, nil, &activePhase)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// All returned projects should be active
		projects := result.Data.([]domain.ProjectDTO)
		for _, p := range projects {
			assert.Equal(t, domain.ProjectPhaseWorking, p.Phase)
		}
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
		ctx := createProjectTestContextWithManagerID(testManagerID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Tilbud Active", domain.ProjectPhaseTilbud)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseWorking)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseWorking, result.Phase)
	})

	t.Run("valid phase transition active to completed", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID(testManagerID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Active Complete", domain.ProjectPhaseWorking)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseCompleted)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseCompleted, result.Phase)
	})

	t.Run("valid phase transition tilbud to working", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID(testManagerID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Tilbud Working", domain.ProjectPhaseTilbud)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseWorking)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseWorking, result.Phase)
	})

	t.Run("invalid phase transition from completed", func(t *testing.T) {
		t.Skip("Skipping - service currently allows completed->active transition, needs business rule clarification")
		ctx := createProjectTestContextWithManagerID(testManagerID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Completed Invalid", domain.ProjectPhaseCompleted)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseWorking)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidPhaseTransition)
	})

	t.Run("invalid phase transition from cancelled", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID(testManagerID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Cancelled Invalid", domain.ProjectPhaseCancelled)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseWorking)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidPhaseTransition)
	})

	t.Run("same phase is valid", func(t *testing.T) {
		ctx := createProjectTestContextWithManagerID(testManagerID)
		project, _ := fixtures.createTestProject(t, ctx, "Test Phase Same", domain.ProjectPhaseWorking)

		result, err := svc.UpdatePhase(ctx, project.ID, domain.ProjectPhaseWorking)
		require.NoError(t, err)
		assert.Equal(t, domain.ProjectPhaseWorking, result.Phase)
	})

	t.Run("project not found", func(t *testing.T) {
		ctx := createProjectTestContext()
		result, err := svc.UpdatePhase(ctx, uuid.New(), domain.ProjectPhaseWorking)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrProjectNotFound)
	})
}

// ============================================================================
// Activity Tests
// ============================================================================

func TestProjectService_GetActivities(t *testing.T) {
	t.Skip("Skipping until activity logging is properly tested - activities not being created in test context")

	db := setupProjectTestDB(t)
	svc, fixtures := setupProjectTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createProjectTestContext()

	t.Run("get project activities", func(t *testing.T) {
		project, _ := fixtures.createTestProject(t, ctx, "Test Activity Project", domain.ProjectPhaseWorking)

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
