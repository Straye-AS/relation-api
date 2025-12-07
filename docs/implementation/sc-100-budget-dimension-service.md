# Implementation Plan: sc-100 BudgetDimension Service

## Story Details
- **ID**: sc-100
- **Title**: Implement BudgetDimension service with business logic
- **Epic**: 39 - Offer & Budget Management
- **Branch**: `feature/sc-100/implement-budgetdimension-service`
- **Depends On**: sc-99 (BudgetDimension Repository - completed)

## Description
Create BudgetDimensionService for managing budget dimensions with validation, calculations, activity logging, and parent entity synchronization.

## Acceptance Criteria Checklist

- [ ] **AC1**: BudgetDimensionService with Create, Update, Delete methods
- [ ] **AC2**: ValidateMarginOverride logic
- [ ] **AC3**: CalculateTotals for offer/project
- [ ] **AC4**: CloneDimensions (offer to project)
- [ ] **AC5**: ReorderDimensions with sort_order
- [ ] **AC6**: Activity logging for all mutations
- [ ] **AC7**: Service tests with good coverage

## Technical Context

### Existing Infrastructure

**Repository (implemented in sc-99)**:
- File: `internal/repository/budget_dimension_repository.go`
- Available methods:
  - `Create(ctx, *BudgetDimension) error`
  - `GetByID(ctx, uuid.UUID) (*BudgetDimension, error)`
  - `Update(ctx, *BudgetDimension) error`
  - `Delete(ctx, uuid.UUID) error`
  - `GetByParent(ctx, parentType, parentID) ([]BudgetDimension, error)`
  - `GetByParentPaginated(ctx, parentType, parentID, page, pageSize) ([]BudgetDimension, int64, error)`
  - `DeleteByParent(ctx, parentType, parentID) error`
  - `GetBudgetSummary(ctx, parentType, parentID) (*BudgetSummary, error)`
  - `GetTotalCost(ctx, parentType, parentID) (float64, error)`
  - `GetTotalRevenue(ctx, parentType, parentID) (float64, error)`
  - `ReorderDimensions(ctx, parentType, parentID, orderedIDs) error`
  - `GetMaxDisplayOrder(ctx, parentType, parentID) (int, error)`
  - `Count(ctx, parentType, parentID) (int, error)`

**Domain Models & DTOs (from domain/models.go and domain/dto.go)**:
- `BudgetDimension` model with `ParentType`, `ParentID`, `CategoryID`, `CustomName`, `Cost`, `Revenue`, `TargetMarginPercent`, `MarginOverride`
- `BudgetDimensionDTO`, `BudgetSummaryDTO`
- `CreateBudgetDimensionRequest`, `UpdateBudgetDimensionRequest`

**Mapper Functions (from mapper/mapper.go)**:
- `ToBudgetDimensionDTO(dim *domain.BudgetDimension) domain.BudgetDimensionDTO`
- `ToBudgetSummaryDTO(parentType, parentID, dimensions) domain.BudgetSummaryDTO`

**Database Trigger**: Margin calculations handled by PostgreSQL trigger (calculates `revenue` from `cost` when `margin_override=true`, computes `margin_percent`)

### Related Repositories Needed
- `BudgetDimensionCategoryRepository` - for category validation
- `OfferRepository` - for verifying offer parent exists
- `ProjectRepository` - for verifying project parent exists
- `ActivityRepository` - for audit logging

---

## Implementation Requirements

### 1. Service Struct & Constructor

Create `internal/service/budget_dimension_service.go`:

