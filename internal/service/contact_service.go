package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ContactService provides business logic for contact management
type ContactService struct {
	contactRepo  *repository.ContactRepository
	customerRepo *repository.CustomerRepository
	dealRepo     *repository.DealRepository
	projectRepo  *repository.ProjectRepository
	activityRepo *repository.ActivityRepository
	logger       *zap.Logger
}

// NewContactService creates a new ContactService with all dependencies
func NewContactService(
	contactRepo *repository.ContactRepository,
	customerRepo *repository.CustomerRepository,
	dealRepo *repository.DealRepository,
	projectRepo *repository.ProjectRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *ContactService {
	return &ContactService{
		contactRepo:  contactRepo,
		customerRepo: customerRepo,
		dealRepo:     dealRepo,
		projectRepo:  projectRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

// emailRegex is a simple email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// validateEmail validates email format
func validateEmail(email string) error {
	if email == "" {
		return nil // Email is optional
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// validateEntityType validates the entity type
func validateEntityType(entityType domain.ContactEntityType) error {
	switch entityType {
	case domain.ContactEntityCustomer, domain.ContactEntityDeal, domain.ContactEntityProject:
		return nil
	default:
		return fmt.Errorf("invalid entity type: %s", entityType)
	}
}

// validateEntityExists checks if the entity exists in the database
func (s *ContactService) validateEntityExists(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) error {
	switch entityType {
	case domain.ContactEntityCustomer:
		_, err := s.customerRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("customer with ID %s not found", entityID)
			}
			return fmt.Errorf("failed to validate customer: %w", err)
		}
	case domain.ContactEntityDeal:
		_, err := s.dealRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("deal with ID %s not found", entityID)
			}
			return fmt.Errorf("failed to validate deal: %w", err)
		}
	case domain.ContactEntityProject:
		_, err := s.projectRepo.GetByID(ctx, entityID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("project with ID %s not found", entityID)
			}
			return fmt.Errorf("failed to validate project: %w", err)
		}
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}
	return nil
}

// logActivity creates an activity log entry
func (s *ContactService) logActivity(ctx context.Context, targetID uuid.UUID, title, body string) {
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetContact,
			TargetID:    targetID,
			Title:       title,
			Body:        body,
			CreatorName: userCtx.DisplayName,
			CreatorID:   userCtx.UserID.String(),
		}
		if err := s.activityRepo.Create(ctx, activity); err != nil {
			s.logger.Error("failed to log activity",
				zap.String("title", title),
				zap.Error(err))
		}
	}
}

// Create creates a new contact with validation
func (s *ContactService) Create(ctx context.Context, req *domain.CreateContactRequest) (*domain.ContactDTO, error) {
	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Check email uniqueness (case-insensitive) if email is provided
	if req.Email != "" {
		exists, err := s.contactRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("contact with email %s already exists", req.Email)
		}
	}

	// Validate primary customer exists if provided
	if req.PrimaryCustomerID != nil {
		if err := s.validateEntityExists(ctx, domain.ContactEntityCustomer, *req.PrimaryCustomerID); err != nil {
			return nil, err
		}
	}

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

	// Log activity
	s.logActivity(ctx, contact.ID, "Contact created",
		fmt.Sprintf("Contact '%s' was created", contact.FullName()))

	dto := mapper.ToContactDTO(contact)
	return &dto, nil
}

// GetByID retrieves a contact by ID
func (s *ContactService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ContactDTO, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	dto := mapper.ToContactDTO(contact)
	return &dto, nil
}

// ListByCustomer retrieves contacts for a specific customer
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

// List retrieves contacts with pagination
func (s *ContactService) List(ctx context.Context, page, pageSize int) ([]domain.ContactDTO, int64, error) {
	// Clamp page size
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page < 1 {
		page = 1
	}

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

// Update updates an existing contact
func (s *ContactService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateContactRequest) (*domain.ContactDTO, error) {
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Validate email format if provided
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Check email uniqueness if email is being changed
	if req.Email != "" && !strings.EqualFold(req.Email, contact.Email) {
		exists, err := s.contactRepo.ExistsByEmailExcluding(ctx, req.Email, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("contact with email %s already exists", req.Email)
		}
	}

	// Validate primary customer exists if provided
	if req.PrimaryCustomerID != nil {
		if err := s.validateEntityExists(ctx, domain.ContactEntityCustomer, *req.PrimaryCustomerID); err != nil {
			return nil, err
		}
	}

	// Update fields
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

	// Log activity
	s.logActivity(ctx, contact.ID, "Contact updated",
		fmt.Sprintf("Contact '%s' was updated", contact.FullName()))

	dto := mapper.ToContactDTO(contact)
	return &dto, nil
}

// Delete deletes a contact
func (s *ContactService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get contact first for logging
	contact, err := s.contactRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("contact not found")
		}
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Check if contact has relationships - warn but allow deletion
	hasRelationships, err := s.contactRepo.HasRelationships(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check contact relationships: %w", err)
	}

	if hasRelationships {
		s.logger.Warn("deleting contact with active relationships",
			zap.String("contact_id", id.String()),
			zap.String("contact_name", contact.FullName()))
	}

	if err := s.contactRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	// Log activity (note: this logs against a now-deleted contact, but still useful for audit)
	s.logActivity(ctx, id, "Contact deleted",
		fmt.Sprintf("Contact '%s' was deleted", contact.FullName()))

	return nil
}

