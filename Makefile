run .PHONY: help build run run-dev api dev test test-all test-coverage test-integration lint docker-build docker-up docker-down docker-logs migrate-up migrate-down migrate-status migrate-create clean deps format security swagger seed test-db-up test-db-down test-db-reset

GO ?= go
BIN_DIR ?= bin
APP_CMD ?= ./cmd/api
MIGRATE_CMD ?= ./cmd/migrate
AIR_BIN := $(shell command -v air 2>/dev/null)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application binaries
	$(GO) build -o $(BIN_DIR)/api $(APP_CMD)
	$(GO) build -o $(BIN_DIR)/migrate $(MIGRATE_CMD)

run: swagger ## Run the application locally (single run, regenerates Swagger docs)
	$(GO) run $(APP_CMD)

dev: ## Run the API with Air hot reload
ifeq ($(strip $(AIR_BIN)),)
	$(error Air is not installed. Install with `go install github.com/air-verse/air@latest`)
endif
	$(AIR_BIN)

run-dev: dev ## Alias for `make dev`

api: run ## Alias for `make run`

test: ## Run fast unit tests (no database required)
	$(GO) test -v -race -cover ./tests/auth ./tests/mapper ./tests/domain

test-all: ## Run all tests including integration tests (uses test database on port 5433)
	$(GO) test -v -race -cover ./tests/auth ./tests/mapper ./tests/domain
	DATABASE_PORT=5433 DATABASE_NAME=relation_test $(GO) test -v -p 1 -cover ./tests/repository ./tests/service ./tests/handler ./tests/middleware

test-coverage: ## Run tests with coverage report
	$(GO) test -v -race -coverprofile=coverage.out ./tests/auth ./tests/mapper ./tests/domain
	$(GO) tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests (uses test database on port 5433)
	DATABASE_PORT=5433 DATABASE_NAME=relation_test $(GO) test -v -p 1 -cover ./tests/repository ./tests/service

lint: ## Run linters
	golangci-lint run ./...

swagger: ## Generate Swagger documentation
	swag init -g cmd/api/main.go -o ./docs

docker-build: swagger ## Build Docker images (regenerates Swagger docs first)
	docker compose build

docker-up: swagger ## Start all services with Docker Compose (regenerates Swagger docs first, loads .env)
	docker compose --env-file .env up -d

docker-down: ## Stop all services
	docker compose down

docker-logs: ## View logs from all services
	docker compose logs -f

migrate-up: ## Run database migrations up
	$(GO) run $(MIGRATE_CMD) up

migrate-down: ## Run database migration down
	$(GO) run $(MIGRATE_CMD) down

migrate-status: ## Check migration status
	$(GO) run $(MIGRATE_CMD) status

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	$(GO) run $(MIGRATE_CMD) create $(name)

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf coverage.out coverage.html
	rm -rf storage/*

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

format: ## Format code
	$(GO) fmt ./...
	goimports -w .

security: ## Run security checks
	gosec ./...

seed: ## Load seed data (LOCAL DEVELOPMENT ONLY - will not run in staging/production)
	@if [ "$$APP_ENVIRONMENT" = "staging" ] || [ "$$APP_ENVIRONMENT" = "production" ] || [ "$$APP_ENVIRONMENT" = "prod" ]; then \
		echo "ERROR: Seed data cannot be loaded in staging or production environments!"; \
		exit 1; \
	fi
	@echo "Loading seed data into local database..."
	PGPASSWORD=relation_password psql -h localhost -U relation_user -d relation -f testdata/seed.sql
	@echo "Seed data loaded successfully!"

# Test Database Management (separate from development database)
test-db-up: ## Start test database (PostgreSQL on port 5433)
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for test database to be ready..."
	@sleep 3
	@echo "Test database is running on port 5433"

test-db-down: ## Stop test database
	docker compose -f docker-compose.test.yml down

test-db-reset: ## Reset test database (stop, remove volume, start fresh)
	docker compose -f docker-compose.test.yml down -v
	docker compose -f docker-compose.test.yml up -d
	@echo "Waiting for test database to be ready..."
	@sleep 3
	@echo "Test database has been reset"

.DEFAULT_GOAL := help
