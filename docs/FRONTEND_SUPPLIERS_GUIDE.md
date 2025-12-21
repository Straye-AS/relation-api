# Suppliers Feature - Frontend Implementation Guide

This document provides all the information needed by the frontend team to implement the new Suppliers feature in the Relation API.

## Table of Contents

1. [Overview](#overview)
2. [API Base URL](#api-base-url)
3. [Authentication](#authentication)
4. [Supplier Endpoints](#supplier-endpoints)
5. [Data Types](#data-types)
6. [Usage Examples](#usage-examples)
7. [Error Handling](#error-handling)
8. [Integration with Existing Features](#integration-with-existing-features)

---

## Overview

The Suppliers feature allows managing supplier/vendor relationships for construction projects. Suppliers can be:
- Created, updated, and soft-deleted
- Linked to offers (many-to-many relationship)
- Searched via global search
- Filtered and sorted in list views

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Supplier** | A vendor/contractor that provides materials or services |
| **Supplier Status** | `active`, `inactive`, `pending`, `blacklisted` |
| **Supplier Contact** | Contact person(s) associated with a supplier |
| **Offer-Supplier** | Junction linking suppliers to offers with status tracking |

---

## API Base URL

```
Production: https://api.relation.straye.no/api/v1
Development: http://localhost:8080/api/v1
```

---

## Authentication

All endpoints require authentication via:
- **Bearer Token**: `Authorization: Bearer <jwt_token>`
- **API Key**: `x-api-key: <api_key>` (for system integrations)

---

## Supplier Endpoints

### 1. List Suppliers

```http
GET /suppliers
```

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `pageSize` | int | 20 | Items per page (max 200) |
| `search` | string | - | Search by name or org number |
| `city` | string | - | Filter by city |
| `country` | string | - | Filter by country |
| `status` | string | - | Filter by status: `active`, `inactive`, `pending`, `blacklisted` |
| `category` | string | - | Filter by category |
| `sortBy` | string | `createdAt` | Sort field: `createdAt`, `updatedAt`, `name`, `city`, `country`, `status`, `category`, `orgNumber` |
| `sortOrder` | string | `desc` | Sort order: `asc`, `desc` |

**Response:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Bygg AS",
      "orgNumber": "912345678",
      "email": "post@bygg.no",
      "phone": "+47 22 33 44 55",
      "address": "Storgata 1",
      "city": "Oslo",
      "postalCode": "0151",
      "country": "Norge",
      "municipality": "Oslo",
      "county": "Oslo",
      "contactPerson": "Per Hansen",
      "contactEmail": "per@bygg.no",
      "contactPhone": "+47 91 23 45 67",
      "status": "active",
      "category": "Taktekking",
      "notes": "Foretrukket leverandor for takprosjekter",
      "paymentTerms": "30 dager netto",
      "website": "https://bygg.no",
      "createdAt": "2025-01-15T10:30:00Z",
      "updatedAt": "2025-01-20T14:22:00Z",
      "createdById": "user-uuid",
      "createdByName": "Ola Nordmann",
      "updatedById": "user-uuid",
      "updatedByName": "Kari Hansen"
    }
  ],
  "page": 1,
  "pageSize": 20,
  "totalItems": 45,
  "totalPages": 3
}
```

---

### 2. Get Supplier by ID

```http
GET /suppliers/{id}
```

Returns supplier with full details including stats, contacts, and recent offers.

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Bygg AS",
  "orgNumber": "912345678",
  "email": "post@bygg.no",
  "status": "active",
  "category": "Taktekking",
  "...": "...other supplier fields...",
  "stats": {
    "totalOffers": 15,
    "activeOffers": 3,
    "completedOffers": 12,
    "totalProjects": 8
  },
  "contacts": [
    {
      "id": "uuid",
      "supplierId": "supplier-uuid",
      "name": "Per Hansen",
      "title": "Daglig leder",
      "email": "per@bygg.no",
      "phone": "+47 91 23 45 67",
      "isPrimary": true,
      "notes": "",
      "createdAt": "2025-01-15T10:30:00Z",
      "updatedAt": "2025-01-15T10:30:00Z"
    }
  ],
  "recentOffers": [
    {
      "id": "uuid",
      "offerId": "offer-uuid",
      "offerTitle": "Takarbeider Bygning A",
      "supplierId": "supplier-uuid",
      "supplierName": "Bygg AS",
      "status": "active",
      "notes": "Avtalt pris 250 000 NOK",
      "createdAt": "2025-01-18T09:00:00Z",
      "updatedAt": "2025-01-18T09:00:00Z"
    }
  ]
}
```

---

### 3. Create Supplier

```http
POST /suppliers
```

**Request Body:**

```json
{
  "name": "Bygg AS",
  "orgNumber": "912345678",
  "email": "post@bygg.no",
  "phone": "+47 22 33 44 55",
  "address": "Storgata 1",
  "city": "Oslo",
  "postalCode": "0151",
  "country": "Norge",
  "municipality": "Oslo",
  "county": "Oslo",
  "contactPerson": "Per Hansen",
  "contactEmail": "per@bygg.no",
  "contactPhone": "+47 91 23 45 67",
  "status": "active",
  "category": "Taktekking",
  "notes": "Foretrukket leverandor",
  "paymentTerms": "30 dager netto",
  "website": "https://bygg.no"
}
```

**Required Fields:**
- `name` (max 200 chars)
- `country` (max 100 chars)

**Optional Fields:** All other fields are optional.

**Response:** `201 Created` with the created supplier object.

---

### 4. Update Supplier

```http
PUT /suppliers/{id}
```

**Request Body:** Same structure as Create, all fields included.

**Response:** `200 OK` with the updated supplier object.

---

### 5. Delete Supplier (Soft Delete)

```http
DELETE /suppliers/{id}
```

**Response:** `204 No Content`

**Note:** Returns `409 Conflict` if supplier has active offer relationships.

---

### 6. Property Update Endpoints

For updating individual properties without sending the full object:

#### Update Status
```http
PUT /suppliers/{id}/status
```
```json
{
  "status": "active"  // active, inactive, pending, blacklisted
}
```

#### Update Notes
```http
PUT /suppliers/{id}/notes
```
```json
{
  "notes": "Updated notes text"
}
```

#### Update Category
```http
PUT /suppliers/{id}/category
```
```json
{
  "category": "Elektro"
}
```

#### Update Payment Terms
```http
PUT /suppliers/{id}/payment-terms
```
```json
{
  "paymentTerms": "45 dager netto"
}
```

---

## Data Types

### SupplierStatus

| Value | Description |
|-------|-------------|
| `active` | Active supplier, can be used in offers |
| `inactive` | Temporarily inactive, not shown in selections |
| `pending` | Under review/approval |
| `blacklisted` | Do not use (hidden from searches) |

### SupplierDTO

```typescript
interface SupplierDTO {
  id: string;                    // UUID
  name: string;
  orgNumber?: string;            // Norwegian org number
  email?: string;
  phone?: string;
  address?: string;
  city?: string;
  postalCode?: string;
  country: string;
  municipality?: string;
  county?: string;
  contactPerson?: string;
  contactEmail?: string;
  contactPhone?: string;
  status: 'active' | 'inactive' | 'pending' | 'blacklisted';
  category?: string;             // e.g., "Taktekking", "Elektro", "VVS"
  notes?: string;
  paymentTerms?: string;         // e.g., "30 dager netto"
  website?: string;
  createdAt: string;             // ISO 8601
  updatedAt: string;             // ISO 8601
  createdById?: string;
  createdByName?: string;
  updatedById?: string;
  updatedByName?: string;
}
```

### SupplierWithDetailsDTO

```typescript
interface SupplierWithDetailsDTO extends SupplierDTO {
  stats?: {
    totalOffers: number;
    activeOffers: number;
    completedOffers: number;
    totalProjects: number;
  };
  contacts?: SupplierContactDTO[];
  recentOffers?: OfferSupplierDTO[];
}
```

### SupplierContactDTO

```typescript
interface SupplierContactDTO {
  id: string;
  supplierId: string;
  name: string;
  title?: string;
  email?: string;
  phone?: string;
  isPrimary: boolean;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}
```

### OfferSupplierDTO

```typescript
interface OfferSupplierDTO {
  id: string;
  offerId: string;
  offerTitle?: string;
  supplierId: string;
  supplierName?: string;
  status: 'active' | 'done';
  notes?: string;
  createdAt: string;
  updatedAt: string;
}
```

---

## Usage Examples

### React/TypeScript Examples

#### Fetching Suppliers with Filters

```typescript
async function fetchSuppliers(params: {
  page?: number;
  pageSize?: number;
  search?: string;
  status?: string;
  category?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}) {
  const searchParams = new URLSearchParams();

  if (params.page) searchParams.set('page', params.page.toString());
  if (params.pageSize) searchParams.set('pageSize', params.pageSize.toString());
  if (params.search) searchParams.set('search', params.search);
  if (params.status) searchParams.set('status', params.status);
  if (params.category) searchParams.set('category', params.category);
  if (params.sortBy) searchParams.set('sortBy', params.sortBy);
  if (params.sortOrder) searchParams.set('sortOrder', params.sortOrder);

  const response = await fetch(`/api/v1/suppliers?${searchParams}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error('Failed to fetch suppliers');
  }

  return response.json();
}
```

#### Creating a Supplier

```typescript
async function createSupplier(supplier: CreateSupplierRequest) {
  const response = await fetch('/api/v1/suppliers', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(supplier),
  });

  if (response.status === 409) {
    throw new Error('A supplier with this organization number already exists');
  }

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to create supplier');
  }

  return response.json();
}
```

#### Quick Status Update

```typescript
async function updateSupplierStatus(id: string, status: SupplierStatus) {
  const response = await fetch(`/api/v1/suppliers/${id}/status`, {
    method: 'PUT',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ status }),
  });

  if (!response.ok) {
    throw new Error('Failed to update status');
  }

  return response.json();
}
```

---

## Error Handling

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `204` | No Content (successful delete) |
| `400` | Bad Request - validation error |
| `401` | Unauthorized - missing/invalid auth |
| `404` | Not Found - supplier doesn't exist |
| `409` | Conflict - duplicate org number or active relationships |
| `500` | Internal Server Error |

### Error Response Format

```json
{
  "error": "Bad Request",
  "message": "Invalid email format"
}
```

### Common Errors

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `"A supplier with this organization number already exists"` | Duplicate org number | Use different org number or update existing |
| `"Supplier not found"` | Invalid ID | Verify supplier ID exists |
| `"Cannot delete supplier with active offer relationships"` | Supplier linked to offers | Remove offer links first or mark as inactive |
| `"Invalid email format"` | Bad email | Fix email format |

---

## Integration with Existing Features

### Global Search

Suppliers are included in the global search endpoint:

```http
GET /search?q=bygg
```

Response includes suppliers matching the search query in the `suppliers` array.

### Offers Integration

When viewing an offer, you can fetch linked suppliers. When creating/editing offers, you can link suppliers.

### Dashboard

Supplier statistics are available for dashboard widgets (coming soon).

---

## UI Recommendations

### Supplier List View

- Show: Name, Org Number, City, Status, Category
- Filters: Status dropdown, Category dropdown, Search input
- Sorting: Name, City, Status, Created date
- Actions: View, Edit, Delete, Quick status change

### Supplier Detail View

- Header: Name, Status badge, Category tag
- Tabs: Details, Contacts, Offers, Files, Activity
- Stats cards: Active Offers, Total Offers, Projects

### Status Colors

| Status | Color | Icon |
|--------|-------|------|
| `active` | Green | Check circle |
| `inactive` | Gray | Pause circle |
| `pending` | Yellow | Clock |
| `blacklisted` | Red | Ban |

---

## Questions?

Contact the backend team or check the Swagger documentation at `/swagger/index.html` for the most up-to-date API reference.
