-- +goose Up
ALTER TABLE projects ADD COLUMN external_reference VARCHAR(100);

-- +goose Down
ALTER TABLE projects DROP COLUMN external_reference;
