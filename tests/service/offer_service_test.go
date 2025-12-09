package service_test

import (
	"context"
	"testing"

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

// Helper functions for pointer values
func boolPtr(b bool) *bool { return &b }

// TestOfferService is an integration test suite for OfferService
// Requires a running PostgreSQL database with migrations applied

func setupOfferTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=relation_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}
	return db
}

func setupOfferTestService(t *testing.T, db *gorm.DB) (*service.OfferService, *offerTestFixtures) {
	log := zap.NewNop()

	offerRepo := repository.NewOfferRepository(db)
	offerItemRepo := repository.NewOfferItemRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	dimensionRepo := repository.NewBudgetDimensionRepository(db)
	fileRepo := repository.NewFileRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	companyRepo := repository.NewCompanyRepository(db)

	companyService := service.NewCompanyService(companyRepo, log)

	svc := service.NewOfferService(
		offerRepo,
		offerItemRepo,
		customerRepo,
		projectRepo,
		dimensionRepo,
		fileRepo,
		activityRepo,
		companyService,
		log,
		db,
	)

	fixtures := &offerTestFixtures{
		db:            db,
		customerRepo:  customerRepo,
		offerRepo:     offerRepo,
		dimensionRepo: dimensionRepo,
	}

	return svc, fixtures
}

type offerTestFixtures struct {
	db            *gorm.DB
	customerRepo  *repository.CustomerRepository
	offerRepo     *repository.OfferRepository
	dimensionRepo *repository.BudgetDimensionRepository
}

func (f *offerTestFixtures) createTestCustomer(t *testing.T, ctx context.Context, name string) *domain.Customer {
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

func (f *offerTestFixtures) createTestOffer(t *testing.T, ctx context.Context, title string, phase domain.OfferPhase) *domain.Offer {
	customer := f.createTestCustomer(t, ctx, "Customer for "+title)

	offer := &domain.Offer{
		Title:             title,
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         domain.CompanyStalbygg,
		Phase:             phase,
		Probability:       50,
		Value:             10000,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: "test-user-id",
		Description:       "Test offer description",
	}
	err := f.offerRepo.Create(ctx, offer)
	require.NoError(t, err)
	return offer
}

func (f *offerTestFixtures) createTestBudgetDimension(t *testing.T, ctx context.Context, offerID uuid.UUID, name string, cost, revenue float64, order int) *domain.BudgetDimension {
	dimension := &domain.BudgetDimension{
		ParentType:   domain.BudgetParentOffer,
		ParentID:     offerID,
		CustomName:   name,
		Cost:         cost,
		Revenue:      revenue,
		DisplayOrder: order,
	}
	err := f.dimensionRepo.Create(ctx, dimension)
	require.NoError(t, err)
	return dimension
}

func (f *offerTestFixtures) cleanup(t *testing.T) {
	f.db.Exec("DELETE FROM activities WHERE target_type = 'Offer' OR target_type = 'Project' OR target_type = 'Customer'")
	f.db.Exec("DELETE FROM budget_dimensions WHERE parent_type = 'offer' OR parent_type = 'project'")
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
		offer := fixtures.createTestOffer(t, ctx, "Test Won Offer Send", domain.OfferPhaseWon)

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
		assert.Equal(t, domain.OfferPhaseWon, result.Offer.Phase)
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
		assert.Equal(t, domain.OfferPhaseWon, result.Offer.Phase)
		assert.Equal(t, "New Project from Offer", result.Project.Name)
		assert.Equal(t, offer.Value, result.Project.Budget)
	})

	t.Run("accept offer with project creation clones budget dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Clone Dims", domain.OfferPhaseSent)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dimension 1", 1000, 1500, 0)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dimension 2", 2000, 3000, 1)

		req := &domain.AcceptOfferRequest{
			CreateProject: true,
		}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result.Project)

		// Verify dimensions were cloned to project
		projectDims, err := fixtures.dimensionRepo.GetByParent(ctx, domain.BudgetParentProject, result.Project.ID)
		require.NoError(t, err)
		assert.Len(t, projectDims, 2)
	})

	t.Run("cannot accept draft offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Draft", domain.OfferPhaseDraft)

		req := &domain.AcceptOfferRequest{CreateProject: false}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotSent)
	})

	t.Run("cannot accept already won offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Accept Won", domain.OfferPhaseWon)

		req := &domain.AcceptOfferRequest{CreateProject: false}

		result, err := svc.AcceptOffer(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotSent)
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
		offer := fixtures.createTestOffer(t, ctx, "Test Expire Won", domain.OfferPhaseWon)

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
			IncludeDimensions: boolPtr(true),
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
			NewTitle:          "My Custom Clone Title",
			IncludeDimensions: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "My Custom Clone Title", result.Title)
		assert.Equal(t, domain.OfferPhaseDraft, result.Phase) // Cloned offers start as draft
	})

	t.Run("clone offer with budget dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone With Dims", domain.OfferPhaseDraft)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Clone Dim 1", 1000, 1500, 0)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Clone Dim 2", 2000, 3000, 1)

		req := &domain.CloneOfferRequest{
			IncludeDimensions: boolPtr(true),
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify dimensions were cloned
		clonedDims, err := fixtures.dimensionRepo.GetByParent(ctx, domain.BudgetParentOffer, result.ID)
		require.NoError(t, err)
		assert.Len(t, clonedDims, 2)
		assert.Equal(t, "Clone Dim 1", clonedDims[0].CustomName)
		assert.Equal(t, "Clone Dim 2", clonedDims[1].CustomName)
	})

	t.Run("clone offer without budget dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone No Dims", domain.OfferPhaseDraft)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "No Clone Dim", 1000, 1500, 0)

		req := &domain.CloneOfferRequest{
			IncludeDimensions: boolPtr(false), // Explicitly don't clone dimensions
		}

		result, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify dimensions were NOT cloned
		clonedDims, err := fixtures.dimensionRepo.GetByParent(ctx, domain.BudgetParentOffer, result.ID)
		require.NoError(t, err)
		assert.Len(t, clonedDims, 0)
	})

	t.Run("clone won offer starts as draft", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Clone Won", domain.OfferPhaseWon)

		req := &domain.CloneOfferRequest{
			IncludeDimensions: boolPtr(true),
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

	t.Run("get summary with dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Summary", domain.OfferPhaseDraft)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dim 1", 1000, 1500, 0)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dim 2", 2000, 3000, 1)
		// Total: Cost=3000, Revenue=4500

		result, err := svc.GetBudgetSummary(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3000.0, result.TotalCost)
		assert.Equal(t, 4500.0, result.TotalRevenue)
		assert.Equal(t, 1500.0, result.TotalProfit)
		assert.Equal(t, 2, result.DimensionCount)
	})

	t.Run("get summary without dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Empty Summary", domain.OfferPhaseDraft)

		result, err := svc.GetBudgetSummary(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0.0, result.TotalCost)
		assert.Equal(t, 0.0, result.TotalRevenue)
		assert.Equal(t, 0, result.DimensionCount)
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

	t.Run("recalculate totals from dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Recalc", domain.OfferPhaseDraft)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dim 1", 1000, 1500, 0)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dim 2", 2000, 3000, 1)
		// Total revenue: 4500

		result, err := svc.RecalculateTotals(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 4500.0, result.Value)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.RecalculateTotals(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrOfferNotFound)
	})
}

