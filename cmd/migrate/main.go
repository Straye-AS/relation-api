package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/straye-as/relation-api/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Open database connection
	db, err := sql.Open("postgres", cfg.Database.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Get command and arguments
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("usage: migrate [up|down|status|version|create]")
	}

	command := args[0]
	arguments := args[1:]

	// Set migrations directory
	migrationsDir := "./migrations"

	// Run goose command
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	switch command {
	case "up":
		if err := goose.Up(db, migrationsDir); err != nil {
			return fmt.Errorf("failed to run up migrations: %w", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := goose.Down(db, migrationsDir); err != nil {
			return fmt.Errorf("failed to run down migration: %w", err)
		}
		fmt.Println("Migration rolled back successfully")

	case "status":
		if err := goose.Status(db, migrationsDir); err != nil {
			return fmt.Errorf("failed to get migration status: %w", err)
		}

	case "version":
		if err := goose.Version(db, migrationsDir); err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}

	case "create":
		if len(arguments) == 0 {
			return fmt.Errorf("create requires a migration name")
		}
		if err := goose.Create(db, migrationsDir, arguments[0], "sql"); err != nil {
			return fmt.Errorf("failed to create migration: %w", err)
		}
		fmt.Printf("Migration created: %s\n", arguments[0])

	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}
