package jobs

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DWSyncJobName is the name of the data warehouse sync job
const DWSyncJobName = "dw_sync"

// OfferSyncService defines the interface for syncing offers from the data warehouse.
// This interface allows the job to call the service without importing the service package directly.
type OfferSyncService interface {
	// SyncAllOffersFromDataWarehouse syncs all offers with external_reference from the data warehouse.
	// Returns counts for successfully synced and failed offers.
	SyncAllOffersFromDataWarehouse(ctx context.Context) (synced int, failed int, err error)
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

// RegisterDWSyncJob registers the data warehouse sync job with the scheduler.
// The cronExpr should be a valid cron expression (e.g., "0 15 * * * *" for 15 minutes past every hour).
func RegisterDWSyncJob(scheduler *Scheduler, service OfferSyncService, logger *zap.Logger, cronExpr string, timeout time.Duration) error {
	job := NewDWSyncJob(service, logger, timeout)
	return scheduler.AddJob(DWSyncJobName, cronExpr, job.Run)
}
