package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/mapper"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Activity service errors
var (
	ErrActivityNotFound                = errors.New("activity not found")
	ErrActivityForbidden               = errors.New("access to activity denied")
	ErrActivityAlreadyCompleted        = errors.New("activity is already completed")
	ErrActivityCannotCompleteCancelled = errors.New("cannot complete a cancelled activity")
)

// ActivityService handles business logic for activities (meetings, tasks, calls, emails, notes)
type ActivityService struct {
	activityRepo        *repository.ActivityRepository
	notificationService *NotificationService
	logger              *zap.Logger
}

// NewActivityService creates a new ActivityService instance
func NewActivityService(
	activityRepo *repository.ActivityRepository,
	notificationService *NotificationService,
	logger *zap.Logger,
) *ActivityService {
	return &ActivityService{
		activityRepo:        activityRepo,
		notificationService: notificationService,
		logger:              logger,
	}
}

// Create creates a new activity and sends notification if assigned to a user
func (s *ActivityService) Create(ctx context.Context, req *domain.CreateActivityRequest) (*domain.ActivityDTO, error) {
	// Get user context for creator info
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = domain.ActivityStatusPlanned
	}

	// Determine company ID (from request or user context)
	var companyID *domain.CompanyID
	if req.CompanyID != nil {
		companyID = req.CompanyID
	} else {
		// Use the user's company filter (returns nil for super admins/Gruppen users)
		companyID = userCtx.GetCompanyFilter()
		// If user is not filtering (super admin/Gruppen), explicitly set their company
		if companyID == nil && userCtx.CompanyID != "" {
			companyID = &userCtx.CompanyID
		}
	}

	// Log warning if due date is in the past for tasks (but allow it)
	if req.DueDate != nil && req.DueDate.Before(time.Now()) && req.ActivityType == domain.ActivityTypeTask {
		s.logger.Warn("task created with due date in the past",
			zap.String("creator_id", userCtx.UserID.String()),
			zap.Time("due_date", *req.DueDate),
		)
	}

	now := time.Now()
	activity := &domain.Activity{
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		Title:           req.Title,
		Body:            req.Body,
		OccurredAt:      now,
		ActivityType:    req.ActivityType,
		Status:          status,
		ScheduledAt:     req.ScheduledAt,
		DueDate:         req.DueDate,
		DurationMinutes: req.DurationMinutes,
		Priority:        req.Priority,
		IsPrivate:       req.IsPrivate,
		CreatorID:       userCtx.UserID.String(),
		CreatorName:     userCtx.DisplayName,
		AssignedToID:    req.AssignedToID,
		CompanyID:       companyID,
	}

	if err := s.activityRepo.Create(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	s.logger.Info("activity created",
		zap.String("activity_id", activity.ID.String()),
		zap.String("type", string(activity.ActivityType)),
		zap.String("creator_id", activity.CreatorID),
	)

	// Send notification if task is assigned to someone other than creator
	if req.AssignedToID != "" && req.AssignedToID != userCtx.UserID.String() {
		s.sendTaskAssignmentNotification(ctx, activity, userCtx.DisplayName)
	}

	dto := mapper.ToActivityDTO(activity)
	return &dto, nil
}

