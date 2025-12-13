-- =============================================================================
-- Straye Tak Projects Import: WON Offers
-- =============================================================================
-- Generated: 2025-12-12
--
-- Creates 92 projects from won offers:
--   - status: 'active'
--   - project_number: inherited from offer.external_reference (e.g., "22000")
--
-- IMPORTANT: Run AFTER import_projects_not_won.sql
-- =============================================================================

-- Create projects for won offers
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
    value,
    cost,
    spent,
    manager_id,
    manager_name,
    offer_id,
    project_number,
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
    'active' as status,
    o.sent_date::date as start_date,  -- NULL if no sent_date, fix later from ERP
    o.value as value,
    o.cost as cost,
    0 as spent,
    NULL as manager_id,
    o.responsible_user_name as manager_name,
    o.id as offer_id,
    o.external_reference as project_number,  -- Inherits "22000" etc. from offer
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM offers o
WHERE o.company_id = 'tak'
  AND o.phase = 'won';

-- Update offers to link back to their projects (also set project_name)
UPDATE offers o
SET project_id = p.id,
    project_name = p.name
FROM projects p
WHERE p.offer_id = o.id
  AND o.company_id = 'tak'
  AND o.phase = 'won';

-- =============================================================================
-- Summary
-- =============================================================================
-- Projects created: 92
--   - All with status 'active'
--   - project_number = offer.external_reference (e.g., "22000", "23044")
--
-- Notes:
--   - Won offers "convert" to projects, inheriting the reference number
--   - start_date may be NULL if offer had no sent_date
--   - Bidirectional link: project.offer_id <-> offer.project_id
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won.sql
-- =============================================================================
