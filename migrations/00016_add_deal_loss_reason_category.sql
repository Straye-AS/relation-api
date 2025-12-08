-- +goose Up
-- +goose StatementBegin

-- Add loss_reason_category column to deals table for categorizing why deals were lost
-- The existing lost_reason column will be used for detailed notes
ALTER TABLE deals ADD COLUMN IF NOT EXISTS loss_reason_category VARCHAR(50);

-- Add comment for documentation
COMMENT ON COLUMN deals.loss_reason_category IS 'Categorized reason for deal loss: price, timing, competitor, requirements, other';

-- Create index for filtering/reporting on loss reasons
CREATE INDEX IF NOT EXISTS idx_deals_loss_reason_category ON deals(loss_reason_category) WHERE loss_reason_category IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_deals_loss_reason_category;
ALTER TABLE deals DROP COLUMN IF EXISTS loss_reason_category;

-- +goose StatementEnd
