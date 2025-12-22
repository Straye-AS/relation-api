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

	// Note: OrgNumber is immutable after creation and cannot be changed via update
	// Only update fields that are explicitly provided (non-empty values)

	if req.Name != "" {
		supplier.Name = req.Name
	}
	if req.Email != "" {
		supplier.Email = req.Email
	}
	if req.Phone != "" {
		supplier.Phone = req.Phone
	}
	if req.Address != "" {
		supplier.Address = req.Address
	}
	if req.City != "" {
		supplier.City = req.City
	}
	if req.PostalCode != "" {
		supplier.PostalCode = req.PostalCode
	}
	if req.Country != "" {
		supplier.Country = req.Country
	}
	if req.Municipality != "" {
		supplier.Municipality = req.Municipality
	}
	if req.County != "" {
		supplier.County = req.County
	}
	if req.ContactPerson != "" {
		supplier.ContactPerson = req.ContactPerson
	}
	if req.ContactEmail != "" {
		supplier.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != "" {
		supplier.ContactPhone = req.ContactPhone
	}
	if req.Category != "" {
		supplier.Category = req.Category
	}
	if req.Notes != "" {
		supplier.Notes = req.Notes
	}
	if req.PaymentTerms != "" {
		supplier.PaymentTerms = req.PaymentTerms
	}
	if req.Website != "" {
		supplier.Website = req.Website
	}
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

// UpdateEmail updates only the supplier email
func (s *SupplierService) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (*domain.SupplierDTO, error) {
	// Validate email format
	if err := validateEmail(email); err != nil {
		return nil, fmt.Errorf("%w: email", ErrInvalidEmailFormat)
	}

	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	oldEmail := supplier.Email
	supplier.Email = email

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier email: %w", err)
	}

	activityMsg := "Leverandorens e-post ble fjernet"
	if email != "" {
		if oldEmail == "" {
			activityMsg = fmt.Sprintf("Leverandorens e-post satt til '%s'", email)
		} else {
			activityMsg = fmt.Sprintf("Leverandorens e-post endret fra '%s' til '%s'", oldEmail, email)
		}
	}
	s.logActivity(ctx, supplier.ID, supplier.Name, "E-post oppdatert", activityMsg)

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdatePhone updates only the supplier phone
func (s *SupplierService) UpdatePhone(ctx context.Context, id uuid.UUID, phone string) (*domain.SupplierDTO, error) {
	// Validate phone format
	if err := validatePhone(phone); err != nil {
		return nil, fmt.Errorf("%w: phone", ErrInvalidPhoneFormat)
	}

	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	oldPhone := supplier.Phone
	supplier.Phone = phone

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier phone: %w", err)
	}

	activityMsg := "Leverandorens telefonnummer ble fjernet"
	if phone != "" {
		if oldPhone == "" {
			activityMsg = fmt.Sprintf("Leverandorens telefonnummer satt til '%s'", phone)
		} else {
			activityMsg = fmt.Sprintf("Leverandorens telefonnummer endret fra '%s' til '%s'", oldPhone, phone)
		}
	}
	s.logActivity(ctx, supplier.ID, supplier.Name, "Telefon oppdatert", activityMsg)

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdateWebsite updates only the supplier website
func (s *SupplierService) UpdateWebsite(ctx context.Context, id uuid.UUID, website string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	oldWebsite := supplier.Website
	supplier.Website = website

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier website: %w", err)
	}

	activityMsg := "Leverandorens nettside ble fjernet"
	if website != "" {
		if oldWebsite == "" {
			activityMsg = fmt.Sprintf("Leverandorens nettside satt til '%s'", website)
		} else {
			activityMsg = fmt.Sprintf("Leverandorens nettside endret fra '%s' til '%s'", oldWebsite, website)
		}
	}
	s.logActivity(ctx, supplier.ID, supplier.Name, "Nettside oppdatert", activityMsg)

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdateAddress updates only the supplier address
func (s *SupplierService) UpdateAddress(ctx context.Context, id uuid.UUID, address string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	oldAddress := supplier.Address
	supplier.Address = address

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier address: %w", err)
	}

	activityMsg := "Leverandorens adresse ble fjernet"
	if address != "" {
		if oldAddress == "" {
			activityMsg = fmt.Sprintf("Leverandorens adresse satt til '%s'", address)
		} else {
			activityMsg = fmt.Sprintf("Leverandorens adresse endret fra '%s' til '%s'", oldAddress, address)
		}
	}
	s.logActivity(ctx, supplier.ID, supplier.Name, "Adresse oppdatert", activityMsg)

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdatePostalCode updates only the supplier postal code
func (s *SupplierService) UpdatePostalCode(ctx context.Context, id uuid.UUID, postalCode string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	oldPostalCode := supplier.PostalCode
	supplier.PostalCode = postalCode

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier postal code: %w", err)
	}

	activityMsg := "Leverandorens postnummer ble fjernet"
	if postalCode != "" {
		if oldPostalCode == "" {
			activityMsg = fmt.Sprintf("Leverandorens postnummer satt til '%s'", postalCode)
		} else {
			activityMsg = fmt.Sprintf("Leverandorens postnummer endret fra '%s' til '%s'", oldPostalCode, postalCode)
		}
	}
	s.logActivity(ctx, supplier.ID, supplier.Name, "Postnummer oppdatert", activityMsg)

	dto := mapper.SupplierToDTO(supplier)
	return &dto, nil
}

