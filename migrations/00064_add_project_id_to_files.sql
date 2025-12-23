-- +goose Up
-- +goose StatementBegin

-- Add project_id column to files table for polymorphic file associations
-- This completes the polymorphic design allowing files to be attached to projects
ALTER TABLE files ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE SET NULL;

-- Create index for efficient file lookups by project
CREATE INDEX idx_files_project_id ON files(project_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_files_project_id;
ALTER TABLE files DROP COLUMN IF EXISTS project_id;
-- +goose StatementEnd
