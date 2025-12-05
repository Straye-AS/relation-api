-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- DEALS TABLE
-- ============================================================================

-- Create deal stage enum type
CREATE TYPE deal_stage AS ENUM (
    'lead',           -- Initial contact/interest
    'qualified',      -- Qualified lead
    'proposal',       -- Proposal sent
    'negotiation',    -- In negotiation
    'won',            -- Deal closed won
    'lost'            -- Deal closed lost
);

-- Create deals table (replaces/extends offers for sales pipeline)
CREATE TABLE deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(200) NOT NULL,
    description TEXT,

    -- Relationships
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    company_id VARCHAR(50) NOT NULL REFERENCES companies(id),

    -- Denormalized fields for performance
    customer_name VARCHAR(200),

    -- Sales pipeline fields
    stage deal_stage NOT NULL DEFAULT 'lead',
    probability INT NOT NULL DEFAULT 0 CHECK (probability >= 0 AND probability <= 100),

    -- Financial fields
    value DECIMAL(15, 2) NOT NULL DEFAULT 0,
    weighted_value DECIMAL(15, 2) GENERATED ALWAYS AS (value * probability / 100) STORED,
    currency VARCHAR(3) NOT NULL DEFAULT 'NOK',

    -- Dates
    expected_close_date DATE,
    actual_close_date DATE,

    -- Assignment
    owner_id VARCHAR(100) NOT NULL,
    owner_name VARCHAR(200),

    -- Source tracking
    source VARCHAR(100),  -- e.g., 'referral', 'website', 'cold_call'

    -- Notes and metadata
    notes TEXT,
    lost_reason VARCHAR(500),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for deals
CREATE INDEX idx_deals_customer_id ON deals(customer_id);
CREATE INDEX idx_deals_company_id ON deals(company_id);
CREATE INDEX idx_deals_stage ON deals(stage);
CREATE INDEX idx_deals_owner_id ON deals(owner_id);
CREATE INDEX idx_deals_expected_close_date ON deals(expected_close_date);
CREATE INDEX idx_deals_created_at ON deals(created_at DESC);

-- Composite index for pipeline queries
CREATE INDEX idx_deals_company_stage ON deals(company_id, stage);

-- Create trigger for deals updated_at
CREATE TRIGGER update_deals_updated_at BEFORE UPDATE ON deals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- DEAL STAGE HISTORY (Audit trail for stage changes)
-- ============================================================================

CREATE TABLE deal_stage_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    from_stage deal_stage,
    to_stage deal_stage NOT NULL,
    changed_by_id VARCHAR(100) NOT NULL,
    changed_by_name VARCHAR(200),
    notes TEXT,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_deal_stage_history_deal_id ON deal_stage_history(deal_id);
CREATE INDEX idx_deal_stage_history_changed_at ON deal_stage_history(changed_at DESC);

-- ============================================================================
-- ENHANCED CONTACTS TABLE
-- ============================================================================

-- Create temporary table to backup existing contacts
CREATE TABLE contacts_backup AS SELECT * FROM contacts;

-- Drop old foreign key constraints and indexes
ALTER TABLE contacts DROP CONSTRAINT IF EXISTS contacts_customer_id_fkey;
ALTER TABLE contacts DROP CONSTRAINT IF EXISTS contacts_project_id_fkey;
DROP INDEX IF EXISTS idx_contacts_customer_id;
DROP INDEX IF EXISTS idx_contacts_project_id;
DROP INDEX IF EXISTS idx_contacts_name;

-- Drop old trigger
DROP TRIGGER IF EXISTS update_contacts_updated_at ON contacts;

-- Drop old contacts table
DROP TABLE contacts;

-- Create new contacts table with enhanced structure
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Name fields (split for better search/display)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,

    -- Contact info
    email VARCHAR(255),
    phone VARCHAR(50),
    mobile VARCHAR(50),

    -- Professional info
    title VARCHAR(100),           -- Job title
    department VARCHAR(100),

    -- Primary company relationship (for lookup/default)
    primary_customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,

    -- Address (optional, for direct contacts)
    address VARCHAR(500),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'Norway',

    -- Social/communication preferences
    linkedin_url VARCHAR(500),
    preferred_contact_method VARCHAR(50) DEFAULT 'email',

    -- Notes
    notes TEXT,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for contacts
CREATE INDEX idx_contacts_name ON contacts(last_name, first_name);
CREATE INDEX idx_contacts_email ON contacts(email);
CREATE INDEX idx_contacts_primary_customer ON contacts(primary_customer_id);
CREATE INDEX idx_contacts_is_active ON contacts(is_active);

