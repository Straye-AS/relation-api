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

// Helper functions for pointer values
func boolPtr(b bool) *bool { return &b }

// TestOfferService is an integration test suite for OfferService
// Requires a running PostgreSQL database with migrations applied

func setupOfferTestDB(t *testing.T) *gorm.DB {
	host := getEnvOrDefaultOffer("DATABASE_HOST", "localhost")
	port := getEnvOrDefaultOffer("DATABASE_PORT", "5433")
	user := getEnvOrDefaultOffer("DATABASE_USER", "relation_user")
	password := getEnvOrDefaultOffer("DATABASE_PASSWORD", "relation_password")
	dbname := getEnvOrDefaultOffer("DATABASE_NAME", "relation_test")

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

func getEnvOrDefaultOffer(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupOfferTestService(t *testing.T, db *gorm.DB) (*service.OfferService, *offerTestFixtures) {
	log := zap.NewNop()

	offerRepo := repository.NewOfferRepository(db)
	offerItemRepo := repository.NewOfferItemRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	budgetItemRepo := repository.NewBudgetItemRepository(db)
	fileRepo := repository.NewFileRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	userRepo := repository.NewUserRepository(db)
	numberSequenceRepo := repository.NewNumberSequenceRepository(db)

	companyService := service.NewCompanyServiceWithRepo(companyRepo, userRepo, log)
	numberSequenceService := service.NewNumberSequenceService(numberSequenceRepo, log)

	fileStorage, _ := storage.NewStorage(&config.StorageConfig{Mode: "disk", LocalBasePath: os.TempDir()}, log)
	supplierRepo := repository.NewSupplierRepository(db)
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

	svc := service.NewOfferService(
		offerRepo,
		offerItemRepo,
		customerRepo,
		projectRepo,
		budgetItemRepo,
		fileRepo,
		activityRepo,
		userRepo,
		companyService,
		numberSequenceService,
		fileService,
		log,
		db,
	)

	fixtures := &offerTestFixtures{
		db:             db,
		customerRepo:   customerRepo,
		offerRepo:      offerRepo,
		budgetItemRepo: budgetItemRepo,
	}

	return svc, fixtures
}

type offerTestFixtures struct {
	db             *gorm.DB
	customerRepo   *repository.CustomerRepository
	offerRepo      *repository.OfferRepository
	budgetItemRepo *repository.BudgetItemRepository
}

func (f *offerTestFixtures) createTestCustomer(t *testing.T, ctx context.Context, name string) *domain.Customer {
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

func (f *offerTestFixtures) createTestOffer(t *testing.T, ctx context.Context, title string, phase domain.OfferPhase) *domain.Offer {
	customer := f.createTestCustomer(t, ctx, "Customer for "+title)

	offer := &domain.Offer{
		Title:             title,
		CustomerID:        &customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             phase,
		Probability:       50,
		Value:             10000,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: testUserID,
		Description:       "Test offer description",
	}

	// Non-draft offers should have an offer number
	// Draft offers should NOT have an offer number
	if phase != domain.OfferPhaseDraft {
		offer.OfferNumber = fmt.Sprintf("TEST-%s-%d", domain.GetCompanyPrefix(domain.CompanyStalbygg), time.Now().UnixNano())
	}

	err := f.offerRepo.Create(ctx, offer)
	require.NoError(t, err)
	return offer
}

func (f *offerTestFixtures) createTestBudgetItem(t *testing.T, ctx context.Context, offerID uuid.UUID, name string, expectedCost float64, marginPercent float64, order int) *domain.BudgetItem {
	item := &domain.BudgetItem{
		ParentType:     domain.BudgetParentOffer,
		ParentID:       offerID,
		Name:           name,
		ExpectedCost:   expectedCost,
		ExpectedMargin: marginPercent,
		DisplayOrder:   order,
	}
	err := f.budgetItemRepo.Create(ctx, item)
	require.NoError(t, err)
	return item
}

func (f *offerTestFixtures) cleanup(t *testing.T) {
	f.db.Exec("DELETE FROM activities WHERE target_type = 'Offer' OR target_type = 'Project' OR target_type = 'Customer'")
	f.db.Exec("DELETE FROM budget_items WHERE parent_type = 'offer' OR parent_type = 'project'")
	f.db.Exec("DELETE FROM files WHERE offer_id IS NOT NULL")
	f.db.Exec("DELETE FROM offer_items")
	f.db.Exec("DELETE FROM projects WHERE name LIKE 'Test%' OR name LIKE 'Copy of%'")
	f.db.Exec("DELETE FROM offers WHERE title LIKE 'Test%' OR title LIKE 'Copy of%'")
	f.db.Exec("DELETE FROM customers WHERE name LIKE 'Customer for%' OR name LIKE 'Test%'")
}

func createOfferTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@straye.no",
		Roles:       []domain.UserRoleType{domain.RoleManager},
		CompanyID:   domain.CompanyStalbygg,
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// ============================================================================
// Lifecycle Method Tests
// ============================================================================

func TestOfferService_SendOffer(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("send offer from draft phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Send Draft Offer", domain.OfferPhaseDraft)

		result, err := svc.SendOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
	})

	t.Run("send offer from in_progress phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Send InProgress Offer", domain.OfferPhaseInProgress)

		result, err := svc.SendOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
	})

	t.Run("cannot send already sent offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Already Sent Offer", domain.OfferPhaseSent)

		result, err := svc.SendOffer(ctx, offer.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInDraftPhase)
	})

	t.Run("cannot send won offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Won Offer Send", domain.OfferPhaseOrder)

		result, err := svc.SendOffer(ctx, offer.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInDraftPhase)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.SendOffer(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_AcceptOffer(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("accept offer without project creation", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept No Project", domain.OfferPhaseSent)

		req := &domain.AcceptOfferRequest{
			CreateProject: false,
		}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Offer)
		assert.Nil(t, result.Project)
		assert.Equal(t, domain.OfferPhaseOrder, result.Offer.Phase)
	})

	t.Run("accept offer with project creation", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept With Project", domain.OfferPhaseSent)
		offer.Value = 50000
		fixtures.db.Save(offer)

		req := &domain.AcceptOfferRequest{
			CreateProject: true,
			ProjectName:   "New Project from Offer",
		}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Offer)
		assert.NotNil(t, result.Project)
		assert.Equal(t, domain.OfferPhaseOrder, result.Offer.Phase)
		assert.Equal(t, "New Project from Offer", result.Project.Name)
		// Project no longer has Value field - economic tracking is on Offer
	})

	t.Run("accept offer with project creation clones budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Clone Items", domain.OfferPhaseSent)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 1", 1000, 50, 0)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 2", 2000, 50, 1)

		req := &domain.AcceptOfferRequest{
			CreateProject: true,
		}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result.Project)

		// Verify items were cloned to project
		projectItems, err := fixtures.budgetItemRepo.ListByParent(ctx, domain.BudgetParentProject, result.Project.ID)
		require.NoError(t, err)
		assert.Len(t, projectItems, 2)
	})

	t.Run("cannot accept draft offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Draft", domain.OfferPhaseDraft)

		req := &domain.AcceptOfferRequest{CreateProject: false}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInSentPhase)
	})

	t.Run("cannot accept already won offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Won", domain.OfferPhaseOrder)

		req := &domain.AcceptOfferRequest{CreateProject: false}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInSentPhase)
	})

	t.Run("not found", func(t *testing.T) {
		req := &domain.AcceptOfferRequest{CreateProject: false}

		result, err := svc.AcceptOffer(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_RejectOffer(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("reject offer with reason", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Reject With Reason", domain.OfferPhaseSent)

		req := &domain.RejectOfferRequest{
			Reason: "Customer chose competitor",
		}

		result, err := svc.RejectOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseLost, result.Phase)
		assert.Contains(t, result.Notes, "Lost reason: Customer chose competitor")
	})

	t.Run("reject offer without reason", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Reject No Reason", domain.OfferPhaseSent)

		req := &domain.RejectOfferRequest{}

		result, err := svc.RejectOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseLost, result.Phase)
	})

	t.Run("cannot reject draft offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Reject Draft", domain.OfferPhaseDraft)

		req := &domain.RejectOfferRequest{}

		result, err := svc.RejectOffer(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotSent)
	})

	t.Run("not found", func(t *testing.T) {
		req := &domain.RejectOfferRequest{}

		result, err := svc.RejectOffer(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_ExpireOffer(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("expire draft offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Expire Draft", domain.OfferPhaseDraft)

		result, err := svc.ExpireOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseExpired, result.Phase)
	})

	t.Run("expire sent offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Expire Sent", domain.OfferPhaseSent)

		result, err := svc.ExpireOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseExpired, result.Phase)
	})

	t.Run("cannot expire won offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Expire Won", domain.OfferPhaseOrder)

		result, err := svc.ExpireOffer(ctx, offer.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferAlreadyClosed)
	})

	t.Run("cannot expire lost offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Expire Lost", domain.OfferPhaseLost)

		result, err := svc.ExpireOffer(ctx, offer.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferAlreadyClosed)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.ExpireOffer(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_CloneOffer(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("clone offer with default title", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone Default", domain.OfferPhaseDraft)

		req := &domain.CloneOfferRequest{
			IncludeBudget: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Copy of Test Clone Default", result.Title)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase)
		assert.NotEqual(t, offer.ID, result.ID)
	})

	t.Run("clone offer with custom title", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone Custom", domain.OfferPhaseSent)

		req := &domain.CloneOfferRequest{
			NewTitle:      "My Custom Clone Title",
			IncludeBudget: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "My Custom Clone Title", result.Title)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase) // Cloned offers start as draft
	})

	t.Run("clone offer with budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone With Items", domain.OfferPhaseDraft)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Clone Item 1", 1000, 50, 0)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Clone Item 2", 2000, 50, 1)

		req := &domain.CloneOfferRequest{
			IncludeBudget: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify items were cloned
		clonedItems, err := fixtures.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, result.ID)
		require.NoError(t, err)
		assert.Len(t, clonedItems, 2)
		assert.Equal(t, "Clone Item 1", clonedItems[0].Name)
		assert.Equal(t, "Clone Item 2", clonedItems[1].Name)
	})

	t.Run("clone offer without budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone No Items", domain.OfferPhaseDraft)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "No Clone Item", 1000, 50, 0)

		req := &domain.CloneOfferRequest{
			IncludeBudget: boolPtr(false), // Explicitly don't clone budget items
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify items were NOT cloned
		clonedItems, err := fixtures.budgetItemRepo.ListByParent(ctx, domain.BudgetParentOffer, result.ID)
		require.NoError(t, err)
		assert.Len(t, clonedItems, 0)
	})

	t.Run("clone won offer starts as draft", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone Won", domain.OfferPhaseOrder)

		req := &domain.CloneOfferRequest{
			IncludeBudget: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase)
	})

	t.Run("not found", func(t *testing.T) {
		req := &domain.CloneOfferRequest{}

		result, err := svc.CloneOffer(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

// ============================================================================
// Budget Method Tests
// ============================================================================

func TestOfferService_GetBudgetSummary(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("get summary with budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Summary", domain.OfferPhaseDraft)
		// Using gross margin formula: Revenue = Cost / (1 - Margin/100)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 1", 1000, 50, 0) // Cost=1000, Revenue=2000, Profit=1000
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 2", 2000, 50, 1) // Cost=2000, Revenue=4000, Profit=2000
		// Total: Cost=3000, Revenue=6000, Profit=3000

		result, err := svc.GetBudgetSummary(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3000.0, result.TotalCost)
		assert.Equal(t, 6000.0, result.TotalRevenue)
		assert.Equal(t, 3000.0, result.TotalProfit)
		assert.Equal(t, 2, result.ItemCount)
	})

	t.Run("get summary without budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Empty Summary", domain.OfferPhaseDraft)

		result, err := svc.GetBudgetSummary(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0.0, result.TotalCost)
		assert.Equal(t, 0.0, result.TotalRevenue)
		assert.Equal(t, 0, result.ItemCount)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.GetBudgetSummary(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_RecalculateTotals(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("recalculate totals from budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Recalc", domain.OfferPhaseDraft)
		// Using gross margin formula: Revenue = Cost / (1 - Margin/100)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 1", 1000, 50, 0) // Revenue=2000
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 2", 2000, 50, 1) // Revenue=4000
		// Total revenue: 6000

		result, err := svc.RecalculateTotals(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 6000.0, result.Value)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.RecalculateTotals(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

// ============================================================================
// GetByIDWithBudgetItems Tests
// ============================================================================

func TestOfferService_GetByIDWithBudgetItems(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("get offer with budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Get With Items", domain.OfferPhaseDraft)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 1", 1000, 50, 0)
		fixtures.createTestBudgetItem(t, ctx, offer.ID, "Item 2", 2000, 50, 1)

		result, err := svc.GetByIDWithBudgetItems(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, offer.ID, result.ID)
		assert.Len(t, result.BudgetItems, 2)
		assert.NotNil(t, result.BudgetSummary)
		assert.Equal(t, 2, result.BudgetSummary.ItemCount)
	})

	t.Run("get offer without budget items", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Get No Items", domain.OfferPhaseDraft)

		result, err := svc.GetByIDWithBudgetItems(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.BudgetItems, 0)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.GetByIDWithBudgetItems(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

// ============================================================================
// Update Tests with Closed Phase Check
// ============================================================================

func TestOfferService_Update_ClosedPhaseCheck(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("cannot update won offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Won", domain.OfferPhaseOrder)

		req := &domain.UpdateOfferRequest{
			Title: "Updated Title",
			Phase: domain.OfferPhaseOrder,
		}

		result, err := svc.Update(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferAlreadyClosed)
	})

	t.Run("cannot update lost offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Lost", domain.OfferPhaseLost)

		req := &domain.UpdateOfferRequest{
			Title: "Updated Title",
			Phase: domain.OfferPhaseLost,
		}

		result, err := svc.Update(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferAlreadyClosed)
	})

	t.Run("cannot update expired offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Expired", domain.OfferPhaseExpired)

		req := &domain.UpdateOfferRequest{
			Title: "Updated Title",
			Phase: domain.OfferPhaseExpired,
		}

		result, err := svc.Update(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferAlreadyClosed)
	})

	t.Run("can update draft offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Draft", domain.OfferPhaseDraft)

		req := &domain.UpdateOfferRequest{
			Title: "Updated Draft Title",
			Phase: domain.OfferPhaseDraft,
		}

		result, err := svc.Update(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Draft Title", result.Title)
	})
}

