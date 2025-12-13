-- +goose Up
-- +goose StatementBegin

-- Create project_phase enum type
-- tilbud = offer/bidding phase (default)
-- active = project is active/in progress
-- completed = project is finished
-- cancelled = project was cancelled
CREATE TYPE project_phase AS ENUM ('tilbud', 'active', 'completed', 'cancelled');

-- Add new columns to projects table
ALTER TABLE projects
    ADD COLUMN phase project_phase NOT NULL DEFAULT 'tilbud',
    ADD COLUMN winning_offer_id uuid REFERENCES offers(id) ON DELETE SET NULL,
    ADD COLUMN inherited_offer_number varchar(50),
    ADD COLUMN calculated_offer_value decimal(15,2) DEFAULT 0,
    ADD COLUMN won_at timestamp;

-- Create index on phase for filtering
CREATE INDEX idx_projects_phase ON projects(phase);

-- Create index on winning_offer_id for lookups
CREATE INDEX idx_projects_winning_offer_id ON projects(winning_offer_id);

-- Migrate existing projects based on their status:
-- - planning, active, on_hold -> retain current status but set phase based on status
-- - completed -> set phase to 'completed'
-- - cancelled -> set phase to 'cancelled'
UPDATE projects SET phase = 'completed' WHERE status = 'completed';
UPDATE projects SET phase = 'cancelled' WHERE status = 'cancelled';
UPDATE projects SET phase = 'active' WHERE status IN ('active', 'on_hold', 'planning');

-- Add comment explaining the relationship
COMMENT ON COLUMN projects.phase IS 'Project lifecycle phase: tilbud (bidding), active, completed, cancelled. During tilbud phase, economic values mirror highest offer.';
COMMENT ON COLUMN projects.winning_offer_id IS 'Reference to the offer that won this project and triggered the transition to active phase';
COMMENT ON COLUMN projects.inherited_offer_number IS 'The original offer number inherited when project was won';
COMMENT ON COLUMN projects.calculated_offer_value IS 'Cached highest offer value during tilbud phase; editable base value after winning';
COMMENT ON COLUMN projects.won_at IS 'Timestamp when the project was won (offer accepted)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS idx_projects_winning_offer_id;
DROP INDEX IF EXISTS idx_projects_phase;

-- Remove columns from projects
ALTER TABLE projects
    DROP COLUMN IF EXISTS won_at,
    DROP COLUMN IF EXISTS calculated_offer_value,
    DROP COLUMN IF EXISTS inherited_offer_number,
    DROP COLUMN IF EXISTS winning_offer_id,
    DROP COLUMN IF EXISTS phase;

-- Drop the enum type
DROP TYPE IF EXISTS project_phase;

-- +goose StatementEnd
