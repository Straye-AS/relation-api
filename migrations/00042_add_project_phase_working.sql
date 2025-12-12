-- +goose Up
-- +goose StatementBegin
-- Add 'working' value to project_phase enum
-- This represents a project that has started work (has StartDate) but doesn't have a WinningOfferID yet
ALTER TYPE project_phase ADD VALUE IF NOT EXISTS 'working' AFTER 'tilbud';
-- +goose StatementEnd

-- +goose Down
-- Note: PostgreSQL does not support removing enum values directly.
-- The 'working' phase will remain in the enum but will not be used if this migration is rolled back.
-- To properly remove it would require recreating the enum type which is destructive.
-- +goose StatementBegin
SELECT 1; -- No-op placeholder
-- +goose StatementEnd