// ============================================================================
// Activity Logging Tests
// ============================================================================

func TestOfferService_ActivityLogging(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("send offer logs activity", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Activity Send", domain.OfferPhaseDraft)

		_, err := svc.SendOffer(ctx, offer.ID)
		require.NoError(t, err)

		// Check activity was logged
		activities, err := svc.GetActivities(ctx, offer.ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(activities), 1)

		// Find the send activity (Norwegian: "Tilbud sendt")
		var found bool
		for _, a := range activities {
			if a.Title == "Tilbud sendt" {
				found = true
				break
			}
		}
		assert.True(t, found, "expected 'Tilbud sendt' activity")
	})

	t.Run("accept offer logs activity", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Activity Accept", domain.OfferPhaseSent)

		req := &domain.AcceptOfferRequest{CreateProject: false}
		_, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)

		activities, err := svc.GetActivities(ctx, offer.ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(activities), 1)

		// Find the accept activity (Norwegian: "Ordre akseptert")
		var found bool
		for _, a := range activities {
			if a.Title == "Ordre akseptert" {
				found = true
				break
			}
		}
		assert.True(t, found, "expected 'Ordre akseptert' activity")
	})

	t.Run("clone offer logs activities on both offers", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Activity Clone", domain.OfferPhaseDraft)

		req := &domain.CloneOfferRequest{IncludeBudget: boolPtr(true)}
		cloned, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)

		// Check source offer has clone activity (Norwegian: "Tilbud klonet")
		sourceActivities, err := svc.GetActivities(ctx, offer.ID, 10)
		require.NoError(t, err)
		var foundSource bool
		for _, a := range sourceActivities {
			if a.Title == "Tilbud klonet" {
				foundSource = true
				break
			}
		}
		assert.True(t, foundSource, "expected 'Tilbud klonet' activity on source")

		// Check cloned offer has creation activity (Norwegian: "Tilbud opprettet fra klone")
		clonedActivities, err := svc.GetActivities(ctx, cloned.ID, 10)
		require.NoError(t, err)
		var foundClone bool
		for _, a := range clonedActivities {
			if a.Title == "Tilbud opprettet fra klone" {
				foundClone = true
				break
			}
		}
		assert.True(t, foundClone, "expected 'Tilbud opprettet fra klone' activity on clone")
	})
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestOfferService_EdgeCases(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("reject offer appends reason to existing notes", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Reject Append", domain.OfferPhaseSent)
		offer.Notes = "Some existing notes"
		fixtures.db.Save(offer)

		req := &domain.RejectOfferRequest{
			Reason: "Too expensive",
		}

		result, err := svc.RejectOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.Contains(t, result.Notes, "Some existing notes")
		assert.Contains(t, result.Notes, "Lost reason: Too expensive")
	})

	t.Run("accept with project uses offer title if project name not provided", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Default Project Name", domain.OfferPhaseSent)

		req := &domain.AcceptOfferRequest{
			CreateProject: true,
			// ProjectName not provided
		}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result.Project)
		assert.Equal(t, "Test Default Project Name", result.Project.Name)
	})
}

