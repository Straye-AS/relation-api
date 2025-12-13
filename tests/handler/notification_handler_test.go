package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/http/handler"
	"github.com/straye-as/relation-api/internal/repository"
	"github.com/straye-as/relation-api/internal/service"
	"github.com/straye-as/relation-api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func setupNotificationHandlerTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupCleanTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createNotificationHandler(t *testing.T, db *gorm.DB) *handler.NotificationHandler {
	logger := zap.NewNop()
	notificationRepo := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo, logger)

	return handler.NewNotificationHandler(notificationService, logger)
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

func createTestNotification(t *testing.T, db *gorm.DB, userID uuid.UUID, notificationType string, title string, read bool) *domain.Notification {
	notification := &domain.Notification{
		UserID:     userID,
		Type:       notificationType,
		Title:      title,
		Message:    "Test notification message",
		Read:       read,
		EntityType: "project",
	}
	err := db.Create(notification).Error
	require.NoError(t, err)
	return notification
}

// TestNotificationHandler_List tests the List endpoint with various filters
func TestNotificationHandler_List(t *testing.T) {
	db := setupNotificationHandlerTestDB(t)
	h := createNotificationHandler(t, db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	// Create test notifications
	createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "Task Assigned 1", false)
	createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "Task Assigned 2", false)
	createTestNotification(t, db, userID, string(domain.NotificationTypeBudgetAlert), "Budget Alert 1", true)
	createTestNotification(t, db, userID, string(domain.NotificationTypeProjectUpdate), "Project Update 1", false)
	createTestNotification(t, db, userID, string(domain.NotificationTypeOfferAccepted), "Offer Accepted 1", true)

	t.Run("list all notifications", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 20, result.PageSize)
	})

	t.Run("list with pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications?page=1&pageSize=2", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Equal(t, 2, result.PageSize)
		assert.Equal(t, 3, result.TotalPages)

		// Check that data contains 2 items
		dataBytes, _ := json.Marshal(result.Data)
		var notifications []domain.NotificationDTO
		err = json.Unmarshal(dataBytes, &notifications)
		assert.NoError(t, err)
		assert.Len(t, notifications, 2)
	})

	t.Run("list unread only", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications?unreadOnly=true", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), result.Total) // 3 unread notifications
	})

	t.Run("list by type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications?type=task_assigned", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Total) // 2 task_assigned notifications
	})

	t.Run("list by type filter combined with unreadOnly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications?type=task_assigned&unreadOnly=true", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result.Total) // Both task_assigned notifications are unread
	})

	t.Run("empty list when no notifications match filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications?type=deal_stage_changed", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)

		// Data should be an empty array
		dataBytes, _ := json.Marshal(result.Data)
		var notifications []domain.NotificationDTO
		err = json.Unmarshal(dataBytes, &notifications)
		assert.NoError(t, err)
		assert.Len(t, notifications, 0)
	})

	t.Run("empty list for user with no notifications", func(t *testing.T) {
		otherUserID := uuid.New()
		otherCtx := createNotificationTestContext(otherUserID)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		req = req.WithContext(otherCtx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.PaginatedResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result.Total)
	})

	t.Run("400 for invalid type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications?type=invalid_type", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.List(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Bad Request", result.Error)
		assert.Contains(t, result.Message, "invalid notification type")
		assert.Contains(t, result.Message, "task_assigned")
		assert.Contains(t, result.Message, "budget_alert")
	})
}

// TestNotificationHandler_GetUnreadCount tests the GetUnreadCount endpoint
func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	db := setupNotificationHandlerTestDB(t)
	h := createNotificationHandler(t, db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	// Create test notifications
	createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "Unread 1", false)
	createTestNotification(t, db, userID, string(domain.NotificationTypeBudgetAlert), "Unread 2", false)
	createTestNotification(t, db, userID, string(domain.NotificationTypeProjectUpdate), "Read 1", true)

	t.Run("returns correct count", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications/count", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.GetUnreadCount(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.UnreadCountDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 2, result.Count) // 2 unread notifications
	})

	t.Run("returns 0 when all read", func(t *testing.T) {
		otherUserID := uuid.New()
		otherCtx := createNotificationTestContext(otherUserID)

		// Create only read notifications for this user
		createTestNotification(t, db, otherUserID, string(domain.NotificationTypeTaskAssigned), "Read 1", true)
		createTestNotification(t, db, otherUserID, string(domain.NotificationTypeBudgetAlert), "Read 2", true)

		req := httptest.NewRequest(http.MethodGet, "/notifications/count", nil)
		req = req.WithContext(otherCtx)

		rr := httptest.NewRecorder()
		h.GetUnreadCount(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.UnreadCountDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 0, result.Count)
	})
}

