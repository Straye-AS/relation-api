-- +goose Up
-- +goose StatementBegin

-- Add status, tier, industry columns to customers table
ALTER TABLE customers ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS tier VARCHAR(50) NOT NULL DEFAULT 'bronze';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS industry VARCHAR(50);

-- Create indexes for filtering on new customer columns
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_tier ON customers(tier);
CREATE INDEX IF NOT EXISTS idx_customers_industry ON customers(industry) WHERE industry IS NOT NULL;

-- Add contact_type column to contacts table
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS contact_type VARCHAR(50) NOT NULL DEFAULT 'primary';

-- Create index for contact_type filtering
CREATE INDEX IF NOT EXISTS idx_contacts_contact_type ON contacts(contact_type);

-- Add unique constraint on contact email (case-insensitive)
-- First create a unique index on lower(email) for non-empty emails
CREATE UNIQUE INDEX IF NOT EXISTS idx_contacts_email_unique ON contacts(LOWER(email)) WHERE email IS NOT NULL AND email != '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove unique constraint on contact email
DROP INDEX IF EXISTS idx_contacts_email_unique;

-- Remove contact_type column from contacts
DROP INDEX IF EXISTS idx_contacts_contact_type;
ALTER TABLE contacts DROP COLUMN IF EXISTS contact_type;

-- Remove customer enhancements
DROP INDEX IF EXISTS idx_customers_industry;
DROP INDEX IF EXISTS idx_customers_tier;
DROP INDEX IF EXISTS idx_customers_status;
ALTER TABLE customers DROP COLUMN IF EXISTS industry;
ALTER TABLE customers DROP COLUMN IF EXISTS tier;
ALTER TABLE customers DROP COLUMN IF EXISTS status;

-- +goose StatementEnd
