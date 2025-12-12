-- +goose Up
-- +goose StatementBegin

-- Dashboard Metrics Aggregation View
-- Solves double-counting problem: when a project has multiple offers,
-- we should only count the MAX(value) per project per phase, not sum all offers.
-- Orphan offers (no project_id) are included at full value.
--
-- This view provides pre-computed metrics for the dashboard:
-- - For offers WITH project_id: GROUP BY project_id, phase, company_id - select MAX(value) as best offer
-- - For orphan offers (project_id IS NULL): include at full value
-- - Aggregate metrics per phase per company

CREATE OR REPLACE VIEW dashboard_metrics_aggregation AS
WITH
-- Get best (highest value) offer per project per phase per company
-- Only for offers that have a project_id
project_best_offers AS (
    SELECT
        o.company_id,
        o.phase,
        o.project_id,
        MAX(o.value) AS best_value,
        -- For weighted value, we take the max value and multiply by the probability of that offer
        -- We need to identify which offer has the max value to get its probability
        (
            SELECT o2.probability
            FROM offers o2
            WHERE o2.project_id = o.project_id
              AND o2.phase = o.phase
              AND o2.company_id = o.company_id
              AND o2.value = MAX(o.value)
            LIMIT 1
        ) AS best_probability,
        COUNT(*) AS offer_count
    FROM offers o
    WHERE o.project_id IS NOT NULL
      AND o.phase NOT IN ('draft', 'expired')
    GROUP BY o.company_id, o.phase, o.project_id
),
-- Get orphan offers (no project_id) - these are counted at full value
orphan_offers AS (
    SELECT
        o.company_id,
        o.phase,
        o.value,
        o.probability,
        1 AS offer_count
    FROM offers o
    WHERE o.project_id IS NULL
      AND o.phase NOT IN ('draft', 'expired')
),
-- Combine project best offers and orphan offers for final aggregation
combined_metrics AS (
    -- Project-based offers (using best value per project)
    SELECT
        company_id,
        phase,
        project_id,
        best_value AS value,
        best_probability AS probability,
        offer_count,
        1 AS project_count  -- Each row represents one unique project
    FROM project_best_offers

    UNION ALL

    -- Orphan offers (no project)
    SELECT
        company_id,
        phase,
        NULL AS project_id,
        value,
        probability,
        offer_count,
        0 AS project_count  -- Orphans don't count as projects
    FROM orphan_offers
)
-- Final aggregation by phase and company
SELECT
    company_id,
    phase,
    SUM(project_count) AS project_count,          -- Unique projects in this phase
    SUM(offer_count) AS offer_count,              -- Total offers (including all offers per project)
    SUM(value) AS total_value,                    -- Sum of best values (no double-counting)
    SUM(value * COALESCE(probability, 0) / 100.0) AS weighted_value,  -- Weighted by probability
    NOW() AS computed_at
FROM combined_metrics
GROUP BY company_id, phase;

-- Add comment explaining the view
COMMENT ON VIEW dashboard_metrics_aggregation IS
'Pre-computed dashboard metrics that avoid double-counting.
For projects with multiple offers, only the highest value offer per phase is counted.
Orphan offers (without project) are included at full value.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS dashboard_metrics_aggregation;
-- +goose StatementEnd
