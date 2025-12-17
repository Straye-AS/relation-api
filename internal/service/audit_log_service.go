package service

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/auth"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/repository"
	"go.uber.org/zap"
)

// AuditLogService handles audit logging operations
type AuditLogService struct {
	auditRepo *repository.AuditLogRepository
	logger    *zap.Logger
}

// NewAuditLogService creates a new audit log service
func NewAuditLogService(auditRepo *repository.AuditLogRepository, logger *zap.Logger) *AuditLogService {
	return &AuditLogService{
		auditRepo: auditRepo,
		logger:    logger,
	}
}

// LogEntry represents the input for creating an audit log entry
type LogEntry struct {
	Action     domain.AuditAction
	EntityType string
	EntityID   *uuid.UUID
	EntityName string
	CompanyID  *domain.CompanyID
	OldValues  interface{}
	NewValues  interface{}
	Metadata   map[string]interface{}
}

// Log creates an audit log entry from context and request
func (s *AuditLogService) Log(ctx context.Context, r *http.Request, entry LogEntry) error {
	auditLog := &domain.AuditLog{
		Action:      entry.Action,
		EntityType:  entry.EntityType,
		EntityID:    entry.EntityID,
		EntityName:  entry.EntityName,
		CompanyID:   entry.CompanyID,
		PerformedAt: time.Now(),
	}

	// Extract user info from context
	if userCtx, ok := auth.FromContext(ctx); ok && userCtx != nil {
		auditLog.UserID = userCtx.UserID.String()
		auditLog.UserEmail = userCtx.Email
		auditLog.UserName = userCtx.DisplayName
	}

	// Extract request info
	if r != nil {
		auditLog.IPAddress = s.getClientIP(r)
		auditLog.UserAgent = r.UserAgent()
		auditLog.RequestID = r.Header.Get("X-Request-ID")
	}

	// Serialize old values (use "null" for JSONB compatibility when no value)
	if entry.OldValues != nil {
		if oldJSON, err := json.Marshal(entry.OldValues); err == nil {
			auditLog.OldValues = string(oldJSON)
		} else {
			auditLog.OldValues = "null"
		}
	} else {
		auditLog.OldValues = "null"
	}

	// Serialize new values (use "null" for JSONB compatibility when no value)
	if entry.NewValues != nil {
		if newJSON, err := json.Marshal(entry.NewValues); err == nil {
			auditLog.NewValues = string(newJSON)
		} else {
			auditLog.NewValues = "null"
		}
	} else {
		auditLog.NewValues = "null"
	}

	// Calculate changes if both old and new values exist
	if entry.OldValues != nil && entry.NewValues != nil {
		changes := s.calculateChanges(entry.OldValues, entry.NewValues)
		if changesJSON, err := json.Marshal(changes); err == nil {
			auditLog.Changes = string(changesJSON)
		} else {
			auditLog.Changes = "null"
		}
	} else {
		auditLog.Changes = "null"
	}

	// Serialize metadata (use "null" for JSONB compatibility when no value)
	if entry.Metadata != nil {
		if metaJSON, err := json.Marshal(entry.Metadata); err == nil {
			auditLog.Metadata = string(metaJSON)
		} else {
			auditLog.Metadata = "null"
		}
	} else {
		auditLog.Metadata = "null"
	}

	err := s.auditRepo.Create(ctx, auditLog)
	if err != nil {
		s.logger.Error("failed to create audit log",
			zap.String("action", string(entry.Action)),
			zap.String("entity_type", entry.EntityType),
			zap.Error(err))
		return err
	}

	return nil
}

// LogCreate logs a create operation
func (s *AuditLogService) LogCreate(ctx context.Context, r *http.Request, entityType string, entityID uuid.UUID, entityName string, newValues interface{}, companyID *domain.CompanyID) error {
	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionCreate,
		EntityType: entityType,
		EntityID:   &entityID,
		EntityName: entityName,
		CompanyID:  companyID,
		NewValues:  newValues,
	})
}

