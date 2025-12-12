# Straye Tak Projects Import Plan

Last updated: 2025-12-12

## Overview

This document describes the strategy for creating projects linked to imported offers.

**Key insight**: Won offers inherit their project number from the offer's `external_reference`. Therefore, we process in this order:
1. First: Create projects for **non-won** offers (sent, in_progress, lost)
2. Second: Create projects for **won** offers (separate SQL)

## Offer Distribution

| Phase | Count | Project Status | Notes |
|-------|-------|----------------|-------|
| sent | 264 | `planning` | Active opportunities |
| in_progress | 10 | `planning` | Being worked on |
| lost | 55 | `cancelled` | Did not win |
| won | 92 | `active` | Will inherit offer number |

---

## Phase 1: Non-Won Offers (329 projects)

### Status Mapping

| Offer Phase | Project Status | Rationale |
|-------------|----------------|-----------|
| `sent` | `planning` | Opportunity sent to customer, awaiting response |
| `in_progress` | `planning` | Still being prepared |
| `lost` | `cancelled` | Did not materialize |

### Project Fields Mapping

| Project Field | Source | Notes |
|---------------|--------|-------|
| `name` | `offer.title` | Project name from offer |
| `summary` | Generated | "Prosjekt fra tilbud: {title}" |
| `customer_id` | `offer.customer_id` | Same customer |
| `customer_name` | `offer.customer_name` | Denormalized |
| `company_id` | `'tak'` | All Straye Tak |
| `status` | Mapped from phase | See table above |
| `start_date` | `offer.sent_date` or `NULL` | When offer was sent (fix later from ERP) |
| `value` | `offer.value` | Offer value (revenue) |
| `cost` | `offer.cost` | Offer cost |
| `margin_percent` | Auto-calculated | (value - cost) / value * 100 |
| `manager_name` | `offer.responsible_user_name` | HSK, etc. |
| `offer_id` | `offer.id` | Link back to offer |

### SQL Strategy

1. Insert projects with generated UUIDs
2. Update offers to set `project_id` linking to the new projects

---

## SQL: Non-Won Projects

```sql
-- =============================================================================
-- PHASE 1: Create projects for NON-WON offers (sent, in_progress, lost)
-- =============================================================================
--
-- This creates 329 projects:
--   - 264 from 'sent' offers → status 'prospect'
--   - 10 from 'in_progress' offers → status 'prospect'
--   - 55 from 'lost' offers → status 'cancelled'
--
-- Run this AFTER import_offers.sql
-- =============================================================================

-- Create projects for non-won offers
INSERT INTO projects (
    id,
    name,
    summary,
    description,
    customer_id,
    customer_name,
    company_id,
    status,
    start_date,
    budget,
    spent,
    manager_id,
    manager_name,
    offer_id,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() as id,
    o.title as name,
    'Prosjekt fra tilbud: ' || o.title as summary,
    o.description,
    o.customer_id,
    o.customer_name,
    o.company_id,
    CASE
        WHEN o.phase = 'lost' THEN 'cancelled'
        ELSE 'planning'
    END as status,
    o.sent_date::date as start_date,  -- NULL if no sent_date, fix later from ERP
    o.value as budget,
    0 as spent,
    NULL as manager_id,
    o.responsible_user_name as manager_name,
    o.id as offer_id,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM offers o
WHERE o.company_id = 'tak'
  AND o.phase IN ('sent', 'in_progress', 'lost');

-- Update offers to link back to their projects
UPDATE offers o
SET project_id = p.id
FROM projects p
WHERE p.offer_id = o.id
  AND o.company_id = 'tak'
  AND o.phase IN ('sent', 'in_progress', 'lost');
```

---

## Phase 2: Won Offers (92 projects)

Won offers are special because:
- Project inherits the offer's `external_reference` as `project_number`
- Status: `active`
- This is the "conversion" of offer → project

**File:** `import_projects_won.sql`

```sql
-- Key difference: project_number = offer.external_reference
INSERT INTO projects (
    ...
    project_number,  -- "22000", "23044", etc.
    ...
)
SELECT
    ...
    o.external_reference as project_number,
    ...
FROM offers o
WHERE o.company_id = 'tak'
  AND o.phase = 'won';
```

---

## Execution Order

```bash
# 1. First, import customers (if not already done)
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_customers.sql

# 2. Import offers
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_offers.sql

# 3. Create projects for non-won offers
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_not_won.sql

# 4. Create projects for won offers (inherits project numbers)
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won.sql
```

---

## Verification Queries

After running, verify with:

```sql
-- Count projects by status
SELECT status, COUNT(*)
FROM projects
WHERE company_id = 'tak'
GROUP BY status;

-- Verify all non-won offers have linked projects
SELECT o.phase, COUNT(*) as offers, COUNT(o.project_id) as with_project
FROM offers o
WHERE o.company_id = 'tak'
GROUP BY o.phase;

-- Check bidirectional links
SELECT COUNT(*) as orphaned_links
FROM projects p
LEFT JOIN offers o ON o.id = p.offer_id
WHERE p.company_id = 'tak' AND o.id IS NULL;
```

---

## Decisions Made

1. **Lost offers**: Keep them for history → status `cancelled`

2. **Project numbers**: Only won projects get `project_number` (inherited from offer's `external_reference`)

3. **Start date**: Use `sent_date` if available, otherwise `NULL` (fix later from ERP data)

4. **Won projects**: Inherit `external_reference` as `project_number` (e.g., "22000") - this is the "conversion" from offer to project
