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

// ErrSupplierNotFound is returned when a supplier is not found
var ErrSupplierNotFound = errors.New("supplier not found")

// ErrDuplicateSupplierOrgNumber is returned when trying to create a supplier with an existing org number
var ErrDuplicateSupplierOrgNumber = errors.New("supplier with this organization number already exists")

// ErrSupplierHasActiveRelations is returned when trying to delete a supplier with active offer relationships
var ErrSupplierHasActiveRelations = errors.New("cannot delete supplier with active offer relationships")

// SupplierService handles business logic for suppliers
type SupplierService struct {
	supplierRepo *repository.SupplierRepository
	activityRepo *repository.ActivityRepository
	logger       *zap.Logger
}

// NewSupplierService creates a new supplier service instance
func NewSupplierService(
	supplierRepo *repository.SupplierRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *SupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
		activityRepo: activityRepo,
		logger:       logger,
	}
}

// Create creates a new supplier
func (s *SupplierService) Create(ctx context.Context, req *domain.CreateSupplierRequest) (*domain.SupplierDTO, error) {
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
		existing, err := s.supplierRepo.GetByOrgNumber(ctx, req.OrgNumber)
		if err == nil && existing != nil {
			return nil, ErrDuplicateSupplierOrgNumber
		}
		// Ignore not found errors
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check org number: %w", err)
		}
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = domain.SupplierStatusActive
	}

	supplier := &domain.Supplier{
		Name:          req.Name,
		OrgNumber:     req.OrgNumber,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		City:          req.City,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		Municipality:  req.Municipality,
		County:        req.County,
		ContactPerson: req.ContactPerson,
		ContactEmail:  req.ContactEmail,
		ContactPhone:  req.ContactPhone,
		Status:        status,
		Category:      req.Category,
		Notes:         req.Notes,
		PaymentTerms:  req.PaymentTerms,
		Website:       req.Website,
	}

	// Set user tracking fields on creation
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.CreatedByID = userCtx.UserID.String()
		supplier.CreatedByName = userCtx.DisplayName
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		// Check for unique constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrDuplicateSupplierOrgNumber
		}
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	// Log activity
	s.logActivity(ctx, supplier.ID, supplier.Name, "Leverandor opprettet", fmt.Sprintf("Leverandoren '%s' ble opprettet", supplier.Name))

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// GetByID retrieves a supplier by ID
func (s *SupplierService) GetByID(ctx context.Context, id uuid.UUID) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// GetByIDWithDetails retrieves a supplier with full details including stats, contacts, and recent offers
func (s *SupplierService) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.SupplierWithDetailsDTO, error) {
	supplier, err := s.supplierRepo.GetWithContacts(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Get supplier stats
	repoStats, err := s.supplierRepo.GetSupplierStats(ctx, id)
	if err != nil {
		s.logger.Warn("failed to get supplier stats", zap.Error(err))
		repoStats = &repository.SupplierStats{}
	}

	// Convert to mapper input type
	stats := &mapper.SupplierStatsInput{
		TotalOffers:     repoStats.TotalOffers,
		ActiveOffers:    repoStats.ActiveOffers,
		CompletedOffers: repoStats.CompletedOffers,
		TotalProjects:   repoStats.TotalProjects,
	}

	// Get recent offer-supplier relationships
	recentOffers, err := s.supplierRepo.GetRecentOfferSuppliers(ctx, id, 5)
	if err != nil {
		s.logger.Warn("failed to get recent offers for supplier", zap.Error(err))
		recentOffers = []domain.OfferSupplier{}
	}

	dto := mapper.SupplierToWithDetailsDTO(supplier, stats, recentOffers)
	return &dto, nil
}

// Update updates an existing supplier
func (s *SupplierService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateSupplierRequest) (*domain.SupplierDTO, error) {
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

	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Check for duplicate org number if it's being changed
	if req.OrgNumber != "" && req.OrgNumber != supplier.OrgNumber {
		existing, err := s.supplierRepo.GetByOrgNumber(ctx, req.OrgNumber)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrDuplicateSupplierOrgNumber
		}
	}

	supplier.Name = req.Name
	supplier.OrgNumber = req.OrgNumber
	supplier.Email = req.Email
	supplier.Phone = req.Phone
	supplier.Address = req.Address
	supplier.City = req.City
	supplier.PostalCode = req.PostalCode
	supplier.Country = req.Country
	supplier.Municipality = req.Municipality
	supplier.County = req.County
	supplier.ContactPerson = req.ContactPerson
	supplier.ContactEmail = req.ContactEmail
	supplier.ContactPhone = req.ContactPhone
	supplier.Category = req.Category
	supplier.Notes = req.Notes
	supplier.PaymentTerms = req.PaymentTerms
	supplier.Website = req.Website

	// Update status if provided, keep existing if empty
	if req.Status != "" {
		supplier.Status = req.Status
	}

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, ErrDuplicateSupplierOrgNumber
		}
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	// Log activity
	s.logActivity(ctx, supplier.ID, supplier.Name, "Leverandor oppdatert", fmt.Sprintf("Leverandoren '%s' ble oppdatert", supplier.Name))

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// Delete performs a soft delete on a supplier
func (s *SupplierService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSupplierNotFound
		}
		return fmt.Errorf("failed to get supplier: %w", err)
	}

	// Check for active relationships
	hasActive, reason, err := s.supplierRepo.HasActiveRelations(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check supplier relations: %w", err)
	}
	if hasActive {
		return fmt.Errorf("%w: %s", ErrSupplierHasActiveRelations, reason)
	}

	if err := s.supplierRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	// Log activity
	s.logActivity(ctx, id, supplier.Name, "Leverandor slettet", fmt.Sprintf("Leverandoren '%s' ble slettet", supplier.Name))

	return nil
}

