package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/straye-as/relation-api/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Storage Interface Tests
// ============================================================================

// TestStorageInterfaceCompliance verifies that all storage implementations
// properly implement the Storage interface.
func TestStorageInterfaceCompliance(t *testing.T) {
	// This test ensures compile-time interface compliance
	var _ storage.Storage = (*storage.LocalStorage)(nil)
	var _ storage.Storage = (*storage.AzureBlobStorage)(nil)
}

// ============================================================================
// LocalStorage Tests
// ============================================================================

func TestNewLocalStorage_CreatesDirectory(t *testing.T) {
	tempDir := t.TempDir()
	basePath := filepath.Join(tempDir, "uploads")

	ls, err := storage.NewLocalStorage(basePath)

	require.NoError(t, err)
	assert.NotNil(t, ls)

	// Verify directory was created
	info, err := os.Stat(basePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewLocalStorage_ExistingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	ls, err := storage.NewLocalStorage(tempDir)

	require.NoError(t, err)
	assert.NotNil(t, ls)
}

func TestLocalStorage_Upload(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	tests := []struct {
		name        string
		filename    string
		contentType string
		content     []byte
	}{
		{
			name:        "upload PDF file",
			filename:    "document.pdf",
			contentType: "application/pdf",
			content:     []byte("fake pdf content"),
		},
		{
			name:        "upload image file",
			filename:    "photo.jpg",
			contentType: "image/jpeg",
			content:     []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG magic bytes
		},
		{
			name:        "upload text file",
			filename:    "notes.txt",
			contentType: "text/plain",
			content:     []byte("Hello, World!"),
		},
		{
			name:        "upload file with special characters",
			filename:    "file with spaces.docx",
			contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			content:     []byte("docx content"),
		},
		{
			name:        "upload empty file",
			filename:    "empty.txt",
			contentType: "text/plain",
			content:     []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.content)

			storagePath, size, err := ls.Upload(context.Background(), tt.filename, tt.contentType, reader)

			require.NoError(t, err)
			assert.NotEmpty(t, storagePath)
			assert.Equal(t, int64(len(tt.content)), size)

			// Verify file was created and has correct extension
			ext := filepath.Ext(tt.filename)
			assert.True(t, filepath.Ext(storagePath) == ext || (ext == "" && filepath.Ext(storagePath) == ""))
		})
	}
}

func TestLocalStorage_Download(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Upload a file first
	content := []byte("test content for download")
	storagePath, _, err := ls.Upload(context.Background(), "test.txt", "text/plain", bytes.NewReader(content))
	require.NoError(t, err)

	// Download the file
	reader, err := ls.Download(context.Background(), storagePath)
	require.NoError(t, err)
	defer reader.Close()

	// Read and verify content
	downloaded, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, content, downloaded)
}

func TestLocalStorage_Download_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	reader, err := ls.Download(context.Background(), "nonexistent/file.txt")

	assert.Error(t, err)
	assert.Nil(t, reader)
	assert.Contains(t, err.Error(), "file not found")
}

func TestLocalStorage_Delete(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Upload a file first
	content := []byte("file to be deleted")
	storagePath, _, err := ls.Upload(context.Background(), "delete-me.txt", "text/plain", bytes.NewReader(content))
	require.NoError(t, err)

	// Delete the file
	err = ls.Delete(context.Background(), storagePath)
	require.NoError(t, err)

	// Verify file is gone
	fullPath := filepath.Join(tempDir, storagePath)
	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err))
}

func TestLocalStorage_Delete_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Deleting a non-existent file should not return an error
	err = ls.Delete(context.Background(), "nonexistent/file.txt")
	assert.NoError(t, err)
}

func TestLocalStorage_Delete_Idempotent(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Upload a file
	content := []byte("delete me twice")
	storagePath, _, err := ls.Upload(context.Background(), "double-delete.txt", "text/plain", bytes.NewReader(content))
	require.NoError(t, err)

	// Delete twice - second delete should succeed (idempotent)
	err = ls.Delete(context.Background(), storagePath)
	require.NoError(t, err)

	err = ls.Delete(context.Background(), storagePath)
	assert.NoError(t, err)
}

func TestLocalStorage_UploadDownloadRoundtrip(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Test with various content sizes
	testCases := []struct {
		name    string
		size    int
		content []byte
	}{
		{"small file", 10, []byte("small file")},
		{"medium file", 1024, bytes.Repeat([]byte("x"), 1024)},
		{"large file", 1024 * 100, bytes.Repeat([]byte("L"), 1024*100)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Upload
			storagePath, size, err := ls.Upload(context.Background(), "test.bin", "application/octet-stream", bytes.NewReader(tc.content))
			require.NoError(t, err)
			assert.Equal(t, int64(len(tc.content)), size)

			// Download
			reader, err := ls.Download(context.Background(), storagePath)
			require.NoError(t, err)
			defer reader.Close()

			downloaded, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, tc.content, downloaded)

			// Cleanup
			err = ls.Delete(context.Background(), storagePath)
			require.NoError(t, err)
		})
	}
}

func TestLocalStorage_UniqueStoragePaths(t *testing.T) {
	tempDir := t.TempDir()
	ls, err := storage.NewLocalStorage(tempDir)
	require.NoError(t, err)

	// Upload the same filename multiple times
	paths := make(map[string]bool)
	content := []byte("same content")

	for i := 0; i < 5; i++ {
		storagePath, _, err := ls.Upload(context.Background(), "same-name.txt", "text/plain", bytes.NewReader(content))
		require.NoError(t, err)

		// Each path should be unique
		assert.False(t, paths[storagePath], "storage path should be unique: %s", storagePath)
		paths[storagePath] = true
	}

	assert.Len(t, paths, 5)
}

// ============================================================================
// AzureBlobStorage Tests (Interface Verification Only)
// ============================================================================

// Note: Full integration tests for Azure Blob Storage require a real Azure
// connection and are handled by integration tests. These unit tests verify
// the struct implements the interface and test error scenarios that can be
// mocked.

func TestAzureBlobStorage_ImplementsStorageInterface(t *testing.T) {
	// This is a compile-time check that AzureBlobStorage implements Storage
	var _ storage.Storage = (*storage.AzureBlobStorage)(nil)
}

// ============================================================================
// NewStorage Factory Tests
// ============================================================================

func TestNewStorage_LocalMode(t *testing.T) {
	// Note: This test would require importing config package and setting up
	// a proper config. Since we're testing the storage package in isolation,
	// we test the LocalStorage directly instead.

	// The NewStorage factory is tested implicitly through the LocalStorage tests above
	// and through integration tests that use the full config.
	t.Skip("Factory function tested through integration tests with full config")
}

func TestNewStorage_InvalidMode(t *testing.T) {
	// The NewStorage factory returns an error for unsupported modes.
	// This is tested in integration tests with the full config.
	t.Skip("Factory function tested through integration tests with full config")
}
