package service

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/storage"
	"go.uber.org/zap"
)

type FileService struct {
	fileRepo     *repository.FileRepository
	offerRepo    *repository.OfferRepository
	activityRepo *repository.ActivityRepository
	storage      storage.Storage
	logger       *zap.Logger
}

func NewFileService(
	fileRepo *repository.FileRepository,
	offerRepo *repository.OfferRepository,
	activityRepo *repository.ActivityRepository,
	storage storage.Storage,
	logger *zap.Logger,
) *FileService {
	return &FileService{
		fileRepo:     fileRepo,
		offerRepo:    offerRepo,
		activityRepo: activityRepo,
		storage:      storage,
		logger:       logger,
	}
}

func (s *FileService) Upload(ctx context.Context, filename, contentType string, data io.Reader, offerID *uuid.UUID) (*domain.FileDTO, error) {
	// Verify offer exists if provided
	if offerID != nil {
		if _, err := s.offerRepo.GetByID(ctx, *offerID); err != nil {
			return nil, fmt.Errorf("offer not found: %w", err)
		}
	}

	// Upload to storage
	storagePath, size, err := s.storage.Upload(ctx, filename, contentType, data)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Create file record
	file := &domain.File{
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		StoragePath: storagePath,
		OfferID:     offerID,
	}

	if err := s.fileRepo.Create(ctx, file); err != nil {
		// Try to delete from storage (best effort cleanup)
		_ = s.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Create activity
	if userCtx, ok := auth.FromContext(ctx); ok {
		activity := &domain.Activity{
			TargetType:  domain.ActivityTargetFile,
			TargetID:    file.ID,
			TargetName:  filename,
			Title:       "Fil lastet opp",
			Body:        fmt.Sprintf("Filen '%s' ble lastet opp", filename),
			CreatorName: userCtx.DisplayName,
		}
		_ = s.activityRepo.Create(ctx, activity)
	}

	dto := mapper.ToFileDTO(file)
	return &dto, nil
}

func (s *FileService) GetByID(ctx context.Context, id uuid.UUID) (*domain.FileDTO, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	dto := mapper.ToFileDTO(file)
	return &dto, nil
}

func (s *FileService) Download(ctx context.Context, id uuid.UUID) (io.ReadCloser, string, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get file: %w", err)
	}

	reader, err := s.storage.Download(ctx, file.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}

	return reader, file.Filename, nil
}

func (s *FileService) Delete(ctx context.Context, id uuid.UUID) error {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}

	// Delete from storage
	if err := s.storage.Delete(ctx, file.StoragePath); err != nil {
		s.logger.Warn("failed to delete file from storage", zap.Error(err))
	}

	// Delete from database
	if err := s.fileRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}
