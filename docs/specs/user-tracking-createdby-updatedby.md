# Specification: User Tracking (CreatedBy/UpdatedBy)

**Status**: Draft - Pending Review
**Author**: Scrum Master (AI)
**Date**: 2025-12-12
**Related Stories**: Standalone feature (not related to sc-254)

---

## 1. Overview

### 1.1 Problem Statement

Currently, the Straye Relation API tracks **when** entities are created and modified (`CreatedAt`, `UpdatedAt` via `BaseModel`), but not **who** performed these actions. While the Activity log captures `CreatorID` and `CreatorName` for audit trail entries, the entities themselves do not store this information directly.

This creates gaps in:
- Quick attribution queries ("Who created this customer?")
- Filtering by creator/modifier without joining activity tables
- Direct entity-level audit visibility

### 1.2 Proposed Solution

Add `CreatedBy` and `UpdatedBy` fields to main business entities to track user attribution directly on the entity. This complements (not replaces) the Activity log system.

### 1.3 Scope

**In Scope:**
- Customer entity
- Project entity
- Offer entity
- Contact entity

**Out of Scope (Future Consideration):**
- Deal entity (already has `OwnerID/OwnerName` for similar purpose)
- File entity (could be added later)
- Activity entity (already has `CreatorID`)
- BudgetItem entity (derived from parent)
- Notification entity (system-generated)

---

## 2. Technical Design

### 2.1 Design Decision: Embedded Struct vs Per-Entity Fields

**Option A: Create AuditableModel (Embedded Struct)**
```go
type AuditableModel struct {
    BaseModel
    CreatedByID   string `gorm:"type:varchar(100);column:created_by_id"`
    CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
    UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
    UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
}
```
- Pros: DRY, consistent across entities
- Cons: Requires refactoring all entities to use AuditableModel instead of BaseModel

**Option B: Add Fields to Specific Entities**
```go
type Customer struct {
    BaseModel
    CreatedByID   string `gorm:"type:varchar(100);column:created_by_id"`
    CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
    UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
    UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
    // ... other fields
}
```
- Pros: No breaking changes to BaseModel, selective application
- Cons: Some code duplication

**Recommendation**: **Option B** - Add fields to specific entities. This allows:
1. Incremental rollout
2. No breaking changes to BaseModel (used elsewhere)
3. Flexibility to exclude entities where tracking doesn't make sense
4. Aligns with existing patterns (Activity already has separate creator fields)

### 2.2 Field Specifications

| Field | Type | Column Name | Nullable | Description |
|-------|------|-------------|----------|-------------|
| CreatedByID | string | created_by_id | YES | User ID from UserContext (can be API key or JWT) |
| CreatedByName | string | created_by_name | YES | Display name at time of creation (denormalized) |
| UpdatedByID | string | updated_by_id | YES | User ID of last modifier |
| UpdatedByName | string | updated_by_name | YES | Display name of last modifier (denormalized) |

**Why Nullable?**
- Historical data will have NULL values
- System/migration operations may not have user context
- Aligns with Activity.CreatorID which is also optional

**Why String (not UUID)?**
- UserContext.UserID is uuid.UUID but stored as string in Activity.CreatorID
- API key users don't have UUIDs
- Consistent with existing patterns in the codebase

**Why Denormalized Names?**
- Follows existing pattern (Activity.CreatorName, Offer.CustomerName, etc.)
- Avoids joins for display purposes
- Name at time of action is preserved even if user name changes

### 2.3 Behavioral Rules

1. **CreatedBy fields**: Set ONCE on entity creation, NEVER modified thereafter
2. **UpdatedBy fields**: Updated on EVERY modification
3. **Null handling**: If no UserContext available, fields remain NULL (system operations)
4. **Name resolution**: Use `UserContext.DisplayName` for name fields

---

## 3. Database Migration

### 3.1 Migration File: `00042_add_user_tracking_fields.sql`

