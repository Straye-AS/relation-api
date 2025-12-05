package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/straye-as/relation-api/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// Default retry configuration
	defaultMaxRetries     = 5
	defaultInitialBackoff = 1 * time.Second
	defaultMaxBackoff     = 30 * time.Second
	defaultBackoffFactor  = 2.0
)

// RetryConfig holds configuration for connection retry logic
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     defaultMaxRetries,
		InitialBackoff: defaultInitialBackoff,
		MaxBackoff:     defaultMaxBackoff,
		BackoffFactor:  defaultBackoffFactor,
	}
}

// NewDatabase creates a new database connection with retry logic and structured logging
func NewDatabase(cfg *config.DatabaseConfig, log *zap.Logger) (*gorm.DB, error) {
	return NewDatabaseWithRetry(cfg, log, DefaultRetryConfig())
}

// NewDatabaseWithRetry creates a new database connection with custom retry configuration
func NewDatabaseWithRetry(cfg *config.DatabaseConfig, log *zap.Logger, retryConfig *RetryConfig) (*gorm.DB, error) {
	dsn := cfg.ConnectionString()

	log.Info("Initiating database connection",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
		zap.String("user", cfg.User),
		zap.String("ssl_mode", cfg.SSLMode),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("conn_max_lifetime_seconds", cfg.ConnMaxLifetime),
	)

	var db *gorm.DB
	var err error
	backoff := retryConfig.InitialBackoff

	for attempt := 1; attempt <= retryConfig.MaxRetries; attempt++ {
		log.Info("Attempting database connection",
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", retryConfig.MaxRetries),
		)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})

		if err == nil {
			// Connection opened successfully, now configure and test it
			sqlDB, poolErr := db.DB()
			if poolErr != nil {
				log.Error("Failed to get underlying database instance",
					zap.Error(poolErr),
					zap.Int("attempt", attempt),
				)
				err = fmt.Errorf("failed to get database instance: %w", poolErr)
			} else {
				// Configure connection pool
				sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
				sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
				sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetimeDuration())

				// Test connection with ping
				if pingErr := sqlDB.Ping(); pingErr != nil {
					log.Warn("Database ping failed",
						zap.Error(pingErr),
						zap.Int("attempt", attempt),
					)
					err = fmt.Errorf("failed to ping database: %w", pingErr)
				} else {
					// Connection successful
					log.Info("Database connection established successfully",
						zap.Int("attempts_taken", attempt),
						zap.String("host", cfg.Host),
						zap.String("database", cfg.Name),
					)
					return db, nil
				}
			}
		} else {
			log.Warn("Database connection attempt failed",
				zap.Error(err),
				zap.Int("attempt", attempt),
				zap.Duration("next_retry_in", backoff),
			)
		}

		// If this wasn't the last attempt, wait before retrying
		if attempt < retryConfig.MaxRetries {
			log.Info("Waiting before retry",
				zap.Duration("backoff", backoff),
			)
			time.Sleep(backoff)

			// Calculate next backoff with exponential increase, capped at max
			backoff = time.Duration(float64(backoff) * retryConfig.BackoffFactor)
			if backoff > retryConfig.MaxBackoff {
				backoff = retryConfig.MaxBackoff
			}
		}
	}

	log.Error("All database connection attempts failed",
		zap.Int("total_attempts", retryConfig.MaxRetries),
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Name),
		zap.Error(err),
	)

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", retryConfig.MaxRetries, err)
}

// HealthCheck performs a database health check
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// HealthCheckWithStats performs a health check and returns connection pool statistics
func HealthCheckWithStats(db *gorm.DB) (*sql.DBStats, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	stats := sqlDB.Stats()
	return &stats, nil
}

// Close gracefully closes the database connection
func Close(db *gorm.DB, log *zap.Logger) error {
	log.Info("Closing database connection")

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("Failed to get database instance for closing", zap.Error(err))
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Error("Failed to close database connection", zap.Error(err))
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	log.Info("Database connection closed successfully")
	return nil
}
