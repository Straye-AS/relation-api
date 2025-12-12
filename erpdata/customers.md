# Straye Tak Customers

Last updated: 2025-12-12

## Source of Truth

All customer data is now stored in `import_customers.sql`.

**Total customers: 121**
- 104 companies with org numbers (including 6 internal Straye companies)
- 17 private persons

## Usage

To restore customers after a database wipe:
```bash
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_customers.sql
```

Or run manually in psql:
```bash
docker exec -it relation-postgres psql -U relation_user -d relation
\i /path/to/import_customers.sql
```

## Notes

- Companies use `ON CONFLICT (org_number)` for upsert logic
- Private persons use `ON CONFLICT (id)` for upsert logic
- Internal Straye companies are marked with `is_internal = true`
