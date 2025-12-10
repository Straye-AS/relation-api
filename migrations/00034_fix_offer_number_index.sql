-- +goose Up
-- +goose StatementBegin

-- Drop the old index that only checks for NULL
DROP INDEX IF EXISTS idx_offers_company_offer_number;

-- Create improved unique index that excludes both NULL and empty string values
CREATE UNIQUE INDEX idx_offers_company_offer_number
    ON offers(company_id, offer_number)
    WHERE offer_number IS NOT NULL AND offer_number <> '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original index
DROP INDEX IF EXISTS idx_offers_company_offer_number;

CREATE UNIQUE INDEX idx_offers_company_offer_number
    ON offers(company_id, offer_number)
    WHERE offer_number IS NOT NULL;

-- +goose StatementEnd
