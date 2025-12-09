package service

import (
	"context"
	"fmt"
	"time"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

// NumberSequenceService handles generation of unique, formatted numbers
// for both offers and projects. Numbers are shared within a company/year
// to ensure uniqueness across both entity types.
//
// Format: {PREFIX}-{YEAR}-{SEQUENCE}
// Example: ST-2025-001, HB-2025-042
type NumberSequenceService struct {
	repo   *repository.NumberSequenceRepository
	logger *zap.Logger
}

// NewNumberSequenceService creates a new NumberSequenceService
func NewNumberSequenceService(
	repo *repository.NumberSequenceRepository,
	logger *zap.Logger,
) *NumberSequenceService {
	return &NumberSequenceService{
		repo:   repo,
		logger: logger,
	}
}

// GenerateOfferNumber generates a unique offer number for a company.
// This should be called when an offer transitions from draft to in_progress.
// Format: {PREFIX}-{YEAR}-{SEQUENCE} e.g., "ST-2025-001"
//
// The number is SHARED with projects - both use the same sequence counter
// per company/year to ensure global uniqueness.
func (s *NumberSequenceService) GenerateOfferNumber(ctx context.Context, companyID domain.CompanyID) (string, error) {
	return s.generateNumber(ctx, companyID, "offer")
}

// GenerateProjectNumber generates a unique project number for a company.
// This should be called when a new project is created.
// Format: {PREFIX}-{YEAR}-{SEQUENCE} e.g., "ST-2025-002"
//
// The number is SHARED with offers - both use the same sequence counter
// per company/year to ensure global uniqueness.
func (s *NumberSequenceService) GenerateProjectNumber(ctx context.Context, companyID domain.CompanyID) (string, error) {
	return s.generateNumber(ctx, companyID, "project")
}

// generateNumber is the internal method that generates a formatted number.
// entityType is used only for logging purposes.
func (s *NumberSequenceService) generateNumber(ctx context.Context, companyID domain.CompanyID, entityType string) (string, error) {
	// Validate company ID
	if !domain.IsValidCompanyID(string(companyID)) {
		return "", fmt.Errorf("%w: %s", ErrInvalidCompanyID, companyID)
	}

	year := time.Now().Year()
	prefix := domain.GetCompanyPrefix(companyID)

	// Get the next sequence number (atomic operation)
	nextSeq, err := s.repo.GetNextNumber(ctx, companyID, year)
	if err != nil {
		s.logger.Error("failed to get next sequence number",
			zap.String("companyID", string(companyID)),
			zap.Int("year", year),
			zap.String("entityType", entityType),
			zap.Error(err))
		return "", fmt.Errorf("failed to generate %s number: %w", entityType, err)
	}

	// Format: PREFIX-YYYY-NNN (zero-padded to 3 digits)
	number := fmt.Sprintf("%s-%d-%03d", prefix, year, nextSeq)

	s.logger.Info("generated number",
		zap.String("number", number),
		zap.String("companyID", string(companyID)),
		zap.Int("year", year),
		zap.Int("sequence", nextSeq),
		zap.String("entityType", entityType))

	return number, nil
}

// GetCompanyPrefix returns the 2-letter prefix for a company.
// This is a convenience method that delegates to domain.GetCompanyPrefix.
func (s *NumberSequenceService) GetCompanyPrefix(companyID domain.CompanyID) string {
	return domain.GetCompanyPrefix(companyID)
}

// GetCurrentSequence returns the current sequence value for a company/year
// without incrementing it. Returns 0 if no sequence exists.
func (s *NumberSequenceService) GetCurrentSequence(ctx context.Context, companyID domain.CompanyID, year int) (int, error) {
	return s.repo.GetCurrentSequence(ctx, companyID, year)
}

// InitializeSequence sets the sequence to a specific value.
// This is useful for data migrations to ensure the sequence
// accounts for existing numbered entities.
// The value should be the LAST USED sequence number.
func (s *NumberSequenceService) InitializeSequence(ctx context.Context, companyID domain.CompanyID, year int, value int) error {
	return s.repo.SetSequence(ctx, companyID, year, value)
}

// ValidateOfferNumber checks if an offer number follows the expected format.
// Returns true if the format is valid: PREFIX-YYYY-NNN
func (s *NumberSequenceService) ValidateOfferNumber(number string) bool {
	// Simple format validation: XX-YYYY-NNN
	if len(number) < 10 {
		return false
	}
	// More thorough validation could be added here
	return true
}
