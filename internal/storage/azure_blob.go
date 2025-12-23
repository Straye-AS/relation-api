package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AzureBlobStorage implements Storage interface for Azure Blob Storage
type AzureBlobStorage struct {
	client        *azblob.Client
	containerName string
	logger        *zap.Logger
}

// NewAzureBlobStorage creates a new Azure Blob Storage instance
func NewAzureBlobStorage(connectionString, containerName string, logger *zap.Logger) (*AzureBlobStorage, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client: %w", err)
	}

	// Ensure container exists
	_, err = client.CreateContainer(context.Background(), containerName, nil)
	if err != nil && !strings.Contains(err.Error(), "ContainerAlreadyExists") {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	logger.Info("Azure Blob Storage initialized",
		zap.String("container", containerName),
	)

	return &AzureBlobStorage{
		client:        client,
		containerName: containerName,
		logger:        logger,
	}, nil
}

// Upload uploads a file to Azure Blob Storage
func (s *AzureBlobStorage) Upload(ctx context.Context, filename string, contentType string, data io.Reader) (string, int64, error) {
	// Generate unique blob name with UUID and original extension
	fileID := uuid.New().String()
	ext := filepath.Ext(filename)
	blobName := fileID + ext

	// Upload the blob with content type
	uploadOptions := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	}

	resp, err := s.client.UploadStream(ctx, s.containerName, blobName, data, uploadOptions)
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload blob: %w", err)
	}

	// Get the content length from response headers if available
	// Note: For stream uploads, we may not have the exact size in the response
	// The caller typically knows the size from the original file
	var size int64
	if resp.ContentMD5 != nil {
		// Size is typically tracked by the caller for stream uploads
		// We return 0 here as the actual size should be tracked by the File record
		size = 0
	}

	s.logger.Info("File uploaded to Azure Blob Storage",
		zap.String("blobName", blobName),
		zap.String("container", s.containerName),
		zap.String("contentType", contentType),
		zap.String("originalFilename", filename),
	)

	return blobName, size, nil
}

// Download downloads a file from Azure Blob Storage
func (s *AzureBlobStorage) Download(ctx context.Context, storagePath string) (io.ReadCloser, error) {
	resp, err := s.client.DownloadStream(ctx, s.containerName, storagePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %w", err)
	}

	return resp.Body, nil
}

// Delete deletes a file from Azure Blob Storage
func (s *AzureBlobStorage) Delete(ctx context.Context, storagePath string) error {
	_, err := s.client.DeleteBlob(ctx, s.containerName, storagePath, nil)
	if err != nil {
		// Check if blob doesn't exist (already deleted)
		if strings.Contains(err.Error(), "BlobNotFound") {
			s.logger.Debug("Blob already deleted or not found",
				zap.String("blobName", storagePath),
				zap.String("container", s.containerName),
			)
			return nil
		}
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	s.logger.Info("File deleted from Azure Blob Storage",
		zap.String("blobName", storagePath),
		zap.String("container", s.containerName),
	)

	return nil
}
