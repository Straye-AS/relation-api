package service_test

import (
	"context"
	"testing"
	"time"

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

func setupActivityServiceTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupCleanTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestData(t, db)
	})
	return db
}

func createActivityServiceTestCustomer(t *testing.T, db *gorm.DB) *domain.Customer {
	return testutil.CreateTestCustomer(t, db, "Activity Service Test Customer")
}

func createActivityService(t *testing.T, db *gorm.DB) *service.ActivityService {
	logger := zap.NewNop()
	activityRepo := repository.NewActivityRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo, logger)

	return service.NewActivityService(activityRepo, notificationService, logger)
}

func createActivityTestContext() context.Context {
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
		CompanyID:   domain.CompanyStalbygg,
		Roles:       []domain.UserRoleType{domain.RoleSuperAdmin},
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func createActivityTestContextWithUser(userID uuid.UUID, displayName string, roles []domain.UserRoleType) context.Context {
	userCtx := &auth.UserContext{
		UserID:      userID,
		DisplayName: displayName,
		Email:       displayName + "@example.com",
		CompanyID:   domain.CompanyStalbygg,
		Roles:       roles,
	}
	return auth.WithUserContext(context.Background(), userCtx)
}

func TestActivityService_Create(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("create activity with minimal fields", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "New Activity",
			ActivityType: domain.ActivityTypeNote,
		}

		activity, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, "New Activity", activity.Title)
		assert.Equal(t, domain.ActivityStatusPlanned, activity.Status) // Default status
		assert.Equal(t, domain.ActivityTypeNote, activity.ActivityType)
	})

	t.Run("create activity with all fields", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		dueDate := time.Now().Add(48 * time.Hour)
		duration := 60

		req := &domain.CreateActivityRequest{
			TargetType:      domain.ActivityTargetCustomer,
			TargetID:        customer.ID,
			Title:           "Full Activity",
			Body:            "Activity with all fields",
			ActivityType:    domain.ActivityTypeMeeting,
			Status:          domain.ActivityStatusInProgress,
			ScheduledAt:     &scheduledAt,
			DueDate:         &dueDate,
			DurationMinutes: &duration,
			Priority:        3,
			IsPrivate:       true,
		}

		activity, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, "Full Activity", activity.Title)
		assert.Equal(t, "Activity with all fields", activity.Body)
		assert.Equal(t, domain.ActivityStatusInProgress, activity.Status)
		assert.Equal(t, 3, activity.Priority)
		assert.True(t, activity.IsPrivate)
	})

	t.Run("create without user context fails", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "No Context",
			ActivityType: domain.ActivityTypeNote,
		}

		activity, err := svc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, activity)
	})
}

func TestActivityService_GetByID(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	// Create an activity first
	req := &domain.CreateActivityRequest{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Test Activity",
		ActivityType: domain.ActivityTypeTask,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("get existing activity", func(t *testing.T) {
		activity, err := svc.GetByID(ctx, created.ID)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, created.ID, activity.ID)
		assert.Equal(t, "Test Activity", activity.Title)
	})

	t.Run("get non-existent activity", func(t *testing.T) {
		activity, err := svc.GetByID(ctx, uuid.New())
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})

	t.Run("get without user context fails", func(t *testing.T) {
		activity, err := svc.GetByID(context.Background(), created.ID)
		assert.Error(t, err)
		assert.Nil(t, activity)
	})
}

