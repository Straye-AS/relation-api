# Feature Specification: Flexible Offer Creation with Project Association

**Story ID**: SC-XXX (to be assigned)
**Type**: Enhancement
**Priority**: High
**Author**: Scrum Master
**Date**: 2025-12-13
**Status**: Ready for Review

---

## 1. Summary

Modify the `POST /offers` endpoint to:
1. **Default phase to `in_progress`** instead of `draft`
2. **Allow flexible customer/project association** with three scenarios:
   - Provide `customerId` only (existing behavior)
   - Provide `projectId` only (NEW - inherit customer from project)
   - Provide both `customerId` AND `projectId` (use provided customer, connect to project)

---

## 2. Business Context

### 2.1 Current State

Currently, `CreateOfferRequest` requires `customerId` and defaults to `draft` phase:
```go
type CreateOfferRequest struct {
    Title             string     `json:"title" validate:"required,max=200"`
    CustomerID        uuid.UUID  `json:"customerId" validate:"required"`  // <-- Required
    CompanyID         CompanyID  `json:"companyId,omitempty"`
    Phase             OfferPhase `json:"phase,omitempty"`  // Defaults to "draft"
    ProjectID         *uuid.UUID `json:"projectId,omitempty"`
    // ... other fields
}
```

### 2.2 Why This Change

1. **Default Phase**: Users almost never want to create offers in draft phase - they want to start working on them immediately (`in_progress`). The draft phase was intended for "inquiries" but that workflow has evolved.

2. **Project-First Workflow**: Sometimes users create a project first (as an "offer folder") and then want to add competing offers from different customers to that project. Currently, they must always specify a customer even when the project already has context.

3. **Multi-Customer Project Support**: A project in `tilbud` (offer) phase can have multiple offers from different customers (e.g., competing bids to different general contractors for the same physical project/location). The project's customer is only finalized when an offer is won.

---

## 3. Detailed Requirements

### 3.1 Default Phase Change

| Aspect | Current | Proposed |
|--------|---------|----------|
| Default phase when not specified | `draft` | `in_progress` |
| Offer number generation | Only for non-draft | Unchanged (generates for `in_progress`) |
| Project auto-creation | Only for non-draft | Unchanged (creates for `in_progress`) |

**Impact**: Offers created without explicit `phase` will:
- Get an offer number immediately
- Have/create a linked project automatically
- Require `companyId` or inherit from customer

### 3.2 Customer/Project Association Scenarios