// TestNotificationHandler_GetByID tests the GetByID endpoint
func TestNotificationHandler_GetByID(t *testing.T) {
	db := setupNotificationHandlerTestDB(t)
	h := createNotificationHandler(t, db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	notification := createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "Test Notification", false)

	t.Run("get own notification successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications/"+notification.ID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", notification.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result domain.NotificationDTO
		err := json.Unmarshal(rr.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, notification.ID, result.ID)
		assert.Equal(t, "Test Notification", result.Title)
		assert.Equal(t, string(domain.NotificationTypeTaskAssigned), result.Type)
	})

	t.Run("404 for non-existent notification", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/notifications/"+nonExistentID.String(), nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("400 for invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notifications/invalid-uuid", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("403 for notification owned by another user", func(t *testing.T) {
		// Create notification for a different user
		otherUserID := uuid.New()
		otherNotification := createTestNotification(t, db, otherUserID, string(domain.NotificationTypeTaskAssigned), "Other User Notification", false)

		req := httptest.NewRequest(http.MethodGet, "/notifications/"+otherNotification.ID.String(), nil)
		req = req.WithContext(ctx) // Using original user's context

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", otherNotification.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.GetByID(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

// TestNotificationHandler_MarkAsRead tests the MarkAsRead endpoint
func TestNotificationHandler_MarkAsRead(t *testing.T) {
	db := setupNotificationHandlerTestDB(t)
	h := createNotificationHandler(t, db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	t.Run("mark notification as read successfully", func(t *testing.T) {
		notification := createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "To Mark Read", false)

		req := httptest.NewRequest(http.MethodPut, "/notifications/"+notification.ID.String()+"/read", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", notification.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.MarkAsRead(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify notification is marked as read in database
		var updated domain.Notification
		err := db.Where("id = ?", notification.ID).First(&updated).Error
		require.NoError(t, err)
		assert.True(t, updated.Read)
		assert.NotNil(t, updated.ReadAt)
	})

	t.Run("404 for non-existent notification", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodPut, "/notifications/"+nonExistentID.String()+"/read", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", nonExistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.MarkAsRead(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("400 for invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/notifications/invalid-uuid/read", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.MarkAsRead(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("success when marking already read notification", func(t *testing.T) {
		notification := createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "Already Read", true)

		req := httptest.NewRequest(http.MethodPut, "/notifications/"+notification.ID.String()+"/read", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", notification.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.MarkAsRead(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code) // Should still succeed
	})

	t.Run("403 for notification owned by another user", func(t *testing.T) {
		otherUserID := uuid.New()
		otherNotification := createTestNotification(t, db, otherUserID, string(domain.NotificationTypeTaskAssigned), "Other User Notification", false)

		req := httptest.NewRequest(http.MethodPut, "/notifications/"+otherNotification.ID.String()+"/read", nil)
		req = req.WithContext(ctx)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", otherNotification.ID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.MarkAsRead(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

// TestNotificationHandler_MarkAllAsRead tests the MarkAllAsRead endpoint
func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	db := setupNotificationHandlerTestDB(t)
	h := createNotificationHandler(t, db)

	userID := uuid.New()
	ctx := createNotificationTestContext(userID)

	t.Run("mark all as read successfully", func(t *testing.T) {
		// Create multiple unread notifications
		n1 := createTestNotification(t, db, userID, string(domain.NotificationTypeTaskAssigned), "Unread 1", false)
		n2 := createTestNotification(t, db, userID, string(domain.NotificationTypeBudgetAlert), "Unread 2", false)
		n3 := createTestNotification(t, db, userID, string(domain.NotificationTypeProjectUpdate), "Already Read", true)

		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		h.MarkAllAsRead(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify all notifications are marked as read
		var notifications []domain.Notification
		err := db.Where("user_id = ?", userID).Find(&notifications).Error
		require.NoError(t, err)

		for _, n := range notifications {
			assert.True(t, n.Read, "notification %s should be read", n.ID)
		}

		// Verify specific notifications
		var updated1, updated2, updated3 domain.Notification
		db.First(&updated1, "id = ?", n1.ID)
		db.First(&updated2, "id = ?", n2.ID)
		db.First(&updated3, "id = ?", n3.ID)

		assert.True(t, updated1.Read)
		assert.True(t, updated2.Read)
		assert.True(t, updated3.Read)
		assert.NotNil(t, updated1.ReadAt)
		assert.NotNil(t, updated2.ReadAt)
	})

	t.Run("success when user has no unread notifications", func(t *testing.T) {
		otherUserID := uuid.New()
		otherCtx := createNotificationTestContext(otherUserID)

		// Create only read notifications
		createTestNotification(t, db, otherUserID, string(domain.NotificationTypeTaskAssigned), "Already Read", true)

		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		req = req.WithContext(otherCtx)

		rr := httptest.NewRecorder()
		h.MarkAllAsRead(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	t.Run("success when user has no notifications", func(t *testing.T) {
		emptyUserID := uuid.New()
		emptyCtx := createNotificationTestContext(emptyUserID)

		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		req = req.WithContext(emptyCtx)

		rr := httptest.NewRecorder()
		h.MarkAllAsRead(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})
}
