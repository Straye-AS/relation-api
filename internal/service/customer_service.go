package service

import (
	"context"
	"errors"
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

// Validation patterns
var (
	// Norwegian org numbers are 9 digits
	orgNumberRegex = regexp.MustCompile(`^\d{9}$`)
	// Standard email validation
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	// Norwegian phone number validation (flexible format)
	phoneRegex = regexp.MustCompile(`^(\+47)?[\s]?[0-9\s-]{8,}$`)
)

// Validation errors
var (
	ErrInvalidOrgNumber      = errors.New("invalid organization number: must be 9 digits")
	ErrDuplicateOrgNumber    = errors.New("organization number already exists")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidPhone          = errors.New("invalid phone format")
	ErrCustomerHasActiveDeps = errors.New("customer has active deals or projects and cannot be deleted")
	ErrCustomerNotFound      = errors.New("customer not found")
)

type CustomerService struct {
	customerRepo *repository.CustomerRepository
	activityRepo *repository.ActivityRepository
	logger       *zap.Logger
}

func NewCustomerService(
	customerRepo *repository.CustomerRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

// validateOrgNumber validates Norwegian organization number format (9 digits)
func validateOrgNumber(orgNumber string) error {
	orgNumber = strings.TrimSpace(orgNumber)
	if !orgNumberRegex.MatchString(orgNumber) {
		return ErrInvalidOrgNumber
	}
	return nil
}

// validateEmail validates email format
func validateEmail(email string) error {
	if email == "" {
		return nil // Email is optional
	}
	email = strings.TrimSpace(email)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// validatePhone validates phone number format
func validatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}
	// Remove common formatting characters for validation
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	if len(cleanPhone) < 8 {
		return ErrInvalidPhone
	}
	return nil
}

// Create creates a new customer with validation
func (s *CustomerService) Create(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.CustomerDTO, error) {
	// Validate org number format
	if err := validateOrgNumber(req.OrgNumber); err != nil {
		s.logger.Warn("Invalid org number format",
			zap.String("orgNumber", req.OrgNumber),
			zap.Error(err))
		return nil, err
	}

	// Check org number uniqueness
	existingCustomer, err := s.customerRepo.GetByOrgNumber(ctx, req.OrgNumber)
	if err == nil && existingCustomer != nil {
		s.logger.Warn("Duplicate org number",
			zap.String("orgNumber", req.OrgNumber),
			zap.String("existingCustomerID", existingCustomer.ID.String()))
		return nil, ErrDuplicateOrgNumber
	}
	// If error is "record not found", that's expected and we can proceed
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check org number uniqueness: %w", err)
	}

	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		s.logger.Warn("Invalid email format",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, err
	}

	// Validate contact email format if provided
	if err := validateEmail(req.ContactEmail); err != nil {
		s.logger.Warn("Invalid contact email format",
			zap.String("contactEmail", req.ContactEmail),
			zap.Error(err))
		return nil, fmt.Errorf("invalid contact email: %w", err)
	}

	// Validate phone format
	if err := validatePhone(req.Phone); err != nil {
		s.logger.Warn("Invalid phone format",
			zap.String("phone", req.Phone),
			zap.Error(err))
		return nil, err
	}

	customer := &domain.Customer{
		Name:          req.Name,
		OrgNumber:     req.OrgNumber,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		City:          req.City,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		ContactPerson: req.ContactPerson,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Create activity log
	s.logActivity(ctx, customer.ID, "Customer created",
		fmt.Sprintf("Customer '%s' was created", customer.Name))

	s.logger.Info("Customer created",
		zap.String("customerID", customer.ID.String()),
		zap.String("name", customer.Name),
		zap.String("orgNumber", customer.OrgNumber))

	dto := mapper.ToCustomerDTO(customer, 0.0, 0)
	return &dto, nil
}

// GetByID retrieves a customer by ID with calculated metrics
func (s *CustomerService) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomerDTO, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get stats for the customer
	stats, err := s.customerRepo.GetCustomerStats(ctx, id)
	if err != nil {
		s.logger.Warn("Failed to get customer stats",
			zap.String("customerID", id.String()),
			zap.Error(err))
		// Continue with zero stats if we can't get them
		stats = &repository.CustomerStats{}
	}

	dto := mapper.ToCustomerDTO(customer, stats.TotalDealValue, int(stats.ActiveDealsCount))
	return &dto, nil
}

// GetWithStats retrieves a customer with full statistics
func (s *CustomerService) GetWithStats(ctx context.Context, id uuid.UUID) (*domain.CustomerWithStatsResponse, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	stats, err := s.customerRepo.GetCustomerStats(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer stats: %w", err)
	}

	customerDTO := mapper.ToCustomerDTO(customer, stats.TotalDealValue, int(stats.ActiveDealsCount))

	return &domain.CustomerWithStatsResponse{
		CustomerDTO: customerDTO,
		Stats: domain.CustomerStatsResponse{
			ActiveDealsCount:    stats.ActiveDealsCount,
			TotalDealsCount:     stats.TotalDealsCount,
			TotalDealValue:      stats.TotalDealValue,
			WonDealsValue:       stats.WonDealsValue,
			ActiveProjectsCount: stats.ActiveProjectsCount,
			TotalProjectsCount:  stats.TotalProjectsCount,
		},
	}, nil
}

