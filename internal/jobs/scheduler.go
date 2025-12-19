// Package jobs provides background job scheduling for the Relation API.
// It uses robfig/cron for cron-based job scheduling.
package jobs

import (
	"context"
	"fmt"
	"sync"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler manages background jobs using cron scheduling.
type Scheduler struct {
	cron   *cron.Cron
	logger *zap.Logger
	mu     sync.Mutex
	jobs   map[string]cron.EntryID
}

// NewScheduler creates a new job scheduler with the given logger.
func NewScheduler(logger *zap.Logger) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds(), cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger),
			cron.Recover(cron.DefaultLogger),
		)),
		logger: logger,
		jobs:   make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler. Jobs added before this call will begin running.
func (s *Scheduler) Start() {
	s.logger.Info("starting job scheduler")
	s.cron.Start()
}

// Stop gracefully stops the scheduler. Running jobs will complete.
func (s *Scheduler) Stop() context.Context {
	s.logger.Info("stopping job scheduler")
	return s.cron.Stop()
}

// AddJob adds a job with the given name and cron expression.
// The cronExpr follows standard cron format with optional seconds field.
// Examples:
//   - "0 15 * * * *" - At minute 15 of every hour (with seconds field)
//   - "15 * * * *"   - At minute 15 of every hour (standard 5-field format)
//   - "@hourly"      - At minute 0 of every hour
//   - "@every 1h"    - Every hour
func (s *Scheduler) AddJob(name string, cronExpr string, job func()) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if job already exists
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job %s already exists", name)
	}

	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.logger.Info("running scheduled job",
			zap.String("job_name", name))
		job()
		s.logger.Info("completed scheduled job",
			zap.String("job_name", name))
	})
	if err != nil {
		return fmt.Errorf("failed to add job %s: %w", name, err)
	}

	s.jobs[name] = entryID
	s.logger.Info("added scheduled job",
		zap.String("job_name", name),
		zap.String("cron_expr", cronExpr))

	return nil
}

// RemoveJob removes a job by name.
func (s *Scheduler) RemoveJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entryID, exists := s.jobs[name]
	if !exists {
		return fmt.Errorf("job %s not found", name)
	}

	s.cron.Remove(entryID)
	delete(s.jobs, name)

	s.logger.Info("removed scheduled job",
		zap.String("job_name", name))

	return nil
}

// GetJobNames returns the names of all registered jobs.
func (s *Scheduler) GetJobNames() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	names := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		names = append(names, name)
	}
	return names
}
