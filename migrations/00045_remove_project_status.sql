-- +goose Up
-- +goose StatementBegin

-- Drop views that depend on projects.status
DROP VIEW IF EXISTS v_project_health_summary CASCADE;
DROP VIEW IF EXISTS v_budget_vs_actual CASCADE;
DROP VIEW IF EXISTS v_customer_lifetime_value CASCADE;
DROP VIEW IF EXISTS v_team_performance CASCADE;

-- Drop index that references status
DROP INDEX IF EXISTS idx_projects_customer_status;

-- Migrate status values to phase
UPDATE projects SET phase = 'working' WHERE status = 'active';
UPDATE projects SET phase = 'cancelled' WHERE status = 'cancelled';
UPDATE projects SET phase = 'completed' WHERE status = 'completed';
-- planning stays as tilbud (already default)
-- on_hold maps to working (if any exist)
UPDATE projects SET phase = 'working' WHERE status = 'on_hold';

-- Drop the status column
ALTER TABLE projects DROP COLUMN IF EXISTS status;

-- Recreate views using phase instead of status
CREATE OR REPLACE VIEW v_project_health_summary AS
SELECT
    p.company_id,
    co.name AS company_name,
    p.phase,
    p.health,
    count(p.id) AS project_count,
    sum(p.value) AS total_value,
    sum(p.cost) AS total_cost,
    sum(p.spent) AS total_spent,
    sum(p.value - p.spent) AS total_remaining,
    avg(p.completion_percent) AS avg_completion_percent,
    count(CASE WHEN p.health = 'on_track'::project_health THEN 1 ELSE NULL END) AS on_track_count,
    count(CASE WHEN p.health = 'at_risk'::project_health THEN 1 ELSE NULL END) AS at_risk_count,
    count(CASE WHEN p.health = 'delayed'::project_health THEN 1 ELSE NULL END) AS delayed_count,
    count(CASE WHEN p.health = 'over_budget'::project_health THEN 1 ELSE NULL END) AS over_budget_count,
    count(CASE WHEN p.end_date < CURRENT_DATE AND p.phase NOT IN ('completed', 'cancelled') THEN 1 ELSE NULL END) AS overdue_count
FROM projects p
JOIN companies co ON p.company_id = co.id
GROUP BY p.company_id, co.name, p.phase, p.health;

CREATE OR REPLACE VIEW v_budget_vs_actual AS
SELECT
    p.id AS project_id,
    p.name AS project_name,
    p.company_id,
    co.name AS company_name,
    p.customer_id,
    cu.name AS customer_name,
    p.phase AS project_phase,
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
    CASE
        WHEN p.value > 0::numeric THEN (p.value - COALESCE(ac.total_actual_cost, 0::numeric)) / p.value * 100::numeric
        ELSE 0::numeric
    END AS value_variance_percent,
    COALESCE(bd.total_planned_cost, 0::numeric) - COALESCE(ac.total_actual_cost, 0::numeric) AS planned_vs_actual_variance,
    CASE
        WHEN COALESCE(bd.total_planned_revenue, 0::numeric) > 0::numeric
        THEN (COALESCE(bd.total_planned_revenue, 0::numeric) - COALESCE(bd.total_planned_cost, 0::numeric)) / COALESCE(bd.total_planned_revenue, 0::numeric) * 100::numeric
        ELSE 0::numeric
    END AS planned_margin_percent,
    CASE
        WHEN COALESCE(bd.total_planned_revenue, 0::numeric) > 0::numeric
        THEN (COALESCE(bd.total_planned_revenue, 0::numeric) - COALESCE(ac.total_actual_cost, 0::numeric)) / COALESCE(bd.total_planned_revenue, 0::numeric) * 100::numeric
        ELSE 0::numeric
    END AS actual_margin_percent,
    p.start_date,
    p.end_date,
    p.completion_percent
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN customers cu ON p.customer_id = cu.id
LEFT JOIN (
    SELECT parent_id, count(*) AS dimension_count, sum(cost) AS total_planned_cost, sum(revenue) AS total_planned_revenue
    FROM budget_dimensions WHERE parent_type = 'project' GROUP BY parent_id
) bd ON bd.parent_id = p.id
LEFT JOIN (
    SELECT project_id, count(*) AS cost_entry_count, sum(amount) AS total_actual_cost
    FROM project_actual_costs GROUP BY project_id
) ac ON ac.project_id = p.id;

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
    SELECT customer_id,
        count(*) AS total_deals,
        count(CASE WHEN stage = 'won' THEN 1 END) AS won_deals,
        count(CASE WHEN stage = 'lost' THEN 1 END) AS lost_deals,
        count(CASE WHEN stage NOT IN ('won', 'lost') THEN 1 END) AS active_deals,
        sum(value) AS total_deal_value,
        sum(CASE WHEN stage = 'won' THEN value ELSE 0 END) AS won_deal_value
    FROM deals GROUP BY customer_id
) dm ON dm.customer_id = cu.id
LEFT JOIN (
    SELECT customer_id,
        count(*) AS total_projects,
        count(CASE WHEN phase = 'working' THEN 1 END) AS active_projects,
        count(CASE WHEN phase = 'completed' THEN 1 END) AS completed_projects,
        sum(value) AS total_project_value,
        sum(spent) AS total_project_spent
    FROM projects GROUP BY customer_id
) pm ON pm.customer_id = cu.id
LEFT JOIN (
    SELECT target_id AS customer_id, count(*) AS total_activities, max(occurred_at) AS last_activity_date
    FROM activities WHERE target_type = 'customer' GROUP BY target_id
) am ON am.customer_id = cu.id;

