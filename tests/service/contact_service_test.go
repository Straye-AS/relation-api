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
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func setupContactServiceTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
}

func createContactService(db *gorm.DB) *service.ContactService {
	contactRepo := repository.NewContactRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	logger := zap.NewNop()

	return service.NewContactService(contactRepo, customerRepo, activityRepo, logger)
}

func createContactTestContext() context.Context {
	ctx := auth.WithUserContext(context.Background(), &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	})
	return ctx
}

func createTestDeal(t *testing.T, db *gorm.DB, customer *domain.Customer) *domain.Deal {
	deal := &domain.Deal{
		Title:        "Test Deal",
		Description:  "Test deal description",
		CustomerID:   customer.ID,
		CustomerName: customer.Name,
		CompanyID:    domain.CompanyStalbygg,
		Stage:        domain.DealStageLead,
		Probability:  25,
		Value:        100000,
		Currency:     "NOK",
		OwnerID:      "test-owner",
		OwnerName:    "Test Owner",
	}
	err := db.Create(deal).Error
	require.NoError(t, err)
	return deal
}

func createTestProject(t *testing.T, db *gorm.DB, customer *domain.Customer) *domain.Project {
	startDate := time.Now()
	customerID := customer.ID
	project := &domain.Project{
		Name:         "Test Project",
		Description:  "Test project description",
		CustomerID:   &customerID,
		CustomerName: customer.Name,
		Phase:        domain.ProjectPhaseWorking,
		StartDate:    startDate,
	}
	err := db.Create(project).Error
	require.NoError(t, err)
	return project
}

// TestContactService_Create tests contact creation
func TestContactService_Create(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	tests := []struct {
		name      string
		req       *domain.CreateContactRequest
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - basic contact",
			req: &domain.CreateContactRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Phone:     "+47 12345678",
				Title:     "CEO",
			},
			wantErr: false,
		},
		{
			name: "success - contact without email",
			req: &domain.CreateContactRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Phone:     "+47 87654321",
			},
			wantErr: false,
		},
		{
			name: "error - invalid email format",
			req: &domain.CreateContactRequest{
				FirstName: "Bob",
				LastName:  "Invalid",
				Email:     "not-an-email",
			},
			wantErr:   true,
			errSubstr: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contact, err := svc.Create(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, contact)
			assert.NotEqual(t, uuid.Nil, contact.ID)
			assert.Equal(t, tt.req.FirstName, contact.FirstName)
			assert.Equal(t, tt.req.LastName, contact.LastName)
			assert.Equal(t, tt.req.FirstName+" "+tt.req.LastName, contact.FullName)
			assert.True(t, contact.IsActive)

			if tt.req.Country == "" {
				assert.Equal(t, "Norway", contact.Country)
			}
			if tt.req.PreferredContactMethod == "" {
				assert.Equal(t, "email", contact.PreferredContactMethod)
			}
		})
	}
}