#### Scenario A: `customerId` only (existing behavior, mostly unchanged)
```json
{
  "title": "Steel Structure Proposal",
  "customerId": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Behavior**:
1. Validate customer exists
2. If `phase` is non-draft (default `in_progress`): auto-create project with same customer
3. Offer's customer = provided customer
4. Project's customer = offer's customer

#### Scenario B: `projectId` only (NEW)
```json
{
  "title": "Alternative Bid",
  "projectId": "987fcdeb-51a2-43e8-9f12-abcdef123456"
}
```

**Behavior**:
1. Validate project exists
2. **Inherit customer from project** (project's current CustomerID)
3. If project has no customer (CustomerID is NULL), return error
4. Offer's customer = project's customer
5. Link offer to project

**Error Scenarios**:
- Project not found: `404 Not Found`
- Project has no customer: `400 Bad Request` - "Project has no customer assigned. Provide customerId explicitly."
- Project is cancelled: `400 Bad Request` - "Cannot add offer to cancelled project"

#### Scenario C: Both `customerId` AND `projectId` (enhanced)
```json
{
  "title": "Subcontractor Bid",
  "customerId": "123e4567-e89b-12d3-a456-426614174000",
  "projectId": "987fcdeb-51a2-43e8-9f12-abcdef123456"
}
```

**Behavior**:
1. Validate customer exists
2. Validate project exists
3. Validate project is not cancelled
4. Validate project's `companyId` matches offer's `companyId`
5. Offer's customer = provided `customerId` (NOT project's customer)
6. Link offer to project
7. Project's customer recalculated based on all active offers (existing `recalculateProjectCustomer` logic)

**Key Point**: The offer can have a DIFFERENT customer than the project. This is intentional for multi-bid scenarios.

### 3.3 Validation Rules

| Field | Scenario A | Scenario B | Scenario C |
|-------|-----------|-----------|-----------|
| `customerId` | Required | Not provided | Required |
| `projectId` | Optional | Required | Required |
| Customer must exist | Yes | N/A (inherited) | Yes |
| Project must exist | If provided | Yes | Yes |
| Project not cancelled | If provided | Yes | Yes |
| Company must match project | If provided | Auto-matched | Yes |

### 3.4 Project Customer Behavior During `tilbud` Phase

The existing `recalculateProjectCustomer` function handles this:
- If all active offers (`in_progress`, `sent`) are to the **same** customer: project's customer = that customer
- If active offers are to **different** customers: project's customer = NULL (cannot infer)
- If **no** active offers: project's customer = NULL

**No changes needed** to this logic - it already supports multi-customer projects.

### 3.5 When Offer is Won

**IMPORTANT GAP IDENTIFIED**: Currently, when an offer is won via `WinOffer()`, the project's customer is NOT updated from the winning offer. This needs to be fixed.

**Required Change to `WinOffer`**:
```go
// In the transaction, after SetWinningOffer:
// 6. Set project's customer from winning offer
if err := tx.Model(&domain.Project{}).
    Where("id = ?", project.ID).
    Updates(map[string]interface{}{
        "customer_id":   offer.CustomerID,
        "customer_name": offer.CustomerName,
    }).Error; err != nil {
    return fmt.Errorf("failed to update project customer: %w", err)
}
```

---

## 4. API Contract Changes

### 4.1 CreateOfferRequest DTO

```go
type CreateOfferRequest struct {
    Title             string                   `json:"title" validate:"required,max=200"`
    CustomerID        *uuid.UUID               `json:"customerId,omitempty"`  // Changed from required to optional
    CompanyID         CompanyID                `json:"companyId,omitempty"`
    Phase             OfferPhase               `json:"phase,omitempty"`       // Default: in_progress (was draft)
    Probability       *int                     `json:"probability,omitempty" validate:"omitempty,min=0,max=100"`
    Status            OfferStatus              `json:"status,omitempty"`
    ResponsibleUserID string                   `json:"responsibleUserId,omitempty"`
    ProjectID         *uuid.UUID               `json:"projectId,omitempty"`
    Items             []CreateOfferItemRequest `json:"items,omitempty"`
    Description       string                   `json:"description,omitempty"`
    Notes             string                   `json:"notes,omitempty"`
    DueDate           *time.Time               `json:"dueDate,omitempty"`
    Cost              float64                  `json:"cost,omitempty" validate:"gte=0"`
    Location          string                   `json:"location,omitempty" validate:"max=200"`
    SentDate          *time.Time               `json:"sentDate,omitempty"`
    ExpirationDate    *time.Time               `json:"expirationDate,omitempty"`
}
```

**Key Changes**:
1. `CustomerID` changes from `uuid.UUID` (required) to `*uuid.UUID` (optional pointer)
2. Remove `validate:"required"` tag from `CustomerID`
3. Default phase changes from `draft` to `in_progress` in service layer

### 4.2 New Error Responses

| HTTP Status | Error Message | Condition |
|-------------|---------------|-----------|
| 400 | "Either customerId or projectId must be provided" | Neither provided |
| 400 | "Project has no customer assigned. Provide customerId explicitly." | `projectId` only, but project has null CustomerID |
| 400 | "Cannot add offer to cancelled project" | Project is cancelled |
| 400 | "Project company does not match offer company" | Company mismatch |
| 404 | "Customer not found" | `customerId` provided but doesn't exist |
| 404 | "Project not found" | `projectId` provided but doesn't exist |

---

## 5. Service Layer Changes

### 5.1 Offer Service: Create/CreateWithProjectResponse

```go
func (s *OfferService) CreateWithProjectResponse(ctx context.Context, req *domain.CreateOfferRequest) (*domain.OfferWithProjectResponse, error) {
    var customer *domain.Customer
    var project *domain.Project

    // VALIDATION: At least one of customerId or projectId must be provided
    if req.CustomerID == nil && req.ProjectID == nil {
        return nil, ErrCustomerOrProjectRequired
    }

    // Handle Scenario B: projectId only - inherit customer
    if req.CustomerID == nil && req.ProjectID != nil {
        project, err = s.projectRepo.GetByID(ctx, *req.ProjectID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return nil, ErrProjectNotFound
            }
            return nil, fmt.Errorf("failed to get project: %w", err)
        }

        // Project must have a customer to inherit from
        if project.CustomerID == uuid.Nil {
            return nil, ErrProjectHasNoCustomer
        }

        // Inherit customer from project
        customer, err = s.customerRepo.GetByID(ctx, project.CustomerID)
        if err != nil {
            return nil, fmt.Errorf("failed to get project's customer: %w", err)
        }

        // Set the request's customerID for downstream processing
        req.CustomerID = &project.CustomerID
    }

    // Handle Scenario A & C: customerId provided
    if req.CustomerID != nil && customer == nil {
        customer, err = s.customerRepo.GetByID(ctx, *req.CustomerID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return nil, ErrCustomerNotFound
            }
            return nil, fmt.Errorf("failed to verify customer: %w", err)
        }
    }

    // ... rest of existing logic with updated default phase

    // Set default phase to in_progress (not draft)
    phase := req.Phase
    if phase == "" {
        phase = domain.OfferPhaseInProgress  // Changed from OfferPhaseDraft
    }

    // ... continue with existing flow
}
```

### 5.2 New Error Variables

```go
var (
    // ... existing errors ...
    ErrCustomerOrProjectRequired = errors.New("either customerId or projectId must be provided")
    ErrProjectHasNoCustomer      = errors.New("project has no customer assigned")
)
```

### 5.3 WinOffer Enhancement

In `WinOffer()`, add step to update project's customer:

```go
// Inside the transaction, after step 4 (SetWinningOffer):