// Update updates an existing customer with validation
func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateCustomerRequest) (*domain.CustomerDTO, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// If org number is changing, validate it
	if req.OrgNumber != customer.OrgNumber {
		if err := validateOrgNumber(req.OrgNumber); err != nil {
			return nil, err
		}

		// Check uniqueness of new org number
		existingCustomer, err := s.customerRepo.GetByOrgNumber(ctx, req.OrgNumber)
		if err == nil && existingCustomer != nil && existingCustomer.ID != id {
			s.logger.Warn("Duplicate org number on update",
				zap.String("orgNumber", req.OrgNumber),
				zap.String("existingCustomerID", existingCustomer.ID.String()))
			return nil, ErrDuplicateOrgNumber
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check org number uniqueness: %w", err)
		}
	}

	// Validate email if provided
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate contact email if provided
	if err := validateEmail(req.ContactEmail); err != nil {
		return nil, fmt.Errorf("invalid contact email: %w", err)
	}

	// Validate phone if provided
	if err := validatePhone(req.Phone); err != nil {
		return nil, err
	}

	// Track what changed for activity log
	changes := []string{}
	if customer.Name != req.Name {
		changes = append(changes, fmt.Sprintf("name: '%s' -> '%s'", customer.Name, req.Name))
	}
	if customer.OrgNumber != req.OrgNumber {
		changes = append(changes, fmt.Sprintf("orgNumber: '%s' -> '%s'", customer.OrgNumber, req.OrgNumber))
	}
	if customer.Email != req.Email {
		changes = append(changes, "email updated")
	}
	if customer.Phone != req.Phone {
		changes = append(changes, "phone updated")
	}

	// Update customer fields
	customer.Name = req.Name
	customer.OrgNumber = req.OrgNumber
	customer.Email = req.Email
	customer.Phone = req.Phone
	customer.Address = req.Address
	customer.City = req.City
	customer.PostalCode = req.PostalCode
	customer.Country = req.Country
	customer.ContactPerson = req.ContactPerson
	customer.ContactEmail = req.ContactEmail
	customer.ContactPhone = req.ContactPhone

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Create activity log with changes summary
	changesSummary := "Customer updated"
	if len(changes) > 0 {
		changesSummary = fmt.Sprintf("Customer '%s' was updated: %s", customer.Name, strings.Join(changes, ", "))
	} else {
		changesSummary = fmt.Sprintf("Customer '%s' was updated", customer.Name)
	}
	s.logActivity(ctx, customer.ID, "Customer updated", changesSummary)

	s.logger.Info("Customer updated",
		zap.String("customerID", customer.ID.String()),
		zap.String("name", customer.Name))

	// Get updated stats
	stats, err := s.customerRepo.GetCustomerStats(ctx, id)
	if err != nil {
		s.logger.Warn("Failed to get customer stats after update",
			zap.String("customerID", id.String()),
			zap.Error(err))
		stats = &repository.CustomerStats{}
	}

	dto := mapper.ToCustomerDTO(customer, stats.TotalDealValue, int(stats.ActiveDealsCount))
	return &dto, nil
}

// Delete performs a soft delete of a customer after checking for active dependencies
func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID) error {
	// First check if customer exists
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Check for active deals or projects
	hasActive, reason, err := s.customerRepo.HasActiveDealsOrProjects(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check customer dependencies: %w", err)
	}
	if hasActive {
		s.logger.Warn("Cannot delete customer with active dependencies",
			zap.String("customerID", id.String()),
			zap.String("reason", reason))
		return fmt.Errorf("%w: %s", ErrCustomerHasActiveDeps, reason)
	}

	// Log activity before deletion
	s.logActivity(ctx, id, "Customer deleted",
		fmt.Sprintf("Customer '%s' was deleted", customer.Name))

	if err := s.customerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	s.logger.Info("Customer deleted",
		zap.String("customerID", id.String()),
		zap.String("name", customer.Name))

	return nil
}

// List returns a paginated list of customers
func (s *CustomerService) List(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResponse, error) {
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

	// Build filters from search parameter
	var filters *repository.CustomerFilters
	if search != "" {
		filters = &repository.CustomerFilters{
			SearchQuery: &search,
		}
	}

	customers, total, err := s.customerRepo.List(ctx, page, pageSize, filters, repository.CustomerSortByCreatedDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	dtos := make([]domain.CustomerDTO, len(customers))
	for i, customer := range customers {
		// Get stats for each customer
		stats, err := s.customerRepo.GetCustomerStats(ctx, customer.ID)
		if err != nil {
			s.logger.Warn("Failed to get customer stats in list",
				zap.String("customerID", customer.ID.String()),
				zap.Error(err))
			stats = &repository.CustomerStats{}
		}
		dtos[i] = mapper.ToCustomerDTO(&customer, stats.TotalDealValue, int(stats.ActiveDealsCount))
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

// logActivity is a helper to create activity entries
func (s *CustomerService) logActivity(ctx context.Context, customerID uuid.UUID, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		s.logger.Warn("No user context for activity log",
			zap.String("customerID", customerID.String()))
		return
	}

	activity := &domain.Activity{
		TargetType:  domain.ActivityTargetCustomer,
		TargetID:    customerID,
		Title:       title,
		Body:        body,
		CreatorName: userCtx.DisplayName,
		CreatorID:   userCtx.UserID.String(),
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Error("Failed to create activity log",
			zap.String("customerID", customerID.String()),
			zap.Error(err))
	}
}
