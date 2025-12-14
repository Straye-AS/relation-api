-- =============================================================================
-- Straye Tak Projects Import: WON Offers
-- =============================================================================
-- Generated: 2025-12-14 (UPDATED for new numbering convention)
--
-- Creates projects from won offers:
--   - phase: 'active'
--   - project_number: from offer's original offer_number (e.g., "TK-2023-001")
--   - inherited_offer_number: same as project_number
--
-- Also updates won offers to add "W" suffix to their offer_number
--   - e.g., "TK-2023-001" becomes "TK-2023-001W"
--
-- IMPORTANT: Run AFTER import_offers.sql
-- =============================================================================

-- Step 1: Create projects for won offers
-- Project claims the offer's original number as project_number
INSERT INTO projects (
    id,
    name,
    summary,
    description,
    customer_id,
    customer_name,
    company_id,
    phase,
    start_date,
    value,
    cost,
    spent,
    manager_id,
    manager_name,
    offer_id,
    project_number,
    inherited_offer_number,
    location,
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
    'active'::project_phase as phase,
    o.sent_date::date as start_date,
    o.value as value,
    o.cost as cost,
    0 as spent,
    NULL as manager_id,
    o.responsible_user_name as manager_name,
    o.id as offer_id,
    o.offer_number as project_number,          -- Project claims offer's original number
    o.offer_number as inherited_offer_number,  -- Same as project_number
    o.location as location,                    -- Inherit location from offer
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM offers o
WHERE o.company_id = 'tak'
  AND o.phase = 'won';

-- Step 2: Update offers to link back to their projects
UPDATE offers o
SET project_id = p.id,
    project_name = p.name
FROM projects p
WHERE p.offer_id = o.id
  AND o.company_id = 'tak'
  AND o.phase = 'won';

-- Step 3: Add "W" suffix to won offer numbers (only if not already suffixed)
UPDATE offers
SET offer_number = offer_number || 'W'
WHERE company_id = 'tak'
  AND phase = 'won'
  AND offer_number NOT LIKE '%W';

-- =============================================================================
-- Summary
-- =============================================================================
-- Projects created from won offers:
--   - phase: 'active'
--   - project_number: offer's original number (e.g., "TK-2023-001")
--   - inherited_offer_number: same as project_number
--   - location: inherited from offer
--
-- Won offers updated:
--   - offer_number: now has "W" suffix (e.g., "TK-2023-001W")
--   - project_id: linked to created project
--
-- Example:
--   Before: Offer "TK-2023-001" (won)
--   After:  Offer "TK-2023-001W" -> Project "TK-2023-001"
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won.sql
-- =============================================================================
