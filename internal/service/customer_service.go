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

// ErrCustomerNotFound is returned when a customer is not found
var ErrCustomerNotFound = errors.New("customer not found")

// ErrDuplicateOrgNumber is returned when trying to create a customer with an existing org number
var ErrDuplicateOrgNumber = errors.New("customer with this organization number already exists")

// ErrCustomerHasActiveDependencies is returned when trying to delete a customer with active relations
var ErrCustomerHasActiveDependencies = errors.New("cannot delete customer with active projects, deals, or offers")

// ErrInvalidEmailFormat is returned when an email address has invalid format
var ErrInvalidEmailFormat = errors.New("invalid email format")

// ErrInvalidPhoneFormat is returned when a phone number has invalid format
var ErrInvalidPhoneFormat = errors.New("invalid phone format")

// Email and phone validation patterns
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// Norwegian phone pattern: allows +47 prefix, spaces, and common formats
	phoneRegex = regexp.MustCompile(`^(\+47\s?)?[\d\s\-]{8,15}$`)
)

// validateEmail checks if the email has a valid format
func validateEmail(email string) error {
	if email == "" {
		return nil // Empty emails are allowed (optional fields may have empty email)
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmailFormat
	}
	return nil
}

// validatePhone checks if the phone number has a valid format
func validatePhone(phone string) error {
	if phone == "" {
		return nil // Empty phones are allowed
	}
	// Remove common formatting characters for validation
	cleaned := strings.ReplaceAll(strings.ReplaceAll(phone, " ", ""), "-", "")
	if len(cleaned) < 8 || !phoneRegex.MatchString(phone) {
		return ErrInvalidPhoneFormat
	}
	return nil
}

// calculateTierFromValue determines the customer tier based on total business value
// Thresholds (in NOK):
// - Platinum: >= 10,000,000 (10M)
// - Gold: >= 1,000,000 (1M)
// - Silver: >= 100,000 (100K)
// - Bronze: < 100,000
func calculateTierFromValue(totalValue float64) domain.CustomerTier {
	switch {
	case totalValue >= 10_000_000:
		return domain.CustomerTierPlatinum
	case totalValue >= 1_000_000:
		return domain.CustomerTierGold
	case totalValue >= 100_000:
		return domain.CustomerTierSilver
	default:
		return domain.CustomerTierBronze
	}
}

