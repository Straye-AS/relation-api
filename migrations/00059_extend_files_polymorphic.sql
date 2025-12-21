-- +goose Up
-- +goose StatementBegin

-- Extend files table to support attachments for suppliers and customers
-- This enables polymorphic file associations beyond just offers

-- Add supplier_id column with foreign key reference
ALTER TABLE files ADD COLUMN supplier_id UUID REFERENCES suppliers(id) ON DELETE SET NULL;

-- Add customer_id column with foreign key reference
ALTER TABLE files ADD COLUMN customer_id UUID REFERENCES customers(id) ON DELETE SET NULL;

-- Create indexes for efficient file lookups by entity
CREATE INDEX idx_files_supplier_id ON files(supplier_id);
CREATE INDEX idx_files_customer_id ON files(customer_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_files_customer_id;
DROP INDEX IF EXISTS idx_files_supplier_id;
ALTER TABLE files DROP COLUMN IF EXISTS customer_id;
ALTER TABLE files DROP COLUMN IF EXISTS supplier_id;
-- +goose StatementEnd
