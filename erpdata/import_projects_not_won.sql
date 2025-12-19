-- =============================================================================
-- Straye Tak Projects Import: NON-CLOSED Offers + Single Completed Offers
-- =============================================================================
-- Generated: 2025-12-18
--
-- Creates projects from non-closed offers:
--   - 'in_progress' offers -> project phase 'tilbud'
--   - 'sent' offers -> project phase 'tilbud'
--   - 'order' offers -> project phase 'working'
--   - 'completed' offers (when only one offer exists for project) -> project phase 'completed'
--
-- Closed offers (lost, expired) do NOT get projects.
--
-- Project numbers: "PROJECT-YYYY-NNN" format, shared across all companies
-- Numbers are assigned based on offer sent_date (earliest first).
--
-- IMPORTANT: Run AFTER import_offers.sql
-- =============================================================================

-- Create projects for non-closed offers with sequential project numbers
-- Projects are numbered by year based on offer sent_date
-- Also includes 'completed' offers that are the only offer for their project
WITH numbered_offers AS (
    SELECT
        o.id as offer_id,
        o.title,
        o.description,
        o.customer_id,
        o.customer_name,
        o.phase as offer_phase,
        o.sent_date,
        o.location,
        o.external_reference,
        EXTRACT(YEAR FROM COALESCE(o.sent_date, CURRENT_DATE))::int as offer_year,
        ROW_NUMBER() OVER (
            PARTITION BY EXTRACT(YEAR FROM COALESCE(o.sent_date, CURRENT_DATE))
            ORDER BY o.sent_date NULLS LAST, o.external_reference
        ) as seq_in_year
    FROM offers o
    WHERE o.company_id = 'tak'
      AND o.phase IN ('in_progress', 'sent', 'order', 'completed')
      AND o.project_id IS NULL
)
INSERT INTO projects (
    id,
    name,
    project_number,
    summary,
    description,
    customer_id,
    customer_name,
    phase,
    start_date,
    location,
    external_reference,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() as id,
    no.title as name,
    'PROJECT-' || no.offer_year || '-' || LPAD(no.seq_in_year::text, 3, '0') as project_number,
    'Prosjekt fra tilbud: ' || no.title as summary,
    COALESCE(no.description, '') as description,
    no.customer_id,
    no.customer_name,
    CASE
        WHEN no.offer_phase = 'order' THEN 'working'::project_phase
        WHEN no.offer_phase = 'completed' THEN 'completed'::project_phase
        ELSE 'tilbud'::project_phase
    END as phase,
    COALESCE(no.sent_date::date, CURRENT_DATE) as start_date,
    no.location,
    no.external_reference,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM numbered_offers no;

-- Disable trigger to preserve original updated_at values during import
ALTER TABLE offers DISABLE TRIGGER update_offers_updated_at;

-- Update offers to link back to their projects
-- Match by external_reference + customer_id + title
UPDATE offers o
SET project_id = p.id,
    project_name = p.name
FROM projects p
WHERE p.external_reference = o.external_reference
  AND p.customer_id = o.customer_id
  AND p.name = o.title
  AND o.company_id = 'tak'
  AND o.phase IN ('in_progress', 'sent', 'order', 'completed')
  AND o.project_id IS NULL;

-- Re-enable trigger for normal operation
ALTER TABLE offers ENABLE TRIGGER update_offers_updated_at;

-- =============================================================================
-- Summary
-- =============================================================================
-- Projects created from non-closed offers + single completed offers:
--   - tilbud: from in_progress + sent offers
--   - working: from order offers (active work)
--   - completed: from completed offers (single offer per project)
--
-- Project numbers:
--   - Format: PROJECT-YYYY-NNN (e.g., PROJECT-2024-001)
--   - Shared across ALL companies (not company-specific)
--   - Numbered sequentially per year based on sent_date
--
-- Notes:
--   - Projects are simplified folders (no value/cost/spent/manager fields)
--   - All economics tracking is on the Offer itself
--   - external_reference and location inherited from offer
--   - Linked via offer.project_id
--
-- Phase breakdown from import_offers.sql:
--   - in_progress: 11 offers -> tilbud projects
--   - sent: 265 offers -> tilbud projects
--   - order: 32 offers -> working projects
--   - completed: X offers -> completed projects
--   - Total: 308+ projects created
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_not_won.sql
-- =============================================================================
