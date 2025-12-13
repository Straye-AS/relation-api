package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCustomerTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupCleanTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func TestCustomerRepository_Create(t *testing.T) {
	db := setupCustomerTestDB(t)
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
	db := setupCustomerTestDB(t)
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

func TestCustomerRepository_GetByOrgNumber(t *testing.T) {
	db := setupCustomerTestDB(t)
	repo := repository.NewCustomerRepository(db)

	t.Run("found", func(t *testing.T) {
		customer := &domain.Customer{
			Name:      "Unique Org Company",
			OrgNumber: "9876543210",
			Email:     "test@example.com",
			Phone:     "12345678",
			Country:   "Norway",
		}
		err := repo.Create(context.Background(), customer)
		require.NoError(t, err)

		found, err := repo.GetByOrgNumber(context.Background(), "9876543210")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, customer.ID, found.ID)
		assert.Equal(t, "Unique Org Company", found.Name)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByOrgNumber(context.Background(), "0000000000")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestCustomerRepository_List(t *testing.T) {
	db := setupCustomerTestDB(t)
	repo := repository.NewCustomerRepository(db)

	// Create test customers with unique OrgNumbers
	customers := []*domain.Customer{
		{Name: "Tech Corp", OrgNumber: "123456001", Email: "test@example.com", Phone: "1234567890", Address: "123 Main St", City: "Anytown", PostalCode: "12345", Country: "Norway", ContactPerson: "John Doe", ContactEmail: "john.doe@example.com", ContactPhone: "+1234567890"},
		{Name: "Finance Inc", OrgNumber: "123456002", Email: "test@example.com", Phone: "1234567890", Address: "123 Main St", City: "Anytown", PostalCode: "12345", Country: "Norway", ContactPerson: "John Doe", ContactEmail: "john.doe@example.com", ContactPhone: "+1234567890"},
		{Name: "Tech Solutions", OrgNumber: "123456003", Email: "test@example.com", Phone: "1234567890", Address: "123 Main St", City: "Anytown", PostalCode: "12345", Country: "Norway", ContactPerson: "John Doe", ContactEmail: "john.doe@example.com", ContactPhone: "+1234567890"},
	}

	for _, c := range customers {
		err := repo.Create(context.Background(), c)
		require.NoError(t, err)
	}

	t.Run("list all", func(t *testing.T) {
		result, total, err := repo.List(context.Background(), 1, 10, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		result, total, err := repo.List(context.Background(), 1, 2, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 2)

		result, total, err = repo.List(context.Background(), 2, 2, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 1)
	})
}

func TestCustomerRepository_ListWithFilters(t *testing.T) {
	db := setupCustomerTestDB(t)
	repo := repository.NewCustomerRepository(db)

	// Create customers with different attributes
	customers := []*domain.Customer{
		{Name: "Oslo Company", OrgNumber: "111111001", Email: "oslo@example.com", Phone: "12345678", City: "Oslo", Country: "Norway"},
		{Name: "Bergen AS", OrgNumber: "111111002", Email: "bergen@example.com", Phone: "12345678", City: "Bergen", Country: "Norway"},
		{Name: "Stockholm AB", OrgNumber: "111111003", Email: "stockholm@example.com", Phone: "12345678", City: "Stockholm", Country: "Sweden"},
		{Name: "Oslo Tech", OrgNumber: "111111004", Email: "oslotech@example.com", Phone: "12345678", City: "Oslo", Country: "Norway"},
	}

	for _, c := range customers {
		err := repo.Create(context.Background(), c)
		require.NoError(t, err)
	}

	t.Run("filter by city", func(t *testing.T) {
		filters := &repository.CustomerFilters{City: "Oslo"}
		result, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.CustomerSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
		for _, c := range result {
			assert.Equal(t, "Oslo", c.City)
		}
	})

	t.Run("filter by country", func(t *testing.T) {
		filters := &repository.CustomerFilters{Country: "Sweden"}
		result, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.CustomerSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "Stockholm AB", result[0].Name)
	})

	t.Run("filter by search - name match", func(t *testing.T) {
		filters := &repository.CustomerFilters{Search: "Tech"}
		result, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.CustomerSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "Oslo Tech", result[0].Name)
	})

	t.Run("filter by search - org number match", func(t *testing.T) {
		filters := &repository.CustomerFilters{Search: "111111002"}
		result, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.CustomerSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
		assert.Equal(t, "Bergen AS", result[0].Name)
	})

	t.Run("filter by city and country combined", func(t *testing.T) {
		filters := &repository.CustomerFilters{City: "Oslo", Country: "Norway"}
		result, total, err := repo.ListWithFilters(context.Background(), 1, 10, filters, repository.CustomerSortByCreatedDesc)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})
}

func TestCustomerRepository_ListWithSorting(t *testing.T) {
	db := setupCustomerTestDB(t)
	repo := repository.NewCustomerRepository(db)

	// Create customers with predictable ordering
	customers := []*domain.Customer{
		{Name: "Alpha Corp", OrgNumber: "222222001", Email: "alpha@example.com", Phone: "12345678", City: "Bergen", Country: "Norway"},
		{Name: "Beta Inc", OrgNumber: "222222002", Email: "beta@example.com", Phone: "12345678", City: "Oslo", Country: "Norway"},
		{Name: "Gamma AS", OrgNumber: "222222003", Email: "gamma@example.com", Phone: "12345678", City: "Trondheim", Country: "Norway"},
	}

	for _, c := range customers {
		err := repo.Create(context.Background(), c)
		require.NoError(t, err)
		// Add small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	t.Run("sort by name ascending", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, repository.CustomerSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, "Alpha Corp", result[0].Name)
		assert.Equal(t, "Beta Inc", result[1].Name)
		assert.Equal(t, "Gamma AS", result[2].Name)
	})

	t.Run("sort by name descending", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, repository.CustomerSortByNameDesc)
		assert.NoError(t, err)
		assert.Equal(t, "Gamma AS", result[0].Name)
		assert.Equal(t, "Beta Inc", result[1].Name)
		assert.Equal(t, "Alpha Corp", result[2].Name)
	})

	t.Run("sort by created ascending", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, repository.CustomerSortByCreatedAsc)
		assert.NoError(t, err)
		// First created should be first
		assert.Equal(t, "Alpha Corp", result[0].Name)
	})

	t.Run("sort by created descending", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, repository.CustomerSortByCreatedDesc)
		assert.NoError(t, err)
		// Last created should be first
		assert.Equal(t, "Gamma AS", result[0].Name)
	})

	t.Run("sort by city ascending", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, repository.CustomerSortByCityAsc)
		assert.NoError(t, err)
		assert.Equal(t, "Bergen", result[0].City)
		assert.Equal(t, "Oslo", result[1].City)
		assert.Equal(t, "Trondheim", result[2].City)
	})

	t.Run("sort by city descending", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, repository.CustomerSortByCityDesc)
		assert.NoError(t, err)
		assert.Equal(t, "Trondheim", result[0].City)
		assert.Equal(t, "Oslo", result[1].City)
		assert.Equal(t, "Bergen", result[2].City)
	})

	t.Run("default sorting when empty", func(t *testing.T) {
		result, _, err := repo.ListWithFilters(context.Background(), 1, 10, nil, "")
		assert.NoError(t, err)
		// Default is created_at DESC, so newest first
		assert.Equal(t, "Gamma AS", result[0].Name)
	})
}

