# Story: Project Reopening and State Management

**Type:** Feature
**Team:** Straye Relation
**Priority:** High

## Summary

Implement project reopening capabilities and add a new "working" phase to the project lifecycle. This feature enables users to reopen completed or cancelled projects, properly synchronizes offer states during reopening, and introduces a "working" phase to distinguish between won-but-not-started projects and projects with active work.

---

## Background & Context

### Current State

**Project Phases:** `tilbud` -> `active` -> `completed` / `cancelled`

**Project Statuses:** `planning`, `active`, `on_hold`, `completed`, `cancelled`

**Offer Phases:** `draft` -> `in_progress` -> `sent` -> `won` / `lost` / `expired`

**Current Limitations:**
1. Projects in `completed` or `cancelled` phase cannot be reopened
2. No distinction between "won but not started" and "actively being worked on"
3. When reopening would be needed, users must create new projects manually
4. Offer-project state consistency is not maintained during hypothetical reopening scenarios

### Problem Statement

Users need the ability to:
1. **Reopen closed projects** - A completed project may need additional work, or a cancelled project may be revived
2. **Track actual work start** - Currently `active` phase means "won" but doesn't indicate whether work has actually begun
3. **Maintain state consistency** - When a project is reopened, linked offers should reflect appropriate states

### Proposed Solution

1. Add `working` phase between `active` and `completed`
2. Implement `ReopenProject` functionality with offer state synchronization
3. Enforce business rules around phase transitions

---

## Acceptance Criteria

### 1. New Project Phase: "working" (I arbeid)

**Phase Definition:**
- [ ] Add new phase constant: `ProjectPhaseWorking = "working"`
- [ ] Phase position in lifecycle: `tilbud` -> `active` -> `working` -> `completed` / `cancelled`
- [ ] Display label (Norwegian): "I arbeid"

**Phase Semantics:**
| Phase | Meaning | Work Started? | Finished? |
|-------|---------|---------------|-----------|
| `tilbud` | Bidding/planning phase | No | No |
| `active` | Project won, not yet started | No | No |
| `working` | Work has begun | **Yes** | No |
| `completed` | Work finished | Yes | **Yes** |
| `cancelled` | Project cancelled | N/A | **Yes** |

**Start Date Requirement:**
- [ ] Transitioning to `working` phase MUST require a non-null `StartDate`
- [ ] If `StartDate` is null, return error: `"Cannot transition to working phase without a start date"`
- [ ] Auto-set `StartDate` to current date if not already set when transitioning to `working`

**Dashboard/Reporting Behavior:**
- [ ] `working` phase counts as "active" for metrics (not finished)
- [ ] `working` phase projects appear in "Active Projects" lists
- [ ] Filter by `isActive` should include both `active` AND `working` phases
- [ ] Order reserve calculations should include `working` phase projects

### 2. Project Reopening from Closed States

**Reopening Completed Projects:**
- [ ] Allow transition: `completed` -> `working`
- [ ] Preserve all project data (budget, economics, team, etc.)
- [ ] Log activity: "Project '{name}' was reopened from completed state"

**Reopening Cancelled Projects:**
- [ ] Allow transition: `cancelled` -> `tilbud`
- [ ] Cancelled projects reopen to `tilbud` (must go through offer/winning process again)
- [ ] Clear `WinningOfferID`, `WonAt` if set (project is no longer "won")
- [ ] Preserve economic data but make it editable again
- [ ] Log activity: "Project '{name}' was reopened from cancelled state"

### 3. Automatic Offer State Synchronization on Reopen

**When reopening a project with a "won" offer:**

| Project Transition | Offer Action | Reason |
|--------------------|--------------|--------|
| `completed` -> `working` | `won` -> `sent` | Project needs more work, offer should be re-sendable |
| `cancelled` -> `tilbud` | `won` -> `sent` | Project cancelled, offer returned to open state |

**Business Rules:**
- [ ] ONLY offers in `won` phase are affected
- [ ] `lost` offers remain `lost` - users must manually change these if needed
- [ ] `expired` offers remain `expired` - users must manually change these if needed
- [ ] Log activity on offer: "Offer reverted to 'sent' due to project reopening"
- [ ] Log activity on project: "Linked offer '{offerTitle}' reverted to 'sent'"

