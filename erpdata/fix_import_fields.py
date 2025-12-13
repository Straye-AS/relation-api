#!/usr/bin/env python3
"""
Fix import SQL files to use correct field mappings:
- offer_number -> external_reference (for ERP references like "22000")
- Generate proper offer_number values (format: TK-YYYY-NNN)
- Generate proper project_number values (format: TK-YYYY-NNN)
"""

import re

def fix_import_offers():
    """Fix import_offers.sql to use external_reference instead of offer_number"""

    with open('import_offers.sql', 'r') as f:
        lines = f.readlines()

    fixed_lines = []
    sequence_counter = {}  # Track sequence per year

    for line in lines:
        # Fix the INSERT column list (appears before each VALUES)
        if "INSERT INTO offers" in line and "offer_number, cost" in line:
            line = line.replace("offer_number, cost", "external_reference, offer_number, cost")
            fixed_lines.append(line)
            continue

        # Fix the VALUES line - add the generated offer_number
        if line.startswith("VALUES ("):
            # Extract the sent_date to determine year
            sent_date_match = re.search(r"'(\d{4})-\d{2}-\d{2} \d{2}:\d{2}:\d{2}'\)", line)
            if sent_date_match:
                year = sent_date_match.group(1)
            else:
                year = "2024"  # Default

            # Find the ERP reference - it's after "NOW(), NOW(), " and before the cost value
            # Use \d* to also match empty ERP references ('')
            erp_match = re.search(r"NOW\(\), NOW\(\), '(\d*)', ([\d.]+),", line)
            if erp_match:
                erp_ref = erp_match.group(1)
                cost_val = erp_match.group(2)

                # Get next sequence number for this year
                if year not in sequence_counter:
                    sequence_counter[year] = 1
                else:
                    sequence_counter[year] += 1

                seq = sequence_counter[year]
                offer_number = f"TK-{year}-{seq:03d}"

                # Replace: NOW(), NOW(), 'ERPREF', cost -> NOW(), NOW(), 'ERPREF', 'TK-YYYY-NNN', cost
                line = line.replace(
                    f"NOW(), NOW(), '{erp_ref}', {cost_val},",
                    f"NOW(), NOW(), '{erp_ref}', '{offer_number}', {cost_val},"
                )

        fixed_lines.append(line)

    # Write fixed content
    with open('import_offers_fixed.sql', 'w') as f:
        f.writelines(fixed_lines)

    print(f"Fixed import_offers.sql -> import_offers_fixed.sql")
    print(f"Sequences generated: {sequence_counter}")


def fix_import_projects_won():
    """Fix import_projects_won.sql to generate proper project_number"""

    new_content = """-- =============================================================================
-- Straye Tak Projects Import: WON Offers
-- =============================================================================
-- Generated: 2025-12-13 (FIXED)
--
-- Creates projects from won offers:
--   - status: 'active'
--   - project_number: generated from sequence (TK-YYYY-PNN)
--   - inherited_offer_number: from winning offer's offer_number
--
-- IMPORTANT: Run AFTER import_offers_fixed.sql
-- =============================================================================

-- Create projects for won offers with proper number generation
-- Using a CTE to generate sequential project numbers
WITH numbered_offers AS (
    SELECT
        o.*,
        ROW_NUMBER() OVER (
            PARTITION BY EXTRACT(YEAR FROM COALESCE(o.sent_date, CURRENT_DATE))
            ORDER BY o.sent_date, o.id
        ) as seq_num,
        EXTRACT(YEAR FROM COALESCE(o.sent_date, CURRENT_DATE))::int as offer_year
    FROM offers o
    WHERE o.company_id = 'tak'
      AND o.phase = 'won'
)
INSERT INTO projects (
    id,
    name,
    summary,
    description,
    customer_id,
    customer_name,
    company_id,
    status,
    start_date,
    value,
    cost,
    spent,
    manager_id,
    manager_name,
    offer_id,
    project_number,
    inherited_offer_number,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() as id,
    o.title as name,
    'Prosjekt fra tilbud: ' || o.title as summary,
    o.description,
    o.customer_id,
    o.customer_name,
    o.company_id,
    'active' as status,
    o.sent_date::date as start_date,
    o.value as value,
    o.cost as cost,
    0 as spent,
    NULL as manager_id,
    o.responsible_user_name as manager_name,
    o.id as offer_id,
    'TK-' || o.offer_year::text || '-P' || LPAD(o.seq_num::text, 3, '0') as project_number,
    o.offer_number as inherited_offer_number,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM numbered_offers o;

-- Update offers to link back to their projects (also set project_name)
UPDATE offers o
SET project_id = p.id,
    project_name = p.name
FROM projects p
WHERE p.offer_id = o.id
  AND o.company_id = 'tak'
  AND o.phase = 'won';

-- =============================================================================
-- Summary
-- =============================================================================
-- Projects created from won offers
--   - All with status 'active'
--   - project_number = generated (e.g., "TK-2023-P001")
--   - inherited_offer_number = offer's internal number (e.g., "TK-2023-001")
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_won_fixed.sql
-- =============================================================================
"""

    with open('import_projects_won_fixed.sql', 'w') as f:
        f.write(new_content)

    print("Fixed import_projects_won.sql -> import_projects_won_fixed.sql")


