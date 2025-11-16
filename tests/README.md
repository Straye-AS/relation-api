# Tests

This directory contains all tests for the Straye Relation API, organized by component type.

## Directory Structure

```
tests/
├── auth/              # Authentication and JWT tests
├── mapper/            # DTO mapper tests  
├── repository/        # Database repository tests (requires PostgreSQL)
├── service/           # Business logic service tests (requires PostgreSQL)
└── integration/       # End-to-end integration tests
```

## Running Tests

### Run All Tests (Unit Tests Only)
```bash
make test
```

### Run Specific Test Suites
```bash
# Auth and mapper tests (no database required)
go test ./tests/auth ./tests/mapper -v

# Repository and service tests (require PostgreSQL)
go test ./tests/repository ./tests/service -v
```

### Run Tests with Coverage
```bash
make test-coverage
```

## Test Requirements

### Unit Tests (auth, mapper)
- **No external dependencies**
- Run with: `go test ./tests/auth ./tests/mapper -v`

### Integration Tests (repository, service)
- **Require PostgreSQL database**
- Use Docker to run database: `make docker-up`
- Repository tests use in-memory SQLite (limited PostgreSQL feature support)
- Service tests require full PostgreSQL instance

## Writing Tests

### Package Naming Convention
- Tests use `_test` suffix on package names (e.g., `package auth_test`)
- This allows testing public APIs from an external perspective
- Use internal imports for testing: `github.com/straye-as/relation-api/internal/...`

### Test File Locations
- All test files are in `tests/` directory, separated from source code
- Test files follow the naming pattern: `*_test.go`
- Helper functions should be exported (capitalized) if used in tests

## CI/CD

Tests are automatically run in CI/CD pipelines:
- Unit tests: Run on every commit
- Integration tests: Run on pull requests to main
- Coverage reports: Generated and uploaded to code coverage service

## Notes

- Repository and service tests may fail with SQLite due to PostgreSQL-specific features (UUID type, pg_array, etc.)
- For full test suite, use PostgreSQL test database
- Mock external dependencies where possible
- Keep tests fast and isolated

