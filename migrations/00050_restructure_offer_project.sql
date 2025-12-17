-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- RESTRUCTURE OFFER-PROJECT RELATIONSHIP
-- ============================================================================
-- This migration restructures the relationship between offers and projects:
-- - Offers gain execution fields (manager, team, spent, invoiced, health, etc.)
-- - Offers gain new phases: 'order' (won and being executed), 'completed'
-- - Offers lose 'won' phase (replaced by 'order')
-- - Projects become lightweight containers, losing economic/execution fields
-- - Projects gain 'on_hold' phase, lose 'active' phase
-- ============================================================================

-- ============================================================================
-- PART 1: DROP DEPENDENT VIEWS
-- ============================================================================
-- Drop all views that depend on columns we're modifying
DROP VIEW IF EXISTS dashboard_metrics_aggregation CASCADE;
DROP VIEW IF EXISTS project_cost_summary CASCADE;
DROP VIEW IF EXISTS v_project_health_summary CASCADE;
DROP VIEW IF EXISTS v_budget_vs_actual CASCADE;
DROP VIEW IF EXISTS v_customer_lifetime_value CASCADE;
DROP VIEW IF EXISTS v_team_performance CASCADE;

-- ============================================================================
-- PART 2: ADD NEW COLUMNS TO OFFERS TABLE
-- ============================================================================

-- Add execution/order phase fields to offers
ALTER TABLE offers ADD COLUMN manager_id VARCHAR(100);
ALTER TABLE offers ADD COLUMN manager_name VARCHAR(200);
ALTER TABLE offers ADD COLUMN team_members TEXT[];
ALTER TABLE offers ADD COLUMN spent DECIMAL(15,2) NOT NULL DEFAULT 0;
ALTER TABLE offers ADD COLUMN invoiced DECIMAL(15,2) NOT NULL DEFAULT 0;
ALTER TABLE offers ADD COLUMN health VARCHAR(20) DEFAULT 'on_track';
ALTER TABLE offers ADD COLUMN completion_percent DECIMAL(5,2) DEFAULT 0;
ALTER TABLE offers ADD COLUMN start_date DATE;
ALTER TABLE offers ADD COLUMN end_date DATE;
ALTER TABLE offers ADD COLUMN estimated_completion_date DATE;

-- Add order_reserve as a generated column (value - invoiced)
ALTER TABLE offers ADD COLUMN order_reserve DECIMAL(15,2) GENERATED ALWAYS AS (value - invoiced) STORED;

-- Add comments for new offer columns
COMMENT ON COLUMN offers.manager_id IS 'Project manager user ID for offers in order/completed phase';
COMMENT ON COLUMN offers.manager_name IS 'Denormalized manager name for display';
COMMENT ON COLUMN offers.team_members IS 'Array of team member user IDs';
COMMENT ON COLUMN offers.spent IS 'Amount spent on execution (actual costs)';
COMMENT ON COLUMN offers.invoiced IS 'Amount invoiced to customer (hittil fakturert)';
COMMENT ON COLUMN offers.order_reserve IS 'Generated column: value - invoiced (remaining to invoice)';
COMMENT ON COLUMN offers.health IS 'Health status for order phase: on_track, at_risk, delayed, over_budget';
COMMENT ON COLUMN offers.completion_percent IS 'Percentage of work completed (0-100)';
COMMENT ON COLUMN offers.start_date IS 'Actual start date of work';
COMMENT ON COLUMN offers.end_date IS 'Actual or planned end date of work';
COMMENT ON COLUMN offers.estimated_completion_date IS 'Estimated completion date';

-- Index for order_reserve will be created after enum type change (see below)

-- ============================================================================
-- PART 3: MODIFY OFFER PHASE ENUM
-- ============================================================================
-- Current phases: draft, in_progress, sent, won, lost, expired
-- New phases: draft, in_progress, sent, order, completed, lost, expired
-- Changes: Add 'order', 'completed'. Remove 'won'.

-- First, migrate any 'won' offers to 'order' (must happen before enum change)
UPDATE offers SET phase = 'sent' WHERE phase = 'won';

-- PostgreSQL doesn't allow removing enum values directly, so we need to:
-- 1. Create a new enum type with desired values
-- 2. Update the column to use the new type
-- 3. Drop the old type

-- Create new offer_phase enum type
CREATE TYPE offer_phase_new AS ENUM ('draft', 'in_progress', 'sent', 'order', 'completed', 'lost', 'expired');