```go
package service

import (
    "context"
    "errors"
    "fmt"

    "github.com/google/uuid"
    "github.com/straye-as/relation-api/internal/auth"
    "github.com/straye-as/relation-api/internal/domain"
    "github.com/straye-as/relation-api/internal/mapper"
    "github.com/straye-as/relation-api/internal/repository"
    "go.uber.org/zap"
    "gorm.io/gorm"
)

// Service-level errors
var (
    ErrBudgetDimensionNotFound     = errors.New("budget dimension not found")
    ErrInvalidParentType           = errors.New("invalid parent type")
    ErrParentNotFound              = errors.New("parent entity not found")
    ErrInvalidCost                 = errors.New("cost must be greater than 0")
    ErrInvalidRevenue              = errors.New("revenue must be greater than or equal to 0")
    ErrInvalidTargetMargin         = errors.New("target margin percent must be between 0 and 100")
    ErrInvalidCategory             = errors.New("category not found or inactive")
    ErrMissingName                 = errors.New("either categoryId or customName must be provided")
    ErrSourceDimensionsNotFound    = errors.New("source has no budget dimensions to clone")
)

type BudgetDimensionService struct {
    dimensionRepo *repository.BudgetDimensionRepository
    categoryRepo  *repository.BudgetDimensionCategoryRepository
    offerRepo     *repository.OfferRepository
    projectRepo   *repository.ProjectRepository
    activityRepo  *repository.ActivityRepository
    logger        *zap.Logger
}

func NewBudgetDimensionService(
    dimensionRepo *repository.BudgetDimensionRepository,
    categoryRepo *repository.BudgetDimensionCategoryRepository,
    offerRepo *repository.OfferRepository,
    projectRepo *repository.ProjectRepository,
    activityRepo *repository.ActivityRepository,
    logger *zap.Logger,
) *BudgetDimensionService {
    return &BudgetDimensionService{
        dimensionRepo: dimensionRepo,
        categoryRepo:  categoryRepo,
        offerRepo:     offerRepo,
        projectRepo:   projectRepo,
        activityRepo:  activityRepo,
        logger:        logger,
    }
}
```

---

### 2. Core CRUD Methods

#### 2.1 Create

```go
func (s *BudgetDimensionService) Create(ctx context.Context, req *domain.CreateBudgetDimensionRequest) (*domain.BudgetDimensionDTO, error) {
    // 1. Validate parent exists
    if err := s.validateParentExists(ctx, req.ParentType, req.ParentID); err != nil {
        return nil, err
    }

    // 2. Validate name source
    if req.CategoryID == nil && req.CustomName == "" {
        return nil, ErrMissingName
    }

    // 3. Validate category if provided
    if req.CategoryID != nil {
        if err := s.validateCategory(ctx, *req.CategoryID); err != nil {
            return nil, err
        }
    }

    // 4. Validate cost/revenue/margin
    if err := s.validateFinancials(req.Cost, req.Revenue, req.TargetMarginPercent, req.MarginOverride); err != nil {
        return nil, err
    }

    // 5. Set display order if not provided
    displayOrder := req.DisplayOrder
    if displayOrder == 0 {
        maxOrder, err := s.dimensionRepo.GetMaxDisplayOrder(ctx, req.ParentType, req.ParentID)
        if err != nil {
            return nil, fmt.Errorf("failed to get max display order: %w", err)
        }
        displayOrder = maxOrder + 1
    }

    // 6. Create dimension
    dimension := &domain.BudgetDimension{
        ParentType:          req.ParentType,
        ParentID:            req.ParentID,
        CategoryID:          req.CategoryID,
        CustomName:          req.CustomName,
        Cost:                req.Cost,
        Revenue:             req.Revenue,
        TargetMarginPercent: req.TargetMarginPercent,
        MarginOverride:      req.MarginOverride,
        Description:         req.Description,
        Quantity:            req.Quantity,
        Unit:                req.Unit,
        DisplayOrder:        displayOrder,
    }

    if err := s.dimensionRepo.Create(ctx, dimension); err != nil {
        return nil, fmt.Errorf("failed to create budget dimension: %w", err)
    }

    // 7. Reload with category preloaded
    dimension, err := s.dimensionRepo.GetByID(ctx, dimension.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to reload dimension: %w", err)
    }

    // 8. Update parent totals
    if err := s.updateParentTotals(ctx, req.ParentType, req.ParentID); err != nil {
        s.logger.Warn("failed to update parent totals", zap.Error(err))
    }

    // 9. Log activity
    s.logActivity(ctx, domain.ActivityTargetOffer, req.ParentID, "Budget dimension added",
        fmt.Sprintf("Added budget line: %s (Cost: %.2f, Revenue: %.2f)", dimension.GetName(), dimension.Cost, dimension.Revenue))

    dto := mapper.ToBudgetDimensionDTO(dimension)
    return &dto, nil
}
```

