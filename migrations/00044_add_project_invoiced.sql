-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN IF NOT EXISTS invoiced DECIMAL(15,2) NOT NULL DEFAULT 0;

COMMENT ON COLUMN projects.invoiced IS 'Amount invoiced to customer (hittil fakturert)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN IF EXISTS invoiced;
-- +goose StatementEnd
