-- +goose Up
-- +goose StatementBegin

-- Add data warehouse synced financial fields to offers table.
-- These fields store the latest financial data synchronized from the data warehouse
-- for offers that have an external_reference linking them to ERP projects.

-- Total income from data warehouse (accounts 3000-3999)
ALTER TABLE offers ADD COLUMN dw_total_income DECIMAL(15,2) NOT NULL DEFAULT 0;
COMMENT ON COLUMN offers.dw_total_income IS 'Total income synchronized from data warehouse (accounts 3000-3999)';

-- Material costs from data warehouse (accounts 4000-4999)
ALTER TABLE offers ADD COLUMN dw_material_costs DECIMAL(15,2) NOT NULL DEFAULT 0;
COMMENT ON COLUMN offers.dw_material_costs IS 'Material costs synchronized from data warehouse (accounts 4000-4999)';

-- Employee costs from data warehouse (accounts 5000-5999)
ALTER TABLE offers ADD COLUMN dw_employee_costs DECIMAL(15,2) NOT NULL DEFAULT 0;
COMMENT ON COLUMN offers.dw_employee_costs IS 'Employee costs synchronized from data warehouse (accounts 5000-5999)';

-- Other costs from data warehouse (accounts >= 6000)
ALTER TABLE offers ADD COLUMN dw_other_costs DECIMAL(15,2) NOT NULL DEFAULT 0;
COMMENT ON COLUMN offers.dw_other_costs IS 'Other costs synchronized from data warehouse (accounts >= 6000)';

-- Net result from data warehouse (income - all costs)
ALTER TABLE offers ADD COLUMN dw_net_result DECIMAL(15,2) NOT NULL DEFAULT 0;
COMMENT ON COLUMN offers.dw_net_result IS 'Net result synchronized from data warehouse (income - costs)';

-- Timestamp of last successful data warehouse sync
ALTER TABLE offers ADD COLUMN dw_last_synced_at TIMESTAMP;
COMMENT ON COLUMN offers.dw_last_synced_at IS 'Timestamp of last successful data warehouse synchronization';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE offers DROP COLUMN IF EXISTS dw_total_income;
ALTER TABLE offers DROP COLUMN IF EXISTS dw_material_costs;
ALTER TABLE offers DROP COLUMN IF EXISTS dw_employee_costs;
ALTER TABLE offers DROP COLUMN IF EXISTS dw_other_costs;
ALTER TABLE offers DROP COLUMN IF EXISTS dw_net_result;
ALTER TABLE offers DROP COLUMN IF EXISTS dw_last_synced_at;

-- +goose StatementEnd
