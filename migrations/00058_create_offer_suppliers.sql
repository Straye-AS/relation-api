-- +goose Up
-- +goose StatementBegin

-- Create offer-supplier relationship status enum
CREATE TYPE offer_supplier_status AS ENUM ('active', 'done');

-- Create junction table for many-to-many relationship between offers and suppliers
CREATE TABLE offer_suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    offer_id UUID NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
    supplier_name VARCHAR(200),
    offer_title VARCHAR(200),
    status offer_supplier_status NOT NULL DEFAULT 'active',
    notes TEXT,
    created_by_id VARCHAR(100),
    created_by_name VARCHAR(200),
    updated_by_id VARCHAR(100),
    updated_by_name VARCHAR(200),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_offer_supplier UNIQUE (offer_id, supplier_id)
);

-- Create indexes for efficient lookups from both sides of the relationship
CREATE INDEX idx_offer_suppliers_offer_id ON offer_suppliers(offer_id);
CREATE INDEX idx_offer_suppliers_supplier_id ON offer_suppliers(supplier_id);
CREATE INDEX idx_offer_suppliers_status ON offer_suppliers(status);

-- Create trigger for automatic updated_at timestamp
CREATE TRIGGER update_offer_suppliers_updated_at
    BEFORE UPDATE ON offer_suppliers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_offer_suppliers_updated_at ON offer_suppliers;
DROP TABLE IF EXISTS offer_suppliers;
DROP TYPE IF EXISTS offer_supplier_status;
-- +goose StatementEnd
