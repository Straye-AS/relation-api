-- +goose Up
-- +goose StatementBegin

-- Create a custom trigger function for offers that checks a session variable
-- to determine whether to skip updating updated_at (used for DW sync operations)
CREATE OR REPLACE FUNCTION update_offers_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    -- Check if we should skip updating updated_at (set by application code for DW sync)
    -- current_setting with 'true' as second param returns NULL instead of error if not set
    IF current_setting('app.skip_updated_at', true) = 'true' THEN
        NEW.updated_at = OLD.updated_at;
    ELSE
        NEW.updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop the old trigger and create a new one using our custom function
DROP TRIGGER IF EXISTS update_offers_updated_at ON offers;
CREATE TRIGGER update_offers_updated_at
    BEFORE UPDATE ON offers
    FOR EACH ROW
    EXECUTE FUNCTION update_offers_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore the original trigger using the generic function
DROP TRIGGER IF EXISTS update_offers_updated_at ON offers;
CREATE TRIGGER update_offers_updated_at
    BEFORE UPDATE ON offers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Drop the custom function
DROP FUNCTION IF EXISTS update_offers_updated_at_column();

-- +goose StatementEnd
