-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- ACTIVITY TABLE INDEX OPTIMIZATION
-- ============================================================================
-- This migration adds indexes to improve query performance on the activities
-- table for common access patterns identified from QA reports.
--
-- Existing indexes (for reference):
--   idx_activities_target (target_type, target_id)
--   idx_activities_occurred_at (occurred_at DESC)
--   idx_activities_target_type_date (target_type, target_id, occurred_at DESC)
--   idx_activities_creator_id (creator_id)
--   idx_activities_company_id (company_id)
--   idx_activities_activity_type (activity_type)
--   idx_activities_status (status)
--   idx_activities_scheduled_at (scheduled_at)
--   idx_activities_due_date (due_date)
--   idx_activities_assigned_to (assigned_to_id)
-- ============================================================================

-- Index on created_at for time-based sorting and filtering
-- Complements the existing occurred_at index for audit/creation time queries
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);

-- Composite index for multi-tenant time-based queries
-- Optimizes: "Get all activities for company X in time range"
CREATE INDEX idx_activities_company_created ON activities(company_id, created_at DESC);

-- Composite index for user activity history with time ordering
-- Optimizes: "Get all activities by user X ordered by time"
CREATE INDEX idx_activities_creator_created ON activities(creator_id, created_at DESC);

-- Composite index for assigned user queries with time ordering
-- Optimizes: "Get all activities assigned to user X ordered by time"
CREATE INDEX idx_activities_assigned_created ON activities(assigned_to_id, created_at DESC);

-- Composite index for company-scoped target entity queries
-- Optimizes: "Get all activities for entity type X in company Y"
CREATE INDEX idx_activities_company_target ON activities(company_id, target_type, target_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_activities_company_target;
DROP INDEX IF EXISTS idx_activities_assigned_created;
DROP INDEX IF EXISTS idx_activities_creator_created;
DROP INDEX IF EXISTS idx_activities_company_created;
DROP INDEX IF EXISTS idx_activities_created_at;

-- +goose StatementEnd