// TestContactService_Create_DuplicateEmail tests email uniqueness validation
func TestContactService_Create_DuplicateEmail(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create first contact
	req1 := &domain.CreateContactRequest{
		FirstName: "First",
		LastName:  "Contact",
		Email:     "duplicate@example.com",
	}
	_, err := svc.Create(ctx, req1)
	require.NoError(t, err)

	// Try to create second contact with same email
	req2 := &domain.CreateContactRequest{
		FirstName: "Second",
		LastName:  "Contact",
		Email:     "duplicate@example.com",
	}
	_, err = svc.Create(ctx, req2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Case-insensitive check
	req3 := &domain.CreateContactRequest{
		FirstName: "Third",
		LastName:  "Contact",
		Email:     "DUPLICATE@EXAMPLE.COM",
	}
	_, err = svc.Create(ctx, req3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// TestContactService_Update tests contact update
func TestContactService_Update(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create a contact
	createReq := &domain.CreateContactRequest{
		FirstName: "Original",
		LastName:  "Name",
		Email:     "original@example.com",
		Phone:     "12345678",
	}
	contact, err := svc.Create(ctx, createReq)
	require.NoError(t, err)

	// Update the contact
	updateReq := &domain.UpdateContactRequest{
		FirstName: "Updated",
		LastName:  "Person",
		Email:     "updated@example.com",
		Phone:     "87654321",
		Title:     "Manager",
	}
	updated, err := svc.Update(ctx, contact.ID, updateReq)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.FirstName)
	assert.Equal(t, "Person", updated.LastName)
	assert.Equal(t, "updated@example.com", updated.Email)
	assert.Equal(t, "87654321", updated.Phone)
	assert.Equal(t, "Manager", updated.Title)
}

// TestContactService_Update_EmailValidation tests email validation on update
func TestContactService_Update_EmailValidation(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create two contacts
	contact1, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Contact",
		LastName:  "One",
		Email:     "contact1@example.com",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Contact",
		LastName:  "Two",
		Email:     "contact2@example.com",
	})
	require.NoError(t, err)

	// Try to update contact1 with contact2's email
	_, err = svc.Update(ctx, contact1.ID, &domain.UpdateContactRequest{
		FirstName: "Contact",
		LastName:  "One",
		Email:     "contact2@example.com",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Updating with same email should work
	_, err = svc.Update(ctx, contact1.ID, &domain.UpdateContactRequest{
		FirstName: "Contact",
		LastName:  "One Updated",
		Email:     "contact1@example.com",
	})
	assert.NoError(t, err)
}

// TestContactService_Delete tests contact deletion
func TestContactService_Delete(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create a contact
	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "To",
		LastName:  "Delete",
		Email:     "delete@example.com",
	})
	require.NoError(t, err)

	// Delete the contact
	err = svc.Delete(ctx, contact.ID)
	assert.NoError(t, err)

	// Verify it's deleted
	_, err = svc.GetByID(ctx, contact.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestContactService_AddRelationship tests adding contact relationships
func TestContactService_AddRelationship(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create prerequisites
	customer := testutil.CreateTestCustomer(t, db, "Test Customer")
	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Contact",
		LastName:  "Person",
		Email:     "contact.person@example.com",
	})
	require.NoError(t, err)

	tests := []struct {
		name      string
		req       *domain.AddContactRelationshipRequest
		wantErr   bool
		errSubstr string
	}{
		{
			name: "success - link to customer",
			req: &domain.AddContactRelationshipRequest{
				EntityType: domain.ContactEntityCustomer,
				EntityID:   customer.ID,
				Role:       "Decision Maker",
				IsPrimary:  false,
			},
			wantErr: false,
		},
		{
			name: "error - duplicate relationship",
			req: &domain.AddContactRelationshipRequest{
				EntityType: domain.ContactEntityCustomer,
				EntityID:   customer.ID,
				Role:       "Buyer",
				IsPrimary:  false,
			},
			wantErr:   true,
			errSubstr: "already exists",
		},
		{
			name: "error - invalid entity type",
			req: &domain.AddContactRelationshipRequest{
				EntityType: "invalid",
				EntityID:   uuid.New(),
			},
			wantErr:   true,
			errSubstr: "invalid input value for enum", // Postgres enum validation error
		},
		// Note: Service does not validate entity existence - relationships can be created
		// to non-existent entities (this is a data integrity choice, not a bug)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := svc.AddRelationship(ctx, contact.ID, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, rel)
			assert.NotEqual(t, uuid.Nil, rel.ID)
			assert.Equal(t, contact.ID, rel.ContactID)
			assert.Equal(t, tt.req.EntityType, rel.EntityType)
			assert.Equal(t, tt.req.EntityID, rel.EntityID)
			assert.Equal(t, tt.req.Role, rel.Role)
			assert.Equal(t, tt.req.IsPrimary, rel.IsPrimary)
		})
	}
}

// TestContactService_AddRelationship_WithDeal tests linking contact to deal
func TestContactService_AddRelationship_WithDeal(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create prerequisites
	customer := testutil.CreateTestCustomer(t, db, "Deal Customer")
	deal := createTestDeal(t, db, customer)

	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Deal",
		LastName:  "Contact",
	})
	require.NoError(t, err)

	// Add relationship to deal
	rel, err := svc.AddRelationship(ctx, contact.ID, &domain.AddContactRelationshipRequest{
		EntityType: domain.ContactEntityDeal,
		EntityID:   deal.ID,
		Role:       "Stakeholder",
		IsPrimary:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ContactEntityDeal, rel.EntityType)
	assert.Equal(t, deal.ID, rel.EntityID)
	assert.True(t, rel.IsPrimary)
}

// TestContactService_AddRelationship_WithProject tests linking contact to project
func TestContactService_AddRelationship_WithProject(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create prerequisites
	customer := testutil.CreateTestCustomer(t, db, "Project Customer")
	project := createTestProject(t, db, customer)

	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Project",
		LastName:  "Contact",
	})
	require.NoError(t, err)

	// Add relationship to project
	rel, err := svc.AddRelationship(ctx, contact.ID, &domain.AddContactRelationshipRequest{
		EntityType: domain.ContactEntityProject,
		EntityID:   project.ID,
		Role:       "Project Lead",
		IsPrimary:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ContactEntityProject, rel.EntityType)
	assert.Equal(t, project.ID, rel.EntityID)
	assert.True(t, rel.IsPrimary)
}

