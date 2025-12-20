-- +goose Up
-- +goose StatementBegin

-- Create supplier status enum
CREATE TYPE supplier_status AS ENUM ('active', 'inactive', 'pending', 'blacklisted');

-- Create suppliers table
CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    org_number VARCHAR(20) UNIQUE,
    email VARCHAR(255),
    phone VARCHAR(50),
    address VARCHAR(500),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Norway',
    municipality VARCHAR(100),
    county VARCHAR(100),
    contact_person VARCHAR(200),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    status supplier_status NOT NULL DEFAULT 'active',
    category VARCHAR(200),
    notes TEXT,
    payment_terms VARCHAR(200),
    website VARCHAR(500),
    company_id VARCHAR(50),
    created_by_id VARCHAR(100),
    created_by_name VARCHAR(200),
    updated_by_id VARCHAR(100),
    updated_by_name VARCHAR(200),
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for common query patterns
CREATE INDEX idx_suppliers_name ON suppliers(name);
CREATE INDEX idx_suppliers_org_number ON suppliers(org_number);
CREATE INDEX idx_suppliers_company_id ON suppliers(company_id);
CREATE INDEX idx_suppliers_status ON suppliers(status);
CREATE INDEX idx_suppliers_category ON suppliers(category);
CREATE INDEX idx_suppliers_city ON suppliers(city);
CREATE INDEX idx_suppliers_deleted_at ON suppliers(deleted_at);
CREATE INDEX idx_suppliers_created_by_id ON suppliers(created_by_id);

-- Create trigger for automatic updated_at timestamp
CREATE TRIGGER update_suppliers_updated_at
    BEFORE UPDATE ON suppliers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_suppliers_updated_at ON suppliers;
DROP TABLE IF EXISTS suppliers;
DROP TYPE IF EXISTS supplier_status;
-- +goose StatementEnd
