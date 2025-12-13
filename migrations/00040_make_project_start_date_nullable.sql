-- +goose Up
-- +goose StatementBegin

-- Make start_date nullable for historical imports
-- Some offers don't have sent_date, so we can't determine start_date
ALTER TABLE projects ALTER COLUMN start_date DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore NOT NULL (set current date for nulls first)
UPDATE projects SET start_date = CURRENT_DATE WHERE start_date IS NULL;
ALTER TABLE projects ALTER COLUMN start_date SET NOT NULL;

-- +goose StatementEnd
