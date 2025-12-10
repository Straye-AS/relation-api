# Story: Offer-Project Lifecycle Management

**Type:** Feature
**Team:** Straye Relation
**Priority:** High

## Summary

Implement automatic project creation and lifecycle synchronization between offers and projects. This feature ensures that all non-draft offers are linked to a project, and that project states automatically update based on offer outcomes.

---

## Background & Context

### Current State

**Offer Phases:** `draft` -> `in_progress` -> `sent` -> `won`/`lost`/`expired`

**Project Phases:** `tilbud` -> `active` -> `completed` -> `cancelled`

**Existing Behavior:**
- Offers have optional `ProjectID` field (nullable)
- Projects have optional `OfferID` field for single-offer linking
- `WinOffer()` already transitions project from `tilbud` -> `active`
- Projects have `CalculatedOfferValue`, `WinningOfferID`, `InheritedOfferNumber` fields

### Problem Statement

Currently, offers and projects exist independently with manual linking. This creates:
1. Friction in the workflow
2. Risk of orphaned offers without proper project context
3. Manual synchronization of economics between offers and projects
4. Inconsistent state management when offers are won/lost

### Proposed Solution

Enforce a tighter coupling where:
- Non-draft offers MUST belong to a project
- Projects automatically track their "best" offer economics during `tilbud` phase
- Project lifecycle states are driven by offer outcomes

---

## Acceptance Criteria

### 1. Auto-Project Creation for Non-Draft Offers

**When an offer transitions to `in_progress` state (via create or update from draft):**

#### Case A: ProjectID IS provided in request
- [ ] Link offer to the specified existing project
- [ ] **Validation:** Project must exist
- [ ] **Validation:** Project must be in `tilbud` phase
- [ ] **Validation:** Project must belong to same customer as offer
- [ ] **Validation:** Project must belong to same company as offer
- [ ] Return error 400 with descriptive message if any validation fails

#### Case B: ProjectID IS NOT provided
- [ ] Auto-create a new project with the following attributes:
  - **Name:** `[AUTO] {offer.Title}` (prompts user to rename later)
  - **Phase:** `tilbud`
  - **Status:** `planning`
  - **CustomerID:** inherited from offer
  - **CustomerName:** inherited from offer (denormalized)
  - **CompanyID:** inherited from offer
  - **ManagerID:** offer's `ResponsibleUserID`, or company default if not set
  - **StartDate:** current date
  - **Budget:** 0 (will be synced from offer economics)
- [ ] Link the offer to this newly created project
- [ ] Return the project in the API response alongside the offer
- [ ] Log activity on offer: "Offer linked to auto-created project '{projectName}'"
- [ ] Log activity on project: "Project created from offer '{offerTitle}'"

### 2. Enforce Project Requirement for Non-Draft Offers

- [ ] All offers in non-draft states (`in_progress`, `sent`, `won`, `lost`, `expired`) MUST have a valid `ProjectID`
- [ ] Draft offers CAN exist without a `ProjectID`
- [ ] Return HTTP 400 error if attempting to advance offer to non-draft without project link
- [ ] Error message: `"Non-draft offers must be linked to a project"`
- [ ] Unlinking an offer from a project (`UnlinkFromProject()`) is ONLY allowed if offer is in `draft` phase
- [ ] Return HTTP 400 error if attempting to unlink a non-draft offer
- [ ] Error message: `"Cannot unlink non-draft offer from project"`

### 3. Project Economics Sync (Tilbud Phase Only)

**While project is in `tilbud` phase:**

- [ ] Project's `CalculatedOfferValue` should reflect the HIGHEST value offer linked to it
- [ ] **"Best" offer definition:** Highest `Value` among offers that are NOT in terminal states (`won`, `lost`, `expired`)
- [ ] If no qualifying offers exist, `CalculatedOfferValue` = 0
- [ ] Project's `Budget` field should mirror `CalculatedOfferValue` during `tilbud` phase

**Sync should occur when:**
- [ ] Offer is linked to project (new or existing)
- [ ] Offer's `Value` field is updated
- [ ] Offer transitions to `lost` or `expired` state
- [ ] Offer is unlinked from project (draft only)

**Economics editability:**
- [ ] During `tilbud` phase: `Budget` and `Spent` are read-only (mirror offer values)
- [ ] After `active` phase: `Budget` and `Spent` become editable
- [ ] Existing `IsEditablePhase()` method already handles this

### 4. Offer Won -> Project Active

