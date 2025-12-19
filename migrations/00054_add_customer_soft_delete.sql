-- +goose Up
-- +goose StatementBegin

-- Add soft delete support for customers
-- When a customer is deleted, they are hidden from normal queries but preserved
-- for historical reference in related projects and offers.
ALTER TABLE customers ADD COLUMN deleted_at TIMESTAMPTZ;

-- Index for efficient filtering of non-deleted customers
CREATE INDEX idx_customers_deleted_at ON customers(deleted_at);

-- Remove CASCADE delete constraints from related tables
-- Projects and offers should be preserved when a customer is soft deleted

-- Drop existing foreign key constraints and recreate without CASCADE
ALTER TABLE contacts DROP CONSTRAINT IF EXISTS fk_customers_contacts;
ALTER TABLE contacts ADD CONSTRAINT fk_customers_contacts
    FOREIGN KEY (primary_customer_id) REFERENCES customers(id) ON DELETE SET NULL;

-- For projects, customer_id is already nullable, so SET NULL is appropriate
ALTER TABLE projects DROP CONSTRAINT IF EXISTS fk_customers_projects;
-- Note: projects.customer_id doesn't have a named constraint, GORM creates it automatically
-- We'll just ensure the column allows null (which it already does)

-- For offers, customer_id is NOT NULL, so we keep the reference but don't cascade
-- The soft delete ensures the customer record still exists
ALTER TABLE offers DROP CONSTRAINT IF EXISTS fk_customers_offers;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the soft delete column
DROP INDEX IF EXISTS idx_customers_deleted_at;
ALTER TABLE customers DROP COLUMN IF EXISTS deleted_at;

-- Note: We don't restore CASCADE constraints in down migration
-- as that would be destructive behavior

-- +goose StatementEnd
