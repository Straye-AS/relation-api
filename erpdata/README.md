# ERP Data Import Files

SQL scripts for importing data from external ERP systems into Relation API.

## Run Order

**IMPORTANT:** Files must be run in this specific order due to foreign key dependencies.

```bash
# 1. First: Import customers (no dependencies)
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_customers.sql

# 2. Second: Import offers (depends on customers)
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_offers.sql

# 3. Third: Create projects from non-won offers (depends on offers)
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_not_won.sql

# 4. Fourth: Create projects from won offers (depends on offers)
docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won.sql
```

## File Descriptions

| File | Description | Dependencies |
|------|-------------|--------------|
| `import_customers.sql` | Imports customer records with UPSERT on org_number | None |
| `import_offers.sql` | Imports offer records with original offer numbers | Customers |
| `import_projects_not_won.sql` | Creates projects from sent/in_progress/lost offers | Offers |
| `import_projects_won.sql` | Creates projects from won offers, adds "W" suffix | Offers |

## Numbering Convention

When an offer is won:
- **Project** gets the original offer number as both `project_number` and `inherited_offer_number`
- **Offer** gets a "W" suffix added to its `offer_number`

### Example
```
Before:  Offer "TK-2023-001" (phase: won)
After:   Offer "TK-2023-001W" -> Project "TK-2023-001"
```

## Project Phases by Offer Status

| Offer Phase | Project Phase |
|-------------|---------------|
| `sent` | `tilbud` |
| `in_progress` | `tilbud` |
| `won` | `active` |
| `lost` | `cancelled` |

## Field Inheritance

When creating projects from offers:

| Offer Field | Project Field |
|-------------|---------------|
| `title` | `name` |
| `description` | `description` |
| `customer_id` | `customer_id` |
| `customer_name` | `customer_name` |
| `company_id` | `company_id` |
| `value` | `value` |
| `cost` | `cost` |
| `responsible_user_name` | `manager_name` |
| `location` | `location` |
| `offer_number` (won only) | `project_number`, `inherited_offer_number` |

## Notes

- Scripts are idempotent for customers (UPSERT on org_number)
- Offers use `ON CONFLICT (id) DO NOTHING` to avoid duplicates
- Projects are created fresh each time - clear existing data before re-import
- Location field requires migration `00047_add_location_to_projects.sql`
