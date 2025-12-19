package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DWSyncJobName is the name of the data warehouse sync job
const DWSyncJobName = "dw_sync"

// DefaultStaleMaxAge is the default maximum age before an offer is considered stale (1 hour)
const DefaultStaleMaxAge = time.Hour

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

// DWSyncJob runs the data warehouse sync for all offers with external_reference.
type DWSyncJob struct {
	service OfferSyncService
	logger  *zap.Logger
	timeout time.Duration
}

// NewDWSyncJob creates a new data warehouse sync job.
// The timeout controls how long the sync operation is allowed to run.
func NewDWSyncJob(service OfferSyncService, logger *zap.Logger, timeout time.Duration) *DWSyncJob {
	return &DWSyncJob{
		service: service,
		logger:  logger,
		timeout: timeout,
	}
}

// Run executes the data warehouse sync job.
// This is called by the scheduler according to the cron expression.
func (j *DWSyncJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()

	start := time.Now()
	j.logger.Info("starting data warehouse sync job")

	synced, failed, err := j.service.SyncAllOffersFromDataWarehouse(ctx)

	duration := time.Since(start)

	if err != nil {
		j.logger.Error("data warehouse sync job failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return
	}

	j.logger.Info("data warehouse sync job completed",
		zap.Int("synced", synced),
		zap.Int("failed", failed),
		zap.Int("total", synced+failed),
		zap.Duration("duration", duration))
}

// RunStartupSync runs a sync for stale offers on startup.
// This syncs offers where dw_last_synced_at is NULL or older than maxAge.
// Returns the number of synced and failed offers.
func (j *DWSyncJob) RunStartupSync(maxAge time.Duration) (synced int, failed int) {
	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()

	start := time.Now()
	j.logger.Info("starting data warehouse startup sync for stale offers",
		zap.Duration("max_age", maxAge))

	synced, failed, err := j.service.SyncStaleOffersFromDataWarehouse(ctx, maxAge)

	duration := time.Since(start)

	if err != nil {
		j.logger.Error("data warehouse startup sync failed",
			zap.Error(err),
			zap.Duration("duration", duration))
		return 0, 0
	}

	if synced > 0 || failed > 0 {
		j.logger.Info("data warehouse startup sync completed",
			zap.Int("synced", synced),
			zap.Int("failed", failed),
			zap.Duration("duration", duration))
	}

	return synced, failed
}

// RegisterDWSyncJob registers the data warehouse sync job with the scheduler.
// The cronExpr should be a valid cron expression (e.g., "0 15 * * * *" for 15 minutes past every hour).
// If runStartupSync is true, it will also run a sync for stale offers (null or > 1 hour old) immediately
// in a background goroutine so it doesn't block API startup.
func RegisterDWSyncJob(scheduler *Scheduler, service OfferSyncService, logger *zap.Logger, cronExpr string, timeout time.Duration, runStartupSync bool) error {
	job := NewDWSyncJob(service, logger, timeout)

	// Run startup sync for stale offers asynchronously if requested
	if runStartupSync {
		go job.RunStartupSync(DefaultStaleMaxAge)
	}

	return scheduler.AddJob(DWSyncJobName, cronExpr, job.Run)
}
