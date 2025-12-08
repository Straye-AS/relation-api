package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrNotificationNotFound is returned when a notification is not found
var ErrNotificationNotFound = errors.New("notification not found")

// ErrNotificationNotOwned is returned when trying to access a notification owned by another user
var ErrNotificationNotOwned = errors.New("notification does not belong to current user")

// ErrUserContextRequired is returned when user context is not available
var ErrUserContextRequired = errors.New("user context required")

// NotificationService handles business logic for notifications
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	logger           *zap.Logger
}

// NewNotificationService creates a new NotificationService instance
func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

// CreateForUser creates a notification for a specific user
func (s *NotificationService) CreateForUser(
	ctx context.Context,
	userID uuid.UUID,
	notificationType domain.NotificationType,
	title string,
	message string,
	entityType string,
	entityID *uuid.UUID,
) (*domain.NotificationDTO, error) {
	notification := &domain.Notification{
		UserID:     userID,
		Type:       string(notificationType),
		Title:      title,
		Message:    message,
		EntityType: entityType,
		EntityID:   entityID,
		Read:       false,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("notification created",
		zap.String("notificationID", notification.ID.String()),
		zap.String("userID", userID.String()),
		zap.String("type", string(notificationType)),
	)

	dto := mapper.ToNotificationDTO(notification)
	return &dto, nil
}

// CreateBatch creates notifications for multiple users
func (s *NotificationService) CreateBatch(
	ctx context.Context,
	userIDs []uuid.UUID,
	notificationType domain.NotificationType,
	title string,
	message string,
	entityType string,
	entityID *uuid.UUID,
) ([]domain.NotificationDTO, error) {
	if len(userIDs) == 0 {
		return []domain.NotificationDTO{}, nil
	}

	results := make([]domain.NotificationDTO, 0, len(userIDs))
	var failedCount int

	for _, userID := range userIDs {
		notification := &domain.Notification{
			UserID:     userID,
			Type:       string(notificationType),
			Title:      title,
			Message:    message,
			EntityType: entityType,
			EntityID:   entityID,
			Read:       false,
		}

		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			s.logger.Warn("failed to create notification for user",
				zap.String("userID", userID.String()),
				zap.Error(err),
			)
			failedCount++
			continue
		}

		results = append(results, mapper.ToNotificationDTO(notification))
	}

	if failedCount > 0 {
		s.logger.Warn("batch notification creation completed with failures",
			zap.Int("total", len(userIDs)),
			zap.Int("failed", failedCount),
			zap.Int("succeeded", len(results)),
		)
	} else {
		s.logger.Info("batch notification creation completed",
			zap.Int("count", len(results)),
			zap.String("type", string(notificationType)),
		)
	}

	return results, nil
}

// GetForCurrentUser returns notifications for the current user with pagination
func (s *NotificationService) GetForCurrentUser(
	ctx context.Context,
	page int,
	pageSize int,
	unreadOnly bool,
	notificationType string,
) (*domain.PaginatedResponse, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	// Clamp page size
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	if page < 1 {
		page = 1
	}

	notifications, total, err := s.notificationRepo.ListByUser(ctx, userCtx.UserID, page, pageSize, unreadOnly, notificationType)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	dtos := make([]domain.NotificationDTO, len(notifications))
	for i, notification := range notifications {
		dtos[i] = mapper.ToNotificationDTO(&notification)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &domain.PaginatedResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetByID returns a notification by ID, verifying ownership
func (s *NotificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.NotificationDTO, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	notification, err := s.notificationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Verify the notification belongs to the current user
	if notification.UserID != userCtx.UserID {
		return nil, ErrNotificationNotOwned
	}

	dto := mapper.ToNotificationDTO(notification)
	return &dto, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return ErrUserContextRequired
	}

	// Verify the notification exists and belongs to the current user
	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotificationNotFound
		}
		return fmt.Errorf("failed to get notification: %w", err)
	}

	if notification.UserID != userCtx.UserID {
		return ErrNotificationNotOwned
	}

	// Already read, nothing to do
	if notification.Read {
		return nil
	}

	if err := s.notificationRepo.MarkAsRead(ctx, notificationID); err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	s.logger.Debug("notification marked as read",
		zap.String("notificationID", notificationID.String()),
		zap.String("userID", userCtx.UserID.String()),
	)

	return nil
}

// MarkAllAsReadForUser marks all notifications for the current user as read
func (s *NotificationService) MarkAllAsReadForUser(ctx context.Context) error {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return ErrUserContextRequired
	}

	if err := s.notificationRepo.MarkAllAsRead(ctx, userCtx.UserID); err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	s.logger.Info("all notifications marked as read",
		zap.String("userID", userCtx.UserID.String()),
	)

	return nil
}

// GetUnreadCount returns the count of unread notifications for the current user
func (s *NotificationService) GetUnreadCount(ctx context.Context) (*domain.UnreadCountDTO, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	count, err := s.notificationRepo.CountUnread(ctx, userCtx.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return &domain.UnreadCountDTO{Count: count}, nil
}
