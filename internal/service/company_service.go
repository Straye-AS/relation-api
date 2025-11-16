package service

import (
	"context"

	"github.com/straye-as/relation-api/internal/domain"
	"go.uber.org/zap"
)

type CompanyService struct {
	companies []domain.Company
	logger    *zap.Logger
}

func NewCompanyService(logger *zap.Logger) *CompanyService {
	// Static company data for Straye group
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
			Name:      "Straye Stålbygg",
			ShortName: "Stålbygg",
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
		companies: companies,
		logger:    logger,
	}
}

func (s *CompanyService) List(ctx context.Context) []domain.Company {
	return s.companies
}

func (s *CompanyService) GetByID(ctx context.Context, id domain.CompanyID) (*domain.Company, error) {
	for _, company := range s.companies {
		if company.ID == id {
			return &company, nil
		}
	}
	return nil, nil
}
