# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Straye Relation API is a production-grade REST API for Customer Relationship Management (CRM), providing customer, project, and sales pipeline management for Straye Gruppen's construction companies (Stålbygg, Hybridbygg, Industri, Tak, Montasje).

**Tech Stack**: Go 1.21+, PostgreSQL 15+, Chi router, GORM ORM, Zap logger, Azure AD JWT + API Key auth, Docker

## Common Commands

### Development
```bash
make run              # Single run
make dev              # Hot reload with Air (install: go install github.com/air-verse/air@latest)
make build            # Build binaries to bin/
make deps             # Download and tidy dependencies
```

### Testing
```bash
make test             # Fast unit tests (auth, mapper - no DB)
make test-all         # All tests including integration
make test-integration # Integration tests (requires DB)
make test-coverage    # Generate coverage.html
```

### Database
```bash
make migrate-up       # Run pending migrations
make migrate-down     # Rollback one migration
make migrate-status   # Check migration status
make migrate-create name=add_field  # Create new migration
```

### Docker
```bash
make docker-up        # Start PostgreSQL + API via docker-compose
make docker-down      # Stop all services
make docker-logs      # Follow logs
```

### Code Quality
```bash
make format           # go fmt + goimports
make lint             # golangci-lint
make security         # gosec security scan
make swagger          # Regenerate OpenAPI docs
```

## Architecture

### Clean Architecture Pattern (3-layer)
```
HTTP Request → Handler → Service → Repository → Database
             ↓         ↓          ↓
             DTO       Business   Domain Model
                       Logic      (GORM)
```

**Critical**: Always follow this flow. Handlers parse/validate requests and return DTOs. Services contain business logic. Repositories handle data access.

### Dependency Injection Flow (cmd/api/main.go)
1. Load config (config.json + env vars)
2. Initialize logger (Zap)
3. Connect database (GORM)
4. Initialize storage (local/cloud)
5. Create repositories → services → handlers
6. Setup auth middleware
7. Wire up router with Chi

**Important**: All dependencies flow downward. Services depend on repositories, handlers depend on services. Never reverse this.

### Key Layers