-- Update offers table to use new enum
-- Note: We cast through text since direct enum-to-enum cast isn't possible
ALTER TABLE offers
    ALTER COLUMN phase TYPE offer_phase_new
    USING (phase::text::offer_phase_new);

-- No old enum type to drop - offers.phase was VARCHAR(50), not an enum type

-- Create simple index for order_reserve (without partial index due to enum immutability issues)
CREATE INDEX idx_offers_order_reserve ON offers (order_reserve);

-- ============================================================================
-- PART 4: DROP COLUMNS FROM PROJECTS TABLE
-- ============================================================================

-- First, drop the trigger that depends on value/cost columns
DROP TRIGGER IF EXISTS trigger_calculate_project_margin_percent ON projects;

-- Drop the generated column first (depends on value and invoiced)
ALTER TABLE projects DROP COLUMN IF EXISTS order_reserve;

-- Drop the index that references order_reserve
DROP INDEX IF EXISTS idx_projects_order_reserve;

-- Drop economic/financial columns
ALTER TABLE projects DROP COLUMN IF EXISTS value;
ALTER TABLE projects DROP COLUMN IF EXISTS cost;
ALTER TABLE projects DROP COLUMN IF EXISTS margin_percent;
ALTER TABLE projects DROP COLUMN IF EXISTS spent;
ALTER TABLE projects DROP COLUMN IF EXISTS invoiced;
ALTER TABLE projects DROP COLUMN IF EXISTS calculated_offer_value;

-- Drop management columns
ALTER TABLE projects DROP COLUMN IF EXISTS manager_id;
ALTER TABLE projects DROP COLUMN IF EXISTS manager_name;
ALTER TABLE projects DROP COLUMN IF EXISTS team_members;

-- Drop status/tracking columns
ALTER TABLE projects DROP COLUMN IF EXISTS health;
ALTER TABLE projects DROP COLUMN IF EXISTS completion_percent;
ALTER TABLE projects DROP COLUMN IF EXISTS estimated_completion_date;

-- Drop offer relationship columns
ALTER TABLE projects DROP COLUMN IF EXISTS has_detailed_budget;
ALTER TABLE projects DROP COLUMN IF EXISTS offer_id;
ALTER TABLE projects DROP COLUMN IF EXISTS winning_offer_id;
ALTER TABLE projects DROP COLUMN IF EXISTS inherited_offer_number;
ALTER TABLE projects DROP COLUMN IF EXISTS won_at;

-- Drop company_id column (offers handle company assignment)
ALTER TABLE projects DROP COLUMN IF EXISTS company_id;

-- Drop indexes that reference removed columns
DROP INDEX IF EXISTS idx_projects_company_id;
DROP INDEX IF EXISTS idx_projects_offer_id;
DROP INDEX IF EXISTS idx_projects_winning_offer_id;

-- ============================================================================
-- PART 5: MAKE PROJECT CUSTOMER_ID NULLABLE
-- ============================================================================
ALTER TABLE projects ALTER COLUMN customer_id DROP NOT NULL;

-- ============================================================================
-- PART 6: MODIFY PROJECT PHASE ENUM
-- ============================================================================
-- Current phases: tilbud, working, active, completed, cancelled
-- New phases: tilbud, working, on_hold, completed, cancelled
-- Changes: Remove 'active', Add 'on_hold'

-- First, migrate any 'active' projects to 'working'
UPDATE projects SET phase = 'working' WHERE phase = 'active';

-- Create new project_phase enum type
CREATE TYPE project_phase_new AS ENUM ('tilbud', 'working', 'on_hold', 'completed', 'cancelled');

