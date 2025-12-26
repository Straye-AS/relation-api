package mapper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/stretchr/testify/assert"
)

func TestToCustomerDTO(t *testing.T) {
	now := time.Now()
	customer := &domain.Customer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
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

	dto := mapper.ToCustomerDTO(customer, 0.0, 0.0, 0)

	assert.Equal(t, customer.ID, dto.ID)
	assert.Equal(t, customer.Name, dto.Name)
	assert.Equal(t, customer.OrgNumber, dto.OrgNumber)
	assert.Equal(t, customer.Email, dto.Email)
	assert.Equal(t, customer.Phone, dto.Phone)
	assert.Equal(t, customer.Address, dto.Address)
	assert.Equal(t, customer.City, dto.City)
	assert.Equal(t, customer.PostalCode, dto.PostalCode)
	assert.Equal(t, customer.Country, dto.Country)
	assert.Equal(t, customer.ContactPerson, dto.ContactPerson)
	assert.Equal(t, customer.ContactEmail, dto.ContactEmail)
	assert.Equal(t, customer.ContactPhone, dto.ContactPhone)
	assert.NotEmpty(t, dto.CreatedAt)
	assert.NotEmpty(t, dto.UpdatedAt)
}

func TestToContactDTO(t *testing.T) {
	customerID := uuid.New()
	contact := &domain.Contact{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName:         "John",
		LastName:          "Doe",
		Email:             "john.doe@example.com",
		Phone:             "+1234567890",
		Title:             "CEO",
		PrimaryCustomerID: &customerID,
		IsActive:          true,
	}

	dto := mapper.ToContactDTO(contact)

	assert.Equal(t, contact.ID, dto.ID)
	assert.Equal(t, "John", dto.FirstName)
	assert.Equal(t, "Doe", dto.LastName)
	assert.Equal(t, "John Doe", dto.FullName)
	assert.Equal(t, "john.doe@example.com", dto.Email)
	assert.Equal(t, "+1234567890", dto.Phone)
	assert.Equal(t, "CEO", dto.Title)
	assert.Equal(t, &customerID, dto.PrimaryCustomerID)
	assert.True(t, dto.IsActive)
}

// TestToProjectBudgetDTO verifies that the deprecated ToProjectBudgetDTO function
// returns zeros since Project no longer has economic fields (moved to Offer)
func TestToProjectBudgetDTO(t *testing.T) {
	project := &domain.Project{
		BaseModel: domain.BaseModel{
			ID: uuid.New(),
		},
		Name: "Test Project",
	}

	dto := mapper.ToProjectBudgetDTO(project)

	// Project no longer has economic fields - function returns zeros
	assert.Equal(t, 0.0, dto.Value)
	assert.Equal(t, 0.0, dto.Cost)
	assert.Equal(t, 0.0, dto.MarginPercent)
	assert.Equal(t, 0.0, dto.Spent)
	assert.Equal(t, 0.0, dto.Remaining)
	assert.Equal(t, 0.0, dto.PercentUsed)
}

// TestToProjectBudgetDTO_ReturnsZeros confirms the deprecated function always returns zeros
func TestToProjectBudgetDTO_ReturnsZeros(t *testing.T) {
	project := &domain.Project{
		BaseModel: domain.BaseModel{
			ID: uuid.New(),
		},
		Name: "Another Project",
	}

	dto := mapper.ToProjectBudgetDTO(project)

	// All economic fields should be zero since they've moved to Offer
	assert.Equal(t, 0.0, dto.Value)
	assert.Equal(t, 0.0, dto.Cost)
	assert.Equal(t, 0.0, dto.MarginPercent)
	assert.Equal(t, 0.0, dto.Spent)
	assert.Equal(t, 0.0, dto.Remaining)
	assert.Equal(t, 0.0, dto.PercentUsed)
}

func TestToOfferItemDTO(t *testing.T) {
	offerID := uuid.New()
	item := &domain.OfferItem{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		OfferID:     offerID,
		Discipline:  "Product A",
		Description: "Description of Product A",
		Quantity:    10,
		Unit:        "Product A",
		Cost:        100,
		Revenue:     1000,
		Margin:      10,
	}

	dto := mapper.ToOfferItemDTO(item)

	assert.Equal(t, item.ID, dto.ID)
	assert.Equal(t, "Product A", dto.Discipline)
	assert.Equal(t, "Description of Product A", dto.Description)
	assert.Equal(t, 10.0, dto.Quantity)
	assert.Equal(t, "Product A", dto.Unit)
	assert.Equal(t, 100.0, dto.Cost)
	assert.Equal(t, 1000.0, dto.Revenue)
	assert.Equal(t, 10.0, dto.Margin)
}

