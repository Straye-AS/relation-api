-- +goose Up
-- +goose StatementBegin

-- Add company_id column to files table (required for multi-tenant file tracking)
ALTER TABLE files
ADD COLUMN company_id VARCHAR(50);

-- Create index for faster lookups by company
CREATE INDEX idx_files_company_id ON files(company_id);

-- Populate existing files with company_id from their parent entities
-- Files linked to offers: inherit from offer's company_id
UPDATE files f
SET company_id = o.company_id
FROM offers o
WHERE f.offer_id = o.id
AND f.company_id IS NULL;

-- Files linked to offer-supplier relationships: inherit from offer's company_id
UPDATE files f
SET company_id = o.company_id
FROM offer_suppliers os
JOIN offers o ON os.offer_id = o.id
WHERE f.offer_supplier_id = os.id
AND f.company_id IS NULL;

-- For remaining files (customers, projects, suppliers without explicit company),
-- default to 'gruppen' as it represents the parent company
UPDATE files
SET company_id = 'gruppen'
WHERE company_id IS NULL;

-- Now make the column NOT NULL since all files should have a company
ALTER TABLE files
ALTER COLUMN company_id SET NOT NULL;

-- Add foreign key constraint to ensure company_id is valid
ALTER TABLE files
ADD CONSTRAINT fk_files_company_id
FOREIGN KEY (company_id) REFERENCES companies(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove foreign key constraint
ALTER TABLE files
DROP CONSTRAINT IF EXISTS fk_files_company_id;

-- Remove NOT NULL constraint and default
ALTER TABLE files
ALTER COLUMN company_id DROP NOT NULL;

-- Remove index
DROP INDEX IF EXISTS idx_files_company_id;

-- Remove the column
ALTER TABLE files
DROP COLUMN IF EXISTS company_id;

-- +goose StatementEnd
