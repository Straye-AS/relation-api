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
		CustomerID: &customerID,
		Name:       "John Doe",
		Email:      "john.doe@example.com",
		Phone:      "+1234567890",
		Role:       "CEO",
	}

	dto := mapper.ToContactDTO(contact)

	assert.Equal(t, contact.ID, dto.ID)
	assert.Equal(t, &customerID, dto.CustomerID)
	assert.Equal(t, "John Doe", dto.Name)
	assert.Equal(t, "john.doe@example.com", dto.Email)
	assert.Equal(t, "+1234567890", dto.Phone)
	assert.Equal(t, "CEO", dto.Role)
}

func TestToProjectBudgetDTO(t *testing.T) {
	project := &domain.Project{
		Budget: 100000,
		Spent:  25000,
	}

	dto := mapper.ToProjectBudgetDTO(project)

	assert.Equal(t, 100000.0, dto.Budget)
	assert.Equal(t, 25000.0, dto.Spent)
	assert.Equal(t, 75000.0, dto.Remaining)
	assert.Equal(t, 25.0, dto.PercentUsed)
}

func TestToProjectBudgetDTO_ZeroBudget(t *testing.T) {
	project := &domain.Project{
		Budget: 0,
		Spent:  0,
	}

	dto := mapper.ToProjectBudgetDTO(project)

	assert.Equal(t, 0.0, dto.Budget)
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
		TargetType:  domain.ActivityTypeCustomer,
		TargetID:    targetID,
		Title:       "Customer created",
		Body:        "New customer was created",
		OccurredAt:  occurredAt,
		CreatorName: "Test User",
	}

	dto := mapper.ToActivityDTO(activity)

	assert.Equal(t, activity.ID, dto.ID)
	assert.Equal(t, domain.ActivityTypeCustomer, dto.TargetType)
	assert.Equal(t, targetID, dto.TargetID)
	assert.Equal(t, "Customer created", dto.Title)
	assert.Equal(t, "New customer was created", dto.Body)
	assert.Equal(t, "Test User", dto.CreatorName)
}