#### 2.2 GetByID

```go
func (s *BudgetDimensionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.BudgetDimensionDTO, error) {
    dimension, err := s.dimensionRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrBudgetDimensionNotFound
        }
        return nil, fmt.Errorf("failed to get budget dimension: %w", err)
    }

    dto := mapper.ToBudgetDimensionDTO(dimension)
    return &dto, nil
}
```

#### 2.3 Update

```go
func (s *BudgetDimensionService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateBudgetDimensionRequest) (*domain.BudgetDimensionDTO, error) {
    // 1. Get existing dimension
    dimension, err := s.dimensionRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrBudgetDimensionNotFound
        }
        return nil, fmt.Errorf("failed to get budget dimension: %w", err)
    }

    // 2. Validate name source
    if req.CategoryID == nil && req.CustomName == "" {
        return nil, ErrMissingName
    }

    // 3. Validate category if changed
    if req.CategoryID != nil {
        if err := s.validateCategory(ctx, *req.CategoryID); err != nil {
            return nil, err
        }
    }

    // 4. Validate financials
    if err := s.validateFinancials(req.Cost, req.Revenue, req.TargetMarginPercent, req.MarginOverride); err != nil {
        return nil, err
    }

    // 5. Update fields
    dimension.CategoryID = req.CategoryID
    dimension.CustomName = req.CustomName
    dimension.Cost = req.Cost
    dimension.Revenue = req.Revenue
    dimension.TargetMarginPercent = req.TargetMarginPercent
    dimension.MarginOverride = req.MarginOverride
    dimension.Description = req.Description
    dimension.Quantity = req.Quantity
    dimension.Unit = req.Unit
    dimension.DisplayOrder = req.DisplayOrder

    if err := s.dimensionRepo.Update(ctx, dimension); err != nil {
        return nil, fmt.Errorf("failed to update budget dimension: %w", err)
    }

    // 6. Reload to get computed fields
    dimension, err = s.dimensionRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to reload dimension: %w", err)
    }

    // 7. Update parent totals
    if err := s.updateParentTotals(ctx, dimension.ParentType, dimension.ParentID); err != nil {
        s.logger.Warn("failed to update parent totals", zap.Error(err))
    }

    // 8. Log activity
    s.logActivity(ctx, s.getActivityTargetType(dimension.ParentType), dimension.ParentID,
        "Budget dimension updated",
        fmt.Sprintf("Updated budget line: %s (Cost: %.2f, Revenue: %.2f)", dimension.GetName(), dimension.Cost, dimension.Revenue))

    dto := mapper.ToBudgetDimensionDTO(dimension)
    return &dto, nil
}
```

#### 2.4 Delete

```go
func (s *BudgetDimensionService) Delete(ctx context.Context, id uuid.UUID) error {
    // 1. Get dimension for activity logging
    dimension, err := s.dimensionRepo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrBudgetDimensionNotFound
        }
        return fmt.Errorf("failed to get budget dimension: %w", err)
    }

    parentType := dimension.ParentType
    parentID := dimension.ParentID
    name := dimension.GetName()

    // 2. Delete
    if err := s.dimensionRepo.Delete(ctx, id); err != nil {
        return fmt.Errorf("failed to delete budget dimension: %w", err)
    }

    // 3. Update parent totals
    if err := s.updateParentTotals(ctx, parentType, parentID); err != nil {
        s.logger.Warn("failed to update parent totals", zap.Error(err))
    }

    // 4. Log activity
    s.logActivity(ctx, s.getActivityTargetType(parentType), parentID,
        "Budget dimension deleted",
        fmt.Sprintf("Deleted budget line: %s", name))

    return nil
}
```

---

### 3. Listing & Summary Methods

#### 3.1 ListByParent

