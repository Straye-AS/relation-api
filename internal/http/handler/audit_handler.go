package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// AuditHandler handles audit log related HTTP requests
type AuditHandler struct {
	auditService *service.AuditLogService
	logger       *zap.Logger
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *service.AuditLogService, logger *zap.Logger) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		logger:       logger,
	}
}

// AuditLogDTO represents an audit log entry for API response
type AuditLogDTO struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"userId,omitempty"`
	UserEmail   string                 `json:"userEmail,omitempty"`
	UserName    string                 `json:"userName,omitempty"`
	Action      string                 `json:"action"`
	EntityType  string                 `json:"entityType"`
	EntityID    string                 `json:"entityId,omitempty"`
	EntityName  string                 `json:"entityName,omitempty"`
	CompanyID   string                 `json:"companyId,omitempty"`
	OldValues   map[string]interface{} `json:"oldValues,omitempty"`
	NewValues   map[string]interface{} `json:"newValues,omitempty"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
	IPAddress   string                 `json:"ipAddress,omitempty"`
	UserAgent   string                 `json:"userAgent,omitempty"`
	RequestID   string                 `json:"requestId,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	PerformedAt string                 `json:"performedAt"`
}

// AuditLogListResponse represents a paginated list of audit logs
type AuditLogListResponse struct {
	Data       []AuditLogDTO `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}

// AuditStatsResponse represents audit log statistics
type AuditStatsResponse struct {
	ActionCounts map[string]int64 `json:"actionCounts"`
	StartTime    string           `json:"startTime"`
	EndTime      string           `json:"endTime"`
}

// List godoc
// @Summary List audit logs
// @Description Returns a paginated list of audit log entries with optional filters
// @Tags Audit
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 20, max: 100)"
// @Param userId query string false "Filter by user ID"
// @Param action query string false "Filter by action type"
// @Param entityType query string false "Filter by entity type"
// @Param entityId query string false "Filter by entity ID"
// @Param startTime query string false "Filter by start time (RFC3339)"
// @Param endTime query string false "Filter by end time (RFC3339)"
// @Success 200 {object} AuditLogListResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /audit [get]
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Check permission
	if !userCtx.HasPermission(domain.PermissionSystemAuditLogs) {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	// Parse query parameters
	page := parseIntQuery(r, "page", 1)
	pageSize := parseIntQuery(r, "pageSize", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	params := service.AuditLogQueryParams{
		UserID:     r.URL.Query().Get("userId"),
		EntityType: r.URL.Query().Get("entityType"),
		Page:       page,
		PageSize:   pageSize,
	}

	// Parse action filter
	if actionStr := r.URL.Query().Get("action"); actionStr != "" {
		action := domain.AuditAction(actionStr)
		params.Action = &action
	}

	// Parse entity ID filter
	if entityIDStr := r.URL.Query().Get("entityId"); entityIDStr != "" {
		if entityID, err := uuid.Parse(entityIDStr); err == nil {
			params.EntityID = &entityID
		}
	}

	// Parse time range filters
	if startStr := r.URL.Query().Get("startTime"); startStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startStr); err == nil {
			params.StartTime = &startTime
		}
	}

	if endStr := r.URL.Query().Get("endTime"); endStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endStr); err == nil {
			params.EndTime = &endTime
		}
	}

	// Get audit logs
	logs, total, err := h.auditService.List(r.Context(), params)
	if err != nil {
		h.logger.Error("failed to list audit logs", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve audit logs"})
		return
	}

	// Convert to DTOs
	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = h.toDTO(log)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	respondJSON(w, http.StatusOK, AuditLogListResponse{
		Data:       dtos,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetByID godoc
// @Summary Get audit log by ID
// @Description Returns a specific audit log entry
// @Tags Audit
// @Produce json
// @Param id path string true "Audit log ID"
// @Success 200 {object} AuditLogDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not found"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /audit/{id} [get]
func (h *AuditHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Check permission
	if !userCtx.HasPermission(domain.PermissionSystemAuditLogs) {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	// Parse ID
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid audit log ID"})
		return
	}

	// Get audit log
	log, err := h.auditService.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get audit log", zap.String("id", idStr), zap.Error(err))
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "audit log not found"})
		return
	}

	respondJSON(w, http.StatusOK, h.toDTO(*log))
}

