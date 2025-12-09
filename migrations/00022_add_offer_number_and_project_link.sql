-- +goose Up
-- +goose StatementBegin

-- Add offer_number column for unique per-company offer identification
ALTER TABLE offers ADD COLUMN IF NOT EXISTS offer_number VARCHAR(50);

-- Add project_id to link offers to projects (nullable - offer can exist without project)
ALTER TABLE offers ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL;

-- Create index for project_id foreign key
CREATE INDEX IF NOT EXISTS idx_offers_project_id ON offers(project_id);

-- Create unique constraint on (company_id, offer_number) - offer numbers are unique per company
-- Only applies to non-null offer_numbers (inquiries in draft phase won't have offer_numbers yet)
CREATE UNIQUE INDEX IF NOT EXISTS idx_offers_company_offer_number
    ON offers(company_id, offer_number)
    WHERE offer_number IS NOT NULL;

-- Create sequence table for tracking offer numbers per company per year
CREATE TABLE IF NOT EXISTS offer_number_sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id VARCHAR(50) NOT NULL,
    year INTEGER NOT NULL,
    last_sequence INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, year)
);

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_offer_number_sequences_company_year
    ON offer_number_sequences(company_id, year);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the sequence table
DROP TABLE IF EXISTS offer_number_sequences;

-- Drop indexes
DROP INDEX IF EXISTS idx_offers_company_offer_number;
DROP INDEX IF EXISTS idx_offers_project_id;

-- Drop columns
ALTER TABLE offers DROP COLUMN IF EXISTS project_id;
ALTER TABLE offers DROP COLUMN IF EXISTS offer_number;

-- +goose StatementEnd