**When an offer is marked as `won`:**

- [ ] Linked project transitions from `tilbud` -> `active` phase
- [ ] Project's `WinningOfferID` is set to the won offer's ID
- [ ] Project's `Budget` is locked to the winning offer's value
- [ ] Project's `InheritedOfferNumber` is set from winning offer's number
- [ ] Project's `WonAt` timestamp is set to current time
- [ ] All OTHER offers in the same project transition to `expired` phase *(existing behavior)*
- [ ] Budget items from winning offer are cloned to project *(existing behavior)*
- [ ] Activity logged on project: "Project activated - offer '{offerTitle}' won"
- [ ] Activity logged on offer: "Offer won - project '{projectName}' activated"

**Note:** This behavior mostly exists in `WinOffer()` - verify and extend as needed.

### 5. Offer Lost -> Conditional Project Expiration

**When an offer is marked as `lost`:**

- [ ] Query: Does the project have OTHER active offers? (phase = `in_progress` OR `sent`)

#### Case A: Other active offers EXIST
- [ ] Project remains in `tilbud` phase
- [ ] Recalculate `CalculatedOfferValue` from remaining active offers
- [ ] Update project `Budget` to new calculated value
- [ ] Activity logged: "Offer lost - project continues with {N} remaining offers"

#### Case B: NO other active offers exist
- [ ] Project transitions to `cancelled` phase
- [ ] Project `Status` set to `cancelled`
- [ ] Activity logged: "Project cancelled - all offers lost"
- [ ] Project becomes read-only (future: reactivation feature)
- [ ] Return descriptive response indicating project was cancelled

### 6. API Response Changes

#### Request DTOs

**`CreateOfferRequest`** - add field:
```go
ProjectID *uuid.UUID `json:"projectId,omitempty"` // Optional: link to existing project
```

**`UpdateOfferRequest`** - add field:
```go
ProjectID *uuid.UUID `json:"projectId,omitempty"` // Optional: link to existing project (only for draft->non-draft transition)
```

**`AdvanceOfferRequest`** - add field:
```go
ProjectID *uuid.UUID `json:"projectId,omitempty"` // Optional: link when transitioning from draft
```

#### Response DTOs

**New `OfferWithProjectResponse`:**
```go
type OfferWithProjectResponse struct {
    Offer          *OfferDTO   `json:"offer"`
    Project        *ProjectDTO `json:"project,omitempty"`
    ProjectCreated bool        `json:"projectCreated"`
}
```

**Update affected endpoints to optionally return this response type when project is created.**

### 7. Edge Cases

- [ ] **Multiple offers won simultaneously:** First processed wins, others expire *(existing WinOffer behavior)*
- [ ] **Offer lost but others pending:** Project stays in `tilbud`, economics recalculated from remaining offers
- [ ] **All offers lost in sequence:** When last offer is lost, project transitions to `cancelled`
- [ ] **Adding offer to cancelled project:** Return error 400 - "Cannot add offer to cancelled project"
- [ ] **Changing offer customer after project link:** Return error 400 - "Cannot change customer of offer linked to project"
- [ ] **Changing offer company after project link:** Return error 400 - "Cannot change company of offer linked to project"
- [ ] **Draft offer without project transitioned to in_progress via bulk update:** Auto-create project for each

---

## Technical Implementation Notes

### Files to Modify

| File | Changes |
|------|---------|
| `internal/domain/dto.go` | Add `ProjectID` to request DTOs, new `OfferWithProjectResponse` |
| `internal/domain/models.go` | No changes expected (fields exist) |
| `internal/service/offer_service.go` | Auto-project creation logic, project requirement validation |
| `internal/service/project_service.go` | Economics sync method, phase transition methods |
| `internal/service/errors.go` | New error types |
| `internal/repository/offer_repository.go` | Query for active offers in project |
| `internal/repository/project_repository.go` | Update economics from offers |
| `internal/http/handler/offer_handler.go` | Handle new response type |

### New Service Methods

```go
// offer_service.go
func (s *OfferService) ensureProjectForOffer(ctx context.Context, offer *domain.Offer, projectID *uuid.UUID) (*domain.Project, bool, error)
func (s *OfferService) validateProjectLinkForOffer(ctx context.Context, offer *domain.Offer, projectID uuid.UUID) error
func (s *OfferService) canUnlinkFromProject(offer *domain.Offer) error

// project_service.go
func (s *ProjectService) RecalculateBestOfferEconomics(ctx context.Context, projectID uuid.UUID) error
func (s *ProjectService) HasActiveOffers(ctx context.Context, projectID uuid.UUID) (bool, int, error)
func (s *ProjectService) TransitionToCancelledFromLostOffers(ctx context.Context, projectID uuid.UUID, reason string) error
func (s *ProjectService) GetActiveOffersInProject(ctx context.Context, projectID uuid.UUID) ([]domain.Offer, error)
```

