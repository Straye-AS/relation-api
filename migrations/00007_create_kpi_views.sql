-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- KPI VIEWS FOR DASHBOARD AND ANALYTICS
-- Migration 00007: Create optimized database views for reporting
-- ============================================================================

-- ============================================================================
-- 1. SALES PIPELINE SUMMARY VIEW
-- Aggregates deal data by company and stage for pipeline visualization
-- ============================================================================

CREATE VIEW v_sales_pipeline_summary AS
SELECT
    d.company_id,
    c.name as company_name,
    d.stage,
    COUNT(d.id) as deal_count,
    SUM(d.value) as total_value,
    SUM(d.weighted_value) as total_weighted_value,
    AVG(d.probability) as avg_probability,
    AVG(d.value) as avg_deal_value,
    MIN(d.expected_close_date) as earliest_close_date,
    MAX(d.expected_close_date) as latest_close_date,
    COUNT(CASE WHEN d.expected_close_date < CURRENT_DATE AND d.stage NOT IN ('won', 'lost') THEN 1 END) as overdue_count
FROM deals d
JOIN companies c ON d.company_id = c.id
GROUP BY d.company_id, c.name, d.stage;

-- Create index to support the view
CREATE INDEX idx_deals_pipeline_query ON deals(company_id, stage, expected_close_date);

-- ============================================================================
-- 2. PROJECT HEALTH SUMMARY VIEW
-- Aggregates project status and budget health across companies
-- ============================================================================

CREATE VIEW v_project_health_summary AS
SELECT
    p.company_id,
    co.name as company_name,
    p.status,
    p.health,
    COUNT(p.id) as project_count,
    SUM(p.budget) as total_budget,
    SUM(p.spent) as total_spent,
    SUM(p.budget - p.spent) as total_remaining,
    AVG(p.completion_percent) as avg_completion_percent,
    COUNT(CASE WHEN p.health = 'on_track' THEN 1 END) as on_track_count,
    COUNT(CASE WHEN p.health = 'at_risk' THEN 1 END) as at_risk_count,
    COUNT(CASE WHEN p.health = 'delayed' THEN 1 END) as delayed_count,
    COUNT(CASE WHEN p.health = 'over_budget' THEN 1 END) as over_budget_count,
    COUNT(CASE WHEN p.end_date < CURRENT_DATE AND p.status NOT IN ('completed', 'cancelled') THEN 1 END) as overdue_count
FROM projects p
JOIN companies co ON p.company_id = co.id
GROUP BY p.company_id, co.name, p.status, p.health;

-- ============================================================================
-- 3. BUDGET VS ACTUAL VIEW
-- Compares budgeted amounts with actual costs for projects
-- ============================================================================

CREATE VIEW v_budget_vs_actual AS
SELECT
    p.id as project_id,
    p.name as project_name,
    p.company_id,
    co.name as company_name,
    p.customer_id,
    cu.name as customer_name,
    p.status as project_status,
    p.health as project_health,
    p.budget as total_budget,
    p.spent as legacy_spent,

    -- Budget dimensions (planned)
    COALESCE(bd.total_planned_cost, 0) as planned_cost,
    COALESCE(bd.total_planned_revenue, 0) as planned_revenue,
    COALESCE(bd.dimension_count, 0) as budget_line_count,

    -- Actual costs from ERP/manual entry
    COALESCE(ac.total_actual_cost, 0) as actual_cost,
    COALESCE(ac.cost_entry_count, 0) as cost_entry_count,

    -- Variance calculations
    p.budget - COALESCE(ac.total_actual_cost, 0) as budget_variance,
    CASE
        WHEN p.budget > 0 THEN ((p.budget - COALESCE(ac.total_actual_cost, 0)) / p.budget) * 100
        ELSE 0
    END as budget_variance_percent,

    COALESCE(bd.total_planned_cost, 0) - COALESCE(ac.total_actual_cost, 0) as planned_vs_actual_variance,

    -- Margin analysis
    CASE
        WHEN COALESCE(bd.total_planned_revenue, 0) > 0
        THEN ((COALESCE(bd.total_planned_revenue, 0) - COALESCE(bd.total_planned_cost, 0)) / COALESCE(bd.total_planned_revenue, 0)) * 100
        ELSE 0
    END as planned_margin_percent,

    CASE
        WHEN COALESCE(bd.total_planned_revenue, 0) > 0
        THEN ((COALESCE(bd.total_planned_revenue, 0) - COALESCE(ac.total_actual_cost, 0)) / COALESCE(bd.total_planned_revenue, 0)) * 100
        ELSE 0
    END as actual_margin_percent,

    p.start_date,
    p.end_date,
    p.completion_percent
