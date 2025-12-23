package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/storage"
	"go.uber.org/zap"
)

// FileService handles file operations with entity validation and activity logging
type FileService struct {
	fileRepo     *repository.FileRepository
	offerRepo    *repository.OfferRepository
	customerRepo *repository.CustomerRepository
	projectRepo  *repository.ProjectRepository
	supplierRepo *repository.SupplierRepository
	activityRepo *repository.ActivityRepository
	storage      storage.Storage
	logger       *zap.Logger
}

// NewFileService creates a new FileService instance with all required dependencies
func NewFileService(
	fileRepo *repository.FileRepository,
	offerRepo *repository.OfferRepository,
	customerRepo *repository.CustomerRepository,
	projectRepo *repository.ProjectRepository,
	supplierRepo *repository.SupplierRepository,
	activityRepo *repository.ActivityRepository,
	storage storage.Storage,
	logger *zap.Logger,
) *FileService {
	return &FileService{
		fileRepo:     fileRepo,
		offerRepo:    offerRepo,
		customerRepo: customerRepo,
		projectRepo:  projectRepo,
		supplierRepo: supplierRepo,
		activityRepo: activityRepo,
		storage:      storage,
		logger:       logger,
	}
}

// ============================================================================
// Entity-Specific Upload Methods
// ============================================================================