// List returns a paginated list of suppliers with default filters
func (s *SupplierService) List(ctx context.Context, page, pageSize int, search string) (*domain.PaginatedResponse, error) {
	filters := &repository.SupplierFilters{Search: search}
	return s.ListWithSort(ctx, page, pageSize, filters, repository.DefaultSortConfig())
}

// ListWithSort returns a paginated list of suppliers with filter and sort options
func (s *SupplierService) ListWithSort(ctx context.Context, page, pageSize int, filters *repository.SupplierFilters, sort repository.SortConfig) (*domain.PaginatedResponse, error) {
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

	suppliers, total, err := s.supplierRepo.ListWithSortConfig(ctx, page, pageSize, filters, sort)
	if err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}

	dtos := mapper.SuppliersToDTO(suppliers)

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &domain.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateStatus updates only the supplier status
func (s *SupplierService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SupplierStatus) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	supplier.Status = status

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier status: %w", err)
	}

	s.logActivity(ctx, supplier.ID, supplier.Name, "Status oppdatert", fmt.Sprintf("Leverandorstatus endret til '%s'", status))

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdateNotes updates only the supplier notes
func (s *SupplierService) UpdateNotes(ctx context.Context, id uuid.UUID, notes string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	supplier.Notes = notes

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier notes: %w", err)
	}

	s.logActivity(ctx, supplier.ID, supplier.Name, "Notater oppdatert", "Leverandorens notater ble oppdatert")

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdateCategory updates only the supplier category
func (s *SupplierService) UpdateCategory(ctx context.Context, id uuid.UUID, category string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	supplier.Category = category

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier category: %w", err)
	}

	s.logActivity(ctx, supplier.ID, supplier.Name, "Kategori oppdatert", fmt.Sprintf("Leverandorkategori endret til '%s'", category))

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdatePaymentTerms updates only the supplier payment terms
func (s *SupplierService) UpdatePaymentTerms(ctx context.Context, id uuid.UUID, paymentTerms string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	supplier.PaymentTerms = paymentTerms

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier payment terms: %w", err)
	}

	s.logActivity(ctx, supplier.ID, supplier.Name, "Betalingsbetingelser oppdatert", fmt.Sprintf("Leverandorens betalingsbetingelser endret til '%s'", paymentTerms))

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// logActivity is a helper to log supplier activities
func (s *SupplierService) logActivity(ctx context.Context, supplierID uuid.UUID, supplierName, title, body string) {
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetSupplier,
			TargetID:    supplierID,
			TargetName:  supplierName,
			Title:       title,
			Body:        body,
			CreatorName: userCtx.DisplayName,
		}
		_ = s.activityRepo.Create(ctx, activity)
	}
}
