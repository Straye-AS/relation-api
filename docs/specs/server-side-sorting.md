# Server-Side Sorting Specification

**Feature**: Server-side sorting for List endpoints
**Endpoints**: `/offers`, `/projects`, `/customers`
**Status**: Specification
**Date**: 2025-12-10

---

## Overview

Add server-side sorting capability to the Offers, Projects, and Customers list endpoints. The frontend is already sending sorting parameters (`sortBy`, `sortOrder`) that the backend needs to handle.

**Request Format**:
```
GET /offers?page=1&pageSize=20&sortBy=value&sortOrder=desc
```

---

## Current State Analysis

### Customers (`/customers`)
- **Already has sorting** via `sortBy` parameter using `CustomerSortOption` enum
- Current options: `name_asc`, `name_desc`, `created_desc`, `created_asc`, `city_asc`, `city_desc`
- **Gap**: Does not support the `sortBy` + `sortOrder` pattern the frontend is sending
- **Gap**: Missing `status` and `tier` sort options

### Offers (`/offers`)
- **No sorting support** - hardcoded to `ORDER BY created_at DESC`
- Needs full implementation

### Projects (`/projects`)
- **No sorting support** - hardcoded to `ORDER BY created_at DESC`
- Needs full implementation

---

## Acceptance Criteria

### AC-1: Offers Sorting
- [ ] Accept `sortBy` query parameter with allowed values: `title`, `customerName`, `phase`, `dueDate`, `value`, `probability`, `updatedAt`, `createdAt`
- [ ] Accept `sortOrder` query parameter with allowed values: `asc`, `desc` (default: `desc`)
- [ ] Default sort when no parameters provided: `updated_at DESC`
- [ ] Invalid `sortBy` values must fall back to default (do not error)
- [ ] Invalid `sortOrder` values must fall back to `desc`

### AC-2: Projects Sorting
- [ ] Accept `sortBy` query parameter with allowed values: `name`, `customerName`, `budget`, `startDate`, `createdAt`, `status`
- [ ] Accept `sortOrder` query parameter with allowed values: `asc`, `desc` (default: `desc`)
- [ ] Default sort when no parameters provided: `updated_at DESC`
- [ ] Invalid `sortBy` values must fall back to default (do not error)
- [ ] Invalid `sortOrder` values must fall back to `desc`

### AC-3: Customers Sorting (Update Existing)
- [ ] Change `sortBy` parameter to accept: `name`, `city`, `status`, `tier`, `createdAt`
- [ ] Add `sortOrder` query parameter with allowed values: `asc`, `desc` (default: `asc` for name, `desc` for others)
- [ ] Default sort when no parameters provided: `name ASC`
- [ ] Maintain backward compatibility with existing `name_asc`, `name_desc` format (deprecated but functional)

### AC-4: Security
- [ ] All `sortBy` values must be whitelisted - no raw SQL injection possible
- [ ] Column names must be mapped from camelCase parameters to snake_case DB columns

### AC-5: Swagger Documentation
- [ ] Update Swagger annotations on all three handlers to document new parameters
- [ ] Document allowed enum values for `sortBy` on each endpoint
- [ ] Document `sortOrder` parameter

---

## Technical Design

### Column Mapping (sortBy parameter -> Database column)

**Offers**:
| Parameter | Database Column | Type |
|-----------|-----------------|------|
| `title` | `title` | Direct |
| `customerName` | `customer_name` | Denormalized |
| `phase` | `phase` | Direct |
| `dueDate` | `due_date` | Direct |
| `value` | `value` | Direct |
| `probability` | `probability` | Direct |
| `updatedAt` | `updated_at` | Direct |
| `createdAt` | `created_at` | Direct |

**Projects**:
| Parameter | Database Column | Type |
|-----------|-----------------|------|
| `name` | `name` | Direct |
| `customerName` | `customer_name` | Denormalized |
| `budget` | `budget` | Direct |
| `startDate` | `start_date` | Direct |
| `createdAt` | `created_at` | Direct |
| `status` | `status` | Direct |

**Customers**:
| Parameter | Database Column | Type |
|-----------|-----------------|------|
| `name` | `name` | Direct |
| `city` | `city` | Direct |
| `status` | `status` | Direct |
| `tier` | `tier` | Direct |
| `createdAt` | `created_at` | Direct |

### Implementation Pattern

Follow the existing Customer repository pattern with improvements:

```go
// SortConfig defines sorting configuration for a list query
type SortConfig struct {
    Field     string // Database column name (validated)
    Direction string // "ASC" or "DESC"
}

// Example whitelist map for offers
var offerSortFields = map[string]string{
    "title":        "title",
    "customerName": "customer_name",
    "phase":        "phase",
    "dueDate":      "due_date",
    "value":        "value",
    "probability":  "probability",
    "updatedAt":    "updated_at",
    "createdAt":    "created_at",
}
```

### Layer Changes

1. **Handler Layer**: Parse `sortBy` and `sortOrder` query params, validate `sortOrder`
2. **Service Layer**: Pass through sort config (no business logic needed for sorting)
3. **Repository Layer**: Build ORDER BY clause from validated sort config

### Files to Modify

| File | Change |
|------|--------|
| `internal/repository/offer_repository.go` | Add `ListWithFilters` sort support, add sort field whitelist |
| `internal/repository/project_repository.go` | Add `ListWithFilters` sort support, add sort field whitelist |
| `internal/repository/customer_repository.go` | Update to support new `sortBy`/`sortOrder` pattern |
| `internal/service/offer_service.go` | Pass sort config to repository |
| `internal/service/project_service.go` | Pass sort config to repository |
| `internal/service/customer_service.go` | Update to pass new sort config |
| `internal/http/handler/offer_handler.go` | Parse sort params, update Swagger |
| `internal/http/handler/project_handler.go` | Parse sort params, update Swagger |
| `internal/http/handler/customer_handler.go` | Update sort param parsing, update Swagger |

---

## Test Cases

### Unit Tests (Handler)
- Parse valid `sortBy` and `sortOrder` combinations
- Default `sortOrder` to `desc` when not provided
- Handle invalid `sortBy` gracefully (use default)
- Handle invalid `sortOrder` gracefully (use default)

### Integration Tests (Repository)
- Verify sort order is correctly applied to results
- Verify each allowed sort field produces valid SQL
- Verify default sorting when no params provided
- Verify pagination works correctly with sorting

---

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| SQL injection via sortBy | Whitelist validation - only allow mapped column names |
| Performance on large datasets | All sort columns should have indexes (verify) |
| Breaking existing frontend | Use defaults that match current behavior |

---

## Out of Scope

- Multi-column sorting (e.g., `sortBy=phase,value`)
- Custom null handling (NULLS FIRST/LAST)
- Case-insensitive sorting for text fields

---

## Implementation Notes

1. **No JOINs needed**: `customer_name` is denormalized on both `offers` and `projects` tables
2. **Consistent API**: All three endpoints should follow the same `sortBy`/`sortOrder` pattern
3. **Graceful degradation**: Invalid values should use defaults, not return errors
4. **Database indexes**: Verify indexes exist for commonly sorted columns (especially `value`, `due_date`, `created_at`)
