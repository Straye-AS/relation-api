package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func setupNotificationServiceTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createNotificationService(db *gorm.DB) *service.NotificationService {
	logger := zap.NewNop()
	notificationRepo := repository.NewNotificationRepository(db)
	return service.NewNotificationService(notificationRepo, logger)
}

func createNotificationTestContext(userID uuid.UUID) context.Context {
	userCtx := &auth.UserContext{
		UserID:      userID,
		DisplayName: "Test User",
		Email:       "test@example.com",
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func TestNotificationService_CreateForUser(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	t.Run("create notification successfully", func(t *testing.T) {
		userID := uuid.New()
		entityID := uuid.New()

		dto, err := svc.CreateForUser(
			context.Background(),
			userID,
			domain.NotificationTypeTaskAssigned,
			"Task Assigned",
			"You have a new task",
			"Activity",
			&entityID,
		)

		assert.NoError(t, err)
		assert.NotNil(t, dto)
		assert.NotEqual(t, uuid.Nil, dto.ID)
		assert.Equal(t, "Task Assigned", dto.Title)
		assert.Equal(t, "You have a new task", dto.Message)
		assert.Equal(t, string(domain.NotificationTypeTaskAssigned), dto.Type)
		assert.Equal(t, "Activity", dto.EntityType)
		assert.Equal(t, &entityID, dto.EntityID)
		assert.False(t, dto.Read)
	})

	t.Run("create notification without entity", func(t *testing.T) {
		userID := uuid.New()

		dto, err := svc.CreateForUser(
			context.Background(),
			userID,
			domain.NotificationTypeActivityReminder,
			"Reminder",
			"Meeting in 15 minutes",
			"",
			nil,
		)

		assert.NoError(t, err)
		assert.NotNil(t, dto)
		assert.Equal(t, "Reminder", dto.Title)
		assert.Empty(t, dto.EntityType)
		assert.Nil(t, dto.EntityID)
	})

	t.Run("create notifications with different types", func(t *testing.T) {
		types := []domain.NotificationType{
			domain.NotificationTypeTaskAssigned,
			domain.NotificationTypeBudgetAlert,
			domain.NotificationTypeDealStageChanged,
			domain.NotificationTypeOfferAccepted,
			domain.NotificationTypeOfferRejected,
			domain.NotificationTypeActivityReminder,
			domain.NotificationTypeProjectUpdate,
		}

		for _, notificationType := range types {
			userID := uuid.New()
			dto, err := svc.CreateForUser(
				context.Background(),
				userID,
				notificationType,
				"Test Title",
				"Test Message",
				"",
				nil,
			)
			assert.NoError(t, err)
			assert.Equal(t, string(notificationType), dto.Type)
		}
	})
}

func TestNotificationService_CreateBatch(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	t.Run("create batch notifications", func(t *testing.T) {
		userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
		entityID := uuid.New()

		dtos, err := svc.CreateBatch(
			context.Background(),
			userIDs,
			domain.NotificationTypeDealStageChanged,
			"Deal Updated",
			"The deal has moved to a new stage",
			"Deal",
			&entityID,
		)

		assert.NoError(t, err)
		assert.Len(t, dtos, 3)
		for _, dto := range dtos {
			assert.Equal(t, "Deal Updated", dto.Title)
			assert.Equal(t, string(domain.NotificationTypeDealStageChanged), dto.Type)
		}
	})

	t.Run("create batch with empty user list", func(t *testing.T) {
		dtos, err := svc.CreateBatch(
			context.Background(),
			[]uuid.UUID{},
			domain.NotificationTypeTaskAssigned,
			"Title",
			"Message",
			"",
			nil,
		)

		assert.NoError(t, err)
		assert.Empty(t, dtos)
	})

	t.Run("create batch for single user", func(t *testing.T) {
		userIDs := []uuid.UUID{uuid.New()}
		entityID := uuid.New()

		dtos, err := svc.CreateBatch(
			context.Background(),
			userIDs,
			domain.NotificationTypeProjectUpdate,
			"Project Update",
			"Project status has changed",
			"Project",
			&entityID,
		)

		assert.NoError(t, err)
		assert.Len(t, dtos, 1)
		assert.Equal(t, "Project Update", dtos[0].Title)
	})
}

func TestNotificationService_GetForCurrentUser(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	// Create test notifications
	for i := 0; i < 5; i++ {
		read := i%2 == 0
		notification := &domain.Notification{
			UserID:  userID,
			Type:    string(domain.NotificationTypeTaskAssigned),
			Title:   "Test Notification",
			Message: "Test message",
			Read:    read,
		}
		err := db.Create(notification).Error
		require.NoError(t, err)
	}

	t.Run("get all notifications", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 10, false, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Data, 5)
	})

	t.Run("get only unread notifications", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 10, true, "")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(2), result.Total) // Only 2 unread (indexes 1 and 3)
	})

	t.Run("pagination", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 2, false, "")
		assert.NoError(t, err)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 3, result.TotalPages)

		result, err = svc.GetForCurrentUser(ctx, 2, 2, false, "")
		assert.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})

	t.Run("clamp page size to minimum", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 0, false, "")
		assert.NoError(t, err)
		assert.Equal(t, 20, result.PageSize)
	})

	t.Run("clamp page size to maximum", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 500, false, "")
		assert.NoError(t, err)
		assert.Equal(t, 200, result.PageSize)
	})

	t.Run("clamp page to minimum", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 0, 10, false, "")
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})

	t.Run("error without user context", func(t *testing.T) {
		_, err := svc.GetForCurrentUser(context.Background(), 1, 10, false, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrUserContextRequired)
	})

	t.Run("filter by notification type", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 10, false, string(domain.NotificationTypeTaskAssigned))
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total) // All notifications are task_assigned
	})

	t.Run("filter by type returns empty when no match", func(t *testing.T) {
		result, err := svc.GetForCurrentUser(ctx, 1, 10, false, string(domain.NotificationTypeBudgetAlert))
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestNotificationService_GetByID(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	userID := uuid.New()
	otherUserID := uuid.New()
	ctx := createNotificationTestContext(userID)

	// Create notification for the user
	entityID := uuid.New()
	notification := &domain.Notification{
		UserID:     userID,
		Type:       string(domain.NotificationTypeDealStageChanged),
		Title:      "Test Notification",
		Message:    "Test message",
		EntityType: "Deal",
		EntityID:   &entityID,
	}
	err := db.Create(notification).Error
	require.NoError(t, err)

	// Create notification for another user
	otherNotification := &domain.Notification{
		UserID:  otherUserID,
		Type:    string(domain.NotificationTypeTaskAssigned),
		Title:   "Other Notification",
		Message: "Other message",
	}
	err = db.Create(otherNotification).Error
	require.NoError(t, err)

	t.Run("get own notification", func(t *testing.T) {
		dto, err := svc.GetByID(ctx, notification.ID)
		assert.NoError(t, err)
		assert.NotNil(t, dto)
		assert.Equal(t, notification.ID, dto.ID)
		assert.Equal(t, "Test Notification", dto.Title)
		assert.Equal(t, "Deal", dto.EntityType)
	})

	t.Run("cannot get notification owned by another user", func(t *testing.T) {
		dto, err := svc.GetByID(ctx, otherNotification.ID)
		assert.Error(t, err)
		assert.Nil(t, dto)
		assert.ErrorIs(t, err, service.ErrNotificationNotOwned)
	})

	t.Run("get non-existent notification", func(t *testing.T) {
		dto, err := svc.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, dto)
		assert.ErrorIs(t, err, service.ErrNotificationNotFound)
	})

	t.Run("error without user context", func(t *testing.T) {
		dto, err := svc.GetByID(context.Background(), notification.ID)
		assert.Error(t, err)
		assert.Nil(t, dto)
		assert.ErrorIs(t, err, service.ErrUserContextRequired)
	})
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	userID := uuid.New()
	otherUserID := uuid.New()
	ctx := createNotificationTestContext(userID)

	// Create unread notification for the user
	notification := &domain.Notification{
		UserID:  userID,
		Type:    string(domain.NotificationTypeTaskAssigned),
		Title:   "Test Notification",
		Message: "Test message",
		Read:    false,
	}
	err := db.Create(notification).Error
	require.NoError(t, err)

	// Create notification for another user
	otherNotification := &domain.Notification{
		UserID:  otherUserID,
		Type:    string(domain.NotificationTypeTaskAssigned),
		Title:   "Other Notification",
		Message: "Other message",
		Read:    false,
	}
	err = db.Create(otherNotification).Error
	require.NoError(t, err)

	t.Run("mark own notification as read", func(t *testing.T) {
		err := svc.MarkAsRead(ctx, notification.ID)
		assert.NoError(t, err)

		// Verify it was marked as read
		dto, err := svc.GetByID(ctx, notification.ID)
		assert.NoError(t, err)
		assert.True(t, dto.Read)
	})

	t.Run("mark already read notification (idempotent)", func(t *testing.T) {
		// Create a notification and mark it as read
		readNotification := &domain.Notification{
			UserID:  userID,
			Type:    string(domain.NotificationTypeTaskAssigned),
			Title:   "Already Read",
			Message: "Already read message",
			Read:    true,
		}
		err := db.Create(readNotification).Error
		require.NoError(t, err)

		// Marking again should not error
		err = svc.MarkAsRead(ctx, readNotification.ID)
		assert.NoError(t, err)
	})

	t.Run("cannot mark another user's notification as read", func(t *testing.T) {
		err := svc.MarkAsRead(ctx, otherNotification.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrNotificationNotOwned)
	})

	t.Run("mark non-existent notification as read", func(t *testing.T) {
		err := svc.MarkAsRead(ctx, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrNotificationNotFound)
	})

	t.Run("error without user context", func(t *testing.T) {
		err := svc.MarkAsRead(context.Background(), notification.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrUserContextRequired)
	})
}

func TestNotificationService_MarkAllAsReadForUser(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	userID := uuid.New()
	otherUserID := uuid.New()
	ctx := createNotificationTestContext(userID)

	// Create unread notifications for the user
	for i := 0; i < 3; i++ {
		notification := &domain.Notification{
			UserID:  userID,
			Type:    string(domain.NotificationTypeTaskAssigned),
			Title:   "Unread Notification",
			Message: "Unread message",
			Read:    false,
		}
		err := db.Create(notification).Error
		require.NoError(t, err)
	}

	// Create unread notification for another user
	otherNotification := &domain.Notification{
		UserID:  otherUserID,
		Type:    string(domain.NotificationTypeTaskAssigned),
		Title:   "Other Notification",
		Message: "Other message",
		Read:    false,
	}
	err := db.Create(otherNotification).Error
	require.NoError(t, err)

	t.Run("mark all notifications as read", func(t *testing.T) {
		err := svc.MarkAllAsReadForUser(ctx)
		assert.NoError(t, err)

		// Verify all user's notifications are read
		result, err := svc.GetForCurrentUser(ctx, 1, 10, true, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result.Total) // No unread notifications

		// Verify other user's notification is still unread
		otherCtx := createNotificationTestContext(otherUserID)
		otherResult, err := svc.GetForCurrentUser(otherCtx, 1, 10, true, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), otherResult.Total)
	})

	t.Run("mark all when already read (idempotent)", func(t *testing.T) {
		err := svc.MarkAllAsReadForUser(ctx)
		assert.NoError(t, err) // Should not error even if already all read
	})

	t.Run("error without user context", func(t *testing.T) {
		err := svc.MarkAllAsReadForUser(context.Background())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrUserContextRequired)
	})
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	t.Run("count with no notifications", func(t *testing.T) {
		count, err := svc.GetUnreadCount(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, count)
		assert.Equal(t, 0, count.Count)
	})

	t.Run("count with mixed read/unread", func(t *testing.T) {
		// Create 3 unread and 2 read notifications
		for i := 0; i < 3; i++ {
			notification := &domain.Notification{
				UserID:  userID,
				Type:    string(domain.NotificationTypeTaskAssigned),
				Title:   "Unread",
				Message: "Message",
				Read:    false,
			}
			err := db.Create(notification).Error
			require.NoError(t, err)
		}
		for i := 0; i < 2; i++ {
			notification := &domain.Notification{
				UserID:  userID,
				Type:    string(domain.NotificationTypeTaskAssigned),
				Title:   "Read",
				Message: "Message",
				Read:    true,
			}
			err := db.Create(notification).Error
			require.NoError(t, err)
		}

		count, err := svc.GetUnreadCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 3, count.Count)
	})

	t.Run("count does not include other user's notifications", func(t *testing.T) {
		otherUserID := uuid.New()
		notification := &domain.Notification{
			UserID:  otherUserID,
			Type:    string(domain.NotificationTypeTaskAssigned),
			Title:   "Other User",
			Message: "Message",
			Read:    false,
		}
		err := db.Create(notification).Error
		require.NoError(t, err)

		// Count for original user should not change
		count, err := svc.GetUnreadCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 3, count.Count) // Still 3 from previous test
	})

	t.Run("error without user context", func(t *testing.T) {
		count, err := svc.GetUnreadCount(context.Background())
		assert.Error(t, err)
		assert.Nil(t, count)
		assert.ErrorIs(t, err, service.ErrUserContextRequired)
	})
}