func TestActivityService_Update(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create an activity first
	req := &domain.CreateActivityRequest{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Original Title",
		Body:         "Original body",
		ActivityType: domain.ActivityTypeTask,
	}
	created, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("update activity by creator", func(t *testing.T) {
		updateReq := &domain.UpdateActivityRequest{
			Title:    "Updated Title",
			Body:     "Updated body",
			Status:   domain.ActivityStatusInProgress,
			Priority: 5,
		}

		activity, err := svc.Update(ctx, created.ID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, "Updated Title", activity.Title)
		assert.Equal(t, "Updated body", activity.Body)
		assert.Equal(t, domain.ActivityStatusInProgress, activity.Status)
		assert.Equal(t, 5, activity.Priority)
	})

	t.Run("update non-existent activity", func(t *testing.T) {
		updateReq := &domain.UpdateActivityRequest{
			Title: "Non-existent",
		}

		activity, err := svc.Update(ctx, uuid.New(), updateReq)
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})

	t.Run("update by non-owner without permission fails", func(t *testing.T) {
		otherUserCtx := createActivityTestContextWithUser(uuid.New(), "Other User", []domain.UserRoleType{domain.RoleViewer})

		updateReq := &domain.UpdateActivityRequest{
			Title: "Unauthorized Update",
		}

		activity, err := svc.Update(otherUserCtx, created.ID, updateReq)
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityForbidden)
	})

	t.Run("update by assigned user succeeds", func(t *testing.T) {
		// First assign to another user
		assignedUserID := uuid.New()
		updateReq := &domain.UpdateActivityRequest{
			Title:        "Assigned Update",
			AssignedToID: assignedUserID.String(),
		}
		_, err := svc.Update(ctx, created.ID, updateReq)
		require.NoError(t, err)

		// Now the assigned user should be able to update
		assignedCtx := createActivityTestContextWithUser(assignedUserID, "Assigned User", []domain.UserRoleType{domain.RoleViewer})
		updateReq2 := &domain.UpdateActivityRequest{
			Title:        "Updated by Assignee",
			AssignedToID: assignedUserID.String(),
		}
		activity, err := svc.Update(assignedCtx, created.ID, updateReq2)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, "Updated by Assignee", activity.Title)
	})

	t.Run("update by manager succeeds", func(t *testing.T) {
		// Create a new activity
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Manager Test Activity",
			ActivityType: domain.ActivityTypeTask,
		}
		managerTestActivity, err := svc.Create(ctx, req)
		require.NoError(t, err)

		managerCtx := createActivityTestContextWithUser(uuid.New(), "Manager User", []domain.UserRoleType{domain.RoleManager})

		updateReq := &domain.UpdateActivityRequest{
			Title: "Updated by Manager",
		}

		activity, err := svc.Update(managerCtx, managerTestActivity.ID, updateReq)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, "Updated by Manager", activity.Title)
	})

	_ = userCtx // Used to establish owner context
}

func TestActivityService_Delete(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("delete activity by creator", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Activity to Delete",
			ActivityType: domain.ActivityTypeNote,
		}
		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		err = svc.Delete(ctx, created.ID)
		assert.NoError(t, err)

		// Verify it's gone
		_, err = svc.GetByID(ctx, created.ID)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})

	t.Run("delete non-existent activity", func(t *testing.T) {
		err := svc.Delete(ctx, uuid.New())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})

	t.Run("delete by non-owner without permission fails", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Protected Activity",
			ActivityType: domain.ActivityTypeNote,
		}
		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		otherUserCtx := createActivityTestContextWithUser(uuid.New(), "Other User", []domain.UserRoleType{domain.RoleViewer})
		err = svc.Delete(otherUserCtx, created.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrActivityForbidden)
	})
}

func TestActivityService_Complete(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("complete activity", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Task to Complete",
			ActivityType: domain.ActivityTypeTask,
			Status:       domain.ActivityStatusPlanned,
		}
		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		activity, err := svc.Complete(ctx, created.ID, "")
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, domain.ActivityStatusCompleted, activity.Status)
		assert.NotEmpty(t, activity.CompletedAt)
	})

	t.Run("complete activity with outcome", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Meeting to Complete",
			Body:         "Meeting notes",
			ActivityType: domain.ActivityTypeMeeting,
			Status:       domain.ActivityStatusInProgress,
		}
		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		activity, err := svc.Complete(ctx, created.ID, "Meeting went well, follow-up scheduled")
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Equal(t, domain.ActivityStatusCompleted, activity.Status)
		assert.Contains(t, activity.Body, "Outcome")
		assert.Contains(t, activity.Body, "Meeting went well")
	})

	t.Run("complete already completed activity fails", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Already Completed",
			ActivityType: domain.ActivityTypeTask,
			Status:       domain.ActivityStatusPlanned,
		}
		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Complete it first
		_, err = svc.Complete(ctx, created.ID, "")
		require.NoError(t, err)

		// Try to complete again
		activity, err := svc.Complete(ctx, created.ID, "")
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityAlreadyCompleted)
	})

	t.Run("complete cancelled activity fails", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Cancelled Task",
			ActivityType: domain.ActivityTypeTask,
			Status:       domain.ActivityStatusCancelled,
		}
		created, err := svc.Create(ctx, req)
		require.NoError(t, err)

		activity, err := svc.Complete(ctx, created.ID, "")
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityCannotCompleteCancelled)
	})

	t.Run("complete non-existent activity fails", func(t *testing.T) {
		activity, err := svc.Complete(ctx, uuid.New(), "")
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})
}

