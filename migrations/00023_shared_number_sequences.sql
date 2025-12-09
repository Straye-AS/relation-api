-- +goose Up
-- +goose StatementBegin

-- Rename offer_number_sequences to number_sequences for shared use between offers and projects
-- The sequence is now shared: both offer numbers and project numbers use the same counter per company/year
ALTER TABLE IF EXISTS offer_number_sequences RENAME TO number_sequences;

-- Update existing sequence data to account for any existing project numbers
-- Projects with manual project_number need their sequences accounted for
-- This ensures new auto-generated numbers don't conflict

-- Add index on offers.offer_number for fast lookup
CREATE INDEX IF NOT EXISTS idx_offers_offer_number ON offers(offer_number) WHERE offer_number IS NOT NULL;

-- Add index on projects.project_number for fast lookup
CREATE INDEX IF NOT EXISTS idx_projects_project_number ON projects(project_number) WHERE project_number IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_projects_project_number;
DROP INDEX IF EXISTS idx_offers_offer_number;

-- Rename table back
ALTER TABLE IF EXISTS number_sequences RENAME TO offer_number_sequences;

-- +goose StatementEnd
