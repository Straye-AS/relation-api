-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- ACTIVITY TYPE ENUM
-- ============================================================================

CREATE TYPE activity_type AS ENUM (
    'meeting',
    'call',
    'email',
    'task',
    'note',
    'system'
);

-- ============================================================================
-- ACTIVITY STATUS ENUM
-- ============================================================================

CREATE TYPE activity_status AS ENUM (
    'planned',
    'in_progress',
    'completed',
    'cancelled'
);

-- ============================================================================
-- ENHANCE ACTIVITIES TABLE
-- ============================================================================

-- Add new columns to activities table
ALTER TABLE activities
    ADD COLUMN activity_type activity_type NOT NULL DEFAULT 'note',
    ADD COLUMN status activity_status NOT NULL DEFAULT 'completed',
    ADD COLUMN scheduled_at TIMESTAMP,
    ADD COLUMN due_date DATE,
    ADD COLUMN completed_at TIMESTAMP,
    ADD COLUMN duration_minutes INT,
    ADD COLUMN priority INT DEFAULT 0,
    ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN creator_id VARCHAR(100),
    ADD COLUMN assigned_to_id VARCHAR(100),
    ADD COLUMN company_id VARCHAR(50) REFERENCES companies(id);

-- Create indexes for new activity columns
CREATE INDEX idx_activities_activity_type ON activities(activity_type);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_scheduled_at ON activities(scheduled_at);
CREATE INDEX idx_activities_due_date ON activities(due_date);
CREATE INDEX idx_activities_creator_id ON activities(creator_id);
CREATE INDEX idx_activities_assigned_to ON activities(assigned_to_id);
CREATE INDEX idx_activities_company_id ON activities(company_id);

-- Composite index for common queries
CREATE INDEX idx_activities_target_type_date ON activities(target_type, target_id, occurred_at DESC);

-- ============================================================================
-- ROLE ENUM
-- ============================================================================

CREATE TYPE user_role AS ENUM (
    'super_admin',
    'company_admin',
    'manager',
    'sales',
    'project_manager',
    'viewer',
    'api_service'
);

-- ============================================================================
-- USER ROLES JUNCTION TABLE
-- ============================================================================

CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role user_role NOT NULL,
    company_id VARCHAR(50) REFERENCES companies(id) ON DELETE CASCADE,
    granted_by VARCHAR(100),
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint: user can have each role once per company (or once globally if company_id is null)
    CONSTRAINT uq_user_role_company UNIQUE (user_id, role, company_id)
);

-- Create indexes for user_roles
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_company_id ON user_roles(company_id);
CREATE INDEX idx_user_roles_role ON user_roles(role);
CREATE INDEX idx_user_roles_is_active ON user_roles(is_active);

-- Trigger for user_roles updated_at
CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE ON user_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- PERMISSION ENUM
-- ============================================================================

CREATE TYPE permission_type AS ENUM (
    -- Customer permissions
    'customers:read',
    'customers:write',
    'customers:delete',

    -- Contact permissions
    'contacts:read',
    'contacts:write',
    'contacts:delete',

    -- Deal permissions
    'deals:read',
    'deals:write',
    'deals:delete',

    -- Offer permissions
    'offers:read',
    'offers:write',
    'offers:delete',
    'offers:approve',

    -- Project permissions
    'projects:read',
    'projects:write',
    'projects:delete',

    -- Budget permissions
    'budgets:read',
    'budgets:write',

    -- Activity permissions
    'activities:read',
    'activities:write',
    'activities:delete',

    -- User management permissions
    'users:read',
    'users:write',
    'users:manage_roles',

    -- Company management permissions
    'companies:read',
    'companies:write',

    -- Reports and analytics
    'reports:view',
    'reports:export',

    -- System administration
    'system:admin',
    'system:audit_logs'
);

-- ============================================================================
-- USER PERMISSIONS TABLE (for overrides)
-- ============================================================================

