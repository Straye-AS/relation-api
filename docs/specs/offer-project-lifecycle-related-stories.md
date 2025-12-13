# Related Stories: Offer-Project Lifecycle Management

This document captures scope gaps and related features identified during the analysis of the Offer-Project Lifecycle Management feature. Each section represents a potential separate story.

---

## Story Proposal: Project Reactivation After Cancellation

### Discovery Context
During analysis of the "Offer Lost -> Project Cancelled" flow, we identified that cancelled projects become permanently read-only. Users may need to reactivate projects if circumstances change.

### Description
Enable reactivation of cancelled projects to allow adding new offers and resuming the sales process.

### Acceptance Criteria
- [ ] Add `ReactivateProject()` method to ProjectService
- [ ] Transition cancelled project back to `tilbud` phase
- [ ] Clear `WinningOfferID`, `WonAt` fields on reactivation
- [ ] Require explicit user action (not automatic)
- [ ] Activity logged: "Project reactivated"
- [ ] Permission check: only project manager or admin can reactivate

### Technical Notes
- New endpoint: `POST /projects/{id}/reactivate`
- Consider adding "reactivation reason" field for audit

### Priority: Medium
### Dependencies: Offer-Project Lifecycle Management (parent story)

---

## Story Proposal: Bulk Migration for Existing Orphaned Offers

### Discovery Context
When implementing the project requirement enforcement, existing offers in non-draft states without project links will become invalid. These need migration.

### Description
Create a data migration to auto-create projects for existing non-draft offers that lack project links.

### Acceptance Criteria
- [ ] Identify all non-draft offers without ProjectID
- [ ] Create project for each with name `[MIGRATED] {offer.Title}`
- [ ] Link offer to newly created project
- [ ] Log migration activity on both entities
- [ ] Migration should be idempotent (safe to re-run)
- [ ] Generate report of migrated offers

### Technical Notes
- One-time migration script, not a recurring process
- Consider running in batches for large datasets
- Backup database before running

### Priority: High (blocks deployment of parent story)
### Dependencies: Offer-Project Lifecycle Management (parent story)

---

## Story Proposal: Custom "Best Offer" Calculation Rules

### Discovery Context
Current implementation uses "highest Value" as the best offer criterion. Some teams may want different rules (most recent, highest margin, etc.).

### Description
Allow configuration of how the "best" offer is determined for project economics calculation.

### Acceptance Criteria
- [ ] Support multiple calculation strategies: highest_value, highest_margin, most_recent
- [ ] Company-level configuration setting
- [ ] Default to highest_value for backward compatibility
- [ ] Apply strategy when syncing project economics
- [ ] Display current strategy in project details

### Technical Notes
- Strategy pattern for calculation
- Add `best_offer_strategy` column to companies table
- Consider future UI for strategy selection

### Priority: Low
### Dependencies: Offer-Project Lifecycle Management (parent story)

---

## Story Proposal: Project Cancellation Notifications

### Discovery Context
When a project is auto-cancelled due to all offers being lost, stakeholders should be notified.

### Description
Send notifications to relevant users when a project transitions to cancelled state due to offer outcomes.

### Acceptance Criteria
- [ ] Notify project manager when project is cancelled
- [ ] Notify all team members listed on project
- [ ] Notification includes: project name, reason, last offer details
- [ ] Support email and in-app notifications
- [ ] Allow users to configure notification preferences

### Technical Notes
- Integrate with existing notification system
- Consider notification batching for multiple rapid changes

### Priority: Medium
### Dependencies:
- Offer-Project Lifecycle Management (parent story)
- Notification system infrastructure

---

## Story Proposal: Expired Project Phase

### Discovery Context
Original requirements mentioned an "expired" state for projects, but current model uses `cancelled`. Need to clarify if these should be separate states.

### Description
Evaluate whether projects need a separate `expired` phase distinct from `cancelled`, specifically for the "all offers lost" scenario.

### Acceptance Criteria
- [ ] Define business difference between expired and cancelled
- [ ] If distinct: Add `expired` to ProjectPhase enum
- [ ] Update transitions: all offers lost -> expired (not cancelled)
- [ ] Cancelled reserved for manual user action
- [ ] Document state definitions

### Technical Notes
- Requires schema migration if adding new phase
- Update all state machine logic

### Priority: Low (clarification needed)
### Dependencies: Business requirements clarification

---

## Story Proposal: UI Auto-Project Naming Prompt

### Discovery Context
Auto-created projects use `[AUTO] {offer.Title}` as name, which is not user-friendly long-term. Users should be prompted to rename.

### Description
Frontend feature to prompt users to rename auto-created projects.

### Acceptance Criteria
- [ ] Detect projects with `[AUTO]` prefix in name
- [ ] Show inline prompt suggesting rename
- [ ] Allow one-click access to rename dialog
- [ ] Track whether prompt was dismissed
- [ ] Don't show prompt repeatedly if dismissed

### Technical Notes
- Frontend-only change
- Consider adding `is_auto_created` flag to project model

### Priority: Low
### Dependencies:
- Offer-Project Lifecycle Management (parent story)
- Frontend team availability

---

## Story Proposal: Offer-Project Link Change Audit Trail

### Discovery Context
When offers are linked/unlinked from projects, there's no historical record of these associations.

### Description
Track history of offer-project relationships for audit purposes.

### Acceptance Criteria
- [ ] Log when offer is linked to project (including auto-create)
- [ ] Log when offer is unlinked from project
- [ ] Store previous project ID when relationship changes
- [ ] Queryable audit history per offer and per project
- [ ] Include timestamp and user who made change

### Technical Notes
- New `offer_project_history` table
- Or leverage existing Activity system with structured metadata

### Priority: Medium
### Dependencies: Offer-Project Lifecycle Management (parent story)

---

## Story Proposal: Project Economics History

### Discovery Context
During tilbud phase, project economics change as offers are added/updated/lost. No history of these changes is maintained.

### Description
Track history of project economics changes for analysis and auditing.

### Acceptance Criteria
- [ ] Log each time CalculatedOfferValue changes
- [ ] Store: old value, new value, trigger (offer added/updated/lost), offer ID
- [ ] Queryable history per project
- [ ] Consider visualization in project dashboard

### Technical Notes
- New `project_economics_history` table
- Trigger from economics sync method

### Priority: Low
### Dependencies: Offer-Project Lifecycle Management (parent story)

---

## Summary

| Story | Priority | Blocker? |
|-------|----------|----------|
| Bulk Migration for Existing Offers | High | Yes - blocks deployment |
| Project Reactivation | Medium | No |
| Project Cancellation Notifications | Medium | No |
| Offer-Project Link Audit Trail | Medium | No |
| Expired vs Cancelled Clarification | Low | No - needs BA input |
| Custom Best Offer Rules | Low | No |
| UI Auto-Project Naming | Low | No |
| Project Economics History | Low | No |

**Recommended order:**
1. Bulk Migration (required before deployment)
2. Main feature deployment
3. Project Reactivation
4. Notifications
5. Audit trail
6. Lower priority items