### 4. State Transition Rules - Project Phase Matrix

**Valid Phase Transitions:**

| From | To | Condition | Offer Impact |
|------|----|-----------|--------------|
| `tilbud` | `active` | Offer won | None (handled by WinOffer) |
| `tilbud` | `cancelled` | All offers lost/expired | None |
| `active` | `working` | Start date required | None |
| `active` | `completed` | Skip working allowed | None |
| `active` | `cancelled` | User decision | Won offer -> `expired` |
| `working` | `completed` | Normal completion | None |
| `working` | `cancelled` | User decision | Won offer -> `expired` |
| `completed` | `working` | Reopen | Won offer -> `sent` |
| `cancelled` | `tilbud` | Reopen | Won offer -> `sent`, clear winning offer fields |

**Invalid Transitions (return 400 error):**

| From | To | Error Message |
|------|----|---------------|
| `tilbud` | `working` | "Cannot transition directly to working from tilbud - project must be won first" |
| `tilbud` | `completed` | "Cannot complete project that hasn't been won" |
| `active` | `tilbud` | "Cannot revert active project to tilbud - use cancel and reopen instead" |
| `working` | `tilbud` | "Cannot revert working project to tilbud - use cancel and reopen instead" |
| `working` | `active` | "Cannot revert from working to active - work has begun" |
| `completed` | `active` | "Cannot reopen to active - use working phase for resumed projects" |
| `completed` | `tilbud` | "Cannot reopen completed project to tilbud - use working phase" |

### 5. Constraint: Project with Won Offer Cannot Be in Tilbud

**Business Rule:**
- [ ] A project CANNOT be in `tilbud` phase if it has a `won` offer linked to it
- [ ] This prevents data inconsistency where project shows "bidding" but offer shows "won"

**Implementation:**
- [ ] Check on project phase update: if new phase is `tilbud` and `WinningOfferID` is set, return error
- [ ] Error message: "Project cannot be in tilbud phase - it has a won offer"
- [ ] The `ReopenProject` to `tilbud` endpoint should first revert the offer state BEFORE changing project phase

### 6. API Endpoints

**New Endpoint: Reopen Project**
```
POST /api/v1/projects/{id}/reopen
```

**Request Body:**
```json
{
  "targetPhase": "working" | "tilbud",  // Required
  "startDate": "2025-01-15",             // Optional, required if targetPhase=working and not set
  "notes": "Customer requested additional scope"  // Optional
}
```

**Response (Success - 200):**
```json
{
  "project": { ... ProjectDTO ... },
  "previousPhase": "completed",
  "affectedOffers": [
    {
      "offerId": "uuid",
      "offerTitle": "Offer ABC",
      "previousPhase": "won",
      "newPhase": "sent"
    }
  ],
  "activityLogged": true
}
```

**Validation Errors (400):**
- `"Cannot reopen project - it is not in a closed state (completed or cancelled)"`
- `"Cannot reopen completed project to tilbud - use working phase"`
- `"Cannot reopen cancelled project to working - must go through tilbud first"`
- `"Cannot transition to working phase without a start date"`

**Updated Endpoint: Update Project Phase**
```
PATCH /api/v1/projects/{id}/phase
```

**Request Body:**
```json
{
  "phase": "working",
  "startDate": "2025-01-15"  // Required if transitioning to working
}
```

**New Validation:**
- [ ] Validate against phase transition matrix
- [ ] Require `startDate` for `working` transition
- [ ] Prevent `tilbud` if project has winning offer

### 7. Edge Cases

**Scenario: Reopen completed project that has NO linked offer**
- [ ] Allow reopening to `working`
- [ ] No offer state changes needed
- [ ] Log activity normally

**Scenario: Reopen project with multiple offers (one won, others expired)**
- [ ] Only the `won` offer reverts to `sent`
- [ ] `expired` offers remain `expired`
- [ ] User can manually reactivate expired offers if needed

**Scenario: Transition active -> cancelled with won offer**
- [ ] Won offer should transition to `expired` (not `sent`)
- [ ] This is cancellation, not reopening - different semantics
- [ ] Clear `WinningOfferID` on project

**Scenario: User tries to set phase to tilbud via regular update**
- [ ] Check if `WinningOfferID` is set
- [ ] If set, return error directing user to use reopen endpoint
- [ ] Reopen endpoint handles the offer state change atomically

