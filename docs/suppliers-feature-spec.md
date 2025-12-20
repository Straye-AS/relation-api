# Suppliers Feature Implementation Plan

## Overview

Add a comprehensive "Suppliers" feature to the Straye Relation API, enabling management of organizations that provide goods/services (steel, windows, concrete, etc.) with many-to-many relationships to offers, document attachments, and global search integration.

## Key Requirements

- Suppliers as organizations with org numbers, locations, names (like customers)
- Multiple suppliers per offer (many-to-many relationship)
- Free-text category field (optional) - e.g., "steel", "windows", "concrete"
- View all suppliers with connected offers/projects
- Multiple contacts per supplier (like customers have contacts)
- Simple active/done status on supplier-offer relationships
- Document uploads for: Suppliers, Customers, Offers
- Global search integration

## Shortcut Stories

| Story | Title | Type | Depends On |
|-------|-------|------|------------|
| sc-294 | Create database migrations for Suppliers feature | Chore | - |
| sc-295 | Implement Supplier CRUD API | Feature | sc-294 |
| sc-296 | Implement Supplier Contacts API | Feature | sc-295 |
| sc-297 | Implement Offer-Supplier relationship API | Feature | sc-295 |
| sc-299 | Extend file uploads for Suppliers and Customers | Feature | sc-295 |
| sc-300 | Add Suppliers to global search | Feature | sc-295 |
| sc-301 | Add Supplier stats and dashboard integration | Feature | sc-297 |
| sc-302 | Add Supplier repository and service tests | Chore | sc-295 |
| sc-303 | Add Supplier handler integration tests | Chore | sc-302 |

---

## Database Migrations

### Migration 00056: Create Suppliers Table

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TYPE supplier_status AS ENUM ('active', 'inactive', 'pending', 'blacklisted');

CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    org_number VARCHAR(20) UNIQUE,
    email VARCHAR(255),
    phone VARCHAR(50),
    address VARCHAR(500),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Norway',
    municipality VARCHAR(100),
    county VARCHAR(100),
    contact_person VARCHAR(200),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    status supplier_status NOT NULL DEFAULT 'active',
    category VARCHAR(200),
    notes TEXT,
    payment_terms VARCHAR(200),
    website VARCHAR(500),
    company_id VARCHAR(50),
    created_by_id VARCHAR(100),
    created_by_name VARCHAR(200),
    updated_by_id VARCHAR(100),
    updated_by_name VARCHAR(200),
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_suppliers_name ON suppliers(name);
CREATE INDEX idx_suppliers_org_number ON suppliers(org_number);
CREATE INDEX idx_suppliers_company_id ON suppliers(company_id);
CREATE INDEX idx_suppliers_status ON suppliers(status);
CREATE INDEX idx_suppliers_category ON suppliers(category);
CREATE INDEX idx_suppliers_city ON suppliers(city);
CREATE INDEX idx_suppliers_deleted_at ON suppliers(deleted_at);
CREATE INDEX idx_suppliers_created_by_id ON suppliers(created_by_id);

CREATE TRIGGER update_suppliers_updated_at
    BEFORE UPDATE ON suppliers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_suppliers_updated_at ON suppliers;
DROP TABLE IF EXISTS suppliers;
DROP TYPE IF EXISTS supplier_status;
-- +goose StatementEnd
```

### Migration 00057: Create Supplier Contacts Table

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE supplier_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    title VARCHAR(200),
    email VARCHAR(255),
    phone VARCHAR(50),
    is_primary BOOLEAN NOT NULL DEFAULT false,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_supplier_contacts_supplier_id ON supplier_contacts(supplier_id);

CREATE TRIGGER update_supplier_contacts_updated_at
    BEFORE UPDATE ON supplier_contacts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_supplier_contacts_updated_at ON supplier_contacts;
DROP TABLE IF EXISTS supplier_contacts;
-- +goose StatementEnd
```

### Migration 00058: Create Offer-Supplier Junction Table

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TYPE offer_supplier_status AS ENUM ('active', 'done');

