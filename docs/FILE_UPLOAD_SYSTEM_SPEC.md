# File Upload System - Implementation Specification

**Epic**: [sc-304] File Upload System
**Created**: 2025-12-20
**Status**: Planned

## Overview

Add modular file upload/download functionality with Azure Blob Storage support. Files can be attached to Customer, Project, Offer, and Supplier entities using a polymorphic design pattern.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Storage | Azure Blob Storage | REST API, scalable, cost-effective for web apps |
| Environment Separation | Container per env | `relation-files-dev`, `relation-files-staging`, `relation-files-prod` |
| Access Control | Private (authenticated) | All downloads through API with Bearer token |
| Deletion | Hard delete | Remove from both DB and blob storage |
| Max File Size | 50MB (configurable) | Via `STORAGE_MAXUPLOADSIZEMB` |

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/customers/{id}/files` | Upload file to customer |
| GET | `/api/v1/customers/{id}/files` | List customer files |
| POST | `/api/v1/projects/{id}/files` | Upload file to project |
| GET | `/api/v1/projects/{id}/files` | List project files |
| POST | `/api/v1/offers/{id}/files` | Upload file to offer |
| GET | `/api/v1/offers/{id}/files` | List offer files |
| GET | `/api/v1/files/{id}` | Get file metadata |
| GET | `/api/v1/files/{id}/download` | Download file |
| DELETE | `/api/v1/files/{id}` | Delete file |

---

## Shortcut Stories

| ID | Type | Story | Blocked By |
|----|------|-------|------------|
| sc-305 | Chore | Add Azure Blob Storage SDK dependency | - |
| sc-306 | Feature | Database migration for polymorphic files | - |
| sc-307 | Feature | Update File domain model for polymorphic design | sc-306 |
| sc-308 | Feature | Implement Azure Blob Storage adapter | sc-305 |
| sc-309 | Feature | Update storage factory to support Azure Blob Storage | sc-308 |
| sc-310 | Feature | Update File repository for polymorphic queries | sc-307 |
| sc-311 | Feature | Update File service for entity-aware operations | sc-309, sc-310, sc-312 |
| sc-312 | Feature | Update File mapper for polymorphic DTOs | sc-307 |
| sc-313 | Feature | Update File handler with entity-specific endpoints | sc-311 |
| sc-314 | Feature | Add file routes for entity-specific endpoints | sc-313 |
| sc-315 | Feature | Update dependency injection for File service | sc-309, sc-314 |
| sc-316 | Chore | Add unit tests for file upload system | sc-307, sc-312 |
| sc-317 | Chore | Add integration tests for file upload system | sc-315 |

---

## Implementation Phases

### Phase 1: Foundation (Parallelizable)
- **sc-305**: Add Azure SDK dependency to go.mod
- **sc-306**: Create database migration for polymorphic files

### Phase 2: Core Components
- **sc-307**: Update domain model (File struct, FileEntityType enum)
- **sc-308**: Implement Azure Blob Storage adapter

### Phase 3: Infrastructure
- **sc-309**: Update storage factory
- **sc-310**: Update repository for polymorphic queries
- **sc-312**: Update mapper for new DTO fields

### Phase 4: Business Logic
- **sc-311**: Update service layer for entity-aware operations

### Phase 5: API Layer
- **sc-313**: Update handler with entity-specific endpoints
- **sc-314**: Add routes for entity file endpoints
- **sc-315**: Update dependency injection in main.go

### Phase 6: Testing
- **sc-316**: Unit tests (can start after Phase 3)
- **sc-317**: Integration tests (after Phase 5)

---

## Technical Implementation Details

### Database Migration

**File**: `migrations/00056_polymorphic_files.sql`

```sql
-- +goose Up
ALTER TABLE files ADD COLUMN entity_type VARCHAR(50);
ALTER TABLE files ADD COLUMN entity_id UUID;

-- Migrate existing offer files
UPDATE files SET entity_type = 'offer', entity_id = offer_id WHERE offer_id IS NOT NULL;
DELETE FROM files WHERE offer_id IS NULL;

CREATE INDEX idx_files_entity ON files(entity_type, entity_id);

ALTER TABLE files DROP CONSTRAINT IF EXISTS files_offer_id_fkey;
DROP INDEX IF EXISTS idx_files_offer_id;
ALTER TABLE files DROP COLUMN offer_id;

ALTER TABLE files ALTER COLUMN entity_type SET NOT NULL;
ALTER TABLE files ALTER COLUMN entity_id SET NOT NULL;

