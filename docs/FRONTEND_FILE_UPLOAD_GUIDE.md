# Frontend File Upload Implementation Guide

This guide explains how to integrate file upload functionality into the Straye Relation frontend application.

## Overview

The file upload system allows attaching files to four entity types:
- **Customers** - Documents, contracts, correspondence
- **Projects** - Project documents, plans, reports
- **Offers** - Proposals, quotes, specifications
- **Suppliers** - Agreements, certifications, invoices

Files are stored in Azure Blob Storage (production) or local filesystem (development).

## API Endpoints

### Entity-Specific Endpoints

Each entity type has dedicated upload and list endpoints:

| Entity | Upload | List |
|--------|--------|------|
| Customer | `POST /customers/{id}/files` | `GET /customers/{id}/files` |
| Project | `POST /projects/{id}/files` | `GET /projects/{id}/files` |
| Offer | `POST /offers/{id}/files` | `GET /offers/{id}/files` |
| Supplier | `POST /suppliers/{id}/files` | `GET /suppliers/{id}/files` |

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
  id: string;           // UUID
  filename: string;     // Original filename
  contentType: string;  // MIME type (e.g., "application/pdf")
  size: number;         // File size in bytes
  offerId?: string;     // UUID if attached to offer
  customerId?: string;  // UUID if attached to customer
  projectId?: string;   // UUID if attached to project
  supplierId?: string;  // UUID if attached to supplier
  createdAt: string;    // ISO 8601 timestamp
}
```

## Implementation Examples

### 1. Upload File to Entity

```typescript
// Generic upload function for any entity type
async function uploadFile(
  entityType: 'customers' | 'projects' | 'offers' | 'suppliers',
  entityId: string,
  file: File
): Promise<FileDTO> {
  const formData = new FormData();
  formData.append('file', file);

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
await uploadFile('customers', customerId, file);
await uploadFile('offers', offerId, file);
await uploadFile('projects', projectId, file);
await uploadFile('suppliers', supplierId, file);
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

interface FileUploadProps {
  entityType: 'customers' | 'projects' | 'offers' | 'suppliers';
  entityId: string;
  onUploadComplete?: (file: FileDTO) => void;
}

export function FileUpload({ entityType, entityId, onUploadComplete }: FileUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleUpload = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setUploading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append('file', file);

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
  }, [entityType, entityId, onUploadComplete]);

  return (
    <div>
      <input
        type="file"
        onChange={handleUpload}
        disabled={uploading}
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
| 400 | Bad request (invalid UUID, missing file) |
| 404 | Entity or file not found |
| 413 | File too large (default limit: 50MB) |
| 500 | Server error |

Error response format:
```json
{
  "error": "Error message here"
}
```

## Best Practices

1. **Show upload progress** - For large files, consider using `XMLHttpRequest` with progress events
2. **Validate file types** - Check file extensions/MIME types before upload if needed
3. **Handle errors gracefully** - Display user-friendly error messages
4. **Confirm deletes** - Always confirm before deleting files
5. **Refresh lists** - Reload file list after upload/delete operations
6. **File size display** - Format file sizes in human-readable format (KB, MB)

## Integration Points

Add file upload/list components to these views:

| View | Entity Type | Component Location |
|------|-------------|---------------------|
| Customer Detail | `customers` | Files tab or section |
| Project Detail | `projects` | Documents section |
| Offer Detail | `offers` | Attachments section |
| Supplier Detail | `suppliers` | Documents section |

## Authentication

All endpoints require authentication via:
- **Bearer Token**: `Authorization: Bearer <jwt_token>` (user requests)
- **API Key**: `x-api-key: <api_key>` (system/integration requests)

Use the same authentication method as other API calls in your application.