// LogUpdate logs an update operation
func (s *AuditLogService) LogUpdate(ctx context.Context, r *http.Request, entityType string, entityID uuid.UUID, entityName string, oldValues, newValues interface{}, companyID *domain.CompanyID) error {
	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionUpdate,
		EntityType: entityType,
		EntityID:   &entityID,
		EntityName: entityName,
		CompanyID:  companyID,
		OldValues:  oldValues,
		NewValues:  newValues,
	})
}

// LogDelete logs a delete operation
func (s *AuditLogService) LogDelete(ctx context.Context, r *http.Request, entityType string, entityID uuid.UUID, entityName string, oldValues interface{}, companyID *domain.CompanyID) error {
	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionDelete,
		EntityType: entityType,
		EntityID:   &entityID,
		EntityName: entityName,
		CompanyID:  companyID,
		OldValues:  oldValues,
	})
}

// LogPermissionGrant logs a permission grant
func (s *AuditLogService) LogPermissionGrant(ctx context.Context, r *http.Request, userID string, permission string, companyID *domain.CompanyID, reason string) error {
	targetUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Warn("could not parse user ID for audit log", zap.String("user_id", userID))
		return nil
	}

	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionPermissionGrant,
		EntityType: "User",
		EntityID:   &targetUUID,
		CompanyID:  companyID,
		NewValues: map[string]interface{}{
			"permission": permission,
			"reason":     reason,
		},
	})
}

// LogPermissionRevoke logs a permission revocation
func (s *AuditLogService) LogPermissionRevoke(ctx context.Context, r *http.Request, userID string, permission string, companyID *domain.CompanyID, reason string) error {
	targetUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Warn("could not parse user ID for audit log", zap.String("user_id", userID))
		return nil
	}

	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionPermissionRevoke,
		EntityType: "User",
		EntityID:   &targetUUID,
		CompanyID:  companyID,
		OldValues: map[string]interface{}{
			"permission": permission,
			"reason":     reason,
		},
	})
}

// LogRoleAssign logs a role assignment
func (s *AuditLogService) LogRoleAssign(ctx context.Context, r *http.Request, userID string, role string, companyID *domain.CompanyID) error {
	targetUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Warn("could not parse user ID for audit log", zap.String("user_id", userID))
		return nil
	}

	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionRoleAssign,
		EntityType: "User",
		EntityID:   &targetUUID,
		CompanyID:  companyID,
		NewValues: map[string]interface{}{
			"role": role,
		},
	})
}

// LogRoleRemove logs a role removal
func (s *AuditLogService) LogRoleRemove(ctx context.Context, r *http.Request, userID string, role string, companyID *domain.CompanyID) error {
	targetUUID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Warn("could not parse user ID for audit log", zap.String("user_id", userID))
		return nil
	}

	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionRoleRemove,
		EntityType: "User",
		EntityID:   &targetUUID,
		CompanyID:  companyID,
		OldValues: map[string]interface{}{
			"role": role,
		},
	})
}

// LogExport logs an export operation
func (s *AuditLogService) LogExport(ctx context.Context, r *http.Request, entityType string, count int, format string, companyID *domain.CompanyID) error {
	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionExport,
		EntityType: entityType,
		CompanyID:  companyID,
		Metadata: map[string]interface{}{
			"count":  count,
			"format": format,
		},
	})
}

// LogImport logs an import operation
func (s *AuditLogService) LogImport(ctx context.Context, r *http.Request, entityType string, count int, source string, companyID *domain.CompanyID) error {
	return s.Log(ctx, r, LogEntry{
		Action:     domain.AuditActionImport,
		EntityType: entityType,
		CompanyID:  companyID,
		Metadata: map[string]interface{}{
			"count":  count,
			"source": source,
		},
	})
}