```sql
-- +goose Up
-- Add user tracking fields to main business entities

-- Customer table
ALTER TABLE customers
ADD COLUMN IF NOT EXISTS created_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(200),
ADD COLUMN IF NOT EXISTS updated_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS updated_by_name VARCHAR(200);

-- Project table
ALTER TABLE projects
ADD COLUMN IF NOT EXISTS created_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(200),
ADD COLUMN IF NOT EXISTS updated_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS updated_by_name VARCHAR(200);

-- Offer table
ALTER TABLE offers
ADD COLUMN IF NOT EXISTS created_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(200),
ADD COLUMN IF NOT EXISTS updated_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS updated_by_name VARCHAR(200);

-- Contact table
ALTER TABLE contacts
ADD COLUMN IF NOT EXISTS created_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS created_by_name VARCHAR(200),
ADD COLUMN IF NOT EXISTS updated_by_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS updated_by_name VARCHAR(200);

-- Create indexes for filtering by creator (common query pattern)
CREATE INDEX IF NOT EXISTS idx_customers_created_by_id ON customers(created_by_id);
CREATE INDEX IF NOT EXISTS idx_projects_created_by_id ON projects(created_by_id);
CREATE INDEX IF NOT EXISTS idx_offers_created_by_id ON offers(created_by_id);
CREATE INDEX IF NOT EXISTS idx_contacts_created_by_id ON contacts(created_by_id);

-- +goose Down
-- Remove indexes first
DROP INDEX IF EXISTS idx_customers_created_by_id;
DROP INDEX IF EXISTS idx_projects_created_by_id;
DROP INDEX IF EXISTS idx_offers_created_by_id;
DROP INDEX IF EXISTS idx_contacts_created_by_id;

-- Remove columns
ALTER TABLE customers
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

ALTER TABLE projects
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

ALTER TABLE offers
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

ALTER TABLE contacts
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;
```

### 3.2 Data Backfill Considerations

Historical data will have NULL values for these fields. Options:
1. **Accept NULL** (recommended) - cleanest, honest about data provenance
2. **Backfill from Activity log** - possible but complex, may not be accurate
3. **Set to "system"** - could mislead

**Recommendation**: Accept NULL for historical data. New data going forward will have proper attribution.

---

## 4. Domain Model Changes

### 4.1 Updated Models (internal/domain/models.go)

Add to Customer, Project, Offer, Contact:

```go
// User tracking fields (denormalized for performance)
CreatedByID   string `gorm:"type:varchar(100);column:created_by_id"`
CreatedByName string `gorm:"type:varchar(200);column:created_by_name"`
UpdatedByID   string `gorm:"type:varchar(100);column:updated_by_id"`
UpdatedByName string `gorm:"type:varchar(200);column:updated_by_name"`
```

---

## 5. DTO Changes

### 5.1 Response DTOs (internal/domain/dto.go)

Add to CustomerDTO, ProjectDTO, OfferDTO, ContactDTO:

```go
// User tracking
CreatedByID   string `json:"createdById,omitempty"`
CreatedByName string `json:"createdByName,omitempty"`
UpdatedByID   string `json:"updatedById,omitempty"`
UpdatedByName string `json:"updatedByName,omitempty"`
```

### 5.2 Request DTOs

**No changes required** - CreatedBy/UpdatedBy are set by the system, not by API callers.

---

## 6. Mapper Changes

### 6.1 Update Mapper Functions (internal/mapper/mapper.go)

For each entity mapper (ToCustomerDTO, ToProjectDTO, ToOfferDTO, ToContactDTO), add:

```go
CreatedByID:   model.CreatedByID,
CreatedByName: model.CreatedByName,
UpdatedByID:   model.UpdatedByID,
UpdatedByName: model.UpdatedByName,
```

---

## 7. Service Layer Changes

### 7.1 Pattern for Create Operations

In each service's Create method:

```go
func (s *CustomerService) Create(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.CustomerDTO, error) {
    // ... existing validation ...

    customer := &domain.Customer{
        // ... existing fields ...
    }

    // Set user tracking fields
    if userCtx, ok := auth.FromContext(ctx); ok {
        customer.CreatedByID = userCtx.UserID.String()
        customer.CreatedByName = userCtx.DisplayName
        customer.UpdatedByID = userCtx.UserID.String()
        customer.UpdatedByName = userCtx.DisplayName
    }

    // ... rest of method ...
}
```

### 7.2 Pattern for Update Operations

In each service's Update method:

```go
func (s *CustomerService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateCustomerRequest) (*domain.CustomerDTO, error) {
    // ... existing fetch and validation ...

    // Update fields (but NOT CreatedBy)
    customer.Name = req.Name
    // ... other field updates ...

    // Set updated by
    if userCtx, ok := auth.FromContext(ctx); ok {
        customer.UpdatedByID = userCtx.UserID.String()
        customer.UpdatedByName = userCtx.DisplayName
    }

    // ... rest of method ...
}
```

### 7.3 Services to Modify

1. `internal/service/customer_service.go`
2. `internal/service/project_service.go`
3. `internal/service/offer_service.go`
4. `internal/service/contact_service.go`

Each service needs updates to:
- Create method
- Update method
- Any property-specific update methods (e.g., UpdateStatus, UpdatePhase)

---

