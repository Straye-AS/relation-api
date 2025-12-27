package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/datawarehouse"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Assignment service errors
var (
	ErrAssignmentNotFound       = errors.New("assignment not found")
	ErrOfferNoExternalReference = errors.New("offer has no external reference for assignment sync")
	ErrDWClientNotAvailable     = errors.New("data warehouse client not available")
)

// AssignmentService handles business logic for assignments
type AssignmentService struct {
	assignmentRepo *repository.AssignmentRepository
	offerRepo      *repository.OfferRepository
	activityRepo   *repository.ActivityRepository
	dwClient       *datawarehouse.Client
	logger         *zap.Logger
}

// NewAssignmentService creates a new assignment service
func NewAssignmentService(
	assignmentRepo *repository.AssignmentRepository,
	offerRepo *repository.OfferRepository,
	activityRepo *repository.ActivityRepository,
	logger *zap.Logger,
) *AssignmentService {
	return &AssignmentService{
		assignmentRepo: assignmentRepo,
		offerRepo:      offerRepo,
		activityRepo:   activityRepo,
		logger:         logger,
	}
}

// SetDataWarehouseClient sets the data warehouse client for syncing.
// Called after construction because the DW client is optional.
func (s *AssignmentService) SetDataWarehouseClient(client *datawarehouse.Client) {
	s.dwClient = client
}