// GetByEntity godoc
// @Summary Get audit logs for an entity
// @Description Returns audit logs for a specific entity
// @Tags Audit
// @Produce json
// @Param entityType path string true "Entity type (e.g., Customer, Project)"
// @Param entityId path string true "Entity ID"
// @Param limit query int false "Maximum number of entries (default: 50)"
// @Success 200 {array} AuditLogDTO
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /audit/entity/{entityType}/{entityId} [get]
func (h *AuditHandler) GetByEntity(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Check permission
	if !userCtx.HasPermission(domain.PermissionSystemAuditLogs) {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	entityType := chi.URLParam(r, "entityType")
	entityIDStr := chi.URLParam(r, "entityId")

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid entity ID"})
		return
	}

	limit := parseIntQuery(r, "limit", 50)
	if limit > 200 {
		limit = 200
	}

	logs, err := h.auditService.GetByEntity(r.Context(), entityType, entityID, limit)
	if err != nil {
		h.logger.Error("failed to get entity audit logs",
			zap.String("entity_type", entityType),
			zap.String("entity_id", entityIDStr),
			zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve audit logs"})
		return
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = h.toDTO(log)
	}

	respondJSON(w, http.StatusOK, dtos)
}

// GetStats godoc
// @Summary Get audit log statistics
// @Description Returns statistics about audit log actions for a time range
// @Tags Audit
// @Produce json
// @Param startTime query string true "Start time (RFC3339)"
// @Param endTime query string true "End time (RFC3339)"
// @Success 200 {object} AuditStatsResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 400 {object} map[string]string "Bad request"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /audit/stats [get]
func (h *AuditHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Check permission
	if !userCtx.HasPermission(domain.PermissionSystemAuditLogs) {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	startStr := r.URL.Query().Get("startTime")
	endStr := r.URL.Query().Get("endTime")

	if startStr == "" || endStr == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "startTime and endTime are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid startTime format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endTime format"})
		return
	}

	stats, err := h.auditService.GetStats(r.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("failed to get audit stats", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to retrieve statistics"})
		return
	}

	// Convert to string keys for JSON
	actionCounts := make(map[string]int64)
	for action, count := range stats {
		actionCounts[string(action)] = count
	}

	respondJSON(w, http.StatusOK, AuditStatsResponse{
		ActionCounts: actionCounts,
		StartTime:    startTime.Format(time.RFC3339),
		EndTime:      endTime.Format(time.RFC3339),
	})
}

// Export godoc
// @Summary Export audit logs
// @Description Exports audit logs for a time range as JSON
// @Tags Audit
// @Produce application/json
// @Param startTime query string true "Start time (RFC3339)"
// @Param endTime query string true "End time (RFC3339)"
// @Success 200 {array} domain.AuditLog
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 400 {object} map[string]string "Bad request"
// @Security BearerAuth
// @Security ApiKeyAuth
// @Router /audit/export [get]
func (h *AuditHandler) Export(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.FromContext(r.Context())
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	// Check permission
	if !userCtx.HasPermission(domain.PermissionSystemAuditLogs) && !userCtx.HasPermission(domain.PermissionReportsExport) {
		respondJSON(w, http.StatusForbidden, map[string]string{"error": "insufficient permissions"})
		return
	}

	startStr := r.URL.Query().Get("startTime")
	endStr := r.URL.Query().Get("endTime")

	if startStr == "" || endStr == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "startTime and endTime are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid startTime format"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid endTime format"})
		return
	}

	// Log the export action (best effort - ignore errors)
	_ = h.auditService.LogExport(r.Context(), r, "AuditLog", 0, "json", nil)

	data, err := h.auditService.ExportLogs(r.Context(), startTime, endTime)
	if err != nil {
		h.logger.Error("failed to export audit logs", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to export audit logs"})
		return
	}

	// Set headers for file download
	filename := "audit_logs_" + startTime.Format("2006-01-02") + "_" + endTime.Format("2006-01-02") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// toDTO converts an audit log to a DTO
func (h *AuditHandler) toDTO(log domain.AuditLog) AuditLogDTO {
	dto := AuditLogDTO{
		ID:          log.ID.String(),
		UserID:      log.UserID,
		UserEmail:   log.UserEmail,
		UserName:    log.UserName,
		Action:      string(log.Action),
		EntityType:  log.EntityType,
		EntityName:  log.EntityName,
		IPAddress:   log.IPAddress,
		UserAgent:   log.UserAgent,
		RequestID:   log.RequestID,
		PerformedAt: log.PerformedAt.Format(time.RFC3339),
	}

	if log.EntityID != nil {
		dto.EntityID = log.EntityID.String()
	}

	if log.CompanyID != nil {
		dto.CompanyID = string(*log.CompanyID)
	}

	// Parse JSON fields
	if log.OldValues != "" {
		var oldValues map[string]interface{}
		if err := parseJSON(log.OldValues, &oldValues); err == nil {
			dto.OldValues = oldValues
		}
	}

	if log.NewValues != "" {
		var newValues map[string]interface{}
		if err := parseJSON(log.NewValues, &newValues); err == nil {
			dto.NewValues = newValues
		}
	}

	if log.Changes != "" {
		var changes map[string]interface{}
		if err := parseJSON(log.Changes, &changes); err == nil {
			dto.Changes = changes
		}
	}

	if log.Metadata != "" {
		var metadata map[string]interface{}
		if err := parseJSON(log.Metadata, &metadata); err == nil {
			dto.Metadata = metadata
		}
	}

	return dto
}

// parseIntQuery parses an integer query parameter with a default value
func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
