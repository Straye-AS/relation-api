# Frontend Integration: Offer-Project Business Rules (sc-289)

## Executive Summary

This release fundamentally improves how offers and projects work together. The key insight is that **a project is like a folder for offers** - multiple offers can compete for the same project, but only one can win. When an offer wins, the project "graduates" from bidding phase to active work, inheriting key data from the winning offer.

---

## Table of Contents

1. [The Big Picture: Why These Changes?](#1-the-big-picture-why-these-changes)
2. [New Project Fields](#2-new-project-fields)
3. [Offer-Project Linking Rules](#3-offer-project-linking-rules)
4. [Auto-Expire Competing Offers](#4-auto-expire-competing-offers)
5. [Numbering Convention](#5-numbering-convention)
6. [Smart Field Inheritance](#6-smart-field-inheritance)
7. [What NOT to Change](#7-what-not-to-change)
8. [API Changes Reference](#8-api-changes-reference)
9. [Frontend Implementation Guide](#9-frontend-implementation-guide)

---

## 1. The Big Picture: Why These Changes?

### The Problem We're Solving

Previously, the relationship between offers and projects was loose:
- Offers could be linked to projects at any stage
- When one offer won, other competing offers stayed in limbo
- Project data (manager, location, etc.) had to be manually entered even when the offer already had it
- No clear numbering system tied offers to their resulting projects

### The Solution: Project as an "Offer Folder"

Think of a project in "tilbud" phase as a **container for competing offers**:

```
Project: "Nytt lagerbygg Fredrikstad" (phase: tilbud)
├── Offer A: TK-2025-001 (sent, 50% probability)
├── Offer B: TK-2025-002 (sent, 30% probability)
└── Offer C: TK-2025-003 (draft)
```

When Offer A wins:
```
Project: "Nytt lagerbygg Fredrikstad" (phase: active, project_number: TK-2025-001)
├── Offer A: TK-2025-001W (won) ← "W" = Won
├── Offer B: TK-2025-002 (expired) ← Auto-expired
└── Offer C: TK-2025-003 (expired) ← Auto-expired
```

**Benefits:**
- Clear lifecycle management
- Automatic cleanup of competing offers
- Data flows naturally from offer to project
- Unified numbering system

---

## 2. New Project Fields

### `location` (string, optional)

**Why:** Construction projects always have a physical location. Previously this was only on offers, so when an offer became a project, the location information was lost.

**How to use:**
```tsx
// Display in project header
<ProjectHeader>
  <h1>{project.name}</h1>
  {project.location && (
    <span className="location">
      <MapPinIcon /> {project.location}
    </span>
  )}
</ProjectHeader>

// Use for filtering/grouping
const projectsByLocation = groupBy(projects, 'location');

// Show on project cards
<ProjectCard>
  <div className="meta">
    {project.location || 'Ingen lokasjon'}
  </div>
</ProjectCard>
```

**Where it comes from:** Inherited from winning offer's `location` field (only if project's location was empty).

---

### `externalReference` (string, optional)

**Why:** Customers often have their own reference/order numbers (e.g., "22001", "PO-2025-1234"). This needs to live on the project for invoicing, reporting, and customer communication.

**How to use:**
```tsx
// Display alongside internal number
<ProjectNumber>
  <span className="internal">{project.projectNumber}</span>
  {project.externalReference && (
    <span className="external">
      KundeRef: {project.externalReference}
    </span>
  )}
</ProjectNumber>

// Use for customer-facing documents
const generateInvoice = (project) => ({
  ourReference: project.projectNumber,
  yourReference: project.externalReference,
  // ...
});

// Enable search by external reference
<SearchInput
  placeholder="Søk på prosjektnummer eller kundereferanse..."
  onSearch={(q) => searchProjects({
    projectNumber: q,
    externalReference: q
  })}
/>
```

**Where it comes from:** Inherited from winning offer's `externalReference` field.

---

## 3. Offer-Project Linking Rules

### The Rule

**Offers can ONLY be linked to projects in "tilbud" phase.**

### Why This Rule Exists

Once a project has progressed past the bidding stage, its fate is sealed:
- `active` = Work has started based on a won offer
- `completed` = Project is done
- `cancelled` = Project was abandoned

Adding new offers to these projects doesn't make business sense and would create data inconsistencies.

### API Behavior

| Scenario | API Response |
|----------|--------------|
| Link offer to `tilbud` project | `200 OK` |
| Link offer to `active` project | `400 Bad Request` |
| Link offer to `completed` project | `400 Bad Request` |
| Link offer to `cancelled` project | `400 Bad Request` |

**Error response:**
```json
{
  "error": "Project must be in tilbud (offer) phase to link offers"
}
```

### Frontend Implementation

```tsx
// Option 1: Filter the dropdown
const ProjectSelector = ({ onSelect }) => {
  const { projects } = useProjects();

  // Only show projects that can accept offers
  const linkableProjects = projects.filter(p => p.phase === 'tilbud');

  return (
    <Select
      options={linkableProjects}
      getOptionLabel={(p) => `${p.projectNumber || ''} ${p.name}`}
      onChange={onSelect}
    />
  );
};

// Option 2: Show all but disable non-linkable (better UX - user sees why)
const ProjectSelector = ({ onSelect }) => {
  const { projects } = useProjects();

  return (
    <Select
      options={projects}
      getOptionLabel={(p) => `${p.projectNumber || ''} ${p.name}`}
      isOptionDisabled={(p) => p.phase !== 'tilbud'}
      formatOptionLabel={(p) => (
        <div className={p.phase !== 'tilbud' ? 'disabled' : ''}>
          <span>{p.name}</span>
          {p.phase !== 'tilbud' && (
            <span className="hint">
              Prosjekt er i {translatePhase(p.phase)} fase
            </span>
          )}
        </div>
      )}
      onChange={onSelect}
    />
  );
};

// Option 3: Handle error gracefully
const linkOfferToProject = async (offerId, projectId) => {
  try {
    await api.post(`/offers/${offerId}/link-project`, { projectId });
    toast.success('Tilbud koblet til prosjekt');
  } catch (error) {
    if (error.response?.status === 400) {
      toast.error('Kan kun koble tilbud til prosjekter i tilbudsfase');
    } else {
      toast.error('Noe gikk galt');
    }
  }
};
```

---

## 4. Auto-Expire Competing Offers

### The Rule

**When an offer wins, ALL other offers on the same project are automatically set to "expired".**

### Why This Exists

A project can only have ONE winning offer. Previously, when an offer won:
- Other offers stayed in `sent` or `in_progress` state
- Users had to manually close them
- Reports showed inflated pipeline values
- Confusion about which offer actually won

Now the system automatically cleans up, ensuring data integrity.

### API Response

When you call `POST /offers/:id/win`:

```json
{
  "offer": {
    "id": "abc-123",
    "title": "Hovedtilbud lagerbygg",
    "offerNumber": "TK-2025-001W",
    "phase": "won"
  },
  "project": {
    "id": "def-456",
    "name": "Nytt lagerbygg",
    "projectNumber": "TK-2025-001",
    "phase": "active"
  },
  "expiredOffers": [
    {
      "id": "ghi-789",
      "title": "Alternativt tilbud",
      "offerNumber": "TK-2025-002",
      "phase": "expired"
    },
    {
      "id": "jkl-012",
      "title": "Budsjettilbud",
      "offerNumber": "TK-2025-003",
      "phase": "expired"
    }
  ],
  "expiredCount": 2
}
```

### Frontend Implementation

```tsx
const WinOfferButton = ({ offer }) => {
  const { offers } = useProjectOffers(offer.projectId);
  const siblingOffers = offers.filter(o =>
    o.id !== offer.id &&
    !['won', 'lost', 'expired'].includes(o.phase)
  );

  const handleWin = async () => {
    // Show warning if there are sibling offers
    if (siblingOffers.length > 0) {
      const confirmed = await confirmDialog({
        title: 'Bekreft vinn tilbud',
        message: `Dette vil sette ${siblingOffers.length} andre tilbud til utløpt:`,
        details: siblingOffers.map(o => `• ${o.title} (${o.offerNumber})`),
        confirmText: 'Ja, vinn tilbud',
        cancelText: 'Avbryt'
      });

      if (!confirmed) return;
    }

    try {
      const result = await api.post(`/offers/${offer.id}/win`);

      // Show success with context
      if (result.expiredCount > 0) {
        toast.success(
          `Tilbud vunnet! ${result.expiredCount} konkurrerende tilbud ble satt til utløpt.`
        );
      } else {
        toast.success('Tilbud vunnet!');
      }

      // Refresh data
      invalidateQueries(['offers', 'projects']);

    } catch (error) {
      toast.error('Kunne ikke vinne tilbud');
    }
  };

  return (
    <Button onClick={handleWin} variant="success">
      <TrophyIcon /> Vinn tilbud
      {siblingOffers.length > 0 && (
        <Badge>{siblingOffers.length} vil utløpe</Badge>
      )}
    </Button>
  );
};
```

### Activity Feed

The system logs activities for transparency:
- On winning offer: "Offer won"
- On each expired offer: "Offer was auto-expired because offer 'X' (TK-2025-001) won on project 'Y'"
- On project: "Project activated with winning offer 'X'. N sibling offer(s) were expired."

```tsx
// Activity feed will show these automatically
<ActivityFeed entityType="offer" entityId={offerId} />
```

---

## 5. Numbering Convention

### The System

When an offer wins, the numbering works like this:

| Entity | Before Win | After Win |
|--------|------------|-----------|
| Offer | `TK-2025-001` | `TK-2025-001W` |
| Project | *(no number)* | `TK-2025-001` |

**The "W" suffix** = Won (Vunnet)

### Why This Design?

1. **Traceability:** You can always trace a project back to its winning offer
2. **Uniqueness:** The offer keeps a unique number (with W) while project gets the "clean" number
3. **Customer-facing:** Project number (without W) is what customers see on invoices
4. **Internal tracking:** The W suffix instantly tells staff this offer was successful

### Frontend Display

```tsx
// Offer number display
const OfferNumber = ({ offer }) => {
  const isWon = offer.phase === 'won';

  return (
    <span className={`offer-number ${isWon ? 'won' : ''}`}>
      {offer.offerNumber}
      {isWon && !offer.offerNumber.endsWith('W') && 'W'}
      {isWon && <CheckCircleIcon className="won-icon" />}
    </span>
  );
};

// Project number with origin
const ProjectNumber = ({ project }) => (
  <div className="project-number">
    <span className="number">{project.projectNumber}</span>
    {project.inheritedOfferNumber && (
      <span className="origin">
        fra tilbud {project.inheritedOfferNumber}W
      </span>
    )}
  </div>
);

// In a table/list showing relationship
<Table>
  <TableRow>
    <TableCell>Prosjekt</TableCell>
    <TableCell>{project.projectNumber}</TableCell>
  </TableRow>
  <TableRow>
    <TableCell>Vunnet tilbud</TableCell>
    <TableCell>{project.inheritedOfferNumber}W</TableCell>
  </TableRow>
  <TableRow>
    <TableCell>Kundereferanse</TableCell>
    <TableCell>{project.externalReference || '-'}</TableCell>
  </TableRow>
</Table>
```

---

## 6. Smart Field Inheritance

### The Philosophy

When an offer wins, the project should automatically get relevant data from the offer. **BUT** we don't want to overwrite data that was manually entered on the project.

### Inheritance Rules

| Offer Field | Project Field | Rule |
|-------------|---------------|------|
| `value` | `value` | **Always** overwrites |
| `cost` | `cost` | **Always** overwrites |
| `customerId` | `customerId` | **Always** overwrites |
| `customerName` | `customerName` | **Always** overwrites |
| `offerNumber` | `projectNumber` | **Always** sets |
| `offerNumber` | `inheritedOfferNumber` | **Always** sets |
| `responsibleUserId` | `managerId` | **Only if** project has no manager |
| `responsibleUserName` | `managerName` | **Only if** project has no manager |
| `description` | `description` | **Only if** project description is empty |
| `location` | `location` | **Only if** project location is empty |
| `externalReference` | `externalReference` | **Only if** project has none |

### Why "Only If Empty"?

Consider this scenario:
1. Project created with description: "Lagerbygg 5000m² med kontor"
2. Offer created with description: "Tilbud på takarbeid"
3. Offer wins

Without conditional inheritance, the project's detailed description would be replaced by the offer's shorter one. With conditional inheritance, the project keeps its original description.

### Frontend Implications

**You don't need to do anything special.** The API handles this automatically. But you can inform users:

```tsx
const WinOfferConfirmation = ({ offer, project }) => {
  const willInherit = [];
  const willKeep = [];

  // Check what will be inherited vs kept
  if (!project.managerId && offer.responsibleUserId) {
    willInherit.push(`Prosjektleder: ${offer.responsibleUserName}`);
  } else if (project.managerId) {
    willKeep.push(`Prosjektleder: ${project.managerName} (beholdes)`);
  }

  if (!project.description && offer.description) {
    willInherit.push('Beskrivelse fra tilbud');
  } else if (project.description) {
    willKeep.push('Beskrivelse (beholdes)');
  }

  // ... similar for location, externalReference

  return (
    <ConfirmDialog>
      <h3>Ved å vinne dette tilbudet:</h3>

      <h4>Vil arves fra tilbud:</h4>
      <ul>
        <li>Verdi: {formatCurrency(offer.value)}</li>
        <li>Kostnad: {formatCurrency(offer.cost)}</li>
        <li>Kunde: {offer.customerName}</li>
        {willInherit.map(item => <li key={item}>{item}</li>)}
      </ul>

      {willKeep.length > 0 && (
        <>
          <h4>Beholdes på prosjekt:</h4>
          <ul>
            {willKeep.map(item => <li key={item}>{item}</li>)}
          </ul>
        </>
      )}
    </ConfirmDialog>
  );
};
```

---

## 7. What NOT to Change

### Existing Endpoints Still Work

All existing API endpoints work exactly as before. This release is **backwards compatible**.

| Endpoint | Change |
|----------|--------|
| `GET /projects` | ✅ Same (now includes `location`, `externalReference`) |
| `GET /projects/:id` | ✅ Same (now includes `location`, `externalReference`) |
| `POST /projects` | ✅ Same |
| `PUT /projects/:id` | ✅ Same |
| `GET /offers` | ✅ Same |
| `GET /offers/:id` | ✅ Same |
| `POST /offers` | ⚠️ Will fail if `projectId` refers to non-tilbud project |
| `POST /offers/:id/win` | ✅ Same (now auto-expires siblings, returns more data) |

### Don't Remove Existing Fields

These fields still exist and work:
- `project.inheritedOfferNumber` - Still set, same as `projectNumber` when won
- `project.winningOfferId` - Still set, references the winning offer
- `project.offerId` - Still exists for direct offer reference

### Don't Change Phase Handling

The project phases are unchanged:
- `tilbud` (was: offer phase)
- `active` (was: active phase)
- `completed`
- `cancelled`

### Don't Modify Offer Phases

Offer phases are unchanged:
- `draft`
- `in_progress`
- `sent`
- `won`
- `lost`
- `expired`

---

## 8. API Changes Reference

### New Response Fields

**ProjectDTO additions:**
```typescript
interface ProjectDTO {
  // Existing fields...

  // NEW
  location?: string;
  externalReference?: string;
}
```

**WinOfferResponse (enhanced):**
```typescript
interface WinOfferResponse {
  offer: OfferDTO;
  project: ProjectDTO;
  expiredOffers: OfferDTO[];  // NEW - list of auto-expired offers
  expiredCount: number;        // NEW - count for convenience
}
```

### New Error Responses

**Linking to non-tilbud project:**
```
POST /offers (with projectId to non-tilbud project)
POST /offers/:id/link-project (to non-tilbud project)

Response: 400 Bad Request
{
  "error": "Project must be in tilbud (offer) phase to link offers"
}
```

---

## 9. Frontend Implementation Guide

### Checklist

- [ ] Update TypeScript types to include `location` and `externalReference` on ProjectDTO
- [ ] Add location display to project cards/details
- [ ] Add external reference display where appropriate
- [ ] Filter/disable project selector when linking offers (only tilbud phase)
- [ ] Handle 400 error when linking to non-tilbud project
- [ ] Update win offer flow to show/handle expired siblings
- [ ] Display "W" suffix appropriately for won offers
- [ ] Update any project number displays to use new `projectNumber` field

### Migration Notes

**No breaking changes.** You can deploy frontend changes gradually:

1. **Phase 1:** Update types, add new field displays (safe, just shows more data)
2. **Phase 2:** Add project filtering for offer linking (improves UX)
3. **Phase 3:** Enhance win offer flow with expiry warnings (improves UX)

### Testing Scenarios

1. **Create offer with project link**
   - Link to tilbud project → Should succeed
   - Link to active project → Should fail with clear error

2. **Win offer with siblings**
   - Win offer that has sibling offers → Siblings should expire
   - Check response includes `expiredOffers` array
   - Check activities logged correctly

3. **Field inheritance**
   - Win offer on project with no manager → Project gets offer's responsible
   - Win offer on project with existing manager → Project keeps its manager
   - Same for description, location, externalReference

4. **Numbering**
   - Win offer → Offer gets W suffix
   - Win offer → Project gets original number as `projectNumber`
   - Numbers display correctly in UI

---

## Questions?

Contact the backend team or check the API documentation at `/swagger/index.html`.