func TestActivityService_List(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	// Create test activities
	for i := 0; i < 5; i++ {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "List Activity " + string(rune('A'+i)),
			ActivityType: domain.ActivityTypeNote,
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	t.Run("list all activities", func(t *testing.T) {
		result, err := svc.List(ctx, nil, 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Data, 5)
	})

	t.Run("list with pagination", func(t *testing.T) {
		result, err := svc.List(ctx, nil, 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, 3, result.TotalPages)

		result, err = svc.List(ctx, nil, 2, 2)
		assert.NoError(t, err)
		assert.Len(t, result.Data, 2)
	})

	t.Run("list with filters", func(t *testing.T) {
		status := domain.ActivityStatusPlanned
		filters := &domain.ActivityFilters{Status: &status}
		result, err := svc.List(ctx, filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), result.Total)
	})

	t.Run("clamps page size", func(t *testing.T) {
		result, err := svc.List(ctx, nil, 1, 500)
		assert.NoError(t, err)
		assert.Equal(t, 200, result.PageSize)
	})

	t.Run("defaults page to 1", func(t *testing.T) {
		result, err := svc.List(ctx, nil, 0, 10)
		assert.NoError(t, err)
		assert.Equal(t, 1, result.Page)
	})
}

func TestActivityService_GetMyTasks(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create tasks assigned to the current user
	for i := 0; i < 3; i++ {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "My Task " + string(rune('A'+i)),
			ActivityType: domain.ActivityTypeTask,
			Status:       domain.ActivityStatusPlanned,
			AssignedToID: userCtx.UserID.String(),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	// Create a completed task (should not appear)
	req := &domain.CreateActivityRequest{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Completed Task",
		ActivityType: domain.ActivityTypeTask,
		Status:       domain.ActivityStatusCompleted,
		AssignedToID: userCtx.UserID.String(),
	}
	_, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("get my tasks excludes completed", func(t *testing.T) {
		result, err := svc.GetMyTasks(ctx, 1, 10)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(3), result.Total)
		assert.Len(t, result.Data, 3)
	})

	t.Run("without user context fails", func(t *testing.T) {
		result, err := svc.GetMyTasks(context.Background(), 1, 10)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestActivityService_GetUpcoming(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create upcoming scheduled activity
	scheduledAt := time.Now().Add(3 * 24 * time.Hour)
	req := &domain.CreateActivityRequest{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Upcoming Meeting",
		ActivityType: domain.ActivityTypeMeeting,
		Status:       domain.ActivityStatusPlanned,
		ScheduledAt:  &scheduledAt,
		AssignedToID: userCtx.UserID.String(),
	}
	_, err := svc.Create(ctx, req)
	require.NoError(t, err)

	t.Run("get upcoming activities", func(t *testing.T) {
		activities, err := svc.GetUpcoming(ctx, 7, 10)
		require.NoError(t, err)
		require.Len(t, activities, 1, "expected 1 upcoming activity for user %s", userCtx.UserID)
		assert.Equal(t, "Upcoming Meeting", activities[0].Title)
	})

	t.Run("limits days ahead", func(t *testing.T) {
		activities, err := svc.GetUpcoming(ctx, 100, 10)
		assert.NoError(t, err)
		// Should clamp to 90 days
		assert.NotNil(t, activities)
	})

	t.Run("limits result count", func(t *testing.T) {
		activities, err := svc.GetUpcoming(ctx, 7, 200)
		assert.NoError(t, err)
		// Should clamp to 100
		assert.NotNil(t, activities)
	})
}

func TestActivityService_GetStatusCounts(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()
	userCtx, _ := auth.FromContext(ctx)

	// Create activities with different statuses
	statuses := []domain.ActivityStatus{
		domain.ActivityStatusPlanned,
		domain.ActivityStatusPlanned,
		domain.ActivityStatusInProgress,
		domain.ActivityStatusCompleted,
		domain.ActivityStatusCompleted,
		domain.ActivityStatusCompleted,
	}

	for _, status := range statuses {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Status Count Activity",
			ActivityType: domain.ActivityTypeTask,
			Status:       status,
			AssignedToID: userCtx.UserID.String(),
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	t.Run("get status counts", func(t *testing.T) {
		counts, err := svc.GetStatusCounts(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, counts)
		assert.Equal(t, 2, counts.Planned)
		assert.Equal(t, 1, counts.InProgress)
		assert.Equal(t, 3, counts.Completed)
		assert.Equal(t, 0, counts.Cancelled)
	})
}

func TestActivityService_GetByTarget(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	// Create activities for the customer
	for i := 0; i < 3; i++ {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Target Activity " + string(rune('A'+i)),
			ActivityType: domain.ActivityTypeNote,
		}
		_, err := svc.Create(ctx, req)
		require.NoError(t, err)
	}

	t.Run("get activities by target", func(t *testing.T) {
		activities, err := svc.GetByTarget(ctx, domain.ActivityTargetCustomer, customer.ID, 10)
		assert.NoError(t, err)
		assert.NotNil(t, activities)
		assert.Len(t, activities, 3)
	})

	t.Run("respects limit", func(t *testing.T) {
		activities, err := svc.GetByTarget(ctx, domain.ActivityTargetCustomer, customer.ID, 2)
		assert.NoError(t, err)
		assert.Len(t, activities, 2)
	})

	t.Run("clamps limit to max", func(t *testing.T) {
		activities, err := svc.GetByTarget(ctx, domain.ActivityTargetCustomer, customer.ID, 200)
		assert.NoError(t, err)
		// Should clamp to 100 max, but we only have 3
		assert.Len(t, activities, 3)
	})
}

func TestActivityService_PrivateActivityAccess(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)

	// Create a private activity as one user
	creatorID := uuid.New()
	creatorCtx := createActivityTestContextWithUser(creatorID, "Creator User", []domain.UserRoleType{domain.RoleMarket})

	req := &domain.CreateActivityRequest{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Private Activity",
		ActivityType: domain.ActivityTypeNote,
		IsPrivate:    true,
	}
	privateActivity, err := svc.Create(creatorCtx, req)
	require.NoError(t, err)

	t.Run("creator can view private activity", func(t *testing.T) {
		activity, err := svc.GetByID(creatorCtx, privateActivity.ID)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
	})

	t.Run("other user cannot view private activity", func(t *testing.T) {
		otherCtx := createActivityTestContextWithUser(uuid.New(), "Other User", []domain.UserRoleType{domain.RoleViewer})
		activity, err := svc.GetByID(otherCtx, privateActivity.ID)
		assert.Error(t, err)
		assert.Nil(t, activity)
		assert.ErrorIs(t, err, service.ErrActivityForbidden)
	})

	t.Run("manager can view private activity", func(t *testing.T) {
		managerCtx := createActivityTestContextWithUser(uuid.New(), "Manager User", []domain.UserRoleType{domain.RoleManager})
		activity, err := svc.GetByID(managerCtx, privateActivity.ID)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
	})
}

func TestActivityService_AddAttendee(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("add attendee to meeting", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Team Meeting",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
		}
		meeting, err := svc.Create(ctx, req)
		require.NoError(t, err)

		attendeeID := uuid.New().String()
		activity, err := svc.AddAttendee(ctx, meeting.ID, attendeeID)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.Contains(t, activity.Attendees, attendeeID)
	})

	t.Run("cannot add attendee to non-meeting activity", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Task Activity",
			ActivityType: domain.ActivityTypeTask,
		}
		task, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.AddAttendee(ctx, task.ID, uuid.New().String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrActivityNotMeeting)
	})

	t.Run("cannot add same attendee twice", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Another Meeting",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
		}
		meeting, err := svc.Create(ctx, req)
		require.NoError(t, err)

		attendeeID := uuid.New().String()
		_, err = svc.AddAttendee(ctx, meeting.ID, attendeeID)
		require.NoError(t, err)

		_, err = svc.AddAttendee(ctx, meeting.ID, attendeeID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrAttendeeAlreadyAdded)
	})

	t.Run("add attendee to non-existent activity", func(t *testing.T) {
		_, err := svc.AddAttendee(ctx, uuid.New(), uuid.New().String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})
}

