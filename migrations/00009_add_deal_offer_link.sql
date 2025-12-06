-- +goose Up
-- +goose StatementBegin

-- Add offer_id column to deals table to link deals with offers
ALTER TABLE deals ADD COLUMN IF NOT EXISTS offer_id UUID REFERENCES offers(id);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_deals_offer_id ON deals(offer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the index first
DROP INDEX IF EXISTS idx_deals_offer_id;

-- Remove the offer_id column
ALTER TABLE deals DROP COLUMN IF EXISTS offer_id;

-- +goose StatementEnd