type CustomerService struct {
	customerRepo *repository.CustomerRepository
	dealRepo     *repository.DealRepository
	projectRepo  *repository.ProjectRepository
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

// NewCustomerServiceWithDeps creates a CustomerService with all dependencies for full feature support
func NewCustomerServiceWithDeps(
	customerRepo *repository.CustomerRepository,
	dealRepo *repository.DealRepository,
	projectRepo *repository.ProjectRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
		dealRepo:     dealRepo,
		projectRepo:  projectRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

func (s *CustomerService) Create(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.CustomerDTO, error) {
	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%w: email", ErrInvalidEmailFormat)
	}
	if err := validateEmail(req.ContactEmail); err != nil {
		return nil, fmt.Errorf("%w: contactEmail", ErrInvalidEmailFormat)
	}

	// Validate phone format
	if err := validatePhone(req.Phone); err != nil {
		return nil, fmt.Errorf("%w: phone", ErrInvalidPhoneFormat)
	}
	if err := validatePhone(req.ContactPhone); err != nil {
		return nil, fmt.Errorf("%w: contactPhone", ErrInvalidPhoneFormat)
	}

	// Check for duplicate org number
	if req.OrgNumber != "" {
		existing, err := s.customerRepo.GetByOrgNumber(ctx, req.OrgNumber)
		if err == nil && existing != nil {
			return nil, ErrDuplicateOrgNumber
		}
		// Ignore not found errors
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check org number: %w", err)
		}
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = domain.CustomerStatusActive
	}

	// Set default tier if not provided
	tier := req.Tier
	if tier == "" {
		tier = domain.CustomerTierBronze
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
		Status:        status,
		Tier:          tier,
		Industry:      req.Industry,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrDuplicateOrgNumber
		}
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetCustomer,
			TargetID:    customer.ID,
			Title:       "Customer created",
			Body:        fmt.Sprintf("Customer '%s' was created", customer.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToCustomerDTO(customer, 0.0, 0)
	return &dto, nil
}

func (s *CustomerService) GetByID(ctx context.Context, id uuid.UUID) (*domain.CustomerDTO, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get customer stats
	stats, err := s.customerRepo.GetCustomerStats(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get customer stats", zap.Error(err))
		stats = &repository.CustomerStats{}
	}

	// Auto-update tier based on total value
	s.updateTierIfNeeded(ctx, customer, stats.TotalValue)

	dto := mapper.ToCustomerDTO(customer, stats.TotalValue, stats.ActiveOffers)
	return &dto, nil
}

// updateTierIfNeeded checks if the customer tier should be updated based on total value
// and updates it if the calculated tier is higher than the current tier
func (s *CustomerService) updateTierIfNeeded(ctx context.Context, customer *domain.Customer, totalValue float64) {
	calculatedTier := calculateTierFromValue(totalValue)

	// Only upgrade tiers, never downgrade automatically
	tierOrder := map[domain.CustomerTier]int{
		domain.CustomerTierBronze:   1,
		domain.CustomerTierSilver:   2,
		domain.CustomerTierGold:     3,
		domain.CustomerTierPlatinum: 4,
	}

	if tierOrder[calculatedTier] > tierOrder[customer.Tier] {
		customer.Tier = calculatedTier
		if err := s.customerRepo.Update(ctx, customer); err != nil {
			s.logger.Warn("failed to auto-update customer tier",
				zap.String("customerID", customer.ID.String()),
				zap.String("newTier", string(calculatedTier)),
				zap.Error(err))
		} else {
			s.logger.Info("auto-upgraded customer tier",
				zap.String("customerID", customer.ID.String()),
				zap.String("newTier", string(calculatedTier)),
				zap.Float64("totalValue", totalValue))
		}
	}
}

// GetByIDWithDetails returns a customer with full details including contacts, deals, and projects
func (s *CustomerService) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.CustomerWithDetailsDTO, error) {
	customer, err := s.customerRepo.GetCustomerWithRelations(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get customer stats
	stats, err := s.customerRepo.GetCustomerStats(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get customer stats", zap.Error(err))
		stats = &repository.CustomerStats{}
	}

	// Auto-update tier based on total value
	s.updateTierIfNeeded(ctx, customer, stats.TotalValue)

	// Build base customer DTO
	customerDTO := mapper.ToCustomerDTO(customer, stats.TotalValue, stats.ActiveOffers)

	result := &domain.CustomerWithDetailsDTO{
		CustomerDTO: customerDTO,
		Stats: &domain.CustomerStatsDTO{
			TotalValue:     stats.TotalValue,
			ActiveOffers:   stats.ActiveOffers,
			ActiveDeals:    stats.ActiveDeals,
			ActiveProjects: stats.ActiveProjects,
			TotalContacts:  stats.TotalContacts,
		},
	}

	// Map contacts
	if len(customer.Contacts) > 0 {
		result.Contacts = make([]domain.ContactDTO, len(customer.Contacts))
		for i, contact := range customer.Contacts {
			result.Contacts[i] = mapper.ToContactDTO(&contact)
		}
	}

	// Get active deals if dealRepo is available
	if s.dealRepo != nil {
		deals, _, err := s.dealRepo.List(ctx, 1, 5, &repository.DealFilters{
			CustomerID: &id,
		}, repository.DealSortByCreatedDesc)
		if err == nil {
			result.ActiveDeals = make([]domain.DealDTO, len(deals))
			for i, deal := range deals {
				result.ActiveDeals[i] = mapper.ToDealDTO(&deal)
			}
		}
	}

	// Get active projects if projectRepo is available
	if s.projectRepo != nil {
		activeStatus := domain.ProjectStatusActive
		projects, _, err := s.projectRepo.List(ctx, 1, 5, &id, &activeStatus)
		if err == nil {
			result.ActiveProjects = make([]domain.ProjectDTO, len(projects))
			for i, project := range projects {
				result.ActiveProjects[i] = mapper.ToProjectDTO(&project)
			}
		}
	}

	return result, nil
}

func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateCustomerRequest) (*domain.CustomerDTO, error) {
	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%w: email", ErrInvalidEmailFormat)
	}
	if err := validateEmail(req.ContactEmail); err != nil {
		return nil, fmt.Errorf("%w: contactEmail", ErrInvalidEmailFormat)
	}

	// Validate phone format
	if err := validatePhone(req.Phone); err != nil {
		return nil, fmt.Errorf("%w: phone", ErrInvalidPhoneFormat)
	}
	if err := validatePhone(req.ContactPhone); err != nil {
		return nil, fmt.Errorf("%w: contactPhone", ErrInvalidPhoneFormat)
	}

	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Check for duplicate org number if it's being changed
	if req.OrgNumber != customer.OrgNumber {
		existing, err := s.customerRepo.GetByOrgNumber(ctx, req.OrgNumber)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrDuplicateOrgNumber
		}
	}

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

	// Update status if provided, keep existing if empty
	if req.Status != "" {
		customer.Status = req.Status
	}

	// Update tier if provided, keep existing if empty
	if req.Tier != "" {
		customer.Tier = req.Tier
	}

	// Update industry (can be empty to clear)
	customer.Industry = req.Industry

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrDuplicateOrgNumber
		}
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetCustomer,
			TargetID:    customer.ID,
			Title:       "Customer updated",
			Body:        fmt.Sprintf("Customer '%s' was updated", customer.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	// Get customer stats
	stats, err := s.customerRepo.GetCustomerStats(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get customer stats", zap.Error(err))
		stats = &repository.CustomerStats{}
	}

	dto := mapper.ToCustomerDTO(customer, stats.TotalValue, stats.ActiveOffers)
	return &dto, nil
}

