# Frontend Integration: Offer-Project Business Rules (sc-289)

## Overview

This update introduces business rules for offer-project relationships, automatic state management when offers are won, and improved field inheritance.

---

## 1. New Project Fields

Two new fields added to Project responses:

```typescript
interface ProjectDTO {
  // ... existing fields
  location?: string;           // NEW - inherited from offer
  externalReference?: string;  // NEW - customer's reference number (e.g., "22001")
}
```

### Usage
- `location`: Display where the project is located (e.g., "Fredrikstad", "Oslo")
- `externalReference`: Customer's own reference/order number for cross-referencing

---

## 2. Offer Linking Restrictions

### Rule: Offers can only be linked to projects in "tilbud" phase

**API Behavior:**
- `POST /offers` with `projectId` → Returns `400 Bad Request` if project is not in "tilbud" phase
- `POST /offers/:id/link-project` → Returns `400 Bad Request` if project is not in "tilbud" phase

**Error Response:**
```json
{
  "error": "Project must be in tilbud (offer) phase to link offers"
}
```

**Frontend Recommendation:**
- When selecting a project for an offer, filter to show only projects with `phase: "tilbud"`
- Or show all projects but disable/gray out non-tilbud projects with tooltip explaining why

---

## 3. Auto-Expire Competing Offers

### Rule: When an offer wins, all other offers on the same project are automatically expired

**What happens when `POST /offers/:id/win` is called:**
1. The winning offer's phase → `won`
2. All sibling offers (same project, different offer) → `expired`
3. Project phase → `active`

**API Response includes expired offers:**
```json
{
  "offer": { /* winning offer */ },
  "project": { /* updated project */ },
  "expiredOffers": [
    { "id": "...", "title": "Competing Offer 1", "phase": "expired" },
    { "id": "...", "title": "Competing Offer 2", "phase": "expired" }
  ],
  "expiredCount": 2
}
```

**Frontend Recommendation:**
- Show a notification/toast: "Tilbud vunnet! 2 andre tilbud ble automatisk utløpt."
- Refresh the project's offer list to show updated phases

---

## 4. Numbering Convention

### When an offer wins:

| Before | After |
|--------|-------|
| Offer: `TK-2023-001` | Offer: `TK-2023-001W` |
| Project: no project_number | Project: `TK-2023-001` |

**Key Points:**
- The **"W" suffix** indicates a won offer
- The **project claims the original number** as its `projectNumber`
- Both `projectNumber` and `inheritedOfferNumber` will have the same value

**Display Example:**
```
Prosjekt: TK-2023-001 - Nytt lagerbygg
Vunnet tilbud: TK-2023-001W
```

---

## 5. Field Inheritance (Offer → Project)

When an offer is won, the project inherits these fields from the offer:

| Offer Field | Project Field | Condition |
|-------------|---------------|-----------|
| `value` | `value` | Always |
| `cost` | `cost` | Always |
| `customerId` | `customerId` | Always |
| `customerName` | `customerName` | Always |
| `offerNumber` | `projectNumber` | Always |
| `offerNumber` | `inheritedOfferNumber` | Always |
| `responsibleUserId` | `managerId` | Only if project has no manager |
| `responsibleUserName` | `managerName` | Only if project has no manager |
| `description` | `description` | Only if project description is empty |
| `location` | `location` | Only if project location is empty |
| `externalReference` | `externalReference` | Only if project has none |

**Note:** Fields marked "Only if empty" respect manually-entered project data.

---

## 6. Project Phase Flow

```
tilbud → active → completed
           ↓
       cancelled
```

| Phase | Description | Can link offers? |
|-------|-------------|------------------|
| `tilbud` | Bidding phase, awaiting offer outcome | ✅ Yes |
| `active` | Work in progress (offer was won) | ❌ No |
| `completed` | Project finished | ❌ No |
| `cancelled` | Project cancelled/lost | ❌ No |

---

## 7. TypeScript Types Update

```typescript
// Updated ProjectDTO
interface ProjectDTO {
  id: string;
  name: string;
  projectNumber?: string;        // e.g., "TK-2023-001"
  inheritedOfferNumber?: string; // Same as projectNumber when won
  externalReference?: string;    // NEW - Customer's reference
  location?: string;             // NEW - Project location
  phase: 'tilbud' | 'active' | 'completed' | 'cancelled';
  customerId: string;
  customerName: string;
  managerId?: string;
  managerName?: string;
  value: number;
  cost: number;
  // ... other fields
}

// WinOffer response
interface WinOfferResponse {
  offer: OfferDTO;
  project: ProjectDTO;
  expiredOffers: OfferDTO[];
  expiredCount: number;
}

// Error response for linking restrictions
interface ErrorResponse {
  error: string;
}
```

---

## 8. UI Recommendations

### Offer Creation/Linking
```tsx
// Filter projects to only show those in tilbud phase
const linkableProjects = projects.filter(p => p.phase === 'tilbud');

// Or show all with disabled state
<ProjectSelect
  projects={projects}
  isOptionDisabled={(p) => p.phase !== 'tilbud'}
  getOptionTooltip={(p) =>
    p.phase !== 'tilbud'
      ? 'Kan kun koble tilbud til prosjekter i tilbudsfase'
      : undefined
  }
/>
```

### Win Offer Confirmation
```tsx
// Show warning if project has other offers
const siblingOffers = offers.filter(o =>
  o.projectId === selectedOffer.projectId &&
  o.id !== selectedOffer.id &&
  !['won', 'lost', 'expired'].includes(o.phase)
);

if (siblingOffers.length > 0) {
  showConfirmDialog({
    title: 'Bekreft vinn tilbud',
    message: `${siblingOffers.length} andre tilbud på dette prosjektet vil bli satt til utløpt.`,
    onConfirm: () => winOffer(selectedOffer.id)
  });
}
```

### Display Won Offer Number
```tsx
// Show the W suffix for won offers
const displayOfferNumber = (offer: OfferDTO) => {
  if (offer.phase === 'won' && !offer.offerNumber.endsWith('W')) {
    return `${offer.offerNumber}W`;
  }
  return offer.offerNumber;
};
```

---

## 9. Activity Log

New activities are logged:
- "Offer auto-expired" - When sibling offers are expired due to another winning
- "Project activated" - When project transitions to active phase

These appear in the activity feed for both offers and projects.