// ============================================================================
// Offer Number Business Rules Tests
// ============================================================================

func TestOfferService_OfferNumberRules(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("draft offer should not have offer number on creation", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Customer Draft Number Test")

		req := &domain.CreateOfferRequest{
			Title:      "Test Draft No Number",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase)
		assert.Empty(t, result.OfferNumber, "draft offer should not have an offer number")
	})

	t.Run("non-draft offer should get offer number on creation", func(t *testing.T) {
		customer := fixtures.createTestCustomer(t, ctx, "Customer NonDraft Number Test")

		req := &domain.CreateOfferRequest{
			Title:             "Test InProgress With Number",
			CustomerID:        &customer.ID,
			CompanyID:         domain.CompanyStalbygg,
			Phase:             domain.OfferPhaseInProgress,
			ResponsibleUserID: testUserID,
		}

		result, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseInProgress, result.Phase)
		assert.NotEmpty(t, result.OfferNumber, "non-draft offer should have an offer number")
	})

	t.Run("send offer from draft generates offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Send Generates Number", domain.OfferPhaseDraft)
		assert.Empty(t, offer.OfferNumber, "draft offer should start without number")

		result, err := svc.SendOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
		assert.NotEmpty(t, result.OfferNumber, "sent offer should have an offer number")
	})

	t.Run("expire offer from draft generates offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Expire Generates Number", domain.OfferPhaseDraft)
		assert.Empty(t, offer.OfferNumber, "draft offer should start without number")

		result, err := svc.ExpireOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseExpired, result.Phase)
		assert.NotEmpty(t, result.OfferNumber, "expired offer should have an offer number")
	})

	t.Run("update draft to non-draft generates offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Generates Number", domain.OfferPhaseDraft)
		assert.Empty(t, offer.OfferNumber, "draft offer should start without number")

		req := &domain.UpdateOfferRequest{
			Title:             offer.Title,
			Phase:             domain.OfferPhaseInProgress,
			ResponsibleUserID: testUserID,
		}

		result, err := svc.Update(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseInProgress, result.Phase)
		assert.NotEmpty(t, result.OfferNumber, "non-draft offer should have an offer number")
	})

	t.Run("advance from draft to in_progress generates offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Advance Generates Number", domain.OfferPhaseDraft)
		assert.Empty(t, offer.OfferNumber, "draft offer should start without number")

		req := &domain.AdvanceOfferRequest{
			Phase: domain.OfferPhaseInProgress,
		}

		result, err := svc.Advance(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseInProgress, result.Phase)
		assert.NotEmpty(t, result.OfferNumber, "advanced offer should have an offer number")
	})

	t.Run("advance from draft to sent generates offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Advance Sent Generates Number", domain.OfferPhaseDraft)
		assert.Empty(t, offer.OfferNumber, "draft offer should start without number")

		req := &domain.AdvanceOfferRequest{
			Phase: domain.OfferPhaseSent,
		}

		result, err := svc.Advance(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
		assert.NotEmpty(t, result.OfferNumber, "advanced offer should have an offer number")
	})

	t.Run("cannot manually set offer number on draft offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Cannot Set Number On Draft", domain.OfferPhaseDraft)

		_, err := svc.UpdateOfferNumber(ctx, offer.ID, "MANUAL-001")
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrDraftOfferCannotHaveNumber)
	})

	t.Run("cannot clear offer number on non-draft offer", func(t *testing.T) {
		// Create offer in non-draft state (it will get an offer number)
		offer := fixtures.createTestOffer(t, ctx, "Test Cannot Clear Number", domain.OfferPhaseInProgress)

		_, err := svc.UpdateOfferNumber(ctx, offer.ID, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrNonDraftOfferMustHaveNumber)
	})

	t.Run("cloned offer starts as draft without offer number", func(t *testing.T) {
		// Start with a non-draft offer that has an offer number
		offer := fixtures.createTestOffer(t, ctx, "Test Clone No Number", domain.OfferPhaseInProgress)

		req := &domain.CloneOfferRequest{
			IncludeBudget: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase)
		assert.Empty(t, result.OfferNumber, "cloned offer should not have an offer number (starts as draft)")
	})
}

