# Sprint 1: Security Foundation Sprint - Implementation Order

**Sprint Dates:** December 6-20, 2024
**Total Stories:** 18 stories, ~55 story points

This document defines the recommended implementation order for Sprint 1 stories, organized by dependency chains and logical groupings.

---

## Phase 1: Database Foundation (Day 1-2)
*These are foundational - everything else depends on them. Likely already partially done.*

| Order | Story ID | Name | Est | Priority | Status |
|-------|----------|------|-----|----------|--------|
| 1 | [sc-51](https://app.shortcut.com/straye/story/51) | Setup PostgreSQL database with connection pooling | 2 | Highest | Validate |
| 2 | [sc-52](https://app.shortcut.com/straye/story/52) | Implement Goose migration system | 2 | Highest | Validate |

**Notes:**
- These are likely already implemented based on existing codebase
- Validate against acceptance criteria and mark as Done if complete
- If incomplete, finish before proceeding

---

## Phase 2: Schema Migrations (Day 2-4)
*Run migrations in sequence - each depends on the previous.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 3 | [sc-53](https://app.shortcut.com/straye/story/53) | Create initial schema migration (companies, users, customers) | 8 | Highest | sc-52 |
| 4 | [sc-54](https://app.shortcut.com/straye/story/54) | Create deals and contacts migration | 4 | High | sc-53 |
| 5 | [sc-55](https://app.shortcut.com/straye/story/55) | Create budget dimensions migration | 4 | High | sc-53 |
| 6 | [sc-56](https://app.shortcut.com/straye/story/56) | Create project enhancements and actual costs migration | 4 | High | sc-54 |
| 7 | [sc-57](https://app.shortcut.com/straye/story/57) | Create activities and permissions migration | 4 | Highest | sc-53 |

**Notes:**
- Migrations must run in order (00002, 00003, 00004, 00005, 00006)
- sc-53 is a larger migration - may partially exist
- sc-57 creates user_roles, user_permissions, audit_logs tables needed for auth

---

## Phase 3: Auth Infrastructure (Day 4-6)
*Build auth system on top of database schema.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 8 | [sc-59](https://app.shortcut.com/straye/story/59) | Implement Azure AD JWT token validation | 4 | Highest | sc-57 |
| 9 | [sc-60](https://app.shortcut.com/straye/story/60) | Implement API Key authentication | 2 | Highest | sc-57 |
| 10 | [sc-61](https://app.shortcut.com/straye/story/61) | Create authentication middleware | 2 | Highest | sc-59, sc-60 |
| 11 | [sc-65](https://app.shortcut.com/straye/story/65) | Implement user context helpers and utilities | 1 | Medium | sc-61 |
| 12 | [sc-62](https://app.shortcut.com/straye/story/62) | Implement multi-tenant data isolation | 4 | Medium | sc-65 |
| 13 | [sc-63](https://app.shortcut.com/straye/story/63) | Implement RBAC permission system | 4 | Low | sc-57, sc-65 |
| 14 | [sc-80](https://app.shortcut.com/straye/story/80) | Implement user role and permission management | 4 | Low | sc-63 |
| 15 | [sc-64](https://app.shortcut.com/straye/story/64) | Create auth endpoints (GET /auth/me, GET /auth/permissions) | 2 | Highest | sc-63, sc-80 |

**Notes:**
- sc-59 implements JWT validation for Azure AD tokens
- sc-60 implements API key auth for system-to-system integration
- sc-61 combines both auth methods into unified middleware
- sc-65 is small but foundational - UserContext used everywhere
- sc-62 adds company filtering middleware
- sc-63 and sc-80 work together for full RBAC
- sc-64 exposes auth info via API endpoints

---

## Phase 4: Security Hardening (Day 6-8)
*Can be done in parallel or after auth. Independent of each other.*

| Order | Story ID | Name | Est | Priority | Depends On |
|-------|----------|------|-----|----------|------------|
| 16 | [sc-138](https://app.shortcut.com/straye/story/138) | Implement CORS configuration and security headers | 2 | Highest | None |
| 17 | [sc-139](https://app.shortcut.com/straye/story/139) | Implement rate limiting and request throttling | 4 | Highest | None |
| 18 | [sc-140](https://app.shortcut.com/straye/story/140) | Implement audit logging for all modifications | 4 | Highest | sc-57 |

**Notes:**
- sc-138 and sc-139 are middleware - can be done anytime
- sc-140 depends on audit_logs table from sc-57
- These provide security hardening for production readiness

---

## Dependency Graph

```
sc-51 (PostgreSQL)
  └── sc-52 (Goose migrations)
        └── sc-53 (companies, users, customers)
              ├── sc-54 (deals, contacts)
              │     └── sc-56 (project enhancements)
              ├── sc-55 (budget dimensions)
              └── sc-57 (activities, permissions)
                    ├── sc-59 (JWT validation)
                    │     └── sc-61 (auth middleware)
                    ├── sc-60 (API key auth)
                    │     └── sc-61 (auth middleware)
                    │           └── sc-65 (user context)
                    │                 ├── sc-62 (multi-tenant)
                    │                 └── sc-63 (RBAC)
                    │                       └── sc-80 (role management)
                    │                             └── sc-64 (auth endpoints)
                    └── sc-140 (audit logging)

Independent:
  sc-138 (CORS/headers)
  sc-139 (rate limiting)
```

---

## Quick Reference: Story Details

| ID | Name | Epic | Est | Priority |
|----|------|------|-----|----------|
| sc-51 | Setup PostgreSQL database with connection pooling | Database Infrastructure | 2 | Highest |
| sc-52 | Implement Goose migration system | Database Infrastructure | 2 | Highest |
| sc-53 | Create initial schema migration (companies, users, customers) | Database Infrastructure | 8 | Highest |
| sc-54 | Create deals and contacts migration | Database Infrastructure | 4 | High |
| sc-55 | Create budget dimensions migration | Database Infrastructure | 4 | High |
| sc-56 | Create project enhancements and actual costs migration | Database Infrastructure | 4 | High |
| sc-57 | Create activities and permissions migration | Database Infrastructure | 4 | Highest |
| sc-59 | Implement Azure AD JWT token validation | Auth & Authorization | 4 | Highest |
| sc-60 | Implement API Key authentication | Auth & Authorization | 2 | Highest |
| sc-61 | Create authentication middleware | Auth & Authorization | 2 | Highest |
| sc-62 | Implement multi-tenant data isolation | Auth & Authorization | 4 | Medium |
| sc-63 | Implement RBAC permission system | Auth & Authorization | 4 | Low |
| sc-64 | Create auth endpoints (GET /auth/me, GET /auth/permissions) | Auth & Authorization | 2 | Highest |
| sc-65 | Implement user context helpers and utilities | Auth & Authorization | 1 | Medium |
| sc-80 | Implement user role and permission management | Company & User Mgmt | 4 | Low |
| sc-138 | Implement CORS configuration and security headers | Security & Compliance | 2 | Highest |
| sc-139 | Implement rate limiting and request throttling | Security & Compliance | 4 | Highest |
| sc-140 | Implement audit logging for all modifications | Security & Compliance | 4 | Highest |

---

## Implementation Tips

### Grouping for PRs
Stories can be grouped into logical PRs:

1. **PR 1: Database Foundation** - sc-51, sc-52 (validate existing)
2. **PR 2: Core Migrations** - sc-53, sc-54, sc-55, sc-56, sc-57
3. **PR 3: Core Authentication** - sc-59, sc-60, sc-61 (JWT, API key, middleware)
4. **PR 4: Auth Extensions** - sc-65, sc-62, sc-63, sc-80, sc-64 (context, RBAC, endpoints)
5. **PR 5: Security Hardening** - sc-138, sc-139, sc-140

### Parallel Work Opportunities
- sc-138 and sc-139 can be done anytime (no dependencies)
- sc-54 and sc-55 can be done in parallel (both depend on sc-53)
- sc-59 and sc-60 can be done in parallel (both depend on sc-57)
- After Phase 2, security hardening can run parallel to auth work

### Validation Points
After each phase, validate:
1. **Phase 1:** `make migrate-status` shows migrations applied
2. **Phase 2:** All tables exist, test data can be inserted
3. **Phase 3:** Auth endpoints return correct user/permissions
4. **Phase 4:** Security headers present, rate limits enforced

---

*Document updated: 2025-12-06*
*Sprint: Security Foundation Sprint (Iteration 151)*
