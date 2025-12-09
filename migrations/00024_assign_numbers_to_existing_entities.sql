-- +goose Up
-- +goose StatementBegin

-- This migration assigns numbers to existing offers and projects that don't have them yet.
-- Numbers are assigned in chronological order (by created_at) to maintain consistency.
-- Draft offers are explicitly set to NULL (they should never have a number).
-- The shared sequence is updated to account for all assigned numbers.

-- First, ensure draft offers have NULL offer_number (business rule: drafts never have numbers)
UPDATE offers SET offer_number = NULL WHERE phase = 'draft' AND offer_number IS NOT NULL;

-- Create a function to assign numbers to existing entities
-- This processes offers and projects together to maintain a shared sequence
CREATE OR REPLACE FUNCTION assign_entity_numbers() RETURNS void AS $$
DECLARE
    rec RECORD;
    current_seq INT;
    company_year_key TEXT;
    processed_keys TEXT[] := ARRAY[]::TEXT[];
BEGIN
    -- Process each company/year combination
    FOR rec IN (
        -- Get all non-draft offers and all projects ordered by created_at
        -- Union them to process chronologically
        SELECT 'offer' as entity_type, id, company_id, EXTRACT(YEAR FROM created_at)::INT as year, created_at, offer_number as current_number
        FROM offers
        WHERE phase != 'draft'
        UNION ALL
        SELECT 'project' as entity_type, id, company_id, EXTRACT(YEAR FROM created_at)::INT as year, created_at, project_number as current_number
        FROM projects
        WHERE company_id IS NOT NULL AND company_id != ''
        ORDER BY company_id, year, created_at
    ) LOOP
        -- Skip if already has a number assigned
        IF rec.current_number IS NOT NULL AND rec.current_number != '' THEN
            CONTINUE;
        END IF;

        -- Skip if company_id is invalid
        IF rec.company_id IS NULL OR rec.company_id = '' THEN
            CONTINUE;
        END IF;

        company_year_key := rec.company_id || '-' || rec.year::TEXT;

        -- Get or create sequence for this company/year
        SELECT last_sequence INTO current_seq
        FROM number_sequences
        WHERE company_id = rec.company_id AND year = rec.year
        FOR UPDATE;

        IF NOT FOUND THEN
            current_seq := 0;
            INSERT INTO number_sequences (company_id, year, last_sequence, created_at, updated_at)
            VALUES (rec.company_id, rec.year, 0, NOW(), NOW());
        END IF;

        -- Increment sequence
        current_seq := current_seq + 1;

        -- Update the sequence
        UPDATE number_sequences
        SET last_sequence = current_seq, updated_at = NOW()
        WHERE company_id = rec.company_id AND year = rec.year;

        -- Assign number to entity
        IF rec.entity_type = 'offer' THEN
            UPDATE offers
            SET offer_number = CASE rec.company_id
                WHEN 'stalbygg' THEN 'ST'
                WHEN 'hybridbygg' THEN 'HB'
                WHEN 'industri' THEN 'IN'
                WHEN 'tak' THEN 'TK'
                WHEN 'montasje' THEN 'MO'
                WHEN 'gruppen' THEN 'GR'
                ELSE 'GR'
            END || '-' || rec.year::TEXT || '-' || LPAD(current_seq::TEXT, 3, '0')
            WHERE id = rec.id;
        ELSIF rec.entity_type = 'project' THEN
            UPDATE projects
            SET project_number = CASE rec.company_id
                WHEN 'stalbygg' THEN 'ST'
                WHEN 'hybridbygg' THEN 'HB'
                WHEN 'industri' THEN 'IN'
                WHEN 'tak' THEN 'TK'
                WHEN 'montasje' THEN 'MO'
                WHEN 'gruppen' THEN 'GR'
                ELSE 'GR'
            END || '-' || rec.year::TEXT || '-' || LPAD(current_seq::TEXT, 3, '0')
            WHERE id = rec.id;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Execute the function
SELECT assign_entity_numbers();

-- Drop the function after use
DROP FUNCTION IF EXISTS assign_entity_numbers();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Note: We cannot reliably reverse this migration as we don't know which entities
-- had manually assigned numbers vs auto-assigned ones.
-- This is a data migration that's safe to leave in place.

-- However, we can clear auto-assigned numbers if needed (this should be done manually if required)
-- UPDATE offers SET offer_number = NULL WHERE phase != 'draft';
-- UPDATE projects SET project_number = NULL;
-- TRUNCATE number_sequences;

-- +goose StatementEnd
