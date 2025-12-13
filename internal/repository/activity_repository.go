package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// ActivityRepository handles database operations for activities.
//
// Index recommendations for optimal query performance:
// - CREATE INDEX idx_activities_assigned_to_id ON activities(assigned_to_id) WHERE assigned_to_id IS NOT NULL;
// - CREATE INDEX idx_activities_status ON activities(status);
// - CREATE INDEX idx_activities_due_date ON activities(due_date) WHERE due_date IS NOT NULL;
// - CREATE INDEX idx_activities_scheduled_at ON activities(scheduled_at) WHERE scheduled_at IS NOT NULL;
// - CREATE INDEX idx_activities_company_id ON activities(company_id) WHERE company_id IS NOT NULL;
// - CREATE INDEX idx_activities_target ON activities(target_type, target_id);
type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *ActivityRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Activity, error) {
	var activity domain.Activity
	query := r.db.WithContext(ctx).Where("id = ?", id)
	query = ApplyCompanyFilter(ctx, query)
	err := query.First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// Update updates an existing activity
func (r *ActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	// Verify the activity exists and user has access
	existing, err := r.GetByID(ctx, activity.ID)
	if err != nil {
		return fmt.Errorf("activity not found: %w", err)
	}

	// Preserve created_at and company_id from original
	activity.CreatedAt = existing.CreatedAt
	activity.CompanyID = existing.CompanyID

	return r.db.WithContext(ctx).Save(activity).Error
}

// Delete removes an activity by ID
func (r *ActivityRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Verify the activity exists and user has access before deleting
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("activity not found: %w", err)
	}

	return r.db.WithContext(ctx).Delete(&domain.Activity{}, "id = ?", id).Error
}

func (r *ActivityRepository) List(ctx context.Context, page, pageSize int, targetType *domain.ActivityTargetType, targetID *uuid.UUID) ([]domain.Activity, int64, error) {
	var activities []domain.Activity
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Activity{})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if targetType != nil {
		query = query.Where("target_type = ?", *targetType)
	}

	if targetID != nil {
		query = query.Where("target_id = ?", *targetID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("occurred_at DESC").Find(&activities).Error

	return activities, total, err
}

func (r *ActivityRepository) ListByTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, limit int) ([]domain.Activity, error) {
	var activities []domain.Activity
	query := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ?", targetType, targetID)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Order("occurred_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}

// GetMyTasks retrieves tasks assigned to the specified user that are not completed or cancelled.
// Tasks are ordered by due_date (nulls last), then by priority (descending), then by created_at.
func (r *ActivityRepository) GetMyTasks(ctx context.Context, userID string, page, pageSize int) ([]domain.Activity, int64, error) {
	var activities []domain.Activity
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Activity{}).
		Where("assigned_to_id = ?", userID).
		Where("status NOT IN ?", []domain.ActivityStatus{
			domain.ActivityStatusCompleted,
			domain.ActivityStatusCancelled,
		})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting tasks: %w", err)
	}

	offset := (page - 1) * pageSize
	// Order by due_date (nulls last), priority desc, then created_at
	err := query.Offset(offset).Limit(pageSize).
		Order("CASE WHEN due_date IS NULL THEN 1 ELSE 0 END, due_date ASC, priority DESC, created_at DESC").
		Find(&activities).Error

	if err != nil {
		return nil, 0, fmt.Errorf("fetching tasks: %w", err)
	}

	return activities, total, nil
}

// GetUpcoming retrieves activities with scheduled_at within the specified number of days ahead.
// Results are ordered by scheduled_at ascending.
func (r *ActivityRepository) GetUpcoming(ctx context.Context, userID string, daysAhead int, limit int) ([]domain.Activity, error) {
	var activities []domain.Activity

	now := time.Now()
	endDate := now.AddDate(0, 0, daysAhead)

	query := r.db.WithContext(ctx).
		Where("assigned_to_id = ?", userID).
		Where("scheduled_at IS NOT NULL").
		Where("scheduled_at >= ?", now).
		Where("scheduled_at <= ?", endDate).
		Where("status NOT IN ?", []domain.ActivityStatus{
			domain.ActivityStatusCompleted,
			domain.ActivityStatusCancelled,
		})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	err := query.Order("scheduled_at ASC").
		Limit(limit).
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("fetching upcoming activities: %w", err)
	}

	return activities, nil
}