func TestToActivityDTO(t *testing.T) {
	targetID := uuid.New()
	occurredAt := time.Now()
	activity := &domain.Activity{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     targetID,
		Title:        "Customer created",
		Body:         "New customer was created",
		OccurredAt:   occurredAt,
		CreatorName:  "Test User",
		ActivityType: domain.ActivityTypeNote,
		Status:       domain.ActivityStatusCompleted,
		Priority:     0,
		IsPrivate:    false,
	}

	dto := mapper.ToActivityDTO(activity)

	assert.Equal(t, activity.ID, dto.ID)
	assert.Equal(t, domain.ActivityTargetCustomer, dto.TargetType)
	assert.Equal(t, targetID, dto.TargetID)
	assert.Equal(t, "Customer created", dto.Title)
	assert.Equal(t, "New customer was created", dto.Body)
	assert.Equal(t, "Test User", dto.CreatorName)
	assert.Equal(t, domain.ActivityTypeNote, dto.ActivityType)
	assert.Equal(t, domain.ActivityStatusCompleted, dto.Status)
}

func TestToActivityDTO_WithAttendees(t *testing.T) {
	targetID := uuid.New()
	parentActivityID := uuid.New()
	occurredAt := time.Now()
	attendee1 := uuid.New().String()
	attendee2 := uuid.New().String()

	activity := &domain.Activity{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		TargetType:       domain.ActivityTargetCustomer,
		TargetID:         targetID,
		Title:            "Team Meeting",
		Body:             "Weekly sync meeting",
		OccurredAt:       occurredAt,
		CreatorName:      "Test User",
		ActivityType:     domain.ActivityTypeMeeting,
		Status:           domain.ActivityStatusPlanned,
		Priority:         2,
		IsPrivate:        false,
		Attendees:        []string{attendee1, attendee2},
		ParentActivityID: &parentActivityID,
	}

	dto := mapper.ToActivityDTO(activity)

	assert.Equal(t, activity.ID, dto.ID)
	assert.Equal(t, domain.ActivityTypeMeeting, dto.ActivityType)
	assert.Equal(t, "Team Meeting", dto.Title)
	assert.Len(t, dto.Attendees, 2)
	assert.Contains(t, dto.Attendees, attendee1)
	assert.Contains(t, dto.Attendees, attendee2)
	assert.Equal(t, &parentActivityID, dto.ParentActivityID)
}

func TestToActivityDTO_EmptyAttendees(t *testing.T) {
	targetID := uuid.New()
	occurredAt := time.Now()

	activity := &domain.Activity{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     targetID,
		Title:        "Task",
		OccurredAt:   occurredAt,
		ActivityType: domain.ActivityTypeTask,
		Status:       domain.ActivityStatusPlanned,
	}

	dto := mapper.ToActivityDTO(activity)

	assert.Empty(t, dto.Attendees)
	assert.Nil(t, dto.ParentActivityID)
}

// ============================================================================
// ToOfferDTO Tests with Order Phase Fields
// ============================================================================

func TestToOfferDTO_BasicFields(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	offer := &domain.Offer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:         "Test Offer",
		OfferNumber:   "SB-2024-001",
		CustomerID:    &customerID,
		CustomerName:  "Test Customer",
		CompanyID:     domain.CompanyStalbygg,
		Phase:         domain.OfferPhaseSent,
		Probability:   75,
		Value:         100000,
		Cost:          80000,
		MarginPercent: 20,
		Status:        domain.OfferStatusActive,
	}

	dto := mapper.ToOfferDTO(offer)

	assert.Equal(t, offer.ID, dto.ID)
	assert.Equal(t, "Test Offer", dto.Title)
	assert.Equal(t, "SB-2024-001", dto.OfferNumber)
	assert.Equal(t, &customerID, dto.CustomerID)
	assert.Equal(t, "Test Customer", dto.CustomerName)
	assert.Equal(t, domain.CompanyStalbygg, dto.CompanyID)
	assert.Equal(t, domain.OfferPhaseSent, dto.Phase)
	assert.Equal(t, 75, dto.Probability)
	assert.Equal(t, 100000.0, dto.Value)
	assert.Equal(t, 80000.0, dto.Cost)
	assert.Equal(t, 20000.0, dto.Margin) // Value - Cost
	assert.Equal(t, 20.0, dto.MarginPercent)
}

