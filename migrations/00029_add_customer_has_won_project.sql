-- +goose Up
-- Add customer_has_won_project field to offers table
-- This boolean flag indicates whether Straye's customer has won their project/offer,
-- which is a positive indicator for Straye's probability of winning.

ALTER TABLE offers ADD COLUMN customer_has_won_project BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN offers.customer_has_won_project IS 'Indicates whether the customer has won their project, which increases likelihood of Straye winning';

-- +goose Down
ALTER TABLE offers DROP COLUMN customer_has_won_project;