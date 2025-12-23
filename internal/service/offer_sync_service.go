package service

// This file contains data warehouse sync methods extracted from offer_service.go
// for better code organization. These methods handle:
// - Syncing financial data from the data warehouse
// - Bulk and stale offer syncing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ============================================================================
// Data Warehouse Sync Methods
// ============================================================================

// ErrNoExternalReference indicates the offer has no external reference for DW sync
var ErrNoExternalReference = errors.New("offer has no external reference")

// ErrDataWarehouseNotAvailable indicates the data warehouse client is not available
var ErrDataWarehouseNotAvailable = errors.New("data warehouse not available")

// SyncFromDataWarehouse syncs financial data from the data warehouse for a single offer.
// Returns the sync response with updated financials and sync status.
// The offer must have an external_reference to be synced.
func (s *OfferService) SyncFromDataWarehouse(ctx context.Context, id uuid.UUID) (*domain.OfferExternalSyncResponse, error) {
	// Get the offer
	offer, err := s.offerRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	// Build base response
	response := &domain.OfferExternalSyncResponse{
		OfferID:           offer.ID,
		ExternalReference: offer.ExternalReference,
		CompanyID:         offer.CompanyID,
		DataWarehouse: &domain.DataWarehouseFinancialsDTO{
			TotalIncome:   0,
			MaterialCosts: 0,
			EmployeeCosts: 0,
			OtherCosts:    0,
			NetResult:     0,
			Connected:     false,
		},
		Persisted: false,
	}

	// Validate external reference exists
	if offer.ExternalReference == "" {
		s.logger.Info("offer has no external reference, skipping DW sync",
			zap.String("offer_id", id.String()))
		return response, nil
	}

	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping sync",
			zap.String("offer_id", id.String()))
		return response, nil
	}

	// Query the data warehouse for project financials
	financials, err := s.dwClient.GetProjectFinancials(ctx, string(offer.CompanyID), offer.ExternalReference)
	if err != nil {
		s.logger.Error("failed to query data warehouse",
			zap.Error(err),
			zap.String("offer_id", id.String()),
			zap.String("external_reference", offer.ExternalReference),
			zap.String("company_id", string(offer.CompanyID)))
		// Return disconnected response on error (don't fail the request)
		return response, nil
	}

	// Update the offer with DW financials
	dwFinancials := &repository.DWFinancials{
		TotalIncome:   financials.TotalIncome,
		MaterialCosts: financials.MaterialCosts,
		EmployeeCosts: financials.EmployeeCosts,
		OtherCosts:    financials.OtherCosts,
		NetResult:     financials.NetResult,
	}

	updatedOffer, err := s.offerRepo.UpdateDWFinancials(ctx, id, dwFinancials)
	if err != nil {
		s.logger.Error("failed to update offer with DW financials",
			zap.Error(err),
			zap.String("offer_id", id.String()))
		// Still return the data even if persistence failed
		response.DataWarehouse = &domain.DataWarehouseFinancialsDTO{
			TotalIncome:   financials.TotalIncome,
			MaterialCosts: financials.MaterialCosts,
			EmployeeCosts: financials.EmployeeCosts,
			OtherCosts:    financials.OtherCosts,
			NetResult:     financials.NetResult,
			Connected:     true,
		}
		return response, nil
	}

	// Get the synced timestamp in UTC with proper timezone
	now := time.Now().UTC().Format(time.RFC3339)

	// Log activity for the sync
	activity := &domain.Activity{
		TargetType:   domain.ActivityTargetOffer,
		TargetID:     offer.ID,
		TargetName:   offer.Title,
		Title:        "Finansdata synkronisert",
		Body:         "Finansdata synkronisert fra datavarehus",
		ActivityType: domain.ActivityTypeSystem,
		Status:       domain.ActivityStatusCompleted,
		OccurredAt:   time.Now(),
		CreatorName:  "System",
		CompanyID:    &offer.CompanyID,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log DW sync activity",
			zap.Error(err),
			zap.String("offer_id", id.String()))
		// Don't fail the sync for activity logging failure
	}

	s.logger.Info("successfully synced DW financials for offer",
		zap.String("offer_id", id.String()),
		zap.String("external_reference", offer.ExternalReference),
		zap.Float64("total_income", financials.TotalIncome),
		zap.Float64("material_costs", financials.MaterialCosts),
		zap.Float64("employee_costs", financials.EmployeeCosts),
		zap.Float64("other_costs", financials.OtherCosts),
		zap.Float64("net_result", financials.NetResult))

	// Return success response with updated offer values
	response.DataWarehouse = &domain.DataWarehouseFinancialsDTO{
		TotalIncome:   financials.TotalIncome,
		MaterialCosts: financials.MaterialCosts,
		EmployeeCosts: financials.EmployeeCosts,
		OtherCosts:    financials.OtherCosts,
		NetResult:     financials.NetResult,
		Connected:     true,
	}
	response.SyncedAt = &now
	response.Persisted = true

	// Include the persisted DW values and updated Spent/Invoiced from the offer
	response.DWTotalIncome = updatedOffer.DWTotalIncome
	response.DWMaterialCosts = updatedOffer.DWMaterialCosts
	response.DWEmployeeCosts = updatedOffer.DWEmployeeCosts
	response.DWOtherCosts = updatedOffer.DWOtherCosts
	response.DWNetResult = updatedOffer.DWNetResult
	response.Spent = updatedOffer.Spent
	response.Invoiced = updatedOffer.Invoiced

	return response, nil
}

