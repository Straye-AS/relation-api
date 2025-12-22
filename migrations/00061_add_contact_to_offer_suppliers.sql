-- +goose Up
-- +goose StatementBegin

-- Add contact person reference to offer_suppliers junction table
-- This allows selecting a specific contact person for a supplier on each offer
ALTER TABLE offer_suppliers
ADD COLUMN contact_id UUID REFERENCES supplier_contacts(id) ON DELETE SET NULL,
ADD COLUMN contact_name VARCHAR(200);

-- Create index for contact lookups
CREATE INDEX idx_offer_suppliers_contact_id ON offer_suppliers(contact_id) WHERE contact_id IS NOT NULL;

-- Add comment explaining the relationship
COMMENT ON COLUMN offer_suppliers.contact_id IS 'Optional contact person from the supplier for this specific offer';
COMMENT ON COLUMN offer_suppliers.contact_name IS 'Denormalized contact name for display purposes';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_offer_suppliers_contact_id;

ALTER TABLE offer_suppliers
DROP COLUMN IF EXISTS contact_name,
DROP COLUMN IF EXISTS contact_id;

-- +goose StatementEnd