// 6. Update project's customer from winning offer
if err := tx.Model(&domain.Project{}).
    Where("id = ?", project.ID).
    Updates(map[string]interface{}{
        "customer_id":   offer.CustomerID,
        "customer_name": offer.CustomerName,
    }).Error; err != nil {
    return fmt.Errorf("failed to update project customer from winning offer: %w", err)
}
```

---

## 6. Acceptance Criteria

### AC1: Default Phase is `in_progress`
- [ ] Creating an offer without specifying `phase` results in `phase: "in_progress"`
- [ ] Offer number is generated immediately (not waiting for phase transition)
- [ ] Project is auto-created (if not provided) when creating the offer

### AC2: Scenario A - customerId Only
- [ ] Works exactly as before when only `customerId` is provided
- [ ] Customer is validated to exist
- [ ] Project auto-created with same customer (if non-draft phase)

### AC3: Scenario B - projectId Only (NEW)
- [ ] Offer can be created with only `projectId` specified
- [ ] Customer is inherited from the project
- [ ] If project has NULL CustomerID, returns 400 error with clear message
- [ ] If project is cancelled, returns 400 error
- [ ] Offer is linked to the provided project

### AC4: Scenario C - Both customerId AND projectId
- [ ] Offer uses the provided `customerId` (not project's customer)
- [ ] Offer is linked to the provided `projectId`
- [ ] Project's customer is recalculated based on all active offers
- [ ] Company ID validation between project and offer still applies

### AC5: Validation Errors
- [ ] Neither `customerId` nor `projectId` provided: 400 error
- [ ] `customerId` not found: 404 error
- [ ] `projectId` not found: 404 error
- [ ] Project has no customer (Scenario B): 400 error with message
- [ ] Project is cancelled: 400 error

### AC6: WinOffer Updates Project Customer
- [ ] When offer is won, project's CustomerID and CustomerName are set from the winning offer
- [ ] This works correctly even if project had NULL customer or different customer

---

## 7. Test Scenarios

### Unit Tests

1. **CreateOffer_DefaultPhase_IsInProgress**
   - Create offer without phase
   - Assert phase is `in_progress`
   - Assert offer number is generated

2. **CreateOffer_ProjectIdOnly_InheritsCustomer**
   - Create project with Customer A
   - Create offer with only `projectId`
   - Assert offer's customer is Customer A

3. **CreateOffer_ProjectIdOnly_ProjectHasNoCustomer_ReturnsError**
   - Create project with NULL customer
   - Attempt to create offer with only `projectId`
   - Assert 400 error with appropriate message

4. **CreateOffer_BothIds_UseProvidedCustomer**
   - Create project with Customer A
   - Create offer with `projectId` AND `customerId` = Customer B
   - Assert offer's customer is Customer B (not A)

5. **CreateOffer_NeitherProvided_ReturnsError**
   - Attempt to create offer without `customerId` or `projectId`
   - Assert 400 error

6. **WinOffer_UpdatesProjectCustomer**
   - Create project with Customer A
   - Create offer to Customer B linked to project
   - Win the offer
   - Assert project's customer is now Customer B

### Integration Tests

1. **FullFlow_MultiCustomerProject**
   - Create project
   - Add Offer 1 to Customer A
   - Add Offer 2 to Customer B (both linked to same project)
   - Verify project customer recalculation shows NULL (multiple customers)
   - Win Offer 1
   - Verify project customer is now Customer A
   - Verify Offer 2 is expired

---

## 8. Migration Considerations

**No database migrations required** - this is purely a business logic change.

The DTO change (`CustomerID` from required to optional) is backwards compatible:
- Existing API clients providing `customerId` continue to work
- No breaking changes to stored data

---

## 9. Swagger/OpenAPI Updates

Update `CreateOfferRequest` documentation:
```yaml
CreateOfferRequest:
  type: object
  required:
    - title
    # Note: customerId is no longer required - either customerId OR projectId must be provided
  properties:
    title:
      type: string
      maxLength: 200
    customerId:
      type: string
      format: uuid
      description: |
        Customer ID for the offer. Optional if projectId is provided
        (customer will be inherited from project).
    projectId:
      type: string
      format: uuid
      description: |
        Project ID to link the offer to. If provided without customerId,
        the customer is inherited from the project.
    phase:
      type: string
      enum: [draft, in_progress, sent, won, lost, expired]
      default: in_progress
      description: "Phase of the offer. Defaults to 'in_progress' (previously 'draft')."
