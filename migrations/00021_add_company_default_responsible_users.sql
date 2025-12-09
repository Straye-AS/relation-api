-- +goose Up
-- +goose StatementBegin

-- Add default responsible user columns to companies table
-- These columns store the default users to assign as responsible for new offers and projects

ALTER TABLE companies
    ADD COLUMN default_offer_responsible_id VARCHAR(100) REFERENCES users(id),
    ADD COLUMN default_project_responsible_id VARCHAR(100) REFERENCES users(id);

-- Add indexes for efficient lookups
CREATE INDEX idx_companies_default_offer_responsible ON companies(default_offer_responsible_id);
CREATE INDEX idx_companies_default_project_responsible ON companies(default_project_responsible_id);

-- Add comments for documentation
COMMENT ON COLUMN companies.default_offer_responsible_id IS 'Default user ID to assign as responsible for new offers created under this company';
COMMENT ON COLUMN companies.default_project_responsible_id IS 'Default user ID to assign as manager for new projects created under this company';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS idx_companies_default_project_responsible;
DROP INDEX IF EXISTS idx_companies_default_offer_responsible;

-- Remove columns
ALTER TABLE companies
    DROP COLUMN IF EXISTS default_project_responsible_id,
    DROP COLUMN IF EXISTS default_offer_responsible_id;

-- +goose StatementEnd