def fix_import_projects_not_won():
    """Fix import_projects_not_won.sql - non-won projects don't need project_number"""

    new_content = """-- =============================================================================
-- Straye Tak Projects Import: NON-WON Offers
-- =============================================================================
-- Generated: 2025-12-13 (FIXED)
--
-- Creates projects from non-won offers:
--   - 'sent' offers -> status 'planning'
--   - 'in_progress' offers -> status 'planning'
--   - 'lost' offers -> status 'cancelled'
--
-- Note: Non-won projects do NOT get a project_number assigned
-- (only projects from won offers get numbered)
--
-- IMPORTANT: Run AFTER import_offers_fixed.sql
-- =============================================================================

-- Create projects for non-won offers
INSERT INTO projects (
    id,
    name,
    summary,
    description,
    customer_id,
    customer_name,
    company_id,
    status,
    start_date,
    value,
    cost,
    spent,
    manager_id,
    manager_name,
    offer_id,
    created_at,
    updated_at
)
SELECT
    gen_random_uuid() as id,
    o.title as name,
    'Prosjekt fra tilbud: ' || o.title as summary,
    o.description,
    o.customer_id,
    o.customer_name,
    o.company_id,
    CASE
        WHEN o.phase = 'lost' THEN 'cancelled'
        ELSE 'planning'
    END as status,
    o.sent_date::date as start_date,
    o.value as value,
    o.cost as cost,
    0 as spent,
    NULL as manager_id,
    o.responsible_user_name as manager_name,
    o.id as offer_id,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM offers o
WHERE o.company_id = 'tak'
  AND o.phase IN ('sent', 'in_progress', 'lost');

-- Update offers to link back to their projects (also set project_name)
UPDATE offers o
SET project_id = p.id,
    project_name = p.name
FROM projects p
WHERE p.offer_id = o.id
  AND o.company_id = 'tak'
  AND o.phase IN ('sent', 'in_progress', 'lost');

-- =============================================================================
-- Summary
-- =============================================================================
-- Projects created from non-won offers:
--   - planning: from sent + in_progress offers
--   - cancelled: from lost offers
--
-- Notes:
--   - No project_number assigned (only won projects get numbers)
--   - start_date may be NULL if offer had no sent_date
--   - Bidirectional link: project.offer_id <-> offer.project_id
--
-- To run:
-- docker exec -i relation-postgres psql -U relation_user -d relation < erpdata/import_projects_not_won_fixed.sql
-- =============================================================================
"""

    with open('import_projects_not_won_fixed.sql', 'w') as f:
        f.write(new_content)

    print("Fixed import_projects_not_won.sql -> import_projects_not_won_fixed.sql")


if __name__ == "__main__":
    import os
    os.chdir(os.path.dirname(os.path.abspath(__file__)))

    fix_import_offers()
    fix_import_projects_won()
    fix_import_projects_not_won()

    print("\nAll files fixed! New files created with '_fixed' suffix.")
    print("Review and rename to replace originals when ready.")
