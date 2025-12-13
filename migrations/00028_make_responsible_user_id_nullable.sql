-- +goose Up
-- +goose StatementBegin

-- Make responsible_user_id nullable for historical imports
-- We have the responsible_user_name which is sufficient for display
ALTER TABLE offers ALTER COLUMN responsible_user_id DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore NOT NULL (set empty string for nulls first)
UPDATE offers SET responsible_user_id = '' WHERE responsible_user_id IS NULL;
ALTER TABLE offers ALTER COLUMN responsible_user_id SET NOT NULL;

-- +goose StatementEnd