FROM projects p
JOIN companies co ON p.company_id = co.id
LEFT JOIN customers cu ON p.customer_id = cu.id
LEFT JOIN (
    SELECT
        parent_id,
        COUNT(*) as dimension_count,
        SUM(cost) as total_planned_cost,
        SUM(revenue) as total_planned_revenue
    FROM budget_dimensions
    WHERE parent_type = 'project'
    GROUP BY parent_id
) bd ON bd.parent_id = p.id
LEFT JOIN (
    SELECT
        project_id,
        COUNT(*) as cost_entry_count,
        SUM(amount) as total_actual_cost
    FROM project_actual_costs
    GROUP BY project_id
) ac ON ac.project_id = p.id;

-- ============================================================================
-- 4. CUSTOMER LIFETIME VALUE VIEW
-- Calculates customer value based on deals won and project revenue
-- ============================================================================

CREATE VIEW v_customer_lifetime_value AS
SELECT
    cu.id as customer_id,
    cu.name as customer_name,
    cu.company_id,
    co.name as company_name,
    cu.org_number,
    cu.created_at as customer_since,

    -- Deal metrics
    COALESCE(dm.total_deals, 0) as total_deals,
    COALESCE(dm.won_deals, 0) as won_deals,
    COALESCE(dm.lost_deals, 0) as lost_deals,
    COALESCE(dm.active_deals, 0) as active_deals,
    COALESCE(dm.total_deal_value, 0) as total_deal_value,
    COALESCE(dm.won_deal_value, 0) as won_deal_value,
    CASE
        WHEN COALESCE(dm.total_deals, 0) > 0
        THEN (COALESCE(dm.won_deals, 0)::decimal / dm.total_deals) * 100
        ELSE 0
    END as win_rate_percent,

    -- Project metrics
    COALESCE(pm.total_projects, 0) as total_projects,
    COALESCE(pm.active_projects, 0) as active_projects,
    COALESCE(pm.completed_projects, 0) as completed_projects,
    COALESCE(pm.total_project_budget, 0) as total_project_budget,
    COALESCE(pm.total_project_spent, 0) as total_project_spent,

    -- Lifetime value calculation (won deals + project budgets)
    COALESCE(dm.won_deal_value, 0) + COALESCE(pm.total_project_budget, 0) as lifetime_value,

    -- Activity metrics
    COALESCE(am.total_activities, 0) as total_activities,
    am.last_activity_date,

    -- Engagement score (simple: based on activities and deals)
    CASE
        WHEN COALESCE(am.total_activities, 0) >= 20 AND COALESCE(dm.active_deals, 0) > 0 THEN 'high'
        WHEN COALESCE(am.total_activities, 0) >= 5 OR COALESCE(dm.active_deals, 0) > 0 THEN 'medium'
        ELSE 'low'
    END as engagement_level

