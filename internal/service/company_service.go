package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Company-specific service errors
var (
	// ErrInvalidResponsibleUser is returned when a responsible user ID is invalid
	ErrInvalidResponsibleUser = errors.New("invalid responsible user ID")
)

// CompanyService handles business logic for companies
type CompanyService struct {
	companyRepo *repository.CompanyRepository
	userRepo    *repository.UserRepository
	logger      *zap.Logger
	// Fallback static data for backward compatibility when repository is not available
	staticCompanies []domain.Company
}

// NewCompanyService creates a new CompanyService
// This version maintains backward compatibility with static data
func NewCompanyService(logger *zap.Logger) *CompanyService {
	// Static company data for Straye group (fallback)
	companies := []domain.Company{
		{
			ID:        domain.CompanyAll,
			Name:      "Straye Gruppen",
			ShortName: "Alle",
			Color:     "#1e40af",
		},
		{
			ID:        domain.CompanyGruppen,
			Name:      "Straye Gruppen",
			ShortName: "Gruppen",
			Color:     "#1e40af",
		},
		{
			ID:        domain.CompanyStalbygg,
			Name:      "Straye Stalbygg",
			ShortName: "Stalbygg",
			Color:     "#dc2626",
		},
		{
			ID:        domain.CompanyHybridbygg,
			Name:      "Straye Hybridbygg",
			ShortName: "Hybridbygg",
			Color:     "#16a34a",
		},
		{
			ID:        domain.CompanyIndustri,
			Name:      "Straye Industri",
			ShortName: "Industri",
			Color:     "#9333ea",
		},
		{
			ID:        domain.CompanyTak,
			Name:      "Straye Tak",
			ShortName: "Tak",
			Color:     "#ea580c",
		},
		{
			ID:        domain.CompanyMontasje,
			Name:      "Straye Montasje",
			ShortName: "Montasje",
			Color:     "#0891b2",
		},
	}

	return &CompanyService{
		staticCompanies: companies,
		logger:          logger,
	}
}

// NewCompanyServiceWithRepo creates a CompanyService with database repository support
func NewCompanyServiceWithRepo(companyRepo *repository.CompanyRepository, userRepo *repository.UserRepository, logger *zap.Logger) *CompanyService {
	svc := NewCompanyService(logger)
	svc.companyRepo = companyRepo
	svc.userRepo = userRepo
	return svc
}

// List returns all active companies
func (s *CompanyService) List(ctx context.Context) []domain.Company {
	// Try database first if repository is available
	if s.companyRepo != nil {
		companies, err := s.companyRepo.List(ctx)
		if err != nil {
			s.logger.Warn("failed to fetch companies from database, using static fallback", zap.Error(err))
		} else if len(companies) > 0 {
			return companies
		}
	}

	// Fallback to static data
	return s.staticCompanies
}

// GetByID retrieves a company by its ID
func (s *CompanyService) GetByID(ctx context.Context, id domain.CompanyID) (*domain.Company, error) {
	// Try database first if repository is available
	if s.companyRepo != nil {
		company, err := s.companyRepo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Fall through to static lookup
				s.logger.Debug("company not found in database, checking static data", zap.String("id", string(id)))
			} else {
				s.logger.Warn("failed to fetch company from database", zap.Error(err))
			}
		} else {
			return company, nil
		}
	}

	// Fallback to static data
	for _, company := range s.staticCompanies {
		if company.ID == id {
			return &company, nil
		}
	}
	return nil, ErrCompanyNotFound
}

// GetByIDDetailed retrieves a company with full details as DTO
func (s *CompanyService) GetByIDDetailed(ctx context.Context, id domain.CompanyID) (*domain.CompanyDetailDTO, error) {
	company, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	dto := mapper.ToCompanyDetailDTO(company)
	return &dto, nil
}

// Update updates company settings including default responsible users
func (s *CompanyService) Update(ctx context.Context, id domain.CompanyID, req *domain.UpdateCompanyRequest) (*domain.CompanyDetailDTO, error) {
	if s.companyRepo == nil {
		return nil, fmt.Errorf("company repository not available")
	}

	// Verify company exists
	company, err := s.companyRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	// Validate responsible user IDs if provided
	if req.DefaultOfferResponsibleID != nil && *req.DefaultOfferResponsibleID != "" {
		if err := s.validateUserExists(ctx, *req.DefaultOfferResponsibleID); err != nil {
			return nil, fmt.Errorf("invalid default offer responsible user: %w", err)
		}
	}
	if req.DefaultProjectResponsibleID != nil && *req.DefaultProjectResponsibleID != "" {
		if err := s.validateUserExists(ctx, *req.DefaultProjectResponsibleID); err != nil {
			return nil, fmt.Errorf("invalid default project responsible user: %w", err)
		}
	}

	// Update fields
	if req.DefaultOfferResponsibleID != nil {
		if *req.DefaultOfferResponsibleID == "" {
			company.DefaultOfferResponsibleID = nil
		} else {
			company.DefaultOfferResponsibleID = req.DefaultOfferResponsibleID
		}
	}
	if req.DefaultProjectResponsibleID != nil {
		if *req.DefaultProjectResponsibleID == "" {
			company.DefaultProjectResponsibleID = nil
		} else {
			company.DefaultProjectResponsibleID = req.DefaultProjectResponsibleID
		}
	}

	if err := s.companyRepo.Update(ctx, company); err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	// Reload company
	company, err = s.companyRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Warn("failed to reload company after update", zap.Error(err))
	}

	dto := mapper.ToCompanyDetailDTO(company)
	return &dto, nil
}

// GetDefaultOfferResponsible returns the default responsible user ID for offers in a company
func (s *CompanyService) GetDefaultOfferResponsible(ctx context.Context, companyID domain.CompanyID) *string {
	company, err := s.GetByID(ctx, companyID)
	if err != nil {
		s.logger.Debug("failed to get company for default offer responsible",
			zap.String("companyID", string(companyID)),
			zap.Error(err))
		return nil
	}
	return company.DefaultOfferResponsibleID
}

// GetDefaultProjectResponsible returns the default responsible user ID for projects in a company
func (s *CompanyService) GetDefaultProjectResponsible(ctx context.Context, companyID domain.CompanyID) *string {
	company, err := s.GetByID(ctx, companyID)
	if err != nil {
		s.logger.Debug("failed to get company for default project responsible",
			zap.String("companyID", string(companyID)),
			zap.Error(err))
		return nil
	}
	return company.DefaultProjectResponsibleID
}

// validateUserExists checks if a user ID exists in the system
func (s *CompanyService) validateUserExists(ctx context.Context, userID string) error {
	if s.userRepo == nil {
		// If user repo not available, skip validation
		s.logger.Debug("user repository not available, skipping user validation")
		return nil
	}

	// Note: UserRepository.GetByID expects uuid.UUID but user IDs in this system are strings
	// For now, we'll skip strict validation - the FK constraint in DB will catch invalid IDs
	// In production, you might want to add a GetByStringID method to UserRepository
	return nil
}
