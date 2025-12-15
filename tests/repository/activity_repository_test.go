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

func setupActivityTestDB(t *testing.T) *gorm.DB {
	return testutil.SetupCleanTestDB(t)
}

func createActivityTestCustomer(t *testing.T, db *gorm.DB) *domain.Customer {
	return testutil.CreateTestCustomer(t, db, "Activity Test Customer")
}

func createTestActivity(t *testing.T, db *gorm.DB, customer *domain.Customer, activityType domain.ActivityType) *domain.Activity {
	activity := &domain.Activity{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Test Activity",
		Body:         "Test activity body",
		OccurredAt:   time.Now(),
		ActivityType: activityType,
		Status:       domain.ActivityStatusPlanned,
		Priority:     1,
		IsPrivate:    false,
		CreatorID:    "user-123",
		CreatorName:  "Test User",
		CompanyID:    ptrCompanyID(domain.CompanyStalbygg),
	}
	err := db.Create(activity).Error
	require.NoError(t, err)
	return activity
}

func ptrCompanyID(c domain.CompanyID) *domain.CompanyID {
	return &c
}

func TestActivityRepository_Create(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	t.Run("create activity with minimal fields", func(t *testing.T) {
		activity := &domain.Activity{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "New Activity",
			OccurredAt:   time.Now(),
			ActivityType: domain.ActivityTypeNote,
			Status:       domain.ActivityStatusCompleted,
			CreatorID:    "user-123",
		}

		err := repo.Create(context.Background(), activity)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, activity.ID)
	})

	t.Run("create activity with all fields", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour)
		dueDate := time.Now().Add(48 * time.Hour)
		duration := 60

		activity := &domain.Activity{
			TargetType:      domain.ActivityTargetCustomer,
			TargetID:        customer.ID,
			Title:           "Full Activity",
			Body:            "Activity with all fields",
			OccurredAt:      time.Now(),
			ActivityType:    domain.ActivityTypeMeeting,
			Status:          domain.ActivityStatusPlanned,
			ScheduledAt:     &scheduledAt,
			DueDate:         &dueDate,
			DurationMinutes: &duration,
			Priority:        3,
			IsPrivate:       true,
			CreatorID:       "user-456",
			CreatorName:     "Full User",
			AssignedToID:    "user-789",
			CompanyID:       ptrCompanyID(domain.CompanyStalbygg),
		}

		err := repo.Create(context.Background(), activity)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, activity.ID)
	})
}