// UpdateCity updates only the supplier city
func (s *SupplierService) UpdateCity(ctx context.Context, id uuid.UUID, city string) (*domain.SupplierDTO, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	oldCity := supplier.City
	supplier.City = city

	// Set updated by fields
	if userCtx, ok := auth.FromContext(ctx); ok {
		supplier.UpdatedByID = userCtx.UserID.String()
		supplier.UpdatedByName = userCtx.DisplayName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier city: %w", err)
	}

	activityMsg := "Leverandorens by ble fjernet"
	if city != "" {
		if oldCity == "" {
			activityMsg = fmt.Sprintf("Leverandorens by satt til '%s'", city)
		} else {
			activityMsg = fmt.Sprintf("Leverandorens by endret fra '%s' til '%s'", oldCity, city)
		}
	}
	s.logActivity(ctx, supplier.ID, supplier.Name, "By oppdatert", activityMsg)

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

// ListOffers returns a paginated list of offers linked to a supplier
func (s *SupplierService) ListOffers(ctx context.Context, supplierID uuid.UUID, page, pageSize int, phase *domain.OfferPhase, sort repository.SortConfig) (*domain.PaginatedResponse, error) {
	// Verify supplier exists
	_, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

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

	offers, total, err := s.supplierRepo.GetOffersBySupplier(ctx, supplierID, page, pageSize, phase, sort)
	if err != nil {
		return nil, fmt.Errorf("failed to list supplier offers: %w", err)
	}

	dtos := make([]domain.OfferDTO, len(offers))
	for i, offer := range offers {
		dtos[i] = mapper.ToOfferDTO(&offer)
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

// ============================================================================
// Supplier Contact Methods
// ============================================================================

// ErrSupplierContactNotFound is returned when a supplier contact is not found
var ErrSupplierContactNotFound = errors.New("supplier contact not found")

// ErrContactUsedInActiveOffers is returned when trying to delete a contact that is used in active offers
var ErrContactUsedInActiveOffers = errors.New("contact is assigned to active offers and cannot be deleted")

// ListContacts returns all contacts for a supplier
func (s *SupplierService) ListContacts(ctx context.Context, supplierID uuid.UUID) ([]domain.SupplierContactDTO, error) {
	// Verify supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	contacts, err := s.supplierRepo.ListContacts(ctx, supplier.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contacts: %w", err)
	}

	dtos := make([]domain.SupplierContactDTO, len(contacts))
	for i, contact := range contacts {
		dtos[i] = mapper.SupplierContactToDTO(&contact)
	}

	return dtos, nil
}

// GetContact retrieves a supplier contact by ID
func (s *SupplierService) GetContact(ctx context.Context, supplierID, contactID uuid.UUID) (*domain.SupplierContactDTO, error) {
	// Verify supplier exists
	_, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	contact, err := s.supplierRepo.GetContactByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify contact belongs to the supplier
	if contact.SupplierID != supplierID {
		return nil, ErrSupplierContactNotFound
	}

	dto := mapper.SupplierContactToDTO(contact)
	return &dto, nil
}

// CreateContact creates a new contact for a supplier
func (s *SupplierService) CreateContact(ctx context.Context, supplierID uuid.UUID, req *domain.CreateSupplierContactRequest) (*domain.SupplierContactDTO, error) {
	// Verify supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%w: email", ErrInvalidEmailFormat)
	}

	// Validate phone format
	if err := validatePhone(req.Phone); err != nil {
		return nil, fmt.Errorf("%w: phone", ErrInvalidPhoneFormat)
	}

	// If this contact is marked as primary, clear existing primary contacts
	if req.IsPrimary {
		if err := s.supplierRepo.ClearPrimaryContacts(ctx, supplierID); err != nil {
			return nil, fmt.Errorf("failed to clear primary contacts: %w", err)
		}
	}

	contact := &domain.SupplierContact{
		SupplierID: supplierID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Title:      req.Title,
		Email:      req.Email,
		Phone:      req.Phone,
		IsPrimary:  req.IsPrimary,
		Notes:      req.Notes,
	}

	if err := s.supplierRepo.CreateContact(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Log activity
	s.logActivity(ctx, supplier.ID, supplier.Name, "Kontaktperson lagt til", fmt.Sprintf("Kontaktperson '%s' ble lagt til leverandor '%s'", contact.FullName(), supplier.Name))

	dto := mapper.SupplierContactToDTO(contact)
	return &dto, nil
}

// UpdateContact updates an existing supplier contact
func (s *SupplierService) UpdateContact(ctx context.Context, supplierID, contactID uuid.UUID, req *domain.UpdateSupplierContactRequest) (*domain.SupplierContactDTO, error) {
	// Verify supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierNotFound
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Get existing contact
	contact, err := s.supplierRepo.GetContactByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSupplierContactNotFound
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify contact belongs to the supplier
	if contact.SupplierID != supplierID {
		return nil, ErrSupplierContactNotFound
	}

	// Validate email format
	if err := validateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("%w: email", ErrInvalidEmailFormat)
	}

	// Validate phone format
	if err := validatePhone(req.Phone); err != nil {
		return nil, fmt.Errorf("%w: phone", ErrInvalidPhoneFormat)
	}

	// If this contact is being marked as primary, clear existing primary contacts
	if req.IsPrimary && !contact.IsPrimary {
		if err := s.supplierRepo.ClearPrimaryContacts(ctx, supplierID); err != nil {
			return nil, fmt.Errorf("failed to clear primary contacts: %w", err)
		}
	}

	// Update contact fields
	contact.FirstName = req.FirstName
	contact.LastName = req.LastName
	contact.Title = req.Title
	contact.Email = req.Email
	contact.Phone = req.Phone
	contact.IsPrimary = req.IsPrimary
	contact.Notes = req.Notes

	if err := s.supplierRepo.UpdateContact(ctx, contact); err != nil {
		return nil, fmt.Errorf("failed to update contact: %w", err)
	}

	// Log activity
	s.logActivity(ctx, supplier.ID, supplier.Name, "Kontaktperson oppdatert", fmt.Sprintf("Kontaktperson '%s' ble oppdatert", contact.FullName()))

	dto := mapper.SupplierContactToDTO(contact)
	return &dto, nil
}

// DeleteContact deletes a supplier contact
func (s *SupplierService) DeleteContact(ctx context.Context, supplierID, contactID uuid.UUID) error {
	// Verify supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSupplierNotFound
		}
		return fmt.Errorf("failed to get supplier: %w", err)
	}

	// Get existing contact
	contact, err := s.supplierRepo.GetContactByID(ctx, contactID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSupplierContactNotFound
		}
		return fmt.Errorf("failed to get contact: %w", err)
	}

	// Verify contact belongs to the supplier
	if contact.SupplierID != supplierID {
		return ErrSupplierContactNotFound
	}

	// Check if contact is used in active offers
	isUsed, err := s.supplierRepo.IsContactUsedInOffers(ctx, contactID)
	if err != nil {
		return fmt.Errorf("failed to check contact usage: %w", err)
	}
	if isUsed {
		return ErrContactUsedInActiveOffers
	}

	if err := s.supplierRepo.DeleteContact(ctx, contactID); err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	// Log activity
	s.logActivity(ctx, supplier.ID, supplier.Name, "Kontaktperson slettet", fmt.Sprintf("Kontaktperson '%s' ble slettet fra leverandor '%s'", contact.FullName(), supplier.Name))

	return nil
}
