package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrDuplicateContactEmail is returned when trying to create a contact with an existing email
var ErrDuplicateContactEmail = errors.New("contact with this email already exists")

// Note: ErrInvalidEmailFormat and emailRegex are defined in customer_service.go

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
	// Validate email format if provided
	if req.Email != "" {
		if !emailRegex.MatchString(req.Email) {
			return nil, ErrInvalidEmailFormat
		}

		// Check for duplicate email
		existing, err := s.contactRepo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil {
			return nil, ErrDuplicateContactEmail
		}
		// Only non-not-found errors are actual errors
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}
	}

	// Set default contact type if not provided
	contactType := req.ContactType
	if contactType == "" {
		contactType = domain.ContactTypePrimary
	}

	contact := &domain.Contact{
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Phone:                  req.Phone,
		Mobile:                 req.Mobile,
		Title:                  req.Title,
		Department:             req.Department,
		ContactType:            contactType,
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
		// Check for unique constraint violation on email
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrDuplicateContactEmail
		}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("contact not found: %w", err)
		}
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

func (s *ContactService) List(ctx context.Context, page, pageSize int) ([]domain.ContactDTO, int64, error) {
	contacts, total, err := s.contactRepo.List(ctx, page, pageSize)
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

	// Check for duplicate email if it's being changed
	if req.Email != "" && req.Email != contact.Email {
		existing, err := s.contactRepo.GetByEmail(ctx, req.Email)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrDuplicateContactEmail
		}
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

	// Update contact type if provided, keep existing if empty
	if req.ContactType != "" {
		contact.ContactType = req.ContactType
	}

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
	// Get the contact first for activity logging
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("contact not found: %w", err)
	}

	if err := s.contactRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetContact,
			TargetID:    id,
			Title:       "Contact deleted",
			Body:        fmt.Sprintf("Contact '%s' was deleted", contact.FullName()),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	return nil
}

// ListWithFilters returns contacts with filters and pagination
func (s *ContactService) ListWithFilters(ctx context.Context, page, pageSize int, filters *repository.ContactFilters, sortBy repository.ContactSortOption) (*domain.PaginatedResponse, error) {
	contacts, total, err := s.contactRepo.ListWithFilters(ctx, page, pageSize, filters, sortBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dto := mapper.ToContactDTO(&contact)
		// Add primary customer name if available
		if contact.PrimaryCustomer != nil {
			dto.PrimaryCustomerName = contact.PrimaryCustomer.Name
		}
		dtos[i] = dto
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &domain.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListByEntity returns contacts related to a specific entity
func (s *ContactService) ListByEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) ([]domain.ContactDTO, error) {
	contacts, err := s.contactRepo.ListByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.ToContactDTO(&contact)
	}

	return dtos, nil
}

// AddRelationship adds a relationship between a contact and an entity
func (s *ContactService) AddRelationship(ctx context.Context, contactID uuid.UUID, req *domain.AddContactRelationshipRequest) (*domain.ContactRelationshipDTO, error) {
	// Check if contact exists
	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}

	// Check if relationship already exists
	exists, err := s.contactRepo.CheckRelationshipExists(ctx, contactID, req.EntityType, req.EntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check relationship: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("relationship already exists")
	}

	// Create the relationship
	rel := &domain.ContactRelationship{
		ContactID:  contactID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Role:       req.Role,
		IsPrimary:  req.IsPrimary,
	}

	if err := s.contactRepo.AddRelationship(ctx, rel); err != nil {
		return nil, fmt.Errorf("failed to add relationship: %w", err)
	}

	// If this is primary, ensure no other relationship of this type is primary
	if req.IsPrimary {
		if err := s.contactRepo.SetPrimaryRelationship(ctx, contactID, req.EntityType, req.EntityID); err != nil {
			s.logger.Warn("failed to set primary relationship", zap.Error(err))
		}
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetContact,
			TargetID:    contactID,
			Title:       "Relationship added",
			Body:        fmt.Sprintf("Contact '%s' was linked to %s", contact.FullName(), req.EntityType),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToContactRelationshipDTO(rel)
	return &dto, nil
}

// RemoveRelationship removes a relationship by its ID
func (s *ContactService) RemoveRelationship(ctx context.Context, contactID, relationshipID uuid.UUID) error {
	// Verify the relationship exists and belongs to this contact
	rel, err := s.contactRepo.GetRelationshipByID(ctx, relationshipID)
	if err != nil {
		return fmt.Errorf("relationship not found: %w", err)
	}

	if rel.ContactID != contactID {
		return fmt.Errorf("relationship does not belong to this contact")
	}

	// Get contact for activity logging
	contact, _ := s.contactRepo.GetByID(ctx, contactID)

	if err := s.contactRepo.RemoveRelationshipByID(ctx, relationshipID); err != nil {
		return fmt.Errorf("failed to remove relationship: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok && contact != nil {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetContact,
			TargetID:    contactID,
			Title:       "Relationship removed",
			Body:        fmt.Sprintf("Contact '%s' was unlinked from %s", contact.FullName(), rel.EntityType),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	return nil
}
