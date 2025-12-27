package service_test

import (
	"context"
	"fmt"
	"os"
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

func setupInquiryTestDB(t *testing.T) *gorm.DB {
	host := getEnvOrDefaultInquiry("DATABASE_HOST", "localhost")
	port := getEnvOrDefaultInquiry("DATABASE_PORT", "5433")
	user := getEnvOrDefaultInquiry("DATABASE_USER", "relation_user")
	password := getEnvOrDefaultInquiry("DATABASE_PASSWORD", "relation_password")
	dbname := getEnvOrDefaultInquiry("DATABASE_NAME", "relation_test")

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

func getEnvOrDefaultInquiry(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupInquiryTestService(t *testing.T, db *gorm.DB) (*service.InquiryService, *inquiryTestFixtures) {
	log := zap.NewNop()

	offerRepo := repository.NewOfferRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	companyRepo := repository.NewCompanyRepository(db)
	userRepo := repository.NewUserRepository(db)

	companyService := service.NewCompanyServiceWithRepo(companyRepo, userRepo, log)

	svc := service.NewInquiryService(
		offerRepo,
		customerRepo,
		activityRepo,
		userRepo,
		companyService,
		log,
		db,
	)

	fixtures := &inquiryTestFixtures{
		db:           db,
		customerRepo: customerRepo,
		offerRepo:    offerRepo,
	}

	return svc, fixtures
}

type inquiryTestFixtures struct {
	db           *gorm.DB
	customerRepo *repository.CustomerRepository
	offerRepo    *repository.OfferRepository
}

func (f *inquiryTestFixtures) createTestCustomer(t *testing.T, ctx context.Context, name string) *domain.Customer {
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

func (f *inquiryTestFixtures) createTestInquiry(t *testing.T, ctx context.Context, title string, companyID domain.CompanyID) *domain.Offer {
	customer := f.createTestCustomer(t, ctx, "Customer for "+title)

	inquiry := &domain.Offer{
		Title:        title,
		CustomerID:   &customer.ID,
		CustomerName: customer.Name,
		CompanyID:    companyID,
		Phase:        domain.OfferPhaseDraft, // Inquiries are always draft
		Probability:  0,
		Value:        0,
		Status:       domain.OfferStatusActive,
		Description:  "Test inquiry description",
	}

	err := f.offerRepo.Create(ctx, inquiry)
	require.NoError(t, err)
	return inquiry
}

func (f *inquiryTestFixtures) createTestOffer(t *testing.T, ctx context.Context, title string, phase domain.OfferPhase, companyID domain.CompanyID) *domain.Offer {
	customer := f.createTestCustomer(t, ctx, "Customer for "+title)

	offer := &domain.Offer{
		Title:             title,
		CustomerID:        &customer.ID,
		CustomerName:      customer.Name,
		CompanyID:         companyID,
		Phase:             phase,
		Probability:       50,
		Value:             10000,
		Status:            domain.OfferStatusActive,
		ResponsibleUserID: "test-user-id",
		Description:       "Test offer description",
	}

	// Non-draft offers should have an offer number
	if phase != domain.OfferPhaseDraft {
		offer.OfferNumber = fmt.Sprintf("TEST-%s-%d", domain.GetCompanyPrefix(companyID), time.Now().UnixNano())
	}

	err := f.offerRepo.Create(ctx, offer)
	require.NoError(t, err)
	return offer
}

func (f *inquiryTestFixtures) cleanup(t *testing.T) {
	f.db.Exec("DELETE FROM activities WHERE target_type = 'Offer' OR target_type = 'Customer'")
	f.db.Exec("DELETE FROM offers WHERE title LIKE 'Test%'")
	f.db.Exec("DELETE FROM customers WHERE name LIKE 'Customer for%'")
}

func createInquiryTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@straye.no",
		Roles:       []domain.UserRoleType{domain.RoleManager},
		CompanyID:   domain.CompanyGruppen,
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

// ============================================================================
// UpdateCompany Tests
// ============================================================================

func TestInquiryService_UpdateCompany(t *testing.T) {
	db := setupInquiryTestDB(t)
	svc, fixtures := setupInquiryTestService(t, db)
	t.Cleanup(func() { fixtures.cleanup(t) })

	ctx := createInquiryTestContext()

	t.Run("update company on draft inquiry", func(t *testing.T) {
		inquiry := fixtures.createTestInquiry(t, ctx, "Test Update Company Draft", domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, inquiry.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.CompanyStalbygg, result.CompanyID)
	})

	t.Run("update company to different valid company", func(t *testing.T) {
		inquiry := fixtures.createTestInquiry(t, ctx, "Test Update Company Different", domain.CompanyStalbygg)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyTak,
		}

		result, err := svc.UpdateCompany(ctx, inquiry.ID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, domain.CompanyTak, result.CompanyID)
	})

	t.Run("cannot update company on in_progress offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Company InProgress", domain.OfferPhaseInProgress, domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrNotAnInquiry)
	})

	t.Run("cannot update company on sent offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Company Sent", domain.OfferPhaseSent, domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrNotAnInquiry)
	})

	t.Run("cannot update company on order", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Company Order", domain.OfferPhaseOrder, domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrNotAnInquiry)
	})

	t.Run("cannot update company on completed offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Company Completed", domain.OfferPhaseCompleted, domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrNotAnInquiry)
	})

	t.Run("cannot update company on lost offer", func(t *testing.T) {
		offer := fixtures.createTestOffer(t, ctx, "Test Update Company Lost", domain.OfferPhaseLost, domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, offer.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrNotAnInquiry)
	})

	t.Run("invalid company ID", func(t *testing.T) {
		inquiry := fixtures.createTestInquiry(t, ctx, "Test Update Company Invalid", domain.CompanyGruppen)

		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyID("invalid_company"),
		}

		result, err := svc.UpdateCompany(ctx, inquiry.ID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInvalidCompanyID)
	})

	t.Run("inquiry not found", func(t *testing.T) {
		req := &domain.UpdateInquiryCompanyRequest{
			CompanyID: domain.CompanyStalbygg,
		}

		result, err := svc.UpdateCompany(ctx, uuid.New(), req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, service.ErrInquiryNotFound)
	})

	t.Run("update company to all valid companies", func(t *testing.T) {
		validCompanies := []domain.CompanyID{
			domain.CompanyGruppen,
			domain.CompanyStalbygg,
			domain.CompanyHybridbygg,
			domain.CompanyIndustri,
			domain.CompanyTak,
			domain.CompanyMontasje,
		}

		for _, companyID := range validCompanies {
			t.Run(string(companyID), func(t *testing.T) {
				inquiry := fixtures.createTestInquiry(t, ctx, "Test Update Company "+string(companyID), domain.CompanyGruppen)

				req := &domain.UpdateInquiryCompanyRequest{
					CompanyID: companyID,
				}

				result, err := svc.UpdateCompany(ctx, inquiry.ID, req)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, companyID, result.CompanyID)
			})
		}
	})
}