CREATE OR REPLACE VIEW v_team_performance AS
SELECT
    p.manager_id,
    COALESCE(u.name, p.manager_name, 'Unknown') AS manager_name,
    p.company_id,
    co.name AS company_name,
    count(p.id) AS total_projects,
    count(CASE WHEN p.phase = 'working' THEN 1 END) AS active_projects,
    count(CASE WHEN p.phase = 'completed' THEN 1 END) AS completed_projects,
    count(CASE WHEN p.phase = 'cancelled' THEN 1 END) AS cancelled_projects,
    sum(p.value) AS total_value,
    sum(p.cost) AS total_cost,
    sum(p.spent) AS total_spent,
    avg(p.completion_percent) AS avg_completion,
    count(CASE WHEN p.health = 'on_track' THEN 1 END) AS on_track_count,
    count(CASE WHEN p.health IN ('at_risk', 'delayed', 'over_budget') THEN 1 END) AS problematic_count,
    CASE
        WHEN count(p.id) > 0 THEN count(CASE WHEN p.health = 'on_track' THEN 1 END)::numeric / count(p.id)::numeric * 100
        ELSE 0
    END AS health_rate
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN users u ON p.manager_id = u.id
GROUP BY p.manager_id, u.name, p.manager_name, p.company_id, co.name;

-- Create new index on phase
CREATE INDEX IF NOT EXISTS idx_projects_customer_phase ON projects(customer_id, phase);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop index
DROP INDEX IF EXISTS idx_projects_customer_phase;

-- Drop views
DROP VIEW IF EXISTS v_project_health_summary CASCADE;
DROP VIEW IF EXISTS v_budget_vs_actual CASCADE;
DROP VIEW IF EXISTS v_customer_lifetime_value CASCADE;
DROP VIEW IF EXISTS v_team_performance CASCADE;

-- Re-add status column
ALTER TABLE projects ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'planning';

-- Create index
CREATE INDEX IF NOT EXISTS idx_projects_customer_status ON projects(customer_id, status);

-- Migrate phase back to status
UPDATE projects SET status = 'active' WHERE phase IN ('working', 'active');
UPDATE projects SET status = 'cancelled' WHERE phase = 'cancelled';
UPDATE projects SET status = 'completed' WHERE phase = 'completed';
UPDATE projects SET status = 'planning' WHERE phase = 'tilbud';

-- Recreate views with status
CREATE OR REPLACE VIEW v_project_health_summary AS
SELECT
    p.company_id,
    co.name AS company_name,
    p.status,
    p.health,
    count(p.id) AS project_count,
    sum(p.value) AS total_value,
    sum(p.cost) AS total_cost,
    sum(p.spent) AS total_spent,
    sum(p.value - p.spent) AS total_remaining,
    avg(p.completion_percent) AS avg_completion_percent,
    count(CASE WHEN p.health = 'on_track'::project_health THEN 1 ELSE NULL END) AS on_track_count,
    count(CASE WHEN p.health = 'at_risk'::project_health THEN 1 ELSE NULL END) AS at_risk_count,
    count(CASE WHEN p.health = 'delayed'::project_health THEN 1 ELSE NULL END) AS delayed_count,
    count(CASE WHEN p.health = 'over_budget'::project_health THEN 1 ELSE NULL END) AS over_budget_count,
    count(CASE WHEN p.end_date < CURRENT_DATE AND p.status NOT IN ('completed', 'cancelled') THEN 1 ELSE NULL END) AS overdue_count
FROM projects p
JOIN companies co ON p.company_id = co.id
GROUP BY p.company_id, co.name, p.status, p.health;

