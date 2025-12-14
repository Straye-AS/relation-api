# Frontend Integration: User Tracking Fields

## Overview

The API now includes user tracking fields on Customer, Project, Offer, and Contact entities. These fields show who created and last modified each record.

## New Response Fields

All four entities now include these fields in API responses:

```typescript
interface UserTracking {
  createdById?: string;    // User ID who created the entity
  createdByName?: string;  // Display name at creation time
  updatedById?: string;    // User ID who last modified
  updatedByName?: string;  // Display name at last modification
}
```

## Affected Endpoints

| Entity | Endpoints |
|--------|-----------|
| Customer | `GET /customers`, `GET /customers/:id`, `POST /customers`, `PUT /customers/:id` |
| Project | `GET /projects`, `GET /projects/:id`, `POST /projects`, `PUT /projects/:id` |
| Offer | `GET /offers`, `GET /offers/:id`, `POST /offers`, `PUT /offers/:id` |
| Contact | `GET /contacts`, `GET /contacts/:id`, `POST /contacts`, `PUT /contacts/:id` |

## Example Response

```json
{
  "id": "a1b2c3d4-...",
  "name": "Bygg AS",
  "createdAt": "2025-01-15T10:30:00Z",
  "updatedAt": "2025-01-20T14:22:00Z",
  "createdById": "user-123",
  "createdByName": "Ola Nordmann",
  "updatedById": "user-456",
  "updatedByName": "Kari Hansen"
}
```

## Handling Null Values

Historical records (created before this feature) will have `null` values for all tracking fields. The frontend should handle this gracefully:

```tsx
// Example: Display created by info
const CreatedByInfo = ({ createdByName, createdAt }) => {
  if (createdByName) {
    return <span>Opprettet av {createdByName}</span>;
  }
  return <span>Opprettet {formatDate(createdAt)}</span>;
};
```

## Suggested UI Patterns

### 1. Entity Detail View
Display tracking info in a metadata section:

```
Opprettet: 15. jan 2025 av Ola Nordmann
Sist endret: 20. jan 2025 av Kari Hansen
```

### 2. List/Table View
Add optional columns or hover tooltips:

| Navn | Opprettet av | Sist endret |
|------|--------------|-------------|
| Bygg AS | Ola Nordmann | Kari Hansen |

### 3. Activity Indicator
Show "Sist endret av [name]" as subtle text below entity title.

## TypeScript Types

Add to your API types:

```typescript
interface CustomerDTO {
  // ... existing fields
  createdById?: string;
  createdByName?: string;
  updatedById?: string;
  updatedByName?: string;
}

interface ProjectDTO {
  // ... existing fields
  createdById?: string;
  createdByName?: string;
  updatedById?: string;
  updatedByName?: string;
}

interface OfferDTO {
  // ... existing fields
  createdById?: string;
  createdByName?: string;
  updatedById?: string;
  updatedByName?: string;
}

interface ContactDTO {
  // ... existing fields
  createdById?: string;
  createdByName?: string;
  updatedById?: string;
  updatedByName?: string;
}
```

## Notes

- Fields are automatically populated by the API - no changes needed to create/update requests
- `createdBy` fields are immutable (set once, never change)
- `updatedBy` fields update on every modification
- Names are denormalized (stored at time of action) so they remain accurate even if user changes their display name later