```go
func (s *BudgetDimensionService) ListByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) ([]domain.BudgetDimensionDTO, error) {
    dimensions, err := s.dimensionRepo.GetByParent(ctx, parentType, parentID)
    if err != nil {
        return nil, fmt.Errorf("failed to list budget dimensions: %w", err)
    }

    dtos := make([]domain.BudgetDimensionDTO, len(dimensions))
    for i, dim := range dimensions {
        dtos[i] = mapper.ToBudgetDimensionDTO(&dim)
    }

    return dtos, nil
}
```

#### 3.2 ListByParentPaginated

```go
func (s *BudgetDimensionService) ListByParentPaginated(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, page, pageSize int) (*domain.PaginatedResponse, error) {
    // Clamp page size
    if pageSize < 1 {
        pageSize = 20
    }
    if pageSize > 200 {
        pageSize = 200
    }
    if page < 1 {
        page = 1
    }

    dimensions, total, err := s.dimensionRepo.GetByParentPaginated(ctx, parentType, parentID, page, pageSize)
    if err != nil {
        return nil, fmt.Errorf("failed to list budget dimensions: %w", err)
    }

    dtos := make([]domain.BudgetDimensionDTO, len(dimensions))
    for i, dim := range dimensions {
        dtos[i] = mapper.ToBudgetDimensionDTO(&dim)
    }

    totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
    return &domain.PaginatedResponse{
        Data:       dtos,
        Total:      total,
        Page:       page,
        PageSize:   pageSize,
        TotalPages: totalPages,
    }, nil
}
```

#### 3.3 GetSummary (AC3: CalculateTotals)

```go
func (s *BudgetDimensionService) GetSummary(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) (*domain.BudgetSummaryDTO, error) {
    summary, err := s.dimensionRepo.GetBudgetSummary(ctx, parentType, parentID)
    if err != nil {
        return nil, fmt.Errorf("failed to get budget summary: %w", err)
    }

    dto := &domain.BudgetSummaryDTO{
        ParentType:           parentType,
        ParentID:             parentID,
        DimensionCount:       summary.DimensionCount,
        TotalCost:            summary.TotalCost,
        TotalRevenue:         summary.TotalRevenue,
        OverallMarginPercent: summary.MarginPercent,
        TotalProfit:          summary.TotalMargin,
    }

    return dto, nil
}
```

---

### 4. Reorder Dimensions (AC5)

```go
func (s *BudgetDimensionService) ReorderDimensions(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID, orderedIDs []uuid.UUID) error {
    // 1. Validate parent exists
    if err := s.validateParentExists(ctx, parentType, parentID); err != nil {
        return err
    }

    // 2. Validate count matches
    count, err := s.dimensionRepo.Count(ctx, parentType, parentID)
    if err != nil {
        return fmt.Errorf("failed to count dimensions: %w", err)
    }

    if len(orderedIDs) != count {
        return fmt.Errorf("ordered IDs count (%d) does not match existing dimensions count (%d)", len(orderedIDs), count)
    }

    // 3. Perform reorder
    if err := s.dimensionRepo.ReorderDimensions(ctx, parentType, parentID, orderedIDs); err != nil {
        return fmt.Errorf("failed to reorder dimensions: %w", err)
    }

    // 4. Log activity
    s.logActivity(ctx, s.getActivityTargetType(parentType), parentID,
        "Budget dimensions reordered",
        fmt.Sprintf("Reordered %d budget lines", len(orderedIDs)))

    return nil
}
```

---

### 5. Clone Dimensions (AC4)