// SyncAssignmentsForOffer fetches assignments from the datawarehouse and syncs them to the local database.
// The offer must have an external_reference that matches the project Code in the ERP.
// Returns a sync result with counts of created/updated/deleted assignments.
func (s *AssignmentService) SyncAssignmentsForOffer(ctx context.Context, offerID uuid.UUID) (*domain.AssignmentSyncResultDTO, error) {
	// Get the offer
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("get offer: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)

	// Build base result
	result := &domain.AssignmentSyncResultDTO{
		OfferID:           offer.ID,
		ExternalReference: offer.ExternalReference,
		CompanyID:         offer.CompanyID,
		SyncedAt:          now,
	}

	// Validate external reference exists
	if offer.ExternalReference == "" {
		s.logger.Info("offer has no external reference, skipping assignment sync",
			zap.String("offer_id", offerID.String()))
		result.Error = "offer has no external reference"
		return result, nil
	}

	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping assignment sync",
			zap.String("offer_id", offerID.String()))
		result.Error = "data warehouse not available"
		return result, nil
	}

	// Fetch assignments from datawarehouse
	erpAssignments, err := s.dwClient.GetProjectAssignments(ctx, string(offer.CompanyID), offer.ExternalReference)
	if err != nil {
		s.logger.Error("failed to fetch assignments from datawarehouse",
			zap.Error(err),
			zap.String("offer_id", offerID.String()),
			zap.String("external_reference", offer.ExternalReference),
			zap.String("company_id", string(offer.CompanyID)))
		result.Error = fmt.Sprintf("datawarehouse query failed: %v", err)
		return result, nil
	}

	s.logger.Info("fetched assignments from datawarehouse",
		zap.String("offer_id", offerID.String()),
		zap.String("external_reference", offer.ExternalReference),
		zap.Int("count", len(erpAssignments)))

	// Prepare upsert inputs
	inputs := make([]repository.AssignmentUpsertInput, 0, len(erpAssignments))
	validDWIDs := make([]int64, 0, len(erpAssignments))

	for _, erp := range erpAssignments {
		// Convert ProgressPercent (float64) to ProgressID (int) for storage
		var progressID *int
		if erp.ProgressPercent != nil {
			p := int(*erp.ProgressPercent)
			progressID = &p
		}

		inputs = append(inputs, repository.AssignmentUpsertInput{
			DWAssignmentID:   erp.AssignmentID,
			DWProjectID:      erp.ProjectID,
			OfferID:          &offer.ID,
			CompanyID:        offer.CompanyID,
			AssignmentNumber: erp.AssignmentNumber,
			Description:      erp.Description,
			FixedPriceAmount: erp.FixedPriceAmount,
			StatusID:         erp.StatusID,
			ProgressID:       progressID,
			RawData:          erp.RawData,
		})
		validDWIDs = append(validDWIDs, erp.AssignmentID)
	}

	// Upsert assignments
	created, updated, err := s.assignmentRepo.UpsertFromDW(ctx, inputs)
	if err != nil {
		s.logger.Error("failed to upsert assignments",
			zap.Error(err),
			zap.String("offer_id", offerID.String()))
		result.Error = fmt.Sprintf("upsert failed: %v", err)
		return result, nil
	}

	// Delete stale assignments (ones in our DB but not in DW)
	deleted, err := s.assignmentRepo.DeleteStaleByOfferID(ctx, offer.ID, validDWIDs)
	if err != nil {
		s.logger.Error("failed to delete stale assignments",
			zap.Error(err),
			zap.String("offer_id", offerID.String()))
		// Don't fail the whole sync, just log
	}

	result.Synced = len(erpAssignments)
	result.Created = created
	result.Updated = updated
	result.Deleted = deleted

	// Calculate and update the total FixedPriceAmount on the offer
	totalFixedPrice, _, _, err := s.assignmentRepo.GetAggregatedByOfferID(ctx, offer.ID)
	if err != nil {
		s.logger.Error("failed to get aggregated assignment data",
			zap.Error(err),
			zap.String("offer_id", offerID.String()))
	} else {
		if err := s.offerRepo.UpdateDWTotalFixedPrice(ctx, offer.ID, totalFixedPrice); err != nil {
			s.logger.Error("failed to update offer dw_total_fixed_price",
				zap.Error(err),
				zap.String("offer_id", offerID.String()))
		}
	}

	// Log activity for the sync
	activity := &domain.Activity{
		TargetType:   domain.ActivityTargetOffer,
		TargetID:     offer.ID,
		TargetName:   offer.Title,
		Title:        "Oppdrag synkronisert",
		Body:         fmt.Sprintf("Synkroniserte %d oppdrag fra datavarehus (%d nye, %d oppdatert, %d slettet)", result.Synced, result.Created, result.Updated, result.Deleted),
		ActivityType: domain.ActivityTypeSystem,
		Status:       domain.ActivityStatusCompleted,
		OccurredAt:   time.Now(),
		CreatorName:  "System",
		CompanyID:    &offer.CompanyID,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		s.logger.Warn("failed to log assignment sync activity",
			zap.Error(err),
			zap.String("offer_id", offerID.String()))
		// Don't fail the sync for activity logging failure
	}

	s.logger.Info("assignment sync completed",
		zap.String("offer_id", offerID.String()),
		zap.Int("synced", result.Synced),
		zap.Int("created", result.Created),
		zap.Int("updated", result.Updated),
		zap.Int("deleted", result.Deleted))

	return result, nil
}

// GetAssignmentsForOffer returns all assignments for an offer
func (s *AssignmentService) GetAssignmentsForOffer(ctx context.Context, offerID uuid.UUID) ([]domain.AssignmentDTO, error) {
	// Verify offer exists
	_, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("get offer: %w", err)
	}

	assignments, err := s.assignmentRepo.GetByOfferID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("get assignments: %w", err)
	}

	dtos := make([]domain.AssignmentDTO, 0, len(assignments))
	for _, a := range assignments {
		dtos = append(dtos, mapper.AssignmentToDTO(&a))
	}

	return dtos, nil
}

