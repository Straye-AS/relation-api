package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// validNotificationTypes contains all valid notification type values
var validNotificationTypes = map[string]bool{
	string(domain.NotificationTypeTaskAssigned):     true,
	string(domain.NotificationTypeBudgetAlert):      true,
	string(domain.NotificationTypeDealStageChanged): true,
	string(domain.NotificationTypeOfferAccepted):    true,
	string(domain.NotificationTypeOfferRejected):    true,
	string(domain.NotificationTypeActivityReminder): true,
	string(domain.NotificationTypeProjectUpdate):    true,
}

// isValidNotificationType checks if the given type string is a valid NotificationType
func isValidNotificationType(t string) bool {
	return validNotificationTypes[t]
}

// NotificationHandler handles HTTP requests for notifications
type NotificationHandler struct {
	notificationService *service.NotificationService
	logger              *zap.Logger
}

// NewNotificationHandler creates a new NotificationHandler instance
func NewNotificationHandler(notificationService *service.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// List godoc
// @Summary List notifications
// @Description Get paginated list of notifications for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Items per page (max 200)" default(20)
// @Param unreadOnly query bool false "Filter to show only unread notifications" default(false)
// @Param type query string false "Filter by notification type" Enums(task_assigned, budget_alert, deal_stage_changed, offer_accepted, offer_rejected, activity_reminder, project_update)
// @Success 200 {object} domain.PaginatedResponse{data=[]domain.NotificationDTO}
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /notifications [get]
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
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

	unreadOnly := r.URL.Query().Get("unreadOnly") == "true"
	notificationType := r.URL.Query().Get("type")

	// Validate notification type if provided
	if notificationType != "" && !isValidNotificationType(notificationType) {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "invalid notification type: must be one of task_assigned, budget_alert, deal_stage_changed, offer_accepted, offer_rejected, activity_reminder, project_update",
		})
		return
	}

	result, err := h.notificationService.GetForCurrentUser(r.Context(), page, pageSize, unreadOnly, notificationType)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to list notifications", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to list notifications",
		})
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get the count of unread notifications for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} domain.UnreadCountDTO
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /notifications/count [get]
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.notificationService.GetUnreadCount(r.Context())
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to get unread count", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get unread count",
		})
		return
	}

	respondJSON(w, http.StatusOK, count)
}

// GetByID godoc
// @Summary Get notification by ID
// @Description Get a single notification by its ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID" format(uuid)
// @Success 200 {object} domain.NotificationDTO
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid notification ID format",
		})
		return
	}

	notification, err := h.notificationService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrNotificationNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Notification not found",
			})
			return
		}
		if errors.Is(err, service.ErrNotificationNotOwned) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have access to this notification",
			})
			return
		}
		h.logger.Error("failed to get notification", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to get notification",
		})
		return
	}

	respondJSON(w, http.StatusOK, notification)
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a single notification as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID" format(uuid)
// @Success 204 "No Content"
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 403 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, domain.ErrorResponse{
			Error:   "Bad Request",
			Message: "Invalid notification ID format",
		})
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		if errors.Is(err, service.ErrNotificationNotFound) {
			respondJSON(w, http.StatusNotFound, domain.ErrorResponse{
				Error:   "Not Found",
				Message: "Notification not found",
			})
			return
		}
		if errors.Is(err, service.ErrNotificationNotOwned) {
			respondJSON(w, http.StatusForbidden, domain.ErrorResponse{
				Error:   "Forbidden",
				Message: "You do not have access to this notification",
			})
			return
		}
		h.logger.Error("failed to mark notification as read", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to mark notification as read",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications for the current user as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if err := h.notificationService.MarkAllAsReadForUser(r.Context()); err != nil {
		if errors.Is(err, service.ErrUserContextRequired) {
			respondJSON(w, http.StatusUnauthorized, domain.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentication required",
			})
			return
		}
		h.logger.Error("failed to mark all notifications as read", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Failed to mark all notifications as read",
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
