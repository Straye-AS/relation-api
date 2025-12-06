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

// ErrCustomerNotFound is returned when a customer is not found
var ErrCustomerNotFound = errors.New("customer not found")

// ErrDuplicateOrgNumber is returned when trying to create a customer with an existing org number
var ErrDuplicateOrgNumber = errors.New("customer with this organization number already exists")

// ErrCustomerHasActiveDependencies is returned when trying to delete a customer with active relations
var ErrCustomerHasActiveDependencies = errors.New("cannot delete customer with active projects, deals, or offers")

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

	dto := mapper.ToCustomerDTO(customer, stats.TotalValue, stats.ActiveOffers)
	return &dto, nil
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
