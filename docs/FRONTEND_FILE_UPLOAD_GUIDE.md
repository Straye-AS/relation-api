# Frontend File Upload Implementation Guide

This guide explains how to integrate file upload functionality into the Straye Relation frontend application.

## Overview

The file upload system allows attaching files to:
- **Customers** - Documents, contracts, correspondence
- **Projects** - Project documents, plans, reports
- **Offers** - Proposals, quotes, specifications
- **Suppliers** - Agreements, certifications, invoices
- **Offer-Supplier** - Supplier quotes/documents specific to an offer

Files are stored in Azure Blob Storage (production) or local filesystem (development).

## Company Association (Required)

**All files must be associated with a company.** This enables proper multi-tenant filtering and access control.

### Valid Company IDs

| ID | Company |
|----|---------|
| `gruppen` | Straye Gruppen (parent company / "all") |
| `stalbygg` | Stålbygg |
| `hybridbygg` | Hybridbygg |
| `industri` | Industri |
| `tak` | Tak |
| `montasje` | Montasje |

### Company Inheritance Rules

The API automatically inherits the company from the parent entity when possible:

| Entity Type | Company Behavior |
|-------------|------------------|
| **Customer files** | Inherits from customer's company, defaults to `gruppen` |
| **Offer files** | Inherits from offer's company |
| **Project files** | ⚠️ **Requires explicit `company_id`** (projects are cross-company) |
| **Supplier files** | Inherits from supplier's company, defaults to `gruppen` |
| **Offer-Supplier files** | Inherits from offer's company |

**Important:** For project uploads, you MUST provide `company_id` in the form data.

## API Endpoints

### Entity-Specific Endpoints

Each entity type has dedicated upload and list endpoints:

| Entity | Upload | List |
|--------|--------|------|
| Customer | `POST /customers/{id}/files` | `GET /customers/{id}/files` |
| Project | `POST /projects/{id}/files` | `GET /projects/{id}/files` |
| Offer | `POST /offers/{id}/files` | `GET /offers/{id}/files` |
| Supplier | `POST /suppliers/{id}/files` | `GET /suppliers/{id}/files` |
| Offer-Supplier | `POST /offers/{offerId}/suppliers/{supplierId}/files` | `GET /offers/{offerId}/suppliers/{supplierId}/files` |

### Generic File Operations

| Operation | Endpoint | Description |
|-----------|----------|-------------|
| Get metadata | `GET /files/{id}` | Get file details without downloading |
| Download | `GET /files/{id}/download` | Download file content |
| Delete | `DELETE /files/{id}` | Delete file permanently |

## Response Format

### FileDTO

All file operations return this structure:

```typescript
interface FileDTO {
  id: string;              // UUID
  filename: string;        // Original filename
  contentType: string;     // MIME type (e.g., "application/pdf")
  size: number;            // File size in bytes
  companyId: string;       // Company ID (gruppen, stalbygg, etc.)
  offerId?: string;        // UUID if attached to offer
  customerId?: string;     // UUID if attached to customer
  projectId?: string;      // UUID if attached to project
  supplierId?: string;     // UUID if attached to supplier
  offerSupplierId?: string; // UUID if attached to offer-supplier relationship
  createdAt: string;       // ISO 8601 timestamp
}
```

## Implementation Examples

### 1. Upload File to Entity

```typescript
// Valid company IDs
type CompanyID = 'gruppen' | 'stalbygg' | 'hybridbygg' | 'industri' | 'tak' | 'montasje';

// Generic upload function for any entity type
async function uploadFile(
  entityType: 'customers' | 'projects' | 'offers' | 'suppliers',
  entityId: string,
  file: File,
  companyId?: CompanyID  // Required for projects, optional for others
): Promise<FileDTO> {
  const formData = new FormData();
  formData.append('file', file);

  // Add company_id if provided (required for projects!)
  if (companyId) {
    formData.append('company_id', companyId);
  }

  const response = await fetch(`/api/${entityType}/${entityId}/files`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
      // Note: Do NOT set Content-Type header - browser sets it with boundary
    },
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Upload failed');
  }

  return response.json();
}

// Usage examples
await uploadFile('customers', customerId, file);                    // Inherits from customer
await uploadFile('offers', offerId, file);                          // Inherits from offer
await uploadFile('projects', projectId, file, 'stalbygg');          // REQUIRED: explicit company
await uploadFile('suppliers', supplierId, file);                    // Inherits from supplier
await uploadFile('customers', customerId, file, 'hybridbygg');      // Override inherited company
```