CREATE TABLE user_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(100) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission permission_type NOT NULL,
    company_id VARCHAR(50) REFERENCES companies(id) ON DELETE CASCADE,
    is_granted BOOLEAN NOT NULL DEFAULT true,  -- true = grant, false = deny (override)
    granted_by VARCHAR(100),
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Unique constraint: user can have each permission once per company
    CONSTRAINT uq_user_permission_company UNIQUE (user_id, permission, company_id)
);

-- Create indexes for user_permissions
CREATE INDEX idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX idx_user_permissions_company_id ON user_permissions(company_id);
CREATE INDEX idx_user_permissions_permission ON user_permissions(permission);
CREATE INDEX idx_user_permissions_is_granted ON user_permissions(is_granted);

-- Trigger for user_permissions updated_at
CREATE TRIGGER update_user_permissions_updated_at BEFORE UPDATE ON user_permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- AUDIT LOG ACTION ENUM
-- ============================================================================

CREATE TYPE audit_action AS ENUM (
    'create',
    'update',
    'delete',
    'login',
    'logout',
    'permission_grant',
    'permission_revoke',
    'role_assign',
    'role_remove',
    'export',
    'import',
    'api_call'
);

-- ============================================================================
-- AUDIT LOGS TABLE
-- ============================================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Who performed the action
    user_id VARCHAR(100),
    user_email VARCHAR(255),
    user_name VARCHAR(200),

    -- What action was performed
    action audit_action NOT NULL,

    -- What entity was affected
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    entity_name VARCHAR(200),

    -- Context
    company_id VARCHAR(50),

    -- Change details
    old_values JSONB,
    new_values JSONB,
    changes JSONB,

    -- Request metadata
    ip_address INET,
    user_agent TEXT,
    request_id VARCHAR(100),

    -- Additional context
    metadata JSONB,

    -- Timestamps
    performed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for audit_logs (optimized for common queries)
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_company_id ON audit_logs(company_id);
CREATE INDEX idx_audit_logs_performed_at ON audit_logs(performed_at DESC);

-- Composite index for time-range queries per entity
CREATE INDEX idx_audit_logs_entity_time ON audit_logs(entity_type, entity_id, performed_at DESC);

-- Composite index for user activity queries
CREATE INDEX idx_audit_logs_user_time ON audit_logs(user_id, performed_at DESC);

-- ============================================================================
-- ROLE DEFAULT PERMISSIONS VIEW
-- ============================================================================

-- This view defines which permissions each role has by default
-- Actual implementation will be in Go code, this is for documentation
CREATE VIEW role_default_permissions AS
SELECT
    'super_admin'::user_role as role,
    'All permissions'::text as description,
    ARRAY[
        'customers:read', 'customers:write', 'customers:delete',
        'contacts:read', 'contacts:write', 'contacts:delete',
        'deals:read', 'deals:write', 'deals:delete',
        'offers:read', 'offers:write', 'offers:delete', 'offers:approve',
        'projects:read', 'projects:write', 'projects:delete',
        'budgets:read', 'budgets:write',
        'activities:read', 'activities:write', 'activities:delete',
        'users:read', 'users:write', 'users:manage_roles',
        'companies:read', 'companies:write',
        'reports:view', 'reports:export',
        'system:admin', 'system:audit_logs'
    ]::text[] as permissions
UNION ALL
SELECT
    'company_admin'::user_role,
    'Company-level administration',
    ARRAY[
        'customers:read', 'customers:write', 'customers:delete',
        'contacts:read', 'contacts:write', 'contacts:delete',
        'deals:read', 'deals:write', 'deals:delete',
        'offers:read', 'offers:write', 'offers:delete', 'offers:approve',
        'projects:read', 'projects:write', 'projects:delete',
        'budgets:read', 'budgets:write',
        'activities:read', 'activities:write', 'activities:delete',
        'users:read', 'users:write', 'users:manage_roles',
        'reports:view', 'reports:export',
        'system:audit_logs'
    ]::text[]
