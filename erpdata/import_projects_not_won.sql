-- =============================================================================
-- Straye Tak Projects Import: NON-WON Offers
-- =============================================================================
-- Generated: 2025-12-14 (UPDATED)
--
-- Creates projects from non-won offers:
--   - 'sent' offers -> phase 'tilbud'
--   - 'in_progress' offers -> phase 'tilbud'
--   - 'lost' offers -> phase 'cancelled'
--
-- Note: Non-won projects do NOT get a project_number assigned
-- (only projects from won offers get numbered)
--
-- IMPORTANT: Run AFTER import_offers.sql
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
    phase,
    start_date,
    value,
    cost,
    spent,
    manager_id,
    manager_name,
    offer_id,
    external_reference,
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
        WHEN o.phase = 'lost' THEN 'cancelled'::project_phase
        ELSE 'tilbud'::project_phase
    END as phase,
    o.sent_date::date as start_date,
    o.value as value,
    o.cost as cost,
    0 as spent,
    NULL as manager_id,
    o.responsible_user_name as manager_name,
    o.id as offer_id,
    o.external_reference as external_reference,  -- Inherit external reference from offer
    o.location as location,                      -- Inherit location from offer
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM offers o
WHERE o.company_id = 'tak'
  AND o.phase IN ('sent', 'in_progress', 'lost');

-- Update offers to link back to their projects (also set project_name)
UPDATE offers o
SET project_id = p.id,
    project_name = p.name
FROM projects p
WHERE p.offer_id = o.id
  AND o.company_id = 'tak'
  AND o.phase IN ('sent', 'in_progress', 'lost');

-- =============================================================================
-- Summary
-- =============================================================================
-- Projects created from non-won offers:
--   - tilbud: from sent + in_progress offers
--   - cancelled: from lost offers
--
-- Notes:
--   - No project_number assigned (only won projects get numbers)
--   - start_date may be NULL if offer had no sent_date
--   - external_reference inherited from offer
--   - location inherited from offer
--   - Bidirectional link: project.offer_id <-> offer.project_id
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_not_won.sql
-- =============================================================================
