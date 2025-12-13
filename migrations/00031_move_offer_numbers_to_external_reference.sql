-- +goose Up
-- +goose StatementBegin

-- This migration:
-- 1. Moves current offer_number values to external_reference (these are customer/external references)
-- 2. Clears offer_number field
-- 3. Regenerates offer_numbers using the internal sequence (e.g., "TK-2025-001")
-- Note: The sequence is shared with projects, so new numbers continue from the existing sequence

-- Step 1: Copy ALL offer_number values to external_reference (preserving customer references)
-- Skip if external_reference already has a value
UPDATE offers
SET external_reference = offer_number
WHERE (external_reference IS NULL OR external_reference = '')
  AND offer_number IS NOT NULL
  AND offer_number != '';

-- Step 2: Clear all offer_numbers to prepare for regeneration
UPDATE offers SET offer_number = NULL;

-- Step 3: Assign new sequential numbers to all non-draft offers
-- Uses the shared sequence with projects to avoid conflicts
CREATE OR REPLACE FUNCTION reassign_offer_numbers() RETURNS void AS $$
DECLARE
    rec RECORD;
    current_seq INT;
    prefix TEXT;
BEGIN
    -- Process each non-draft offer ordered by created_at to maintain chronological order
    FOR rec IN (
        SELECT id, company_id, EXTRACT(YEAR FROM created_at)::INT as year, created_at
        FROM offers
        WHERE phase != 'draft'
        ORDER BY company_id, year, created_at
    ) LOOP
        -- Skip if company_id is invalid
        IF rec.company_id IS NULL OR rec.company_id = '' THEN
            CONTINUE;
        END IF;

        -- Determine prefix based on company
        prefix := CASE rec.company_id
            WHEN 'stalbygg' THEN 'ST'
            WHEN 'hybridbygg' THEN 'HB'
            WHEN 'industri' THEN 'IN'
            WHEN 'tak' THEN 'TK'
            WHEN 'montasje' THEN 'MO'
            WHEN 'gruppen' THEN 'GR'
            ELSE 'GR'
        END;

        -- Get current sequence for this company/year (shared with projects)
        SELECT last_sequence INTO current_seq
        FROM number_sequences
        WHERE company_id = rec.company_id AND year = rec.year
        FOR UPDATE;

        IF NOT FOUND THEN
            -- No sequence exists yet, start at 0
            current_seq := 0;
            INSERT INTO number_sequences (company_id, year, last_sequence, created_at, updated_at)
            VALUES (rec.company_id, rec.year, 0, NOW(), NOW());
        END IF;

        -- Increment sequence
        current_seq := current_seq + 1;

        -- Update the sequence counter
        UPDATE number_sequences
        SET last_sequence = current_seq, updated_at = NOW()
        WHERE company_id = rec.company_id AND year = rec.year;

        -- Assign new internal offer number
        UPDATE offers
        SET offer_number = prefix || '-' || rec.year::TEXT || '-' || LPAD(current_seq::TEXT, 3, '0')
        WHERE id = rec.id;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Execute the function
SELECT reassign_offer_numbers();

-- Drop the function after use
DROP FUNCTION IF EXISTS reassign_offer_numbers();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- This cannot be fully reversed as the original offer_numbers are now in external_reference
-- and the sequence has been updated. Manual intervention would be required.

-- If needed, you can swap back:
-- UPDATE offers SET offer_number = external_reference WHERE external_reference IS NOT NULL;
-- UPDATE offers SET external_reference = NULL;

-- +goose StatementEnd
