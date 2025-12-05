-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- PROJECT HEALTH ENUM
-- ============================================================================

CREATE TYPE project_health AS ENUM (
    'on_track',
    'at_risk',
    'delayed',
    'over_budget'
);

-- ============================================================================
-- ENHANCE PROJECTS TABLE
-- ============================================================================

-- Add deal relationship (a project can be linked to a deal)
ALTER TABLE projects
    ADD COLUMN deal_id UUID REFERENCES deals(id) ON DELETE SET NULL;

-- Add budget tracking fields
ALTER TABLE projects
    ADD COLUMN has_detailed_budget BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN health project_health DEFAULT 'on_track',
    ADD COLUMN completion_percent DECIMAL(5, 2) DEFAULT 0 CHECK (completion_percent >= 0 AND completion_percent <= 100),
    ADD COLUMN estimated_completion_date DATE;

-- Create indexes for new columns
CREATE INDEX idx_projects_deal_id ON projects(deal_id);
CREATE INDEX idx_projects_health ON projects(health);
CREATE INDEX idx_projects_completion ON projects(completion_percent);

-- ============================================================================
-- PROJECT ACTUAL COSTS TABLE (for ERP integration)
-- ============================================================================

-- Create enum for ERP source systems
CREATE TYPE erp_source AS ENUM (
    'manual',
    'tripletex',
    'visma',
    'poweroffice',
    'other'
);

-- Create enum for cost types
CREATE TYPE cost_type AS ENUM (
    'labor',
    'materials',
    'equipment',
    'subcontractor',
    'travel',
    'overhead',
    'other'
);

-- Create project_actual_costs table
CREATE TABLE project_actual_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Project relationship
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Cost details
    cost_type cost_type NOT NULL,
    description VARCHAR(500) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'NOK',

    -- Date tracking
    cost_date DATE NOT NULL,
    posting_date DATE,

    -- Budget dimension link (optional - links to specific budget line)
    budget_dimension_id UUID REFERENCES budget_dimensions(id) ON DELETE SET NULL,

    -- ERP integration fields
    erp_source erp_source NOT NULL DEFAULT 'manual',
    erp_reference VARCHAR(100),       -- External reference ID from ERP
    erp_transaction_id VARCHAR(100),  -- Transaction ID from ERP
    erp_synced_at TIMESTAMP,          -- Last sync timestamp

    -- Approval workflow
    is_approved BOOLEAN NOT NULL DEFAULT false,
    approved_by_id VARCHAR(100),
    approved_at TIMESTAMP,

    -- Notes
    notes TEXT,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for project_actual_costs
CREATE INDEX idx_project_actual_costs_project_id ON project_actual_costs(project_id);
CREATE INDEX idx_project_actual_costs_cost_type ON project_actual_costs(cost_type);
CREATE INDEX idx_project_actual_costs_cost_date ON project_actual_costs(cost_date);
CREATE INDEX idx_project_actual_costs_erp_source ON project_actual_costs(erp_source);
CREATE INDEX idx_project_actual_costs_erp_reference ON project_actual_costs(erp_reference);
CREATE INDEX idx_project_actual_costs_budget_dimension ON project_actual_costs(budget_dimension_id);

-- Composite index for common queries
CREATE INDEX idx_project_actual_costs_project_date ON project_actual_costs(project_id, cost_date DESC);

-- Create trigger for project_actual_costs updated_at
CREATE TRIGGER update_project_actual_costs_updated_at BEFORE UPDATE ON project_actual_costs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- PROJECT COST SUMMARY VIEW
-- ============================================================================

CREATE VIEW project_cost_summary AS
SELECT
    p.id as project_id,
    p.name as project_name,
    p.budget,
    p.spent,
    COALESCE(SUM(pac.amount), 0) as actual_costs,
    p.budget - COALESCE(SUM(pac.amount), 0) as remaining_budget,
    CASE
        WHEN p.budget > 0 THEN (COALESCE(SUM(pac.amount), 0) / p.budget) * 100
        ELSE 0
    END as budget_used_percent,
    COUNT(pac.id) as cost_entry_count
FROM projects p
LEFT JOIN project_actual_costs pac ON pac.project_id = p.id
GROUP BY p.id, p.name, p.budget, p.spent;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop view
DROP VIEW IF EXISTS project_cost_summary;

-- Drop project_actual_costs table
DROP TRIGGER IF EXISTS update_project_actual_costs_updated_at ON project_actual_costs;
DROP TABLE IF EXISTS project_actual_costs;

-- Drop enums
DROP TYPE IF EXISTS cost_type;
DROP TYPE IF EXISTS erp_source;

-- Remove columns from projects
DROP INDEX IF EXISTS idx_projects_completion;
DROP INDEX IF EXISTS idx_projects_health;
DROP INDEX IF EXISTS idx_projects_deal_id;

ALTER TABLE projects
    DROP COLUMN IF EXISTS estimated_completion_date,
    DROP COLUMN IF EXISTS completion_percent,
    DROP COLUMN IF EXISTS health,
    DROP COLUMN IF EXISTS has_detailed_budget,
    DROP COLUMN IF EXISTS deal_id;

-- Drop project health enum
DROP TYPE IF EXISTS project_health;

-- +goose StatementEnd
