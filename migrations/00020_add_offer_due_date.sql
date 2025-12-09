-- +goose Up
-- +goose StatementBegin
ALTER TABLE offers ADD COLUMN due_date TIMESTAMP;
CREATE INDEX idx_offers_due_date ON offers(due_date) WHERE due_date IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_offers_due_date;
ALTER TABLE offers DROP COLUMN IF EXISTS due_date;
-- +goose StatementEnd
