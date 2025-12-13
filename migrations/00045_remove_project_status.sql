-- +goose Up
-- +goose StatementBegin

-- Migrate status values to phase
UPDATE projects SET phase = 'working' WHERE status = 'active';
UPDATE projects SET phase = 'cancelled' WHERE status = 'cancelled';
UPDATE projects SET phase = 'completed' WHERE status = 'completed';
-- planning stays as tilbud (already default)
-- on_hold maps to working (if any exist)
UPDATE projects SET phase = 'working' WHERE status = 'on_hold';

-- Drop the status column
ALTER TABLE projects DROP COLUMN IF EXISTS status;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add status column
ALTER TABLE projects ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'planning';

-- Create index
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

-- Migrate phase back to status
UPDATE projects SET status = 'active' WHERE phase IN ('working', 'active');
UPDATE projects SET status = 'cancelled' WHERE phase = 'cancelled';
UPDATE projects SET status = 'completed' WHERE phase = 'completed';
UPDATE projects SET status = 'planning' WHERE phase = 'tilbud';

-- +goose StatementEnd
