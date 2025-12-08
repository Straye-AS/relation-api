-- +goose Up
-- +goose StatementBegin
ALTER TABLE audit_logs ALTER COLUMN ip_address TYPE varchar(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Note: This may fail if there are IPv6 addresses with brackets stored
ALTER TABLE audit_logs ALTER COLUMN ip_address TYPE inet USING ip_address::inet;
-- +goose StatementEnd