// SyncAllOffersFromDataWarehouse syncs financial data from the data warehouse for all offers
// in "order" phase that have an external_reference. Continues on error for individual offers.
// Returns counts for successfully synced and failed offers.
// Note: Only "order" phase offers are synced automatically; other phases require manual sync.
func (s *OfferService) SyncAllOffersFromDataWarehouse(ctx context.Context) (synced int, failed int, err error) {
	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping bulk sync")
		return 0, 0, ErrDataWarehouseNotAvailable
	}

	// Get "order" phase offers with external_reference that need syncing
	// Sync offers not synced in the last 55 minutes (allows buffer for hourly cron timing variations)
	maxAge := 55 * time.Minute
	offers, err := s.offerRepo.GetOffersNeedingDWSync(ctx, maxAge)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get offers for DW sync: %w", err)
	}

	s.logger.Info("starting bulk DW sync",
		zap.Int("offer_count", len(offers)))

	for _, offer := range offers {
		// Query the data warehouse for project financials
		financials, err := s.dwClient.GetProjectFinancials(ctx, string(offer.CompanyID), offer.ExternalReference)
		if err != nil {
			s.logger.Warn("failed to query DW for offer",
				zap.Error(err),
				zap.String("offer_id", offer.ID.String()),
				zap.String("external_reference", offer.ExternalReference))
			failed++
			continue
		}

		// Update the offer with DW financials
		dwFinancials := &repository.DWFinancials{
			TotalIncome:   financials.TotalIncome,
			MaterialCosts: financials.MaterialCosts,
			EmployeeCosts: financials.EmployeeCosts,
			OtherCosts:    financials.OtherCosts,
			NetResult:     financials.NetResult,
		}

		if _, err := s.offerRepo.UpdateDWFinancials(ctx, offer.ID, dwFinancials); err != nil {
			s.logger.Warn("failed to update offer with DW financials",
				zap.Error(err),
				zap.String("offer_id", offer.ID.String()))
			failed++
			continue
		}

		synced++
	}

	s.logger.Info("completed bulk DW sync",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.Int("total", len(offers)))

	return synced, failed, nil
}

// GetOffersForDWSync returns all offers that have an external_reference and can be synced.
func (s *OfferService) GetOffersForDWSync(ctx context.Context) ([]domain.Offer, error) {
	return s.offerRepo.GetOffersForDWSync(ctx)
}

// SyncStaleOffersFromDataWarehouse syncs only offers that are stale (never synced or older than maxAge).
// This is used for startup sync to catch up on offers that haven't been synced recently.
// Returns counts for successfully synced and failed offers.
func (s *OfferService) SyncStaleOffersFromDataWarehouse(ctx context.Context, maxAge time.Duration) (synced int, failed int, err error) {
	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping stale sync")
		return 0, 0, ErrDataWarehouseNotAvailable
	}

	// Get offers that need syncing (null or older than maxAge)
	offers, err := s.offerRepo.GetOffersNeedingDWSync(ctx, maxAge)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get offers needing DW sync: %w", err)
	}

	if len(offers) == 0 {
		s.logger.Info("no stale offers to sync")
		return 0, 0, nil
	}

	s.logger.Info("starting stale offer DW sync",
		zap.Int("offer_count", len(offers)),
		zap.Duration("max_age", maxAge))

	for _, offer := range offers {
		// Query the data warehouse for project financials
		financials, err := s.dwClient.GetProjectFinancials(ctx, string(offer.CompanyID), offer.ExternalReference)
		if err != nil {
			s.logger.Warn("failed to query DW for stale offer",
				zap.Error(err),
				zap.String("offer_id", offer.ID.String()),
				zap.String("external_reference", offer.ExternalReference))
			failed++
			continue
		}

		// Update the offer with DW financials
		dwFinancials := &repository.DWFinancials{
			TotalIncome:   financials.TotalIncome,
			MaterialCosts: financials.MaterialCosts,
			EmployeeCosts: financials.EmployeeCosts,
			OtherCosts:    financials.OtherCosts,
			NetResult:     financials.NetResult,
		}

		if _, err := s.offerRepo.UpdateDWFinancials(ctx, offer.ID, dwFinancials); err != nil {
			s.logger.Warn("failed to update stale offer with DW financials",
				zap.Error(err),
				zap.String("offer_id", offer.ID.String()))
			failed++
			continue
		}

		synced++
	}

	s.logger.Info("completed stale offer DW sync",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.Int("total", len(offers)))

	return synced, failed, nil
}