func TestToOfferDTO_WithOrderPhaseFields(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()
	managerID := "manager-uuid-123"
	startDate := now.AddDate(0, -1, 0) // 1 month ago
	endDate := now.AddDate(0, 1, 0)    // 1 month from now
	estimatedCompletion := now.AddDate(0, 0, 15)
	health := domain.OfferHealthAtRisk
	completionPct := 65.5

	offer := &domain.Offer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:        "Order Phase Offer",
		OfferNumber:  "SB-2024-002O",
		CustomerID:   &customerID,
		CustomerName: "Test Customer",
		CompanyID:    domain.CompanyStalbygg,
		Phase:        domain.OfferPhaseOrder,
		Value:        150000,
		Cost:         120000,
		Status:       domain.OfferStatusActive,
		// Order phase execution fields
		ManagerID:               &managerID,
		ManagerName:             "John Manager",
		TeamMembers:             []string{"user-1", "user-2", "user-3"},
		Spent:                   50000,
		Invoiced:                75000,
		Health:                  &health,
		CompletionPercent:       &completionPct,
		StartDate:               &startDate,
		EndDate:                 &endDate,
		EstimatedCompletionDate: &estimatedCompletion,
	}

	dto := mapper.ToOfferDTO(offer)

	// Basic fields
	assert.Equal(t, offer.ID, dto.ID)
	assert.Equal(t, domain.OfferPhaseOrder, dto.Phase)
	assert.Equal(t, "SB-2024-002O", dto.OfferNumber)

	// Order phase execution fields
	assert.NotNil(t, dto.ManagerID)
	assert.Equal(t, "manager-uuid-123", *dto.ManagerID)
	assert.Equal(t, "John Manager", dto.ManagerName)
	assert.Len(t, dto.TeamMembers, 3)
	assert.Contains(t, dto.TeamMembers, "user-1")
	assert.Contains(t, dto.TeamMembers, "user-2")
	assert.Contains(t, dto.TeamMembers, "user-3")
	assert.Equal(t, 50000.0, dto.Spent)
	assert.Equal(t, 75000.0, dto.Invoiced)

	// Health and completion
	assert.NotNil(t, dto.Health)
	assert.Equal(t, "at_risk", *dto.Health)
	assert.NotNil(t, dto.CompletionPercent)
	assert.Equal(t, 65.5, *dto.CompletionPercent)

	// Dates
	assert.NotNil(t, dto.StartDate)
	assert.NotNil(t, dto.EndDate)
	assert.NotNil(t, dto.EstimatedCompletionDate)
}

func TestToOfferDTO_WithNilOrderPhaseFields(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	// Draft offer without order phase fields
	offer := &domain.Offer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:        "Draft Offer",
		CustomerID:   &customerID,
		CustomerName: "Test Customer",
		CompanyID:    domain.CompanyStalbygg,
		Phase:        domain.OfferPhaseDraft,
		Value:        50000,
		Status:       domain.OfferStatusActive,
		// All order phase fields are nil/zero
	}

	dto := mapper.ToOfferDTO(offer)

	assert.Equal(t, domain.OfferPhaseDraft, dto.Phase)
	assert.Nil(t, dto.ManagerID)
	assert.Empty(t, dto.ManagerName)
	assert.Empty(t, dto.TeamMembers)
	assert.Equal(t, 0.0, dto.Spent)
	assert.Equal(t, 0.0, dto.Invoiced)
	assert.Nil(t, dto.Health)
	assert.Nil(t, dto.CompletionPercent)
	assert.Nil(t, dto.StartDate)
	assert.Nil(t, dto.EndDate)
	assert.Nil(t, dto.EstimatedCompletionDate)
}

func TestToOfferDTO_AllHealthStatuses(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name           string
		health         domain.OfferHealth
		expectedString string
	}{
		{"on_track", domain.OfferHealthOnTrack, "on_track"},
		{"at_risk", domain.OfferHealthAtRisk, "at_risk"},
		{"delayed", domain.OfferHealthDelayed, "delayed"},
		{"over_budget", domain.OfferHealthOverBudget, "over_budget"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			offer := &domain.Offer{
				BaseModel: domain.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Title:     "Test Offer",
				Phase:     domain.OfferPhaseOrder,
				CompanyID: domain.CompanyStalbygg,
				Health:    &tc.health,
				Status:    domain.OfferStatusActive,
			}

			dto := mapper.ToOfferDTO(offer)

			assert.NotNil(t, dto.Health)
			assert.Equal(t, tc.expectedString, *dto.Health)
		})
	}
}

