package database

import (
	"fmt"
	"time"

	"github.com/straye-as/relation-api/internal/config"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.ConnectionString()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetimeDuration())

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// AutoMigrate runs automatic migrations (for development only)
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.Customer{},
		&domain.Contact{},
		&domain.Project{},
		&domain.Offer{},
		&domain.OfferItem{},
		&domain.Activity{},
		&domain.File{},
		&domain.Notification{},
		&domain.User{},
	)
}