CREATE TABLE offer_suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id UUID NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
    supplier_name VARCHAR(200),
    offer_title VARCHAR(200),
    status offer_supplier_status NOT NULL DEFAULT 'active',
    notes TEXT,
    created_by_id VARCHAR(100),
    created_by_name VARCHAR(200),
    updated_by_id VARCHAR(100),
    updated_by_name VARCHAR(200),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_offer_supplier UNIQUE (offer_id, supplier_id)
);

CREATE INDEX idx_offer_suppliers_offer_id ON offer_suppliers(offer_id);
CREATE INDEX idx_offer_suppliers_supplier_id ON offer_suppliers(supplier_id);
CREATE INDEX idx_offer_suppliers_status ON offer_suppliers(status);

CREATE TRIGGER update_offer_suppliers_updated_at
    BEFORE UPDATE ON offer_suppliers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_offer_suppliers_updated_at ON offer_suppliers;
DROP TABLE IF EXISTS offer_suppliers;
DROP TYPE IF EXISTS offer_supplier_status;
-- +goose StatementEnd
```

### Migration 00059: Extend Files for Polymorphic Attachments

```sql
-- +goose Up
-- +goose StatementBegin

ALTER TABLE files ADD COLUMN supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;
ALTER TABLE files ADD COLUMN customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

