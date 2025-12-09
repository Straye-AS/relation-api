# Straye Relation API - Frontend Development Guide

This guide provides comprehensive documentation for frontend developers integrating with the Straye Relation API. Use this alongside the Swagger documentation (`/swagger/index.html`) for full API details.

## Table of Contents
1. [Authentication & Authorization](#authentication--authorization)
2. [Multi-Tenant Architecture](#multi-tenant-architecture)
3. [Core Entities & Relationships](#core-entities--relationships)
4. [API Endpoints Reference](#api-endpoints-reference)
5. [Common Patterns](#common-patterns)
6. [Workflows & State Machines](#workflows--state-machines)
7. [Frontend Implementation Guide](#frontend-implementation-guide)
8. [Pitfalls & Edge Cases](#pitfalls--edge-cases)
9. [Recommended Features](#recommended-features)

---

## Authentication & Authorization

### Authentication Methods

The API supports two authentication methods:

#### 1. Azure AD JWT Token (Primary - for user sessions)
```http
Authorization: Bearer <jwt_token>
```
- Obtain token via Azure AD OAuth 2.0 flow
- Contains user identity, roles, and company membership
- Required for all user-facing operations

#### 2. API Key (System/Backend integrations)
```http
x-api-key: <api_key>
```
- Used for service-to-service communication
- Has elevated permissions but limited audit context

### Getting Current User Context

After authentication, fetch user context:

```http
GET /api/v1/auth/me
```

**Response:**
```json
{
  "id": "user-uuid",
  "name": "Ola Nordmann",
  "email": "ola@straye.no",
  "roles": ["sales_manager", "user"],
  "company": { "id": "stalbygg", "name": "Straye Stalbygg" },
  "initials": "ON",
  "isSuperAdmin": false,
  "isCompanyAdmin": true
}
```

### Getting User Permissions

```http
GET /api/v1/auth/permissions
```

**Response:**
```json
{
  "permissions": [
    { "resource": "offers", "action": "create", "allowed": true },
    { "resource": "deals", "action": "delete", "allowed": false }
  ],
  "roles": ["sales_manager"],
  "isSuperAdmin": false
}
```

**Frontend Action:** Use this to conditionally render UI elements (edit buttons, delete actions, etc.)

### User Roles & Capabilities

| Role | Capabilities |
|------|-------------|
| `super_admin` | Full system access across all companies |
| `company_admin` | Full access within assigned company |
| `sales_manager` | Manage deals, offers, customers |
| `project_manager` | Manage projects, view related offers |
| `user` | Basic read access, limited write |

---

## Multi-Tenant Architecture

### Company IDs (Tenant Identifiers)

The system supports multiple Straye companies:

| ID | Company |
|----|---------|
| `gruppen` | Straye Gruppen (parent) |
| `stalbygg` | Straye Stalbygg |
| `hybridbygg` | Straye Hybridbygg |
| `industri` | Straye Industri |
| `tak` | Straye Tak |
| `montasje` | Straye Montasje |

### Automatic Filtering

**Critical:** All data is automatically filtered by the user's company context. The middleware extracts company from JWT token and applies it to all queries.

- Users only see data from their assigned company
- Super admins can access all companies
- Creating entities requires explicit `companyId` field matching user's company

**Frontend Action:** When creating entities, always include `companyId` from the current user's context:
```javascript
const createDeal = async (dealData) => {
  const user = await getCurrentUser();
  return api.post('/deals', {
    ...dealData,
    companyId: user.company.id
  });
};
```

---

## Core Entities & Relationships

### Entity Relationship Diagram

```
                                  ┌──────────────┐
                                  │   Customer   │
                                  │              │
                                  └──────┬───────┘
                                         │
              ┌──────────────────────────┼──────────────────────────┐
              │                          │                          │
              ▼                          ▼                          ▼
       ┌──────────┐              ┌──────────────┐           ┌──────────────┐
       │   Deal   │──────────────│    Offer     │───────────│   Project    │
       │          │  creates     │              │  converts │              │
       └────┬─────┘              └──────┬───────┘   to      └──────┬───────┘
            │                           │                          │
            │                           │                          │
            ▼                           ▼                          ▼
   ┌────────────────┐          ┌────────────────┐         ┌────────────────┐
   │ Deal Stage     │          │ Budget         │         │ Budget         │
   │ History        │          │ Dimensions     │────────▶│ Dimensions     │
   └────────────────┘          └────────────────┘ inherit └────────────────┘
                                       │
                                       ▼
                               ┌────────────────┐
                               │     Files      │
                               └────────────────┘

                    ┌────────────────────────────────┐
                    │           Contact              │
                    │  (linked to multiple entities) │
                    └────────────────────────────────┘
                               │
            ┌──────────────────┼──────────────────┐
            ▼                  ▼                  ▼
       Customer             Deal             Project

                    ┌────────────────────────────────┐
                    │           Activity             │
                    │  (linked to multiple entities) │
                    └────────────────────────────────┘
                               │
            ┌──────────────────┼──────────────────┐
            ▼                  ▼                  ▼
         Deal              Offer             Project
```

### Key Entity Types

#### Customer
The root entity for all business relationships.

```typescript
interface Customer {
  id: string;           // UUID
  name: string;
  orgNumber: string;    // Norwegian org number
  email: string;
  phone: string;
  status: 'active' | 'inactive' | 'prospect' | 'churned';
  tier: 'standard' | 'premium' | 'enterprise';
  industry: string;     // 'construction' | 'manufacturing' | etc.
}
```

#### Deal (Sales Pipeline)
Represents a sales opportunity through the pipeline.

```typescript
interface Deal {
  id: string;
  title: string;
  customerId: string;
  customerName: string;     // Denormalized for display
  stage: 'lead' | 'qualified' | 'proposal' | 'negotiation' | 'won' | 'lost';
  probability: number;      // 0-100, auto-updates with stage
  value: number;
  weightedValue: number;    // value * probability/100
  expectedCloseDate?: string;
  ownerId: string;
  ownerName: string;        // Denormalized
  offerId?: string;         // Linked offer (if created)
  lostReason?: string;
  lossReasonCategory?: 'price' | 'timing' | 'competitor' | 'requirements' | 'other';
}
```

#### Offer (Quote/Proposal)
Detailed pricing proposal sent to customer.

```typescript
interface Offer {
  id: string;
  title: string;
  customerId: string;
  customerName: string;
  phase: 'draft' | 'in_progress' | 'sent' | 'won' | 'lost' | 'expired';
  status: 'active' | 'cancelled';
  probability: number;
  value: number;            // Calculated from budget dimensions
  responsibleUserId: string;
  items: OfferItem[];       // Legacy items (deprecated, use budget dimensions)
}
```

#### Project
Active work after winning an offer/deal.

```typescript
interface Project {
  id: string;
  name: string;
  projectNumber?: string;
  customerId: string;
  customerName: string;
  status: 'planning' | 'active' | 'on_hold' | 'completed' | 'cancelled';
  startDate: string;
  endDate?: string;
  budget: number;
  spent: number;
  managerId: string;
  managerName: string;
  teamMembers: string[];    // Array of user IDs
  offerId?: string;         // Source offer
  dealId?: string;          // Source deal
  hasDetailedBudget: boolean;
  health?: 'on_track' | 'at_risk' | 'critical';
  completionPercent?: number;
}
```

#### Contact (Multi-Entity)
Person linked to multiple entities (customer, deal, project).

```typescript
interface Contact {
  id: string;
  firstName: string;
  lastName: string;
  fullName: string;         // Computed
  email?: string;
  phone?: string;
  mobile?: string;
  title?: string;
  department?: string;
  contactType: 'primary' | 'billing' | 'technical' | 'decision_maker' | 'influencer' | 'other';
  primaryCustomerId?: string;
  relationships: ContactRelationship[];
}

interface ContactRelationship {
  entityType: 'customer' | 'deal' | 'offer' | 'project';
  entityId: string;
  role?: string;
  isPrimary: boolean;
}
```

#### Activity (Task/Event)
Track meetings, calls, tasks across entities.

```typescript
interface Activity {
  id: string;
  targetType: 'customer' | 'deal' | 'offer' | 'project';
  targetId: string;
  title: string;
  body?: string;
  activityType: 'call' | 'meeting' | 'email' | 'task' | 'note' | 'follow_up';
  status: 'planned' | 'in_progress' | 'completed' | 'cancelled';
  scheduledAt?: string;
  dueDate?: string;
  completedAt?: string;
  durationMinutes?: number;
  priority: number;         // 0-5
  isPrivate: boolean;
  creatorId: string;
  assignedToId?: string;
  attendees: string[];      // User IDs for meetings
  parentActivityId?: string; // For follow-up chains
}
```

#### Budget Dimension
Line items for offers and projects with cost/revenue breakdown.

```typescript
interface BudgetDimension {
  id: string;
  parentType: 'offer' | 'project';
  parentId: string;
  categoryId?: string;      // Predefined category
  customName?: string;      // Or custom name
  name: string;             // Resolved name
  cost: number;
  revenue: number;
  targetMarginPercent?: number;
  marginOverride: boolean;
  marginPercent: number;    // Calculated
  description?: string;
  quantity?: number;
  unit?: string;
  displayOrder: number;
}
```

---

## API Endpoints Reference

### Customers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/customers` | List customers (paginated) |
| POST | `/api/v1/customers` | Create customer |
| GET | `/api/v1/customers/{id}` | Get customer by ID |
| PUT | `/api/v1/customers/{id}` | Update customer |
| DELETE | `/api/v1/customers/{id}` | Delete customer |
| GET | `/api/v1/customers/{id}/contacts` | Get customer's contacts |
| POST | `/api/v1/customers/{id}/contacts` | Create contact for customer |

**Query Parameters for List:**
- `page` (default: 1)
- `pageSize` (default: 20, max: 200)
- `search` - Search name, email, org number
- `status` - Filter by status
- `tier` - Filter by tier

### Contacts

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/contacts` | List all contacts |
| POST | `/api/v1/contacts` | Create contact |
| GET | `/api/v1/contacts/{id}` | Get contact |
| PUT | `/api/v1/contacts/{id}` | Update contact |
| DELETE | `/api/v1/contacts/{id}` | Delete contact |
| POST | `/api/v1/contacts/{id}/relationships` | Add entity relationship |
| DELETE | `/api/v1/contacts/{id}/relationships/{relId}` | Remove relationship |

### Deals

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/deals` | List deals (paginated) |
| POST | `/api/v1/deals` | Create deal |
| GET | `/api/v1/deals/{id}` | Get deal |
| PUT | `/api/v1/deals/{id}` | Update deal |
| DELETE | `/api/v1/deals/{id}` | Delete deal |
| POST | `/api/v1/deals/{id}/advance` | Advance to next stage |
| POST | `/api/v1/deals/{id}/win` | Mark deal as won |
| POST | `/api/v1/deals/{id}/lose` | Mark deal as lost (requires reason) |
| POST | `/api/v1/deals/{id}/reopen` | Reopen closed deal |
| POST | `/api/v1/deals/{id}/create-offer` | Create offer from deal |
| GET | `/api/v1/deals/{id}/history` | Get stage change history |
| GET | `/api/v1/deals/{id}/activities` | Get related activities |
| GET | `/api/v1/deals/{id}/contacts` | Get related contacts |

**Analytics Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/deals/analytics` | Full pipeline analytics |
| GET | `/api/v1/deals/pipeline` | Pipeline overview by stage |
| GET | `/api/v1/deals/stats` | Quick pipeline statistics |
| GET | `/api/v1/deals/forecast` | Revenue forecast (30/90 days) |

**Query Parameters for List:**
- `page`, `pageSize`
- `stage` - Filter by stage
- `ownerId` - Filter by owner
- `customerId` - Filter by customer

### Offers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/offers` | List offers |
| POST | `/api/v1/offers` | Create offer |
| GET | `/api/v1/offers/{id}` | Get offer |
| PUT | `/api/v1/offers/{id}` | Update offer |
| DELETE | `/api/v1/offers/{id}` | Delete offer |
| POST | `/api/v1/offers/{id}/advance` | Advance phase |
| POST | `/api/v1/offers/{id}/send` | Mark as sent |
| POST | `/api/v1/offers/{id}/accept` | Accept offer (optionally create project) |
| POST | `/api/v1/offers/{id}/reject` | Reject offer |
| POST | `/api/v1/offers/{id}/clone` | Clone offer |
| GET | `/api/v1/offers/{id}/items` | Get items (legacy) |
| POST | `/api/v1/offers/{id}/items` | Add item (legacy) |
| GET | `/api/v1/offers/{id}/files` | Get attached files |
| GET | `/api/v1/offers/{id}/activities` | Get activities |
| GET | `/api/v1/offers/{id}/detail` | Get with budget dimensions |
| GET | `/api/v1/offers/{id}/budget` | Get budget summary |
| POST | `/api/v1/offers/{id}/recalculate` | Recalculate totals |

**Budget Dimension Sub-routes:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/offers/{id}/budget/dimensions` | List dimensions |
| POST | `/api/v1/offers/{id}/budget/dimensions` | Add dimension |
| PUT | `/api/v1/offers/{id}/budget/dimensions/{dimId}` | Update dimension |
| DELETE | `/api/v1/offers/{id}/budget/dimensions/{dimId}` | Delete dimension |
| PUT | `/api/v1/offers/{id}/budget/reorder` | Reorder dimensions |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/projects` | List projects |
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects/{id}` | Get project |
| PUT | `/api/v1/projects/{id}` | Update project |
| DELETE | `/api/v1/projects/{id}` | Delete project |
| PUT | `/api/v1/projects/{id}/status` | Update status and health |
| GET | `/api/v1/projects/{id}/budget` | Get budget details |
| POST | `/api/v1/projects/{id}/inherit-budget` | Copy budget from offer |
| GET | `/api/v1/projects/{id}/activities` | Get activities |
| GET | `/api/v1/projects/{id}/contacts` | Get contacts |

**Query Parameters:**
- `page`, `pageSize`
- `status` - Filter by status
- `managerId` - Filter by manager
- `customerId` - Filter by customer

### Activities

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/activities` | List activities |
| POST | `/api/v1/activities` | Create activity |
| GET | `/api/v1/activities/my-tasks` | Get current user's tasks |
| GET | `/api/v1/activities/upcoming` | Get upcoming scheduled activities |
| GET | `/api/v1/activities/stats` | Get activity statistics |
| GET | `/api/v1/activities/{id}` | Get activity |
| PUT | `/api/v1/activities/{id}` | Update activity |
| DELETE | `/api/v1/activities/{id}` | Delete activity |
| POST | `/api/v1/activities/{id}/complete` | Mark complete |
| POST | `/api/v1/activities/{id}/follow-up` | Create follow-up task |
| POST | `/api/v1/activities/{id}/attendees` | Add meeting attendee |
| DELETE | `/api/v1/activities/{id}/attendees/{userId}` | Remove attendee |

**Query Parameters:**
- `targetType`, `targetId` - Filter by entity
- `activityType` - Filter by type
- `status` - Filter by status
- `assignedToId` - Filter by assignee
- `dueDateFrom`, `dueDateTo` - Date range

### Notifications

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/notifications` | List notifications |
| GET | `/api/v1/notifications/count` | Get unread count |
| PUT | `/api/v1/notifications/read-all` | Mark all as read |
| GET | `/api/v1/notifications/{id}` | Get notification |
| PUT | `/api/v1/notifications/{id}/read` | Mark as read |

**Query Parameters:**
- `type` - Filter by notification type (`deal_update`, `offer_update`, `project_update`, `task_assigned`, `mention`, `system`)
- `unreadOnly` - Only show unread (default: false)

### Files

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/files/upload` | Upload file |
| GET | `/api/v1/files/{id}` | Get file metadata |
| GET | `/api/v1/files/{id}/download` | Download file |

### Dashboard & Search

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/dashboard/metrics` | Get dashboard metrics |
| GET | `/api/v1/search` | Global search |

**Search Query Parameters:**
- `q` - Search query (searches across customers, projects, offers, contacts)
- `type` - Filter results by type

### Audit Logs

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/audit` | List audit logs |
| GET | `/api/v1/audit/stats` | Get audit statistics |
| GET | `/api/v1/audit/export` | Export audit logs (CSV) |
| GET | `/api/v1/audit/entity/{type}/{id}` | Get logs for entity |
| GET | `/api/v1/audit/{id}` | Get specific log entry |

---

## Common Patterns

### Pagination

All list endpoints return paginated responses:

```json
{
  "data": [...],
  "total": 150,
  "page": 1,
  "pageSize": 20,
  "totalPages": 8
}
```

**Frontend Implementation:**
```typescript
const fetchPage = async (page: number, pageSize: number = 20) => {
  const response = await api.get(`/deals?page=${page}&pageSize=${pageSize}`);
  return {
    items: response.data.data,
    total: response.data.total,
    hasMore: page < response.data.totalPages
  };
};
```

### Error Handling

Errors return consistent structure:

```json
{
  "error": "Error type",
  "message": "Human-readable message",
  "code": 400
}
```

**Common Error Codes:**
- `400` - Bad Request (validation failed)
- `401` - Unauthorized (not authenticated)
- `403` - Forbidden (no permission)
- `404` - Not Found
- `409` - Conflict (e.g., duplicate)
- `422` - Unprocessable Entity (business rule violation)
- `500` - Internal Server Error

### Date/Time Format

All dates use ISO 8601 format:
```
2024-01-15T14:30:00Z
```

**Frontend Tip:** Use `date-fns` or `dayjs` for consistent parsing:
```typescript
import { parseISO, format } from 'date-fns';
import { nb } from 'date-fns/locale';

const formatDate = (isoString: string) =>
  format(parseISO(isoString), 'd. MMMM yyyy', { locale: nb });
```

### Denormalized Fields

Many entities include denormalized fields for display (e.g., `customerName`, `ownerName`). These are:
- Read-only in responses
- Automatically updated when related entities change
- Useful for list views without needing additional API calls

---

## Workflows & State Machines

### Deal Pipeline Workflow

```
┌────────┐    ┌───────────┐    ┌──────────┐    ┌─────────────┐    ┌───────┐
│  lead  │───▶│ qualified │───▶│ proposal │───▶│ negotiation │───▶│  won  │
│  (20%) │    │   (40%)   │    │  (60%)   │    │    (80%)    │    │ (100%)│
└────────┘    └───────────┘    └──────────┘    └──────────────┘    └───────┘
     │              │               │                │
     │              │               │                │
     └──────────────┴───────────────┴────────────────┴──────────────▶ lost
                                                                     (0%)
```

**Stage Transitions:**
- Forward: Use `POST /deals/{id}/advance`
- Win: Use `POST /deals/{id}/win`
- Lose: Use `POST /deals/{id}/lose` with reason
- Reopen: Use `POST /deals/{id}/reopen`

**Probability Auto-Update:** When stage changes, probability automatically updates to stage default.

**Losing a Deal - Required:**
```typescript
await api.post(`/deals/${id}/lose`, {
  reason: 'competitor',  // Required enum
  notes: 'Lost to competitor XYZ who offered lower price'  // Required, min 10 chars
});
```

### Offer Lifecycle

```
┌─────────┐    ┌─────────────┐    ┌────────┐    ┌───────┐
│  draft  │───▶│ in_progress │───▶│  sent  │───▶│  won  │
└─────────┘    └─────────────┘    └────────┘    └───────┘
                                       │             │
                                       ▼             │
                                  ┌─────────┐        │
                                  │  lost   │◀───────┘
                                  └─────────┘        │
                                       ▲             │
                                       └─────────────┘
                                             │
                                       ┌──────────┐
                                       │ expired  │
                                       └──────────┘
```

**Actions:**
- `POST /offers/{id}/advance` - Move to next phase
- `POST /offers/{id}/send` - Mark as sent to customer
- `POST /offers/{id}/accept` - Customer accepted
- `POST /offers/{id}/reject` - Customer rejected

**Accept with Project Creation:**
```typescript
const acceptOffer = async (offerId: string, createProject: boolean) => {
  return api.post(`/offers/${offerId}/accept`, {
    createProject: createProject,
    projectName: 'Project from Offer',
    managerId: currentUserId
  });
};
// Response includes both offer and project if created
```

### Project Status

```
┌──────────┐    ┌────────┐    ┌───────────┐
│ planning │───▶│ active │───▶│ completed │
└──────────┘    └────────┘    └───────────┘
                    │
                    ▼
               ┌─────────┐    ┌───────────┐
               │ on_hold │───▶│ cancelled │
               └─────────┘    └───────────┘
```

### Budget Inheritance Flow

When a project is created from an offer, or manually:

```typescript
// Create project from accepted offer (automatic)
const result = await api.post(`/offers/${offerId}/accept`, {
  createProject: true
});
// Budget dimensions are automatically inherited

// Manual inheritance for existing project
await api.post(`/projects/${projectId}/inherit-budget`, {
  offerId: sourceOfferId
});
```

**Important:** Budget dimensions are cloned (copied), not linked. Changes to offer dimensions after inheritance do not affect project dimensions.

---

## Frontend Implementation Guide

### Recommended Component Structure

```
src/
├── components/
│   ├── common/
│   │   ├── Pagination.tsx
│   │   ├── SearchInput.tsx
│   │   ├── StatusBadge.tsx
│   │   └── EntityLink.tsx
│   ├── customers/
│   │   ├── CustomerList.tsx
│   │   ├── CustomerForm.tsx
│   │   └── CustomerCard.tsx
│   ├── deals/
│   │   ├── DealPipeline.tsx      # Kanban view
│   │   ├── DealList.tsx          # Table view
│   │   ├── DealForm.tsx
│   │   └── DealStageHistory.tsx
│   ├── offers/
│   │   ├── OfferList.tsx
│   │   ├── OfferForm.tsx
│   │   ├── BudgetDimensionEditor.tsx
│   │   └── OfferTimeline.tsx
│   └── projects/
│       ├── ProjectList.tsx
│       ├── ProjectForm.tsx
│       ├── ProjectHealth.tsx
│       └── BudgetTracker.tsx
├── hooks/
│   ├── useAuth.ts
│   ├── usePagination.ts
│   ├── useNotifications.ts
│   └── useRealTimeUpdates.ts
└── services/
    ├── api.ts                    # Axios instance
    └── entities/
        ├── customerService.ts
        ├── dealService.ts
        └── ...
```

### Key UI Patterns

#### 1. Deal Pipeline Board (Kanban)

```typescript
const DealPipeline: React.FC = () => {
  const stages: DealStage[] = ['lead', 'qualified', 'proposal', 'negotiation'];

  const { data: pipeline } = useQuery(['pipeline'], () =>
    api.get('/deals/pipeline')
  );

  const advanceDeal = async (dealId: string, toStage: DealStage) => {
    await api.post(`/deals/${dealId}/advance`, { stage: toStage });
    queryClient.invalidateQueries(['pipeline']);
  };

  return (
    <DragDropContext onDragEnd={handleDragEnd}>
      {stages.map(stage => (
        <StageColumn key={stage} stage={stage} deals={pipeline[stage]} />
      ))}
    </DragDropContext>
  );
};
```

#### 2. Notification Polling

```typescript
const useNotifications = () => {
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    const pollInterval = setInterval(async () => {
      const { data } = await api.get('/notifications/count');
      setUnreadCount(data.count);
    }, 30000); // Poll every 30 seconds

    return () => clearInterval(pollInterval);
  }, []);

  return { unreadCount };
};
```

#### 3. Budget Dimension Editor

```typescript
const BudgetDimensionEditor: React.FC<{ offerId: string }> = ({ offerId }) => {
  const { data: dimensions } = useQuery(['dimensions', offerId], () =>
    api.get(`/offers/${offerId}/budget/dimensions`)
  );

  const addDimension = async (data: CreateDimensionRequest) => {
    await api.post(`/offers/${offerId}/budget/dimensions`, data);
    queryClient.invalidateQueries(['dimensions', offerId]);
    // Also refresh offer to get new totals
    queryClient.invalidateQueries(['offer', offerId]);
  };

  const reorderDimensions = async (orderedIds: string[]) => {
    await api.put(`/offers/${offerId}/budget/reorder`, { orderedIds });
    queryClient.invalidateQueries(['dimensions', offerId]);
  };

  return (/* Drag-and-drop list with inline editing */);
};
```

#### 4. Activity Timeline

```typescript
const ActivityTimeline: React.FC<{ entityType: string; entityId: string }> = ({
  entityType,
  entityId
}) => {
  const { data: activities } = useQuery(['activities', entityType, entityId], () =>
    api.get('/activities', {
      params: { targetType: entityType, targetId: entityId }
    })
  );

  return (
    <Timeline>
      {activities.map(activity => (
        <TimelineItem key={activity.id}>
          <ActivityIcon type={activity.activityType} />
          <ActivityContent activity={activity} />
        </TimelineItem>
      ))}
    </Timeline>
  );
};
```

### State Management Recommendations

**For Server State:** Use React Query (TanStack Query)
- Automatic caching and revalidation
- Optimistic updates for better UX
- Background refetching

**For UI State:** Use Zustand or Context
- Current user/auth state
- UI preferences
- Modal/drawer state

```typescript
// React Query setup
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,  // 5 minutes
      refetchOnWindowFocus: true,
    },
  },
});

// Entity hooks
const useDeal = (id: string) => useQuery(['deal', id], () =>
  api.get(`/deals/${id}`).then(r => r.data)
);

const useUpdateDeal = () => useMutation(
  (data: UpdateDealRequest) => api.put(`/deals/${data.id}`, data),
  {
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries(['deal', variables.id]);
      queryClient.invalidateQueries(['deals']);
    }
  }
);
```

---

## Pitfalls & Edge Cases

### 1. UUID Format

All IDs are UUIDs (36 characters with hyphens). Always validate before sending:

```typescript
const isValidUUID = (id: string) =>
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id);
```

### 2. Nullable vs Optional Fields

- `*string` in Go = nullable in JSON (can be `null`)
- Fields with `omitempty` may be absent from response
- Always check for both `null` and `undefined`

```typescript
const closeDate = deal.expectedCloseDate ?? deal.actualCloseDate ?? 'Not set';
```

### 3. Stage/Phase Validation

The API will reject invalid stage transitions:
- Cannot skip stages (lead directly to won)
- Cannot advance closed deals (won/lost)

Handle errors gracefully:
```typescript
try {
  await api.post(`/deals/${id}/advance`);
} catch (error) {
  if (error.response?.status === 422) {
    toast.error('Invalid stage transition');
  }
}
```

### 4. Multi-Tenant Data Isolation

Never assume you can access entities from other companies. The API will return 404 (not 403) for cross-company access to prevent information leakage.

### 5. Concurrent Updates

No built-in optimistic locking. For critical updates, consider:
- Fetch latest before updating
- Show "modified by another user" if `updatedAt` changed

```typescript
const updateWithCheck = async (entity: Entity, updates: Partial<Entity>) => {
  const current = await api.get(`/entities/${entity.id}`);
  if (current.updatedAt !== entity.updatedAt) {
    throw new Error('Entity was modified by another user');
  }
  return api.put(`/entities/${entity.id}`, updates);
};
```

### 6. File Upload Size Limits

- Max file size: 50MB
- Allowed types: Check Swagger docs
- Use multipart/form-data

```typescript
const uploadFile = async (offerId: string, file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('offerId', offerId);

  return api.post('/files/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  });
};
```

### 7. Rate Limiting

The API has IP-based rate limiting. Handle 429 responses:

```typescript
api.interceptors.response.use(null, async (error) => {
  if (error.response?.status === 429) {
    const retryAfter = error.response.headers['retry-after'] || 60;
    await sleep(retryAfter * 1000);
    return api.request(error.config);
  }
  throw error;
});
```

### 8. Notification Type Validation

When filtering notifications by type, invalid types return 400:

Valid types: `deal_update`, `offer_update`, `project_update`, `task_assigned`, `mention`, `system`

### 9. Loss Reason Required for Losing Deals

You cannot lose a deal without providing a reason:

```typescript
// Will fail with 400
await api.post(`/deals/${id}/lose`);

// Correct
await api.post(`/deals/${id}/lose`, {
  reason: 'price',
  notes: 'Customer chose competitor due to lower pricing'
});
```

### 10. Budget Dimension Order

Display order matters for budget dimensions. After reordering, you must send all IDs:

```typescript
// Correct - send ALL dimension IDs in new order
await api.put(`/offers/${id}/budget/reorder`, {
  orderedIds: ['uuid-3', 'uuid-1', 'uuid-2', 'uuid-4']
});

// Wrong - partial list
await api.put(`/offers/${id}/budget/reorder`, {
  orderedIds: ['uuid-3', 'uuid-1']  // Missing items!
});
```

---

## Recommended Features

### Dashboard Overview
- Pipeline value by stage (use `/deals/pipeline`)
- Revenue forecast chart (use `/deals/forecast`)
- Activity feed (use `/activities/upcoming`)
- Unread notifications badge (use `/notifications/count`)

### Quick Actions
- Create deal from customer page
- Create offer from deal
- Create project from accepted offer
- Create follow-up from completed activity

### Search
- Global search across all entities
- Show entity type icons in results
- Deep link to entity detail pages

### Activity Logging
- Auto-create activity when deal stage changes
- Link activities to relevant entities
- Show activity timeline on entity detail pages

### Offline Support
- Cache frequently accessed data
- Queue mutations for retry
- Show sync status indicator

### Analytics Dashboard
- Win rate trends (from `/deals/analytics`)
- Conversion funnel (from `conversionRates`)
- Owner performance metrics

---

## API Base URLs

| Environment | Base URL |
|-------------|----------|
| Development | `http://localhost:8080` |
| Production | `https://api.relation.straye.no` |

**Swagger UI:** `{base_url}/swagger/index.html`

---

## Support & Debugging

### Enable Debug Logging
```typescript
api.interceptors.request.use(request => {
  console.log('Request:', request.method, request.url, request.data);
  return request;
});

api.interceptors.response.use(
  response => {
    console.log('Response:', response.status, response.data);
    return response;
  },
  error => {
    console.error('Error:', error.response?.status, error.response?.data);
    throw error;
  }
);
```

### Health Checks
- Basic liveness: `GET /health` (returns "OK")
- Database check: `GET /health/db` (with connection stats)
- Full readiness: `GET /health/ready` (all dependencies)

---

*Last updated: December 2024*
*API Version: 1.0*
