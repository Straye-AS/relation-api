package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func setupCustomerServiceTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createCustomerService(db *gorm.DB) *service.CustomerService {
	customerRepo := repository.NewCustomerRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	logger := zap.NewNop()

	return service.NewCustomerService(customerRepo, activityRepo, logger)
}

func createCustomerTestContext() context.Context {
	ctx := auth.WithUserContext(context.Background(), &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin}, // SuperAdmin bypasses company filter
	})
	return ctx
}

func TestCustomerService_Create(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	t.Run("success with valid data", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:          "Test Company AS",
			OrgNumber:     "123456789",
			Email:         "contact@testcompany.no",
			Phone:         "22334455",
			Address:       "Storgata 1",
			City:          "Oslo",
			PostalCode:    "0123",
			Country:       "Norway",
			ContactPerson: "Ola Nordmann",
			ContactEmail:  "ola@testcompany.no",
			ContactPhone:  "+47 99887766",
		}

		customer, err := svc.Create(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, customer)

		assert.NotEqual(t, uuid.Nil, customer.ID)
		assert.Equal(t, req.Name, customer.Name)
		assert.Equal(t, req.OrgNumber, customer.OrgNumber)
		assert.Equal(t, req.Email, customer.Email)
		assert.Equal(t, req.Phone, customer.Phone)
		assert.Equal(t, req.Address, customer.Address)
		assert.Equal(t, req.City, customer.City)
		assert.Equal(t, req.PostalCode, customer.PostalCode)
		assert.Equal(t, req.Country, customer.Country)
		assert.Equal(t, req.ContactPerson, customer.ContactPerson)
		assert.Equal(t, req.ContactEmail, customer.ContactEmail)
		assert.Equal(t, req.ContactPhone, customer.ContactPhone)
	})

	t.Run("succeeds with short org number - no validation", func(t *testing.T) {
		// Note: The current service does not validate org number format,
		// it only checks for duplicates. This test documents current behavior.
		req := &domain.CreateCustomerRequest{
			Name:      "Short Org Company",
			OrgNumber: "12345", // Short org number - currently allowed
			Email:     "test-short-org@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		customer, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, customer)
	})

	t.Run("fails with duplicate org number", func(t *testing.T) {
		// Create first customer
		req1 := &domain.CreateCustomerRequest{
			Name:      "First Company",
			OrgNumber: "987654321",
			Email:     "first@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		_, err := svc.Create(ctx, req1)
		require.NoError(t, err)

		// Try to create second customer with same org number
		req2 := &domain.CreateCustomerRequest{
			Name:      "Second Company",
			OrgNumber: "987654321", // Same org number
			Email:     "second@example.com",
			Phone:     "87654321",
			Country:   "Norway",
		}

		customer, err := svc.Create(ctx, req2)
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.ErrorIs(t, err, service.ErrDuplicateOrgNumber)
	})

	t.Run("fails with invalid email format", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:      "Bad Email Company",
			OrgNumber: "111222333",
			Email:     "not-an-email", // Invalid email
			Phone:     "12345678",
			Country:   "Norway",
		}

		customer, err := svc.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.ErrorIs(t, err, service.ErrInvalidEmailFormat)
	})

	t.Run("fails with invalid contact email format", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:         "Bad Contact Email Company",
			OrgNumber:    "222333444",
			Email:        "valid@example.com",
			Phone:        "12345678",
			Country:      "Norway",
			ContactEmail: "invalid-contact-email", // Invalid contact email
		}

		customer, err := svc.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.Contains(t, err.Error(), "invalid contact email")
	})

	t.Run("fails with invalid phone format - too short", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:      "Short Phone Company",
			OrgNumber: "333444555",
			Email:     "valid@example.com",
			Phone:     "123", // Too short
			Country:   "Norway",
		}

		customer, err := svc.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.ErrorIs(t, err, service.ErrInvalidPhoneFormat)
	})

	t.Run("succeeds with Norwegian phone formats", func(t *testing.T) {
		testCases := []struct {
			name  string
			phone string
		}{
			{"8 digits", "12345678"},
			{"with country code", "+4712345678"},
			{"with spaces", "12 34 56 78"},
			{"with dashes", "12-34-56-78"},
			{"mixed formatting", "+47 123 45 678"},
		}

		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Generate unique org number for each test case
				orgNum := "44455566" + string('0'+byte(i))
				req := &domain.CreateCustomerRequest{
					Name:      "Phone Format Test " + tc.name,
					OrgNumber: orgNum,
					Email:     "valid@example.com",
					Phone:     tc.phone,
					Country:   "Norway",
				}

				customer, err := svc.Create(ctx, req)
				assert.NoError(t, err, "Phone format %s should be valid", tc.phone)
				assert.NotNil(t, customer)
			})
		}
	})
}

