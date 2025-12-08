-- +goose Up
-- +goose StatementBegin

-- Add attendees array field to activities table for meeting attendee management
ALTER TABLE activities
    ADD COLUMN IF NOT EXISTS attendees TEXT[] DEFAULT '{}';

-- Add parent_activity_id for follow-up task linking
ALTER TABLE activities
    ADD COLUMN IF NOT EXISTS parent_activity_id UUID REFERENCES activities(id) ON DELETE SET NULL;

-- Create index for efficient attendee lookups
CREATE INDEX IF NOT EXISTS idx_activities_attendees ON activities USING GIN (attendees);

-- Create index for parent activity lookups
CREATE INDEX IF NOT EXISTS idx_activities_parent_activity_id ON activities(parent_activity_id) WHERE parent_activity_id IS NOT NULL;

COMMENT ON COLUMN activities.attendees IS 'Array of user IDs attending the meeting activity';
COMMENT ON COLUMN activities.parent_activity_id IS 'Reference to parent activity for follow-up tasks';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_activities_parent_activity_id;
DROP INDEX IF EXISTS idx_activities_attendees;
ALTER TABLE activities DROP COLUMN IF EXISTS parent_activity_id;
ALTER TABLE activities DROP COLUMN IF EXISTS attendees;

-- +goose StatementEnd