```go
// CloneDimensions copies all budget dimensions from source (offer) to target (project)
// This is typically used when an offer is won and converted to a project
func (s *BudgetDimensionService) CloneDimensions(ctx context.Context, sourceType domain.BudgetParentType, sourceID uuid.UUID, targetType domain.BudgetParentType, targetID uuid.UUID) ([]domain.BudgetDimensionDTO, error) {
    // 1. Validate source exists
    if err := s.validateParentExists(ctx, sourceType, sourceID); err != nil {
        return nil, fmt.Errorf("source validation failed: %w", err)
    }

    // 2. Validate target exists
    if err := s.validateParentExists(ctx, targetType, targetID); err != nil {
        return nil, fmt.Errorf("target validation failed: %w", err)
    }

    // 3. Get source dimensions
    sourceDimensions, err := s.dimensionRepo.GetByParent(ctx, sourceType, sourceID)
    if err != nil {
        return nil, fmt.Errorf("failed to get source dimensions: %w", err)
    }

    if len(sourceDimensions) == 0 {
        return nil, ErrSourceDimensionsNotFound
    }

    // 4. Clone each dimension
    clonedDTOs := make([]domain.BudgetDimensionDTO, 0, len(sourceDimensions))
    for _, src := range sourceDimensions {
        cloned := &domain.BudgetDimension{
            ParentType:          targetType,
            ParentID:            targetID,
            CategoryID:          src.CategoryID,
            CustomName:          src.CustomName,
            Cost:                src.Cost,
            Revenue:             src.Revenue,
            TargetMarginPercent: src.TargetMarginPercent,
            MarginOverride:      src.MarginOverride,
            Description:         src.Description,
            Quantity:            src.Quantity,
            Unit:                src.Unit,
            DisplayOrder:        src.DisplayOrder,
        }

        if err := s.dimensionRepo.Create(ctx, cloned); err != nil {
            return nil, fmt.Errorf("failed to clone dimension: %w", err)
        }

        // Reload to get computed fields and category
        cloned, err = s.dimensionRepo.GetByID(ctx, cloned.ID)
        if err != nil {
            s.logger.Warn("failed to reload cloned dimension", zap.Error(err))
            continue
        }

        clonedDTOs = append(clonedDTOs, mapper.ToBudgetDimensionDTO(cloned))
    }

    // 5. Update target totals
    if err := s.updateParentTotals(ctx, targetType, targetID); err != nil {
        s.logger.Warn("failed to update target totals", zap.Error(err))
    }

    // 6. Log activity on target
    s.logActivity(ctx, s.getActivityTargetType(targetType), targetID,
        "Budget dimensions cloned",
        fmt.Sprintf("Cloned %d budget lines from %s", len(clonedDTOs), sourceType))

    return clonedDTOs, nil
}
```

---

### 6. Validation Methods (AC2)

```go
// validateFinancials validates cost, revenue, and margin settings
func (s *BudgetDimensionService) validateFinancials(cost, revenue float64, targetMargin *float64, marginOverride bool) error {
    // Cost validation - must be positive (allow 0 for some edge cases? Check with PM)
    // Per AC: cost > 0
    if cost <= 0 {
        return ErrInvalidCost
    }

    // When margin override is true, revenue is calculated from cost and target margin
    // So we don't validate revenue in that case
    if !marginOverride {
        // Revenue validation - must be non-negative
        if revenue < 0 {
            return ErrInvalidRevenue
        }
    }

    // Target margin validation when margin override is enabled
    if marginOverride {
        if targetMargin == nil {
            return fmt.Errorf("%w: target margin is required when margin override is enabled", ErrInvalidTargetMargin)
        }
        if *targetMargin < 0 || *targetMargin >= 100 {
            return ErrInvalidTargetMargin
        }
    }

    return nil
}

// validateCategory checks if category exists and is active
func (s *BudgetDimensionService) validateCategory(ctx context.Context, categoryID string) error {
    category, err := s.categoryRepo.GetByID(ctx, categoryID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return ErrInvalidCategory
        }
        return fmt.Errorf("failed to validate category: %w", err)
    }

    if !category.IsActive {
        return ErrInvalidCategory
    }

    return nil
}

// validateParentExists verifies the parent entity (offer or project) exists
func (s *BudgetDimensionService) validateParentExists(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
    switch parentType {
    case domain.BudgetParentOffer:
        _, err := s.offerRepo.GetByID(ctx, parentID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return fmt.Errorf("%w: offer %s not found", ErrParentNotFound, parentID)
            }
            return fmt.Errorf("failed to validate offer: %w", err)
        }
    case domain.BudgetParentProject:
        _, err := s.projectRepo.GetByID(ctx, parentID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return fmt.Errorf("%w: project %s not found", ErrParentNotFound, parentID)
            }
            return fmt.Errorf("failed to validate project: %w", err)
        }
    default:
        return ErrInvalidParentType
    }

    return nil
}
```