func TestCustomerService_GetByID(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	t.Run("success - returns customer with stats", func(t *testing.T) {
		// Create a customer
		req := &domain.CreateCustomerRequest{
			Name:      "Test Company",
			OrgNumber: "555666777",
			Email:     "test@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Get the customer
		customer, err := svc.GetByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, customer)

		assert.Equal(t, created.ID, customer.ID)
		assert.Equal(t, created.Name, customer.Name)
		assert.Equal(t, created.OrgNumber, customer.OrgNumber)
		// TotalValue and ActiveOffers should be 0 for a new customer
		assert.Equal(t, float64(0), customer.TotalValue)
		assert.Equal(t, 0, customer.ActiveOffers)
	})

	t.Run("fails with non-existent ID", func(t *testing.T) {
		customer, err := svc.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.ErrorIs(t, err, service.ErrCustomerNotFound)
	})
}

// Note: GetWithStats is not implemented in CustomerService.
// The GetByID method returns CustomerDTO with TotalValue and ActiveOffers populated.
// For detailed stats, use GetByIDWithDetails which returns CustomerWithDetailsDTO.

func TestCustomerService_Update(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	t.Run("success - update all fields", func(t *testing.T) {
		// Create a customer
		createReq := &domain.CreateCustomerRequest{
			Name:          "Original Company",
			OrgNumber:     "777888999",
			Email:         "original@example.com",
			Phone:         "12345678",
			Address:       "Original Address",
			City:          "Original City",
			PostalCode:    "0001",
			Country:       "Norway",
			ContactPerson: "Original Person",
			ContactEmail:  "original.person@example.com",
			ContactPhone:  "11112222",
		}

		created, err := svc.Create(ctx, createReq)
		require.NoError(t, err)

		// Update the customer
		updateReq := &domain.UpdateCustomerRequest{
			Name:          "Updated Company",
			OrgNumber:     "777888999", // Keep same org number
			Email:         "updated@example.com",
			Phone:         "87654321",
			Address:       "Updated Address",
			City:          "Updated City",
			PostalCode:    "9999",
			Country:       "Norway",
			ContactPerson: "Updated Person",
			ContactEmail:  "updated.person@example.com",
			ContactPhone:  "33334444",
		}

		updated, err := svc.Update(ctx, created.ID, updateReq)
		require.NoError(t, err)
		require.NotNil(t, updated)

		assert.Equal(t, updateReq.Name, updated.Name)
		assert.Equal(t, updateReq.Email, updated.Email)
		assert.Equal(t, updateReq.Phone, updated.Phone)
		assert.Equal(t, updateReq.Address, updated.Address)
		assert.Equal(t, updateReq.City, updated.City)
		assert.Equal(t, updateReq.PostalCode, updated.PostalCode)
		assert.Equal(t, updateReq.ContactPerson, updated.ContactPerson)
		assert.Equal(t, updateReq.ContactEmail, updated.ContactEmail)
		assert.Equal(t, updateReq.ContactPhone, updated.ContactPhone)
	})

	t.Run("success - change org number to valid unique number", func(t *testing.T) {
		createReq := &domain.CreateCustomerRequest{
			Name:      "Org Change Company",
			OrgNumber: "888999000",
			Email:     "orgchange@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, createReq)
		require.NoError(t, err)

		updateReq := &domain.UpdateCustomerRequest{
			Name:      "Org Change Company",
			OrgNumber: "111000999", // New valid org number
			Email:     "orgchange@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		updated, err := svc.Update(ctx, created.ID, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "111000999", updated.OrgNumber)
	})

	t.Run("fails - change org number to existing number", func(t *testing.T) {
		// Create first customer
		req1 := &domain.CreateCustomerRequest{
			Name:      "First Update Company",
			OrgNumber: "999000111",
			Email:     "first.update@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		_, err := svc.Create(ctx, req1)
		require.NoError(t, err)

		// Create second customer
		req2 := &domain.CreateCustomerRequest{
			Name:      "Second Update Company",
			OrgNumber: "999000222",
			Email:     "second.update@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		created2, err := svc.Create(ctx, req2)
		require.NoError(t, err)

		// Try to update second customer with first customer's org number
		updateReq := &domain.UpdateCustomerRequest{
			Name:      "Second Update Company",
			OrgNumber: "999000111", // First customer's org number
			Email:     "second.update@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		updated, err := svc.Update(ctx, created2.ID, updateReq)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.ErrorIs(t, err, service.ErrDuplicateOrgNumber)
	})

	t.Run("fails - invalid email on update", func(t *testing.T) {
		createReq := &domain.CreateCustomerRequest{
			Name:      "Email Update Company",
			OrgNumber: "999000333",
			Email:     "valid@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, createReq)
		require.NoError(t, err)

		updateReq := &domain.UpdateCustomerRequest{
			Name:      "Email Update Company",
			OrgNumber: "999000333",
			Email:     "invalid-email", // Invalid email
			Phone:     "12345678",
			Country:   "Norway",
		}

		updated, err := svc.Update(ctx, created.ID, updateReq)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.ErrorIs(t, err, service.ErrInvalidEmailFormat)
	})

	t.Run("fails - non-existent customer", func(t *testing.T) {
		updateReq := &domain.UpdateCustomerRequest{
			Name:      "Non-existent Company",
			OrgNumber: "999000444",
			Email:     "test@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		updated, err := svc.Update(ctx, uuid.New(), updateReq)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.ErrorIs(t, err, service.ErrCustomerNotFound)
	})
}

func TestCustomerService_Delete(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	t.Run("success - delete customer without dependencies", func(t *testing.T) {
		// Create a customer
		req := &domain.CreateCustomerRequest{
			Name:      "Delete Test Company",
			OrgNumber: "111222331",
			Email:     "delete@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Delete the customer
		err = svc.Delete(ctx, created.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = svc.GetByID(ctx, created.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrCustomerNotFound)
	})

	t.Run("fails - customer has active deal", func(t *testing.T) {
		// Create a customer
		req := &domain.CreateCustomerRequest{
			Name:      "Customer With Deal",
			OrgNumber: "111222332",
			Email:     "deal@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Create an active deal for this customer
		userCtx, _ := auth.FromContext(ctx)
		deal := &domain.Deal{
			Title:      "Active Deal",
			CustomerID: created.ID,
			CompanyID:  domain.CompanyStalbygg,
			Stage:      domain.DealStageLead, // Active stage
			OwnerID:    userCtx.UserID.String(),
			Value:      100000,
		}
		err = db.Omit(clause.Associations).Create(deal).Error
		require.NoError(t, err)

		// Try to delete the customer
		err = svc.Delete(ctx, created.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrCustomerHasActiveDependencies)
		assert.Contains(t, err.Error(), "active deals")
	})

	t.Run("fails - customer has active project", func(t *testing.T) {
		// Create a customer
		req := &domain.CreateCustomerRequest{
			Name:      "Customer With Project",
			OrgNumber: "111222333",
			Email:     "project@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Create an active project for this customer
		userCtx, _ := auth.FromContext(ctx)
		project := &domain.Project{
			Name:       "Active Project",
			CustomerID: created.ID,
			CompanyID:  domain.CompanyStalbygg,
			Status:     domain.ProjectStatusActive,
			StartDate:  time.Now(),
			ManagerID:  userCtx.UserID.String(),
		}
		err = db.Omit(clause.Associations).Create(project).Error
		require.NoError(t, err)

		// Try to delete the customer
		err = svc.Delete(ctx, created.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrCustomerHasActiveDependencies)
		assert.Contains(t, err.Error(), "active projects")
	})

	t.Run("success - delete customer with completed deal", func(t *testing.T) {
		// Create a customer
		req := &domain.CreateCustomerRequest{
			Name:      "Customer With Won Deal",
			OrgNumber: "111222334",
			Email:     "won@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Create a won (completed) deal for this customer
		userCtx, _ := auth.FromContext(ctx)
		deal := &domain.Deal{
			Title:      "Won Deal",
			CustomerID: created.ID,
			CompanyID:  domain.CompanyStalbygg,
			Stage:      domain.DealStageWon, // Completed stage
			OwnerID:    userCtx.UserID.String(),
			Value:      100000,
		}
		err = db.Omit(clause.Associations).Create(deal).Error
		require.NoError(t, err)

		// Should be able to delete the customer
		err = svc.Delete(ctx, created.ID)
		assert.NoError(t, err)
	})

	t.Run("fails - non-existent customer", func(t *testing.T) {
		err := svc.Delete(ctx, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrCustomerNotFound)
	})
}

func TestCustomerService_List(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	// Create multiple customers for testing with unique identifiers
	// Using a unique prefix to avoid conflicts with other tests
	uniquePrefix := "ListTest_" + uuid.New().String()[:8]
	// Generate unique org numbers based on timestamp
	baseOrgNum := time.Now().UnixNano() % 100000000
	customers := []struct {
		name      string
		orgNumber string
	}{
		{uniquePrefix + "_Tech Solutions AS", fmt.Sprintf("%09d", (baseOrgNum+1)%1000000000)},
		{uniquePrefix + "_Finance Corp AS", fmt.Sprintf("%09d", (baseOrgNum+2)%1000000000)},
		{uniquePrefix + "_Tech Innovation AS", fmt.Sprintf("%09d", (baseOrgNum+3)%1000000000)},
		{uniquePrefix + "_Healthcare Ltd", fmt.Sprintf("%09d", (baseOrgNum+4)%1000000000)},
		{uniquePrefix + "_Tech Partners AS", fmt.Sprintf("%09d", (baseOrgNum+5)%1000000000)},
	}

	for _, c := range customers {
		req := &domain.CreateCustomerRequest{
			Name:      c.name,
			OrgNumber: c.orgNumber,
			Email:     "test@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	t.Run("list all customers", func(t *testing.T) {
		// Search by unique prefix to only find our test customers
		result, err := svc.List(ctx, 1, 20, uniquePrefix)
		require.NoError(t, err)

		assert.Equal(t, int64(5), result.Total)
		data, ok := result.Data.([]domain.CustomerDTO)
		assert.True(t, ok)
		assert.Equal(t, 5, len(data))
	})

	t.Run("list with search - by name", func(t *testing.T) {
		// Search for unique prefix + Tech
		result, err := svc.List(ctx, 1, 20, uniquePrefix+"_Tech")
		require.NoError(t, err)

		assert.Equal(t, int64(3), result.Total)
		data, ok := result.Data.([]domain.CustomerDTO)
		assert.True(t, ok)
		assert.Equal(t, 3, len(data))
		for _, c := range data {
			assert.Contains(t, c.Name, "Tech")
		}
	})

	t.Run("list with search - by org number", func(t *testing.T) {
		// Search for the third customer's org number (Tech Innovation)
		orgNum := customers[2].orgNumber
		result, err := svc.List(ctx, 1, 20, orgNum)
		require.NoError(t, err)

		assert.Equal(t, int64(1), result.Total)
		data, ok := result.Data.([]domain.CustomerDTO)
		assert.True(t, ok)
		assert.Equal(t, 1, len(data))
	})

	t.Run("pagination", func(t *testing.T) {
		// Get first page, filtered by unique prefix
		page1, err := svc.List(ctx, 1, 2, uniquePrefix)
		require.NoError(t, err)

		data1, ok := page1.Data.([]domain.CustomerDTO)
		assert.True(t, ok)
		require.Equal(t, 2, len(data1), "First page should have 2 customers")
		assert.Equal(t, 2, page1.PageSize)

		// Get second page
		page2, err := svc.List(ctx, 2, 2, uniquePrefix)
		require.NoError(t, err)

		data2, ok := page2.Data.([]domain.CustomerDTO)
		assert.True(t, ok)
		require.Equal(t, 2, len(data2), "Second page should have 2 customers")

		// Ensure different customers on each page
		assert.NotEqual(t, data1[0].ID, data2[0].ID)
	})

	t.Run("clamps page size to max 200", func(t *testing.T) {
		result, err := svc.List(ctx, 1, 500, "")
		require.NoError(t, err)

		assert.Equal(t, 200, result.PageSize)
	})

	t.Run("uses default page size when too small", func(t *testing.T) {
		result, err := svc.List(ctx, 1, 0, "")
		require.NoError(t, err)

		assert.Equal(t, 20, result.PageSize)
	})

	t.Run("uses default page when too small", func(t *testing.T) {
		result, err := svc.List(ctx, 0, 20, "")
		require.NoError(t, err)

		assert.Equal(t, 1, result.Page)
	})
}

func TestCustomerService_ValidationHelpers(t *testing.T) {
	t.Run("org number validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			orgNumber   string
			shouldError bool
		}{
			{"valid 9 digits", "123456789", false},
			{"too short", "12345678", true},
			{"too long", "1234567890", true},
			{"with letters", "12345678A", true},
			{"with spaces", "123 456 789", true},
			{"with dashes", "123-456-789", true},
			{"empty", "", true},
		}

		db := setupCustomerServiceTestDB(t)
		svc := createCustomerService(db)
		ctx := createCustomerTestContext()

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := &domain.CreateCustomerRequest{
					Name:      "Test Company " + tc.name,
					OrgNumber: tc.orgNumber,
					Email:     "test@example.com",
					Phone:     "12345678",
					Country:   "Norway",
				}

				_, err := svc.Create(ctx, req)
				if tc.shouldError {
					assert.Error(t, err, "Expected error for org number: %s", tc.orgNumber)
				} else {
					// Clean up for next test
					assert.NoError(t, err, "Expected no error for org number: %s", tc.orgNumber)
				}
			})
		}
	})

	t.Run("email validation", func(t *testing.T) {
		testCases := []struct {
			name        string
			email       string
			shouldError bool
		}{
			{"valid email", "test@example.com", false},
			{"valid with subdomain", "test@sub.example.com", false},
			{"valid with plus", "test+tag@example.com", false},
			{"empty - optional", "", false},
			{"no at sign", "testexample.com", true},
			{"no domain", "test@", true},
			{"no tld", "test@example", true},
			{"just at sign", "@", true},
		}

		db := setupCustomerServiceTestDB(t)
		svc := createCustomerService(db)
		ctx := createCustomerTestContext()

		for i, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Generate unique org number for each test
				orgNum := "31122233" + string('0'+byte(i%10))
				req := &domain.CreateCustomerRequest{
					Name:      "Email Test " + tc.name,
					OrgNumber: orgNum,
					Email:     tc.email,
					Phone:     "12345678",
					Country:   "Norway",
				}

				_, err := svc.Create(ctx, req)
				if tc.shouldError {
					assert.Error(t, err, "Expected error for email: %s", tc.email)
				} else {
					assert.NoError(t, err, "Expected no error for email: %s", tc.email)
				}
			})
		}
	})
}

func TestCustomerService_ActivityLogging(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	activityRepo := repository.NewActivityRepository(db)
	ctx := createCustomerTestContext()

	t.Run("logs activity on create", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:      "Activity Log Test Company",
			OrgNumber: "411222331",
			Email:     "activity@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Check for activity log
		activities, err := activityRepo.ListByTarget(ctx, domain.ActivityTargetCustomer, created.ID, 10)
		require.NoError(t, err)
		require.NotEmpty(t, activities)

		// Find the create activity
		var found bool
		for _, a := range activities {
			if a.Title == "Customer created" {
				found = true
				assert.Contains(t, a.Body, created.Name)
				break
			}
		}
		assert.True(t, found, "Should have a 'Customer created' activity")
	})

	t.Run("logs activity on update", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:      "Activity Update Test",
			OrgNumber: "411222332",
			Email:     "update.activity@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		updateReq := &domain.UpdateCustomerRequest{
			Name:      "Activity Update Test - Updated",
			OrgNumber: "411222332",
			Email:     "update.activity@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		_, err = svc.Update(ctx, created.ID, updateReq)
		require.NoError(t, err)

		// Check for activity log
		activities, err := activityRepo.ListByTarget(ctx, domain.ActivityTargetCustomer, created.ID, 10)
		require.NoError(t, err)

		// Find the update activity
		var found bool
		for _, a := range activities {
			if a.Title == "Customer updated" {
				found = true
				assert.Contains(t, a.Body, "name:")
				break
			}
		}
		assert.True(t, found, "Should have a 'Customer updated' activity")
	})

	t.Run("logs activity on delete", func(t *testing.T) {
		req := &domain.CreateCustomerRequest{
			Name:      "Activity Delete Test",
			OrgNumber: "411222333",
			Email:     "delete.activity@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}

		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		customerID := created.ID

		err = svc.Delete(ctx, customerID)
		require.NoError(t, err)

		// Check for activity log (customer is deleted, but activity should remain)
		activities, err := activityRepo.ListByTarget(ctx, domain.ActivityTargetCustomer, customerID, 10)
		require.NoError(t, err)

		// Find the delete activity
		var found bool
		for _, a := range activities {
			if a.Title == "Customer deleted" {
				found = true
				assert.Contains(t, a.Body, "Activity Delete Test")
				break
			}
		}
		assert.True(t, found, "Should have a 'Customer deleted' activity")
	})
}
