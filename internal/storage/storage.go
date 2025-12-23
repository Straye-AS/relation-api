package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/config"
	"go.uber.org/zap"
)

// Storage defines the interface for file storage operations
type Storage interface {
	Upload(ctx context.Context, filename string, contentType string, data io.Reader) (string, int64, error)
	Download(ctx context.Context, storagePath string) (io.ReadCloser, error)
	Delete(ctx context.Context, storagePath string) error
}

// NewStorage creates a new storage instance based on configuration.
// For local mode, files are stored on the local filesystem.
// For cloud/azure mode, files are stored in Azure Blob Storage.
func NewStorage(cfg *config.StorageConfig, logger *zap.Logger) (Storage, error) {
	switch cfg.Mode {
	case "local":
		return NewLocalStorage(cfg.LocalBasePath)
	case "cloud", "azure":
		if cfg.CloudConnectionString == "" {
			return nil, fmt.Errorf("cloud connection string required for azure storage")
		}
		return NewAzureBlobStorage(cfg.CloudConnectionString, cfg.CloudContainer, logger)
	default:
		return nil, fmt.Errorf("unsupported storage mode: %s", cfg.Mode)
	}
}

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base path if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// Upload uploads a file to local storage
func (s *LocalStorage) Upload(ctx context.Context, filename string, contentType string, data io.Reader) (string, int64, error) {
	// Generate unique storage path
	fileID := uuid.New().String()
	ext := filepath.Ext(filename)
	storagePath := filepath.Join(fileID[:2], fileID[2:4], fileID+ext)
	fullPath := filepath.Join(s.basePath, storagePath)

	// Create directory structure
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	size, err := io.Copy(file, data)
	if err != nil {
		os.Remove(fullPath) // Cleanup on error
		return "", 0, fmt.Errorf("failed to write file: %w", err)
	}

	return storagePath, size, nil
}

// Download downloads a file from local storage
func (s *LocalStorage) Download(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, storagePath)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", storagePath)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete deletes a file from local storage
func (s *LocalStorage) Delete(ctx context.Context, storagePath string) error {
	fullPath := filepath.Join(s.basePath, storagePath)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