---

### 7. Helper Methods

```go
// updateParentTotals recalculates and updates the parent entity's budget totals
func (s *BudgetDimensionService) updateParentTotals(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
    summary, err := s.dimensionRepo.GetBudgetSummary(ctx, parentType, parentID)
    if err != nil {
        return err
    }

    switch parentType {
    case domain.BudgetParentOffer:
        // Update offer.Value to match total revenue
        offer, err := s.offerRepo.GetByID(ctx, parentID)
        if err != nil {
            return err
        }
        offer.Value = summary.TotalRevenue
        return s.offerRepo.Update(ctx, offer)

    case domain.BudgetParentProject:
        // Update project.Budget to match total cost (or revenue, depending on business rules)
        project, err := s.projectRepo.GetByID(ctx, parentID)
        if err != nil {
            return err
        }
        project.Budget = summary.TotalRevenue
        project.HasDetailedBudget = summary.DimensionCount > 0
        return s.projectRepo.Update(ctx, project)
    }

    return nil
}

// getActivityTargetType maps budget parent type to activity target type
func (s *BudgetDimensionService) getActivityTargetType(parentType domain.BudgetParentType) domain.ActivityTargetType {
    switch parentType {
    case domain.BudgetParentOffer:
        return domain.ActivityTargetOffer
    case domain.BudgetParentProject:
        return domain.ActivityTargetProject
    default:
        return domain.ActivityTargetOffer
    }
}

// logActivity creates an activity log entry
func (s *BudgetDimensionService) logActivity(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, title, body string) {
    userCtx, ok := auth.FromContext(ctx)
    if !ok {
        s.logger.Warn("no user context for activity logging")
        return
    }

    activity := &domain.Activity{
        TargetType:  targetType,
        TargetID:    targetID,
        Title:       title,
        Body:        body,
        CreatorName: userCtx.DisplayName,
    }

    if err := s.activityRepo.Create(ctx, activity); err != nil {
        s.logger.Warn("failed to log activity", zap.Error(err))
    }
}
```

---

### 8. Delete All By Parent

```go
// DeleteByParent removes all budget dimensions for a parent entity
func (s *BudgetDimensionService) DeleteByParent(ctx context.Context, parentType domain.BudgetParentType, parentID uuid.UUID) error {
    // 1. Get count for activity log
    count, err := s.dimensionRepo.Count(ctx, parentType, parentID)
    if err != nil {
        return fmt.Errorf("failed to count dimensions: %w", err)
    }

    if count == 0 {
        return nil // Nothing to delete
    }

    // 2. Delete all
    if err := s.dimensionRepo.DeleteByParent(ctx, parentType, parentID); err != nil {
        return fmt.Errorf("failed to delete dimensions: %w", err)
    }

    // 3. Update parent totals (will set to 0)
    if err := s.updateParentTotals(ctx, parentType, parentID); err != nil {
        s.logger.Warn("failed to update parent totals", zap.Error(err))
    }

    // 4. Log activity
    s.logActivity(ctx, s.getActivityTargetType(parentType), parentID,
        "Budget dimensions cleared",
        fmt.Sprintf("Removed all %d budget lines", count))

    return nil
}
```

---

## Test Coverage Requirements (AC7)

Create `tests/service/budget_dimension_service_test.go`:

### Test Categories

#### 1. Create Tests
- `TestCreate_Success` - Valid creation with custom name
- `TestCreate_WithCategory` - Valid creation with category reference
- `TestCreate_WithMarginOverride` - Creates with margin override, verifies revenue calculation
- `TestCreate_InvalidParent` - Fails with non-existent parent
- `TestCreate_InvalidCategory` - Fails with non-existent category
- `TestCreate_MissingName` - Fails when no categoryId or customName provided
- `TestCreate_InvalidCost` - Fails when cost <= 0
- `TestCreate_InvalidMargin` - Fails when margin not in 0-100 range
- `TestCreate_AutoDisplayOrder` - Verifies auto-increment of display order

