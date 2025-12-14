-- +goose Up
-- +goose StatementBegin

-- Add user tracking fields to customers table
ALTER TABLE customers
ADD COLUMN created_by_id VARCHAR(100),
ADD COLUMN created_by_name VARCHAR(200),
ADD COLUMN updated_by_id VARCHAR(100),
ADD COLUMN updated_by_name VARCHAR(200);

-- Add user tracking fields to projects table
ALTER TABLE projects
ADD COLUMN created_by_id VARCHAR(100),
ADD COLUMN created_by_name VARCHAR(200),
ADD COLUMN updated_by_id VARCHAR(100),
ADD COLUMN updated_by_name VARCHAR(200);

-- Add user tracking fields to offers table
ALTER TABLE offers
ADD COLUMN created_by_id VARCHAR(100),
ADD COLUMN created_by_name VARCHAR(200),
ADD COLUMN updated_by_id VARCHAR(100),
ADD COLUMN updated_by_name VARCHAR(200);

-- Add user tracking fields to contacts table
ALTER TABLE contacts
ADD COLUMN created_by_id VARCHAR(100),
ADD COLUMN created_by_name VARCHAR(200),
ADD COLUMN updated_by_id VARCHAR(100),
ADD COLUMN updated_by_name VARCHAR(200);

-- Create indexes on created_by_id for efficient querying
CREATE INDEX idx_customers_created_by_id ON customers(created_by_id);
CREATE INDEX idx_projects_created_by_id ON projects(created_by_id);
CREATE INDEX idx_offers_created_by_id ON offers(created_by_id);
CREATE INDEX idx_contacts_created_by_id ON contacts(created_by_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_customers_created_by_id;
DROP INDEX IF EXISTS idx_projects_created_by_id;
DROP INDEX IF EXISTS idx_offers_created_by_id;
DROP INDEX IF EXISTS idx_contacts_created_by_id;

-- Remove columns from customers
ALTER TABLE customers
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

-- Remove columns from projects
ALTER TABLE projects
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

-- Remove columns from offers
ALTER TABLE offers
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

-- Remove columns from contacts
ALTER TABLE contacts
DROP COLUMN IF EXISTS created_by_id,
DROP COLUMN IF EXISTS created_by_name,
DROP COLUMN IF EXISTS updated_by_id,
DROP COLUMN IF EXISTS updated_by_name;

-- +goose StatementEnd
