-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- BUDGET DIMENSION CATEGORIES
-- ============================================================================

-- Create budget dimension categories table for predefined budget line types
CREATE TABLE budget_dimension_categories (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert seed data for standard construction budget categories
INSERT INTO budget_dimension_categories (id, name, description, display_order) VALUES
    ('steel_structure', 'Stålkonstruksjon', 'Steel structure and framework', 1),
    ('hybrid_structure', 'Hybridkonstruksjon', 'Hybrid construction elements', 2),
    ('roofing', 'Tak', 'Roofing materials and installation', 3),
    ('cladding', 'Kledning', 'Wall cladding and facade', 4),
    ('foundation', 'Fundament', 'Foundation work', 5),
    ('assembly', 'Montasje', 'Assembly and installation labor', 6),
    ('transport', 'Transport', 'Transportation and logistics', 7),
    ('engineering', 'Prosjektering', 'Engineering and design', 8),
    ('project_mgmt', 'Prosjektledelse', 'Project management', 9),
    ('crane_rigging', 'Kran og rigg', 'Crane rental and rigging', 10),
    ('miscellaneous', 'Diverse', 'Miscellaneous costs', 11),
    ('contingency', 'Uforutsett', 'Contingency/buffer', 12);

-- Create trigger for budget_dimension_categories updated_at
CREATE TRIGGER update_budget_dimension_categories_updated_at BEFORE UPDATE ON budget_dimension_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- BUDGET DIMENSIONS (Polymorphic - can belong to Offer or Project)
-- ============================================================================

-- Create enum for budget dimension parent type
CREATE TYPE budget_parent_type AS ENUM ('offer', 'project');

-- Create budget_dimensions table
CREATE TABLE budget_dimensions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Polymorphic parent relationship
    parent_type budget_parent_type NOT NULL,
    parent_id UUID NOT NULL,

    -- Category reference (optional - allows custom dimensions)
    category_id VARCHAR(50) REFERENCES budget_dimension_categories(id),

    -- Custom name (used if category_id is null, or for override)
    custom_name VARCHAR(200),

    -- Financial fields
    cost DECIMAL(15, 2) NOT NULL DEFAULT 0,
    revenue DECIMAL(15, 2) NOT NULL DEFAULT 0,

    -- Margin fields
    target_margin_percent DECIMAL(5, 2),  -- If set, revenue is calculated from cost
    margin_override BOOLEAN NOT NULL DEFAULT false,  -- If true, use target_margin_percent to calculate revenue

    -- Calculated margin (stored for query efficiency)
    margin_percent DECIMAL(5, 2) GENERATED ALWAYS AS (
        CASE
            WHEN revenue > 0 THEN ((revenue - cost) / revenue) * 100
            ELSE 0
        END
    ) STORED,

    -- Additional details
    description TEXT,
    quantity DECIMAL(10, 2),
    unit VARCHAR(50),

    -- Ordering within parent
    display_order INT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraint: must have either category_id or custom_name
    CONSTRAINT chk_dimension_name CHECK (
        category_id IS NOT NULL OR custom_name IS NOT NULL
    )
);

-- Create indexes for budget_dimensions
CREATE INDEX idx_budget_dimensions_parent ON budget_dimensions(parent_type, parent_id);
CREATE INDEX idx_budget_dimensions_category ON budget_dimensions(category_id);
CREATE INDEX idx_budget_dimensions_order ON budget_dimensions(parent_type, parent_id, display_order);

-- Create trigger for budget_dimensions updated_at
CREATE TRIGGER update_budget_dimensions_updated_at BEFORE UPDATE ON budget_dimensions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- MARGIN CALCULATION TRIGGER
-- ============================================================================

-- Create function to calculate revenue from cost and target margin
CREATE OR REPLACE FUNCTION calculate_budget_dimension_revenue()
RETURNS TRIGGER AS $$
BEGIN
    -- If margin_override is enabled and target_margin_percent is set,
    -- calculate revenue from cost: revenue = cost / (1 - (margin/100))
    IF NEW.margin_override = true AND NEW.target_margin_percent IS NOT NULL THEN
        -- Prevent division by zero (100% margin would require infinite revenue)
        IF NEW.target_margin_percent >= 100 THEN
            RAISE EXCEPTION 'Target margin percent cannot be 100 or greater';
        END IF;

        NEW.revenue := NEW.cost / (1 - (NEW.target_margin_percent / 100));
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger that fires before insert or update
CREATE TRIGGER trg_calculate_budget_dimension_revenue
    BEFORE INSERT OR UPDATE ON budget_dimensions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_budget_dimension_revenue();

-- ============================================================================
-- DATA MIGRATION FROM OFFER_ITEMS
-- ============================================================================

-- Migrate existing offer_items to budget_dimensions
INSERT INTO budget_dimensions (
    parent_type,
    parent_id,
    custom_name,
    cost,
    revenue,
    description,
    quantity,
    unit,
    display_order,
    created_at,
    updated_at
)
SELECT
    'offer'::budget_parent_type as parent_type,
    offer_id as parent_id,
    discipline as custom_name,
    cost,
    revenue,
    description,
    quantity,
    unit,
    ROW_NUMBER() OVER (PARTITION BY offer_id ORDER BY created_at) as display_order,
    created_at,
    updated_at
FROM offer_items;

-- Note: We keep offer_items table for now to allow rollback
-- It can be dropped in a future migration once budget_dimensions is validated

-- ============================================================================
-- BUDGET SUMMARY VIEW (for convenience)
-- ============================================================================

CREATE VIEW budget_summary AS
SELECT
    parent_type,
    parent_id,
    COUNT(*) as dimension_count,
    SUM(cost) as total_cost,
    SUM(revenue) as total_revenue,
    CASE
        WHEN SUM(revenue) > 0 THEN ((SUM(revenue) - SUM(cost)) / SUM(revenue)) * 100
        ELSE 0
    END as overall_margin_percent,
    SUM(revenue) - SUM(cost) as total_profit
FROM budget_dimensions
GROUP BY parent_type, parent_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop view
DROP VIEW IF EXISTS budget_summary;

-- Drop trigger and function
DROP TRIGGER IF EXISTS trg_calculate_budget_dimension_revenue ON budget_dimensions;
DROP FUNCTION IF EXISTS calculate_budget_dimension_revenue();

-- Drop budget_dimensions table
DROP TRIGGER IF EXISTS update_budget_dimensions_updated_at ON budget_dimensions;
DROP TABLE IF EXISTS budget_dimensions;

-- Drop budget parent type enum
DROP TYPE IF EXISTS budget_parent_type;

-- Drop budget_dimension_categories table
DROP TRIGGER IF EXISTS update_budget_dimension_categories_updated_at ON budget_dimension_categories;
DROP TABLE IF EXISTS budget_dimension_categories;

-- +goose StatementEnd