CREATE INDEX idx_files_supplier_id ON files(supplier_id);
CREATE INDEX idx_files_customer_id ON files(customer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_files_customer_id;
DROP INDEX IF EXISTS idx_files_supplier_id;
ALTER TABLE files DROP COLUMN IF EXISTS customer_id;
ALTER TABLE files DROP COLUMN IF EXISTS supplier_id;
-- +goose StatementEnd
```

---

## Domain Models

### New Models (internal/domain/models.go)

```go
// SupplierStatus represents the status of a supplier
type SupplierStatus string

const (
    SupplierStatusActive      SupplierStatus = "active"
    SupplierStatusInactive    SupplierStatus = "inactive"
    SupplierStatusPending     SupplierStatus = "pending"
    SupplierStatusBlacklisted SupplierStatus = "blacklisted"
)

// OfferSupplierStatus represents the status of a supplier in an offer
type OfferSupplierStatus string

const (
    OfferSupplierStatusActive OfferSupplierStatus = "active"
    OfferSupplierStatusDone   OfferSupplierStatus = "done"
)

// Supplier represents an organization that provides goods/services
type Supplier struct {
    BaseModel
    DeletedAt     gorm.DeletedAt   `gorm:"index"`
    Name          string           `gorm:"type:varchar(200);not null;index"`
    OrgNumber     string           `gorm:"type:varchar(20);unique;index"`
    Email         string           `gorm:"type:varchar(255)"`
    Phone         string           `gorm:"type:varchar(50)"`
    Address       string           `gorm:"type:varchar(500)"`
    City          string           `gorm:"type:varchar(100)"`
    PostalCode    string           `gorm:"type:varchar(20)"`
    Country       string           `gorm:"type:varchar(100);not null;default:'Norway'"`
    Municipality  string           `gorm:"type:varchar(100)"`
    County        string           `gorm:"type:varchar(100)"`
    ContactPerson string           `gorm:"type:varchar(200)"`
    ContactEmail  string           `gorm:"type:varchar(255)"`
    ContactPhone  string           `gorm:"type:varchar(50)"`
    Status        SupplierStatus   `gorm:"type:varchar(50);not null;default:'active';index"`
    Category      string           `gorm:"type:varchar(200)"`
    Notes         string           `gorm:"type:text"`
    PaymentTerms  string           `gorm:"type:varchar(200)"`
    Website       string           `gorm:"type:varchar(500)"`
    CompanyID     *CompanyID       `gorm:"type:varchar(50);column:company_id;index"`
    Company       *Company         `gorm:"foreignKey:CompanyID"`
    CreatedByID   string           `gorm:"type:varchar(100);column:created_by_id;index"`
    CreatedByName string           `gorm:"type:varchar(200);column:created_by_name"`
    UpdatedByID   string           `gorm:"type:varchar(100);column:updated_by_id"`
    UpdatedByName string           `gorm:"type:varchar(200);column:updated_by_name"`
    Contacts       []SupplierContact `gorm:"foreignKey:SupplierID"`
    OfferSuppliers []OfferSupplier   `gorm:"foreignKey:SupplierID"`
}

// SupplierContact represents a contact person for a supplier
type SupplierContact struct {
    BaseModel
    SupplierID uuid.UUID `gorm:"type:uuid;not null;index"`
    Supplier   *Supplier `gorm:"foreignKey:SupplierID"`
    Name       string    `gorm:"type:varchar(200);not null"`
    Title      string    `gorm:"type:varchar(200)"`
    Email      string    `gorm:"type:varchar(255)"`
    Phone      string    `gorm:"type:varchar(50)"`
    IsPrimary  bool      `gorm:"not null;default:false"`
    Notes      string    `gorm:"type:text"`
}

func (SupplierContact) TableName() string {
    return "supplier_contacts"
}

// OfferSupplier represents the many-to-many relationship between offers and suppliers
type OfferSupplier struct {
    BaseModel
    OfferID       uuid.UUID           `gorm:"type:uuid;not null;index"`
    Offer         *Offer              `gorm:"foreignKey:OfferID"`
    SupplierID    uuid.UUID           `gorm:"type:uuid;not null;index"`
    Supplier      *Supplier           `gorm:"foreignKey:SupplierID"`
    SupplierName  string              `gorm:"type:varchar(200)"`
    OfferTitle    string              `gorm:"type:varchar(200)"`
    Status        OfferSupplierStatus `gorm:"type:varchar(50);not null;default:'active'"`
    Notes         string              `gorm:"type:text"`
    CreatedByID   string              `gorm:"type:varchar(100);column:created_by_id"`
    CreatedByName string              `gorm:"type:varchar(200);column:created_by_name"`
    UpdatedByID   string              `gorm:"type:varchar(100);column:updated_by_id"`
    UpdatedByName string              `gorm:"type:varchar(200);column:updated_by_name"`
}

func (OfferSupplier) TableName() string {
    return "offer_suppliers"
}
```

### Update File Model

```go
type File struct {
    BaseModel
    Filename    string     `gorm:"type:varchar(255);not null"`
    ContentType string     `gorm:"type:varchar(100);not null"`
    Size        int64      `gorm:"not null"`
    StoragePath string     `gorm:"type:varchar(500);not null;unique"`
    OfferID     *uuid.UUID `gorm:"type:uuid;index"`
    Offer       *Offer     `gorm:"foreignKey:OfferID"`
    SupplierID  *uuid.UUID `gorm:"type:uuid;index"`      // NEW
    Supplier    *Supplier  `gorm:"foreignKey:SupplierID"` // NEW
    CustomerID  *uuid.UUID `gorm:"type:uuid;index"`      // NEW
    Customer    *Customer  `gorm:"foreignKey:CustomerID"` // NEW
}
```

### Update ActivityTargetType

```go
const (
    // ... existing constants
    ActivityTargetSupplier ActivityTargetType = "Supplier"
)
```

---

## API Endpoints

### Supplier CRUD

| Method | Path | Description |
|--------|------|-------------|
| GET | `/suppliers` | List with pagination, filters |
| POST | `/suppliers` | Create supplier |
| GET | `/suppliers/{id}` | Get with details |
| PUT | `/suppliers/{id}` | Full update |
| DELETE | `/suppliers/{id}` | Soft delete |

**Query Parameters for List:**
- `page`, `pageSize` (pagination)
- `search` (name, org_number)
- `city`, `country`
- `status` (active, inactive, pending, blacklisted)
- `category` (free text filter)
- `sortField`, `sortOrder`

### Supplier Property Updates

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/suppliers/{id}/status` | Update status |
| PUT | `/suppliers/{id}/notes` | Update notes |
| PUT | `/suppliers/{id}/category` | Update category |
| PUT | `/suppliers/{id}/payment-terms` | Update payment terms |

### Supplier Contacts

| Method | Path | Description |
|--------|------|-------------|
| GET | `/suppliers/{id}/contacts` | List contacts |
| POST | `/suppliers/{id}/contacts` | Create contact |
| PUT | `/suppliers/{id}/contacts/{contactId}` | Update contact |
| DELETE | `/suppliers/{id}/contacts/{contactId}` | Delete contact |

### Supplier Files

| Method | Path | Description |
|--------|------|-------------|
| GET | `/suppliers/{id}/files` | List supplier files |

### Offer-Supplier Relationships

| Method | Path | Description |
|--------|------|-------------|
| GET | `/offers/{id}/suppliers` | List suppliers on offer |
| POST | `/offers/{id}/suppliers` | Add supplier to offer |
| PUT | `/offers/{id}/suppliers/{osId}` | Update relationship |
| DELETE | `/offers/{id}/suppliers/{osId}` | Remove supplier |

### Supplier Reverse Lookups

| Method | Path | Description |
|--------|------|-------------|
| GET | `/suppliers/{id}/offers` | List offers with supplier |
| GET | `/suppliers/{id}/projects` | List projects (via offers) |

---

## DTOs

### Core Response DTOs

```go
type SupplierDTO struct {
    ID            uuid.UUID      `json:"id"`
    Name          string         `json:"name"`
    OrgNumber     string         `json:"orgNumber,omitempty"`
    Email         string         `json:"email,omitempty"`
    Phone         string         `json:"phone,omitempty"`
    Address       string         `json:"address,omitempty"`
    City          string         `json:"city,omitempty"`
    PostalCode    string         `json:"postalCode,omitempty"`
    Country       string         `json:"country"`
    Municipality  string         `json:"municipality,omitempty"`
    County        string         `json:"county,omitempty"`
    ContactPerson string         `json:"contactPerson,omitempty"`
    ContactEmail  string         `json:"contactEmail,omitempty"`
    ContactPhone  string         `json:"contactPhone,omitempty"`
    Status        SupplierStatus `json:"status"`
    Category      string         `json:"category,omitempty"`
    Notes         string         `json:"notes,omitempty"`
    PaymentTerms  string         `json:"paymentTerms,omitempty"`
    Website       string         `json:"website,omitempty"`
    CreatedAt     string         `json:"createdAt"`
    UpdatedAt     string         `json:"updatedAt"`
    CreatedByID   string         `json:"createdById,omitempty"`
    CreatedByName string         `json:"createdByName,omitempty"`
    UpdatedByID   string         `json:"updatedById,omitempty"`
    UpdatedByName string         `json:"updatedByName,omitempty"`
}

type SupplierWithDetailsDTO struct {
    SupplierDTO
    Stats        *SupplierStatsDTO    `json:"stats,omitempty"`
    Contacts     []SupplierContactDTO `json:"contacts,omitempty"`
    RecentOffers []OfferSupplierDTO   `json:"recentOffers,omitempty"`
}

type SupplierStatsDTO struct {
    TotalOffers     int `json:"totalOffers"`
    ActiveOffers    int `json:"activeOffers"`
    CompletedOffers int `json:"completedOffers"`
    TotalProjects   int `json:"totalProjects"`
}

type SupplierContactDTO struct {
    ID         uuid.UUID `json:"id"`
    SupplierID uuid.UUID `json:"supplierId"`
    Name       string    `json:"name"`
    Title      string    `json:"title,omitempty"`
    Email      string    `json:"email,omitempty"`
    Phone      string    `json:"phone,omitempty"`
    IsPrimary  bool      `json:"isPrimary"`
    Notes      string    `json:"notes,omitempty"`
    CreatedAt  string    `json:"createdAt"`
    UpdatedAt  string    `json:"updatedAt"`
}

type OfferSupplierDTO struct {
    ID           uuid.UUID           `json:"id"`
    OfferID      uuid.UUID           `json:"offerId"`
    OfferTitle   string              `json:"offerTitle,omitempty"`
    SupplierID   uuid.UUID           `json:"supplierId"`
    SupplierName string              `json:"supplierName,omitempty"`
    Status       OfferSupplierStatus `json:"status"`
    Notes        string              `json:"notes,omitempty"`
    CreatedAt    string              `json:"createdAt"`
    UpdatedAt    string              `json:"updatedAt"`
}

type TopSupplierDTO struct {
    ID         uuid.UUID `json:"id"`
    Name       string    `json:"name"`
    OrgNumber  string    `json:"orgNumber,omitempty"`
    Category   string    `json:"category,omitempty"`
    OfferCount int       `json:"offerCount"`
}
```

### Request DTOs

```go
type CreateSupplierRequest struct {
    Name          string         `json:"name" validate:"required,max=200"`
    OrgNumber     string         `json:"orgNumber,omitempty" validate:"max=20"`
    Email         string         `json:"email,omitempty" validate:"omitempty,email"`
    Phone         string         `json:"phone,omitempty" validate:"max=50"`
    Address       string         `json:"address,omitempty" validate:"max=500"`
    City          string         `json:"city,omitempty" validate:"max=100"`
    PostalCode    string         `json:"postalCode,omitempty" validate:"max=20"`
    Country       string         `json:"country" validate:"required,max=100"`
    Municipality  string         `json:"municipality,omitempty" validate:"max=100"`
    County        string         `json:"county,omitempty" validate:"max=100"`
    ContactPerson string         `json:"contactPerson,omitempty" validate:"max=200"`
    ContactEmail  string         `json:"contactEmail,omitempty" validate:"omitempty,email"`
    ContactPhone  string         `json:"contactPhone,omitempty" validate:"max=50"`
    Status        SupplierStatus `json:"status,omitempty"`
    Category      string         `json:"category,omitempty" validate:"max=200"`
    Notes         string         `json:"notes,omitempty"`
    PaymentTerms  string         `json:"paymentTerms,omitempty" validate:"max=200"`
    Website       string         `json:"website,omitempty" validate:"max=500"`
}

type UpdateSupplierRequest struct {
    Name          string         `json:"name" validate:"required,max=200"`
    OrgNumber     string         `json:"orgNumber,omitempty" validate:"max=20"`
    Email         string         `json:"email,omitempty" validate:"omitempty,email"`
    Phone         string         `json:"phone,omitempty" validate:"max=50"`
    Address       string         `json:"address,omitempty" validate:"max=500"`
    City          string         `json:"city,omitempty" validate:"max=100"`
    PostalCode    string         `json:"postalCode,omitempty" validate:"max=20"`
    Country       string         `json:"country" validate:"required,max=100"`
    Municipality  string         `json:"municipality,omitempty" validate:"max=100"`
    County        string         `json:"county,omitempty" validate:"max=100"`
    ContactPerson string         `json:"contactPerson,omitempty" validate:"max=200"`
    ContactEmail  string         `json:"contactEmail,omitempty" validate:"omitempty,email"`
    ContactPhone  string         `json:"contactPhone,omitempty" validate:"max=50"`
    Status        SupplierStatus `json:"status,omitempty"`
    Category      string         `json:"category,omitempty" validate:"max=200"`
    Notes         string         `json:"notes,omitempty"`
    PaymentTerms  string         `json:"paymentTerms,omitempty" validate:"max=200"`
    Website       string         `json:"website,omitempty" validate:"max=500"`
}

type CreateSupplierContactRequest struct {
    Name      string `json:"name" validate:"required,max=200"`
    Title     string `json:"title,omitempty" validate:"max=200"`
    Email     string `json:"email,omitempty" validate:"omitempty,email"`
    Phone     string `json:"phone,omitempty" validate:"max=50"`
    IsPrimary bool   `json:"isPrimary"`
    Notes     string `json:"notes,omitempty"`
}

type UpdateSupplierContactRequest struct {
    Name      string `json:"name" validate:"required,max=200"`
    Title     string `json:"title,omitempty" validate:"max=200"`
    Email     string `json:"email,omitempty" validate:"omitempty,email"`
    Phone     string `json:"phone,omitempty" validate:"max=50"`
    IsPrimary bool   `json:"isPrimary"`
    Notes     string `json:"notes,omitempty"`
}

type AddOfferSupplierRequest struct {
    SupplierID uuid.UUID           `json:"supplierId" validate:"required"`
    Status     OfferSupplierStatus `json:"status,omitempty"`
    Notes      string              `json:"notes,omitempty"`
}

type UpdateOfferSupplierRequest struct {
    Status OfferSupplierStatus `json:"status,omitempty"`
    Notes  string              `json:"notes,omitempty"`
}

// Property update requests
type UpdateSupplierStatusRequest struct {
    Status SupplierStatus `json:"status" validate:"required,oneof=active inactive pending blacklisted"`
}

type UpdateSupplierNotesRequest struct {
    Notes string `json:"notes"`
}

type UpdateSupplierCategoryRequest struct {
    Category string `json:"category" validate:"max=200"`
}

type UpdateSupplierPaymentTermsRequest struct {
    PaymentTerms string `json:"paymentTerms" validate:"max=200"`
}
```

### Update SearchResults

```go
type SearchResults struct {
    Customers []CustomerDTO `json:"customers"`
    Projects  []ProjectDTO  `json:"projects"`
    Offers    []OfferDTO    `json:"offers"`
    Contacts  []ContactDTO  `json:"contacts"`
    Suppliers []SupplierDTO `json:"suppliers"` // NEW
    Total     int           `json:"total"`
}
```

### Update FileDTO

```go
type FileDTO struct {
    ID          uuid.UUID  `json:"id"`
    Filename    string     `json:"filename"`
    ContentType string     `json:"contentType"`
    Size        int64      `json:"size"`
    OfferID     *uuid.UUID `json:"offerId,omitempty"`
    SupplierID  *uuid.UUID `json:"supplierId,omitempty"` // NEW
    CustomerID  *uuid.UUID `json:"customerId,omitempty"` // NEW
    CreatedAt   string     `json:"createdAt"`
}
```

---

## Files to Create

| File | Description |
|------|-------------|
| `migrations/00056_create_suppliers.sql` | Suppliers table |
| `migrations/00057_create_supplier_contacts.sql` | Supplier contacts table |
| `migrations/00058_create_offer_suppliers.sql` | Junction table |
| `migrations/00059_extend_files_polymorphic.sql` | Add supplier_id, customer_id to files |
| `internal/repository/supplier_repository.go` | Supplier data access |
| `internal/repository/supplier_contact_repository.go` | Contact data access |
| `internal/repository/offer_supplier_repository.go` | Junction table operations |
| `internal/service/supplier_service.go` | Business logic |
| `internal/http/handler/supplier_handler.go` | HTTP handlers |

## Files to Modify

| File | Changes |
|------|---------|
| `internal/domain/models.go` | Add Supplier, SupplierContact, OfferSupplier, enums |
| `internal/domain/dto.go` | Add all DTOs |
| `internal/mapper/mapper.go` | Add mapper functions |
| `internal/repository/file_repository.go` | Add ListBySupplier, ListByCustomer |
| `internal/service/file_service.go` | Support supplierId/customerId in Upload |
| `internal/service/dashboard_service.go` | Add supplier search |
| `internal/http/router/router.go` | Add all supplier routes |
| `cmd/api/main.go` | Wire up dependencies |

---

## Implementation Order

### Phase 1: Database Migrations (sc-294)
1. Create all 4 migration files
2. Run `make migrate-up`
3. Verify schema with `make migrate-status`

### Phase 2: Supplier CRUD (sc-295)
1. Add domain models and enums to `models.go`
2. Add DTOs to `dto.go`
3. Create `supplier_repository.go`
4. Create `supplier_service.go`
5. Create `supplier_handler.go`
6. Add mapper functions
7. Wire up in `main.go` and `router.go`
8. Test with curl/Postman

### Phase 3: Supplier Contacts (sc-296)
1. Add SupplierContact model
2. Add DTOs
3. Create `supplier_contact_repository.go`
4. Add service methods
5. Add handler endpoints
6. Wire up routes

### Phase 4: Offer-Supplier Relationships (sc-297)
1. Add OfferSupplier model
2. Add DTOs
3. Create `offer_supplier_repository.go`
4. Add service methods
5. Add handler endpoints
6. Wire up routes under `/offers/{id}/suppliers`

### Phase 5: File Attachments (sc-299)
1. Update File model
2. Update FileDTO
3. Add repository methods
4. Update file service
5. Add `/suppliers/{id}/files` endpoint

### Phase 6: Global Search (sc-300)
1. Add `SupplierRepository.Search()`
2. Update SearchResults DTO
3. Update `DashboardService.Search()`

### Phase 7: Stats & Dashboard (sc-301)
1. Add `GetSupplierStats()` repository method
2. Add `GetTopSuppliers()` method
3. Add `/suppliers/{id}/offers` and `/suppliers/{id}/projects` endpoints

### Phase 8: Testing (sc-302, sc-303)
1. Repository tests
2. Service tests
3. Handler integration tests