-- +goose Down
-- (reverse migration included in actual file)
```

### Domain Model Updates

**File**: `internal/domain/models.go`

```go
// FileEntityType represents the type of entity a file is attached to
type FileEntityType string

const (
    FileEntityCustomer FileEntityType = "customer"
    FileEntityProject  FileEntityType = "project"
    FileEntityOffer    FileEntityType = "offer"
    FileEntitySupplier FileEntityType = "supplier"
)

func (t FileEntityType) IsValid() bool {
    switch t {
    case FileEntityCustomer, FileEntityProject, FileEntityOffer, FileEntitySupplier:
        return true
    }
    return false
}

// File represents an uploaded file attached to an entity
type File struct {
    BaseModel
    Filename    string         `gorm:"type:varchar(255);not null"`
    ContentType string         `gorm:"type:varchar(100);not null"`
    Size        int64          `gorm:"not null"`
    StoragePath string         `gorm:"type:varchar(500);not null;unique"`
    EntityType  FileEntityType `gorm:"type:varchar(50);not null;index:idx_files_entity"`
    EntityID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_files_entity"`
}
```

### DTO Updates

**File**: `internal/domain/dto.go`

```go
type FileDTO struct {
    ID          uuid.UUID `json:"id"`
    Filename    string    `json:"filename"`
    ContentType string    `json:"contentType"`
    Size        int64     `json:"size"`
    EntityType  string    `json:"entityType"`
    EntityID    uuid.UUID `json:"entityId"`
    CreatedAt   string    `json:"createdAt"`
}
```

### Azure Blob Storage Implementation

**File**: `internal/storage/azure_blob.go` (NEW)

```go
package storage

import (
    "context"
    "fmt"
    "io"
    "path/filepath"
    "strings"

    "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

type AzureBlobStorage struct {
    client        *azblob.Client
    containerName string
    logger        *zap.Logger
}

func NewAzureBlobStorage(connectionString, containerName string, logger *zap.Logger) (*AzureBlobStorage, error) {
    client, err := azblob.NewClientFromConnectionString(connectionString, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create blob client: %w", err)
    }

    // Ensure container exists
    _, err = client.CreateContainer(context.Background(), containerName, nil)
    if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
        return nil, fmt.Errorf("failed to create container: %w", err)
    }

    return &AzureBlobStorage{
        client:        client,
        containerName: containerName,
        logger:        logger,
    }, nil
}

func (s *AzureBlobStorage) Upload(ctx context.Context, filename string, contentType string, data io.Reader) (string, int64, error) {
    ext := filepath.Ext(filename)
    blobName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

    resp, err := s.client.UploadStream(ctx, s.containerName, blobName, data, &azblob.UploadStreamOptions{
        HTTPHeaders: &azblob.HTTPHeaders{
            BlobContentType: &contentType,
        },
    })
    if err != nil {
        return "", 0, fmt.Errorf("failed to upload blob: %w", err)
    }

    props, err := s.client.ServiceClient().NewContainerClient(s.containerName).NewBlobClient(blobName).GetProperties(ctx, nil)
    if err != nil {
        return "", 0, fmt.Errorf("failed to get blob properties: %w", err)
    }

    s.logger.Info("File uploaded to Azure Blob",
        zap.String("blob", blobName),
        zap.String("etag", string(*resp.ETag)),
    )

    return blobName, *props.ContentLength, nil
}

func (s *AzureBlobStorage) Download(ctx context.Context, storagePath string) (io.ReadCloser, error) {
    resp, err := s.client.DownloadStream(ctx, s.containerName, storagePath, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to download blob: %w", err)
    }
    return resp.Body, nil
}

func (s *AzureBlobStorage) Delete(ctx context.Context, storagePath string) error {
    _, err := s.client.DeleteBlob(ctx, s.containerName, storagePath, nil)
    if err != nil {
        return fmt.Errorf("failed to delete blob: %w", err)
    }
    return nil
}
```

### Storage Factory Update

**File**: `internal/storage/storage.go`

```go
func NewStorage(cfg *config.StorageConfig, logger *zap.Logger) (Storage, error) {
    switch cfg.Mode {
    case "local":
        return NewLocalStorage(cfg.LocalBasePath)
    case "cloud", "azure":
        if cfg.CloudConnectionString == "" {
            return nil, fmt.Errorf("cloud connection string required for azure storage")
        }
        return NewAzureBlobStorage(cfg.CloudConnectionString, cfg.CloudContainer, logger)
    default:
        return nil, fmt.Errorf("unsupported storage mode: %s", cfg.Mode)
    }
}
```

### Repository Updates

**File**: `internal/repository/file_repository.go`

```go
func (r *FileRepository) ListByEntity(ctx context.Context, entityType domain.FileEntityType, entityID uuid.UUID) ([]domain.File, error) {
    var files []domain.File
    err := r.db.WithContext(ctx).
        Where("entity_type = ? AND entity_id = ?", entityType, entityID).
        Order("created_at DESC").
        Find(&files).Error
    return files, err
}

