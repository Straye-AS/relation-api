package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/straye-as/relation-api/internal/domain"
	"github.com/straye-as/relation-api/internal/service"
	"go.uber.org/zap"
)

// AuditConfig holds configuration for audit middleware
type AuditConfig struct {
	// SkipPaths contains paths that should not be audited
	SkipPaths []string
	// SkipMethods contains HTTP methods that should not be audited (e.g., GET, OPTIONS)
	SkipMethods []string
	// AuditReads enables auditing of GET requests (defaults to false)
	AuditReads bool
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		SkipPaths: []string{
			"/health",
			"/health/db",
			"/health/ready",
			"/swagger",
		},
		SkipMethods: []string{
			http.MethodOptions,
			http.MethodHead,
		},
		AuditReads: false,
	}
}

// AuditMiddleware provides audit logging for HTTP requests
type AuditMiddleware struct {
	auditService *service.AuditLogService
	config       *AuditConfig
	logger       *zap.Logger
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditService *service.AuditLogService, config *AuditConfig, logger *zap.Logger) *AuditMiddleware {
	if config == nil {
		config = DefaultAuditConfig()
	}
	return &AuditMiddleware{
		auditService: auditService,
		config:       config,
		logger:       logger,
	}
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	auditRequestBodyKey contextKey = "audit_request_body"
)

// Audit returns middleware that logs modifications to the audit log
func (m *AuditMiddleware) Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this request should be audited
		if !m.shouldAudit(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Capture request body for POST/PUT/PATCH
		var requestBody []byte
		if r.Body != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Store request body in context for later use
		ctx := context.WithValue(r.Context(), auditRequestBodyKey, requestBody)
		r = r.WithContext(ctx)

		// Wrap response writer to capture status code
		rw := &responseCapture{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the audit entry after the request is processed
		go m.logAudit(r, rw.statusCode, requestBody)
	})
}

// shouldAudit determines if a request should be audited
func (m *AuditMiddleware) shouldAudit(r *http.Request) bool {
	// Skip excluded methods
	for _, method := range m.config.SkipMethods {
		if r.Method == method {
			return false
		}
	}

	// Skip GET requests unless configured to audit reads
	if r.Method == http.MethodGet && !m.config.AuditReads {
		return false
	}

	// Skip excluded paths
	path := r.URL.Path
	for _, skipPath := range m.config.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return false
		}
	}

	return true
}

// logAudit creates an audit log entry for the request
func (m *AuditMiddleware) logAudit(r *http.Request, statusCode int, requestBody []byte) {
	// Skip if no audit service configured
	if m.auditService == nil {
		return
	}

	// Only log successful modifications
	if statusCode < 200 || statusCode >= 300 {
		return
	}

	// Determine action based on HTTP method
	action := m.methodToAction(r.Method)
	if action == "" {
		return
	}

	// Extract entity info from route
	entityType, entityID := m.extractEntityInfo(r)

	// Parse request body for values
	var values interface{}
	if len(requestBody) > 0 {
		var parsed map[string]interface{}
		if json.Unmarshal(requestBody, &parsed) == nil {
			// Remove sensitive fields
			delete(parsed, "password")
			delete(parsed, "secret")
			delete(parsed, "token")
			delete(parsed, "apiKey")
			values = parsed
		}
	}

	entry := service.LogEntry{
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		NewValues:  values,
	}

	if err := m.auditService.Log(r.Context(), r, entry); err != nil {
		m.logger.Warn("failed to create audit log entry",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.Error(err))
	}
}

// methodToAction converts HTTP method to audit action
func (m *AuditMiddleware) methodToAction(method string) domain.AuditAction {
	switch method {
	case http.MethodPost:
		return domain.AuditActionCreate
	case http.MethodPut, http.MethodPatch:
		return domain.AuditActionUpdate
	case http.MethodDelete:
		return domain.AuditActionDelete
	default:
		return ""
	}
}

// extractEntityInfo extracts entity type and ID from the request path
func (m *AuditMiddleware) extractEntityInfo(r *http.Request) (string, *uuid.UUID) {
	// Try to get route pattern and params from chi
	routeCtx := chi.RouteContext(r.Context())
	if routeCtx == nil {
		return m.parseEntityFromPath(r.URL.Path), nil
	}

	// Get entity ID from route params
	var entityID *uuid.UUID
	if idStr := routeCtx.URLParam("id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err == nil {
			entityID = &id
		}
	}

	// Determine entity type from route pattern
	pattern := routeCtx.RoutePattern()
	entityType := m.parseEntityFromPath(pattern)

	return entityType, entityID
}

// parseEntityFromPath extracts entity type from a URL path
func (m *AuditMiddleware) parseEntityFromPath(path string) string {
	// Map of path segments to entity types
	entityMap := map[string]string{
		"customers":   "Customer",
		"contacts":    "Contact",
		"projects":    "Project",
		"offers":      "Offer",
		"files":       "File",
		"users":       "User",
		"roles":       "UserRole",
		"permissions": "UserPermission",
		"deals":       "Deal",
		"activities":  "Activity",
	}

	// Split path and find entity type
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for _, part := range parts {
		if entityType, ok := entityMap[part]; ok {
			return entityType
		}
	}

	return "Unknown"
}

// responseCapture wraps ResponseWriter to capture the status code
type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseCapture) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// GetRequestBody retrieves the stored request body from context
func GetRequestBody(ctx context.Context) []byte {
	if body, ok := ctx.Value(auditRequestBodyKey).([]byte); ok {
		return body
	}
	return nil
}
