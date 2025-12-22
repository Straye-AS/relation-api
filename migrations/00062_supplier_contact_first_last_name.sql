-- +goose Up
-- +goose StatementBegin

-- Add first_name and last_name columns
ALTER TABLE supplier_contacts ADD COLUMN first_name VARCHAR(100);
ALTER TABLE supplier_contacts ADD COLUMN last_name VARCHAR(100);

-- Migrate data from name column
-- Split name on first space, or use entire name as last_name if no space
UPDATE supplier_contacts SET
    first_name = CASE
        WHEN position(' ' in name) > 0 THEN split_part(name, ' ', 1)
        ELSE ''
    END,
    last_name = CASE
        WHEN position(' ' in name) > 0 THEN substring(name from position(' ' in name) + 1)
        ELSE name
    END;

-- Make columns not null after data migration
ALTER TABLE supplier_contacts ALTER COLUMN first_name SET NOT NULL;
ALTER TABLE supplier_contacts ALTER COLUMN last_name SET NOT NULL;

-- Drop the old name column
ALTER TABLE supplier_contacts DROP COLUMN name;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Add name column back
ALTER TABLE supplier_contacts ADD COLUMN name VARCHAR(200);

-- Migrate data back - concatenate first and last name
UPDATE supplier_contacts SET name = TRIM(first_name || ' ' || last_name);

-- Make name not null
ALTER TABLE supplier_contacts ALTER COLUMN name SET NOT NULL;

-- Drop first_name and last_name columns
ALTER TABLE supplier_contacts DROP COLUMN first_name;
ALTER TABLE supplier_contacts DROP COLUMN last_name;

-- +goose StatementEnd