```

---

## 10. Out of Scope / Related Stories

### 10.1 Identified Gap: WinOffer Does Not Set Project Customer

**This is a bug/gap that should be fixed as part of this story.**

Currently, when `WinOffer()` is called:
- Project phase changes to `active`
- Project gets winning offer's value/cost
- Project does NOT get winning offer's customer

This means if a project had offers to multiple customers and one is won, the project's `CustomerID` remains in its previous state (could be NULL or a different customer).

**Recommendation**: Include the fix in this story's implementation.

### 10.2 Future Consideration: Project Creation Without Customer

A potential future enhancement would be to allow creating projects without any customer (true "offer folder" model). Currently, `CreateProjectRequest` has `CustomerID` as required. This is out of scope for this story.

---

## 11. Implementation Notes for Developer

1. **Order of Operations**: The validation logic should be:
   ```
   1. Check if at least one of customerId/projectId is provided
   2. If projectId only: load project, verify customer exists, inherit
   3. If customerId provided: load customer
   4. Continue with existing flow
   ```

2. **Reuse Existing Functions**: The `ensureProjectForOffer` function already handles project validation - extend it rather than duplicate logic.

3. **Backwards Compatibility**: Ensure all existing tests pass - only the default phase behavior changes for callers not specifying phase.

4. **Error Messages**: Use descriptive error messages that help API consumers understand what to fix.

---

## Appendix A: Current Code References

- **CreateOfferRequest DTO**: `internal/domain/dto.go:690`
- **OfferService.Create**: `internal/service/offer_service.go:61`
- **OfferService.CreateWithProjectResponse**: `internal/service/offer_service.go:70`
- **ensureProjectForOffer**: `internal/service/offer_service.go:1533`
- **recalculateProjectCustomer**: `internal/service/offer_service.go:1643`
- **WinOffer**: `internal/service/offer_service.go:790`
- **SetWinningOffer**: `internal/repository/project_repository.go:498`