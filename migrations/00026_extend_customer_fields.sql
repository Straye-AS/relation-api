-- +goose Up
-- +goose StatementBegin

-- Add extended customer fields for better CRM functionality
ALTER TABLE customers
    ADD COLUMN notes TEXT,
    ADD COLUMN customer_class VARCHAR(50),
    ADD COLUMN credit_limit DECIMAL(15,2),
    ADD COLUMN is_internal BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN municipality VARCHAR(100),
    ADD COLUMN county VARCHAR(100);

-- Add index for commonly queried fields
CREATE INDEX idx_customers_county ON customers(county);
CREATE INDEX idx_customers_municipality ON customers(municipality);
CREATE INDEX idx_customers_is_internal ON customers(is_internal);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_customers_is_internal;
DROP INDEX IF EXISTS idx_customers_municipality;
DROP INDEX IF EXISTS idx_customers_county;

ALTER TABLE customers
    DROP COLUMN IF EXISTS county,
    DROP COLUMN IF EXISTS municipality,
    DROP COLUMN IF EXISTS is_internal,
    DROP COLUMN IF EXISTS credit_limit,
    DROP COLUMN IF EXISTS customer_class,
    DROP COLUMN IF EXISTS notes;

-- +goose StatementEnd
