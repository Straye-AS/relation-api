package service_test

import (
	"context"
	"testing"

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

	req := &domain.CreateCustomerRequest{
		Name:          "Test Company",
		OrgNumber:     "123456789",
		Email:         "test@example.com",
		Phone:         "1234567890",
		Address:       "123 Main St",
		City:          "Anytown",
		PostalCode:    "12345",
		Country:       "Norway",
		ContactPerson: "John Doe",
		ContactEmail:  "john.doe@example.com",
		ContactPhone:  "+1234567890",
	}

	customer, err := svc.Create(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, customer)
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
	assert.NotEqual(t, uuid.Nil, customer.ID)
}

func TestCustomerService_GetByID(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	// Create a customer first
	req := &domain.CreateCustomerRequest{
		Name:      "Test Company",
		OrgNumber: "123456780",
		Email:     "test@example.com",
		Phone:     "1234567890",
		Country:   "Norway",
	}

	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	// Get the customer
	customer, err := svc.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, customer.ID)
	assert.Equal(t, created.Name, customer.Name)
}

func TestCustomerService_Update(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	// Create a customer
	createReq := &domain.CreateCustomerRequest{
		Name:          "Original Name",
		OrgNumber:     "123456781",
		Email:         "test@example.com",
		Phone:         "1234567890",
		Address:       "123 Main St",
		City:          "Anytown",
		PostalCode:    "12345",
		Country:       "Norway",
		ContactPerson: "John Doe",
		ContactEmail:  "john.doe@example.com",
		ContactPhone:  "+1234567890",
	}

	created, err := svc.Create(ctx, createReq)
	require.NoError(t, err)

	// Update the customer
	updateReq := &domain.UpdateCustomerRequest{
		Name:          "Updated Name",
		OrgNumber:     "123456781",
		Email:         "test@example.com",
		Phone:         "1234567890",
		Address:       "123 Main St",
		City:          "Anytown",
		PostalCode:    "12345",
		Country:       "Norway",
		ContactPerson: "John Doe",
		ContactEmail:  "john.doe@example.com",
		ContactPhone:  "+1234567890",
	}

	updated, err := svc.Update(ctx, created.ID, updateReq)
	assert.NoError(t, err)
	assert.Equal(t, updateReq.Name, updated.Name)
	assert.Equal(t, updateReq.OrgNumber, updated.OrgNumber)
	assert.Equal(t, updateReq.Email, updated.Email)
	assert.Equal(t, updateReq.Phone, updated.Phone)
	assert.Equal(t, updateReq.Address, updated.Address)
	assert.Equal(t, updateReq.City, updated.City)
	assert.Equal(t, updateReq.PostalCode, updated.PostalCode)
	assert.Equal(t, updateReq.Country, updated.Country)
	assert.Equal(t, updateReq.ContactPerson, updated.ContactPerson)
	assert.Equal(t, updateReq.ContactEmail, updated.ContactEmail)
	assert.Equal(t, updateReq.ContactPhone, updated.ContactPhone)
}

func TestCustomerService_List(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	// Create multiple customers
	customers := []struct {
		name      string
		orgNumber string
	}{
		{"Tech Corp", "123456782"},
		{"Finance Inc", "123456783"},
		{"Tech Solutions", "123456784"},
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

	// List customers
	result, err := svc.List(ctx, 1, 20, "")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, int64(3))
	assert.GreaterOrEqual(t, len(result.Data), 3)

	// Search customers
	result, err = svc.List(ctx, 1, 20, "Tech")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, int64(2))
	assert.GreaterOrEqual(t, len(result.Data), 2)
}

func TestCustomerService_Delete(t *testing.T) {
	db := setupCustomerServiceTestDB(t)
	svc := createCustomerService(db)
	ctx := createCustomerTestContext()

	// Create a customer
	req := &domain.CreateCustomerRequest{
		Name:      "Test Company",
		OrgNumber: "123456785",
		Email:     "test@example.com",
		Phone:     "12345678",
		Country:   "Norway",
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	// Delete the customer
	err = svc.Delete(ctx, created.ID)
	assert.NoError(t, err)

	// Verify it's deleted
	_, err = svc.GetByID(ctx, created.ID)
	assert.Error(t, err)
}