func TestNotificationService_Integration(t *testing.T) {
	db := setupNotificationServiceTestDB(t)
	svc := createNotificationService(db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	t.Run("full lifecycle", func(t *testing.T) {
		// 1. Create notification
		entityID := uuid.New()
		dto, err := svc.CreateForUser(
			context.Background(),
			userID,
			domain.NotificationTypeDealStageChanged,
			"Deal Moved",
			"Your deal has moved to Qualified stage",
			"Deal",
			&entityID,
		)
		require.NoError(t, err)
		assert.NotNil(t, dto)
		notificationID := dto.ID

		// 2. Verify unread count is 1
		count, err := svc.GetUnreadCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count.Count)

		// 3. Get notification by ID
		retrieved, err := svc.GetByID(ctx, notificationID)
		require.NoError(t, err)
		assert.Equal(t, "Deal Moved", retrieved.Title)
		assert.False(t, retrieved.Read)

		// 4. List notifications
		result, err := svc.GetForCurrentUser(ctx, 1, 10, false, "")
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)

		// 5. Mark as read
		err = svc.MarkAsRead(ctx, notificationID)
		require.NoError(t, err)

		// 6. Verify unread count is 0
		count, err = svc.GetUnreadCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count.Count)

		// 7. Verify notification is read
		retrieved, err = svc.GetByID(ctx, notificationID)
		require.NoError(t, err)
		assert.True(t, retrieved.Read)
	})

	t.Run("batch create and mark all read", func(t *testing.T) {
		newUserID := uuid.New()
		newCtx := createNotificationTestContext(newUserID)

		// Create batch of notifications
		_, err := svc.CreateBatch(
			context.Background(),
			[]uuid.UUID{newUserID, newUserID, newUserID},
			domain.NotificationTypeProjectUpdate,
			"Update",
			"Project updated",
			"Project",
			nil,
		)
		require.NoError(t, err)

		// Verify 3 unread
		count, err := svc.GetUnreadCount(newCtx)
		require.NoError(t, err)
		assert.Equal(t, 3, count.Count)

		// Mark all as read
		err = svc.MarkAllAsReadForUser(newCtx)
		require.NoError(t, err)

		// Verify 0 unread
		count, err = svc.GetUnreadCount(newCtx)
		require.NoError(t, err)
		assert.Equal(t, 0, count.Count)
	})
}