-- Drop the default before changing type (PostgreSQL can't auto-cast defaults)
ALTER TABLE projects ALTER COLUMN phase DROP DEFAULT;

-- Update projects table to use new enum
ALTER TABLE projects
    ALTER COLUMN phase TYPE project_phase_new
    USING (phase::text::project_phase_new);

-- Drop old enum type and rename new one
DROP TYPE IF EXISTS project_phase;
ALTER TYPE project_phase_new RENAME TO project_phase;

-- Restore the default
ALTER TABLE projects ALTER COLUMN phase SET DEFAULT 'tilbud';

-- ============================================================================
-- PART 7: RECREATE VIEWS WITH NEW SCHEMA
-- ============================================================================

-- Dashboard Metrics Aggregation View (updated for new offer phases)
CREATE OR REPLACE VIEW dashboard_metrics_aggregation AS
WITH
-- Get best (highest value) offer per project per phase per company
project_best_offers AS (
    SELECT
        o.company_id,
        o.phase::text AS phase,
        o.project_id,
        MAX(o.value) AS best_value,
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
-- Get orphan offers (no project_id)
orphan_offers AS (
    SELECT
        o.company_id,
        o.phase::text AS phase,
        o.value,
        o.probability,
        1 AS offer_count
    FROM offers o
    WHERE o.project_id IS NULL
      AND o.phase NOT IN ('draft', 'expired')
),
-- Combine project best offers and orphan offers
combined_metrics AS (
    SELECT
        company_id,
        phase,
        project_id,
        best_value AS value,
        best_probability AS probability,
        offer_count,
        1 AS project_count
    FROM project_best_offers
    UNION ALL
    SELECT
        company_id,
        phase,
        NULL AS project_id,
        value,
        probability,
        offer_count,
        0 AS project_count
    FROM orphan_offers
)
SELECT
    company_id,
    phase,
    SUM(project_count) AS project_count,
    SUM(offer_count) AS offer_count,
    SUM(value) AS total_value,
    SUM(value * COALESCE(probability, 0) / 100.0) AS weighted_value,
    NOW() AS computed_at
FROM combined_metrics
GROUP BY company_id, phase;

COMMENT ON VIEW dashboard_metrics_aggregation IS
'Pre-computed dashboard metrics. For projects with multiple offers, only the highest value offer per phase is counted.';

-- Project Health Summary View (simplified - projects no longer have economic fields)
CREATE OR REPLACE VIEW v_project_health_summary AS
SELECT
    o.company_id,
    co.name AS company_name,
    o.phase::text AS phase,
    o.health,
    COUNT(o.id) AS offer_count,
    SUM(o.value) AS total_value,
    SUM(o.cost) AS total_cost,
    SUM(o.spent) AS total_spent,
    SUM(o.value - o.spent) AS total_remaining,
    AVG(o.completion_percent) AS avg_completion_percent,
    COUNT(CASE WHEN o.health = 'on_track' THEN 1 END) AS on_track_count,
    COUNT(CASE WHEN o.health = 'at_risk' THEN 1 END) AS at_risk_count,
    COUNT(CASE WHEN o.health = 'delayed' THEN 1 END) AS delayed_count,
    COUNT(CASE WHEN o.health = 'over_budget' THEN 1 END) AS over_budget_count,
    COUNT(CASE WHEN o.end_date < CURRENT_DATE AND o.phase NOT IN ('completed', 'lost', 'expired') THEN 1 END) AS overdue_count
FROM offers o
JOIN companies co ON o.company_id = co.id
WHERE o.phase IN ('order', 'completed')
GROUP BY o.company_id, co.name, o.phase, o.health;

-- Budget vs Actual View (now based on offers instead of projects)
CREATE OR REPLACE VIEW v_budget_vs_actual AS
SELECT
    o.id AS offer_id,
    o.title AS offer_title,
    o.company_id,
    co.name AS company_name,
    o.customer_id,
    cu.name AS customer_name,
    o.phase::text AS offer_phase,
    o.health AS offer_health,
    o.value AS total_value,
    o.cost AS offer_cost,
    o.margin_percent AS offer_margin_percent,
    o.spent AS total_spent,
    o.invoiced,
    o.order_reserve,
    COALESCE(bi.total_planned_cost, 0::numeric) AS planned_cost,
    COALESCE(bi.total_planned_revenue, 0::numeric) AS planned_revenue,
    COALESCE(bi.item_count, 0::bigint) AS budget_line_count,
    o.value - o.spent AS value_variance,
    CASE
        WHEN o.value > 0::numeric THEN (o.value - o.spent) / o.value * 100::numeric
        ELSE 0::numeric
    END AS value_variance_percent,
    CASE
        WHEN COALESCE(bi.total_planned_revenue, 0::numeric) > 0::numeric
        THEN (COALESCE(bi.total_planned_revenue, 0::numeric) - COALESCE(bi.total_planned_cost, 0::numeric)) / COALESCE(bi.total_planned_revenue, 0::numeric) * 100::numeric
        ELSE 0::numeric
    END AS planned_margin_percent,
    o.start_date,
    o.end_date,
    o.completion_percent
FROM offers o
JOIN companies co ON o.company_id = co.id
LEFT JOIN customers cu ON o.customer_id = cu.id
LEFT JOIN (
    SELECT parent_id, COUNT(*) AS item_count, SUM(expected_cost) AS total_planned_cost, SUM(expected_revenue) AS total_planned_revenue
    FROM budget_items WHERE parent_type = 'offer' GROUP BY parent_id
) bi ON bi.parent_id = o.id
WHERE o.phase IN ('order', 'completed');

-- Customer Lifetime Value View (updated for new structure)
CREATE OR REPLACE VIEW v_customer_lifetime_value AS
SELECT
    cu.id AS customer_id,
    cu.name AS customer_name,
    cu.company_id,
    co.name AS company_name,
    cu.org_number,
    cu.created_at AS customer_since,
    COALESCE(dm.total_deals, 0::bigint) AS total_deals,
    COALESCE(dm.won_deals, 0::bigint) AS won_deals,
    COALESCE(dm.lost_deals, 0::bigint) AS lost_deals,
    COALESCE(dm.active_deals, 0::bigint) AS active_deals,
    COALESCE(dm.total_deal_value, 0::numeric) AS total_deal_value,
    COALESCE(dm.won_deal_value, 0::numeric) AS won_deal_value,
    CASE
        WHEN COALESCE(dm.total_deals, 0::bigint) > 0 THEN COALESCE(dm.won_deals, 0::bigint)::numeric / dm.total_deals::numeric * 100::numeric
        ELSE 0::numeric
    END AS win_rate_percent,
    COALESCE(om.total_offers, 0::bigint) AS total_offers,
    COALESCE(om.active_offers, 0::bigint) AS active_offers,
    COALESCE(om.completed_offers, 0::bigint) AS completed_offers,
    COALESCE(om.total_offer_value, 0::numeric) AS total_offer_value,
    COALESCE(om.total_offer_spent, 0::numeric) AS total_offer_spent,
    COALESCE(dm.won_deal_value, 0::numeric) + COALESCE(om.total_offer_value, 0::numeric) AS lifetime_value,
    COALESCE(am.total_activities, 0::bigint) AS total_activities,
    am.last_activity_date,
    CASE
        WHEN COALESCE(am.total_activities, 0::bigint) >= 20 AND COALESCE(dm.active_deals, 0::bigint) > 0 THEN 'high'
        WHEN COALESCE(am.total_activities, 0::bigint) >= 5 OR COALESCE(dm.active_deals, 0::bigint) > 0 THEN 'medium'
        ELSE 'low'
    END AS engagement_level
FROM customers cu
LEFT JOIN companies co ON cu.company_id = co.id
LEFT JOIN (
    SELECT customer_id,
        COUNT(*) AS total_deals,
        COUNT(CASE WHEN stage = 'won' THEN 1 END) AS won_deals,
        COUNT(CASE WHEN stage = 'lost' THEN 1 END) AS lost_deals,
        COUNT(CASE WHEN stage NOT IN ('won', 'lost') THEN 1 END) AS active_deals,
        SUM(value) AS total_deal_value,
        SUM(CASE WHEN stage = 'won' THEN value ELSE 0 END) AS won_deal_value
    FROM deals GROUP BY customer_id
) dm ON dm.customer_id = cu.id
LEFT JOIN (
    SELECT customer_id,
        COUNT(*) AS total_offers,
        COUNT(CASE WHEN phase IN ('order', 'in_progress', 'sent') THEN 1 END) AS active_offers,
        COUNT(CASE WHEN phase = 'completed' THEN 1 END) AS completed_offers,
        SUM(value) AS total_offer_value,
        SUM(spent) AS total_offer_spent
    FROM offers GROUP BY customer_id
) om ON om.customer_id = cu.id
LEFT JOIN (
    SELECT target_id AS customer_id, COUNT(*) AS total_activities, MAX(occurred_at) AS last_activity_date
    FROM activities WHERE target_type = 'Customer' GROUP BY target_id
) am ON am.customer_id = cu.id;

-- Team Performance View (now based on offer managers)
CREATE OR REPLACE VIEW v_team_performance AS
SELECT
    o.manager_id,
    COALESCE(u.name, o.manager_name, 'Unknown') AS manager_name,
    o.company_id,
    co.name AS company_name,
    COUNT(o.id) AS total_offers,
    COUNT(CASE WHEN o.phase = 'order' THEN 1 END) AS active_offers,
    COUNT(CASE WHEN o.phase = 'completed' THEN 1 END) AS completed_offers,
    COUNT(CASE WHEN o.phase = 'lost' THEN 1 END) AS lost_offers,
    SUM(o.value) AS total_value,
    SUM(o.cost) AS total_cost,
    SUM(o.spent) AS total_spent,
    AVG(o.completion_percent) AS avg_completion,
    COUNT(CASE WHEN o.health = 'on_track' THEN 1 END) AS on_track_count,
    COUNT(CASE WHEN o.health IN ('at_risk', 'delayed', 'over_budget') THEN 1 END) AS problematic_count,
    CASE
        WHEN COUNT(o.id) > 0 THEN COUNT(CASE WHEN o.health = 'on_track' THEN 1 END)::numeric / COUNT(o.id)::numeric * 100
        ELSE 0
    END AS health_rate
FROM offers o
JOIN companies co ON o.company_id = co.id
LEFT JOIN users u ON o.manager_id = u.id
WHERE o.phase IN ('order', 'completed')
GROUP BY o.manager_id, u.name, o.manager_name, o.company_id, co.name;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- ============================================================================
-- ROLLBACK: RESTORE PREVIOUS SCHEMA
-- ============================================================================

-- Drop new views
DROP VIEW IF EXISTS dashboard_metrics_aggregation CASCADE;
DROP VIEW IF EXISTS project_cost_summary CASCADE;
DROP VIEW IF EXISTS v_project_health_summary CASCADE;
DROP VIEW IF EXISTS v_budget_vs_actual CASCADE;
DROP VIEW IF EXISTS v_customer_lifetime_value CASCADE;
DROP VIEW IF EXISTS v_team_performance CASCADE;

-- ============================================================================
-- ROLLBACK PART 6: RESTORE PROJECT PHASE ENUM
-- ============================================================================

-- Migrate on_hold projects back to working
UPDATE projects SET phase = 'working' WHERE phase::text = 'on_hold';

-- Create old project_phase enum
CREATE TYPE project_phase_old AS ENUM ('tilbud', 'working', 'active', 'completed', 'cancelled');

-- Update projects table
ALTER TABLE projects
    ALTER COLUMN phase TYPE project_phase_old
    USING (phase::text::project_phase_old);

-- Drop new enum and rename old
DROP TYPE IF EXISTS project_phase;
ALTER TYPE project_phase_old RENAME TO project_phase;

-- ============================================================================
-- ROLLBACK PART 5: MAKE PROJECT CUSTOMER_ID NOT NULL
-- ============================================================================
-- Note: This will fail if there are projects with NULL customer_id
-- First set a default or handle NULL values
UPDATE projects SET customer_id = (SELECT id FROM customers LIMIT 1) WHERE customer_id IS NULL;
ALTER TABLE projects ALTER COLUMN customer_id SET NOT NULL;

-- ============================================================================
-- ROLLBACK PART 4: RESTORE COLUMNS TO PROJECTS TABLE
-- ============================================================================

-- Re-add company_id column
ALTER TABLE projects ADD COLUMN company_id VARCHAR(50);
-- Attempt to populate from related offers
UPDATE projects p SET company_id = (
    SELECT o.company_id FROM offers o WHERE o.project_id = p.id LIMIT 1
);
-- Set default for any remaining nulls
UPDATE projects SET company_id = 'gruppen' WHERE company_id IS NULL;
ALTER TABLE projects ALTER COLUMN company_id SET NOT NULL;
CREATE INDEX idx_projects_company_id ON projects(company_id);

-- Re-add economic columns
ALTER TABLE projects ADD COLUMN value DECIMAL(15,2) NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN cost DECIMAL(15,2) NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN margin_percent DECIMAL(8,4) NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN spent DECIMAL(15,2) NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN invoiced DECIMAL(15,2) NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN calculated_offer_value DECIMAL(15,2) DEFAULT 0;

-- Re-add management columns
ALTER TABLE projects ADD COLUMN manager_id VARCHAR(100);
ALTER TABLE projects ADD COLUMN manager_name VARCHAR(200);
ALTER TABLE projects ADD COLUMN team_members TEXT[];

-- Re-add status/tracking columns
ALTER TABLE projects ADD COLUMN health project_health DEFAULT 'on_track';
ALTER TABLE projects ADD COLUMN completion_percent DECIMAL(5,2) DEFAULT 0;
ALTER TABLE projects ADD COLUMN estimated_completion_date DATE;

-- Re-add offer relationship columns
ALTER TABLE projects ADD COLUMN has_detailed_budget BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE projects ADD COLUMN offer_id UUID REFERENCES offers(id) ON DELETE SET NULL;
ALTER TABLE projects ADD COLUMN winning_offer_id UUID REFERENCES offers(id) ON DELETE SET NULL;
ALTER TABLE projects ADD COLUMN inherited_offer_number VARCHAR(50);
ALTER TABLE projects ADD COLUMN won_at TIMESTAMP;

-- Re-create indexes
CREATE INDEX idx_projects_offer_id ON projects(offer_id);
CREATE INDEX idx_projects_winning_offer_id ON projects(winning_offer_id);

-- Re-add generated column
ALTER TABLE projects ADD COLUMN order_reserve DECIMAL(15,2) GENERATED ALWAYS AS (value - invoiced) STORED;
CREATE INDEX idx_projects_order_reserve ON projects (order_reserve) WHERE phase IN ('working', 'active');

-- Re-create margin calculation trigger
CREATE OR REPLACE FUNCTION calculate_project_margin_percent()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.value > 0 THEN
        NEW.margin_percent := ((NEW.value - COALESCE(NEW.cost, 0)) / NEW.value) * 100;
    ELSE
        NEW.margin_percent := 0;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_calculate_project_margin_percent
    BEFORE INSERT OR UPDATE OF value, cost ON projects
    FOR EACH ROW
    EXECUTE FUNCTION calculate_project_margin_percent();

-- ============================================================================
-- ROLLBACK PART 3: RESTORE OFFER PHASE (convert order/completed back)
-- ============================================================================

-- Migrate order -> sent (closest equivalent), completed -> sent
UPDATE offers SET phase = 'sent' WHERE phase::text IN ('order', 'completed');

-- Note: Cannot easily restore 'won' phase since we don't know which 'sent' offers were originally 'won'
-- The phase column will remain VARCHAR(50) - no enum type change needed for rollback

-- Drop the new enum type and convert back to VARCHAR
ALTER TABLE offers ALTER COLUMN phase TYPE VARCHAR(50) USING (phase::text);
DROP TYPE IF EXISTS offer_phase_new;

-- ============================================================================
-- ROLLBACK PART 2: REMOVE NEW COLUMNS FROM OFFERS TABLE
-- ============================================================================

DROP INDEX IF EXISTS idx_offers_order_reserve;

ALTER TABLE offers DROP COLUMN IF EXISTS order_reserve;
ALTER TABLE offers DROP COLUMN IF EXISTS estimated_completion_date;
ALTER TABLE offers DROP COLUMN IF EXISTS end_date;
ALTER TABLE offers DROP COLUMN IF EXISTS start_date;
ALTER TABLE offers DROP COLUMN IF EXISTS completion_percent;
ALTER TABLE offers DROP COLUMN IF EXISTS health;
ALTER TABLE offers DROP COLUMN IF EXISTS invoiced;
ALTER TABLE offers DROP COLUMN IF EXISTS spent;
ALTER TABLE offers DROP COLUMN IF EXISTS team_members;
ALTER TABLE offers DROP COLUMN IF EXISTS manager_name;
ALTER TABLE offers DROP COLUMN IF EXISTS manager_id;

-- ============================================================================
-- ROLLBACK PART 7: RECREATE ORIGINAL VIEWS
-- ============================================================================

-- Dashboard Metrics Aggregation View
CREATE OR REPLACE VIEW dashboard_metrics_aggregation AS
WITH
project_best_offers AS (
    SELECT
        o.company_id,
        o.phase,
        o.project_id,
        MAX(o.value) AS best_value,
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
combined_metrics AS (
    SELECT company_id, phase, project_id, best_value AS value, best_probability AS probability, offer_count, 1 AS project_count
    FROM project_best_offers
    UNION ALL
    SELECT company_id, phase, NULL AS project_id, value, probability, offer_count, 0 AS project_count
    FROM orphan_offers
)
SELECT
    company_id,
    phase,
    SUM(project_count) AS project_count,
    SUM(offer_count) AS offer_count,
    SUM(value) AS total_value,
    SUM(value * COALESCE(probability, 0) / 100.0) AS weighted_value,
    NOW() AS computed_at
FROM combined_metrics
GROUP BY company_id, phase;

-- Project Health Summary View
CREATE OR REPLACE VIEW v_project_health_summary AS
SELECT
    p.company_id,
    co.name AS company_name,
    p.phase::text AS phase,
    p.health,
    COUNT(p.id) AS project_count,
    SUM(p.value) AS total_value,
    SUM(p.cost) AS total_cost,
    SUM(p.spent) AS total_spent,
    SUM(p.value - p.spent) AS total_remaining,
    AVG(p.completion_percent) AS avg_completion_percent,
    COUNT(CASE WHEN p.health = 'on_track'::project_health THEN 1 ELSE NULL END) AS on_track_count,
    COUNT(CASE WHEN p.health = 'at_risk'::project_health THEN 1 ELSE NULL END) AS at_risk_count,
    COUNT(CASE WHEN p.health = 'delayed'::project_health THEN 1 ELSE NULL END) AS delayed_count,
    COUNT(CASE WHEN p.health = 'over_budget'::project_health THEN 1 ELSE NULL END) AS over_budget_count,
    COUNT(CASE WHEN p.end_date < CURRENT_DATE AND p.phase NOT IN ('completed', 'cancelled') THEN 1 ELSE NULL END) AS overdue_count
FROM projects p
JOIN companies co ON p.company_id = co.id
GROUP BY p.company_id, co.name, p.phase, p.health;

-- Budget vs Actual View
CREATE OR REPLACE VIEW v_budget_vs_actual AS
SELECT
    p.id AS project_id,
    p.name AS project_name,
    p.company_id,
    co.name AS company_name,
    p.customer_id,
    cu.name AS customer_name,
    p.phase::text AS project_phase,
    p.health AS project_health,
    p.value AS total_value,
    p.cost AS project_cost,
    p.margin_percent AS project_margin_percent,
    p.spent AS legacy_spent,
    COALESCE(bd.total_planned_cost, 0::numeric) AS planned_cost,
    COALESCE(bd.total_planned_revenue, 0::numeric) AS planned_revenue,
    COALESCE(bd.dimension_count, 0::bigint) AS budget_line_count,
    COALESCE(ac.total_actual_cost, 0::numeric) AS actual_cost,
    COALESCE(ac.cost_entry_count, 0::bigint) AS cost_entry_count,
    p.value - COALESCE(ac.total_actual_cost, 0::numeric) AS value_variance,
    CASE WHEN p.value > 0::numeric THEN (p.value - COALESCE(ac.total_actual_cost, 0::numeric)) / p.value * 100::numeric ELSE 0::numeric END AS value_variance_percent,
    COALESCE(bd.total_planned_cost, 0::numeric) - COALESCE(ac.total_actual_cost, 0::numeric) AS planned_vs_actual_variance,
    CASE WHEN COALESCE(bd.total_planned_revenue, 0::numeric) > 0::numeric THEN (COALESCE(bd.total_planned_revenue, 0::numeric) - COALESCE(bd.total_planned_cost, 0::numeric)) / COALESCE(bd.total_planned_revenue, 0::numeric) * 100::numeric ELSE 0::numeric END AS planned_margin_percent,
    CASE WHEN COALESCE(bd.total_planned_revenue, 0::numeric) > 0::numeric THEN (COALESCE(bd.total_planned_revenue, 0::numeric) - COALESCE(ac.total_actual_cost, 0::numeric)) / COALESCE(bd.total_planned_revenue, 0::numeric) * 100::numeric ELSE 0::numeric END AS actual_margin_percent,
    p.start_date,
    p.end_date,
    p.completion_percent
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN customers cu ON p.customer_id = cu.id
LEFT JOIN (
    SELECT parent_id, COUNT(*) AS dimension_count, SUM(cost) AS total_planned_cost, SUM(revenue) AS total_planned_revenue
    FROM budget_dimensions WHERE parent_type = 'project' GROUP BY parent_id
) bd ON bd.parent_id = p.id
LEFT JOIN (
    SELECT project_id, COUNT(*) AS cost_entry_count, SUM(amount) AS total_actual_cost
    FROM project_actual_costs GROUP BY project_id
) ac ON ac.project_id = p.id;

-- Customer Lifetime Value View
CREATE OR REPLACE VIEW v_customer_lifetime_value AS
SELECT
    cu.id AS customer_id,
    cu.name AS customer_name,
    cu.company_id,
    co.name AS company_name,
    cu.org_number,
    cu.created_at AS customer_since,
    COALESCE(dm.total_deals, 0::bigint) AS total_deals,
    COALESCE(dm.won_deals, 0::bigint) AS won_deals,
    COALESCE(dm.lost_deals, 0::bigint) AS lost_deals,
    COALESCE(dm.active_deals, 0::bigint) AS active_deals,
    COALESCE(dm.total_deal_value, 0::numeric) AS total_deal_value,
    COALESCE(dm.won_deal_value, 0::numeric) AS won_deal_value,
    CASE WHEN COALESCE(dm.total_deals, 0::bigint) > 0 THEN COALESCE(dm.won_deals, 0::bigint)::numeric / dm.total_deals::numeric * 100::numeric ELSE 0::numeric END AS win_rate_percent,
    COALESCE(pm.total_projects, 0::bigint) AS total_projects,
    COALESCE(pm.active_projects, 0::bigint) AS active_projects,
    COALESCE(pm.completed_projects, 0::bigint) AS completed_projects,
    COALESCE(pm.total_project_value, 0::numeric) AS total_project_value,
    COALESCE(pm.total_project_spent, 0::numeric) AS total_project_spent,
    COALESCE(dm.won_deal_value, 0::numeric) + COALESCE(pm.total_project_value, 0::numeric) AS lifetime_value,
    COALESCE(am.total_activities, 0::bigint) AS total_activities,
    am.last_activity_date,
    CASE
        WHEN COALESCE(am.total_activities, 0::bigint) >= 20 AND COALESCE(dm.active_deals, 0::bigint) > 0 THEN 'high'
        WHEN COALESCE(am.total_activities, 0::bigint) >= 5 OR COALESCE(dm.active_deals, 0::bigint) > 0 THEN 'medium'
        ELSE 'low'
    END AS engagement_level
FROM customers cu
LEFT JOIN companies co ON cu.company_id = co.id
LEFT JOIN (
    SELECT customer_id, COUNT(*) AS total_deals, COUNT(CASE WHEN stage = 'won' THEN 1 END) AS won_deals, COUNT(CASE WHEN stage = 'lost' THEN 1 END) AS lost_deals, COUNT(CASE WHEN stage NOT IN ('won', 'lost') THEN 1 END) AS active_deals, SUM(value) AS total_deal_value, SUM(CASE WHEN stage = 'won' THEN value ELSE 0 END) AS won_deal_value
    FROM deals GROUP BY customer_id
) dm ON dm.customer_id = cu.id
LEFT JOIN (
    SELECT customer_id, COUNT(*) AS total_projects, COUNT(CASE WHEN phase = 'working' THEN 1 END) AS active_projects, COUNT(CASE WHEN phase = 'completed' THEN 1 END) AS completed_projects, SUM(value) AS total_project_value, SUM(spent) AS total_project_spent
    FROM projects GROUP BY customer_id
) pm ON pm.customer_id = cu.id
LEFT JOIN (
    SELECT target_id AS customer_id, COUNT(*) AS total_activities, MAX(occurred_at) AS last_activity_date
    FROM activities WHERE target_type = 'Customer' GROUP BY target_id
) am ON am.customer_id = cu.id;

-- Team Performance View
CREATE OR REPLACE VIEW v_team_performance AS
SELECT
    p.manager_id,
    COALESCE(u.name, p.manager_name, 'Unknown') AS manager_name,
    p.company_id,
    co.name AS company_name,
    COUNT(p.id) AS total_projects,
    COUNT(CASE WHEN p.phase = 'working' THEN 1 END) AS active_projects,
    COUNT(CASE WHEN p.phase = 'completed' THEN 1 END) AS completed_projects,
    COUNT(CASE WHEN p.phase = 'cancelled' THEN 1 END) AS cancelled_projects,
    SUM(p.value) AS total_value,
    SUM(p.cost) AS total_cost,
    SUM(p.spent) AS total_spent,
    AVG(p.completion_percent) AS avg_completion,
    COUNT(CASE WHEN p.health = 'on_track' THEN 1 END) AS on_track_count,
    COUNT(CASE WHEN p.health IN ('at_risk', 'delayed', 'over_budget') THEN 1 END) AS problematic_count,
    CASE WHEN COUNT(p.id) > 0 THEN COUNT(CASE WHEN p.health = 'on_track' THEN 1 END)::numeric / COUNT(p.id)::numeric * 100 ELSE 0 END AS health_rate
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN users u ON p.manager_id = u.id
GROUP BY p.manager_id, u.name, p.manager_name, p.company_id, co.name;

-- +goose StatementEnd
