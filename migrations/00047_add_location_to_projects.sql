-- +goose Up
ALTER TABLE projects ADD COLUMN location VARCHAR(200);

-- +goose Down
ALTER TABLE projects DROP COLUMN location;