CREATE OR REPLACE VIEW v_budget_vs_actual AS
SELECT
    p.id AS project_id,
    p.name AS project_name,
    p.company_id,
    co.name AS company_name,
    p.customer_id,
    cu.name AS customer_name,
    p.status AS project_status,
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
    CASE
        WHEN p.value > 0::numeric THEN (p.value - COALESCE(ac.total_actual_cost, 0::numeric)) / p.value * 100::numeric
        ELSE 0::numeric
    END AS value_variance_percent,
    COALESCE(bd.total_planned_cost, 0::numeric) - COALESCE(ac.total_actual_cost, 0::numeric) AS planned_vs_actual_variance,
    CASE
        WHEN COALESCE(bd.total_planned_revenue, 0::numeric) > 0::numeric
        THEN (COALESCE(bd.total_planned_revenue, 0::numeric) - COALESCE(bd.total_planned_cost, 0::numeric)) / COALESCE(bd.total_planned_revenue, 0::numeric) * 100::numeric
        ELSE 0::numeric
    END AS planned_margin_percent,
    CASE
        WHEN COALESCE(bd.total_planned_revenue, 0::numeric) > 0::numeric
        THEN (COALESCE(bd.total_planned_revenue, 0::numeric) - COALESCE(ac.total_actual_cost, 0::numeric)) / COALESCE(bd.total_planned_revenue, 0::numeric) * 100::numeric
        ELSE 0::numeric
    END AS actual_margin_percent,
    p.start_date,
    p.end_date,
    p.completion_percent
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN customers cu ON p.customer_id = cu.id
LEFT JOIN (
    SELECT parent_id, count(*) AS dimension_count, sum(cost) AS total_planned_cost, sum(revenue) AS total_planned_revenue
    FROM budget_dimensions WHERE parent_type = 'project' GROUP BY parent_id
) bd ON bd.parent_id = p.id
LEFT JOIN (
    SELECT project_id, count(*) AS cost_entry_count, sum(amount) AS total_actual_cost
    FROM project_actual_costs GROUP BY project_id
) ac ON ac.project_id = p.id;

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
    SELECT customer_id,
        count(*) AS total_deals,
        count(CASE WHEN stage = 'won' THEN 1 END) AS won_deals,
        count(CASE WHEN stage = 'lost' THEN 1 END) AS lost_deals,
        count(CASE WHEN stage NOT IN ('won', 'lost') THEN 1 END) AS active_deals,
        sum(value) AS total_deal_value,
        sum(CASE WHEN stage = 'won' THEN value ELSE 0 END) AS won_deal_value
    FROM deals GROUP BY customer_id
) dm ON dm.customer_id = cu.id
LEFT JOIN (
    SELECT customer_id,
        count(*) AS total_projects,
        count(CASE WHEN status = 'active' THEN 1 END) AS active_projects,
        count(CASE WHEN status = 'completed' THEN 1 END) AS completed_projects,
        sum(value) AS total_project_value,
        sum(spent) AS total_project_spent
    FROM projects GROUP BY customer_id
) pm ON pm.customer_id = cu.id
LEFT JOIN (
    SELECT target_id AS customer_id, count(*) AS total_activities, max(occurred_at) AS last_activity_date
    FROM activities WHERE target_type = 'customer' GROUP BY target_id
) am ON am.customer_id = cu.id;

CREATE OR REPLACE VIEW v_team_performance AS
SELECT
    p.manager_id,
    COALESCE(u.name, p.manager_name, 'Unknown') AS manager_name,
    p.company_id,
    co.name AS company_name,
    count(p.id) AS total_projects,
    count(CASE WHEN p.status = 'active' THEN 1 END) AS active_projects,
    count(CASE WHEN p.status = 'completed' THEN 1 END) AS completed_projects,
    count(CASE WHEN p.status = 'cancelled' THEN 1 END) AS cancelled_projects,
    sum(p.value) AS total_value,
    sum(p.cost) AS total_cost,
    sum(p.spent) AS total_spent,
    avg(p.completion_percent) AS avg_completion,
    count(CASE WHEN p.health = 'on_track' THEN 1 END) AS on_track_count,
    count(CASE WHEN p.health IN ('at_risk', 'delayed', 'over_budget') THEN 1 END) AS problematic_count,
    CASE
        WHEN count(p.id) > 0 THEN count(CASE WHEN p.health = 'on_track' THEN 1 END)::numeric / count(p.id)::numeric * 100
        ELSE 0
    END AS health_rate
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN users u ON p.manager_id = u.id
GROUP BY p.manager_id, u.name, p.manager_name, p.company_id, co.name;

-- +goose StatementEnd
