-- +goose Up
-- Add target_name column to activities table
-- This stores the name of the entity being tracked (e.g., customer name, offer title)
-- to enable more descriptive activity messages on the frontend

ALTER TABLE activities ADD COLUMN IF NOT EXISTS target_name VARCHAR(255);

-- Backfill target_name from existing data where possible
-- Update activities for customers
UPDATE activities a
SET target_name = c.name
FROM customers c
WHERE a.target_type = 'Customer' AND a.target_id = c.id AND a.target_name IS NULL;

-- Update activities for contacts
UPDATE activities a
SET target_name = CONCAT(c.first_name, ' ', c.last_name)
FROM contacts c
WHERE a.target_type = 'Contact' AND a.target_id = c.id AND a.target_name IS NULL;

-- Update activities for offers
UPDATE activities a
SET target_name = o.title
FROM offers o
WHERE a.target_type = 'Offer' AND a.target_id = o.id AND a.target_name IS NULL;

-- Update activities for projects
UPDATE activities a
SET target_name = p.name
FROM projects p
WHERE a.target_type = 'Project' AND a.target_id = p.id AND a.target_name IS NULL;

-- Update activities for deals
UPDATE activities a
SET target_name = d.title
FROM deals d
WHERE a.target_type = 'Deal' AND a.target_id = d.id AND a.target_name IS NULL;

-- Update activities for files
UPDATE activities a
SET target_name = f.filename
FROM files f
WHERE a.target_type = 'File' AND a.target_id = f.id AND a.target_name IS NULL;

-- +goose Down
ALTER TABLE activities DROP COLUMN IF EXISTS target_name;