// ============================================================================
// Phase Transition Tests
// ============================================================================

func TestOfferService_PhaseTransitions(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("advance from sent to in_progress clears sent-related dates", func(t *testing.T) {
		// Create a sent offer with sent_date and expiration_date set
		offer := fixtures.createTestOffer(t, ctx, "Test Sent To InProgress", domain.OfferPhaseSent)
		sentDate := time.Now().AddDate(0, 0, -5)
		expirationDate := time.Now().AddDate(0, 1, 0)
		offer.SentDate = &sentDate
		offer.ExpirationDate = &expirationDate
		fixtures.db.Save(offer)

		// Verify dates are set
		reloaded, _ := fixtures.offerRepo.GetByID(ctx, offer.ID)
		require.NotNil(t, reloaded.SentDate, "sent_date should be set before test")
		require.NotNil(t, reloaded.ExpirationDate, "expiration_date should be set before test")

		// Move back to in_progress
		req := &domain.AdvanceOfferRequest{
			Phase: domain.OfferPhaseInProgress,
		}
		result, err := svc.Advance(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseInProgress, result.Phase)

		// Verify dates are cleared
		assert.Nil(t, result.SentDate, "sent_date should be cleared when moving back to in_progress")
		assert.Nil(t, result.ExpirationDate, "expiration_date should be cleared when moving back to in_progress")
	})

	t.Run("advance from sent to in_progress preserves offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Sent To InProgress Number", domain.OfferPhaseSent)
		originalNumber := offer.OfferNumber
		require.NotEmpty(t, originalNumber, "sent offer should have an offer number")

		req := &domain.AdvanceOfferRequest{
			Phase: domain.OfferPhaseInProgress,
		}
		result, err := svc.Advance(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.Equal(t, originalNumber, result.OfferNumber, "offer number should be preserved when moving back to in_progress")
	})

	t.Run("advance from in_progress to sent is allowed", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test InProgress To Sent", domain.OfferPhaseInProgress)

		req := &domain.AdvanceOfferRequest{
			Phase: domain.OfferPhaseSent,
		}
		result, err := svc.Advance(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseSent, result.Phase)
	})

	t.Run("cannot advance to terminal phases", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Cannot Advance Terminal", domain.OfferPhaseSent)

		// Cannot advance to won
		req := &domain.AdvanceOfferRequest{Phase: domain.OfferPhaseOrder}
		_, err := svc.Advance(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrOfferCannotAdvanceToTerminalPhase)

		// Cannot advance to lost
		req = &domain.AdvanceOfferRequest{Phase: domain.OfferPhaseLost}
		_, err = svc.Advance(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrOfferCannotAdvanceToTerminalPhase)

		// Cannot advance to expired
		req = &domain.AdvanceOfferRequest{Phase: domain.OfferPhaseExpired}
		_, err = svc.Advance(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrOfferCannotAdvanceToTerminalPhase)
	})
}

