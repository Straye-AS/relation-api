package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// ActivityHandler handles HTTP requests for activities (meetings, tasks, calls, emails, notes)
type ActivityHandler struct {
	activityService *service.ActivityService
	logger          *zap.Logger
}

// NewActivityHandler creates a new ActivityHandler instance
func NewActivityHandler(activityService *service.ActivityService, logger *zap.Logger) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
		logger:          logger,
	}
}

// List godoc
// @Summary List activities
// @Description Get paginated list of activities with optional filters
// @Tags Activities
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param type query string false "Filter by activity type (meeting, task, call, email, note)"
// @Param status query string false "Filter by status (planned, in_progress, completed, cancelled)"
// @Param targetType query string false "Filter by target type (customer, project, offer, deal)"
// @Param targetId query string false "Filter by target entity ID" format(uuid)
// @Param assignedTo query string false "Filter by assigned user ID"
// @Param from query string false "Filter activities from this date (YYYY-MM-DD), inclusive from 00:00:00"
// @Param to query string false "Filter activities to this date (YYYY-MM-DD), inclusive until 23:59:59"
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.ActivityDTO}
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities [get]
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	// Build filters from query params
	filters := &domain.ActivityFilters{}

	// Activity type filter
	if activityType := r.URL.Query().Get("type"); activityType != "" {
		at := domain.ActivityType(activityType)
		if !at.IsValid() {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid activity type. Valid values: meeting, call, email, task, note, system",
			})
			return
		}
		filters.ActivityType = &at
	}

	// Status filter
	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ActivityStatus(status)
		if !s.IsValid() {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid status. Valid values: planned, in_progress, completed, cancelled",
			})
			return
		}
		filters.Status = &s
	}

	// Target type filter
	if targetType := r.URL.Query().Get("targetType"); targetType != "" {
		tt := domain.ActivityTargetType(targetType)
		filters.TargetType = &tt
	}

	// Target ID filter
	if targetIDStr := r.URL.Query().Get("targetId"); targetIDStr != "" {
		targetID, err := uuid.Parse(targetIDStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid targetId format, must be a valid UUID",
			})
			return
		}
		filters.TargetID = &targetID
	}

	// Assigned to filter
	if assignedTo := r.URL.Query().Get("assignedTo"); assignedTo != "" {
		filters.AssignedToID = &assignedTo
	}

	// Date range filter for occurred_at - "from" is start of day (00:00:00)
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		fromDate, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid 'from' format, expected YYYY-MM-DD",
			})
			return
		}
		// Set to start of day in UTC
		fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
		filters.OccurredFrom = &fromDate
	}

	// Date range filter for occurred_at - "to" is end of day (23:59:59.999999999)
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		toDate, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid 'to' format, expected YYYY-MM-DD",
			})
			return
		}
		// Set to end of day in UTC for inclusive filtering
		toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 23, 59, 59, 999999999, time.UTC)
		filters.OccurredTo = &toDate
	}

	result, err := h.activityService.List(r.Context(), filters, page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to list activities", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list activities",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetMyTasks godoc
// @Summary Get my tasks
// @Description Get current user's incomplete tasks (assigned tasks that are not completed or cancelled)
// @Tags Activities
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.ActivityDTO}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/my-tasks [get]
func (h *ActivityHandler) GetMyTasks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}

	result, err := h.activityService.GetMyTasks(r.Context(), page, pageSize)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to get my tasks", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get tasks",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetUpcoming godoc
// @Summary Get upcoming activities
// @Description Get upcoming scheduled activities for the current user within a specified number of days
// @Tags Activities
// @Accept json
// @Produce json
// @Param daysAhead query int false "Number of days to look ahead (1-90)" default(7)
// @Param limit query int false "Maximum number of activities to return (1-100)" default(20)
// @Success 200 {array} domain.ActivityDTO
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/upcoming [get]
func (h *ActivityHandler) GetUpcoming(w http.ResponseWriter, r *http.Request) {
	daysAhead, _ := strconv.Atoi(r.URL.Query().Get("daysAhead"))
	if daysAhead < 1 {
		daysAhead = 7
	}
	if daysAhead > 90 {
		daysAhead = 90
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	activities, err := h.activityService.GetUpcoming(r.Context(), daysAhead, limit)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to get upcoming activities", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get upcoming activities",
		})
		return
	}

	respondJSON(w, http.StatusOK, activities)
}

// GetStats godoc
// @Summary Get activity statistics
// @Description Get activity status counts for the current user's dashboard
// @Tags Activities
// @Accept json
// @Produce json
// @Success 200 {object} domain.ActivityStatusCounts
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/stats [get]
func (h *ActivityHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	counts, err := h.activityService.GetStatusCounts(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to get activity stats", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get activity statistics",
		})
		return
	}

	respondJSON(w, http.StatusOK, counts)
}

// GetByID godoc
// @Summary Get activity by ID
// @Description Get a single activity by its ID
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Success 200 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id} [get]
func (h *ActivityHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	activity, err := h.activityService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have access to this activity",
			})
			return
		}
		h.logger.Error("failed to get activity", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get activity",
		})
		return
	}

	respondJSON(w, http.StatusOK, activity)
}

