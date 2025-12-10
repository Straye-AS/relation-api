-- +goose Up
-- +goose StatementBegin

-- Add expiration_date column to offers table
-- This field indicates when the offer expires for the customer to accept
-- Defaults to 60 days after sent_date when the offer is sent
ALTER TABLE offers ADD COLUMN expiration_date TIMESTAMP;

-- Add index for efficient querying of expiring offers
CREATE INDEX idx_offers_expiration_date ON offers(expiration_date) WHERE expiration_date IS NOT NULL;

-- Add comment explaining the field
COMMENT ON COLUMN offers.expiration_date IS 'Date when the offer expires. Defaults to 60 days after sent_date. Frontend shows expired status but no automatic phase change.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_offers_expiration_date;
ALTER TABLE offers DROP COLUMN IF EXISTS expiration_date;

-- +goose StatementEnd