// UploadToCustomer uploads a file and attaches it to a customer
// companyID is always provided from the auth context (X-Company-Id header)
func (s *FileService) UploadToCustomer(ctx context.Context, customerID uuid.UUID, filename, contentType string, data io.Reader, companyID domain.CompanyID) (*domain.FileDTO, error) {
	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Upload file with the provided company
	file := &domain.File{
		CustomerID: &customerID,
		CompanyID:  companyID,
	}

	dto, err := s.uploadFile(ctx, file, filename, contentType, data, "customer", customer.Name)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

// UploadToProject uploads a file and attaches it to a project
// companyID is always provided from the auth context (X-Company-Id header)
func (s *FileService) UploadToProject(ctx context.Context, projectID uuid.UUID, filename, contentType string, data io.Reader, companyID domain.CompanyID) (*domain.FileDTO, error) {
	// Verify project exists
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Upload file with the provided company
	file := &domain.File{
		ProjectID: &projectID,
		CompanyID: companyID,
	}

	dto, err := s.uploadFile(ctx, file, filename, contentType, data, "project", project.Name)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

// UploadToOffer uploads a file and attaches it to an offer
// companyID is always provided from the auth context (X-Company-Id header)
func (s *FileService) UploadToOffer(ctx context.Context, offerID uuid.UUID, filename, contentType string, data io.Reader, companyID domain.CompanyID) (*domain.FileDTO, error) {
	// Verify offer exists
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("offer not found: %w", err)
	}

	// Upload file with the provided company
	file := &domain.File{
		OfferID:   &offerID,
		CompanyID: companyID,
	}

	dto, err := s.uploadFile(ctx, file, filename, contentType, data, "offer", offer.Title)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

// UploadToSupplier uploads a file and attaches it to a supplier
// companyID is always provided from the auth context (X-Company-Id header)
func (s *FileService) UploadToSupplier(ctx context.Context, supplierID uuid.UUID, filename, contentType string, data io.Reader, companyID domain.CompanyID) (*domain.FileDTO, error) {
	// Verify supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, supplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	// Upload file with the provided company
	file := &domain.File{
		SupplierID: &supplierID,
		CompanyID:  companyID,
	}

	dto, err := s.uploadFile(ctx, file, filename, contentType, data, "supplier", supplier.Name)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

// UploadToOfferSupplier uploads a file and attaches it to an offer-supplier relationship
// companyID is always provided from the auth context (X-Company-Id header)
func (s *FileService) UploadToOfferSupplier(ctx context.Context, offerID, supplierID uuid.UUID, filename, contentType string, data io.Reader, companyID domain.CompanyID) (*domain.FileDTO, error) {
	// Verify offer-supplier relationship exists
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		return nil, fmt.Errorf("offer-supplier relationship not found: %w", err)
	}

	// Upload file with the provided company
	file := &domain.File{
		OfferSupplierID: &offerSupplier.ID,
		CompanyID:       companyID,
	}

	entityName := fmt.Sprintf("%s - %s", offerSupplier.OfferTitle, offerSupplier.SupplierName)
	dto, err := s.uploadFile(ctx, file, filename, contentType, data, "offer-supplier", entityName)
	if err != nil {
		return nil, err
	}

	return dto, nil
}

// uploadFile is a helper method that handles the common upload logic
func (s *FileService) uploadFile(ctx context.Context, file *domain.File, filename, contentType string, data io.Reader, entityType, entityName string) (*domain.FileDTO, error) {
	// Upload to storage
	storagePath, size, err := s.storage.Upload(ctx, filename, contentType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Set file properties
	file.Filename = filename
	file.ContentType = contentType
	file.Size = size
	file.StoragePath = storagePath

	// Create file record
	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Try to delete from storage (best effort cleanup)
		if delErr := s.storage.Delete(ctx, storagePath); delErr != nil {
			s.logger.Warn("failed to cleanup file from storage after DB error",
				zap.Error(delErr),
				zap.String("storagePath", storagePath),
			)
		}
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Log activity
	s.logFileActivity(ctx, file, "Fil lastet opp",
		fmt.Sprintf("Filen '%s' ble lastet opp til %s '%s'", filename, entityType, entityName))

	dto := mapper.ToFileDTO(file)
	return &dto, nil
}

// ============================================================================
// Entity-Specific List Methods
// ============================================================================

// ListByCustomer returns all files attached to a customer, filtered by company access
// Files from the user's company and "gruppen" are returned; gruppen users see all files
func (s *FileService) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]domain.FileDTO, error) {
	// Verify customer exists
	if _, err := s.customerRepo.GetByID(ctx, customerID); err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Get company filter based on user context
	companyFilter := s.getCompanyFilter(ctx)

	files, err := s.fileRepo.ListByCustomer(ctx, customerID, companyFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list customer files: %w", err)
	}

	return mapper.ToFileDTOs(files), nil
}

// ListByProject returns all files attached to a project, filtered by company access
// Files from the user's company and "gruppen" are returned; gruppen users see all files
func (s *FileService) ListByProject(ctx context.Context, projectID uuid.UUID) ([]domain.FileDTO, error) {
	// Verify project exists
	if _, err := s.projectRepo.GetByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	// Get company filter based on user context
	companyFilter := s.getCompanyFilter(ctx)

	files, err := s.fileRepo.ListByProject(ctx, projectID, companyFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list project files: %w", err)
	}

	return mapper.ToFileDTOs(files), nil
}

// ListByOffer returns all files attached to an offer, filtered by company access
// Files from the user's company and "gruppen" are returned; gruppen users see all files
func (s *FileService) ListByOffer(ctx context.Context, offerID uuid.UUID) ([]domain.FileDTO, error) {
	// Verify offer exists
	if _, err := s.offerRepo.GetByID(ctx, offerID); err != nil {
		return nil, fmt.Errorf("offer not found: %w", err)
	}

	// Get company filter based on user context
	companyFilter := s.getCompanyFilter(ctx)

	files, err := s.fileRepo.ListByOffer(ctx, offerID, companyFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list offer files: %w", err)
	}

	return mapper.ToFileDTOs(files), nil
}

// ListBySupplier returns all files attached to a supplier, filtered by company access
// Files from the user's company and "gruppen" are returned; gruppen users see all files
func (s *FileService) ListBySupplier(ctx context.Context, supplierID uuid.UUID) ([]domain.FileDTO, error) {
	// Verify supplier exists
	if _, err := s.supplierRepo.GetByID(ctx, supplierID); err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}

	// Get company filter based on user context
	companyFilter := s.getCompanyFilter(ctx)

	files, err := s.fileRepo.ListBySupplier(ctx, supplierID, companyFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list supplier files: %w", err)
	}

	return mapper.ToFileDTOs(files), nil
}

// ListByOfferSupplier returns all files attached to an offer-supplier relationship, filtered by company access
// Files from the user's company and "gruppen" are returned; gruppen users see all files
func (s *FileService) ListByOfferSupplier(ctx context.Context, offerID, supplierID uuid.UUID) ([]domain.FileDTO, error) {
	// Verify offer-supplier relationship exists
	offerSupplier, err := s.offerRepo.GetOfferSupplier(ctx, offerID, supplierID)
	if err != nil {
		return nil, fmt.Errorf("offer-supplier relationship not found: %w", err)
	}

	// Get company filter based on user context
	companyFilter := s.getCompanyFilter(ctx)

	files, err := s.fileRepo.ListByOfferSupplier(ctx, offerSupplier.ID, companyFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list offer-supplier files: %w", err)
	}

	return mapper.ToFileDTOs(files), nil
}

// ============================================================================
// Generic File Operations
// ============================================================================