// AddRelationship adds a relationship between a contact and an entity
func (s *ContactService) AddRelationship(ctx context.Context, contactID uuid.UUID, req *domain.AddContactRelationshipRequest) (*domain.ContactRelationshipDTO, error) {
	// Validate entity type
	if err := validateEntityType(req.EntityType); err != nil {
		return nil, err
	}

	// Validate contact exists
	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Validate entity exists
	if err := s.validateEntityExists(ctx, req.EntityType, req.EntityID); err != nil {
		return nil, err
	}

	// Check if relationship already exists
	existingRel, err := s.contactRepo.GetRelationshipByContactAndEntity(ctx, contactID, req.EntityType, req.EntityID)
	if err == nil && existingRel != nil {
		return nil, fmt.Errorf("relationship between contact and %s already exists", req.EntityType)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing relationship: %w", err)
	}

	// If this is marked as primary, unset other primary contacts for this entity
	if req.IsPrimary {
		if err := s.contactRepo.SetPrimaryContactForEntity(ctx, req.EntityType, req.EntityID, contactID); err != nil {
			s.logger.Warn("failed to unset existing primary contact",
				zap.String("entity_type", string(req.EntityType)),
				zap.String("entity_id", req.EntityID.String()),
				zap.Error(err))
		}
	}

	// Create relationship
	relationship := &domain.ContactRelationship{
		ContactID:  contactID,
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		Role:       req.Role,
		IsPrimary:  req.IsPrimary,
	}

	if err := s.contactRepo.AddRelationship(ctx, relationship); err != nil {
		return nil, fmt.Errorf("failed to add relationship: %w", err)
	}

	// Log activity
	s.logActivity(ctx, contactID, "Contact relationship added",
		fmt.Sprintf("Contact '%s' was linked to %s", contact.FullName(), req.EntityType))

	dto := mapper.ToContactRelationshipDTO(relationship)
	return &dto, nil
}

// RemoveRelationship removes a relationship by its ID
func (s *ContactService) RemoveRelationship(ctx context.Context, relationshipID uuid.UUID) error {
	// Get relationship for logging
	relationship, err := s.contactRepo.GetRelationshipByID(ctx, relationshipID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("relationship not found")
		}
		return fmt.Errorf("failed to get relationship: %w", err)
	}

	if err := s.contactRepo.DeleteRelationshipByID(ctx, relationshipID); err != nil {
		return fmt.Errorf("failed to remove relationship: %w", err)
	}

	// Log activity
	contactName := "Unknown"
	if relationship.Contact != nil {
		contactName = relationship.Contact.FullName()
	}
	s.logActivity(ctx, relationship.ContactID, "Contact relationship removed",
		fmt.Sprintf("Contact '%s' was unlinked from %s", contactName, relationship.EntityType))

	return nil
}

// SetPrimaryContact sets a contact as the primary for a specific entity
func (s *ContactService) SetPrimaryContact(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID, contactID uuid.UUID) error {
	// Validate entity type
	if err := validateEntityType(entityType); err != nil {
		return err
	}

	// Validate entity exists
	if err := s.validateEntityExists(ctx, entityType, entityID); err != nil {
		return err
	}

	// Validate contact exists
	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("contact not found")
		}
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Validate contact is related to entity
	_, err = s.contactRepo.GetRelationshipByContactAndEntity(ctx, contactID, entityType, entityID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("contact is not related to this %s", entityType)
		}
		return fmt.Errorf("failed to check relationship: %w", err)
	}

	// Set as primary (this will unset other primary contacts for this entity)
	if err := s.contactRepo.SetPrimaryContactForEntity(ctx, entityType, entityID, contactID); err != nil {
		return fmt.Errorf("failed to set primary contact: %w", err)
	}

	// Log activity
	s.logActivity(ctx, contactID, "Primary contact changed",
		fmt.Sprintf("Contact '%s' was set as primary for %s", contact.FullName(), entityType))

	return nil
}

// GetContactsForEntity retrieves all contacts for a specific entity
func (s *ContactService) GetContactsForEntity(ctx context.Context, entityType domain.ContactEntityType, entityID uuid.UUID) ([]domain.ContactDTO, error) {
	// Validate entity type
	if err := validateEntityType(entityType); err != nil {
		return nil, err
	}

	contacts, err := s.contactRepo.ListByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts for entity: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.ToContactDTO(&contact)
	}

	return dtos, nil
}

// Search searches contacts by name or email
func (s *ContactService) Search(ctx context.Context, query string, limit int) ([]domain.ContactDTO, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	contacts, err := s.contactRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}

	dtos := make([]domain.ContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.ToContactDTO(&contact)
	}

	return dtos, nil
}
