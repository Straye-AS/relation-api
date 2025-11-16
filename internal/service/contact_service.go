package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

type ContactService struct {
	contactRepo  *repository.ContactRepository
	customerRepo *repository.CustomerRepository
	activityRepo *repository.ActivityRepository
	logger       *zap.Logger
}

func NewContactService(
	contactRepo *repository.ContactRepository,
	customerRepo *repository.CustomerRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *ContactService {
	return &ContactService{
		contactRepo:  contactRepo,
		customerRepo: customerRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

func (s *ContactService) Create(ctx context.Context, customerID uuid.UUID, req *domain.CreateContactRequest) (*domain.ContactDTO, error) {
	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	contact := &domain.Contact{
		CustomerID:   &customerID,
		CustomerName: customer.Name,
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		Role:         req.Role,
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTypeContact,
			TargetID:    contact.ID,
			Title:       "Contact created",
			Body:        fmt.Sprintf("Contact '%s' was created", contact.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToContactDTO(contact)
	return &dto, nil
}

func (s *ContactService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContactDTO, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	dto := mapper.ToContactDTO(contact)
	return &dto, nil
}

func (s *ContactService) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.ContactDTO, error) {
	contacts, err := s.contactRepo.ListByCustomer(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.ToContactDTO(&contact)
	}

	return dtos, nil
}

func (s *ContactService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.contactRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}