**internal/domain/**
- `models.go`: GORM models with relationships (Customer, Project, Offer, OfferItem, Contact, File, Activity, Notification, User)
- `dto.go`: API request/response DTOs with validation tags
- DTOs use camelCase JSON, models use PascalCase Go

**internal/repository/**
- Data access layer, one file per entity
- Methods: Create, GetByID, Update, Delete, List (with pagination)
- Always use `tx *gorm.DB` for transactions
- Return domain models, not DTOs

**internal/service/**
- Business logic layer
- Pattern: Service depends on multiple repos + logger
- Responsibilities: validation, calculation, activity logging, denormalized field updates
- Always log activities for audit trail (via ActivityRepository)
- Return DTOs using mapper functions

**internal/http/handler/**
- HTTP request/response handling
- Parse request body, validate, call service, return JSON
- Use `respondWithJSON`, `respondWithError` from `common.go`
- Extract user context: `userCtx, _ := auth.FromContext(r.Context())`

**internal/mapper/**
- Converts domain models ↔ DTOs
- Format timestamps as ISO 8601: `time.Format("2006-01-02T15:04:05Z")`
- Calculate derived fields (margin, budget %)
- Update denormalized fields (CustomerName, ManagerName, etc.)

**internal/auth/**
- `jwt.go`: Validates Azure AD JWT tokens
- `middleware.go`: Authenticate middleware (checks API key OR Bearer token)
- `context.go`: UserContext stored in request context
- System operations use `x-api-key` header, regular users use `Authorization: Bearer <token>`

### Important Patterns

**Denormalized Fields**: Models store redundant data for performance (e.g., `Offer.CustomerName`, `Project.ManagerName`). Always update these in services when related entities change.

**Activity Logging**: All create/update/delete operations must log activities to `Activity` table for audit trail. Use `ActivityRepository.Create()` in service layer.

**CompanyID**: Enum type representing Straye companies (`gruppen`, `stalbygg`, `hybridbygg`, `industri`, `tak`, `montasje`). Filter multi-tenant data by this field.

**Offer Phase Pipeline**: `draft` → `in_progress` → `sent` → `won`/`lost`/`expired`. Use `AdvanceOffer` to progress through pipeline.

**File Storage**: Abstracted in `internal/storage/storage.go`. Supports local filesystem or cloud (Azure Blob, S3). Files linked to Offers via `File.OfferID`.

### Data Warehouse (internal/datawarehouse/)

Read-only connectivity to the MS SQL Server data warehouse for reporting and financial data integration.

**Architecture**:
- Uses `database/sql` with `github.com/microsoft/go-mssqldb` driver (not GORM)
- Separate from the main repository layer since it's read-only
- Optional connection - app starts normally without it if not configured
- Connection pooling with retry logic for transient failures

**Configuration** (via Azure Key Vault or environment):
- `WAREHOUSE-URL` / `WAREHOUSE_URL`: Connection URL (host:port/database)
- `WAREHOUSE-USERNAME` / `WAREHOUSE_USERNAME`: Database user
- `WAREHOUSE-PASSWORD` / `WAREHOUSE_PASSWORD`: Database password
- Enable via `DATAWAREHOUSE_ENABLED=true` or config

**Company Mapping** (Straye ID -> Table Prefix):
- `tak` -> `strayetak`
- `stalbygg` -> `strayestaal`
- `montasje` -> `strayemontage`
- `hybridbygg` -> `strayehybridbygg`
- `industri` -> `strayeindustri`

**General Ledger Tables**: `dbo.nxt_<prefix>_generalledgertransaction`
- Use `OrgUnit12` column to match against project `external_reference`

**Usage**:
```go
// Get client from main.go initialization
results, err := dwClient.ExecuteQuery(ctx, "SELECT * FROM dbo.nxt_strayetak_generalledgertransaction WHERE OrgUnit12 = @ref", externalRef)

// Get table name for a company
tableName, err := datawarehouse.GetGeneralLedgerTableName("tak")
// Returns: "dbo.nxt_strayetak_generalledgertransaction"
```

**Health Check**: `GET /health/datawarehouse` returns status, latency, and pool stats.

## Configuration

Hierarchy: Default values in config.go → config.json → Environment variables (highest priority)

Environment variables use underscore: `DATABASE_HOST`, `AZURE_TENANT_ID`, `ADMIN_API_KEY`

## Database Migrations

- Use Goose: `migrations/*.sql` files
- Naming: `00001_description.sql`
- Migrations run automatically in Docker, manually with `make migrate-up`
- Always write both `-- +goose Up` and `-- +goose Down` sections

## API Documentation

OpenAPI/Swagger: Annotations in handlers → `make swagger` → `/swagger/index.html`

Security schemes: `BearerAuth` (JWT) or `ApiKeyAuth` (x-api-key header)

## Testing Strategy

- **Fast tests** (tests/auth, tests/mapper): No DB, pure logic
- **Integration tests** (tests/repository, tests/service): Require PostgreSQL
- Use table-driven tests with subtests
- Mock external dependencies (Azure AD in JWT tests)

## Common Gotchas

1. **UUID types**: Use `uuid.UUID` from `github.com/google/uuid`, not strings
2. **Nullable fields**: Use pointers (`*uuid.UUID`, `*time.Time`) for optional foreign keys/dates
3. **String arrays**: Use `pq.StringArray` for PostgreSQL text[] (e.g., `Project.TeamMembers`)
4. **Error wrapping**: Use `fmt.Errorf("context: %w", err)` for stack traces
5. **Request validation**: Validate in handler, return 400 errors early
6. **Pagination**: Use `page`, `pageSize` query params; max 200 items/page
7. **CORS**: Currently allows all origins (`AllowOriginFunc` always true) - configure for production
8. **Graceful shutdown**: Server handles SIGTERM with 30s timeout

## Development Workflow

1. Read relevant code first (handler → service → repository)
2. For new features: Create migration → Update models → Add repository methods → Add service logic → Add handler → Update router
3. For bugs: Start with handler, trace through service to repository
4. Always run `make test` before committing
5. Update Swagger annotations when changing API contracts
6. Log structured errors with Zap, include context fields