// AuditLogQueryParams represents query parameters for listing audit logs
type AuditLogQueryParams struct {
	UserID     string
	Action     *domain.AuditAction
	EntityType string
	EntityID   *uuid.UUID
	CompanyID  *domain.CompanyID
	StartTime  *time.Time
	EndTime    *time.Time
	Page       int
	PageSize   int
}

// List retrieves audit logs with filters
func (s *AuditLogService) List(ctx context.Context, params AuditLogQueryParams) ([]domain.AuditLog, int64, error) {
	filter := &repository.AuditLogFilter{
		UserID:     params.UserID,
		Action:     params.Action,
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		CompanyID:  params.CompanyID,
		StartTime:  params.StartTime,
		EndTime:    params.EndTime,
	}

	return s.auditRepo.List(ctx, filter, params.Page, params.PageSize)
}

// GetByID retrieves a specific audit log entry
func (s *AuditLogService) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	return s.auditRepo.GetByID(ctx, id)
}

// GetByEntity retrieves audit logs for a specific entity
func (s *AuditLogService) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	return s.auditRepo.ListByEntity(ctx, entityType, entityID, limit)
}

// GetByUser retrieves audit logs for a specific user's actions
func (s *AuditLogService) GetByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	return s.auditRepo.ListByUser(ctx, userID, limit)
}

// GetStats returns audit log statistics for a time range
func (s *AuditLogService) GetStats(ctx context.Context, start, end time.Time) (map[domain.AuditAction]int64, error) {
	return s.auditRepo.CountByAction(ctx, start, end)
}

// CleanupOldLogs removes logs older than the specified retention period
func (s *AuditLogService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	before := time.Now().AddDate(0, 0, -retentionDays)
	count, err := s.auditRepo.DeleteOlderThan(ctx, before)
	if err != nil {
		s.logger.Error("failed to cleanup old audit logs",
			zap.Int("retention_days", retentionDays),
			zap.Error(err))
		return 0, err
	}

	if count > 0 {
		s.logger.Info("cleaned up old audit logs",
			zap.Int64("deleted_count", count),
			zap.Int("retention_days", retentionDays))
	}

	return count, nil
}

// ExportLogs exports audit logs for a time range as JSON
func (s *AuditLogService) ExportLogs(ctx context.Context, start, end time.Time) ([]byte, error) {
	var allLogs []domain.AuditLog
	page := 1
	pageSize := 1000

	for {
		logs, _, err := s.auditRepo.ListByTimeRange(ctx, start, end, page, pageSize)
		if err != nil {
			return nil, err
		}

		allLogs = append(allLogs, logs...)

		if len(logs) < pageSize {
			break
		}
		page++
	}

	return json.MarshalIndent(allLogs, "", "  ")
}

// calculateChanges determines what changed between old and new values
func (s *AuditLogService) calculateChanges(oldValues, newValues interface{}) map[string]interface{} {
	changes := make(map[string]interface{})

	oldMap := s.toMap(oldValues)
	newMap := s.toMap(newValues)

	// Find modified and new fields
	for key, newVal := range newMap {
		if oldVal, exists := oldMap[key]; exists {
			if !reflect.DeepEqual(oldVal, newVal) {
				changes[key] = map[string]interface{}{
					"old": oldVal,
					"new": newVal,
				}
			}
		} else {
			changes[key] = map[string]interface{}{
				"old": nil,
				"new": newVal,
			}
		}
	}

	// Find deleted fields
	for key, oldVal := range oldMap {
		if _, exists := newMap[key]; !exists {
			changes[key] = map[string]interface{}{
				"old": oldVal,
				"new": nil,
			}
		}
	}

	return changes
}

// toMap converts an interface to a map for comparison
func (s *AuditLogService) toMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	if v == nil {
		return result
	}

	// If already a map, return it
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	// Try to marshal and unmarshal to get a map
	data, err := json.Marshal(v)
	if err != nil {
		return result
	}

	_ = json.Unmarshal(data, &result)
	return result
}

// getClientIP extracts the client IP from the request
func (s *AuditLogService) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr (remove port)
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
