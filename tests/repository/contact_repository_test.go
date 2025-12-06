package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupContactTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createTestContact(t *testing.T, db *gorm.DB, firstName, lastName, email string) *domain.Contact {
	contact := &domain.Contact{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     "12345678",
		Title:     "Manager",
		IsActive:  true,
	}
	err := db.Create(contact).Error
	require.NoError(t, err)
	return contact
}

func createTestContactWithCustomer(t *testing.T, db *gorm.DB, firstName, lastName, email string, customerID uuid.UUID) *domain.Contact {
	contact := &domain.Contact{
		FirstName:         firstName,
		LastName:          lastName,
		Email:             email,
		Phone:             "12345678",
		Title:             "Manager",
		PrimaryCustomerID: &customerID,
		IsActive:          true,
	}
	err := db.Create(contact).Error
	require.NoError(t, err)
	return contact
}

// =============================================================================
// CRUD Tests
// =============================================================================

func TestContactRepository_Create(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	contact := &domain.Contact{
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		Phone:      "12345678",
		Mobile:     "98765432",
		Title:      "CEO",
		Department: "Executive",
		IsActive:   true,
	}

	err := repo.Create(context.Background(), contact)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, contact.ID)

	// Verify the contact was created
	found, err := repo.GetByID(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Equal(t, "John", found.FirstName)
	assert.Equal(t, "Doe", found.LastName)
	assert.Equal(t, "john.doe@example.com", found.Email)
}