### 1b. Upload File to Offer-Supplier (Supplier within an Offer)

```typescript
// Upload a file specific to a supplier's involvement in an offer
async function uploadOfferSupplierFile(
  offerId: string,
  supplierId: string,
  file: File,
  companyId?: CompanyID  // Optional - inherits from offer
): Promise<FileDTO> {
  const formData = new FormData();
  formData.append('file', file);

  if (companyId) {
    formData.append('company_id', companyId);
  }

  const response = await fetch(`/api/offers/${offerId}/suppliers/${supplierId}/files`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
    body: formData,
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Upload failed');
  }

  return response.json();
}

// List files for a specific supplier within an offer
async function listOfferSupplierFiles(
  offerId: string,
  supplierId: string
): Promise<FileDTO[]> {
  const response = await fetch(`/api/offers/${offerId}/suppliers/${supplierId}/files`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to load files');
  }

  return response.json();
}
```

### 2. List Entity Files

```typescript
async function listFiles(
  entityType: 'customers' | 'projects' | 'offers' | 'suppliers',
  entityId: string
): Promise<FileDTO[]> {
  const response = await fetch(`/api/${entityType}/${entityId}/files`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error('Failed to load files');
  }

  return response.json();
}
```

### 3. Download File

```typescript
async function downloadFile(fileId: string, filename: string): Promise<void> {
  const response = await fetch(`/api/files/${fileId}/download`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error('Download failed');
  }

  // Create download link
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
}
```

### 4. Delete File

```typescript
async function deleteFile(fileId: string): Promise<void> {
  const response = await fetch(`/api/files/${fileId}`, {
    method: 'DELETE',
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  });

  if (!response.ok) {
    throw new Error('Delete failed');
  }
}
```

## React Component Example

```tsx
import { useState, useCallback } from 'react';

type CompanyID = 'gruppen' | 'stalbygg' | 'hybridbygg' | 'industri' | 'tak' | 'montasje';

interface FileUploadProps {
  entityType: 'customers' | 'projects' | 'offers' | 'suppliers';
  entityId: string;
  companyId?: CompanyID;           // Required for projects
  requireCompanySelect?: boolean;  // Show company dropdown
  onUploadComplete?: (file: FileDTO) => void;
}

const COMPANIES: { id: CompanyID; name: string }[] = [
  { id: 'gruppen', name: 'Straye Gruppen (All)' },
  { id: 'stalbygg', name: 'Stålbygg' },
  { id: 'hybridbygg', name: 'Hybridbygg' },
  { id: 'industri', name: 'Industri' },
  { id: 'tak', name: 'Tak' },
  { id: 'montasje', name: 'Montasje' },
];

export function FileUpload({
  entityType,
  entityId,
  companyId: defaultCompanyId,
  requireCompanySelect = false,
  onUploadComplete
}: FileUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedCompany, setSelectedCompany] = useState<CompanyID | undefined>(defaultCompanyId);

  // Projects always require company selection
  const showCompanySelect = requireCompanySelect || entityType === 'projects';

  const handleUpload = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate company for projects
    if (entityType === 'projects' && !selectedCompany) {
      setError('Please select a company for project files');
      return;
    }

    setUploading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append('file', file);

      // Add company_id if available
      if (selectedCompany) {
        formData.append('company_id', selectedCompany);
      }

      const response = await fetch(`/api/${entityType}/${entityId}/files`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.message || 'Upload failed');
      }

      const uploadedFile = await response.json();
      onUploadComplete?.(uploadedFile);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploading(false);
      e.target.value = ''; // Reset input
    }
  }, [entityType, entityId, selectedCompany, onUploadComplete]);

  return (
    <div className="file-upload">
      {showCompanySelect && (
        <select
          value={selectedCompany || ''}
          onChange={(e) => setSelectedCompany(e.target.value as CompanyID)}
          disabled={uploading}
        >
          <option value="">Select company...</option>
          {COMPANIES.map(c => (
            <option key={c.id} value={c.id}>{c.name}</option>
          ))}
        </select>
      )}
      <input
        type="file"
        onChange={handleUpload}
        disabled={uploading || (showCompanySelect && !selectedCompany)}
      />
      {uploading && <span>Uploading...</span>}
      {error && <span className="error">{error}</span>}
    </div>
  );
}
```

## File List Component Example

