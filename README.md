# Straye Relation API

A production-grade REST API service for Customer Relationship Management (CRM), providing comprehensive customer, project, and sales pipeline management capabilities.

## Features

- **Customer Management**: Complete CRUD operations for customers with contacts
- **Project Tracking**: Project management with budget tracking and status monitoring
- **Offer Management**: Sales proposals with line items and phase tracking
- **Activity Logging**: Complete audit trail for all entities
- **File Management**: Upload and download files with offer attachments
- **Dashboard & Metrics**: Real-time business metrics and global search
- **Dual Authentication**: JWT Bearer tokens + API Key authentication
- **Production Ready**: Structured logging, error handling, Docker support

## Architecture

- **Clean Architecture**: Separation of concerns with layers (handlers → services → repositories)
- **Database**: PostgreSQL with GORM ORM
- **HTTP Router**: Chi router with middleware support
- **Authentication**: JWT validation from Azure AD/OAuth + API Key
- **File Storage**: Abstracted storage layer (local filesystem or cloud)
- **Logging**: Structured JSON logging with Zap
- **Migrations**: Database migrations with Goose
- **API Documentation**: OpenAPI 3.0 / Swagger

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

### Running with Docker Compose

```bash
# Start all services (PostgreSQL + API)
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

The API will be available at `http://localhost:8080`

### Running Locally

1. **Install dependencies**:
```bash
go mod download
```

2. **Configure database**:
   - Edit `config.json` with your PostgreSQL connection details
   - Or set environment variables (see Configuration section)

3. **Run migrations**:
```bash
go run ./cmd/migrate up
```

4. **Start the API**:
```bash
make run       # single run
# or
make dev       # hot reload via Air (see below)
```

### Hot Reload (Air)

