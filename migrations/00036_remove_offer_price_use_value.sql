-- +goose Up
-- +goose StatementBegin

-- Remove the old trigger that uses price
DROP TRIGGER IF EXISTS trigger_calculate_offer_margin_percent ON offers;
DROP FUNCTION IF EXISTS calculate_offer_margin_percent();

-- Copy price values to value where value is 0 but price is set
-- This ensures we don't lose any data
UPDATE offers SET value = price WHERE value = 0 AND price > 0;

-- Create new function to calculate margin_percent from value and cost
CREATE OR REPLACE FUNCTION calculate_offer_margin_percent()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate margin_percent based on value and cost
    -- dekningsgrad = (value - cost) / value * 100
    IF NEW.value > 0 THEN
        NEW.margin_percent := ((NEW.value - COALESCE(NEW.cost, 0)) / NEW.value) * 100;
    ELSE
        -- If value is 0 or null, margin is 0
        NEW.margin_percent := 0;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to auto-calculate margin_percent on insert/update of value or cost
CREATE TRIGGER trigger_calculate_offer_margin_percent
    BEFORE INSERT OR UPDATE OF value, cost ON offers
    FOR EACH ROW
    EXECUTE FUNCTION calculate_offer_margin_percent();

-- Recalculate margin_percent for all offers using value
UPDATE offers SET margin_percent = CASE
    WHEN value > 0 THEN ((value - COALESCE(cost, 0)) / value) * 100
    ELSE 0
END;

-- Remove the price column
ALTER TABLE offers DROP COLUMN price;

-- Update column comment
COMMENT ON COLUMN offers.value IS 'The total value/price of the offer (sum of line item revenues or manually set)';
COMMENT ON COLUMN offers.margin_percent IS 'Dekningsgrad: (value - cost) / value * 100, auto-calculated';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add the price column
ALTER TABLE offers ADD COLUMN price DECIMAL(15,2) NOT NULL DEFAULT 0;

-- Copy value to price
UPDATE offers SET price = value;

-- Remove the trigger that uses value
DROP TRIGGER IF EXISTS trigger_calculate_offer_margin_percent ON offers;
DROP FUNCTION IF EXISTS calculate_offer_margin_percent();

-- Recreate original function using price
CREATE OR REPLACE FUNCTION calculate_offer_margin_percent()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.price > 0 THEN
        NEW.margin_percent := ((NEW.price - COALESCE(NEW.cost, 0)) / NEW.price) * 100;
    ELSE
        NEW.margin_percent := 0;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate trigger using price
CREATE TRIGGER trigger_calculate_offer_margin_percent
    BEFORE INSERT OR UPDATE OF price, cost ON offers
    FOR EACH ROW
    EXECUTE FUNCTION calculate_offer_margin_percent();

COMMENT ON COLUMN offers.price IS 'The price charged to the customer';

-- +goose StatementEnd