```tsx
interface FileListProps {
  entityType: 'customers' | 'projects' | 'offers' | 'suppliers';
  entityId: string;
}

export function FileList({ entityType, entityId }: FileListProps) {
  const [files, setFiles] = useState<FileDTO[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadFiles();
  }, [entityType, entityId]);

  const loadFiles = async () => {
    try {
      const response = await fetch(`/api/${entityType}/${entityId}/files`, {
        headers: { 'Authorization': `Bearer ${getAccessToken()}` },
      });
      if (response.ok) {
        setFiles(await response.json());
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async (file: FileDTO) => {
    const response = await fetch(`/api/files/${file.id}/download`, {
      headers: { 'Authorization': `Bearer ${getAccessToken()}` },
    });
    const blob = await response.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = file.filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleDelete = async (fileId: string) => {
    if (!confirm('Are you sure you want to delete this file?')) return;

    await fetch(`/api/files/${fileId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${getAccessToken()}` },
    });
    setFiles(files.filter(f => f.id !== fileId));
  };

  const getCompanyName = (companyId: string): string => {
    const names: Record<string, string> = {
      gruppen: 'Straye Gruppen',
      stalbygg: 'Stålbygg',
      hybridbygg: 'Hybridbygg',
      industri: 'Industri',
      tak: 'Tak',
      montasje: 'Montasje',
    };
    return names[companyId] || companyId;
  };

  if (loading) return <div>Loading files...</div>;

  return (
    <div>
      <h3>Files ({files.length})</h3>
      {files.length === 0 ? (
        <p>No files attached</p>
      ) : (
        <ul>
          {files.map(file => (
            <li key={file.id}>
              <span>{file.filename}</span>
              <span className="company-badge">{getCompanyName(file.companyId)}</span>
              <span>{formatFileSize(file.size)}</span>
              <button onClick={() => handleDownload(file)}>Download</button>
              <button onClick={() => handleDelete(file.id)}>Delete</button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
```

## Error Handling

The API returns standard error responses:

| Status | Meaning |
|--------|---------|
| 400 | Bad request (invalid UUID, missing file, invalid company_id) |
| 404 | Entity or file not found |
| 413 | File too large (default limit: 50MB) |
| 500 | Server error |

Error response format:
```json
{
  "error": "Error message here"
}
```

### Common Errors

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `company_id is required for project files` | Project upload without company | Add `company_id` to form data |
| `invalid company_id` | Unknown company value | Use valid company ID from list |
| `file is required` | No file in form data | Ensure `file` field is populated |

## Best Practices

1. **Always provide company for projects** - Project uploads require explicit `company_id`
2. **Show upload progress** - For large files, consider using `XMLHttpRequest` with progress events
3. **Validate file types** - Check file extensions/MIME types before upload if needed
4. **Handle errors gracefully** - Display user-friendly error messages
5. **Confirm deletes** - Always confirm before deleting files
6. **Refresh lists** - Reload file list after upload/delete operations
7. **File size display** - Format file sizes in human-readable format (KB, MB)
8. **Show company badge** - Display the company association in file lists

## Integration Points

Add file upload/list components to these views:

| View | Entity Type | Company Handling |
|------|-------------|------------------|
| Customer Detail | `customers` | Inherits from customer |
| Project Detail | `projects` | **Show company selector** |
| Offer Detail | `offers` | Inherits from offer |
| Supplier Detail | `suppliers` | Inherits from supplier |
| Offer Supplier Section | `offer-supplier` | Inherits from offer |

### Offer-Supplier Files Use Case

When viewing an offer with suppliers, each supplier card/section should have its own file upload area. These files are specific to that supplier's involvement in THIS offer (e.g., supplier quotes, contracts for this job) - not global supplier documents.

```tsx
// Example: Supplier section within offer detail page
function OfferSupplierCard({ offerId, supplier }: { offerId: string; supplier: OfferSupplier }) {
  return (
    <div className="supplier-card">
      <h4>{supplier.name}</h4>
      <p>Status: {supplier.status}</p>

      {/* Files specific to this supplier for this offer */}
      {/* Company is inherited from the offer automatically */}
      <FileList
        endpoint={`/api/offers/${offerId}/suppliers/${supplier.id}/files`}
      />
      <FileUpload
        endpoint={`/api/offers/${offerId}/suppliers/${supplier.id}/files`}
      />
    </div>
  );
}
```

## Authentication

All endpoints require authentication via:
- **Bearer Token**: `Authorization: Bearer <jwt_token>` (user requests)
- **API Key**: `x-api-key: <api_key>` (system/integration requests)

Use the same authentication method as other API calls in your application.