func TestContactRepository_GetByID(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	contact := createTestContact(t, db, "Jane", "Smith", "jane.smith@example.com")

	found, err := repo.GetByID(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Equal(t, contact.ID, found.ID)
	assert.Equal(t, "Jane", found.FirstName)
	assert.Equal(t, "Smith", found.LastName)
}

func TestContactRepository_GetByID_NotFound(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestContactRepository_Update(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	contact := createTestContact(t, db, "Bob", "Johnson", "bob@example.com")

	// Update the contact
	contact.FirstName = "Robert"
	contact.Title = "Director"
	err := repo.Update(context.Background(), contact)
	assert.NoError(t, err)

	// Verify the update
	found, err := repo.GetByID(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Robert", found.FirstName)
	assert.Equal(t, "Director", found.Title)
}

func TestContactRepository_Delete(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	contact := createTestContact(t, db, "Delete", "Me", "delete.me@example.com")

	// Soft delete
	err := repo.Delete(context.Background(), contact.ID)
	assert.NoError(t, err)

	// Contact should still exist but be inactive
	found, err := repo.GetByID(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.False(t, found.IsActive)
}

func TestContactRepository_HardDelete(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	contact := createTestContact(t, db, "Hard", "Delete", "hard.delete@example.com")

	err := repo.HardDelete(context.Background(), contact.ID)
	assert.NoError(t, err)

	// Contact should not exist
	_, err = repo.GetByID(context.Background(), contact.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// =============================================================================
// List and Filter Tests
// =============================================================================

func TestContactRepository_List(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	// Create test contacts
	createTestContact(t, db, "Alice", "Anderson", "alice@example.com")
	createTestContact(t, db, "Bob", "Brown", "bob@example.com")
	createTestContact(t, db, "Charlie", "Clark", "charlie@example.com")

	t.Run("list all active contacts", func(t *testing.T) {
		contacts, total, err := repo.List(context.Background(), 1, 10, nil, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, contacts, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		contacts, total, err := repo.List(context.Background(), 1, 2, nil, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, contacts, 2)

		contacts, total, err = repo.List(context.Background(), 2, 2, nil, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, contacts, 1)
	})
}

func TestContactRepository_ListWithFilters(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Filter Test Customer")

	// Use unique identifiers to avoid conflicts with other test data
	uniquePrefix := uuid.New().String()[:8]

	// Create contacts with various attributes
	contact1 := &domain.Contact{
		FirstName:         "John" + uniquePrefix,
		LastName:          "Developer",
		Email:             uniquePrefix + ".john.dev@example.com",
		Title:             "Senior Developer",
		Department:        "FilterTestEngineering",
		PrimaryCustomerID: &customer.ID,
		IsActive:          true,
	}
	require.NoError(t, db.Create(contact1).Error)

	contact2 := &domain.Contact{
		FirstName:  "Jane" + uniquePrefix,
		LastName:   "Manager",
		Email:      uniquePrefix + ".jane.manager@example.com",
		Title:      "Project Manager",
		Department: "FilterTestOperations",
		IsActive:   true,
	}
	require.NoError(t, db.Create(contact2).Error)

	contact3 := &domain.Contact{
		FirstName:  "Inactive" + uniquePrefix,
		LastName:   "Person",
		Email:      uniquePrefix + ".inactive@example.com",
		Title:      "Former Employee",
		Department: "FilterTestEngineering",
		IsActive:   true, // Create first, then update to inactive
	}
	require.NoError(t, db.Create(contact3).Error)
	// Must update after creation to set is_active=false (GORM default handling)
	require.NoError(t, db.Model(contact3).Update("is_active", false).Error)

	t.Run("filter by search query - first name", func(t *testing.T) {
		search := "John" + uniquePrefix
		filters := &repository.ContactFilters{SearchQuery: &search}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, contacts, 1)
		assert.Equal(t, "John"+uniquePrefix, contacts[0].FirstName)
	})

	t.Run("filter by search query - last name", func(t *testing.T) {
		// Search for unique prefix + manager
		search := "Jane" + uniquePrefix
		filters := &repository.ContactFilters{SearchQuery: &search}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, contacts, 1)
		assert.Equal(t, "Jane"+uniquePrefix, contacts[0].FirstName)
	})

	t.Run("filter by search query - email", func(t *testing.T) {
		search := uniquePrefix + ".john.dev"
		filters := &repository.ContactFilters{SearchQuery: &search}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, contacts, 1)
		assert.Equal(t, "John"+uniquePrefix, contacts[0].FirstName)
	})

	t.Run("filter by title", func(t *testing.T) {
		// Also filter by unique prefix to isolate
		title := "Senior Developer"
		search := uniquePrefix
		filters := &repository.ContactFilters{Title: &title, SearchQuery: &search}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, contacts, 1)
		assert.Equal(t, "John"+uniquePrefix, contacts[0].FirstName)
	})

	t.Run("filter by department", func(t *testing.T) {
		dept := "FilterTestEngineering"
		filters := &repository.ContactFilters{Department: &dept}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total) // Only active one
		require.Len(t, contacts, 1)
		assert.Equal(t, "John"+uniquePrefix, contacts[0].FirstName)
	})

	t.Run("filter by primary customer", func(t *testing.T) {
		filters := &repository.ContactFilters{PrimaryCustomerID: &customer.ID}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, contacts, 1)
		assert.Equal(t, "John"+uniquePrefix, contacts[0].FirstName)
	})

	t.Run("filter by inactive status", func(t *testing.T) {
		isActive := false
		search := uniquePrefix
		filters := &repository.ContactFilters{IsActive: &isActive, SearchQuery: &search}
		contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, contacts, 1)
		assert.Equal(t, "Inactive"+uniquePrefix, contacts[0].FirstName)
	})
}

