package repository_test

import (
	"context"
	"testing"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&domain.Customer{},
		&domain.Contact{},
		&domain.Project{},
		&domain.Offer{},
	)
	require.NoError(t, err)

	return db
}

func TestCustomerRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCustomerRepository(db)

	customer := &domain.Customer{
		Name:          "Test Company",
		OrgNumber:     "1234567890",
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

	err := repo.Create(context.Background(), customer)
	assert.NoError(t, err)
	assert.NotEqual(t, "", customer.ID.String())
}

func TestCustomerRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCustomerRepository(db)

	customer := &domain.Customer{
		Name:          "Test Company",
		OrgNumber:     "1234567890",
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

	err := repo.Create(context.Background(), customer)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, customer.Name, found.Name)
	assert.Equal(t, customer.OrgNumber, found.OrgNumber)
	assert.Equal(t, customer.Email, found.Email)
	assert.Equal(t, customer.Phone, found.Phone)
	assert.Equal(t, customer.Address, found.Address)
	assert.Equal(t, customer.City, found.City)
	assert.Equal(t, customer.PostalCode, found.PostalCode)
	assert.Equal(t, customer.Country, found.Country)
	assert.Equal(t, customer.ContactPerson, found.ContactPerson)
	assert.Equal(t, customer.ContactEmail, found.ContactEmail)
	assert.Equal(t, customer.ContactPhone, found.ContactPhone)
}

func TestCustomerRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCustomerRepository(db)

	// Create test customers
	customers := []*domain.Customer{
		{Name: "Tech Corp", OrgNumber: "1234567890", Email: "test@example.com", Phone: "1234567890", Address: "123 Main St", City: "Anytown", PostalCode: "12345", Country: "Norway", ContactPerson: "John Doe", ContactEmail: "john.doe@example.com", ContactPhone: "+1234567890"},
		{Name: "Finance Inc", OrgNumber: "1234567890", Email: "test@example.com", Phone: "1234567890", Address: "123 Main St", City: "Anytown", PostalCode: "12345", Country: "Norway", ContactPerson: "John Doe", ContactEmail: "john.doe@example.com", ContactPhone: "+1234567890"},
		{Name: "Tech Solutions", OrgNumber: "1234567890", Email: "test@example.com", Phone: "1234567890", Address: "123 Main St", City: "Anytown", PostalCode: "12345", Country: "Norway", ContactPerson: "John Doe", ContactEmail: "john.doe@example.com", ContactPhone: "+1234567890"},
	}

	for _, c := range customers {
		err := repo.Create(context.Background(), c)
		require.NoError(t, err)
	}

	// Test listing all
	result, total, err := repo.List(context.Background(), 1, 10, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, result, 3)

	// Test search
	result, total, err = repo.List(context.Background(), 1, 10, "Tech")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)

	// Test pagination
	result, total, err = repo.List(context.Background(), 1, 2, "")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, result, 2)
}

func TestCustomerRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCustomerRepository(db)

	customer := &domain.Customer{
		Name:          "Original Name",
		OrgNumber:     "1234567890",
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

	err := repo.Create(context.Background(), customer)
	require.NoError(t, err)

	customer.Name = "Updated Name"
	customer.OrgNumber = "1234567890"
	customer.Email = "test@example.com"
	customer.Phone = "1234567890"
	customer.Address = "123 Main St"
	customer.City = "Anytown"
	customer.PostalCode = "12345"
	customer.Country = "Norway"
	customer.ContactPerson = "John Doe"
	customer.ContactEmail = "john.doe@example.com"
	customer.ContactPhone = "+1234567890"

	err = repo.Update(context.Background(), customer)
	assert.NoError(t, err)

	found, err := repo.GetByID(context.Background(), customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", found.Name)
	assert.Equal(t, "1234567890", found.OrgNumber)
	assert.Equal(t, "test@example.com", found.Email)
	assert.Equal(t, "1234567890", found.Phone)
	assert.Equal(t, "123 Main St", found.Address)
	assert.Equal(t, "Anytown", found.City)
	assert.Equal(t, "12345", found.PostalCode)
	assert.Equal(t, "Norway", found.Country)
	assert.Equal(t, "John Doe", found.ContactPerson)
	assert.Equal(t, "john.doe@example.com", found.ContactEmail)
	assert.Equal(t, "+1234567890", found.ContactPhone)
}

func TestCustomerRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewCustomerRepository(db)

	customer := &domain.Customer{
		Name: "Test Company",
	}

	err := repo.Create(context.Background(), customer)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), customer.ID)
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), customer.ID)
	assert.Error(t, err)
}
