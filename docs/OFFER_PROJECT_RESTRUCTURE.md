# Offer-Project Restructure Guide

This document describes the major restructuring of the Offer and Project relationship in the Relation API. This is a **breaking change** that affects how offers and projects work together.

## Summary of Changes

### Before (Old Model)
- **Offers** were sales proposals that, when won, created a **Project**
- **Projects** contained all execution tracking (manager, team, spent, invoiced, health, completion%)
- Projects were company-specific (had `companyId`)
- Winning an offer automatically created a project and copied budget items

### After (New Model)
- **Offers** are now the full lifecycle entity - from draft to completed order
- **Projects** are simplified to cross-company folders/containers for grouping offers
- All execution tracking (manager, team, spent, invoiced, health) is now on the **Offer**
- Project creation is optional and controlled by the client

---

## New Offer Lifecycle

### Offer Phases

```
┌─────────┐    ┌─────────────┐    ┌──────┐    ┌───────┐    ┌───────────┐
│  draft  │───▶│ in_progress │───▶│ sent │───▶│ order │───▶│ completed │
└─────────┘    └─────────────┘    └──────┘    └───────┘    └───────────┘
                                      │
                                      ├───▶ lost
                                      │
                                      └───▶ expired
```

| Phase | Description | Can Update? | Next Phases |
|-------|-------------|-------------|-------------|
| `draft` | Initial creation, no offer number yet | Yes | `in_progress`, `sent`, `expired` |
| `in_progress` | Being worked on internally | Yes | `draft`, `sent`, `expired` |
| `sent` | Sent to customer, awaiting response | Yes | `in_progress`, `order`, `lost`, `expired` |
| `order` | **NEW** - Customer accepted, work in progress | No* | `completed` |
| `completed` | **NEW** - Order finished | No | - |
| `lost` | Customer rejected | No | - |
| `expired` | Offer expired without response | No | - |

*Orders can only be updated via dedicated endpoints (health, spent, invoiced)

### Key Change: `won` → `order`
The old `won` phase has been replaced by `order`. An offer in `order` phase represents active work being executed for the customer - what was previously tracked on a Project.

---

## New Offer Fields (Order Phase)

When an offer transitions to `order` phase, these fields become relevant:

```typescript
interface OfferDTO {
  // ... existing fields ...

  // NEW: Order execution fields
  managerId?: string;           // UUID of assigned manager
  managerName: string;          // Denormalized manager name
  teamMembers: string[];        // Array of user UUIDs
  spent: number;                // Amount spent so far
  invoiced: number;             // Amount invoiced to customer
  orderReserve: number;         // Calculated: value - invoiced
  health?: 'on_track' | 'at_risk' | 'delayed' | 'over_budget';
  completionPercent?: number;   // 0-100
  startDate?: string;           // ISO date
  endDate?: string;             // ISO date
  estimatedCompletionDate?: string; // ISO date
}
```

---

## New API Endpoints

### Accept Order (Transition sent → order)
```
POST /api/v1/offers/{id}/accept-order
```

**Request Body:**
```json
{
  "notes": "Customer confirmed via email"  // optional
}
```

**Response:**
```json
{
  "offer": { /* OfferDTO with phase: "order" */ }
}
```

This is the primary way to transition an offer to order phase. The offer number will have "O" appended (e.g., `SB-2024-001` → `SB-2024-001O`).

### Update Order Health
```
PUT /api/v1/offers/{id}/health
```

**Request Body:**
```json
{
  "health": "at_risk",        // required: on_track | at_risk | delayed | over_budget
  "completionPercent": 65.5   // optional: 0-100
}
```

Only works for offers in `order` phase.

### Update Order Spent
```
PUT /api/v1/offers/{id}/spent
```

**Request Body:**
```json
{
  "spent": 25000.50
}
```

Only works for offers in `order` phase.

### Update Order Invoiced
```
PUT /api/v1/offers/{id}/invoiced
```

**Request Body:**
```json
{
  "invoiced": 50000.00
}
```

Only works for offers in `order` phase.

### Complete Order (Transition order → completed)
```
POST /api/v1/offers/{id}/complete
```

No request body. Marks the order as finished.

---

## Changed Endpoints

### Accept Offer (with optional project creation)
```
POST /api/v1/offers/{id}/accept
```

**Request Body:**
```json
{
  "createProject": true,                    // optional, default: false
  "projectName": "Customer HQ Renovation"   // optional, defaults to offer title
}
```

**Response:**
```json
{
  "offer": { /* OfferDTO with phase: "order" */ },
  "project": { /* ProjectDTO or null */ }
}
```

**Key Changes:**
- Project creation is now **optional** (controlled by `createProject` flag)
- If `createProject: false` (default), no project is created
- If `createProject: true`, a project folder is created and linked to the offer
- Budget items are cloned to the project when created

### Create Offer
```
POST /api/v1/offers
```

**New optional fields in request:**
```json
{
  // ... existing fields ...
  "createProject": false,  // optional: auto-create project on creation
  "projectId": "uuid"      // optional: link to existing project
}
```

### Advance Offer
```
PUT /api/v1/offers/{id}/advance
```

**Important:** You can no longer advance directly to `order`, `lost`, or `expired` phases. Use the dedicated endpoints:
- `order` → `POST /offers/{id}/accept` or `POST /offers/{id}/accept-order`
- `lost` → `POST /offers/{id}/reject`
- `expired` → `POST /offers/{id}/expire`

---

## Removed/Deprecated Project Endpoints

The following project endpoints have been **removed** (functionality moved to Offer):

