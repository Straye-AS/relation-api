-- +migrate Up
-- Add project number for external system integration (ERP, accounting, etc.)

ALTER TABLE projects ADD COLUMN IF NOT EXISTS project_number VARCHAR(50);

-- Add unique index for project_number (allows NULL values)
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_project_number 
    ON projects (project_number) 
    WHERE project_number IS NOT NULL AND project_number != '';

COMMENT ON COLUMN projects.project_number IS 'External reference number for linking to ERP/accounting systems';

-- +migrate Down
DROP INDEX IF EXISTS idx_projects_project_number;
ALTER TABLE projects DROP COLUMN IF EXISTS project_number;