For nodemon-style auto rebuilds, use [Air](https://github.com/cosmtrek/air):

```bash
go install github.com/air-verse/air@latest  # one-time install (puts `air` in $(go env GOPATH)/bin)
make dev                                    # uses the included .air.toml config
```

Use `make run` for a single run, `make dev` for auto-reloads, or `make docker-up` to start via Docker Compose. Run `make help` to see the full list of shortcuts. If you see “Air is not installed,” install it via `go install github.com/air-verse/air@latest` and ensure `$(go env GOPATH)/bin` is on your PATH.

## Configuration

Configuration can be provided via:
1. `config.json` file
2. Environment variables (highest priority)

### Environment Variables

```bash
# Application
APP_NAME="Straye Relation API"
APP_ENVIRONMENT=development
APP_PORT=8080

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=relation_db
DATABASE_USER=relation_user
DATABASE_PASSWORD=relation_password

# Authentication
AZURE_TENANT_ID=your-tenant-id
AZURE_CLIENT_ID=your-client-id
AZURE_CLIENT_SECRET=your-client-secret
AZURE_REQUIRED_SCOPES=api://relation-api/access
ADMIN_API_KEY=your-secret-api-key

# Storage
STORAGE_MODE=local
STORAGE_LOCALBASEPATH=./storage
STORAGE_MAXUPLOADSIZEMB=50

# Logging
LOGGING_LEVEL=info
LOGGING_FORMAT=json
```

## API Endpoints

### Authentication
- `GET /auth/me` - Get current authenticated user
- `GET /users` - List users

### Customers
- `GET /customers` - List customers (paginated)
- `POST /customers` - Create customer
- `GET /customers/{id}` - Get customer
- `PUT /customers/{id}` - Update customer
- `DELETE /customers/{id}` - Delete customer
- `GET /customers/{id}/contacts` - List customer contacts
- `POST /customers/{id}/contacts` - Create contact

### Projects
- `GET /projects` - List projects (paginated, filterable)
- `POST /projects` - Create project
- `GET /projects/{id}` - Get project
- `PUT /projects/{id}` - Update project
- `GET /projects/{id}/budget` - Get budget summary
- `GET /projects/{id}/activities` - Get activity log

### Offers
- `GET /offers` - List offers (paginated, filterable)
- `POST /offers` - Create offer
- `GET /offers/{id}` - Get offer with items
- `PUT /offers/{id}` - Update offer
- `POST /offers/{id}/advance` - Advance offer phase
- `GET /offers/{id}/items` - Get offer items
- `POST /offers/{id}/items` - Add offer item
- `GET /offers/{id}/files` - Get offer files
- `GET /offers/{id}/activities` - Get activity log

### Files
- `POST /files/upload` - Upload file
- `GET /files/{id}` - Get file metadata
- `GET /files/{id}/download` - Download file

### Dashboard
- `GET /dashboard/metrics` - Get aggregate metrics
- `GET /search?q=query` - Global search

## Authentication

### JWT Bearer Token

Add to request headers:
```
Authorization: Bearer <your-jwt-token>
```

### API Key

For system operations, add to request headers:
```
x-api-key: <your-api-key>
```

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Generate Swagger Docs

```bash
swag init -g cmd/api/main.go -o ./docs
```

Access Swagger UI at: `http://localhost:8080/swagger/index.html`

### Database Migrations

```bash
# Run migrations up
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status

# Create new migration
make migrate-create name=add_new_field
```

### Code Quality

```bash
# Format code
make format

# Run linter
make lint

# Security scan
make security
```

## Project Structure

```
.
├── cmd/
│   ├── api/          # Main application entry point
│   └── migrate/      # Migration runner
├── internal/
│   ├── auth/         # Authentication & authorization
│   ├── config/       # Configuration management
│   ├── database/     # Database connection
│   ├── domain/       # Domain models & DTOs
│   ├── http/         # HTTP handlers & middleware
│   ├── logger/       # Structured logging
│   ├── mapper/       # DTO mappers
│   ├── repository/   # Data access layer
│   ├── service/      # Business logic layer
│   └── storage/      # File storage abstraction
├── migrations/       # Database migrations
├── config.json       # Configuration file
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## CI/CD Pipeline

The project includes a GitHub Actions CI/CD pipeline (`.github/workflows/ci.yml`) that runs automatically on:
- **Pull Requests to main**: Runs linting, security scanning, and tests
- **Pushes to main**: Runs all checks plus builds and pushes Docker image to ACR

### Pipeline Jobs

| Job | Description | Trigger |
|-----|-------------|---------|
| `lint` | Runs golangci-lint for static code analysis | PR + Push |
| `security` | Runs gosec security vulnerability scan | PR + Push |
| `test-unit` | Runs fast unit tests (no DB required) | PR + Push |
| `test-integration` | Runs integration tests with PostgreSQL | PR + Push |
| `build` | Builds application binaries | PR + Push |
| `docker` | Builds and pushes Docker image to ACR | Push to main only |
| `notify-failure` | Sends Slack notification on failure | Push to main only |

### Required GitHub Secrets

Configure these secrets in your GitHub repository settings (`Settings > Secrets and variables > Actions`):

#### Azure Authentication (OIDC - Recommended)
| Secret | Description | Example |
|--------|-------------|---------|
| `AZURE_CLIENT_ID` | Azure service principal client ID | `12345678-1234-1234-1234-123456789012` |
| `AZURE_TENANT_ID` | Azure AD tenant ID | `12345678-1234-1234-1234-123456789012` |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription ID | `12345678-1234-1234-1234-123456789012` |
| `ACR_LOGIN_SERVER` | ACR login server hostname | `myregistry.azurecr.io` |

#### Alternative: ACR Username/Password
If not using OIDC, configure:
| Secret | Description |
|--------|-------------|
| `ACR_USERNAME` | ACR username |
| `ACR_PASSWORD` | ACR password |

#### Optional
| Secret | Description |
|--------|-------------|
| `SLACK_WEBHOOK_URL` | Slack webhook URL for failure notifications |

### Azure Setup for OIDC Authentication

OIDC (Workload Identity Federation) is the recommended approach - **no secrets or passwords are stored**, only IDs. GitHub and Azure establish trust through federated credentials.

#### Step 1: Create App Registration (no secret needed)
```bash
# Create the app registration
az ad app create --display-name "github-actions-relation-api"

# Get the App (Client) ID - save this for AZURE_CLIENT_ID secret
az ad app list --display-name "github-actions-relation-api" --query "[0].appId" -o tsv
```

#### Step 2: Create Service Principal
```bash
# Create service principal for the app
APP_ID=$(az ad app list --display-name "github-actions-relation-api" --query "[0].appId" -o tsv)
az ad sp create --id $APP_ID

# Get the Service Principal Object ID (needed for role assignment)
az ad sp show --id $APP_ID --query "id" -o tsv
```

#### Step 3: Grant ACR Push Permissions
```bash
# Assign AcrPush role to the service principal
SP_OBJECT_ID=$(az ad sp show --id $APP_ID --query "id" -o tsv)

az role assignment create \
  --assignee-object-id $SP_OBJECT_ID \
  --assignee-principal-type ServicePrincipal \
  --role "AcrPush" \
  --scope /subscriptions/{subscription-id}/resourceGroups/{resource-group}/providers/Microsoft.ContainerRegistry/registries/{acr-name}
```

#### Step 4: Configure Federated Credentials for GitHub
```bash
# This establishes trust between GitHub Actions and Azure AD
az ad app federated-credential create --id $APP_ID --parameters '{
  "name": "github-main-branch",
  "issuer": "https://token.actions.githubusercontent.com",
  "subject": "repo:Straye-AS/relation-api:ref:refs/heads/main",
  "audiences": ["api://AzureADTokenExchange"]
}'