// Upload uploads a file with optional offer ID (legacy method for backward compatibility)
// Deprecated: Use entity-specific upload methods (UploadToCustomer, UploadToProject, etc.)
// companyID is required - if offerID is provided and companyID is empty, inherits from the offer
func (s *FileService) Upload(ctx context.Context, filename, contentType string, data io.Reader, offerID *uuid.UUID, companyID domain.CompanyID) (*domain.FileDTO, error) {
	// Verify offer exists if provided and determine company
	var entityName string
	effectiveCompanyID := companyID
	if offerID != nil {
		offer, err := s.offerRepo.GetByID(ctx, *offerID)
		if err != nil {
			return nil, fmt.Errorf("offer not found: %w", err)
		}
		entityName = offer.Title
		// Inherit company from offer if not provided
		if effectiveCompanyID == "" {
			effectiveCompanyID = offer.CompanyID
		}
	}

	// Company is required
	if effectiveCompanyID == "" {
		return nil, fmt.Errorf("company_id is required for file uploads")
	}

	// Upload to storage
	storagePath, size, err := s.storage.Upload(ctx, filename, contentType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create file record
	file := &domain.File{
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		StoragePath: storagePath,
		OfferID:     offerID,
		CompanyID:   effectiveCompanyID,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Try to delete from storage (best effort cleanup)
		_ = s.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Log activity
	if offerID != nil {
		s.logFileActivity(ctx, file, "Fil lastet opp",
			fmt.Sprintf("Filen '%s' ble lastet opp til tilbud '%s'", filename, entityName))
	} else {
		s.logFileActivity(ctx, file, "Fil lastet opp",
			fmt.Sprintf("Filen '%s' ble lastet opp", filename))
	}

	dto := mapper.ToFileDTO(file)
	return &dto, nil
}

// GetByID retrieves a file by its ID
func (s *FileService) GetByID(ctx context.Context, id uuid.UUID) (*domain.FileDTO, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	dto := mapper.ToFileDTO(file)
	return &dto, nil
}

// Download retrieves a file's content for download
// Returns: reader, filename, content-type, error
func (s *FileService) Download(ctx context.Context, id uuid.UUID) (io.ReadCloser, string, string, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get file: %w", err)
	}

	reader, err := s.storage.Download(ctx, file.StoragePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to download file: %w", err)
	}

	return reader, file.Filename, file.ContentType, nil
}

// Delete removes a file from both storage and database
func (s *FileService) Delete(ctx context.Context, id uuid.UUID) error {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Delete from storage (log warning if fails, continue)
	if err := s.storage.Delete(ctx, file.StoragePath); err != nil {
		s.logger.Warn("failed to delete file from storage",
			zap.Error(err),
			zap.String("storagePath", file.StoragePath),
			zap.String("fileID", id.String()),
		)
	}

	// Delete from database
	if err := s.fileRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	// Log activity
	s.logFileActivity(ctx, file, "Fil slettet",
		fmt.Sprintf("Filen '%s' ble slettet", file.Filename))

	return nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// logFileActivity creates an activity log entry for file operations
func (s *FileService) logFileActivity(ctx context.Context, file *domain.File, title, body string) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		// System operation, create activity without user context
		activity := &domain.Activity{
			TargetType: domain.ActivityTargetFile,
			TargetID:   file.ID,
			TargetName: file.Filename,
			Title:      title,
			Body:       body,
		}
		if err := s.activityRepo.Create(ctx, activity); err != nil {
			s.logger.Warn("failed to create file activity",
				zap.Error(err),
				zap.String("fileID", file.ID.String()),
			)
		}
		return
	}

	activity := &domain.Activity{
		TargetType:  domain.ActivityTargetFile,
		TargetID:    file.ID,
		TargetName:  file.Filename,
		Title:       title,
		Body:        body,
		CreatorName: userCtx.DisplayName,
		CreatorID:   userCtx.UserID.String(),
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to create file activity",
			zap.Error(err),
			zap.String("fileID", file.ID.String()),
			zap.String("userID", userCtx.UserID.String()),
		)
	}
}

// getCompanyFilter extracts the company filter from context for file visibility
// Returns nil for gruppen users (can see all files), or the user's company ID for filtering
func (s *FileService) getCompanyFilter(ctx context.Context) *domain.CompanyID {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		// No user context - system operation, no filtering
		return nil
	}

	// Gruppen users and super admins can see all files
	if userCtx.IsGruppenUser() {
		return nil
	}

	// Regular users see only their company's files (plus gruppen files via repository filter)
	return &userCtx.CompanyID
}
