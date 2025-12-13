-- +goose Up
-- +goose StatementBegin

-- Add price column (what we charge the customer)
ALTER TABLE offers ADD COLUMN price DECIMAL(15,2) NOT NULL DEFAULT 0;

-- Add margin_percent column (dekningsgrad - automatically calculated)
ALTER TABLE offers ADD COLUMN margin_percent DECIMAL(8,4) NOT NULL DEFAULT 0;

COMMENT ON COLUMN offers.price IS 'The price charged to the customer';
COMMENT ON COLUMN offers.cost IS 'Internal cost for delivering the offer';
COMMENT ON COLUMN offers.margin_percent IS 'Dekningsgrad: (price - cost) / price * 100, auto-calculated';

-- Move current cost values to price (cost was being used as price)
UPDATE offers SET price = cost;

-- Clear cost (will be entered separately later)
UPDATE offers SET cost = 0;

-- Calculate margin_percent for all offers
-- Formula: (price - cost) / price * 100
-- Edge cases:
--   - cost=0 and price>0: 100%
--   - price=0 and cost>0: 0%
--   - both 0: 0%
UPDATE offers SET margin_percent = CASE
    WHEN price > 0 THEN ((price - cost) / price) * 100
    ELSE 0
END;

-- Create a function to calculate margin_percent
CREATE OR REPLACE FUNCTION calculate_offer_margin_percent()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate margin_percent based on price and cost
    -- dekningsgrad = (price - cost) / price * 100
    IF NEW.price > 0 THEN
        NEW.margin_percent := ((NEW.price - COALESCE(NEW.cost, 0)) / NEW.price) * 100;
    ELSE
        -- If price is 0 or null, margin is 0
        NEW.margin_percent := 0;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-calculate margin_percent on insert/update
CREATE TRIGGER trigger_calculate_offer_margin_percent
    BEFORE INSERT OR UPDATE OF price, cost ON offers
    FOR EACH ROW
    EXECUTE FUNCTION calculate_offer_margin_percent();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trigger_calculate_offer_margin_percent ON offers;
DROP FUNCTION IF EXISTS calculate_offer_margin_percent();

-- Move price back to cost
UPDATE offers SET cost = price;

ALTER TABLE offers DROP COLUMN margin_percent;
ALTER TABLE offers DROP COLUMN price;

-- +goose StatementEnd
