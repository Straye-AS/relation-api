-- +goose Up
-- +goose StatementBegin

-- Create budget_items table for flexible budget line items
-- This replaces the category-based budget_dimensions with user-defined items
CREATE TABLE IF NOT EXISTS budget_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_type budget_parent_type NOT NULL,
    parent_id UUID NOT NULL,
    name VARCHAR(200) NOT NULL,
    expected_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    expected_margin DECIMAL(5,2) NOT NULL DEFAULT 0,
    expected_revenue DECIMAL(15,2) GENERATED ALWAYS AS (
        CASE
            WHEN expected_margin >= 100 THEN expected_cost
            ELSE expected_cost / (1 - expected_margin/100)
        END
    ) STORED,
    expected_profit DECIMAL(15,2) GENERATED ALWAYS AS (
        CASE
            WHEN expected_margin >= 100 THEN 0
            ELSE (expected_cost / (1 - expected_margin/100)) - expected_cost
        END
    ) STORED,
    quantity DECIMAL(10,2),
    price_per_item DECIMAL(15,2),
    description TEXT,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient queries
CREATE INDEX idx_budget_items_parent ON budget_items(parent_type, parent_id);
CREATE INDEX idx_budget_items_display_order ON budget_items(parent_type, parent_id, display_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_budget_items_display_order;
DROP INDEX IF EXISTS idx_budget_items_parent;
DROP TABLE IF EXISTS budget_items;

-- +goose StatementEnd