#### 2. GetByID Tests
- `TestGetByID_Found` - Returns dimension with category preloaded
- `TestGetByID_NotFound` - Returns ErrBudgetDimensionNotFound

#### 3. Update Tests
- `TestUpdate_Success` - Updates all fields
- `TestUpdate_NotFound` - Returns error for non-existent dimension
- `TestUpdate_InvalidCost` - Validation error
- `TestUpdate_InvalidMargin` - Validation error
- `TestUpdate_UpdatesParentTotals` - Verifies parent entity is updated

#### 4. Delete Tests
- `TestDelete_Success` - Deletes and updates parent totals
- `TestDelete_NotFound` - Returns error for non-existent dimension
- `TestDelete_LogsActivity` - Verifies activity is created

#### 5. List Tests
- `TestListByParent_Success` - Returns ordered dimensions
- `TestListByParent_Empty` - Returns empty slice for no dimensions
- `TestListByParentPaginated_FirstPage` - Pagination works correctly
- `TestListByParentPaginated_Clamping` - Page size is clamped to 200

#### 6. Summary Tests (CalculateTotals)
- `TestGetSummary_WithDimensions` - Correct totals calculated
- `TestGetSummary_Empty` - Zero values for no dimensions

#### 7. Reorder Tests
- `TestReorderDimensions_Success` - Order is updated correctly
- `TestReorderDimensions_InvalidCount` - Fails if ID count doesn't match
- `TestReorderDimensions_InvalidID` - Fails with non-existent ID

#### 8. Clone Tests
- `TestCloneDimensions_OfferToProject` - Successfully clones all dimensions
- `TestCloneDimensions_PreservesOrder` - Display order is maintained
- `TestCloneDimensions_PreservesFinancials` - Cost, revenue, margin preserved
- `TestCloneDimensions_SourceEmpty` - Returns error when source has no dimensions
- `TestCloneDimensions_InvalidSource` - Returns error for non-existent source
- `TestCloneDimensions_InvalidTarget` - Returns error for non-existent target

#### 9. Validation Tests
- `TestValidateMarginOverride_Valid` - Margin in range 0-100
- `TestValidateMarginOverride_TooHigh` - Margin >= 100 rejected
- `TestValidateMarginOverride_Negative` - Negative margin rejected
- `TestValidateMarginOverride_MissingWhenRequired` - No target margin with override=true

---

## Wire-up in main.go

Add to `cmd/api/main.go`:

```go
// In repository initialization section:
budgetDimensionRepo := repository.NewBudgetDimensionRepository(db)
budgetDimensionCategoryRepo := repository.NewBudgetDimensionCategoryRepository(db)

// In service initialization section:
budgetDimensionService := service.NewBudgetDimensionService(
    budgetDimensionRepo,
    budgetDimensionCategoryRepo,
    offerRepo,
    projectRepo,
    activityRepo,
    logger,
)
```

---

## File Checklist

- [ ] `internal/service/budget_dimension_service.go`
- [ ] `tests/service/budget_dimension_service_test.go`
- [ ] Update `cmd/api/main.go` (wire-up)

---

## Definition of Done

1. All 7 acceptance criteria explicitly verified
2. Service follows established patterns (see CustomerService, DealService)
3. Proper error wrapping with context
4. Activity logging for all mutations
5. Parent entity totals updated after changes
6. Tests pass with good coverage (target: 80%+)
7. Code follows Clean Architecture (Service depends on repositories only)
8. Code reviewed and approved
9. Branch merged to main

---

## Architecture Compliance Checklist

Based on CLAUDE.md patterns:

- [ ] Handler -> Service -> Repository flow respected
- [ ] DTOs used for API responses (via mapper)
- [ ] Domain models for internal operations
- [ ] Activity logging for audit trail
- [ ] Denormalized field updates (offer.Value, project.Budget)
- [ ] Proper error wrapping: `fmt.Errorf("context: %w", err)`
- [ ] uuid.UUID types (not strings) for IDs
- [ ] Context passed through all layers
- [ ] Pagination respects max 200 items
- [ ] Zap logger for structured logging
