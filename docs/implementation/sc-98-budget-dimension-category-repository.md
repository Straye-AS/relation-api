# Implementation Plan: sc-98 BudgetDimensionCategory Repository

## Story Details
- **ID**: sc-98
- **Title**: Implement BudgetDimensionCategory repository and seed data
- **Epic**: 39 - Offer & Budget Management
- **Estimate**: 4 points
- **Branch**: `feature/sc-98/implement-budgetdimensioncategory`

## Description
Create repository for budget dimension categories with predefined categories. These are construction-specific budget line categories used in offers and projects.

## Acceptance Criteria Checklist

- [ ] **AC1**: BudgetDimensionCategoryRepository with CRUD operations
- [ ] **AC2**: Seed data with construction-specific categories (see list below)
- [ ] **AC3**: List all categories (with optional active filter)
- [ ] **AC4**: GetByID for category lookups
- [ ] **AC5**: Category usage tracking (count of BudgetDimensions using each category)
- [ ] **AC6**: Repository tests with good coverage

## Technical Context

### Existing Infrastructure
- **Database table**: `budget_dimension_categories` already exists (migration 00004)
- **Domain model**: `BudgetDimensionCategory` defined in `internal/domain/models.go`
- **Category constants**: `BudgetCategoryID` enum constants exist in `models.go`

### Database Schema (from migration 00004)
```sql
CREATE TABLE budget_dimension_categories (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Domain Model
```go
type BudgetDimensionCategory struct {
    ID           string    `gorm:"type:varchar(50);primaryKey"`
    Name         string    `gorm:"type:varchar(200);not null"`
    Description  string    `gorm:"type:text"`
    DisplayOrder int       `gorm:"not null;default:0;column:display_order"`
    IsActive     bool      `gorm:"not null;default:true;column:is_active"`
    CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
```

## Implementation Requirements

### 1. Repository Interface & Implementation

Create `internal/repository/budget_dimension_category_repository.go`:

```go
type BudgetDimensionCategoryRepository struct {
    db *gorm.DB
}

func NewBudgetDimensionCategoryRepository(db *gorm.DB) *BudgetDimensionCategoryRepository
```

**Required Methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| Create | `(ctx, *BudgetDimensionCategory) error` | Create new category |
| GetByID | `(ctx, id string) (*BudgetDimensionCategory, error)` | Get by primary key |
| Update | `(ctx, *BudgetDimensionCategory) error` | Update existing category |
| Delete | `(ctx, id string) error` | Soft delete (set is_active=false) or hard delete |
| List | `(ctx, activeOnly bool) ([]BudgetDimensionCategory, error)` | List all, optionally filtered by active |
| GetUsageCount | `(ctx, categoryID string) (int, error)` | Count BudgetDimensions using this category |
| GetAllWithUsageCounts | `(ctx, activeOnly bool) ([]CategoryWithUsage, error)` | List with usage stats |
| EnsureSeeded | `(ctx) error` | Ensure seed data exists (upsert) |

### 2. Seed Data Categories

Based on the existing `BudgetCategoryID` constants in models.go, create seed data for these construction-specific categories:

| ID | Name | Description | Display Order |
|----|------|-------------|---------------|
| `steel_structure` | Steel Structure | Primary steel framework and structural elements | 1 |
| `hybrid_structure` | Hybrid Structure | Combined steel and other material structures | 2 |
| `roofing` | Roofing | Roof installation, materials and labor | 3 |
| `cladding` | Cladding | Wall cladding and facade materials | 4 |
| `foundation` | Foundation | Concrete foundation and groundwork | 5 |
| `assembly` | Assembly | On-site assembly and installation labor | 6 |
| `transport` | Transport | Delivery and logistics costs | 7 |
| `engineering` | Engineering | Design and engineering services | 8 |
| `project_management` | Project Management | PM overhead and coordination | 9 |
| `crane_rigging` | Crane & Rigging | Crane rental and rigging services | 10 |
| `miscellaneous` | Miscellaneous | Other uncategorized costs | 11 |
| `contingency` | Contingency | Risk buffer and unforeseen costs | 12 |

### 3. CategoryWithUsage DTO

Create a struct to return category with usage statistics:

```go
type CategoryWithUsage struct {
    domain.BudgetDimensionCategory
    UsageCount int `json:"usageCount"`
}
```

### 4. Seed Migration

Create a new migration file `migrations/00013_seed_budget_dimension_categories.sql`:

```sql
-- +goose Up
INSERT INTO budget_dimension_categories (id, name, description, display_order, is_active) VALUES
    ('steel_structure', 'Steel Structure', 'Primary steel framework and structural elements', 1, true),
    -- ... etc
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    display_order = EXCLUDED.display_order,
    updated_at = CURRENT_TIMESTAMP;
```

### 5. Repository Tests

Create `tests/repository/budget_dimension_category_repository_test.go`:

**Test cases:**
1. TestCreate - Create new category
2. TestGetByID - Get existing and non-existing
3. TestUpdate - Update category fields
4. TestDelete - Delete category
5. TestList - List all and active-only
6. TestGetUsageCount - Count with and without usage
7. TestEnsureSeeded - Verify idempotent seeding

## Architecture Notes

### Pattern Reference (from CustomerRepository)
```go
// Follow this established pattern:
func (r *BudgetDimensionCategoryRepository) GetByID(ctx context.Context, id string) (*domain.BudgetDimensionCategory, error) {
    var category domain.BudgetDimensionCategory
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
    if err != nil {
        return nil, err
    }
    return &category, nil
}
```

### Key Differences from Other Repositories
1. **No UUID primary key** - Uses string ID (varchar(50))
2. **No company filter** - Categories are system-wide, not tenant-specific
3. **Relatively static data** - Categories change rarely
4. **Seed data requirement** - Must ensure default categories exist

### Error Handling
- Return `gorm.ErrRecordNotFound` when category not found
- Wrap errors with context: `fmt.Errorf("failed to get category %s: %w", id, err)`

## Dependencies

### This story depends on:
- Migration 00004 (already exists)

### Stories that depend on this:
- sc-99: BudgetDimension repository (needs category lookups)
- sc-100: BudgetDimension service
- sc-101-104: Offer management

## File Checklist

- [ ] `internal/repository/budget_dimension_category_repository.go`
- [ ] `migrations/00013_seed_budget_dimension_categories.sql`
- [ ] `tests/repository/budget_dimension_category_repository_test.go`

## Definition of Done

1. All acceptance criteria met
2. Repository follows established patterns
3. Seed migration creates all 12 categories
4. Tests pass with good coverage
5. Code reviewed and approved
6. Branch merged to main
