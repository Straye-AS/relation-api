-- +goose Up
-- +goose StatementBegin

-- Make customer_id nullable on offers table
-- Offers can now exist without a customer when linked to a project
ALTER TABLE offers ALTER COLUMN customer_id DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore NOT NULL constraint (will fail if there are NULL values)
ALTER TABLE offers ALTER COLUMN customer_id SET NOT NULL;

-- +goose StatementEnd
