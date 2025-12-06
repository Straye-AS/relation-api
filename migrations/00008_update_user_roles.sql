-- +goose Up

-- ============================================================================
-- UPDATE USER ROLE ENUM
-- Replace 'sales' with 'market' and add 'project_leader'
-- ============================================================================

-- First, drop all views that depend on the user_role type
DROP VIEW IF EXISTS role_default_permissions;

-- +goose StatementBegin

-- PostgreSQL doesn't allow direct modification of enum values
-- We need to create a new type, migrate data, and replace

-- Create new enum type with updated values
CREATE TYPE user_role_new AS ENUM (
    'super_admin',
    'company_admin',
    'manager',
    'market',
    'project_manager',
    'project_leader',
    'viewer',
    'api_service'
);

-- Update user_roles table to use the new enum
ALTER TABLE user_roles
    ALTER COLUMN role TYPE user_role_new
    USING (
        CASE role::text
            WHEN 'sales' THEN 'market'::user_role_new
            ELSE role::text::user_role_new
        END
    );

-- Drop old enum type
DROP TYPE user_role;

-- Rename new enum to original name
ALTER TYPE user_role_new RENAME TO user_role;

-- +goose StatementEnd

-- ============================================================================
-- RECREATE ROLE DEFAULT PERMISSIONS VIEW
-- ============================================================================

-- +goose StatementBegin

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
    'market'::user_role,
    'Marketing and sales operations',
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
    'project_leader'::user_role,
    'Project leadership and execution',
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

-- First, drop views that depend on the user_role type
DROP VIEW IF EXISTS role_default_permissions;

-- +goose StatementBegin

-- Revert to original enum (sales instead of market, no project_leader)
CREATE TYPE user_role_old AS ENUM (
    'super_admin',
    'company_admin',
    'manager',
    'sales',
    'project_manager',
    'viewer',
    'api_service'
);

-- Update user_roles table to use the old enum
ALTER TABLE user_roles
    ALTER COLUMN role TYPE user_role_old
    USING (
        CASE role::text
            WHEN 'market' THEN 'sales'::user_role_old
            WHEN 'project_leader' THEN 'project_manager'::user_role_old
            ELSE role::text::user_role_old
        END
    );

-- Drop new enum type
DROP TYPE user_role;

-- Rename old enum back
ALTER TYPE user_role_old RENAME TO user_role;

-- +goose StatementEnd

-- +goose StatementBegin

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