// ============================================================================
// ToFileDTO and ToFileDTOs Tests
// ============================================================================

func TestToFileDTO_WithOfferID(t *testing.T) {
	now := time.Now()
	offerID := uuid.New()

	file := &domain.File{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
		},
		Filename:    "document.pdf",
		ContentType: "application/pdf",
		Size:        1024,
		StoragePath: "files/abc123.pdf",
		CompanyID:   domain.CompanyStalbygg,
		OfferID:     &offerID,
	}

	dto := mapper.ToFileDTO(file)

	assert.Equal(t, file.ID, dto.ID)
	assert.Equal(t, "document.pdf", dto.Filename)
	assert.Equal(t, "application/pdf", dto.ContentType)
	assert.Equal(t, int64(1024), dto.Size)
	assert.Equal(t, domain.CompanyStalbygg, dto.CompanyID)
	assert.Equal(t, &offerID, dto.OfferID)
	assert.Nil(t, dto.CustomerID)
	assert.Nil(t, dto.ProjectID)
	assert.Nil(t, dto.SupplierID)
	assert.NotEmpty(t, dto.CreatedAt)
}

func TestToFileDTO_WithCustomerID(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	file := &domain.File{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
		},
		Filename:    "contract.pdf",
		ContentType: "application/pdf",
		Size:        2048,
		StoragePath: "files/def456.pdf",
		CompanyID:   domain.CompanyGruppen,
		CustomerID:  &customerID,
	}

	dto := mapper.ToFileDTO(file)

	assert.Equal(t, file.ID, dto.ID)
	assert.Equal(t, "contract.pdf", dto.Filename)
	assert.Equal(t, domain.CompanyGruppen, dto.CompanyID)
	assert.Equal(t, &customerID, dto.CustomerID)
	assert.Nil(t, dto.OfferID)
	assert.Nil(t, dto.ProjectID)
	assert.Nil(t, dto.SupplierID)
}

func TestToFileDTO_WithProjectID(t *testing.T) {
	now := time.Now()
	projectID := uuid.New()

	file := &domain.File{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
		},
		Filename:    "blueprint.dwg",
		ContentType: "application/acad",
		Size:        5120,
		StoragePath: "files/ghi789.dwg",
		CompanyID:   domain.CompanyHybridbygg,
		ProjectID:   &projectID,
	}

	dto := mapper.ToFileDTO(file)

	assert.Equal(t, file.ID, dto.ID)
	assert.Equal(t, "blueprint.dwg", dto.Filename)
	assert.Equal(t, domain.CompanyHybridbygg, dto.CompanyID)
	assert.Equal(t, &projectID, dto.ProjectID)
	assert.Nil(t, dto.OfferID)
	assert.Nil(t, dto.CustomerID)
	assert.Nil(t, dto.SupplierID)
}

func TestToFileDTO_WithSupplierID(t *testing.T) {
	now := time.Now()
	supplierID := uuid.New()

	file := &domain.File{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
		},
		Filename:    "invoice.pdf",
		ContentType: "application/pdf",
		Size:        512,
		StoragePath: "files/jkl012.pdf",
		CompanyID:   domain.CompanyTak,
		SupplierID:  &supplierID,
	}

	dto := mapper.ToFileDTO(file)

	assert.Equal(t, file.ID, dto.ID)
	assert.Equal(t, "invoice.pdf", dto.Filename)
	assert.Equal(t, domain.CompanyTak, dto.CompanyID)
	assert.Equal(t, &supplierID, dto.SupplierID)
	assert.Nil(t, dto.OfferID)
	assert.Nil(t, dto.CustomerID)
	assert.Nil(t, dto.ProjectID)
}

func TestToFileDTOs_MultipleFiles(t *testing.T) {
	now := time.Now()
	offerID := uuid.New()
	customerID := uuid.New()

	files := []domain.File{
		{
			BaseModel: domain.BaseModel{
				ID:        uuid.New(),
				CreatedAt: now,
			},
			Filename:    "file1.pdf",
			ContentType: "application/pdf",
			Size:        1024,
			StoragePath: "files/file1.pdf",
			CompanyID:   domain.CompanyStalbygg,
			OfferID:     &offerID,
		},
		{
			BaseModel: domain.BaseModel{
				ID:        uuid.New(),
				CreatedAt: now,
			},
			Filename:    "file2.pdf",
			ContentType: "application/pdf",
			Size:        2048,
			StoragePath: "files/file2.pdf",
			CompanyID:   domain.CompanyGruppen,
			CustomerID:  &customerID,
		},
	}

	dtos := mapper.ToFileDTOs(files)

	assert.Len(t, dtos, 2)
	assert.Equal(t, "file1.pdf", dtos[0].Filename)
	assert.Equal(t, domain.CompanyStalbygg, dtos[0].CompanyID)
	assert.Equal(t, &offerID, dtos[0].OfferID)
	assert.Equal(t, "file2.pdf", dtos[1].Filename)
	assert.Equal(t, domain.CompanyGruppen, dtos[1].CompanyID)
	assert.Equal(t, &customerID, dtos[1].CustomerID)
}

