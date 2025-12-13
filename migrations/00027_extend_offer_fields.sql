-- +goose Up
-- +goose StatementBegin

-- Add extended offer fields for better CRM functionality
ALTER TABLE offers
    ADD COLUMN cost DECIMAL(15,2) DEFAULT 0,
    ADD COLUMN location VARCHAR(200),
    ADD COLUMN sent_date TIMESTAMP;

-- Add index for location queries
CREATE INDEX idx_offers_location ON offers(location);
CREATE INDEX idx_offers_sent_date ON offers(sent_date);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_offers_sent_date;
DROP INDEX IF EXISTS idx_offers_location;

ALTER TABLE offers
    DROP COLUMN IF EXISTS sent_date,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS cost;

-- +goose StatementEnd