func TestContactRepository_ListSorting(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	// Create contacts
	createTestContact(t, db, "Charlie", "Zoo", "charlie@example.com")
	createTestContact(t, db, "Alice", "Anderson", "alice@example.com")
	createTestContact(t, db, "Bob", "Baker", "bob@example.com")

	t.Run("sort by name ascending", func(t *testing.T) {
		contacts, _, err := repo.List(context.Background(), 1, 10, nil, repository.ContactSortByNameAsc)
		assert.NoError(t, err)
		assert.Equal(t, "Anderson", contacts[0].LastName)
		assert.Equal(t, "Baker", contacts[1].LastName)
		assert.Equal(t, "Zoo", contacts[2].LastName)
	})

	t.Run("sort by name descending", func(t *testing.T) {
		contacts, _, err := repo.List(context.Background(), 1, 10, nil, repository.ContactSortByNameDesc)
		assert.NoError(t, err)
		assert.Equal(t, "Zoo", contacts[0].LastName)
		assert.Equal(t, "Baker", contacts[1].LastName)
		assert.Equal(t, "Anderson", contacts[2].LastName)
	})

	t.Run("sort by email ascending", func(t *testing.T) {
		contacts, _, err := repo.List(context.Background(), 1, 10, nil, repository.ContactSortByEmailAsc)
		assert.NoError(t, err)
		assert.Equal(t, "alice@example.com", contacts[0].Email)
		assert.Equal(t, "bob@example.com", contacts[1].Email)
		assert.Equal(t, "charlie@example.com", contacts[2].Email)
	})
}

// =============================================================================
// GetByEmail Test
// =============================================================================

func TestContactRepository_GetByEmail(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	contact := createTestContact(t, db, "Email", "Test", "email.test@EXAMPLE.COM")

	t.Run("find by exact email", func(t *testing.T) {
		found, err := repo.GetByEmail(context.Background(), "email.test@EXAMPLE.COM")
		assert.NoError(t, err)
		assert.Equal(t, contact.ID, found.ID)
	})

	t.Run("find by email case insensitive", func(t *testing.T) {
		found, err := repo.GetByEmail(context.Background(), "EMAIL.TEST@example.com")
		assert.NoError(t, err)
		assert.Equal(t, contact.ID, found.ID)
	})

	t.Run("not found for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(context.Background(), "nonexistent@example.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

// =============================================================================
// Search Tests
// =============================================================================

func TestContactRepository_Search(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	createTestContact(t, db, "John", "Doe", "john.doe@example.com")
	createTestContact(t, db, "Jane", "Doe", "jane.doe@example.com")
	createTestContact(t, db, "Bob", "Smith", "bob.smith@example.com")

	t.Run("search by last name", func(t *testing.T) {
		contacts, err := repo.Search(context.Background(), "doe", 10)
		assert.NoError(t, err)
		assert.Len(t, contacts, 2)
	})

	t.Run("search by first name", func(t *testing.T) {
		contacts, err := repo.Search(context.Background(), "john", 10)
		assert.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, "John", contacts[0].FirstName)
	})

	t.Run("search by full name", func(t *testing.T) {
		contacts, err := repo.Search(context.Background(), "jane doe", 10)
		assert.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, "Jane", contacts[0].FirstName)
	})

	t.Run("search by email", func(t *testing.T) {
		contacts, err := repo.Search(context.Background(), "bob.smith", 10)
		assert.NoError(t, err)
		assert.Len(t, contacts, 1)
		assert.Equal(t, "Bob", contacts[0].FirstName)
	})

	t.Run("search with limit", func(t *testing.T) {
		contacts, err := repo.Search(context.Background(), "doe", 1)
		assert.NoError(t, err)
		assert.Len(t, contacts, 1)
	})
}

// =============================================================================
// Relationship Tests
// =============================================================================

func TestContactRepository_Relationships_AddAndGet(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Relationship Test Customer")

	contact := createTestContact(t, db, "Rel", "Test", "rel.test@example.com")

	// Add a relationship
	rel := &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		Role:       "Primary Contact",
		IsPrimary:  true,
	}

	err := repo.AddRelationship(context.Background(), rel)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, rel.ID)

	// Get relationships for contact
	relationships, err := repo.GetContactRelationships(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Len(t, relationships, 1)
	assert.Equal(t, domain.ContactEntityCustomer, relationships[0].EntityType)
	assert.Equal(t, customer.ID, relationships[0].EntityID)
	assert.Equal(t, "Primary Contact", relationships[0].Role)
	assert.True(t, relationships[0].IsPrimary)
}