// Update updates an existing activity
func (s *ActivityService) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateActivityRequest) (*domain.ActivityDTO, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	// Get existing activity
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Check permission: creator, assigned user, or manager can modify
	if !s.canModifyActivity(userCtx, activity) {
		return nil, ErrActivityForbidden
	}

	// Track if assignment changed for notification
	oldAssignedToID := activity.AssignedToID
	assignmentChanged := req.AssignedToID != oldAssignedToID && req.AssignedToID != ""

	// Update fields
	activity.Title = req.Title
	activity.Body = req.Body
	if req.Status != "" {
		activity.Status = req.Status
	}
	activity.ScheduledAt = req.ScheduledAt
	activity.DueDate = req.DueDate
	activity.DurationMinutes = req.DurationMinutes
	activity.Priority = req.Priority
	activity.IsPrivate = req.IsPrivate
	activity.AssignedToID = req.AssignedToID

	if err := s.activityRepo.Update(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to update activity: %w", err)
	}

	s.logger.Info("activity updated",
		zap.String("activity_id", activity.ID.String()),
		zap.String("updated_by", userCtx.UserID.String()),
	)

	// Send notification if assignment changed to a different user (not self)
	if assignmentChanged && req.AssignedToID != userCtx.UserID.String() {
		s.sendTaskAssignmentNotification(ctx, activity, userCtx.DisplayName)
	}

	dto := mapper.ToActivityDTO(activity)
	return &dto, nil
}

// Delete removes an activity
func (s *ActivityService) Delete(ctx context.Context, id uuid.UUID) error {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return ErrUserContextRequired
	}

	// Get existing activity
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrActivityNotFound
		}
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// Check permission
	if !s.canModifyActivity(userCtx, activity) {
		return ErrActivityForbidden
	}

	if err := s.activityRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete activity: %w", err)
	}

	s.logger.Info("activity deleted",
		zap.String("activity_id", id.String()),
		zap.String("deleted_by", userCtx.UserID.String()),
	)

	return nil
}

// GetByID retrieves a single activity by ID
func (s *ActivityService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ActivityDTO, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Check access to private activities
	if activity.IsPrivate && !s.canViewActivity(userCtx, activity) {
		return nil, ErrActivityForbidden
	}

	dto := mapper.ToActivityDTO(activity)
	return &dto, nil
}