| Removed Endpoint | Replacement |
|------------------|-------------|
| `GET /projects/{id}/budget` | Use offer budget endpoints |
| `PUT /projects/{id}/budget` | Use offer budget endpoints |
| `PUT /projects/{id}/spent` | `PUT /offers/{id}/spent` |
| `PUT /projects/{id}/health` | `PUT /offers/{id}/health` |
| `POST /projects/{id}/inherit-budget` | N/A - budget stays on offer |
| `POST /projects/{id}/resync-from-offer` | N/A - no longer needed |

---

## Simplified Project Model

Projects are now simple folders/containers. The following fields have been **removed** from ProjectDTO:

```typescript
// REMOVED from ProjectDTO:
interface RemovedProjectFields {
  companyId: string;          // Projects are now cross-company
  value: number;
  cost: number;
  marginPercent: number;
  spent: number;
  invoiced: number;
  orderReserve: number;
  managerId: string;
  managerName: string;
  teamMembers: string[];
  health: string;
  completionPercent: number;
  estimatedCompletionDate: string;
  hasDetailedBudget: boolean;
  winningOfferId: string;
  inheritedOfferNumber: string;
  calculatedOfferValue: number;
  wonAt: string;
  isEconomicsEditable: boolean;
}
```

**Remaining ProjectDTO fields:**
```typescript
interface ProjectDTO {
  id: string;
  name: string;
  projectNumber: string;
  summary: string;
  description: string;
  customerId?: string;      // Now optional
  customerName: string;
  phase: ProjectPhase;
  startDate: string;
  endDate: string;
  location: string;
  dealId?: string;
  externalReference: string;
  createdAt: string;
  updatedAt: string;
  createdById: string;
  createdByName: string;
  updatedById: string;
  updatedByName: string;
}
```

### New Project Phases

```typescript
type ProjectPhase =
  | 'tilbud'     // Initial/proposal stage
  | 'working'    // Active work (replaces 'active')
  | 'on_hold'    // NEW - Paused
  | 'cancelled'  // Cancelled (terminal)
  | 'completed'; // Finished
```

**Phase Transitions:**
- `tilbud` → `working`, `on_hold`, `cancelled`
- `working` → `on_hold`, `completed`, `cancelled`, `tilbud`
- `on_hold` → `working`, `cancelled`, `completed`
- `completed` → `working` (can reopen)
- `cancelled` → (terminal, no transitions)

---

## Frontend Migration Guide

### 1. Update Offer List/Grid Views

Add new phase badges:
- `order` - Active order being executed (was "won")
- `completed` - Finished order

Consider showing order-specific columns:
- Health status indicator
- Completion percentage
- Spent vs Budget

### 2. Update Offer Detail View

For offers in `order` phase, show execution tracking section:
- Manager assignment
- Team members
- Spent / Invoiced / Order Reserve
- Health status (on_track, at_risk, delayed, over_budget)
- Completion percentage
- Start/End dates

### 3. Update "Accept Offer" Flow

Old flow:
```
Click "Accept" → Offer becomes "won" → Project auto-created
```

New flow:
```
Click "Accept" → Show dialog:
  □ Create project folder for this order
  [Project name: ________]

→ Offer becomes "order"
→ Project created only if checkbox selected
```

### 4. Update Project Views

Remove economic tracking from project views:
- No more budget/spent/invoiced on projects
- No more health/completion on projects
- No more manager/team on projects

Projects are now just organizational folders. Show:
- Basic info (name, customer, dates)
- List of linked offers
- Files attached to project

### 5. Dashboard Changes

Update dashboard metrics to use offer data:

**Old (Project-based):**
- Active projects count
- Project health distribution
- Project budget utilization

**New (Offer/Order-based):**
- Active orders count (offers in `order` phase)
- Order health distribution
- Order budget utilization (spent vs value)
- Completed orders (offers in `completed` phase)

### 6. API Call Updates

| Old API Call | New API Call |
|--------------|--------------|
| `POST /offers/{id}/win` | `POST /offers/{id}/accept` or `POST /offers/{id}/accept-order` |
| `PUT /projects/{id}/spent` | `PUT /offers/{id}/spent` |
| `PUT /projects/{id}/health` | `PUT /offers/{id}/health` |
| `GET /projects/{id}/budget` | `GET /offers/{id}/budget` |

---

## Example Workflows

### Workflow 1: Simple Order (No Project)

```
1. Create offer (phase: draft)
2. Work on offer (phase: in_progress)
3. Send to customer (phase: sent)
4. Customer accepts:
   POST /offers/{id}/accept-order
   → phase: order
5. Track execution:
   PUT /offers/{id}/spent
   PUT /offers/{id}/health
6. Complete:
   POST /offers/{id}/complete
   → phase: completed
```

### Workflow 2: Order with Project Folder

```
1. Create offer (phase: draft)
2. Send to customer (phase: sent)
3. Customer accepts, create project:
   POST /offers/{id}/accept
   { "createProject": true, "projectName": "Customer HQ" }
   → phase: order
   → project created and linked
4. Add more offers to same project folder
5. Track execution on each offer individually
```

### Workflow 3: Multiple Offers, One Project

```
1. Create project folder:
   POST /projects
   { "name": "Customer Multi-Phase Project" }

2. Create offers linked to project:
   POST /offers
   { "title": "Phase 1", "projectId": "{project-id}" }
   POST /offers
   { "title": "Phase 2", "projectId": "{project-id}" }

3. Each offer follows its own lifecycle
4. Project serves as organizational container
```

---

## Questions?

Contact the backend team for clarification on any of these changes.