FROM customers cu
LEFT JOIN companies co ON cu.company_id = co.id
LEFT JOIN (
    SELECT
        customer_id,
        COUNT(*) as total_deals,
        COUNT(CASE WHEN stage = 'won' THEN 1 END) as won_deals,
        COUNT(CASE WHEN stage = 'lost' THEN 1 END) as lost_deals,
        COUNT(CASE WHEN stage NOT IN ('won', 'lost') THEN 1 END) as active_deals,
        SUM(value) as total_deal_value,
        SUM(CASE WHEN stage = 'won' THEN value ELSE 0 END) as won_deal_value
    FROM deals
    GROUP BY customer_id
) dm ON dm.customer_id = cu.id
LEFT JOIN (
    SELECT
        customer_id,
        COUNT(*) as total_projects,
        COUNT(CASE WHEN status NOT IN ('completed', 'cancelled') THEN 1 END) as active_projects,
        COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_projects,
        SUM(budget) as total_project_budget,
        SUM(spent) as total_project_spent
    FROM projects
    GROUP BY customer_id
) pm ON pm.customer_id = cu.id
LEFT JOIN (
    SELECT
        target_id as customer_id,
        COUNT(*) as total_activities,
        MAX(occurred_at) as last_activity_date
    FROM activities
    WHERE target_type = 'customer'
    GROUP BY target_id
) am ON am.customer_id = cu.id;

-- Create index to support customer lifetime value queries
CREATE INDEX idx_deals_customer_stage ON deals(customer_id, stage);
CREATE INDEX idx_projects_customer_status ON projects(customer_id, status);
CREATE INDEX idx_activities_customer ON activities(target_id) WHERE target_type = 'customer';

-- ============================================================================
-- 5. TEAM PERFORMANCE VIEW
-- Tracks sales and project performance by user/team member
-- ============================================================================

CREATE VIEW v_team_performance AS
SELECT
    u.id as user_id,
    u.name as user_name,
    u.email,
    u.company_id,
    co.name as company_name,

    -- Deal ownership metrics
    COALESCE(dm.owned_deals, 0) as owned_deals,
    COALESCE(dm.won_deals, 0) as won_deals,
    COALESCE(dm.lost_deals, 0) as lost_deals,
    COALESCE(dm.total_deal_value, 0) as total_deal_value,
    COALESCE(dm.won_deal_value, 0) as won_deal_value,
    CASE
        WHEN COALESCE(dm.closed_deals, 0) > 0
        THEN (COALESCE(dm.won_deals, 0)::decimal / dm.closed_deals) * 100
        ELSE 0
    END as win_rate_percent,
    COALESCE(dm.pipeline_value, 0) as pipeline_value,
    COALESCE(dm.weighted_pipeline, 0) as weighted_pipeline,

    -- Project management metrics
    COALESCE(pm.managed_projects, 0) as managed_projects,
    COALESCE(pm.active_projects, 0) as active_projects,
    COALESCE(pm.completed_projects, 0) as completed_projects,
    COALESCE(pm.total_budget_managed, 0) as total_budget_managed,
    COALESCE(pm.on_track_projects, 0) as on_track_projects,
    COALESCE(pm.at_risk_projects, 0) as at_risk_projects,

    -- Activity metrics
    COALESCE(am.total_activities, 0) as total_activities,
    COALESCE(am.activities_this_month, 0) as activities_this_month,
    am.last_activity_date,

    -- Performance indicators
    CASE
        WHEN COALESCE(dm.won_deals, 0) >= 5 AND COALESCE(pm.on_track_projects, 0) > COALESCE(pm.at_risk_projects, 0) THEN 'excellent'
        WHEN COALESCE(dm.won_deals, 0) >= 2 OR COALESCE(pm.managed_projects, 0) >= 3 THEN 'good'
        WHEN COALESCE(am.activities_this_month, 0) >= 10 THEN 'active'
        ELSE 'needs_attention'
    END as performance_indicator