func TestCustomerRepository_GetCustomerStats(t *testing.T) {
	db := setupCustomerTestDB(t)
	customerRepo := repository.NewCustomerRepository(db)

	// Create a test customer
	customer := testutil.CreateTestCustomer(t, db, "Stats Test Customer")

	t.Run("customer with no deals or projects", func(t *testing.T) {
		stats, err := customerRepo.GetCustomerStats(context.Background(), customer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats.ActiveDeals)
		assert.Equal(t, float64(0), stats.TotalValue)
		assert.Equal(t, 0, stats.ActiveProjects)
		assert.Equal(t, 0, stats.ActiveOffers)
		assert.Equal(t, 0, stats.TotalContacts)
	})

	t.Run("customer with deals", func(t *testing.T) {
		// Create some deals for the customer
		deals := []*domain.Deal{
			{Title: "Active Deal 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLead, Value: 100000, OwnerID: "user-1", Currency: "NOK"},
			{Title: "Active Deal 2", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageQualified, Value: 200000, OwnerID: "user-1", Currency: "NOK"},
			{Title: "Won Deal", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageWon, Value: 500000, OwnerID: "user-1", Currency: "NOK"},
			{Title: "Lost Deal", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Stage: domain.DealStageLost, Value: 150000, OwnerID: "user-1", Currency: "NOK"},
		}
		for _, d := range deals {
			err := db.Create(d).Error
			require.NoError(t, err)
		}

		stats, err := customerRepo.GetCustomerStats(context.Background(), customer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		// ActiveDeals counts deals NOT in Won or Lost stages
		assert.Equal(t, 2, stats.ActiveDeals) // Lead and Qualified stages
	})

	t.Run("customer with projects", func(t *testing.T) {
		// Create projects for the customer
		startDate := time.Now()
		managerID := "mgr-1"
		projects := []*domain.Project{
			{Name: "Active Project 1", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Phase: domain.ProjectPhaseActive, ManagerID: &managerID, StartDate: startDate},
			{Name: "Tilbud Project", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Phase: domain.ProjectPhaseTilbud, ManagerID: &managerID, StartDate: startDate},
			{Name: "Completed Project", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Phase: domain.ProjectPhaseCompleted, ManagerID: &managerID, StartDate: startDate},
			{Name: "Working Project", CustomerID: customer.ID, CompanyID: domain.CompanyStalbygg, Phase: domain.ProjectPhaseWorking, ManagerID: &managerID, StartDate: startDate},
		}
		for _, p := range projects {
			err := db.Create(p).Error
			require.NoError(t, err)
		}

		stats, err := customerRepo.GetCustomerStats(context.Background(), customer.ID)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		// Active projects = Active + Tilbud + Working (Completed not counted as active per the repo query)
		assert.Equal(t, 3, stats.ActiveProjects)
	})

	t.Run("non-existent customer", func(t *testing.T) {
		nonExistentID := uuid.New()
		stats, err := customerRepo.GetCustomerStats(context.Background(), nonExistentID)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, 0, stats.ActiveDeals)
		assert.Equal(t, float64(0), stats.TotalValue)
		assert.Equal(t, 0, stats.ActiveProjects)
		assert.Equal(t, 0, stats.ActiveOffers)
		assert.Equal(t, 0, stats.TotalContacts)
	})
}

func TestCustomerRepository_Update(t *testing.T) {
	db := setupCustomerTestDB(t)
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
	db := setupCustomerTestDB(t)
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