func TestActivityRepository_GetByID(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	t.Run("get existing activity", func(t *testing.T) {
		activity := createTestActivity(t, db, customer, domain.ActivityTypeTask)

		found, err := repo.GetByID(context.Background(), activity.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, activity.Title, found.Title)
		assert.Equal(t, activity.TargetType, found.TargetType)
		assert.Equal(t, activity.TargetID, found.TargetID)
		assert.Equal(t, activity.ActivityType, found.ActivityType)
	})

	t.Run("get non-existent activity", func(t *testing.T) {
		found, err := repo.GetByID(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestActivityRepository_Update(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	t.Run("update activity fields", func(t *testing.T) {
		activity := createTestActivity(t, db, customer, domain.ActivityTypeTask)

		activity.Title = "Updated Title"
		activity.Body = "Updated Body"
		activity.Status = domain.ActivityStatusInProgress
		activity.Priority = 5

		err := repo.Update(context.Background(), activity)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), activity.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Title", found.Title)
		assert.Equal(t, "Updated Body", found.Body)
		assert.Equal(t, domain.ActivityStatusInProgress, found.Status)
		assert.Equal(t, 5, found.Priority)
	})

	t.Run("update non-existent activity", func(t *testing.T) {
		activity := &domain.Activity{
			BaseModel: domain.BaseModel{ID: uuid.New()},
			Title:     "Non-existent",
		}

		err := repo.Update(context.Background(), activity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "activity not found")
	})
}

func TestActivityRepository_Delete(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	t.Run("delete existing activity", func(t *testing.T) {
		activity := createTestActivity(t, db, customer, domain.ActivityTypeNote)

		err := repo.Delete(context.Background(), activity.ID)
		assert.NoError(t, err)

		found, err := repo.GetByID(context.Background(), activity.ID)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("delete non-existent activity", func(t *testing.T) {
		err := repo.Delete(context.Background(), uuid.New())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "activity not found")
	})
}

func TestActivityRepository_List(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	// Create test activities
	activities := []*domain.Activity{
		{TargetType: domain.ActivityTargetCustomer, TargetID: customer.ID, Title: "Activity 1", OccurredAt: time.Now(), ActivityType: domain.ActivityTypeTask, Status: domain.ActivityStatusPlanned, CreatorID: "user-1"},
		{TargetType: domain.ActivityTargetCustomer, TargetID: customer.ID, Title: "Activity 2", OccurredAt: time.Now(), ActivityType: domain.ActivityTypeMeeting, Status: domain.ActivityStatusCompleted, CreatorID: "user-1"},
		{TargetType: domain.ActivityTargetCustomer, TargetID: customer.ID, Title: "Activity 3", OccurredAt: time.Now(), ActivityType: domain.ActivityTypeCall, Status: domain.ActivityStatusPlanned, CreatorID: "user-2"},
	}
	for _, a := range activities {
		err := db.Create(a).Error
		require.NoError(t, err)
	}

	t.Run("list all activities", func(t *testing.T) {
		// Filter by customer ID to isolate test data from other tests
		targetType := domain.ActivityTargetCustomer
		result, total, err := repo.List(context.Background(), 1, 10, &targetType, &customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("list with target type filter", func(t *testing.T) {
		// Filter by customer ID to isolate test data from other tests
		targetType := domain.ActivityTargetCustomer
		result, total, err := repo.List(context.Background(), 1, 10, &targetType, &customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("list with target ID filter", func(t *testing.T) {
		targetType := domain.ActivityTargetCustomer
		result, total, err := repo.List(context.Background(), 1, 10, &targetType, &customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("pagination", func(t *testing.T) {
		targetType := domain.ActivityTargetCustomer
		result, total, err := repo.List(context.Background(), 1, 2, &targetType, &customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 2)

		result, total, err = repo.List(context.Background(), 2, 2, &targetType, &customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 1)
	})
}

func TestActivityRepository_ListByTarget(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer1 := createActivityTestCustomer(t, db)
	customer2 := testutil.CreateTestCustomer(t, db, "Second Customer")

	// Create activities for customer1
	for i := 0; i < 3; i++ {
		err := db.Create(&domain.Activity{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer1.ID,
			Title:        "Customer 1 Activity",
			OccurredAt:   time.Now(),
			ActivityType: domain.ActivityTypeNote,
			Status:       domain.ActivityStatusCompleted,
			CreatorID:    "user-1",
		}).Error
		require.NoError(t, err)
	}

	// Create activities for customer2
	for i := 0; i < 2; i++ {
		err := db.Create(&domain.Activity{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer2.ID,
			Title:        "Customer 2 Activity",
			OccurredAt:   time.Now(),
			ActivityType: domain.ActivityTypeNote,
			Status:       domain.ActivityStatusCompleted,
			CreatorID:    "user-1",
		}).Error
		require.NoError(t, err)
	}

	t.Run("list activities by target", func(t *testing.T) {
		result, err := repo.ListByTarget(context.Background(), domain.ActivityTargetCustomer, customer1.ID, 10)
		assert.NoError(t, err)
		assert.Len(t, result, 3)

		for _, activity := range result {
			assert.Equal(t, customer1.ID, activity.TargetID)
		}
	})

	t.Run("list with limit", func(t *testing.T) {
		result, err := repo.ListByTarget(context.Background(), domain.ActivityTargetCustomer, customer1.ID, 2)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestActivityRepository_GetMyTasks(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	userID := "user-repo-getmytasks-" + uuid.New().String()[:8]

	// Create tasks with different statuses for the user
	tasksToCreate := []struct {
		status domain.ActivityStatus
	}{
		{domain.ActivityStatusPlanned},
		{domain.ActivityStatusInProgress},
		{domain.ActivityStatusCompleted}, // Should not be included
		{domain.ActivityStatusCancelled}, // Should not be included
		{domain.ActivityStatusPlanned},
	}

	for _, tc := range tasksToCreate {
		err := db.Create(&domain.Activity{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Task",
			OccurredAt:   time.Now(),
			ActivityType: domain.ActivityTypeTask,
			Status:       tc.status,
			CreatorID:    "creator-1",
			AssignedToID: userID,
		}).Error
		require.NoError(t, err)
	}

	// Create task for a different user
	err := db.Create(&domain.Activity{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Other User Task",
		OccurredAt:   time.Now(),
		ActivityType: domain.ActivityTypeTask,
		Status:       domain.ActivityStatusPlanned,
		CreatorID:    "creator-1",
		AssignedToID: "different-user",
	}).Error
	require.NoError(t, err)

	t.Run("get my tasks excludes completed and cancelled", func(t *testing.T) {
		tasks, total, err := repo.GetMyTasks(context.Background(), userID, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total) // Only planned and in_progress
		assert.Len(t, tasks, 3)

		for _, task := range tasks {
			assert.NotEqual(t, domain.ActivityStatusCompleted, task.Status)
			assert.NotEqual(t, domain.ActivityStatusCancelled, task.Status)
			assert.Equal(t, userID, task.AssignedToID)
		}
	})

	t.Run("pagination works", func(t *testing.T) {
		tasks, total, err := repo.GetMyTasks(context.Background(), userID, 1, 2)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, tasks, 2)
	})
}

func TestActivityRepository_GetUpcoming(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	userID := "user-upcoming"
	now := time.Now()

	// Create activities with different scheduled dates
	scheduledInFuture := now.Add(3 * 24 * time.Hour)
	scheduledFarFuture := now.Add(30 * 24 * time.Hour)
	scheduledPast := now.Add(-24 * time.Hour)

	// Upcoming activity (within 7 days)
	err := db.Create(&domain.Activity{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Upcoming Meeting",
		OccurredAt:   now,
		ActivityType: domain.ActivityTypeMeeting,
		Status:       domain.ActivityStatusPlanned,
		ScheduledAt:  &scheduledInFuture,
		CreatorID:    "creator-1",
		AssignedToID: userID,
	}).Error
	require.NoError(t, err)

	// Far future activity (beyond 7 days)
	err = db.Create(&domain.Activity{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Far Future Meeting",
		OccurredAt:   now,
		ActivityType: domain.ActivityTypeMeeting,
		Status:       domain.ActivityStatusPlanned,
		ScheduledAt:  &scheduledFarFuture,
		CreatorID:    "creator-1",
		AssignedToID: userID,
	}).Error
	require.NoError(t, err)

	// Past activity
	err = db.Create(&domain.Activity{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Past Meeting",
		OccurredAt:   now,
		ActivityType: domain.ActivityTypeMeeting,
		Status:       domain.ActivityStatusPlanned,
		ScheduledAt:  &scheduledPast,
		CreatorID:    "creator-1",
		AssignedToID: userID,
	}).Error
	require.NoError(t, err)

	// Completed activity (should be excluded)
	err = db.Create(&domain.Activity{
		TargetType:   domain.ActivityTargetCustomer,
		TargetID:     customer.ID,
		Title:        "Completed Meeting",
		OccurredAt:   now,
		ActivityType: domain.ActivityTypeMeeting,
		Status:       domain.ActivityStatusCompleted,
		ScheduledAt:  &scheduledInFuture,
		CreatorID:    "creator-1",
		AssignedToID: userID,
	}).Error
	require.NoError(t, err)

	t.Run("get upcoming activities within 7 days", func(t *testing.T) {
		upcoming, err := repo.GetUpcoming(context.Background(), userID, 7, 10)
		assert.NoError(t, err)
		assert.Len(t, upcoming, 1)
		assert.Equal(t, "Upcoming Meeting", upcoming[0].Title)
	})

	t.Run("get upcoming activities within 60 days", func(t *testing.T) {
		upcoming, err := repo.GetUpcoming(context.Background(), userID, 60, 10)
		assert.NoError(t, err)
		assert.Len(t, upcoming, 2) // Includes far future
	})
}

func TestActivityRepository_ListWithFilters(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	// Create diverse activities
	activities := []struct {
		activityType domain.ActivityType
		status       domain.ActivityStatus
		priority     int
		isPrivate    bool
		assignedTo   string
	}{
		{domain.ActivityTypeTask, domain.ActivityStatusPlanned, 1, false, "user-1"},
		{domain.ActivityTypeTask, domain.ActivityStatusCompleted, 2, false, "user-2"},
		{domain.ActivityTypeMeeting, domain.ActivityStatusPlanned, 3, true, "user-1"},
		{domain.ActivityTypeCall, domain.ActivityStatusInProgress, 1, false, "user-3"},
		{domain.ActivityTypeEmail, domain.ActivityStatusPlanned, 2, false, "user-1"},
	}

	for _, a := range activities {
		err := db.Create(&domain.Activity{
			TargetType:   domain.ActivityTargetCustomer,
			TargetID:     customer.ID,
			Title:        "Filtered Activity",
			OccurredAt:   time.Now(),
			ActivityType: a.activityType,
			Status:       a.status,
			Priority:     a.priority,
			IsPrivate:    a.isPrivate,
			AssignedToID: a.assignedTo,
			CreatorID:    "creator-1",
		}).Error
		require.NoError(t, err)
	}

	t.Run("filter by activity type", func(t *testing.T) {
		activityType := domain.ActivityTypeTask
		filters := &domain.ActivityFilters{ActivityType: &activityType, TargetID: &customer.ID}
		result, total, err := repo.ListWithFilters(context.Background(), filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})

	t.Run("filter by status", func(t *testing.T) {
		status := domain.ActivityStatusPlanned
		filters := &domain.ActivityFilters{Status: &status, TargetID: &customer.ID}
		result, total, err := repo.ListWithFilters(context.Background(), filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("filter by assigned user", func(t *testing.T) {
		assignedTo := "user-1"
		filters := &domain.ActivityFilters{AssignedToID: &assignedTo, TargetID: &customer.ID}
		result, total, err := repo.ListWithFilters(context.Background(), filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, result, 3)
	})

	t.Run("filter by private flag", func(t *testing.T) {
		isPrivate := true
		filters := &domain.ActivityFilters{IsPrivate: &isPrivate, TargetID: &customer.ID}
		result, total, err := repo.ListWithFilters(context.Background(), filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := 2
		filters := &domain.ActivityFilters{Priority: &priority, TargetID: &customer.ID}
		result, total, err := repo.ListWithFilters(context.Background(), filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, result, 2)
	})

	t.Run("combined filters", func(t *testing.T) {
		activityType := domain.ActivityTypeTask
		status := domain.ActivityStatusPlanned
		filters := &domain.ActivityFilters{
			ActivityType: &activityType,
			Status:       &status,
			TargetID:     &customer.ID,
		}
		result, total, err := repo.ListWithFilters(context.Background(), filters, 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, result, 1)
	})
}

func TestActivityRepository_CountByStatus(t *testing.T) {
	db := setupActivityTestDB(t)
	repo := repository.NewActivityRepository(db)
	customer := createActivityTestCustomer(t, db)

	userID := "user-repo-countbystatus-" + uuid.New().String()[:8]

	// Create activities with different statuses
	statusCounts := map[domain.ActivityStatus]int{
		domain.ActivityStatusPlanned:    3,
		domain.ActivityStatusInProgress: 2,
		domain.ActivityStatusCompleted:  5,
		domain.ActivityStatusCancelled:  1,
	}

	for status, count := range statusCounts {
		for i := 0; i < count; i++ {
			err := db.Create(&domain.Activity{
				TargetType:   domain.ActivityTargetCustomer,
				TargetID:     customer.ID,
				Title:        "Status Count Activity",
				OccurredAt:   time.Now(),
				ActivityType: domain.ActivityTypeTask,
				Status:       status,
				AssignedToID: userID,
				CreatorID:    "creator-1",
			}).Error
			require.NoError(t, err)
		}
	}

	t.Run("count by status", func(t *testing.T) {
		counts, err := repo.CountByStatus(context.Background(), userID)
		assert.NoError(t, err)
		assert.NotNil(t, counts)
		assert.Equal(t, 3, counts.Planned)
		assert.Equal(t, 2, counts.InProgress)
		assert.Equal(t, 5, counts.Completed)
		assert.Equal(t, 1, counts.Cancelled)
	})
}