### New Repository Methods

```go
// offer_repository.go
func (r *OfferRepository) CountActiveByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
func (r *OfferRepository) GetActiveByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Offer, error)
func (r *OfferRepository) GetBestOfferValueByProject(ctx context.Context, projectID uuid.UUID) (float64, error)

// project_repository.go
func (r *ProjectRepository) UpdateEconomicsFromOffers(ctx context.Context, projectID uuid.UUID) error
```

### New Error Types

```go
var (
    ErrNonDraftOfferRequiresProject = errors.New("non-draft offers must be linked to a project")
    ErrCannotUnlinkNonDraftOffer    = errors.New("cannot unlink non-draft offer from project")
    ErrCannotAddOfferToCancelledProject = errors.New("cannot add offer to cancelled project")
    ErrProjectCustomerMismatch      = errors.New("project customer does not match offer customer")
    ErrProjectCompanyMismatch       = errors.New("project company does not match offer company")
    ErrProjectNotInTilbudPhase      = errors.New("project is not in tilbud phase")
)
```

### Database Considerations

- **No schema changes required** - existing fields support this feature
- **Consider adding index:** `CREATE INDEX idx_offers_project_phase ON offers(project_id, phase)` for active offer queries
- **Migration:** Existing offers without project links remain valid (draft behavior)

### Transaction Boundaries

The following operations should be wrapped in transactions:
1. Auto-project creation + offer link
2. Offer won + project activation + sibling expiration
3. Offer lost + project cancellation (when last offer)

---

## Out of Scope (Future Stories)

| Feature | Reason |
|---------|--------|
| Project reactivation after cancellation | Separate feature request needed |
| Bulk offer-to-project migration for existing data | Data migration story |
| Custom "best offer" calculation rules | Beyond current requirements |
| UI changes for auto-project naming prompt | Frontend story |
| Notification when project auto-cancelled | Notification system enhancement |

---

## Questions Resolved

| Question | Answer |
|----------|--------|
| Multiple offers won simultaneously? | First one processed wins, others expire (existing WinOffer behavior) |
| Offer lost with others pending? | Project stays in `tilbud`, economics recalculated |
| What fields does auto-project inherit? | Customer, Company from offer; Manager from ResponsibleUserID or company default |
| What is "best" offer definition? | Highest `Value` among non-terminal (not won/lost/expired) offers |
| Can draft offers link to projects? | Yes, but not required. Linking is optional for drafts. |

---

## Test Scenarios

### Unit Tests

1. **Auto-project creation triggers correctly**
   - Create offer with phase=in_progress, no projectId -> project created
   - Create offer with phase=draft, no projectId -> no project created
   - Update offer from draft to in_progress, no projectId -> project created

2. **Project link validation**
   - Link to project with wrong customer -> error
   - Link to project with wrong company -> error
   - Link to cancelled project -> error
   - Link to project in active phase -> error

3. **Economics sync**
   - Add offer to project -> CalculatedOfferValue updated
   - Update offer value -> CalculatedOfferValue updated
   - Offer lost -> CalculatedOfferValue recalculated from remaining

4. **Project transitions**
   - Win offer -> project active
   - Lose last offer -> project cancelled
   - Lose offer with siblings -> project stays tilbud

### Integration Tests

1. Full offer lifecycle: draft -> in_progress (auto-project) -> sent -> won -> project active
2. Full offer lifecycle: draft -> in_progress (auto-project) -> sent -> lost -> project cancelled
3. Multiple offers in project: one won, others expired
4. Multiple offers in project: one lost, others continue

---

## Definition of Done

- [ ] All acceptance criteria implemented and tested
- [ ] Unit tests cover all new service methods
- [ ] Integration tests cover full lifecycle scenarios
- [ ] Swagger documentation updated for new/modified endpoints
- [ ] Activity logging implemented for all state transitions
- [ ] Error messages are clear and actionable
- [ ] No regression in existing offer/project functionality
- [ ] Code reviewed and approved
- [ ] Manual testing completed in development environment