func TestContactRepository_Relationships_GetByID(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "GetByID Test Customer")

	contact := createTestContact(t, db, "RelID", "Test", "relid.test@example.com")

	rel := &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		Role:       "Decision Maker",
		IsPrimary:  false,
	}
	require.NoError(t, repo.AddRelationship(context.Background(), rel))

	// Get by ID
	found, err := repo.GetRelationshipByID(context.Background(), rel.ID)
	assert.NoError(t, err)
	assert.Equal(t, rel.ID, found.ID)
	assert.Equal(t, contact.ID, found.ContactID)
	assert.NotNil(t, found.Contact)
	assert.Equal(t, "RelID", found.Contact.FirstName)
}

func TestContactRepository_Relationships_Remove(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Remove Test Customer")

	contact := createTestContact(t, db, "Remove", "Test", "remove.test@example.com")

	rel := &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
	}
	require.NoError(t, repo.AddRelationship(context.Background(), rel))

	t.Run("remove by contact and entity", func(t *testing.T) {
		err := repo.RemoveRelationship(context.Background(), contact.ID, domain.ContactEntityCustomer, customer.ID)
		assert.NoError(t, err)

		relationships, err := repo.GetContactRelationships(context.Background(), contact.ID)
		assert.NoError(t, err)
		assert.Len(t, relationships, 0)
	})
}

func TestContactRepository_Relationships_RemoveByID(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "RemoveByID Test Customer")

	contact := createTestContact(t, db, "RemoveByID", "Test", "removebyid.test@example.com")

	rel := &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
	}
	require.NoError(t, repo.AddRelationship(context.Background(), rel))

	err := repo.RemoveRelationshipByID(context.Background(), rel.ID)
	assert.NoError(t, err)

	// Verify removal
	_, err = repo.GetRelationshipByID(context.Background(), rel.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestContactRepository_Relationships_RemoveByID_NotFound(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	err := repo.RemoveRelationshipByID(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestContactRepository_Relationships_MultiplePerContact(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	customer1 := testutil.CreateTestCustomer(t, db, "Customer One")
	customer2 := testutil.CreateTestCustomer(t, db, "Customer Two")

	contact := createTestContact(t, db, "Multi", "Rel", "multi.rel@example.com")

	// Add multiple relationships
	rel1 := &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer1.ID,
		Role:       "Primary Contact",
		IsPrimary:  true,
	}
	rel2 := &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer2.ID,
		Role:       "Secondary Contact",
		IsPrimary:  false,
	}

	require.NoError(t, repo.AddRelationship(context.Background(), rel1))
	require.NoError(t, repo.AddRelationship(context.Background(), rel2))

	relationships, err := repo.GetContactRelationships(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Len(t, relationships, 2)

	// Primary should be first due to ordering
	assert.True(t, relationships[0].IsPrimary)
	assert.False(t, relationships[1].IsPrimary)
}

// =============================================================================
// GetContactsForEntity Tests
// =============================================================================

func TestContactRepository_GetContactsForEntity(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	uniquePrefix := uuid.New().String()[:8]
	customer := testutil.CreateTestCustomer(t, db, "Entity Contacts Customer "+uniquePrefix)

	// Create contacts and link them to the customer
	contact1 := createTestContact(t, db, "Alpha"+uniquePrefix, "Contact", uniquePrefix+".alpha@example.com")
	contact2 := createTestContact(t, db, "Beta"+uniquePrefix, "Contact", uniquePrefix+".beta@example.com")
	contact3 := createTestContact(t, db, "Unrelated"+uniquePrefix, "Contact", uniquePrefix+".unrelated@example.com")

	// Create an inactive contact - must explicitly set is_active to false after creation
	inactiveContact := &domain.Contact{
		FirstName: "Inactive" + uniquePrefix,
		LastName:  "Contact",
		Email:     uniquePrefix + ".inactive.contact@example.com",
		IsActive:  true, // GORM will create with this
	}
	require.NoError(t, db.Create(inactiveContact).Error)
	// Now update to inactive
	require.NoError(t, db.Model(inactiveContact).Update("is_active", false).Error)

	// Link contacts to customer
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact1.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  true,
	}))
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact2.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  false,
	}))
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  inactiveContact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  false,
	}))

	contacts, err := repo.GetContactsForEntity(context.Background(), domain.ContactEntityCustomer, customer.ID)
	assert.NoError(t, err)
	assert.Len(t, contacts, 2) // Only active contacts

	// Primary contact should be first
	assert.Equal(t, "Alpha"+uniquePrefix, contacts[0].FirstName)

	// Unrelated contact should not be included
	for _, c := range contacts {
		assert.NotEqual(t, contact3.ID, c.ID)
	}
}