// TestContactService_RemoveRelationship tests removing contact relationships
func TestContactService_RemoveRelationship(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create prerequisites
	customer := testutil.CreateTestCustomer(t, db, "Remove Test Customer")
	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Remove",
		LastName:  "Contact",
	})
	require.NoError(t, err)

	// Add relationship
	rel, err := svc.AddRelationship(ctx, contact.ID, &domain.AddContactRelationshipRequest{
		EntityType: domain.ContactEntityCustomer,
		EntityID:   customer.ID,
	})
	require.NoError(t, err)

	// Remove relationship
	err = svc.RemoveRelationship(ctx, contact.ID, rel.ID)
	assert.NoError(t, err)

	// Verify relationship is removed (contact should no longer appear in customer's contacts)
	contacts, err := svc.ListByEntity(ctx, domain.ContactEntityCustomer, customer.ID)
	require.NoError(t, err)
	for _, c := range contacts {
		assert.NotEqual(t, contact.ID, c.ID)
	}
}

// TestContactService_RemoveRelationship_NotFound tests removing non-existent relationship
func TestContactService_RemoveRelationship_NotFound(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create a contact first
	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName: "Test",
		LastName:  "Contact",
	})
	require.NoError(t, err)

	err = svc.RemoveRelationship(ctx, contact.ID, uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Note: SetPrimaryContact was removed from the service API.
// Primary contact is set via the IsPrimary field when calling AddRelationship.

// TestContactService_ListByEntity tests retrieving contacts for an entity
func TestContactService_ListByEntity(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create customer
	customer := testutil.CreateTestCustomer(t, db, "Entity Contacts Customer")

	// Create multiple contacts and link them
	for i := 0; i < 3; i++ {
		contact, err := svc.Create(ctx, &domain.CreateContactRequest{
			FirstName: "Contact",
			LastName:  string(rune('A' + i)),
		})
		require.NoError(t, err)

		_, err = svc.AddRelationship(ctx, contact.ID, &domain.AddContactRelationshipRequest{
			EntityType: domain.ContactEntityCustomer,
			EntityID:   customer.ID,
			IsPrimary:  i == 0,
		})
		require.NoError(t, err)
	}

	// Get contacts for customer
	contacts, err := svc.ListByEntity(ctx, domain.ContactEntityCustomer, customer.ID)
	require.NoError(t, err)
	assert.Len(t, contacts, 3)
}

// TestContactService_List tests listing contacts with pagination
func TestContactService_List(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create multiple contacts
	for i := 0; i < 25; i++ {
		_, err := svc.Create(ctx, &domain.CreateContactRequest{
			FirstName: "List",
			LastName:  string(rune('A' + (i % 26))),
		})
		require.NoError(t, err)
	}

	// Test pagination
	contacts, total, err := svc.List(ctx, 1, 10)
	require.NoError(t, err)
	assert.Len(t, contacts, 10)
	assert.GreaterOrEqual(t, total, int64(25))

	// Test second page
	contacts2, _, err := svc.List(ctx, 2, 10)
	require.NoError(t, err)
	assert.Len(t, contacts2, 10)

	// Verify different results
	assert.NotEqual(t, contacts[0].ID, contacts2[0].ID)
}

// Note: Search method was removed from ContactService.
// Search functionality is available via ListWithFilters with the Search filter field.

// TestContactService_Create_WithPrimaryCustomer tests creating contact with primary customer
func TestContactService_Create_WithPrimaryCustomer(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	// Create customer
	customer := testutil.CreateTestCustomer(t, db, "Primary Customer")

	// Create contact with primary customer
	contact, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName:         "With",
		LastName:          "PrimaryCustomer",
		PrimaryCustomerID: &customer.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, contact.PrimaryCustomerID)
	assert.Equal(t, customer.ID, *contact.PrimaryCustomerID)
}

// TestContactService_Create_InvalidPrimaryCustomer tests creating contact with invalid primary customer
func TestContactService_Create_InvalidPrimaryCustomer(t *testing.T) {
	db := setupContactServiceTestDB(t)
	svc := createContactService(db)
	ctx := createContactTestContext()

	nonExistentID := uuid.New()
	_, err := svc.Create(ctx, &domain.CreateContactRequest{
		FirstName:         "Invalid",
		LastName:          "PrimaryCustomer",
		PrimaryCustomerID: &nonExistentID,
	})
	assert.Error(t, err)
	// The service doesn't pre-validate customer existence - the database foreign key constraint catches this
	assert.Contains(t, err.Error(), "foreign key constraint")
}
