# Sprint 3: Core CRM & Offer Management Sprint - Implementation Order

**Sprint Dates:** December 6-8, 2025
**Iteration:** [160](https://app.shortcut.com/straye/iteration/160)
**Total Stories:** 25 stories, ~102 story points

This document defines the recommended implementation order for Sprint 3 stories, organized by dependency chains and logical groupings.

---

## URGENT: Phase 1 Bug Fixes

*These bugs were identified during acceptance criteria verification and must be fixed before continuing to Phase 2.*

| Order | Story ID | Name | Type | Original Story |
|-------|----------|------|------|----------------|
| 0.1 | [sc-161](https://app.shortcut.com/straye/story/161) | Customer filters missing status, tier, and industry fields | Bug | sc-82 |
| 0.2 | [sc-162](https://app.shortcut.com/straye/story/162) | Customer service missing email/phone format validation | Bug | sc-83 |
| 0.3 | [sc-163](https://app.shortcut.com/straye/story/163) | Customer tier assignment logic not implemented | Bug | sc-83 |
| 0.4 | [sc-164](https://app.shortcut.com/straye/story/164) | Contact repository missing contact_type filter | Bug | sc-85 |
| 0.5 | [sc-165](https://app.shortcut.com/straye/story/165) | Contact service missing email uniqueness validation | Bug | sc-87 |

**Notes:**
- These gaps were found during Phase 1 acceptance criteria review
- sc-161 and sc-163 require database migration (new fields on Customer)
- sc-162 adds validation helpers to customer service
- sc-164 needs clarification: contact_type vs entity_type filter
- sc-165 requires unique constraint migration on contact email

---

## Phase 1: Customer & Contact Foundation (COMPLETED)

*These are foundational entities - many other features depend on customers and contacts.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 1 | [sc-82](https://app.shortcut.com/straye/story/82) | Implement Customer repository with search and filtering | 4 | High | None |
| 2 | [sc-83](https://app.shortcut.com/straye/story/83) | Implement Customer service with business validation | 4 | High | sc-82 |
| 3 | [sc-84](https://app.shortcut.com/straye/story/84) | Create Customer API endpoints (CRUD + search) | 4 | Highest | sc-83 |
| 4 | [sc-85](https://app.shortcut.com/straye/story/85) | Implement Contact repository with polymorphic relationships | 4 | Highest | None |
| 5 | [sc-87](https://app.shortcut.com/straye/story/87) | Implement Contact service with relationship logic | 4 | Highest | sc-85 |
| 6 | [sc-88](https://app.shortcut.com/straye/story/88) | Create Contact API endpoints (CRUD + relationships) | 4 | Highest | sc-87 |

**Notes:**
- sc-82 and sc-85 can be done in parallel (independent repositories)
- Customer is a core entity referenced by deals, offers, and projects
- Contact has polymorphic relationships to customer, deal, project
- Activity logging should be integrated from the start

---

## Phase 2: Budget Dimension Infrastructure

*Budget dimensions are used by both offers and projects - build this before offer/project services.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 7 | [sc-98](https://app.shortcut.com/straye/story/98) | Implement BudgetDimensionCategory repository and seed data | 4 | Low | None |
| 8 | [sc-99](https://app.shortcut.com/straye/story/99) | Implement BudgetDimension repository with margin calculations | 4 | Low | sc-98 |
| 9 | [sc-100](https://app.shortcut.com/straye/story/100) | Implement BudgetDimension service with business logic | 4 | Medium | sc-99 |

**Notes:**
- Categories are relatively static (Labor, Materials, Equipment, etc.)
- BudgetDimension supports manual margin override calculations
- These are needed before offer and project services can calculate totals
- Consider caching category list in memory for performance

---

## Phase 3: Offer Management

*Offers depend on customers and budget dimensions. Deals reference offers.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 10 | [sc-101](https://app.shortcut.com/straye/story/101) | Implement Offer repository with status and phase management | 4 | Low | sc-84, sc-100 |
| 11 | [sc-102](https://app.shortcut.com/straye/story/102) | Implement Offer service with lifecycle workflows | 8 | High | sc-101, sc-100 |
| 12 | [sc-103](https://app.shortcut.com/straye/story/103) | Create Offer API endpoints (CRUD + lifecycle) | 4 | Highest | sc-102 |
| 13 | [sc-104](https://app.shortcut.com/straye/story/104) | Create Offer budget dimension management endpoints | 4 | Highest | sc-103, sc-100 |

**Notes:**
- Offer lifecycle: draft → sent → accepted/rejected
- sc-102 is larger (8 pts) - includes SendOffer, AcceptOffer, CloneOffer workflows
- Budget dimension endpoints allow managing cost/revenue line items
- AcceptOffer can optionally create a project

---

## Phase 4: Deal Workflows & Analytics

*Deal workflows depend on offers being in place.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 14 | [sc-95](https://app.shortcut.com/straye/story/95) | Implement Deal Loss workflow with reason tracking | 2 | High | None (uses existing Deal) |
| 15 | [sc-96](https://app.shortcut.com/straye/story/96) | Implement Create Offer from Deal workflow | 4 | Highest | sc-103 |
| 16 | [sc-97](https://app.shortcut.com/straye/story/97) | Create sales pipeline analytics and forecasting | 8 | Highest | sc-95 |

**Notes:**
- sc-95 adds loss reasons (price, timing, competitor, requirements, other)
- sc-96 creates offer linked to deal with optional template support
- sc-97 is larger (8 pts) - includes pipeline summary, forecasting, conversion rates
- Analytics use v_sales_pipeline_summary database view

---

## Phase 5: Project Management

*Projects are created from accepted offers and inherit budget dimensions.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 17 | [sc-107](https://app.shortcut.com/straye/story/107) | Implement Project repository with health tracking and filtering | 4 | Highest | sc-100 |
| 18 | [sc-109](https://app.shortcut.com/straye/story/109) | Implement Project service with budget inheritance logic | 4 | Medium | sc-107, sc-100 |
| 19 | [sc-110](https://app.shortcut.com/straye/story/110) | Create Project API endpoints (CRUD + status) | 4 | Highest | sc-109 |
| 20 | [sc-111](https://app.shortcut.com/straye/story/111) | Implement project budget inheritance from offer | 4 | Medium | sc-110, sc-102 |

**Notes:**
- Project health: on_track (<10% variance), at_risk (10-20%), over_budget (>20%)
- sc-111 clones budget dimensions from offer to project
- Projects track actual costs vs budgeted costs
- Budget inheritance requires both offer service and project service

---

## Phase 6: Activity & CRM Interactions

*Activities have polymorphic relationships to all entities (customer, deal, offer, project).*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 21 | [sc-116](https://app.shortcut.com/straye/story/116) | Implement Activity repository with polymorphic entity support | 4 | High | Phase 1-5 entities |
| 22 | [sc-117](https://app.shortcut.com/straye/story/117) | Implement Activity service with task assignment and notifications | 4 | Highest | sc-116, sc-126 |
| 23 | [sc-118](https://app.shortcut.com/straye/story/118) | Create Activity API endpoints (CRUD + my tasks) | 4 | Highest | sc-117 |

**Notes:**
- Activity types: meeting, task, call, email, note
- Polymorphic: entity_type (customer, deal, offer, project) + entity_id
- Task assignment creates notifications
- Includes my-tasks and upcoming meetings endpoints

---

## Phase 7: Notification System

*Notifications support activity assignments and system alerts.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 24 | [sc-126](https://app.shortcut.com/straye/story/126) | Implement Notification repository and service | 4 | Lowest | None |
| 25 | [sc-127](https://app.shortcut.com/straye/story/127) | Create Notification API endpoints | 4 | Medium | sc-126 |

**Notes:**
- sc-126 can be started early (no dependencies) but is lowest priority
- Notification types: task_assigned, budget_alert, deal_stage_changed
- Future: WebSocket for real-time delivery
- Auto-delete notifications older than 90 days

---

## Dependency Graph

```
Phase 1: Customer & Contact
sc-82 (Customer repo)
  └── sc-83 (Customer service)
        └── sc-84 (Customer API)

sc-85 (Contact repo)          [parallel with sc-82]
  └── sc-87 (Contact service)
        └── sc-88 (Contact API)

Phase 2: Budget Dimensions
sc-98 (BudgetDimensionCategory)
  └── sc-99 (BudgetDimension repo)
        └── sc-100 (BudgetDimension service)

Phase 3: Offers
sc-84 + sc-100
  └── sc-101 (Offer repo)
        └── sc-102 (Offer service)
              └── sc-103 (Offer API)
                    └── sc-104 (Offer budget endpoints)

Phase 4: Deal Workflows
sc-95 (Deal Loss)              [can start early]
sc-103 (Offer API)
  └── sc-96 (Create Offer from Deal)
sc-95
  └── sc-97 (Pipeline analytics)

Phase 5: Projects
sc-100 (Budget service)
  └── sc-107 (Project repo)
        └── sc-109 (Project service)
              └── sc-110 (Project API)
                    └── sc-111 (Budget inheritance)

Phase 6: Activities
All Phase 1-5 entities + sc-126
  └── sc-116 (Activity repo)
        └── sc-117 (Activity service)
              └── sc-118 (Activity API)

Phase 7: Notifications
sc-126 (Notification repo+service)  [independent, can start anytime]
  └── sc-127 (Notification API)
```

---

## Quick Reference: Story Details

| ID | Name | Epic | Est | Priority |
|----|------|------|-----|----------|
| sc-82 | Implement Customer repository with search and filtering | Customer & Contact Management | 4 | High |
| sc-83 | Implement Customer service with business validation | Customer & Contact Management | 4 | High |
| sc-84 | Create Customer API endpoints (CRUD + search) | Customer & Contact Management | 4 | Highest |
| sc-85 | Implement Contact repository with polymorphic relationships | Customer & Contact Management | 4 | Highest |
| sc-87 | Implement Contact service with relationship logic | Customer & Contact Management | 4 | Highest |
| sc-88 | Create Contact API endpoints (CRUD + relationships) | Customer & Contact Management | 4 | Highest |
| sc-95 | Implement Deal Loss workflow with reason tracking | Sales Funnel & Deal Management | 2 | High |
| sc-96 | Implement Create Offer from Deal workflow | Sales Funnel & Deal Management | 4 | Highest |
| sc-97 | Create sales pipeline analytics and forecasting | Sales Funnel & Deal Management | 8 | Highest |
| sc-98 | Implement BudgetDimensionCategory repository and seed data | Offer & Budget Management | 4 | Low |
| sc-99 | Implement BudgetDimension repository with margin calculations | Offer & Budget Management | 4 | Low |
| sc-100 | Implement BudgetDimension service with business logic | Offer & Budget Management | 4 | Medium |
| sc-101 | Implement Offer repository with status and phase management | Offer & Budget Management | 4 | Low |
| sc-102 | Implement Offer service with lifecycle workflows | Offer & Budget Management | 8 | High |
| sc-103 | Create Offer API endpoints (CRUD + lifecycle) | Offer & Budget Management | 4 | Highest |
| sc-104 | Create Offer budget dimension management endpoints | Offer & Budget Management | 4 | Highest |
| sc-107 | Implement Project repository with health tracking and filtering | Project Management & Tracking | 4 | Highest |
| sc-109 | Implement Project service with budget inheritance logic | Project Management & Tracking | 4 | Medium |
| sc-110 | Create Project API endpoints (CRUD + status) | Project Management & Tracking | 4 | Highest |
| sc-111 | Implement project budget inheritance from offer | Project Management & Tracking | 4 | Medium |
| sc-116 | Implement Activity repository with polymorphic entity support | Activity & CRM Interactions | 4 | High |
| sc-117 | Implement Activity service with task assignment and notifications | Activity & CRM Interactions | 4 | Highest |
| sc-118 | Create Activity API endpoints (CRUD + my tasks) | Activity & CRM Interactions | 4 | Highest |
| sc-126 | Implement Notification repository and service | Notification System | 4 | Lowest |
| sc-127 | Create Notification API endpoints | Notification System | 4 | Medium |

---

## Implementation Tips

### Grouping for PRs
Stories can be grouped into logical PRs:

1. **PR 1: Customer Management** - sc-82, sc-83, sc-84
2. **PR 2: Contact Management** - sc-85, sc-87, sc-88
3. **PR 3: Budget Dimensions** - sc-98, sc-99, sc-100
4. **PR 4: Offer Core** - sc-101, sc-102, sc-103
5. **PR 5: Offer Budget** - sc-104
6. **PR 6: Deal Workflows** - sc-95, sc-96, sc-97
7. **PR 7: Project Management** - sc-107, sc-109, sc-110, sc-111
8. **PR 8: Activity System** - sc-116, sc-117, sc-118
9. **PR 9: Notifications** - sc-126, sc-127

### Parallel Work Opportunities
- sc-82 (Customer repo) and sc-85 (Contact repo) can be done in parallel
- sc-126 (Notifications) has no dependencies - can start anytime
- sc-95 (Deal Loss) can start early (uses existing Deal infrastructure)
- After Phase 2, multiple workstreams can proceed in parallel

### Architecture Pattern Reminder
Follow the clean architecture pattern for each entity:
```
Repository (data access) → Service (business logic) → Handler (HTTP)
```

### Validation Points
After each phase, validate:
1. **Phase 1:** Customer and Contact CRUD working, relationships correct
2. **Phase 2:** Budget dimensions calculate margins correctly
3. **Phase 3:** Offer lifecycle works (draft → sent → accepted)
4. **Phase 4:** Deal loss tracking, offer creation from deal
5. **Phase 5:** Project budget inheritance from offer works
6. **Phase 6:** Activities linked to all entity types, tasks assignable
7. **Phase 7:** Notifications created and delivered

### Key Business Logic
- **Margin calculation:** revenue = cost / (1 - (target_margin_percent / 100))
- **Project health:** on_track (<10%), at_risk (10-20%), over_budget (>20%)
- **Deal loss reasons:** price, timing, competitor, requirements, other
- **Offer lifecycle:** draft → in_progress → sent → won/lost/expired

---

*Document created: 2025-12-06*
*Sprint: Core CRM & Offer Management Sprint (Iteration 160)*
