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

// InquiryService handles business logic for inquiries (draft offers)
type InquiryService struct {
	offerRepo      *repository.OfferRepository
	customerRepo   *repository.CustomerRepository
	activityRepo   *repository.ActivityRepository
	userRepo       *repository.UserRepository
	companyService *CompanyService
	logger         *zap.Logger
	db             *gorm.DB
}

// NewInquiryService creates a new InquiryService
func NewInquiryService(
	offerRepo *repository.OfferRepository,
	customerRepo *repository.CustomerRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
	companyService *CompanyService,
	logger *zap.Logger,
	db *gorm.DB,
) *InquiryService {
	return &InquiryService{
		offerRepo:      offerRepo,
		customerRepo:   customerRepo,
		activityRepo:   activityRepo,
		userRepo:       userRepo,
		companyService: companyService,
		logger:         logger,
		db:             db,
	}
}

// Create creates a new inquiry (offer in draft phase)
func (s *InquiryService) Create(ctx context.Context, req *domain.CreateInquiryRequest) (*domain.OfferDTO, error) {
	var customerID *uuid.UUID
	var customerName string
	companyID := domain.CompanyGruppen

	// If customerId is provided, verify customer exists
	if req.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *req.CustomerID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCustomerNotFound
			}
			return nil, fmt.Errorf("failed to verify customer: %w", err)
		}
		customerID = req.CustomerID
		customerName = customer.Name

		// Use customer's company if available
		if customer.CompanyID != nil && *customer.CompanyID != "" {
			companyID = domain.CompanyID(*customer.CompanyID)
		}
	}

	// Create inquiry (offer in draft phase) with minimal required fields
	inquiry := &domain.Offer{
		Title:        req.Title,
		CustomerID:   customerID,
		CustomerName: customerName,
		CompanyID:    companyID,
		Phase:        domain.OfferPhaseDraft, // Always draft for inquiries
		Probability:  0,
		Value:        0,
		Status:       domain.OfferStatusActive,
		Description:  req.Description,
		DueDate:      req.DueDate,
		// ResponsibleUserID and OfferNumber are NOT set - will be set on conversion
	}

	// Set responsible if provided - resolve email to user ID if possible
	if req.Responsible != "" {
		if resolvedID := s.resolveResponsible(ctx, req.Responsible); resolvedID != "" {
			inquiry.ResponsibleUserID = resolvedID
		}
	}

	if err := s.offerRepo.Create(ctx, inquiry); err != nil {
		return nil, fmt.Errorf("failed to create inquiry: %w", err)
	}

	// Reload with relations
	inquiry, err := s.offerRepo.GetByID(ctx, inquiry.ID)
	if err != nil {
		s.logger.Warn("failed to reload inquiry after create", zap.Error(err))
	}

	// Log activity
	activityMessage := fmt.Sprintf("Inquiry '%s' was created", inquiry.Title)
	if customerName != "" {
		activityMessage = fmt.Sprintf("Inquiry '%s' was created for customer %s", inquiry.Title, customerName)
	}
	s.logActivity(ctx, inquiry.ID, "Inquiry created", activityMessage)

	dto := mapper.ToOfferDTO(inquiry)
	return &dto, nil
}

// GetByID retrieves an inquiry by ID
func (s *InquiryService) GetByID(ctx context.Context, id uuid.UUID) (*domain.OfferDTO, error) {
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInquiryNotFound
		}
		return nil, fmt.Errorf("failed to get inquiry: %w", err)
	}

	// Verify it's actually an inquiry (draft phase)
	if offer.Phase != domain.OfferPhaseDraft {
		return nil, ErrNotAnInquiry
	}

	dto := mapper.ToOfferDTO(offer)
	return &dto, nil
}

// List returns a paginated list of inquiries (draft offers)
func (s *InquiryService) List(ctx context.Context, page, pageSize int, customerID *uuid.UUID) (*domain.PaginatedResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	inquiries, total, err := s.offerRepo.ListInquiries(ctx, page, pageSize, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list inquiries: %w", err)
	}

	dtos := make([]domain.OfferDTO, len(inquiries))
	for i, inquiry := range inquiries {
		dtos[i] = mapper.ToOfferDTO(&inquiry)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &domain.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Delete removes an inquiry
func (s *InquiryService) Delete(ctx context.Context, id uuid.UUID) error {
	inquiry, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInquiryNotFound
		}
		return fmt.Errorf("failed to get inquiry: %w", err)
	}

	// Verify it's actually an inquiry (draft phase)
	if inquiry.Phase != domain.OfferPhaseDraft {
		return ErrNotAnInquiry
	}

	if err := s.offerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete inquiry: %w", err)
	}

	// Log activity on customer since inquiry is deleted (only if customer exists)
	if inquiry.CustomerID != nil {
		s.logActivityOnTarget(ctx, domain.ActivityTargetCustomer, *inquiry.CustomerID,
			"Inquiry deleted", fmt.Sprintf("Inquiry '%s' was deleted", inquiry.Title))
	}

	return nil
}

