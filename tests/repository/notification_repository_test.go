package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupNotificationTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
}

func createTestNotification(t *testing.T, db *gorm.DB, userID uuid.UUID, read bool) *domain.Notification {
	notification := &domain.Notification{
		UserID:     userID,
		Type:       string(domain.NotificationTypeTaskAssigned),
		Title:      "Test Notification",
		Message:    "This is a test notification message",
		Read:       read,
		EntityType: "Deal",
	}
	err := db.Create(notification).Error
	require.NoError(t, err)
	return notification
}

func TestNotificationRepository_Create(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	entityID := uuid.New()
	notification := &domain.Notification{
		UserID:     userID,
		Type:       string(domain.NotificationTypeDealStageChanged),
		Title:      "Deal Stage Changed",
		Message:    "Your deal has moved to the next stage",
		Read:       false,
		EntityType: "Deal",
		EntityID:   &entityID,
	}

	err := repo.Create(context.Background(), notification)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, notification.ID)
	assert.False(t, notification.Read)
}

func TestNotificationRepository_GetByID(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	created := createTestNotification(t, db, userID, false)

	t.Run("get existing notification", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), created.ID)
		assert.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Title, found.Title)
		assert.Equal(t, created.Message, found.Message)
		assert.Equal(t, created.UserID, found.UserID)
		assert.False(t, found.Read)
	})

	t.Run("get non-existent notification", func(t *testing.T) {
		_, err := repo.GetByID(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestNotificationRepository_ListByUser(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	otherUserID := uuid.New()

	// Create notifications for the test user
	for i := 0; i < 5; i++ {
		read := i%2 == 0 // Alternate between read and unread
		createTestNotification(t, db, userID, read)
		time.Sleep(time.Millisecond) // Ensure different CreatedAt timestamps
	}

	// Create notification for another user
	createTestNotification(t, db, otherUserID, false)

	t.Run("list all notifications for user", func(t *testing.T) {
		notifications, total, err := repo.ListByUser(context.Background(), userID, 1, 10, false, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, notifications, 5)
	})

	t.Run("list only unread notifications", func(t *testing.T) {
		notifications, total, err := repo.ListByUser(context.Background(), userID, 1, 10, true, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total) // Only 2 unread (indexes 1 and 3)
		assert.Len(t, notifications, 2)
		for _, n := range notifications {
			assert.False(t, n.Read)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		notifications, total, err := repo.ListByUser(context.Background(), userID, 1, 2, false, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, notifications, 2)

		notifications, total, err = repo.ListByUser(context.Background(), userID, 2, 2, false, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, notifications, 2)

		notifications, total, err = repo.ListByUser(context.Background(), userID, 3, 2, false, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Len(t, notifications, 1)
	})

	t.Run("returns empty for user with no notifications", func(t *testing.T) {
		emptyUserID := uuid.New()
		notifications, total, err := repo.ListByUser(context.Background(), emptyUserID, 1, 10, false, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, notifications, 0)
	})

	t.Run("ordered by created_at DESC", func(t *testing.T) {
		notifications, _, err := repo.ListByUser(context.Background(), userID, 1, 10, false, "")
		assert.NoError(t, err)
		for i := 0; i < len(notifications)-1; i++ {
			assert.True(t, notifications[i].CreatedAt.After(notifications[i+1].CreatedAt) ||
				notifications[i].CreatedAt.Equal(notifications[i+1].CreatedAt))
		}
	})

	t.Run("filter by notification type", func(t *testing.T) {
		notifications, total, err := repo.ListByUser(context.Background(), userID, 1, 10, false, string(domain.NotificationTypeTaskAssigned))
		assert.NoError(t, err)
		assert.Equal(t, int64(5), total) // All notifications are task_assigned
		assert.Len(t, notifications, 5)
		for _, n := range notifications {
			assert.Equal(t, string(domain.NotificationTypeTaskAssigned), n.Type)
		}
	})

	t.Run("filter by type returns empty when no match", func(t *testing.T) {
		notifications, total, err := repo.ListByUser(context.Background(), userID, 1, 10, false, string(domain.NotificationTypeBudgetAlert))
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, notifications, 0)
	})
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	notification := createTestNotification(t, db, userID, false)

	assert.False(t, notification.Read)

	err := repo.MarkAsRead(context.Background(), notification.ID)
	assert.NoError(t, err)

	// Verify it was marked as read
	found, err := repo.GetByID(context.Background(), notification.ID)
	assert.NoError(t, err)
	assert.True(t, found.Read)
	assert.NotNil(t, found.ReadAt)
}

func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	otherUserID := uuid.New()

	// Create unread notifications for test user
	n1 := createTestNotification(t, db, userID, false)
	n2 := createTestNotification(t, db, userID, false)
	n3 := createTestNotification(t, db, userID, true) // Already read

	// Create unread notification for other user
	otherNotification := createTestNotification(t, db, otherUserID, false)

	err := repo.MarkAllAsRead(context.Background(), userID)
	assert.NoError(t, err)

	// Verify all user notifications are read
	found1, _ := repo.GetByID(context.Background(), n1.ID)
	assert.True(t, found1.Read)
	assert.NotNil(t, found1.ReadAt)

	found2, _ := repo.GetByID(context.Background(), n2.ID)
	assert.True(t, found2.Read)
	assert.NotNil(t, found2.ReadAt)

	found3, _ := repo.GetByID(context.Background(), n3.ID)
	assert.True(t, found3.Read) // Was already read

	// Other user's notification should remain unread
	foundOther, _ := repo.GetByID(context.Background(), otherNotification.ID)
	assert.False(t, foundOther.Read)
}

func TestNotificationRepository_CountUnread(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	otherUserID := uuid.New()

	// Create notifications
	createTestNotification(t, db, userID, false) // unread
	createTestNotification(t, db, userID, false) // unread
	createTestNotification(t, db, userID, true)  // read
	createTestNotification(t, db, userID, false) // unread
	createTestNotification(t, db, otherUserID, false)

	count, err := repo.CountUnread(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestNotificationRepository_CountUnread_NoNotifications(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()

	count, err := repo.CountUnread(context.Background(), userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestNotificationRepository_Create_WithAllFields(t *testing.T) {
	db := setupNotificationTestDB(t)
	repo := repository.NewNotificationRepository(db)

	userID := uuid.New()
	entityID := uuid.New()

	tests := []struct {
		name         string
		notification *domain.Notification
	}{
		{
			name: "task assigned notification",
			notification: &domain.Notification{
				UserID:     userID,
				Type:       string(domain.NotificationTypeTaskAssigned),
				Title:      "Task Assigned",
				Message:    "You have been assigned a new task",
				EntityType: "Activity",
				EntityID:   &entityID,
			},
		},
		{
			name: "budget alert notification",
			notification: &domain.Notification{
				UserID:     userID,
				Type:       string(domain.NotificationTypeBudgetAlert),
				Title:      "Budget Alert",
				Message:    "Project budget is 90% utilized",
				EntityType: "Project",
				EntityID:   &entityID,
			},
		},
		{
			name: "deal stage changed notification",
			notification: &domain.Notification{
				UserID:     userID,
				Type:       string(domain.NotificationTypeDealStageChanged),
				Title:      "Deal Stage Changed",
				Message:    "Deal moved from Lead to Qualified",
				EntityType: "Deal",
				EntityID:   &entityID,
			},
		},
		{
			name: "offer accepted notification",
			notification: &domain.Notification{
				UserID:     userID,
				Type:       string(domain.NotificationTypeOfferAccepted),
				Title:      "Offer Accepted",
				Message:    "Your offer has been accepted",
				EntityType: "Offer",
				EntityID:   &entityID,
			},
		},
		{
			name: "offer rejected notification",
			notification: &domain.Notification{
				UserID:     userID,
				Type:       string(domain.NotificationTypeOfferRejected),
				Title:      "Offer Rejected",
				Message:    "Your offer has been rejected",
				EntityType: "Offer",
				EntityID:   &entityID,
			},
		},
		{
			name: "notification without entity",
			notification: &domain.Notification{
				UserID:  userID,
				Type:    string(domain.NotificationTypeActivityReminder),
				Title:   "Reminder",
				Message: "You have a meeting in 15 minutes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(context.Background(), tt.notification)
			assert.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, tt.notification.ID)

			found, err := repo.GetByID(context.Background(), tt.notification.ID)
			assert.NoError(t, err)
			assert.Equal(t, tt.notification.Type, found.Type)
			assert.Equal(t, tt.notification.Title, found.Title)
			assert.Equal(t, tt.notification.Message, found.Message)
			assert.Equal(t, tt.notification.EntityType, found.EntityType)
			if tt.notification.EntityID != nil {
				assert.Equal(t, *tt.notification.EntityID, *found.EntityID)
			}
		})
	}
}
