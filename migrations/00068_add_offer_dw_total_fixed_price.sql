-- +goose Up
-- +goose StatementBegin

-- Add column to store the sum of FixedPriceAmount from synced assignments
-- This is automatically updated when assignments are synced from the datawarehouse
ALTER TABLE offers ADD COLUMN dw_total_fixed_price DECIMAL(15,2) DEFAULT 0;

COMMENT ON COLUMN offers.dw_total_fixed_price IS 'Sum of FixedPriceAmount from all synced assignments. Updated by assignment sync.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE offers DROP COLUMN IF EXISTS dw_total_fixed_price;
-- +goose StatementEnd