// ListWithFilters retrieves activities matching the provided filters with pagination.
// All filter fields are optional and will be applied if non-nil.
func (r *ActivityRepository) ListWithFilters(ctx context.Context, filters *domain.ActivityFilters, page, pageSize int) ([]domain.Activity, int64, error) {
	var activities []domain.Activity
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Activity{})

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	// Apply filters if provided
	if filters != nil {
		query = r.applyFilters(query, filters)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting filtered activities: %w", err)
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("occurred_at DESC").
		Find(&activities).Error

	if err != nil {
		return nil, 0, fmt.Errorf("fetching filtered activities: %w", err)
	}

	return activities, total, nil
}

// applyFilters applies all non-nil filter values to the query
func (r *ActivityRepository) applyFilters(query *gorm.DB, filters *domain.ActivityFilters) *gorm.DB {
	if filters.ActivityType != nil {
		query = query.Where("activity_type = ?", *filters.ActivityType)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", *filters.Status)
	}

	if filters.TargetType != nil {
		query = query.Where("target_type = ?", *filters.TargetType)
	}

	if filters.TargetID != nil {
		query = query.Where("target_id = ?", *filters.TargetID)
	}

	if filters.AssignedToID != nil {
		query = query.Where("assigned_to_id = ?", *filters.AssignedToID)
	}

	if filters.CreatorID != nil {
		query = query.Where("creator_id = ?", *filters.CreatorID)
	}

	if filters.DueDateFrom != nil {
		query = query.Where("due_date >= ?", *filters.DueDateFrom)
	}

	if filters.DueDateTo != nil {
		query = query.Where("due_date <= ?", *filters.DueDateTo)
	}

	if filters.ScheduledFrom != nil {
		query = query.Where("scheduled_at >= ?", *filters.ScheduledFrom)
	}

	if filters.ScheduledTo != nil {
		query = query.Where("scheduled_at <= ?", *filters.ScheduledTo)
	}

	if filters.IsPrivate != nil {
		query = query.Where("is_private = ?", *filters.IsPrivate)
	}

	if filters.Priority != nil {
		query = query.Where("priority = ?", *filters.Priority)
	}

	return query
}

// CountByStatus returns a map of activity counts grouped by status for a specific user.
// This is useful for dashboard statistics showing task distribution.
func (r *ActivityRepository) CountByStatus(ctx context.Context, userID string) (*domain.ActivityStatusCounts, error) {
	var results []struct {
		Status domain.ActivityStatus
		Count  int
	}

	query := r.db.WithContext(ctx).Model(&domain.Activity{}).
		Select("status, COUNT(*) as count").
		Where("assigned_to_id = ?", userID).
		Group("status")

	// Apply multi-tenant company filter
	query = ApplyCompanyFilter(ctx, query)

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("counting activities by status: %w", err)
	}

	// Convert results to ActivityStatusCounts
	counts := &domain.ActivityStatusCounts{}
	for _, r := range results {
		switch r.Status {
		case domain.ActivityStatusPlanned:
			counts.Planned = r.Count
		case domain.ActivityStatusInProgress:
			counts.InProgress = r.Count
		case domain.ActivityStatusCompleted:
			counts.Completed = r.Count
		case domain.ActivityStatusCancelled:
			counts.Cancelled = r.Count
		}
	}

	return counts, nil
}

// GetRecentActivities returns the most recent activities
func (r *ActivityRepository) GetRecentActivities(ctx context.Context, limit int) ([]domain.Activity, error) {
	var activities []domain.Activity
	query := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&activities).Error
	return activities, err
}

// GetRecentActivitiesInWindow returns the most recent activities within a time window
// If since is nil, no date filter is applied (all time)
func (r *ActivityRepository) GetRecentActivitiesInWindow(ctx context.Context, since *time.Time, limit int) ([]domain.Activity, error) {
	var activities []domain.Activity
	query := r.db.WithContext(ctx)
	if since != nil {
		query = query.Where("created_at >= ?", *since)
	}
	query = query.Order("created_at DESC").Limit(limit)
	query = ApplyCompanyFilter(ctx, query)
	err := query.Find(&activities).Error
	return activities, err
}
