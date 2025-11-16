# Repository Guidelines

## Project Structure & Module Organization
- `cmd/api` hosts the HTTP application entrypoint; `cmd/migrate` runs Goose migrations.
- Business logic lives in `internal/` (auth, config, http handlers, service, repository, etc.).
- Database scripts reside in `migrations/`, while integration and unit tests sit under `tests/`.
- Local assets (uploaded files) default to `storage/`; keep large binaries out of Git.

## Build, Test, and Development Commands
- `make run` – builds in-memory and starts `cmd/api` once.
- `make dev` – starts Air hot reload using `.air.toml`; requires `go install github.com/air-verse/air@latest`.
- `make build` – produces binaries in `bin/` for the API and migration tool.
- `make test` / `make test-all` – run focused tests or the entire suite, respectively.
- `make docker-up` / `make docker-down` – spin up or tear down the Docker Compose stack (API + Postgres).
- `make migrate-up|down|status` – execute Goose migrations via `cmd/migrate`.

## Coding Style & Naming Conventions
- Go 1.21+ code formatted with `go fmt` and `goimports`; run `make format` before committing.
- Use idiomatic Go naming: exported identifiers in PascalCase, unexported in camelCase, constants in ALL_CAPS only when appropriate.
- Keep handler/service/repository files grouped by domain (e.g., `internal/http/handler/customer_handler.go`).
- Logging uses Uber’s Zap; prefer structured fields (`zap.String`) over string concatenation.

## Testing Guidelines
- Unit tests live in `tests/` packages and use Go’s standard `testing` framework.
- Use descriptive test names following `Test<Component>_<Behavior>`.
- Prefer table-driven tests for services and repositories.
- Run `make test` before submitting small changes; `make test-all` (or `make test-integration`) is expected before merges affecting data access or external dependencies.

## Commit & Pull Request Guidelines
- Follow conventional, descriptive commit messages (e.g., “Fix customer search query” or “Add Air dev workflow”).
- Each PR should include: summary of changes, testing evidence (`make test` output), and references to related issues or tickets.
- Screenshots/GIFs are encouraged for UI-facing or API contract changes (e.g., Swagger updates).
- Keep PRs focused—refactor or cosmetic updates should be isolated from feature work unless tightly coupled.

## Security & Configuration Tips
- Secrets (API keys, Azure AD values) are loaded via environment variables; never commit `.env` or secret JSON files.
- Local storage (`storage/`) may contain uploads; add new subpaths to `.gitignore` if necessary.
- Use `config.json` only for non-sensitive defaults; production settings should come from environment variables or secret stores.