// GetAssignmentSummaryForOffer returns aggregated assignment data with comparison to offer value
func (s *AssignmentService) GetAssignmentSummaryForOffer(ctx context.Context, offerID uuid.UUID) (*domain.AssignmentSummaryDTO, error) {
	// Get the offer for value comparison
	offer, err := s.offerRepo.GetByID(ctx, offerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOfferNotFound
		}
		return nil, fmt.Errorf("get offer: %w", err)
	}

	// Get aggregated assignment data
	totalFixedPriceAmount, count, lastSyncedAt, err := s.assignmentRepo.GetAggregatedByOfferID(ctx, offerID)
	if err != nil {
		return nil, fmt.Errorf("aggregate assignments: %w", err)
	}

	// Calculate difference
	difference := offer.Value - totalFixedPriceAmount
	var differencePercent float64
	if offer.Value > 0 {
		differencePercent = (difference / offer.Value) * 100
	}

	// Format lastSyncedAt
	var lastSyncedAtStr *string
	if lastSyncedAt != nil {
		t := lastSyncedAt.UTC().Format(time.RFC3339)
		lastSyncedAtStr = &t
	}

	return &domain.AssignmentSummaryDTO{
		OfferID:               offerID,
		OfferValue:            offer.Value,
		TotalFixedPriceAmount: totalFixedPriceAmount,
		AssignmentCount:       count,
		Difference:            difference,
		DifferencePercent:     differencePercent,
		LastSyncedAt:          lastSyncedAtStr,
	}, nil
}

// SyncAllAssignmentsFromDataWarehouse syncs assignments for all offers in "order" phase
// that have an external_reference. Uses the same maxAge window as offer sync.
// Continues on error for individual offers.
// Returns counts for successfully synced and failed offers.
func (s *AssignmentService) SyncAllAssignmentsFromDataWarehouse(ctx context.Context) (synced int, failed int, err error) {
	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping bulk assignment sync")
		return 0, 0, ErrDWClientNotAvailable
	}

	// Get "order" phase offers with external_reference (same as offer sync)
	// Uses 55 minute maxAge to match hourly cron timing
	maxAge := 55 * time.Minute
	offers, err := s.offerRepo.GetOffersNeedingDWSync(ctx, maxAge)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get offers for assignment sync: %w", err)
	}

	s.logger.Info("starting bulk assignment sync",
		zap.Int("offer_count", len(offers)))

	for _, offer := range offers {
		_, syncErr := s.syncAssignmentsForOfferInternal(ctx, &offer)
		if syncErr != nil {
			s.logger.Warn("failed to sync assignments for offer",
				zap.Error(syncErr),
				zap.String("offer_id", offer.ID.String()))
			failed++
			continue
		}
		synced++
	}

	s.logger.Info("completed bulk assignment sync",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.Int("total", len(offers)))

	return synced, failed, nil
}

// ForceSyncAllAssignmentsFromDataWarehouse syncs assignments for ALL offers in "order" phase
// with external_reference, regardless of when they were last synced. Used by admin force-sync endpoint.
func (s *AssignmentService) ForceSyncAllAssignmentsFromDataWarehouse(ctx context.Context) (synced int, failed int, err error) {
	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping force assignment sync")
		return 0, 0, ErrDWClientNotAvailable
	}

	// Get ALL "order" phase offers with external_reference (no maxAge filter)
	offers, err := s.offerRepo.GetAllOffersForDWSync(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get offers for force assignment sync: %w", err)
	}

	s.logger.Info("starting force assignment sync (admin)",
		zap.Int("offer_count", len(offers)))

	for _, offer := range offers {
		_, syncErr := s.syncAssignmentsForOfferInternal(ctx, &offer)
		if syncErr != nil {
			s.logger.Warn("failed to sync assignments for offer (force sync)",
				zap.Error(syncErr),
				zap.String("offer_id", offer.ID.String()))
			failed++
			continue
		}
		synced++
	}

	s.logger.Info("completed force assignment sync (admin)",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.Int("total", len(offers)))

	return synced, failed, nil
}

