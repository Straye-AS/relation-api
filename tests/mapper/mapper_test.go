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

	dto := mapper.ToCustomerDTO(customer, 0.0, 0)

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

func TestToProjectBudgetDTO(t *testing.T) {
	project := &domain.Project{
		Value:         100000,
		Cost:          60000,
		MarginPercent: 40.0,
		Spent:         25000,
	}

	dto := mapper.ToProjectBudgetDTO(project)

	assert.Equal(t, 100000.0, dto.Value)
	assert.Equal(t, 60000.0, dto.Cost)
	assert.Equal(t, 40.0, dto.MarginPercent)
	assert.Equal(t, 25000.0, dto.Spent)
	assert.Equal(t, 75000.0, dto.Remaining)        // Value - Spent = 100000 - 25000
	assert.InDelta(t, 25.0, dto.PercentUsed, 0.01) // Spent/Value * 100 = 25000/100000 * 100
}

func TestToProjectBudgetDTO_ZeroBudget(t *testing.T) {
	project := &domain.Project{
		Value:         0,
		Cost:          0,
		MarginPercent: 0,
		Spent:         0,
	}

	dto := mapper.ToProjectBudgetDTO(project)

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
