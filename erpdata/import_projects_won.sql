-- =============================================================================
-- Straye Tak Projects Import: WON Offers
-- =============================================================================
-- Generated: 2025-12-13 (FIXED)
--
-- Creates projects from won offers:
--   - phase: 'active'
--   - project_number: generated from sequence (TK-YYYY-PNN)
--   - inherited_offer_number: from winning offer's offer_number
--
-- IMPORTANT: Run AFTER import_offers_fixed.sql
-- =============================================================================

-- Create projects for won offers with proper number generation
-- Using a CTE to generate sequential project numbers
WITH numbered_offers AS (
    SELECT
        o.*,
        ROW_NUMBER() OVER (
            PARTITION BY EXTRACT(YEAR FROM COALESCE(o.sent_date, CURRENT_DATE))
            ORDER BY o.sent_date, o.id
        ) as seq_num,
        EXTRACT(YEAR FROM COALESCE(o.sent_date, CURRENT_DATE))::int as offer_year
    FROM offers o
    WHERE o.company_id = 'tak'
      AND o.phase = 'won'
)
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
    'TK-' || o.offer_year::text || '-P' || LPAD(o.seq_num::text, 3, '0') as project_number,
    o.offer_number as inherited_offer_number,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM numbered_offers o;

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
-- Projects created from won offers
--   - All with phase 'active'
--   - project_number = generated (e.g., "TK-2023-P001")
--   - inherited_offer_number = offer's internal number (e.g., "TK-2023-001")
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won_fixed.sql
-- =============================================================================
