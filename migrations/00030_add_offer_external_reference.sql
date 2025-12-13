-- +goose Up
-- Add external_reference field to offers table
-- This stores customer/external system reference numbers (e.g., customer's project number)

ALTER TABLE offers ADD COLUMN external_reference VARCHAR(100);

CREATE INDEX idx_offers_external_reference ON offers(external_reference) WHERE external_reference IS NOT NULL;

COMMENT ON COLUMN offers.external_reference IS 'External reference number from customer or other systems';

-- +goose Down
DROP INDEX IF EXISTS idx_offers_external_reference;
ALTER TABLE offers DROP COLUMN external_reference;