-- Full text search index for contact names
CREATE INDEX idx_contacts_fullname ON contacts USING gin(to_tsvector('simple', first_name || ' ' || last_name));

-- Create trigger for contacts updated_at
CREATE TRIGGER update_contacts_updated_at BEFORE UPDATE ON contacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- CONTACT RELATIONSHIPS (Polymorphic associations)
-- ============================================================================

-- Create entity type enum for contact relationships
CREATE TYPE contact_entity_type AS ENUM (
    'customer',
    'project',
    'deal'
);

-- Create contact relationships table
CREATE TABLE contact_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contact_id UUID NOT NULL REFERENCES contacts(id) ON DELETE CASCADE,
    entity_type contact_entity_type NOT NULL,
    entity_id UUID NOT NULL,

    -- Relationship details
    role VARCHAR(100),            -- e.g., 'decision_maker', 'technical_contact', 'billing'
    is_primary BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure unique relationship per contact-entity pair
    UNIQUE(contact_id, entity_type, entity_id)
);

-- Create indexes for contact relationships
CREATE INDEX idx_contact_rel_contact_id ON contact_relationships(contact_id);
CREATE INDEX idx_contact_rel_entity ON contact_relationships(entity_type, entity_id);
CREATE INDEX idx_contact_rel_is_primary ON contact_relationships(is_primary) WHERE is_primary = true;

-- ============================================================================
-- DATA MIGRATION
-- ============================================================================

-- Migrate existing contacts from backup
INSERT INTO contacts (
    id,
    first_name,
    last_name,
    email,
    phone,
    title,
    primary_customer_id,
    notes,
    created_at,
    updated_at
)
SELECT
    id,
    -- Split name into first/last (simple split on first space)
    COALESCE(SPLIT_PART(name, ' ', 1), name) as first_name,
    COALESCE(NULLIF(SUBSTRING(name FROM POSITION(' ' IN name) + 1), ''), '-') as last_name,
    email,
    phone,
    role as title,
    customer_id as primary_customer_id,
    NULL as notes,
    created_at,
    updated_at
FROM contacts_backup;

-- Create contact relationships for migrated contacts (customer relationships)
INSERT INTO contact_relationships (contact_id, entity_type, entity_id, role, is_primary)
SELECT
    id as contact_id,
    'customer'::contact_entity_type as entity_type,
    customer_id as entity_id,
    role,
    true as is_primary
FROM contacts_backup
WHERE customer_id IS NOT NULL;

-- Create contact relationships for migrated contacts (project relationships)
INSERT INTO contact_relationships (contact_id, entity_type, entity_id, role, is_primary)
SELECT
    id as contact_id,
    'project'::contact_entity_type as entity_type,
    project_id as entity_id,
    role,
    false as is_primary
FROM contacts_backup
WHERE project_id IS NOT NULL;

-- Drop backup table
DROP TABLE contacts_backup;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate original contacts table
CREATE TABLE contacts_new (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    role VARCHAR(120),
    customer_id UUID,
    customer_name VARCHAR(200),
    project_id UUID,
    project_name VARCHAR(200),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Migrate data back
INSERT INTO contacts_new (id, name, email, phone, role, customer_id, created_at, updated_at)
SELECT
    c.id,
    c.first_name || ' ' || c.last_name as name,
    COALESCE(c.email, '') as email,
    COALESCE(c.phone, '') as phone,
    c.title as role,
    c.primary_customer_id as customer_id,
    c.created_at,
    c.updated_at
FROM contacts c;

-- Drop new tables
DROP TABLE IF EXISTS contact_relationships;
DROP TYPE IF EXISTS contact_entity_type;
DROP TABLE IF EXISTS contacts;

-- Rename back
ALTER TABLE contacts_new RENAME TO contacts;

-- Recreate indexes and constraints
CREATE INDEX idx_contacts_customer_id ON contacts(customer_id);
CREATE INDEX idx_contacts_name ON contacts(name);
ALTER TABLE contacts ADD CONSTRAINT contacts_customer_id_fkey
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE;

-- Create trigger
CREATE TRIGGER update_contacts_updated_at BEFORE UPDATE ON contacts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Drop deals tables
DROP TABLE IF EXISTS deal_stage_history;
DROP TABLE IF EXISTS deals;
DROP TYPE IF EXISTS deal_stage;

-- +goose StatementEnd
