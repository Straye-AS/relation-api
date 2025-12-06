-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN IF NOT EXISTS summary VARCHAR(500);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN IF EXISTS summary;
-- +goose StatementEnd