func (r *FileRepository) CountByEntity(ctx context.Context, entityType domain.FileEntityType, entityID uuid.UUID) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&domain.File{}).
        Where("entity_type = ? AND entity_id = ?", entityType, entityID).
        Count(&count).Error
    return count, err
}
```

### Service Layer Updates

**File**: `internal/service/file_service.go`

Key changes:
- Add `customerRepo`, `projectRepo` dependencies
- Implement `verifyEntityExists()` method to validate entity before upload
- Update `Upload()` to accept `entityType` and `entityID` parameters
- Add `ListByEntity()` method
- Log activities with entity context

### Handler Updates

**File**: `internal/http/handler/file_handler.go`

New methods:
- `UploadToEntity()` - Handle POST to entity-specific file endpoint
- `ListEntityFiles()` - Handle GET for listing entity files

### Router Updates

**File**: `internal/http/router/router.go`

```go
// Customer files
r.Route("/customers/{customerId}/files", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Post("/", fileHandler.UploadToEntity)
    r.Get("/", fileHandler.ListEntityFiles)
})

// Project files
r.Route("/projects/{projectId}/files", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Post("/", fileHandler.UploadToEntity)
    r.Get("/", fileHandler.ListEntityFiles)
})

// Offer files
r.Route("/offers/{offerId}/files", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Post("/", fileHandler.UploadToEntity)
    r.Get("/", fileHandler.ListEntityFiles)
})

// Generic file operations
r.Route("/files/{fileId}", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Get("/", fileHandler.GetByID)
    r.Get("/download", fileHandler.Download)
    r.Delete("/", fileHandler.Delete)
})
```

---

## Configuration

### Environment Variables

```bash
STORAGE_MODE=cloud                           # or "azure" or "local"
STORAGE_CLOUDCONTAINER=relation-files-dev    # per environment
STORAGE_MAXUPLOADSIZEMB=50                   # configurable limit

# Connection string from Azure Key Vault secret: "storage-connection-string"
```

### Container Naming Convention

| Environment | Container Name |
|-------------|---------------|
| Development | `relation-files-dev` |
| Staging | `relation-files-staging` |
| Production | `relation-files-prod` |

---

## Dependencies

Add to `go.mod`:
```
github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.0
```

---

## Files to Modify

| File | Change Type |
|------|-------------|
| `migrations/00056_polymorphic_files.sql` | NEW |
| `internal/domain/models.go` | MODIFY |
| `internal/domain/dto.go` | MODIFY |
| `internal/storage/azure_blob.go` | NEW |
| `internal/storage/storage.go` | MODIFY |
| `internal/repository/file_repository.go` | MODIFY |
| `internal/service/file_service.go` | MODIFY |
| `internal/http/handler/file_handler.go` | MODIFY |
| `internal/http/router/router.go` | MODIFY |
| `internal/mapper/mapper.go` | MODIFY |
| `cmd/api/main.go` | MODIFY |
| `go.mod` | MODIFY |

---

## Testing Strategy

### Unit Tests
- FileEntityType validation
- Mapper tests for new FileDTO fields
- Azure Blob Storage mock tests

### Integration Tests
- Upload to each entity type (customer, project, offer)
- List files by entity
- Delete file (verify storage + DB cleanup)
- Error handling (entity not found, file too large)

### Local Testing
Use `STORAGE_MODE=local` for development without Azure dependency.

### Staging Testing
Test with real Azure Blob Storage using staging container.

---

## Future Extensions

### Adding Supplier Entity Support
When the Supplier entity is implemented:
1. `FileEntitySupplier` constant is already defined in the enum
2. Add `supplierRepo` dependency to FileService
3. Implement `verifyEntityExists` case for supplier
4. Add supplier file routes to router

### Potential Enhancements
- File type validation (whitelist/blacklist)
- Virus scanning integration
- Thumbnail generation for images
- Signed URL support for temporary public access
- Batch upload support
- File versioning