// ============================================================================
// GetByIDWithBudgetDimensions Tests
// ============================================================================

func TestOfferService_GetByIDWithBudgetDimensions(t *testing.T) {
	db := setupOfferTestDB(t)
	svc, fixtures := setupOfferTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createOfferTestContext()

	t.Run("get offer with dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Get With Dims", domain.OfferPhaseDraft)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dim 1", 1000, 1500, 0)
		fixtures.createTestBudgetDimension(t, ctx, offer.ID, "Dim 2", 2000, 3000, 1)

		result, err := svc.GetByIDWithBudgetDimensions(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, offer.ID, result.ID)
		assert.Len(t, result.BudgetDimensions, 2)
		assert.NotNil(t, result.BudgetSummary)
		assert.Equal(t, 2, result.BudgetSummary.DimensionCount)
	})

	t.Run("get offer without dimensions", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Get No Dims", domain.OfferPhaseDraft)

		result, err := svc.GetByIDWithBudgetDimensions(ctx, offer.ID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.BudgetDimensions, 0)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := svc.GetByIDWithBudgetDimensions(ctx, uuid.New())
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
		offer := fixtures.createTestOffer(t, ctx, "Test Update Won", domain.OfferPhaseWon)

		req := &domain.UpdateOfferRequest{
			Title: "Updated Title",
			Phase: domain.OfferPhaseWon,
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

		// Find the send activity
		var found bool
		for _, a := range activities {
			if a.Title == "Offer sent" {
				found = true
				break
			}
		}
		assert.True(t, found, "expected 'Offer sent' activity")
	})

	t.Run("accept offer logs activity", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Activity Accept", domain.OfferPhaseSent)

		req := &domain.AcceptOfferRequest{CreateProject: false}
		_, err := svc.AcceptOffer(ctx, offer.ID, req)
		require.NoError(t, err)

		activities, err := svc.GetActivities(ctx, offer.ID, 10)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(activities), 1)

		var found bool
		for _, a := range activities {
			if a.Title == "Offer accepted" {
				found = true
				break
			}
		}
		assert.True(t, found, "expected 'Offer accepted' activity")
	})

	t.Run("clone offer logs activities on both offers", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Activity Clone", domain.OfferPhaseDraft)

		req := &domain.CloneOfferRequest{IncludeDimensions: boolPtr(true)}
		cloned, err := svc.CloneOffer(ctx, offer.ID, req)
		require.NoError(t, err)

		// Check source offer has clone activity
		sourceActivities, err := svc.GetActivities(ctx, offer.ID, 10)
		require.NoError(t, err)
		var foundSource bool
		for _, a := range sourceActivities {
			if a.Title == "Offer cloned" {
				foundSource = true
				break
			}
		}
		assert.True(t, foundSource, "expected 'Offer cloned' activity on source")

		// Check cloned offer has creation activity
		clonedActivities, err := svc.GetActivities(ctx, cloned.ID, 10)
		require.NoError(t, err)
		var foundClone bool
		for _, a := range clonedActivities {
			if a.Title == "Offer created from clone" {
				foundClone = true
				break
			}
		}
		assert.True(t, foundClone, "expected 'Offer created from clone' activity on clone")
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