---

## Technical Implementation Notes

### Files to Modify

| File | Changes |
|------|---------|
| `internal/domain/models.go` | Add `ProjectPhaseWorking` constant |
| `internal/domain/dto.go` | Add `ReopenProjectRequest`, `ReopenProjectResponse` DTOs |
| `internal/service/project_service.go` | Add `ReopenProject`, update phase transition validation |
| `internal/service/offer_service.go` | Add `RevertOfferToSent` method for reopening scenarios |
| `internal/service/errors.go` | New error types for invalid transitions |
| `internal/repository/project_repository.go` | Add `ReopenProject` method |
| `internal/http/handler/project_handler.go` | Add `/projects/{id}/reopen` endpoint |
| `internal/http/routes.go` | Register new endpoint |

### New Types

```go
// domain/models.go
const (
    ProjectPhaseTilbud    ProjectPhase = "tilbud"
    ProjectPhaseActive    ProjectPhase = "active"
    ProjectPhaseWorking   ProjectPhase = "working"  // NEW
    ProjectPhaseCompleted ProjectPhase = "completed"
    ProjectPhaseCancelled ProjectPhase = "cancelled"
)

// IsActive returns true if the project is considered active (not finished)
func (p ProjectPhase) IsActive() bool {
    return p == ProjectPhaseActive || p == ProjectPhaseWorking
}

// IsFinished returns true if the project is in a terminal state
func (p ProjectPhase) IsFinished() bool {
    return p == ProjectPhaseCompleted || p == ProjectPhaseCancelled
}

// CanTransitionTo checks if transition to target phase is valid
func (p ProjectPhase) CanTransitionTo(target ProjectPhase) bool {
    // Implementation based on matrix above
}
```

```go
// domain/dto.go
type ReopenProjectRequest struct {
    TargetPhase ProjectPhase `json:"targetPhase" validate:"required,oneof=working tilbud"`
    StartDate   *time.Time   `json:"startDate,omitempty"`
    Notes       string       `json:"notes,omitempty" validate:"max=500"`
}

type ReopenProjectResponse struct {
    Project        *ProjectDTO            `json:"project"`
    PreviousPhase  ProjectPhase           `json:"previousPhase"`
    AffectedOffers []AffectedOfferDTO     `json:"affectedOffers"`
    ActivityLogged bool                   `json:"activityLogged"`
}

type AffectedOfferDTO struct {
    OfferID       uuid.UUID  `json:"offerId"`
    OfferTitle    string     `json:"offerTitle"`
    PreviousPhase OfferPhase `json:"previousPhase"`
    NewPhase      OfferPhase `json:"newPhase"`
}
```

### New Service Methods

```go
// project_service.go

// ReopenProject reopens a closed project (completed or cancelled)
// and synchronizes linked offer states appropriately
func (s *ProjectService) ReopenProject(ctx context.Context, id uuid.UUID, req *domain.ReopenProjectRequest) (*domain.ReopenProjectResponse, error)

// validatePhaseTransition validates that a phase transition is allowed
func (s *ProjectService) validatePhaseTransition(from, to domain.ProjectPhase, project *domain.Project) error

// isValidPhaseTransition checks if transition is valid per the matrix
func (s *ProjectService) isValidPhaseTransition(from, to domain.ProjectPhase) bool

// offer_service.go

// RevertToSent reverts a won offer back to sent state (for project reopening)
func (s *OfferService) RevertToSent(ctx context.Context, id uuid.UUID, reason string) (*domain.OfferDTO, error)
```

### New Error Types

```go
// service/errors.go

var (
    // Phase transition errors
    ErrInvalidPhaseTransition          = errors.New("invalid phase transition")
    ErrCannotReopenActiveProject       = errors.New("cannot reopen project that is not closed")
    ErrCompletedCannotReopenToTilbud   = errors.New("completed project cannot reopen to tilbud - use working phase")
    ErrCancelledCannotReopenToWorking  = errors.New("cancelled project cannot reopen to working - must go through tilbud")
    ErrWorkingRequiresStartDate        = errors.New("working phase requires a start date")
    ErrTilbudWithWonOffer              = errors.New("project cannot be in tilbud phase with a won offer")
    ErrCannotRevertToTilbudFromActive  = errors.New("cannot revert active project to tilbud")
    ErrCannotRevertToActiveFromWorking = errors.New("cannot revert from working to active")
)
```

