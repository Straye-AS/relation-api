-- +goose Up
-- +goose StatementBegin

-- Create companies table
CREATE TABLE companies (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    short_name VARCHAR(50) NOT NULL,
    org_number VARCHAR(20),
    color VARCHAR(20) NOT NULL DEFAULT '#000000',
    logo VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert seed data for Straye companies
INSERT INTO companies (id, name, short_name, org_number, color, is_active) VALUES
    ('gruppen', 'Straye Gruppen AS', 'Gruppen', '929 677 856', '#1a1a2e', true),
    ('stalbygg', 'Straye Stålbygg AS', 'Stålbygg', '929 677 864', '#e63946', true),
    ('hybridbygg', 'Straye Hybridbygg AS', 'Hybridbygg', '929 677 872', '#2a9d8f', true),
    ('industri', 'Straye Industri AS', 'Industri', '929 677 880', '#e9c46a', true),
    ('tak', 'Straye Tak AS', 'Tak', '929 677 899', '#264653', true),
    ('montasje', 'Straye Montasje AS', 'Montasje', '929 677 902', '#f4a261', true);

-- Create trigger for companies updated_at
CREATE TRIGGER update_companies_updated_at BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add new columns to users table
-- Change id from VARCHAR to UUID for consistency
ALTER TABLE users
    ADD COLUMN azure_ad_oid VARCHAR(100) UNIQUE,
    ADD COLUMN first_name VARCHAR(100),
    ADD COLUMN last_name VARCHAR(100),
    ADD COLUMN company_id VARCHAR(50) REFERENCES companies(id),
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN last_login_at TIMESTAMP;

-- Create index on users.company_id for efficient joins
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_users_azure_ad_oid ON users(azure_ad_oid);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Add company_id to customers table for multi-tenancy
ALTER TABLE customers
    ADD COLUMN company_id VARCHAR(50) REFERENCES companies(id);

CREATE INDEX idx_customers_company_id ON customers(company_id);

-- Update offers table to reference companies properly
-- First, add foreign key constraint (company_id column already exists as VARCHAR(50))
ALTER TABLE offers
    ADD CONSTRAINT fk_offers_company_id FOREIGN KEY (company_id) REFERENCES companies(id);

-- Update projects table to reference companies properly
ALTER TABLE projects
    ADD CONSTRAINT fk_projects_company_id FOREIGN KEY (company_id) REFERENCES companies(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove foreign key constraints from projects
ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS fk_projects_company_id;

-- Remove foreign key constraints from offers
ALTER TABLE offers
    DROP CONSTRAINT IF EXISTS fk_offers_company_id;

-- Remove company_id from customers
DROP INDEX IF EXISTS idx_customers_company_id;
ALTER TABLE customers
    DROP COLUMN IF EXISTS company_id;

-- Remove new columns from users
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_azure_ad_oid;
DROP INDEX IF EXISTS idx_users_company_id;
ALTER TABLE users
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS company_id,
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name,
    DROP COLUMN IF EXISTS azure_ad_oid;

-- Drop companies table
DROP TRIGGER IF EXISTS update_companies_updated_at ON companies;
DROP TABLE IF EXISTS companies;

-- +goose StatementEnd
