-- +goose Up
-- +goose StatementBegin
-- Add website URL field to customers table (similar to suppliers)
ALTER TABLE customers ADD COLUMN IF NOT EXISTS website VARCHAR(500);

COMMENT ON COLUMN customers.website IS 'Customer website URL';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers DROP COLUMN IF EXISTS website;
-- +goose StatementEnd