// Convert converts an inquiry to an offer (phase=in_progress)
// Logic:
// - responsibleUserId only: infer company from user's department/companyId
// - companyId only: use company's defaultOfferResponsibleId
// - both provided: use both directly
func (s *InquiryService) Convert(ctx context.Context, id uuid.UUID, req *domain.ConvertInquiryRequest) (*domain.ConvertInquiryResponse, error) {
	inquiry, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInquiryNotFound
		}
		return nil, fmt.Errorf("failed to get inquiry: %w", err)
	}

	// Verify it's actually an inquiry (draft phase)
	if inquiry.Phase != domain.OfferPhaseDraft {
		return nil, ErrNotAnInquiry
	}

	var responsibleUserID string
	var companyID domain.CompanyID

	// Determine responsible user and company
	if req.ResponsibleUserID != nil && *req.ResponsibleUserID != "" {
		responsibleUserID = *req.ResponsibleUserID
		// If company not provided, use the existing company or default
		if req.CompanyID != nil && *req.CompanyID != "" {
			companyID = *req.CompanyID
		} else {
			companyID = inquiry.CompanyID
		}
	} else if req.CompanyID != nil && *req.CompanyID != "" {
		companyID = *req.CompanyID
		// Get default responsible user from company
		if s.companyService != nil {
			defaultResponsible := s.companyService.GetDefaultOfferResponsible(ctx, companyID)
			if defaultResponsible != nil && *defaultResponsible != "" {
				responsibleUserID = *defaultResponsible
			}
		}
	} else {
		// Neither provided - try to use existing company's default
		companyID = inquiry.CompanyID
		if s.companyService != nil {
			defaultResponsible := s.companyService.GetDefaultOfferResponsible(ctx, companyID)
			if defaultResponsible != nil && *defaultResponsible != "" {
				responsibleUserID = *defaultResponsible
			}
		}
	}

	// Ensure we have a responsible user
	if responsibleUserID == "" {
		return nil, ErrInquiryMissingConversionData
	}

	// Validate company ID for offer number generation
	if !domain.IsValidCompanyID(string(companyID)) {
		return nil, ErrInvalidCompanyID
	}

	// Generate offer number for the company
	offerNumber, err := s.offerRepo.GenerateOfferNumber(ctx, companyID)
	if err != nil {
		s.logger.Error("failed to generate offer number during conversion",
			zap.Error(err),
			zap.String("inquiryID", id.String()),
			zap.String("companyID", string(companyID)))
		return nil, fmt.Errorf("%w: %v", ErrOfferNumberGenerationFailed, err)
	}

	s.logger.Info("generated offer number during inquiry conversion",
		zap.String("inquiryID", id.String()),
		zap.String("offerNumber", offerNumber))

	// Update inquiry -> offer
	updates := map[string]interface{}{
		"phase":               domain.OfferPhaseInProgress,
		"company_id":          companyID,
		"responsible_user_id": responsibleUserID,
		"offer_number":        offerNumber,
	}

	if err := s.offerRepo.UpdateFields(ctx, id, updates); err != nil {
		return nil, fmt.Errorf("failed to convert inquiry: %w", err)
	}

	// Reload the offer
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload offer after conversion: %w", err)
	}

	// Log activity
	s.logActivity(ctx, offer.ID, "Inquiry converted to offer",
		fmt.Sprintf("Inquiry '%s' was converted to offer %s (responsible: %s)",
			offer.Title, offerNumber, responsibleUserID))

	dto := mapper.ToOfferDTO(offer)
	return &domain.ConvertInquiryResponse{
		Offer:       &dto,
		OfferNumber: offerNumber,
	}, nil
}

// resolveResponsible resolves a responsible identifier to a user ID
// If the input looks like an email (contains @), it tries to find the user by email
// and returns their ID. Returns empty string if user cannot be found.
// If input doesn't contain @, it's assumed to be a valid user ID and returned as-is.
func (s *InquiryService) resolveResponsible(ctx context.Context, responsible string) string {
	// If it doesn't look like an email, return as-is (assume it's already a user ID)
	if !strings.Contains(responsible, "@") {
		return responsible
	}

	// Try to find user by email
	if s.userRepo != nil {
		user, err := s.userRepo.GetByEmail(ctx, responsible)
		if err == nil && user != nil {
			s.logger.Debug("resolved email to user ID",
				zap.String("email", responsible),
				zap.String("userId", user.ID))
			return user.ID
		}
		// Log if lookup failed (but don't fail the request)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("failed to lookup user by email",
				zap.String("email", responsible),
				zap.Error(err))
		}
	}

	// User not found - return empty (will not be stored)
	s.logger.Debug("could not resolve email to user ID, responsible will be null",
		zap.String("email", responsible))
	return ""
}

// logActivity creates an activity log entry for an offer
func (s *InquiryService) logActivity(ctx context.Context, offerID uuid.UUID, title, body string) {
	s.logActivityOnTarget(ctx, domain.ActivityTargetOffer, offerID, title, body)
}

// logActivityOnTarget creates an activity log entry for any target
func (s *InquiryService) logActivityOnTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		s.logger.Warn("no user context for activity logging")
		return
	}

	activity := &domain.Activity{
		TargetType:  targetType,
		TargetID:    targetID,
		Title:       title,
		Body:        body,
		CreatorName: userCtx.DisplayName,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log activity", zap.Error(err))
	}
}