FROM users u
LEFT JOIN companies co ON u.company_id = co.id
LEFT JOIN (
    SELECT
        owner_id,
        COUNT(*) as owned_deals,
        COUNT(CASE WHEN stage = 'won' THEN 1 END) as won_deals,
        COUNT(CASE WHEN stage = 'lost' THEN 1 END) as lost_deals,
        COUNT(CASE WHEN stage IN ('won', 'lost') THEN 1 END) as closed_deals,
        SUM(value) as total_deal_value,
        SUM(CASE WHEN stage = 'won' THEN value ELSE 0 END) as won_deal_value,
        SUM(CASE WHEN stage NOT IN ('won', 'lost') THEN value ELSE 0 END) as pipeline_value,
        SUM(CASE WHEN stage NOT IN ('won', 'lost') THEN weighted_value ELSE 0 END) as weighted_pipeline
    FROM deals
    GROUP BY owner_id
) dm ON dm.owner_id = u.id
LEFT JOIN (
    SELECT
        manager_id,
        COUNT(*) as managed_projects,
        COUNT(CASE WHEN status NOT IN ('completed', 'cancelled') THEN 1 END) as active_projects,
        COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_projects,
        SUM(budget) as total_budget_managed,
        COUNT(CASE WHEN health = 'on_track' THEN 1 END) as on_track_projects,
        COUNT(CASE WHEN health IN ('at_risk', 'delayed', 'over_budget') THEN 1 END) as at_risk_projects
    FROM projects
    GROUP BY manager_id
) pm ON pm.manager_id = u.id
LEFT JOIN (
    SELECT
        creator_id,
        COUNT(*) as total_activities,
        COUNT(CASE WHEN occurred_at >= DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as activities_this_month,
        MAX(occurred_at) as last_activity_date
    FROM activities
    WHERE creator_id IS NOT NULL
    GROUP BY creator_id
) am ON am.creator_id = u.id
WHERE u.is_active = true;

-- Create indexes to support team performance queries
CREATE INDEX idx_deals_owner ON deals(owner_id);
CREATE INDEX idx_projects_manager ON projects(manager_id);
CREATE INDEX idx_activities_creator ON activities(creator_id) WHERE creator_id IS NOT NULL;

-- ============================================================================
-- 6. MONTHLY SALES TRENDS VIEW (Bonus - useful for charts)
-- Aggregates deal data by month for trend analysis
-- ============================================================================

CREATE VIEW v_monthly_sales_trends AS
SELECT
    d.company_id,
    co.name as company_name,
    DATE_TRUNC('month', d.created_at) as month,
    COUNT(*) as deals_created,
    COUNT(CASE WHEN d.stage = 'won' THEN 1 END) as deals_won,
    COUNT(CASE WHEN d.stage = 'lost' THEN 1 END) as deals_lost,
    SUM(d.value) as total_value,
    SUM(CASE WHEN d.stage = 'won' THEN d.value ELSE 0 END) as won_value,
    AVG(d.value) as avg_deal_value,
    AVG(CASE WHEN d.stage = 'won' THEN d.value END) as avg_won_deal_value
FROM deals d
JOIN companies co ON d.company_id = co.id
GROUP BY d.company_id, co.name, DATE_TRUNC('month', d.created_at)
ORDER BY month DESC;

-- Create index for monthly trends
CREATE INDEX idx_deals_created_month ON deals(company_id, DATE_TRUNC('month', created_at));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_deals_created_month;
DROP INDEX IF EXISTS idx_activities_creator;
DROP INDEX IF EXISTS idx_projects_manager;
DROP INDEX IF EXISTS idx_deals_owner;
DROP INDEX IF EXISTS idx_activities_customer;
DROP INDEX IF EXISTS idx_projects_customer_status;
DROP INDEX IF EXISTS idx_deals_customer_stage;
DROP INDEX IF EXISTS idx_deals_pipeline_query;

-- Drop views
DROP VIEW IF EXISTS v_monthly_sales_trends;
DROP VIEW IF EXISTS v_team_performance;
DROP VIEW IF EXISTS v_customer_lifetime_value;
DROP VIEW IF EXISTS v_budget_vs_actual;
DROP VIEW IF EXISTS v_project_health_summary;
DROP VIEW IF EXISTS v_sales_pipeline_summary;

-- +goose StatementEnd
