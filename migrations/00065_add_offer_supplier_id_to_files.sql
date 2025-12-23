-- +goose Up
-- +goose StatementBegin
ALTER TABLE files ADD COLUMN offer_supplier_id UUID REFERENCES offer_suppliers(id) ON DELETE SET NULL;
CREATE INDEX idx_files_offer_supplier_id ON files(offer_supplier_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_files_offer_supplier_id;
ALTER TABLE files DROP COLUMN IF EXISTS offer_supplier_id;
-- +goose StatementEnd