// ============================================================================
// Order Phase Method Tests
// ============================================================================

func TestOfferService_AcceptOrder(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("accept order from sent phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Order Sent", domain.OfferPhaseSent)

		req := &domain.AcceptOrderRequest{
			Notes: "Customer confirmed",
		}

		result, err := svc.AcceptOrder(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Offer)
		assert.Equal(t, domain.OfferPhaseOrder, result.Offer.Phase)
	})

	t.Run("accept order adds O suffix to offer number", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Order Number", domain.OfferPhaseSent)
		originalNumber := offer.OfferNumber

		req := &domain.AcceptOrderRequest{}

		result, err := svc.AcceptOrder(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		// Offer number should have "O" suffix
		assert.Equal(t, originalNumber+"O", result.Offer.OfferNumber)
	})

	t.Run("cannot accept order from draft phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Order Draft", domain.OfferPhaseDraft)

		req := &domain.AcceptOrderRequest{}

		result, err := svc.AcceptOrder(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInSentPhase)
	})

	t.Run("cannot accept order from in_progress phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Order InProgress", domain.OfferPhaseInProgress)

		req := &domain.AcceptOrderRequest{}

		result, err := svc.AcceptOrder(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInSentPhase)
	})

	t.Run("cannot accept order from already order phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Order Already", domain.OfferPhaseOrder)

		req := &domain.AcceptOrderRequest{}

		result, err := svc.AcceptOrder(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInSentPhase)
	})

	t.Run("not found", func(t *testing.T) {
		req := &domain.AcceptOrderRequest{}

		result, err := svc.AcceptOrder(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_UpdateOfferHealth(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("update health in order phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Health Order", domain.OfferPhaseOrder)

		req := &domain.UpdateOfferHealthRequest{
			Health: domain.OfferHealthAtRisk,
		}

		result, err := svc.UpdateOfferHealth(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Health)
		assert.Equal(t, "at_risk", *result.Health)
	})

	t.Run("update health with completion percent", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Health Complete", domain.OfferPhaseOrder)

		completionPct := 75.5
		req := &domain.UpdateOfferHealthRequest{
			Health:            domain.OfferHealthOnTrack,
			CompletionPercent: &completionPct,
		}

		result, err := svc.UpdateOfferHealth(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Health)
		assert.Equal(t, "on_track", *result.Health)
		assert.NotNil(t, result.CompletionPercent)
		assert.Equal(t, 75.5, *result.CompletionPercent)
	})

	t.Run("cannot update health in non-order phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Health Sent", domain.OfferPhaseSent)

		req := &domain.UpdateOfferHealthRequest{
			Health: domain.OfferHealthOnTrack,
		}

		result, err := svc.UpdateOfferHealth(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInOrderPhase)
	})

	t.Run("cannot update health in draft phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Health Draft", domain.OfferPhaseDraft)

		req := &domain.UpdateOfferHealthRequest{
			Health: domain.OfferHealthDelayed,
		}

		result, err := svc.UpdateOfferHealth(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInOrderPhase)
	})

	t.Run("not found", func(t *testing.T) {
		req := &domain.UpdateOfferHealthRequest{
			Health: domain.OfferHealthOnTrack,
		}

		result, err := svc.UpdateOfferHealth(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

func TestOfferService_UpdateOfferSpent(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	// All UpdateOfferSpent calls should now return ErrOfferFinancialFieldReadOnly
	// as spent/invoiced fields are now managed by data warehouse sync

	t.Run("spent field is read-only - returns error", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Spent Order", domain.OfferPhaseOrder)

		req := &domain.UpdateOfferSpentRequest{
			Spent: 25000.50,
		}

		result, err := svc.UpdateOfferSpent(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferFinancialFieldReadOnly)
	})
}

func TestOfferService_UpdateOfferInvoiced(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	// All UpdateOfferInvoiced calls should now return ErrOfferFinancialFieldReadOnly
	// as spent/invoiced fields are now managed by data warehouse sync

	t.Run("invoiced field is read-only - returns error", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Invoiced Order", domain.OfferPhaseOrder)

		req := &domain.UpdateOfferInvoicedRequest{
			Invoiced: 50000,
		}

		result, err := svc.UpdateOfferInvoiced(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferFinancialFieldReadOnly)
	})
}

func TestOfferService_CompleteOffer(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("complete offer from order phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Complete Order", domain.OfferPhaseOrder)

		result, err := svc.CompleteOffer(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.OfferPhaseCompleted, result.Phase)
	})

	t.Run("cannot complete offer from sent phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Complete Sent", domain.OfferPhaseSent)

		result, err := svc.CompleteOffer(ctx, offer.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInOrderPhase)
	})

	t.Run("cannot complete offer from draft phase", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Complete Draft", domain.OfferPhaseDraft)

		result, err := svc.CompleteOffer(ctx, offer.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotInOrderPhase)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.CompleteOffer(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

// ============================================================================
// Project Location Inference Tests
// ============================================================================

// TestOfferService_SyncProjectLocation tests that project location is inferred from offers
func TestOfferService_SyncProjectLocation(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	// Helper to create project directly in DB
	createProject := func(name string) *domain.Project {
		project := &domain.Project{
			Name:  name,
			Phase: domain.ProjectPhaseTilbud,
		}
		err := fixtures.db.Create(project).Error
		require.NoError(t, err)
		return project
	}

	// Helper to get project location from DB
	getProjectLocation := func(projectID uuid.UUID) string {
		var project domain.Project
		err := fixtures.db.First(&project, "id = ?", projectID).Error
		require.NoError(t, err)
		return project.Location
	}

	t.Run("location inferred when all offers have same location", func(t *testing.T) {
		project := createProject("Test Same Location Project")
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Same Location")

		// Create first offer with location and link to project
		req1 := &domain.CreateOfferRequest{
			Title:      "Test Offer 1 Same Location",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
			ProjectID:  &project.ID,
			Location:   "Oslo",
		}
		_, err := svc.Create(ctx, req1)
		require.NoError(t, err)

		// Verify project location is set to "Oslo"
		location := getProjectLocation(project.ID)
		assert.Equal(t, "Oslo", location, "project location should be 'Oslo' when only offer has location 'Oslo'")

		// Create second offer with same location
		req2 := &domain.CreateOfferRequest{
			Title:      "Test Offer 2 Same Location",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
			ProjectID:  &project.ID,
			Location:   "Oslo",
		}
		_, err = svc.Create(ctx, req2)
		require.NoError(t, err)

		// Verify project location is still "Oslo"
		location = getProjectLocation(project.ID)
		assert.Equal(t, "Oslo", location, "project location should still be 'Oslo' when all offers have location 'Oslo'")
	})

	t.Run("location cleared when offers have different locations", func(t *testing.T) {
		project := createProject("Test Different Locations Project")
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Different Locations")

		// Create first offer with location "Oslo"
		req1 := &domain.CreateOfferRequest{
			Title:      "Test Offer 1 Different",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
			ProjectID:  &project.ID,
			Location:   "Oslo",
		}
		_, err := svc.Create(ctx, req1)
		require.NoError(t, err)

		// Verify project location is set to "Oslo"
		location := getProjectLocation(project.ID)
		assert.Equal(t, "Oslo", location)

		// Create second offer with different location "Bergen"
		req2 := &domain.CreateOfferRequest{
			Title:      "Test Offer 2 Different",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
			ProjectID:  &project.ID,
			Location:   "Bergen",
		}
		_, err = svc.Create(ctx, req2)
		require.NoError(t, err)

		// Verify project location is cleared (different locations)
		location = getProjectLocation(project.ID)
		assert.Empty(t, location, "project location should be cleared when offers have different locations")
	})

	t.Run("location cleared when no offers have location", func(t *testing.T) {
		project := createProject("Test No Location Offers Project")
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer No Location")

		// Create offer without location
		req := &domain.CreateOfferRequest{
			Title:      "Test Offer No Location",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
			ProjectID:  &project.ID,
			// Location not set
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Verify project location is empty
		location := getProjectLocation(project.ID)
		assert.Empty(t, location, "project location should be empty when offers have no location")
	})

	t.Run("location updated when offer location changes via update", func(t *testing.T) {
		project := createProject("Test Location Update Project")
		customer := fixtures.createTestCustomer(t, ctx, "Test Customer Location Update")

		// Create offer with location "Oslo"
		createReq := &domain.CreateOfferRequest{
			Title:      "Test Offer Location Update",
			CustomerID: &customer.ID,
			CompanyID:  domain.CompanyStalbygg,
			Phase:      domain.OfferPhaseDraft,
			ProjectID:  &project.ID,
			Location:   "Oslo",
		}
		offer, err := svc.Create(ctx, createReq)
		require.NoError(t, err)

		// Verify project location is "Oslo"
		location := getProjectLocation(project.ID)
		assert.Equal(t, "Oslo", location)

		// Update offer location to "Trondheim"
		// Note: UpdateOfferRequest uses Location field; ProjectID is already linked
		updateReq := &domain.UpdateOfferRequest{
			Title:    "Test Offer Location Update",
			Phase:    domain.OfferPhaseDraft,
			Location: "Trondheim",
		}
		_, err = svc.Update(ctx, offer.ID, updateReq)
		require.NoError(t, err)

		// Verify project location is now "Trondheim"
		location = getProjectLocation(project.ID)
		assert.Equal(t, "Trondheim", location, "project location should update when offer location changes")
	})
}