func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if customer exists
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCustomerNotFound
		}
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Check for active relations
	hasActive, reason, err := s.customerRepo.HasActiveRelations(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check customer relations: %w", err)
	}
	if hasActive {
		return fmt.Errorf("%w: %s", ErrCustomerHasActiveDependencies, reason)
	}

	if err := s.customerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetCustomer,
			TargetID:    id,
			Title:       "Customer deleted",
			Body:        fmt.Sprintf("Customer '%s' was deleted", customer.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	return nil
}

func (s *CustomerService) List(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResponse, error) {
	filters := &repository.CustomerFilters{Search: search}
	return s.ListWithFilters(ctx, page, pageSize, filters, repository.CustomerSortByCreatedDesc)
}

// ListWithFilters returns a paginated list of customers with filter and sort options
func (s *CustomerService) ListWithFilters(ctx context.Context, page, pageSize int, filters *repository.CustomerFilters, sortBy repository.CustomerSortOption) (*domain.PaginatedResponse, error) {
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

	customers, total, err := s.customerRepo.ListWithFilters(ctx, page, pageSize, filters, sortBy)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	dtos := make([]domain.CustomerDTO, len(customers))
	for i, customer := range customers {
		// Get stats for each customer (optional optimization: batch query)
		stats, err := s.customerRepo.GetCustomerStats(ctx, customer.ID)
		if err != nil {
			s.logger.Warn("failed to get customer stats", zap.String("customerID", customer.ID.String()), zap.Error(err))
			stats = &repository.CustomerStats{}
		}
		dtos[i] = mapper.ToCustomerDTO(&customer, stats.TotalValue, stats.ActiveOffers)
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

// FuzzySearchBestMatch finds the single best matching customer for a query
// Uses multiple matching strategies including exact, prefix, contains, and trigram similarity
// Returns the best match with a confidence score
// Special case: query "all" returns all customers (limited to 1000)
func (s *CustomerService) FuzzySearchBestMatch(ctx context.Context, query string) (*domain.FuzzyCustomerSearchResponse, error) {
	// Validate query length (max 200 characters)
	if len(query) > 200 {
		return nil, fmt.Errorf("query too long: maximum 200 characters allowed")
	}

	// Special case: return all customers when query is "all"
	if strings.ToLower(strings.TrimSpace(query)) == "all" {
		customers, err := s.customerRepo.GetAllMinimal(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get all customers: %w", err)
		}

		dtos := make([]domain.CustomerMinimalDTO, len(customers))
		for i, c := range customers {
			dtos[i] = domain.CustomerMinimalDTO{
				ID:   c.ID,
				Name: c.Name,
			}
		}

		return &domain.FuzzyCustomerSearchResponse{
			Customers: dtos,
			Found:     len(dtos) > 0,
		}, nil
	}

	result, err := s.customerRepo.FuzzySearchBestMatch(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search customer: %w", err)
	}

	if result == nil {
		return &domain.FuzzyCustomerSearchResponse{
			Customer:   nil,
			Confidence: 0,
			Found:      false,
		}, nil
	}

	minimalDTO := domain.CustomerMinimalDTO{
		ID:   result.Customer.ID,
		Name: result.Customer.Name,
	}
	return &domain.FuzzyCustomerSearchResponse{
		Customer:   &minimalDTO,
		Confidence: result.Similarity,
		Found:      true,
	}, nil
}
