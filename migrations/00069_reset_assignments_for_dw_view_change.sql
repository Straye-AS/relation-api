-- +goose Up
-- Reset assignments table due to change in DW source from cw_* tables to dbo.Arbeidsordre view
-- The dw_assignment_id values were incorrectly parsed before; now using Arbeidsordrenr correctly
-- The sync job will repopulate this table with correct data on next run

TRUNCATE TABLE assignments;

-- +goose Down
-- No rollback needed - data will be re-synced from DW
-- The old data was incorrect anyway