func TestActivityService_RemoveAttendee(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("remove attendee from meeting", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		attendeeID := uuid.New().String()
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Meeting with Attendee",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
			Attendees:    []string{attendeeID},
		}
		meeting, err := svc.Create(ctx, req)
		require.NoError(t, err)
		require.Contains(t, meeting.Attendees, attendeeID)

		activity, err := svc.RemoveAttendee(ctx, meeting.ID, attendeeID)
		assert.NoError(t, err)
		assert.NotNil(t, activity)
		assert.NotContains(t, activity.Attendees, attendeeID)
	})

	t.Run("remove non-existent attendee", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Meeting without Attendees",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
		}
		meeting, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.RemoveAttendee(ctx, meeting.ID, uuid.New().String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrAttendeeNotFound)
	})

	t.Run("cannot remove attendee from non-meeting activity", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Note Activity",
			ActivityType: domain.ActivityTypeNote,
		}
		note, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.RemoveAttendee(ctx, note.ID, uuid.New().String())
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrActivityNotMeeting)
	})
}

func TestActivityService_CreateFollowUp(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()
	userCtx, _ := auth.FromContext(ctx)

	t.Run("create follow-up from completed activity", func(t *testing.T) {
		// First create and complete a parent activity
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Parent Task",
			ActivityType: domain.ActivityTypeTask,
			Priority:     3,
		}
		parent, err := svc.Create(ctx, req)
		require.NoError(t, err)

		// Complete the parent
		parent, err = svc.Complete(ctx, parent.ID, "Task completed successfully")
		require.NoError(t, err)
		require.Equal(t, domain.ActivityStatusCompleted, parent.Status)

		// Create follow-up
		followUpReq := &domain.CreateFollowUpRequest{
			Title:       "Follow-up Task",
			Description: "Need to verify the work",
		}
		followUp, err := svc.CreateFollowUp(ctx, parent.ID, followUpReq)
		assert.NoError(t, err)
		assert.NotNil(t, followUp)
		assert.Equal(t, "Follow-up Task", followUp.Title)
		assert.Equal(t, domain.ActivityTypeTask, followUp.ActivityType)
		assert.Equal(t, domain.ActivityStatusPlanned, followUp.Status)
		assert.Equal(t, &parent.ID, followUp.ParentActivityID)
		assert.Equal(t, parent.Priority, followUp.Priority) // Inherited priority
		assert.Contains(t, followUp.Body, "Follow-up from: Parent Task")
	})

	t.Run("follow-up inherits target from parent", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Customer Task",
			ActivityType: domain.ActivityTypeTask,
		}
		parent, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.Complete(ctx, parent.ID, "")
		require.NoError(t, err)

		followUp, err := svc.CreateFollowUp(ctx, parent.ID, &domain.CreateFollowUpRequest{
			Title: "Follow-up for Customer",
		})
		assert.NoError(t, err)
		assert.Equal(t, domain.ActivityTargetCustomer, followUp.TargetType)
		assert.Equal(t, customer.ID, followUp.TargetID)
	})

	t.Run("follow-up with custom assignee", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Team Task",
			ActivityType: domain.ActivityTypeTask,
		}
		parent, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.Complete(ctx, parent.ID, "")
		require.NoError(t, err)

		assigneeID := uuid.New().String()
		followUp, err := svc.CreateFollowUp(ctx, parent.ID, &domain.CreateFollowUpRequest{
			Title:        "Assigned Follow-up",
			AssignedToID: &assigneeID,
		})
		assert.NoError(t, err)
		assert.Equal(t, assigneeID, followUp.AssignedToID)
	})

	t.Run("follow-up defaults assignee to current user", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "My Task",
			ActivityType: domain.ActivityTypeTask,
		}
		parent, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.Complete(ctx, parent.ID, "")
		require.NoError(t, err)

		followUp, err := svc.CreateFollowUp(ctx, parent.ID, &domain.CreateFollowUpRequest{
			Title: "Self-Assigned Follow-up",
		})
		assert.NoError(t, err)
		assert.Equal(t, userCtx.UserID.String(), followUp.AssignedToID)
	})

	t.Run("cannot create follow-up from incomplete activity", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Incomplete Task",
			ActivityType: domain.ActivityTypeTask,
			Status:       domain.ActivityStatusPlanned,
		}
		parent, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.CreateFollowUp(ctx, parent.ID, &domain.CreateFollowUpRequest{
			Title: "Should Fail",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrFollowUpRequiresCompletedParent)
	})

	t.Run("follow-up from non-existent activity", func(t *testing.T) {
		_, err := svc.CreateFollowUp(ctx, uuid.New(), &domain.CreateFollowUpRequest{
			Title: "Should Fail",
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, service.ErrActivityNotFound)
	})

	t.Run("follow-up with due date", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Task with Follow-up Due Date",
			ActivityType: domain.ActivityTypeTask,
		}
		parent, err := svc.Create(ctx, req)
		require.NoError(t, err)

		_, err = svc.Complete(ctx, parent.ID, "")
		require.NoError(t, err)

		dueDate := time.Now().Add(48 * time.Hour)
		followUp, err := svc.CreateFollowUp(ctx, parent.ID, &domain.CreateFollowUpRequest{
			Title:   "Follow-up with Due Date",
			DueDate: &dueDate,
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, followUp.DueDate)
	})
}

