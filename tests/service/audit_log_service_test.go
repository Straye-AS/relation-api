package service_test

import (
	"context"
	"net/http/httptest"
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

func setupAuditLogTestDB(t *testing.T) *gorm.DB {
	db := testutil.SetupCleanTestDB(t)
	t.Cleanup(func() {
		// Clean up audit logs
		db.Exec("DELETE FROM audit_logs WHERE id IS NOT NULL")
	})
	return db
}

func createTestAuditLogService(t *testing.T) (*service.AuditLogService, *gorm.DB) {
	db := setupAuditLogTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewAuditLogRepository(db)
	svc := service.NewAuditLogService(repo, logger)
	return svc, db
}

func TestAuditLogService_LogCreate(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	entityID := uuid.New()
	companyID := domain.CompanyStalbygg

	// Create test request
	req := httptest.NewRequest("POST", "/api/v1/customers", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Create user context
	userCtx := &auth.UserContext{
		UserID:      uuid.New(),
		DisplayName: "Test User",
		Email:       "test@example.com",
	}
	ctx = auth.WithUserContext(ctx, userCtx)

	err := svc.LogCreate(ctx, req, "Customer", entityID, "Test Customer", map[string]string{"name": "Test"}, &companyID)
	require.NoError(t, err)

	// Verify log was created
	var count int64
	db.Model(&domain.AuditLog{}).Count(&count)
	assert.Equal(t, int64(1), count)

	var log domain.AuditLog
	db.First(&log)
	assert.Equal(t, domain.AuditActionCreate, log.Action)
	assert.Equal(t, "Customer", log.EntityType)
	assert.NotNil(t, log.EntityID)
	assert.Equal(t, entityID, *log.EntityID)
	assert.Equal(t, userCtx.UserID.String(), log.UserID)
}

func TestAuditLogService_LogUpdate(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	entityID := uuid.New()

	req := httptest.NewRequest("PUT", "/api/v1/customers/"+entityID.String(), nil)
	req.RemoteAddr = "192.168.1.1:12345"

	oldValues := map[string]string{"name": "Old Name"}
	newValues := map[string]string{"name": "New Name"}

	err := svc.LogUpdate(ctx, req, "Customer", entityID, "Test Customer", oldValues, newValues, nil)
	require.NoError(t, err)

	var log domain.AuditLog
	db.First(&log)
	assert.Equal(t, domain.AuditActionUpdate, log.Action)
	assert.Contains(t, log.OldValues, "Old Name")
	assert.Contains(t, log.NewValues, "New Name")
	assert.NotEmpty(t, log.Changes) // Should have calculated changes
}

func TestAuditLogService_LogDelete(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	entityID := uuid.New()

	req := httptest.NewRequest("DELETE", "/api/v1/customers/"+entityID.String(), nil)
	req.RemoteAddr = "192.168.1.1:12345"

	oldValues := map[string]string{"name": "Deleted Customer"}

	err := svc.LogDelete(ctx, req, "Customer", entityID, "Deleted Customer", oldValues, nil)
	require.NoError(t, err)

	var log domain.AuditLog
	db.First(&log)
	assert.Equal(t, domain.AuditActionDelete, log.Action)
	assert.Contains(t, log.OldValues, "Deleted Customer")
}

func TestAuditLogService_List(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	// Create multiple logs with valid JSONB values and IP address
	userID := "test-user-123"
	for i := 0; i < 5; i++ {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      userID,
			Action:      domain.AuditActionCreate,
			EntityType:  "Customer",
			PerformedAt: time.Now(),
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	// List all
	logs, total, err := svc.List(ctx, service.AuditLogQueryParams{
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, logs, 5)

	// List with user filter
	logs, total, err = svc.List(ctx, service.AuditLogQueryParams{
		UserID:   userID,
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)

	// List with non-existent user
	logs, total, err = svc.List(ctx, service.AuditLogQueryParams{
		UserID:   "non-existent",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, logs, 0)
}

func TestAuditLogService_ListWithPagination(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	// Create 15 logs with valid JSONB values and IP address
	for i := 0; i < 15; i++ {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      "test-user",
			Action:      domain.AuditActionCreate,
			EntityType:  "Customer",
			PerformedAt: time.Now().Add(time.Duration(-i) * time.Minute),
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	// Get first page
	logs, total, err := svc.List(ctx, service.AuditLogQueryParams{
		Page:     1,
		PageSize: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, logs, 5)

	// Get second page
	logs, total, err = svc.List(ctx, service.AuditLogQueryParams{
		Page:     2,
		PageSize: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, logs, 5)

	// Get third page
	logs, total, err = svc.List(ctx, service.AuditLogQueryParams{
		Page:     3,
		PageSize: 5,
	})
	require.NoError(t, err)
	assert.Len(t, logs, 5)
}

func TestAuditLogService_GetByEntity(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	entityID := uuid.New()

	// Create logs for specific entity with valid JSONB values and IP address
	for i := 0; i < 3; i++ {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      "test-user",
			Action:      domain.AuditActionUpdate,
			EntityType:  "Customer",
			EntityID:    &entityID,
			PerformedAt: time.Now(),
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	// Create logs for other entity
	otherID := uuid.New()
	err := db.Create(&domain.AuditLog{
		ID:          uuid.New(),
		UserID:      "test-user",
		Action:      domain.AuditActionCreate,
		EntityType:  "Customer",
		EntityID:    &otherID,
		PerformedAt: time.Now(),
		OldValues:   "null",
		NewValues:   "null",
		Changes:     "null",
		Metadata:    "null",
		IPAddress:   "127.0.0.1",
	}).Error
	require.NoError(t, err)

	logs, err := svc.GetByEntity(ctx, "Customer", entityID, 10)
	require.NoError(t, err)
	assert.Len(t, logs, 3)
}

func TestAuditLogService_GetByUser(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	// Create logs for specific user with valid JSONB values and IP address
	userID := "user-123"
	for i := 0; i < 4; i++ {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      userID,
			Action:      domain.AuditActionCreate,
			EntityType:  "Project",
			PerformedAt: time.Now(),
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	// Create logs for other user
	err := db.Create(&domain.AuditLog{
		ID:          uuid.New(),
		UserID:      "other-user",
		Action:      domain.AuditActionCreate,
		EntityType:  "Project",
		PerformedAt: time.Now(),
		OldValues:   "null",
		NewValues:   "null",
		Changes:     "null",
		Metadata:    "null",
		IPAddress:   "127.0.0.1",
	}).Error
	require.NoError(t, err)

	logs, err := svc.GetByUser(ctx, userID, 10)
	require.NoError(t, err)
	assert.Len(t, logs, 4)
}

func TestAuditLogService_GetStats(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	// Create logs with different actions, valid JSONB values and IP address
	actions := []domain.AuditAction{
		domain.AuditActionCreate,
		domain.AuditActionCreate,
		domain.AuditActionUpdate,
		domain.AuditActionUpdate,
		domain.AuditActionUpdate,
		domain.AuditActionDelete,
	}

	for _, action := range actions {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      "test-user",
			Action:      action,
			EntityType:  "Customer",
			PerformedAt: now,
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	stats, err := svc.GetStats(ctx, start, end)
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats[domain.AuditActionCreate])
	assert.Equal(t, int64(3), stats[domain.AuditActionUpdate])
	assert.Equal(t, int64(1), stats[domain.AuditActionDelete])
}

func TestAuditLogService_CleanupOldLogs(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	now := time.Now()

	// Create old logs (> 30 days) with valid JSONB values and IP address
	for i := 0; i < 3; i++ {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      "test-user",
			Action:      domain.AuditActionCreate,
			EntityType:  "Customer",
			PerformedAt: now.AddDate(0, 0, -60), // 60 days ago
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	// Create recent logs with valid JSONB values and IP address
	for i := 0; i < 2; i++ {
		err := db.Create(&domain.AuditLog{
			ID:          uuid.New(),
			UserID:      "test-user",
			Action:      domain.AuditActionCreate,
			EntityType:  "Customer",
			PerformedAt: now,
			OldValues:   "null",
			NewValues:   "null",
			Changes:     "null",
			Metadata:    "null",
			IPAddress:   "127.0.0.1",
		}).Error
		require.NoError(t, err)
	}

	// Cleanup logs older than 30 days
	count, err := svc.CleanupOldLogs(ctx, 30)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// Verify only recent logs remain
	var remaining int64
	db.Model(&domain.AuditLog{}).Count(&remaining)
	assert.Equal(t, int64(2), remaining)
}

func TestAuditLogService_LogWithClientIP(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		expectedIP    string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.100:12345",
			expectedIP: "192.168.1.100",
		},
		{
			name:          "X-Forwarded-For",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.50",
			expectedIP:    "203.0.113.50",
		},
		{
			name:          "X-Forwarded-For with multiple IPs",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.50, 10.0.0.2",
			expectedIP:    "203.0.113.50",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "198.51.100.25",
			expectedIP: "198.51.100.25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entityID := uuid.New()

			req := httptest.NewRequest("POST", "/api/v1/customers", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			err := svc.LogCreate(ctx, req, "Customer", entityID, "Test", nil, nil)
			require.NoError(t, err)

			var log domain.AuditLog
			db.Where("entity_id = ?", entityID).First(&log)
			assert.Equal(t, tt.expectedIP, log.IPAddress)
		})
	}
}

func TestAuditLogService_CalculatesChanges(t *testing.T) {
	svc, db := createTestAuditLogService(t)
	ctx := context.Background()

	entityID := uuid.New()
	req := httptest.NewRequest("PUT", "/api/v1/customers/"+entityID.String(), nil)
	req.RemoteAddr = "192.168.1.1:12345"

	oldValues := map[string]interface{}{
		"name":    "Old Name",
		"email":   "old@example.com",
		"phone":   "123456789",
		"removed": "this field was removed",
	}
	newValues := map[string]interface{}{
		"name":  "New Name",
		"email": "old@example.com", // unchanged
		"phone": "987654321",
		"added": "new field",
	}

	err := svc.LogUpdate(ctx, req, "Customer", entityID, "Test Customer", oldValues, newValues, nil)
	require.NoError(t, err)

	var log domain.AuditLog
	db.First(&log)

	// Verify changes were calculated
	assert.NotEmpty(t, log.Changes)
	assert.Contains(t, log.Changes, "name")    // changed
	assert.Contains(t, log.Changes, "phone")   // changed
	assert.Contains(t, log.Changes, "removed") // deleted
	assert.Contains(t, log.Changes, "added")   // new
	// email should not be in changes (unchanged)
}