func TestContactRepository_GetPrimaryContactForEntity(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Primary Contact Customer")

	primary := createTestContact(t, db, "Primary", "Person", "primary@example.com")
	secondary := createTestContact(t, db, "Secondary", "Person", "secondary@example.com")

	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  primary.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  true,
	}))
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  secondary.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  false,
	}))

	found, err := repo.GetPrimaryContactForEntity(context.Background(), domain.ContactEntityCustomer, customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, primary.ID, found.ID)
	assert.Equal(t, "Primary", found.FirstName)
}

func TestContactRepository_GetPrimaryContactForEntity_NotFound(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "No Primary Customer")

	contact := createTestContact(t, db, "Not", "Primary", "not.primary@example.com")

	// Add a non-primary relationship
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  false,
	}))

	_, err := repo.GetPrimaryContactForEntity(context.Background(), domain.ContactEntityCustomer, customer.ID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// =============================================================================
// SetPrimaryContact Tests
// =============================================================================

func TestContactRepository_SetPrimaryContact(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Set Primary Customer")

	contact1 := createTestContact(t, db, "First", "Primary", "first@example.com")
	contact2 := createTestContact(t, db, "Second", "Primary", "second@example.com")

	// Add contact1 as primary
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact1.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  true,
	}))

	// Add contact2 as non-primary
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact2.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		IsPrimary:  false,
	}))

	// Set contact2 as primary (should unset contact1)
	err := repo.SetPrimaryContact(context.Background(), domain.ContactEntityCustomer, customer.ID, contact2.ID)
	assert.NoError(t, err)

	// Verify contact2 is now primary
	primary, err := repo.GetPrimaryContactForEntity(context.Background(), domain.ContactEntityCustomer, customer.ID)
	assert.NoError(t, err)
	assert.Equal(t, contact2.ID, primary.ID)

	// Verify contact1 is no longer primary
	rels, err := repo.GetContactRelationships(context.Background(), contact1.ID)
	assert.NoError(t, err)
	assert.Len(t, rels, 1)
	assert.False(t, rels[0].IsPrimary)
}

func TestContactRepository_SetPrimaryContact_CreatesRelationship(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "New Primary Customer")

	contact := createTestContact(t, db, "New", "Primary", "new.primary@example.com")

	// Set as primary without existing relationship
	err := repo.SetPrimaryContact(context.Background(), domain.ContactEntityCustomer, customer.ID, contact.ID)
	assert.NoError(t, err)

	// Verify relationship was created
	rels, err := repo.GetContactRelationships(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Len(t, rels, 1)
	assert.True(t, rels[0].IsPrimary)
	assert.Equal(t, domain.ContactEntityCustomer, rels[0].EntityType)
	assert.Equal(t, customer.ID, rels[0].EntityID)
}

// =============================================================================
// GetRelationshipsForEntity Tests
// =============================================================================

func TestContactRepository_GetRelationshipsForEntity(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Entity Rels Customer")

	contact1 := createTestContact(t, db, "Ent", "One", "ent.one@example.com")
	contact2 := createTestContact(t, db, "Ent", "Two", "ent.two@example.com")

	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact1.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		Role:       "CEO",
		IsPrimary:  true,
	}))
	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact2.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		Role:       "CFO",
		IsPrimary:  false,
	}))

	rels, err := repo.GetRelationshipsForEntity(context.Background(), domain.ContactEntityCustomer, customer.ID)
	assert.NoError(t, err)
	assert.Len(t, rels, 2)

	// Primary should be first
	assert.True(t, rels[0].IsPrimary)
	assert.Equal(t, "CEO", rels[0].Role)

	// Contacts should be preloaded
	assert.NotNil(t, rels[0].Contact)
	assert.NotNil(t, rels[1].Contact)
}

