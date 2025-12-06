package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"gorm.io/gorm"
)

// AuditLogFilter represents filter options for querying audit logs
type AuditLogFilter struct {
	UserID     string
	Action     *domain.AuditAction
	EntityType string
	EntityID   *uuid.UUID
	CompanyID  *domain.CompanyID
	StartTime  *time.Time
	EndTime    *time.Time
	IPAddress  string
	RequestID  string
}

// AuditLogRepository handles audit log data access
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create inserts a new audit log entry (append-only - no updates allowed)
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// CreateBatch inserts multiple audit log entries efficiently
func (r *AuditLogRepository) CreateBatch(ctx context.Context, logs []*domain.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(logs).Error
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	var log domain.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// List retrieves audit logs with pagination and optional filters
func (r *AuditLogRepository) List(ctx context.Context, filter *AuditLogFilter, page, pageSize int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	// Apply filters
	query = r.applyFilters(query, filter)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	err := query.
		Order("performed_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// ListByEntity retrieves audit logs for a specific entity
func (r *AuditLogRepository) ListByEntity(ctx context.Context, entityType string, entityID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("performed_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ListByUser retrieves audit logs for a specific user
func (r *AuditLogRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("performed_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ListByTimeRange retrieves audit logs within a time range
func (r *AuditLogRepository) ListByTimeRange(ctx context.Context, start, end time.Time, page, pageSize int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{}).
		Where("performed_at >= ? AND performed_at <= ?", start, end)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("performed_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}

// CountByAction counts audit logs grouped by action type within a time range
func (r *AuditLogRepository) CountByAction(ctx context.Context, start, end time.Time) (map[domain.AuditAction]int64, error) {
	type result struct {
		Action domain.AuditAction
		Count  int64
	}

	var results []result
	err := r.db.WithContext(ctx).Model(&domain.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("performed_at >= ? AND performed_at <= ?", start, end).
		Group("action").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[domain.AuditAction]int64)
	for _, r := range results {
		counts[r.Action] = r.Count
	}

	return counts, nil
}

// DeleteOlderThan removes audit logs older than a specified duration (for retention policy)
// Note: This should be used with caution and typically only by scheduled cleanup jobs
func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("performed_at < ?", before).
		Delete(&domain.AuditLog{})
	return result.RowsAffected, result.Error
}

// applyFilters applies optional filters to the query
func (r *AuditLogRepository) applyFilters(query *gorm.DB, filter *AuditLogFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}

	if filter.EntityType != "" {
		query = query.Where("entity_type = ?", filter.EntityType)
	}

	if filter.EntityID != nil {
		query = query.Where("entity_id = ?", *filter.EntityID)
	}

	if filter.CompanyID != nil {
		query = query.Where("company_id = ?", *filter.CompanyID)
	}

	if filter.StartTime != nil {
		query = query.Where("performed_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("performed_at <= ?", *filter.EndTime)
	}

	if filter.IPAddress != "" {
		query = query.Where("ip_address = ?", filter.IPAddress)
	}

	if filter.RequestID != "" {
		query = query.Where("request_id = ?", filter.RequestID)
	}

	return query
}