# Optional: Add credential for pull requests too
az ad app federated-credential create --id $APP_ID --parameters '{
  "name": "github-pull-requests",
  "issuer": "https://token.actions.githubusercontent.com",
  "subject": "repo:Straye-AS/relation-api:pull_request",
  "audiences": ["api://AzureADTokenExchange"]
}'
```

#### Step 5: Get Required Values for GitHub Secrets
```bash
# AZURE_CLIENT_ID (App ID)
az ad app list --display-name "github-actions-relation-api" --query "[0].appId" -o tsv

# AZURE_TENANT_ID
az account show --query "tenantId" -o tsv

# AZURE_SUBSCRIPTION_ID
az account show --query "id" -o tsv

# ACR_LOGIN_SERVER (your ACR hostname)
az acr show --name {acr-name} --query "loginServer" -o tsv
```

### Docker Image Tags

Images pushed to ACR are tagged with:
- Full Git commit SHA (e.g., `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0`) - for easy comparison and rollback
- `latest` - always points to the most recent main branch build

Example:
```
myregistry.azurecr.io/relation-api:abc123def456...  # specific version
myregistry.azurecr.io/relation-api:latest           # current version
```

## Deployment

### Docker

```bash
# Build image
docker build -t relation-api:latest .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_HOST=your-db-host \
  -e DATABASE_PASSWORD=your-db-password \
  -e ADMIN_API_KEY=your-api-key \
  relation-api:latest
```

### Kubernetes

Deployment manifests can be generated using the provided Dockerfile and environment variables.

### Environment Considerations

- **Development**: Use local storage, JSON logging to console
- **Production**: Use cloud storage, structured JSON logs, proper secrets management
- Ensure `ADMIN_API_KEY` is stored securely (Azure Key Vault, AWS Secrets Manager, etc.)
- Configure proper CORS origins for your frontend
- Enable HTTPS/TLS at load balancer level

## Performance

- Connection pooling: 25 max open connections, 5 max idle
- Request timeout: 60 seconds
- File upload limit: 50MB (configurable)
- Pagination: Max 200 items per page
- Database indexes on foreign keys and search fields

## Security

- JWT token validation with public key verification
- API key constant-time comparison
- SQL injection protection via prepared statements
- File upload size limits
- Request body size limits
- CORS configuration
- Structured audit logging

## Monitoring

Structured JSON logs include:
- Request ID
- User context (ID, name)
- HTTP method, path, status
- Response time
- Error details with stack traces

Integrate with:
- ELK Stack (Elasticsearch, Logstash, Kibana)
- Azure Application Insights
- Datadog
- CloudWatch

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: [repository-url]/issues
- Email: support@straye.io