// =============================================================================
// ListByPrimaryCustomer Tests
// =============================================================================

func TestContactRepository_ListByPrimaryCustomer(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	customer1 := testutil.CreateTestCustomer(t, db, "Primary Cust One")
	customer2 := testutil.CreateTestCustomer(t, db, "Primary Cust Two")

	createTestContactWithCustomer(t, db, "Cust1", "Contact1", "c1c1@example.com", customer1.ID)
	createTestContactWithCustomer(t, db, "Cust1", "Contact2", "c1c2@example.com", customer1.ID)
	createTestContactWithCustomer(t, db, "Cust2", "Contact1", "c2c1@example.com", customer2.ID)

	contacts, err := repo.ListByPrimaryCustomer(context.Background(), customer1.ID)
	assert.NoError(t, err)
	assert.Len(t, contacts, 2)

	for _, c := range contacts {
		assert.Equal(t, customer1.ID, *c.PrimaryCustomerID)
	}
}

// =============================================================================
// Transaction Tests
// =============================================================================

func TestContactRepository_WithTransaction(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	t.Run("successful transaction", func(t *testing.T) {
		var contactID uuid.UUID
		err := repo.WithTransaction(context.Background(), func(tx *gorm.DB) error {
			contact := &domain.Contact{
				FirstName: "Transaction",
				LastName:  "Test",
				Email:     "tx.test@example.com",
				IsActive:  true,
			}
			if err := tx.Create(contact).Error; err != nil {
				return err
			}
			contactID = contact.ID
			return nil
		})
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), contactID)
		assert.NoError(t, err)
		assert.Equal(t, "Transaction", found.FirstName)
	})

	t.Run("rollback on error", func(t *testing.T) {
		err := repo.WithTransaction(context.Background(), func(tx *gorm.DB) error {
			contact := &domain.Contact{
				FirstName: "Rollback",
				LastName:  "Test",
				Email:     "rollback.test@example.com",
				IsActive:  true,
			}
			if err := tx.Create(contact).Error; err != nil {
				return err
			}
			// Simulate an error to trigger rollback
			return gorm.ErrInvalidTransaction
		})
		assert.Error(t, err)

		// Contact should not exist
		_, err = repo.GetByEmail(context.Background(), "rollback.test@example.com")
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestContactRepository_EmptyFilters(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)

	createTestContact(t, db, "Empty", "Filter", "empty.filter@example.com")

	// Empty filter struct should work the same as nil
	filters := &repository.ContactFilters{}
	contacts, total, err := repo.List(context.Background(), 1, 10, filters, repository.ContactSortByNameAsc)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, contacts, 1)
}

func TestContactRepository_PreloadedRelationships(t *testing.T) {
	db := setupContactTestDB(t)
	repo := repository.NewContactRepository(db)
	customer := testutil.CreateTestCustomer(t, db, "Preload Customer")

	contact := createTestContact(t, db, "Preload", "Test", "preload@example.com")

	require.NoError(t, repo.AddRelationship(context.Background(), &domain.ContactRelationship{
		ContactID:  contact.ID,
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
		Role:       "Tester",
	}))

	// GetByID should preload relationships
	found, err := repo.GetByID(context.Background(), contact.ID)
	assert.NoError(t, err)
	assert.Len(t, found.Relationships, 1)
	assert.Equal(t, "Tester", found.Relationships[0].Role)

	// List should also preload relationships
	contacts, _, err := repo.List(context.Background(), 1, 10, nil, repository.ContactSortByNameAsc)
	assert.NoError(t, err)
	for _, c := range contacts {
		if c.ID == contact.ID {
			assert.Len(t, c.Relationships, 1)
		}
	}
}
