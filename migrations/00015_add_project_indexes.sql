-- +goose Up
-- +goose StatementBegin

-- Composite index for CountByHealth queries (dashboard performance)
-- Optimizes: SELECT health, COUNT(*) FROM projects WHERE status = ? GROUP BY health
CREATE INDEX IF NOT EXISTS idx_projects_status_health ON projects (status, health);

-- Additional useful indexes for common project queries
CREATE INDEX IF NOT EXISTS idx_projects_company_status ON projects (company_id, status);
CREATE INDEX IF NOT EXISTS idx_projects_manager_status ON projects (manager_id, status);
CREATE INDEX IF NOT EXISTS idx_projects_customer_id ON projects (customer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_projects_customer_id;
DROP INDEX IF EXISTS idx_projects_manager_status;
DROP INDEX IF EXISTS idx_projects_company_status;
DROP INDEX IF EXISTS idx_projects_status_health;

-- +goose StatementEnd