// Create godoc
// @Summary Create activity
// @Description Create a new activity (meeting, task, call, email, or note)
// @Tags Activities
// @Accept json
// @Produce json
// @Param body body domain.CreateActivityRequest true "Activity data"
// @Success 201 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities [post]
func (h *ActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Validate activity type enum
	if !req.ActivityType.IsValid() {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity type. Valid values: meeting, call, email, task, note, system",
		})
		return
	}

	// Validate status enum if provided
	if req.Status != "" && !req.Status.IsValid() {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid status. Valid values: planned, in_progress, completed, cancelled",
		})
		return
	}

	activity, err := h.activityService.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to create activity", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create activity",
		})
		return
	}

	respondJSON(w, http.StatusCreated, activity)
}

// Update godoc
// @Summary Update activity
// @Description Update an existing activity
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Param body body domain.UpdateActivityRequest true "Activity data"
// @Success 200 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id} [put]
func (h *ActivityHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	var req domain.UpdateActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	// Validate status enum if provided
	if req.Status != "" && !req.Status.IsValid() {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid status. Valid values: planned, in_progress, completed, cancelled",
		})
		return
	}

	activity, err := h.activityService.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to modify this activity",
			})
			return
		}
		h.logger.Error("failed to update activity", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to update activity",
		})
		return
	}

	respondJSON(w, http.StatusOK, activity)
}

// Delete godoc
// @Summary Delete activity
// @Description Delete an activity
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id} [delete]
func (h *ActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	if err := h.activityService.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to delete this activity",
			})
			return
		}
		h.logger.Error("failed to delete activity", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to delete activity",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddAttendee godoc
// @Summary Add attendee to meeting
// @Description Add a user as an attendee to a meeting activity
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Param body body domain.AddAttendeeRequest true "Attendee user ID"
// @Success 200 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id}/attendees [post]
func (h *ActivityHandler) AddAttendee(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	var req domain.AddAttendeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	activity, err := h.activityService.AddAttendee(r.Context(), id, req.UserID)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to modify this activity",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotMeeting) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Attendees can only be added to meeting type activities",
			})
			return
		}
		if errors.Is(err, service.ErrAttendeeAlreadyAdded) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "User is already an attendee",
			})
			return
		}
		h.logger.Error("failed to add attendee", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to add attendee",
		})
		return
	}

	respondJSON(w, http.StatusOK, activity)
}

// RemoveAttendee godoc
// @Summary Remove attendee from meeting
// @Description Remove a user from the attendees list of a meeting activity
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Param userId path string true "User ID to remove"
// @Success 200 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id}/attendees/{userId} [delete]
func (h *ActivityHandler) RemoveAttendee(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	userID := chi.URLParam(r, "userId")
	if userID == "" {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "User ID is required",
		})
		return
	}

	activity, err := h.activityService.RemoveAttendee(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to modify this activity",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotMeeting) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Attendees can only be managed for meeting type activities",
			})
			return
		}
		if errors.Is(err, service.ErrAttendeeNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Attendee not found",
			})
			return
		}
		h.logger.Error("failed to remove attendee", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to remove attendee",
		})
		return
	}

	respondJSON(w, http.StatusOK, activity)
}

// CreateFollowUp godoc
// @Summary Create follow-up task
// @Description Create a follow-up task from a completed activity
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Param body body domain.CreateFollowUpRequest true "Follow-up task data"
// @Success 201 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id}/follow-up [post]
func (h *ActivityHandler) CreateFollowUp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	var req domain.CreateFollowUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
		return
	}

	// Validate request
	if err := validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	activity, err := h.activityService.CreateFollowUp(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to access this activity",
			})
			return
		}
		if errors.Is(err, service.ErrFollowUpRequiresCompletedParent) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Follow-up can only be created from a completed activity",
			})
			return
		}
		h.logger.Error("failed to create follow-up", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to create follow-up",
		})
		return
	}

	respondJSON(w, http.StatusCreated, activity)
}

// Complete godoc
// @Summary Complete activity
// @Description Mark an activity as completed with an optional outcome
// @Tags Activities
// @Accept json
// @Produce json
// @Param id path string true "Activity ID" format(uuid)
// @Param body body domain.CompleteActivityRequest false "Optional outcome"
// @Success 200 {object} domain.ActivityDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /activities/{id}/complete [post]
func (h *ActivityHandler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid activity ID format",
		})
		return
	}

	// Parse optional request body for outcome
	var req domain.CompleteActivityRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: "Invalid request body",
			})
			return
		}

		// Validate request
		if err := validate.Struct(req); err != nil {
			respondValidationError(w, err)
			return
		}
	}

	activity, err := h.activityService.Complete(r.Context(), id, req.Outcome)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrActivityNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Activity not found",
			})
			return
		}
		if errors.Is(err, service.ErrActivityForbidden) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have permission to complete this activity",
			})
			return
		}
		// Check for specific business logic errors
		if errors.Is(err, service.ErrActivityAlreadyCompleted) || errors.Is(err, service.ErrActivityCannotCompleteCancelled) {
			respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
			})
			return
		}
		h.logger.Error("failed to complete activity", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to complete activity",
		})
		return
	}

	respondJSON(w, http.StatusOK, activity)
}