### Database Migration

```sql
-- Migration: Add 'working' to project_phase enum

-- +goose Up
ALTER TYPE project_phase ADD VALUE IF NOT EXISTS 'working' AFTER 'active';

-- Add comment for clarity
COMMENT ON TYPE project_phase IS 'Project lifecycle phases: tilbud (bidding), active (won but not started), working (in progress), completed, cancelled';

-- +goose Down
-- Note: PostgreSQL does not support removing enum values directly
-- The 'working' value will remain but can be unused
```

### Transaction Boundaries

Operations requiring transactions:
1. `ReopenProject` - project phase change + offer state reversion
2. `TransitionToWorking` - phase change + start date update
3. Cancellation with offer expiration

---

## Dashboard/Reporting Impact

### Metrics that include "working" phase:

| Metric | Includes Working? | Notes |
|--------|-------------------|-------|
| Active Project Count | YES | working = active work |
| Order Reserve | YES | Budget - Spent for working projects |
| Total Invoiced | YES | Spent from working projects |
| Projects "In Progress" | YES | Both active and working |

### Filter Behavior:

```go
// When filtering for "active" projects, include working
func GetActiveProjects() {
    phases := []ProjectPhase{ProjectPhaseActive, ProjectPhaseWorking}
    // ...
}
```

---

## Test Scenarios

### Unit Tests

**Phase Transition Validation:**
1. `tilbud` -> `active` (via WinOffer) - allowed
2. `tilbud` -> `working` - error (must be won first)
3. `tilbud` -> `completed` - error (must be won first)
4. `active` -> `working` with start date - allowed
5. `active` -> `working` without start date - error
6. `active` -> `completed` - allowed (skip working)
7. `working` -> `completed` - allowed
8. `working` -> `active` - error (cannot go backwards)
9. `completed` -> `working` - allowed (reopen)
10. `completed` -> `tilbud` - error (must use working)
11. `cancelled` -> `tilbud` - allowed (reopen)
12. `cancelled` -> `working` - error (must go through tilbud)

**Reopen Project:**
1. Reopen completed project with won offer -> offer reverts to sent
2. Reopen completed project without offer -> no offer changes
3. Reopen cancelled project with won offer -> offer reverts to sent, winning offer fields cleared
4. Reopen cancelled project with lost offer -> lost offer stays lost

**Constraint: Tilbud with Won Offer:**
1. Try to set phase=tilbud when WinningOfferID is set -> error
2. ReopenProject to tilbud clears WinningOfferID first -> success

### Integration Tests

1. Full lifecycle: tilbud -> active -> working -> completed -> working (reopen) -> completed
2. Full lifecycle: tilbud -> cancelled -> tilbud (reopen) -> active -> completed
3. Project with multiple offers: one won, others expired, reopen -> only won reverts

---

## Definition of Done

- [ ] `working` phase added to `ProjectPhase` enum
- [ ] Database migration for new enum value
- [ ] Phase transition validation implemented per matrix
- [ ] `ReopenProject` endpoint implemented
- [ ] Offer state synchronization on reopen implemented
- [ ] Start date requirement for working phase enforced
- [ ] Tilbud-with-won-offer constraint enforced
- [ ] Dashboard queries updated to include working phase
- [ ] Activity logging for all state transitions
- [ ] Unit tests for all transition scenarios
- [ ] Integration tests for full lifecycle
- [ ] Swagger documentation updated
- [ ] Code reviewed and approved

---

## Questions for Stakeholder Review

1. **Working phase auto-start date:** Should we auto-set StartDate to current date when transitioning to working, or require explicit user input?
   - **Recommendation:** Auto-set if null, allow override if provided

2. **Reopen permissions:** Should reopening require manager/admin role, or should any user with project write access be able to reopen?
   - **Recommendation:** Same permissions as project update (manager or admin)

3. **Notification on reopen:** Should we notify the project team when a project is reopened?
   - **Recommendation:** Out of scope for this story, create follow-up if needed

4. **Audit trail:** Should we add a `ReopenedAt` timestamp field to track reopen history?
   - **Recommendation:** Activity log is sufficient, but can add if needed
