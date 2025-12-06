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

func (s *ContactService) Create(ctx context.Context, req *domain.CreateContactRequest) (*domain.ContactDTO, error) {
	contact := &domain.Contact{
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Phone:                  req.Phone,
		Mobile:                 req.Mobile,
		Title:                  req.Title,
		Department:             req.Department,
		PrimaryCustomerID:      req.PrimaryCustomerID,
		Address:                req.Address,
		City:                   req.City,
		PostalCode:             req.PostalCode,
		Country:                req.Country,
		LinkedInURL:            req.LinkedInURL,
		PreferredContactMethod: req.PreferredContactMethod,
		Notes:                  req.Notes,
		IsActive:               true,
	}

	// Set default country if not provided
	if contact.Country == "" {
		contact.Country = "Norway"
	}

	// Set default preferred contact method if not provided
	if contact.PreferredContactMethod == "" {
		contact.PreferredContactMethod = "email"
	}

	if err := s.contactRepo.Create(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetContact,
			TargetID:    contact.ID,
			Title:       "Contact created",
			Body:        fmt.Sprintf("Contact '%s' was created", contact.FullName()),
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
	contacts, err := s.contactRepo.ListByEntity(ctx, domain.ContactEntityCustomer, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.ToContactDTO(&contact)
	}

	return dtos, nil
}

func (s *ContactService) List(ctx context.Context, page, pageSize int, filters *repository.ContactFilters, sortBy repository.ContactSortOption) ([]domain.ContactDTO, int64, error) {
	contacts, total, err := s.contactRepo.List(ctx, page, pageSize, filters, sortBy)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.ToContactDTO(&contact)
	}

	return dtos, total, nil
}

func (s *ContactService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateContactRequest) (*domain.ContactDTO, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}

	contact.FirstName = req.FirstName
	contact.LastName = req.LastName
	contact.Email = req.Email
	contact.Phone = req.Phone
	contact.Mobile = req.Mobile
	contact.Title = req.Title
	contact.Department = req.Department
	contact.PrimaryCustomerID = req.PrimaryCustomerID
	contact.Address = req.Address
	contact.City = req.City
	contact.PostalCode = req.PostalCode
	contact.Country = req.Country
	contact.LinkedInURL = req.LinkedInURL
	contact.PreferredContactMethod = req.PreferredContactMethod
	contact.Notes = req.Notes

	if req.IsActive != nil {
		contact.IsActive = *req.IsActive
	}

	if err := s.contactRepo.Update(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetContact,
			TargetID:    contact.ID,
			Title:       "Contact updated",
			Body:        fmt.Sprintf("Contact '%s' was updated", contact.FullName()),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToContactDTO(contact)
	return &dto, nil
}

func (s *ContactService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.contactRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	return nil
}