func TestToFileDTOs_EmptySlice(t *testing.T) {
	files := []domain.File{}

	dtos := mapper.ToFileDTOs(files)

	assert.NotNil(t, dtos)
	assert.Len(t, dtos, 0)
}

// ============================================================================
// Offer Warnings Tests
// ============================================================================

func TestToOfferDTO_WarningsAddedWhenOrderPhaseAndValuesDiffer(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	offer := &domain.Offer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:         "Order with Value Mismatch",
		CustomerID:    &customerID,
		CustomerName:  "Test Customer",
		CompanyID:     domain.CompanyStalbygg,
		Phase:         domain.OfferPhaseOrder, // Must be in order phase
		Value:         100000,                 // Offer value
		DWTotalIncome: 120000,                 // DW value differs
		Status:        domain.OfferStatusActive,
	}

	dto := mapper.ToOfferDTO(offer)

	// Warnings should be populated
	assert.NotNil(t, dto.Warnings)
	assert.Len(t, dto.Warnings, 1)
	assert.Contains(t, dto.Warnings, domain.OfferWarningValueNotEqualsDWTotalIncome)
}

func TestToOfferDTO_NoWarningWhenOrderPhaseAndValuesEqual(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	offer := &domain.Offer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:         "Order with Matching Values",
		CustomerID:    &customerID,
		CustomerName:  "Test Customer",
		CompanyID:     domain.CompanyStalbygg,
		Phase:         domain.OfferPhaseOrder, // Must be in order phase
		Value:         100000,                 // Offer value
		DWTotalIncome: 100000,                 // DW value matches
		Status:        domain.OfferStatusActive,
	}

	dto := mapper.ToOfferDTO(offer)

	// No warnings should be present
	assert.Nil(t, dto.Warnings)
}

func TestToOfferDTO_NoWarningWhenNotInOrderPhase(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	testCases := []struct {
		name  string
		phase domain.OfferPhase
	}{
		{"draft phase", domain.OfferPhaseDraft},
		{"in_progress phase", domain.OfferPhaseInProgress},
		{"sent phase", domain.OfferPhaseSent},
		{"completed phase", domain.OfferPhaseCompleted},
		{"lost phase", domain.OfferPhaseLost},
		{"expired phase", domain.OfferPhaseExpired},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			offer := &domain.Offer{
				BaseModel: domain.BaseModel{
					ID:        uuid.New(),
					CreatedAt: now,
					UpdatedAt: now,
				},
				Title:         "Offer Not in Order Phase",
				CustomerID:    &customerID,
				CustomerName:  "Test Customer",
				CompanyID:     domain.CompanyStalbygg,
				Phase:         tc.phase, // Not order phase
				Value:         100000,   // Offer value
				DWTotalIncome: 120000,   // DW value differs - but should NOT trigger warning
				Status:        domain.OfferStatusActive,
			}

			dto := mapper.ToOfferDTO(offer)

			// No warnings should be present regardless of value mismatch
			assert.Nil(t, dto.Warnings, "Warnings should be nil for phase: %s", tc.phase)
		})
	}
}

func TestToOfferDTO_NoWarningWhenDWTotalIncomeIsZero(t *testing.T) {
	now := time.Now()
	customerID := uuid.New()

	offer := &domain.Offer{
		BaseModel: domain.BaseModel{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:         "Order with No DW Sync",
		CustomerID:    &customerID,
		CustomerName:  "Test Customer",
		CompanyID:     domain.CompanyStalbygg,
		Phase:         domain.OfferPhaseOrder, // Order phase
		Value:         100000,                 // Offer value
		DWTotalIncome: 0,                      // Not synced yet (zero value)
		Status:        domain.OfferStatusActive,
	}

	dto := mapper.ToOfferDTO(offer)

	// No warnings - DW hasn't been synced yet
	assert.Nil(t, dto.Warnings)
}
