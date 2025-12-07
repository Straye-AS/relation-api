-- +goose Up
-- +goose StatementBegin

-- Add company_id column to support company-specific budget categories
-- NULL company_id = global/default category available to all companies
ALTER TABLE budget_dimension_categories
ADD COLUMN IF NOT EXISTS company_id VARCHAR(50) REFERENCES companies(id) ON DELETE CASCADE;

-- Create index for efficient company-based lookups
CREATE INDEX IF NOT EXISTS idx_budget_dimension_categories_company_id
ON budget_dimension_categories(company_id);

-- Create unique constraint for category name within company scope
-- This allows same category name in different companies, but unique within each company
-- NULL company_id categories (global) must also have unique names
CREATE UNIQUE INDEX IF NOT EXISTS idx_budget_dimension_categories_company_name
ON budget_dimension_categories(COALESCE(company_id, ''), name);

-- Seed global/default budget dimension categories (company_id = NULL)
-- These are available to all companies as a starting template
-- Companies can create their own custom categories with their company_id set

INSERT INTO budget_dimension_categories (id, company_id, name, description, display_order, is_active, created_at, updated_at) VALUES
    ('steel_structure', NULL, 'Steel Structure', 'Primary steel framework and structural elements', 1, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('hybrid_structure', NULL, 'Hybrid Structure', 'Combined steel and other material structures', 2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('roofing', NULL, 'Roofing', 'Roof installation, materials and labor', 3, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('cladding', NULL, 'Cladding', 'Wall cladding and facade materials', 4, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('foundation', NULL, 'Foundation', 'Concrete foundation and groundwork', 5, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('assembly', NULL, 'Assembly', 'On-site assembly and installation labor', 6, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('transport', NULL, 'Transport', 'Delivery and logistics costs', 7, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('engineering', NULL, 'Engineering', 'Design and engineering services', 8, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('project_management', NULL, 'Project Management', 'PM overhead and coordination', 9, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('crane_rigging', NULL, 'Crane & Rigging', 'Crane rental and rigging services', 10, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('miscellaneous', NULL, 'Miscellaneous', 'Other uncategorized costs', 11, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('contingency', NULL, 'Contingency', 'Risk buffer and unforeseen costs', 12, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    company_id = NULL,  -- Ensure global categories stay global
    updated_at = CURRENT_TIMESTAMP;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the seed data (only the specific IDs we inserted)
DELETE FROM budget_dimension_categories
WHERE id IN (
    'steel_structure',
    'hybrid_structure',
    'roofing',
    'cladding',
    'foundation',
    'assembly',
    'transport',
    'engineering',
    'project_management',
    'crane_rigging',
    'miscellaneous',
    'contingency'
);

-- Remove the unique constraint
DROP INDEX IF EXISTS idx_budget_dimension_categories_company_name;

-- Remove the company_id index
DROP INDEX IF EXISTS idx_budget_dimension_categories_company_id;

-- Remove the company_id column
ALTER TABLE budget_dimension_categories DROP COLUMN IF EXISTS company_id;

-- +goose StatementEnd
