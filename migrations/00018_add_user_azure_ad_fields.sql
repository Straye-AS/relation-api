-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS azure_ad_roles TEXT[];
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_ip_address VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS azure_ad_roles;
ALTER TABLE users DROP COLUMN IF EXISTS last_ip_address;
-- +goose StatementEnd