## 8. API Response Examples

### 8.1 Customer Response (After Implementation)

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Byggmester Berg AS",
  "orgNumber": "987654321",
  "status": "active",
  "createdAt": "2025-12-12T10:30:00Z",
  "updatedAt": "2025-12-12T14:15:00Z",
  "createdById": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "createdByName": "Ola Nordmann",
  "updatedById": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
  "updatedByName": "Kari Hansen"
}
```

---

## 9. Acceptance Criteria

### 9.1 Database
- [ ] Migration adds all 4 fields to customers, projects, offers, contacts tables
- [ ] Migration creates indexes on created_by_id columns
- [ ] Migration is reversible (down migration works)
- [ ] Existing data is preserved with NULL values for new fields

### 9.2 Models
- [ ] Customer, Project, Offer, Contact models include CreatedByID, CreatedByName, UpdatedByID, UpdatedByName
- [ ] Fields use correct GORM tags (varchar(100), varchar(200), column names)

### 9.3 DTOs
- [ ] CustomerDTO, ProjectDTO, OfferDTO, ContactDTO include tracking fields
- [ ] Fields use correct JSON tags (camelCase, omitempty)

### 9.4 Mappers
- [ ] All entity mappers copy tracking fields from model to DTO

### 9.5 Services - Create Operations
- [ ] CustomerService.Create sets CreatedByID, CreatedByName, UpdatedByID, UpdatedByName
- [ ] ProjectService.Create sets CreatedByID, CreatedByName, UpdatedByID, UpdatedByName
- [ ] OfferService.Create sets CreatedByID, CreatedByName, UpdatedByID, UpdatedByName
- [ ] ContactService.Create sets CreatedByID, CreatedByName, UpdatedByID, UpdatedByName
- [ ] All create operations handle missing UserContext gracefully (fields remain empty)

### 9.6 Services - Update Operations
- [ ] CustomerService.Update sets UpdatedByID, UpdatedByName (NOT CreatedBy)
- [ ] ProjectService.Update sets UpdatedByID, UpdatedByName (NOT CreatedBy)
- [ ] OfferService.Update sets UpdatedByID, UpdatedByName (NOT CreatedBy)
- [ ] ContactService.Update sets UpdatedByID, UpdatedByName (NOT CreatedBy)
- [ ] All property-specific update methods also set UpdatedBy fields

### 9.7 API Responses
- [ ] GET endpoints return createdById, createdByName, updatedById, updatedByName
- [ ] Fields are omitted (not null) when empty (via omitempty tag)

### 9.8 Testing
- [ ] Unit tests verify CreatedBy is set on creation
- [ ] Unit tests verify UpdatedBy is updated on modification
- [ ] Unit tests verify CreatedBy is NOT modified on updates
- [ ] Integration tests verify fields are persisted to database

---

## 10. Implementation Plan

### Phase 1: Database Migration
1. Create migration file `00042_add_user_tracking_fields.sql`
2. Run migration in development
3. Verify columns and indexes created

### Phase 2: Model Updates
1. Add fields to Customer model
2. Add fields to Project model
3. Add fields to Offer model
4. Add fields to Contact model
5. Update DTO structs
6. Update mapper functions

### Phase 3: Service Layer Updates
1. Update CustomerService (Create + Update)
2. Update ProjectService (Create + Update + property methods)
3. Update OfferService (Create + Update + property methods)
4. Update ContactService (Create + Update)

### Phase 4: Testing
1. Write unit tests for tracking logic
2. Run integration tests
3. Manual verification via API calls

### Phase 5: Documentation
1. Update Swagger annotations if needed
2. Update API documentation

---

## 11. Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking changes to existing API consumers | Medium | Low | Fields are additive, use omitempty |
| Performance impact from additional columns | Low | Low | Indexes on frequently queried columns |
| Migration failure on large tables | Medium | Low | Use IF NOT EXISTS, run during low-traffic |
| Missing UserContext in edge cases | Low | Medium | Handle gracefully with empty strings |

---

## 12. Future Considerations

1. **Filtering by creator**: Add query parameters `?createdBy=userId` to list endpoints
2. **Bulk attribution report**: "All entities created by user X"
3. **Extend to other entities**: File, Deal, BudgetItem
4. **Audit log enhancement**: Cross-reference entity tracking with Activity log

---

## 13. Reviewer Checklist

- [ ] Design decision (Option B) is acceptable
- [ ] Field types and nullability are appropriate
- [ ] Migration approach is safe
- [ ] No security concerns with exposing user IDs
- [ ] Performance impact is acceptable
- [ ] Testing strategy is sufficient