// SyncStaleAssignmentsFromDataWarehouse syncs assignments only for offers that are stale.
// This is used for startup sync to catch up on offers that haven't been synced recently.
// Returns counts for successfully synced and failed offers.
func (s *AssignmentService) SyncStaleAssignmentsFromDataWarehouse(ctx context.Context, maxAge time.Duration) (synced int, failed int, err error) {
	// Check if data warehouse client is available
	if s.dwClient == nil || !s.dwClient.IsEnabled() {
		s.logger.Info("data warehouse not available, skipping stale assignment sync")
		return 0, 0, ErrDWClientNotAvailable
	}

	// Get offers that need syncing (null or older than maxAge)
	offers, err := s.offerRepo.GetOffersNeedingDWSync(ctx, maxAge)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get offers needing assignment sync: %w", err)
	}

	if len(offers) == 0 {
		s.logger.Info("no stale offers for assignment sync")
		return 0, 0, nil
	}

	s.logger.Info("starting stale assignment sync",
		zap.Int("offer_count", len(offers)),
		zap.Duration("max_age", maxAge))

	for _, offer := range offers {
		_, syncErr := s.syncAssignmentsForOfferInternal(ctx, &offer)
		if syncErr != nil {
			s.logger.Warn("failed to sync assignments for stale offer",
				zap.Error(syncErr),
				zap.String("offer_id", offer.ID.String()))
			failed++
			continue
		}
		synced++
	}

	s.logger.Info("completed stale assignment sync",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.Int("total", len(offers)))

	return synced, failed, nil
}

// syncAssignmentsForOfferInternal is an internal method that syncs assignments for a given offer.
// It does not log activities to avoid duplicate activity entries during bulk sync.
func (s *AssignmentService) syncAssignmentsForOfferInternal(ctx context.Context, offer *domain.Offer) (*domain.AssignmentSyncResultDTO, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	result := &domain.AssignmentSyncResultDTO{
		OfferID:           offer.ID,
		ExternalReference: offer.ExternalReference,
		CompanyID:         offer.CompanyID,
		SyncedAt:          now,
	}

	// Validate external reference exists
	if offer.ExternalReference == "" {
		return result, nil // Skip silently, not an error
	}

	// Fetch assignments from datawarehouse
	erpAssignments, err := s.dwClient.GetProjectAssignments(ctx, string(offer.CompanyID), offer.ExternalReference)
	if err != nil {
		s.logger.Warn("failed to fetch assignments from datawarehouse",
			zap.Error(err),
			zap.String("offer_id", offer.ID.String()),
			zap.String("external_reference", offer.ExternalReference))
		return nil, err
	}

	// Prepare upsert inputs
	inputs := make([]repository.AssignmentUpsertInput, 0, len(erpAssignments))
	validDWIDs := make([]int64, 0, len(erpAssignments))

	for _, erp := range erpAssignments {
		// Convert ProgressPercent (float64) to ProgressID (int) for storage
		var progressID *int
		if erp.ProgressPercent != nil {
			p := int(*erp.ProgressPercent)
			progressID = &p
		}

		inputs = append(inputs, repository.AssignmentUpsertInput{
			DWAssignmentID:   erp.AssignmentID,
			DWProjectID:      erp.ProjectID,
			OfferID:          &offer.ID,
			CompanyID:        offer.CompanyID,
			AssignmentNumber: erp.AssignmentNumber,
			Description:      erp.Description,
			FixedPriceAmount: erp.FixedPriceAmount,
			StatusID:         erp.StatusID,
			ProgressID:       progressID,
			RawData:          erp.RawData,
		})
		validDWIDs = append(validDWIDs, erp.AssignmentID)
	}

	// Upsert assignments
	created, updated, err := s.assignmentRepo.UpsertFromDW(ctx, inputs)
	if err != nil {
		return nil, fmt.Errorf("upsert assignments: %w", err)
	}

	// Delete stale assignments
	deleted, _ := s.assignmentRepo.DeleteStaleByOfferID(ctx, offer.ID, validDWIDs)

	result.Synced = len(erpAssignments)
	result.Created = created
	result.Updated = updated
	result.Deleted = deleted

	// Calculate and update the total FixedPriceAmount on the offer
	totalFixedPrice, _, _, err := s.assignmentRepo.GetAggregatedByOfferID(ctx, offer.ID)
	if err == nil {
		_ = s.offerRepo.UpdateDWTotalFixedPrice(ctx, offer.ID, totalFixedPrice)
	}

	return result, nil
}