UNION ALL
SELECT
    'manager'::user_role,
    'Team management and oversight',
    ARRAY[
        'customers:read', 'customers:write',
        'contacts:read', 'contacts:write',
        'deals:read', 'deals:write',
        'offers:read', 'offers:write', 'offers:approve',
        'projects:read', 'projects:write',
        'budgets:read', 'budgets:write',
        'activities:read', 'activities:write',
        'users:read',
        'reports:view', 'reports:export'
    ]::text[]
UNION ALL
SELECT
    'sales'::user_role,
    'Sales operations',
    ARRAY[
        'customers:read', 'customers:write',
        'contacts:read', 'contacts:write',
        'deals:read', 'deals:write',
        'offers:read', 'offers:write',
        'projects:read',
        'budgets:read',
        'activities:read', 'activities:write',
        'reports:view'
    ]::text[]
UNION ALL
SELECT
    'project_manager'::user_role,
    'Project management operations',
    ARRAY[
        'customers:read',
        'contacts:read', 'contacts:write',
        'deals:read',
        'offers:read',
        'projects:read', 'projects:write',
        'budgets:read', 'budgets:write',
        'activities:read', 'activities:write',
        'reports:view'
    ]::text[]
UNION ALL
SELECT
    'viewer'::user_role,
    'Read-only access',
    ARRAY[
        'customers:read',
        'contacts:read',
        'deals:read',
        'offers:read',
        'projects:read',
        'budgets:read',
        'activities:read',
        'reports:view'
    ]::text[]
UNION ALL
SELECT
    'api_service'::user_role,
    'API service account',
    ARRAY[
        'customers:read', 'customers:write',
        'contacts:read', 'contacts:write',
        'deals:read', 'deals:write',
        'offers:read', 'offers:write',
        'projects:read', 'projects:write',
        'budgets:read', 'budgets:write',
        'activities:read', 'activities:write'
    ]::text[];

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop view
DROP VIEW IF EXISTS role_default_permissions;

-- Drop audit_logs table
DROP TABLE IF EXISTS audit_logs;

-- Drop audit action enum
DROP TYPE IF EXISTS audit_action;

-- Drop user_permissions table
DROP TRIGGER IF EXISTS update_user_permissions_updated_at ON user_permissions;
DROP TABLE IF EXISTS user_permissions;

-- Drop permission enum
DROP TYPE IF EXISTS permission_type;

-- Drop user_roles table
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;
DROP TABLE IF EXISTS user_roles;

-- Drop role enum
DROP TYPE IF EXISTS user_role;

-- Remove indexes from activities
DROP INDEX IF EXISTS idx_activities_target_type_date;
DROP INDEX IF EXISTS idx_activities_company_id;
DROP INDEX IF EXISTS idx_activities_assigned_to;
DROP INDEX IF EXISTS idx_activities_creator_id;
DROP INDEX IF EXISTS idx_activities_due_date;
DROP INDEX IF EXISTS idx_activities_scheduled_at;
DROP INDEX IF EXISTS idx_activities_status;
DROP INDEX IF EXISTS idx_activities_activity_type;

-- Remove new columns from activities
ALTER TABLE activities
    DROP COLUMN IF EXISTS company_id,
    DROP COLUMN IF EXISTS assigned_to_id,
    DROP COLUMN IF EXISTS creator_id,
    DROP COLUMN IF EXISTS is_private,
    DROP COLUMN IF EXISTS priority,
    DROP COLUMN IF EXISTS duration_minutes,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS due_date,
    DROP COLUMN IF EXISTS scheduled_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS activity_type;

-- Drop enums
DROP TYPE IF EXISTS activity_status;
DROP TYPE IF EXISTS activity_type;

-- +goose StatementEnd
