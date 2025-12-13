-- +goose Up
-- +goose StatementBegin

-- Make manager_id nullable for historical imports
-- We have manager_name which is sufficient for display
ALTER TABLE projects ALTER COLUMN manager_id DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore NOT NULL (set empty string for nulls first)
UPDATE projects SET manager_id = '' WHERE manager_id IS NULL;
ALTER TABLE projects ALTER COLUMN manager_id SET NOT NULL;

-- +goose StatementEnd
