-- =============================================================================
-- Straye Tak Projects Import: WON Offers
-- =============================================================================
-- Generated: 2025-12-14 (UPDATED for new numbering convention + completed status)
--
-- Creates projects from won offers:
--   - phase: 'completed' for finished projects, 'active' for ongoing
--   - project_number: from offer's original offer_number (e.g., "TK-2023-001")
--   - inherited_offer_number: same as project_number
--
-- Also updates won offers to add "W" suffix to their offer_number
--   - e.g., "TK-2023-001" becomes "TK-2023-001W"
--
-- IMPORTANT: Run AFTER import_offers.sql
-- =============================================================================

-- External references for completed (Ferdig) projects from Excel
-- These will get phase 'completed' instead of 'active'
-- Total: 61 completed projects

-- Step 1: Create projects for won offers
-- Project claims the offer's original number as project_number
-- Phase is 'completed' if project is marked as Ferdig in Excel, otherwise 'active'
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
    CASE
        WHEN o.external_reference IN (
            '22001', '22002', '22003', '22006', '22010', '22014', '22015', '22019', '22025',
            '23002', '23005', '23014', '23016', '23019', '23020', '23023', '23036', '23038',
            '23040', '23042', '23043', '23049', '23050', '23053', '23054', '23055', '23057',
            '23068', '23073', '23080', '23081', '23096', '23097', '23101',
            '24004', '24006', '24023', '24030', '24031', '24035', '24038', '24047', '24055',
            '24062', '24075', '24086', '24093', '24115', '24122', '24125', '24124', '24128',
            '24138', '24139', '24145', '24147', '24156',
            '25042', '25048', '25051', '25503'
        ) THEN 'completed'::project_phase
        ELSE 'active'::project_phase
    END as phase,
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
--   - phase: 'completed' (61 finished projects) or 'active' (ongoing)
--   - project_number: offer's original number (e.g., "TK-2023-001")
--   - inherited_offer_number: same as project_number
--   - location: inherited from offer
--
-- Won offers updated:
--   - offer_number: now has "W" suffix (e.g., "TK-2023-001W")
--   - project_id: linked to created project
--
-- Example:
--   Before: Offer "TK-2023-001" (won, Ferdig)
--   After:  Offer "TK-2023-001W" -> Project "TK-2023-001" (completed)
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won.sql
-- =============================================================================
