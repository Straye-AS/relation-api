-- +goose Up
-- +goose StatementBegin

-- Add project_name as denormalized field for display
ALTER TABLE offers ADD COLUMN project_name VARCHAR(200);

-- Populate existing offers with project names
UPDATE offers o
SET project_name = p.name
FROM projects p
WHERE o.project_id = p.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE offers DROP COLUMN IF EXISTS project_name;

-- +goose StatementEnd
