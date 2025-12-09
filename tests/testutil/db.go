package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/straye-as/relation-api/internal/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// SetupTestDB creates a connection to the test PostgreSQL database
// It uses environment variables or falls back to docker-compose defaults
func SetupTestDB(t *testing.T) *gorm.DB {
	host := getEnvOrDefault("DATABASE_HOST", "localhost")
	port := getEnvOrDefault("DATABASE_PORT", "5432")
	user := getEnvOrDefault("DATABASE_USER", "relation_user")
	password := getEnvOrDefault("DATABASE_PASSWORD", "relation_password")
	dbname := getEnvOrDefault("DATABASE_NAME", "relation")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err, "Failed to connect to test database. Ensure PostgreSQL is running.")

	// Ensure test companies exist
	EnsureTestCompanies(t, db)

	return db
}

// CleanupTestData cleans up test data from all tables
// This should be called after tests to ensure a clean state
func CleanupTestData(t *testing.T, db *gorm.DB) {
	// Delete in order to respect foreign key constraints
	tables := []string{
		"deal_stage_history",
		"deals",
		"notifications",
		"activities",
		"offer_items",
		"files",
		"offers",
		"projects",
		"contact_relationships",
		"contacts",
		"customers",
		"number_sequences",
	}

	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id IS NOT NULL", table)).Error
		if err != nil {
			// Table might not exist, that's ok
			t.Logf("Note: Could not clean table %s: %v", table, err)
		}
	}
}

// CreateTestCustomer creates a test customer and returns it
// Uses Omit to skip association handling that would fail due to model relationship issues
func CreateTestCustomer(t *testing.T, db *gorm.DB, name string) *domain.Customer {
	// Use the last 9 digits of nanoseconds to fit in varchar(20)
	orgNum := fmt.Sprintf("%09d", randomInt()%1000000000)
	customer := &domain.Customer{
		Name:      name,
		OrgNumber: orgNum,
		Email:     "test@example.com",
		Phone:     "12345678",
		Country:   "Norway",
	}
	// Omit associations to avoid GORM trying to validate/create related records
	err := db.Omit(clause.Associations).Create(customer).Error
	require.NoError(t, err)
	return customer
}

// randomInt returns a unique integer for test data
func randomInt() int64 {
	return time.Now().UnixNano()
}

// EnsureTestCompanies creates test company records if they don't exist
func EnsureTestCompanies(t *testing.T, db *gorm.DB) {
	companies := []struct {
		id        string
		name      string
		shortName string
	}{
		{string(domain.CompanyGruppen), "Straye Gruppen", "Gruppen"},
		{string(domain.CompanyStalbygg), "Stålbygg", "Stålbygg"},
		{string(domain.CompanyHybridbygg), "Hybridbygg", "Hybridbygg"},
		{string(domain.CompanyIndustri), "Industri", "Industri"},
		{string(domain.CompanyTak), "Tak", "Tak"},
		{string(domain.CompanyMontasje), "Montasje", "Montasje"},
	}

	for _, c := range companies {
		// Try to insert, ignore if already exists
		err := db.Exec(`
			INSERT INTO companies (id, name, short_name, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, true, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`, c.id, c.name, c.shortName).Error
		if err != nil {
			t.Logf("Note: Could not insert company %s: %v", c.id, err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