func TestActivityService_CreateMeetingWithAttendees(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("create meeting with attendees", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		attendee1 := uuid.New().String()
		attendee2 := uuid.New().String()

		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Team Sync",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
			Attendees:    []string{attendee1, attendee2},
		}
		meeting, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, meeting)
		assert.Len(t, meeting.Attendees, 2)
		assert.Contains(t, meeting.Attendees, attendee1)
		assert.Contains(t, meeting.Attendees, attendee2)
	})

	t.Run("attendees ignored for non-meeting activity", func(t *testing.T) {
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Task with Attendees",
			ActivityType: domain.ActivityTypeTask,
			Attendees:    []string{uuid.New().String()},
		}
		task, err := svc.Create(ctx, req)
		assert.NoError(t, err)
		assert.Empty(t, task.Attendees)
	})
}

func TestActivityService_UpdateMeetingAttendees(t *testing.T) {
	db := setupActivityServiceTestDB(t)
	svc := createActivityService(t, db)
	customer := createActivityServiceTestCustomer(t, db)
	ctx := createActivityTestContext()

	t.Run("update meeting attendees", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		initialAttendee := uuid.New().String()

		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Meeting to Update",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
			Attendees:    []string{initialAttendee},
		}
		meeting, err := svc.Create(ctx, req)
		require.NoError(t, err)
		require.Len(t, meeting.Attendees, 1)

		// Update with new attendees
		newAttendee1 := uuid.New().String()
		newAttendee2 := uuid.New().String()
		updateReq := &domain.UpdateActivityRequest{
			Title:     "Updated Meeting",
			Attendees: []string{newAttendee1, newAttendee2},
		}
		updated, err := svc.Update(ctx, meeting.ID, updateReq)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Len(t, updated.Attendees, 2)
		assert.Contains(t, updated.Attendees, newAttendee1)
		assert.Contains(t, updated.Attendees, newAttendee2)
		assert.NotContains(t, updated.Attendees, initialAttendee)
	})

	t.Run("clear meeting attendees", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		req := &domain.CreateActivityRequest{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Meeting to Clear",
			ActivityType: domain.ActivityTypeMeeting,
			ScheduledAt:  &scheduledAt,
			Attendees:    []string{uuid.New().String()},
		}
		meeting, err := svc.Create(ctx, req)
		require.NoError(t, err)
		require.Len(t, meeting.Attendees, 1)

		// Update with empty attendees
		updateReq := &domain.UpdateActivityRequest{
			Title:     "Cleared Meeting",
			Attendees: []string{},
		}
		updated, err := svc.Update(ctx, meeting.ID, updateReq)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Empty(t, updated.Attendees)
	})
}

