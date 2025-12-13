-- +goose Up
-- +goose StatementBegin
ALTER TABLE notifications ADD COLUMN read_at TIMESTAMP;

-- Set read_at for any existing read notifications to their updated_at time
UPDATE notifications SET read_at = updated_at WHERE read = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE notifications DROP COLUMN IF EXISTS read_at;
-- +goose StatementEnd
