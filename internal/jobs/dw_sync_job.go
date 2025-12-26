package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DWSyncJobName is the name of the data warehouse sync job
const DWSyncJobName = "dw_sync"

// DefaultStaleMaxAge is the default maximum age before an offer is considered stale (55 minutes)
// This is slightly less than 1 hour to allow buffer for hourly cron timing variations
const DefaultStaleMaxAge = 55 * time.Minute

// OfferSyncService defines the interface for syncing offers from the data warehouse.
// This interface allows the job to call the service without importing the service package directly.
type OfferSyncService interface {
	// SyncAllOffersFromDataWarehouse syncs all offers with external_reference from the data warehouse.
	// Returns counts for successfully synced and failed offers.
	SyncAllOffersFromDataWarehouse(ctx context.Context) (synced int, failed int, err error)

	// SyncStaleOffersFromDataWarehouse syncs only offers that are stale (never synced or older than maxAge).
	// Returns counts for successfully synced and failed offers.
	SyncStaleOffersFromDataWarehouse(ctx context.Context, maxAge time.Duration) (synced int, failed int, err error)
}

// AssignmentSyncService defines the interface for syncing assignments from the data warehouse.
type AssignmentSyncService interface {
	// SyncAllAssignmentsFromDataWarehouse syncs assignments for all offers in "order" phase.
	// Returns counts for successfully synced and failed offers.
	SyncAllAssignmentsFromDataWarehouse(ctx context.Context) (synced int, failed int, err error)

	// SyncStaleAssignmentsFromDataWarehouse syncs assignments only for offers that are stale.
	// Returns counts for successfully synced and failed offers.
	SyncStaleAssignmentsFromDataWarehouse(ctx context.Context, maxAge time.Duration) (synced int, failed int, err error)
}

// DWSyncJob runs the data warehouse sync for all offers with external_reference
// and their associated assignments.
type DWSyncJob struct {
	offerService      OfferSyncService
	assignmentService AssignmentSyncService
	logger            *zap.Logger
	timeout           time.Duration
}

// NewDWSyncJob creates a new data warehouse sync job.
// The timeout controls how long the sync operation is allowed to run.
func NewDWSyncJob(offerService OfferSyncService, assignmentService AssignmentSyncService, logger *zap.Logger, timeout time.Duration) *DWSyncJob {
	return &DWSyncJob{
		offerService:      offerService,
		assignmentService: assignmentService,
		logger:            logger,
		timeout:           timeout,
	}
}

// Run executes the data warehouse sync job.
// This is called by the scheduler according to the cron expression.
// It syncs both offer financials and assignments.
func (j *DWSyncJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()

	start := time.Now()
	j.logger.Info("starting data warehouse sync job")

	// Sync offer financials
	offersSynced, offersFailed, err := j.offerService.SyncAllOffersFromDataWarehouse(ctx)
	if err != nil {
		j.logger.Error("data warehouse offer sync failed",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		// Continue with assignment sync even if offer sync fails
	}

	// Sync assignments
	var assignmentsSynced, assignmentsFailed int
	if j.assignmentService != nil {
		assignmentsSynced, assignmentsFailed, err = j.assignmentService.SyncAllAssignmentsFromDataWarehouse(ctx)
		if err != nil {
			j.logger.Error("data warehouse assignment sync failed",
				zap.Error(err),
				zap.Duration("duration", time.Since(start)))
		}
	}

	duration := time.Since(start)

	j.logger.Info("data warehouse sync job completed",
		zap.Int("offers_synced", offersSynced),
		zap.Int("offers_failed", offersFailed),
		zap.Int("assignments_synced", assignmentsSynced),
		zap.Int("assignments_failed", assignmentsFailed),
		zap.Duration("duration", duration))
}

// RunStartupSync runs a sync for stale offers and assignments on startup.
// This syncs offers and assignments where dw_last_synced_at is NULL or older than maxAge.
// Returns the number of synced and failed offers (assignments are synced alongside).
func (j *DWSyncJob) RunStartupSync(maxAge time.Duration) (synced int, failed int) {
	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()

	start := time.Now()
	j.logger.Info("starting data warehouse startup sync for stale offers and assignments",
		zap.Duration("max_age", maxAge))

	// Sync stale offer financials
	offersSynced, offersFailed, err := j.offerService.SyncStaleOffersFromDataWarehouse(ctx, maxAge)
	if err != nil {
		j.logger.Error("data warehouse startup offer sync failed",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)))
		// Continue with assignment sync
	}

	// Sync stale assignments
	var assignmentsSynced, assignmentsFailed int
	if j.assignmentService != nil {
		assignmentsSynced, assignmentsFailed, err = j.assignmentService.SyncStaleAssignmentsFromDataWarehouse(ctx, maxAge)
		if err != nil {
			j.logger.Error("data warehouse startup assignment sync failed",
				zap.Error(err),
				zap.Duration("duration", time.Since(start)))
		}
	}

	duration := time.Since(start)

	if offersSynced > 0 || offersFailed > 0 || assignmentsSynced > 0 || assignmentsFailed > 0 {
		j.logger.Info("data warehouse startup sync completed",
			zap.Int("offers_synced", offersSynced),
			zap.Int("offers_failed", offersFailed),
			zap.Int("assignments_synced", assignmentsSynced),
			zap.Int("assignments_failed", assignmentsFailed),
			zap.Duration("duration", duration))
	}

	return offersSynced, offersFailed
}

// RegisterDWSyncJob registers the data warehouse sync job with the scheduler.
// The cronExpr should be a valid cron expression (e.g., "0 15 * * * *" for 15 minutes past every hour).
// If runStartupSync is true, it will also run a sync for stale offers and assignments (null or > 1 hour old)
// immediately in a background goroutine so it doesn't block API startup.
// assignmentService can be nil if assignment syncing is not needed.
func RegisterDWSyncJob(scheduler *Scheduler, offerService OfferSyncService, assignmentService AssignmentSyncService, logger *zap.Logger, cronExpr string, timeout time.Duration, runStartupSync bool) error {
	job := NewDWSyncJob(offerService, assignmentService, logger, timeout)

	// Run startup sync for stale offers and assignments asynchronously if requested
	if runStartupSync {
		go job.RunStartupSync(DefaultStaleMaxAge)
	}

	return scheduler.AddJob(DWSyncJobName, cronExpr, job.Run)
}