// List returns a paginated list of activities with optional filters
func (s *ActivityService) List(ctx context.Context, filters *domain.ActivityFilters, page, pageSize int) (*domain.PaginatedResponse, error) {
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

	activities, total, err := s.activityRepo.ListWithFilters(ctx, filters, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
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

// Complete marks an activity as completed with an optional outcome
func (s *ActivityService) Complete(ctx context.Context, id uuid.UUID, outcome string) (*domain.ActivityDTO, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	// Get existing activity
	activity, err := s.activityRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrActivityNotFound
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Check permission
	if !s.canModifyActivity(userCtx, activity) {
		return nil, ErrActivityForbidden
	}

	// Cannot complete an already completed or cancelled activity
	if activity.Status == domain.ActivityStatusCompleted {
		return nil, ErrActivityAlreadyCompleted
	}
	if activity.Status == domain.ActivityStatusCancelled {
		return nil, ErrActivityCannotCompleteCancelled
	}

	// Update status and completion time
	now := time.Now()
	activity.Status = domain.ActivityStatusCompleted
	activity.CompletedAt = &now

	// Add outcome to body if provided
	if outcome != "" {
		if activity.Body != "" {
			activity.Body = activity.Body + "\n\n--- Outcome ---\n" + outcome
		} else {
			activity.Body = "--- Outcome ---\n" + outcome
		}
	}

	if err := s.activityRepo.Update(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to complete activity: %w", err)
	}

	s.logger.Info("activity completed",
		zap.String("activity_id", activity.ID.String()),
		zap.String("completed_by", userCtx.UserID.String()),
	)

	dto := mapper.ToActivityDTO(activity)
	return &dto, nil
}

// GetMyTasks retrieves tasks assigned to the current user that are not completed/cancelled
func (s *ActivityService) GetMyTasks(ctx context.Context, page, pageSize int) (*domain.PaginatedResponse, error) {
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

	activities, total, err := s.activityRepo.GetMyTasks(ctx, userCtx.UserID.String(), page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
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

// GetUpcoming retrieves upcoming scheduled activities for the current user
func (s *ActivityService) GetUpcoming(ctx context.Context, daysAhead, limit int) ([]domain.ActivityDTO, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	// Set reasonable defaults and limits
	if daysAhead < 1 {
		daysAhead = 7
	}
	if daysAhead > 90 {
		daysAhead = 90
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	activities, err := s.activityRepo.GetUpcoming(ctx, userCtx.UserID.String(), daysAhead, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming activities: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}

// GetStatusCounts returns activity counts grouped by status for the current user's dashboard
func (s *ActivityService) GetStatusCounts(ctx context.Context) (*domain.ActivityStatusCounts, error) {
	userCtx, ok := auth.FromContext(ctx)
	if !ok {
		return nil, ErrUserContextRequired
	}

	counts, err := s.activityRepo.CountByStatus(ctx, userCtx.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}

	return counts, nil
}

// GetByTarget retrieves activities for a specific target entity
func (s *ActivityService) GetByTarget(ctx context.Context, targetType domain.ActivityTargetType, targetID uuid.UUID, limit int) ([]domain.ActivityDTO, error) {
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	activities, err := s.activityRepo.ListByTarget(ctx, targetType, targetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities by target: %w", err)
	}

	dtos := make([]domain.ActivityDTO, len(activities))
	for i, activity := range activities {
		dtos[i] = mapper.ToActivityDTO(&activity)
	}

	return dtos, nil
}

// sendTaskAssignmentNotification sends a notification when a task is assigned
func (s *ActivityService) sendTaskAssignmentNotification(ctx context.Context, activity *domain.Activity, assignerName string) {
	if s.notificationService == nil {
		s.logger.Warn("notification service not available, skipping task assignment notification")
		return
	}

	// Parse assigned user ID to UUID
	assignedUserID, err := uuid.Parse(activity.AssignedToID)
	if err != nil {
		s.logger.Warn("invalid assigned user ID for notification",
			zap.String("assigned_to_id", activity.AssignedToID),
			zap.Error(err),
		)
		return
	}

	// Build notification message
	title := "New Task Assigned"
	message := fmt.Sprintf("%s assigned you a task: %s", assignerName, activity.Title)
	if activity.DueDate != nil {
		message += fmt.Sprintf(" (due: %s)", activity.DueDate.Format("2006-01-02"))
	}

	_, err = s.notificationService.CreateForUser(
		ctx,
		assignedUserID,
		domain.NotificationTypeTaskAssigned,
		title,
		message,
		"activity",
		&activity.ID,
	)
	if err != nil {
		s.logger.Warn("failed to send task assignment notification",
			zap.String("activity_id", activity.ID.String()),
			zap.String("assigned_to", activity.AssignedToID),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("task assignment notification sent",
		zap.String("activity_id", activity.ID.String()),
		zap.String("assigned_to", activity.AssignedToID),
	)
}

// canModifyActivity checks if the user has permission to modify an activity
func (s *ActivityService) canModifyActivity(userCtx *auth.UserContext, activity *domain.Activity) bool {
	userID := userCtx.UserID.String()

	// Creator can always modify
	if activity.CreatorID == userID {
		return true
	}

	// Assigned user can modify
	if activity.AssignedToID == userID {
		return true
	}

	// Managers and admins can modify
	if userCtx.HasAnyRole(domain.RoleManager, domain.RoleCompanyAdmin, domain.RoleSuperAdmin) {
		return true
	}

	return false
}

// canViewActivity checks if the user has permission to view an activity
func (s *ActivityService) canViewActivity(userCtx *auth.UserContext, activity *domain.Activity) bool {
	userID := userCtx.UserID.String()

	// Non-private activities are viewable by anyone in the company
	if !activity.IsPrivate {
		return true
	}

	// Creator can always view
	if activity.CreatorID == userID {
		return true
	}

	// Assigned user can view
	if activity.AssignedToID == userID {
		return true
	}

	// Managers and admins can view private activities
	if userCtx.HasAnyRole(domain.RoleManager, domain.RoleCompanyAdmin, domain.RoleSuperAdmin) {
		return true
	}

	return false
}
