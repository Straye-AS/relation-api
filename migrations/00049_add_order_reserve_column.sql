-- +goose Up
-- Add order_reserve as a generated column (value - invoiced)
ALTER TABLE projects ADD COLUMN order_reserve decimal(15,2) GENERATED ALWAYS AS (value - invoiced) STORED;

-- Create index for efficient filtering/aggregation on active projects
CREATE INDEX idx_projects_order_reserve ON projects (order_reserve) WHERE phase IN ('working', 'active');

-- +goose Down
DROP INDEX IF EXISTS idx_projects_order_reserve;
ALTER TABLE projects DROP COLUMN IF EXISTS order_reserve;
