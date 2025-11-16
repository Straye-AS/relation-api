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

func (s *CustomerService) Create(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.CustomerDTO, error) {
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

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTypeCustomer,
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
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Calculate total value and active offers
	totalValue := 0.0
	activeOffers := 0
	// TODO: Implement actual calculation from offers

	dto := mapper.ToCustomerDTO(customer, totalValue, activeOffers)
	return &dto, nil
}

func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateCustomerRequest) (*domain.CustomerDTO, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
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
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTypeCustomer,
			TargetID:    customer.ID,
			Title:       "Customer updated",
			Body:        fmt.Sprintf("Customer '%s' was updated", customer.Name),
			CreatorName: userCtx.DisplayName,
		}
		s.activityRepo.Create(ctx, activity)
	}

	// Calculate metrics
	totalValue := 0.0
	activeOffers := 0
	// TODO: Implement actual calculation

	dto := mapper.ToCustomerDTO(customer, totalValue, activeOffers)
	return &dto, nil
}

func (s *CustomerService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.customerRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	return nil
}

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

	customers, total, err := s.customerRepo.List(ctx, page, pageSize, search)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	dtos := make([]domain.CustomerDTO, len(customers))
	for i, customer := range customers {
		// TODO: Calculate actual total value and active offers
		totalValue := 0.0
		activeOffers := 0
		dtos[i] = mapper.ToCustomerDTO(&customer, totalValue, activeOffers)
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
