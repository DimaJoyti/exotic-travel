package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Logger provides audit logging functionality
type Logger struct {
	repository AuditRepository
	tracer     trace.Tracer
}

// AuditRepository interface for audit log persistence
type AuditRepository interface {
	CreateAuditEvent(ctx context.Context, event *models.AuditEvent) error
	GetAuditEvents(ctx context.Context, filters AuditFilters) ([]models.AuditEvent, error)
	GetAuditEventsByUser(ctx context.Context, userID int, limit int) ([]models.AuditEvent, error)
	GetAuditEventsByAction(ctx context.Context, action string, limit int) ([]models.AuditEvent, error)
	DeleteOldAuditEvents(ctx context.Context, olderThan time.Time) error
}

// AuditFilters represents filters for audit event queries
type AuditFilters struct {
	UserID    *int       `json:"user_id,omitempty"`
	Action    string     `json:"action,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// NewLogger creates a new audit logger
func NewLogger(repository AuditRepository) *Logger {
	return &Logger{
		repository: repository,
		tracer:     otel.Tracer("audit.logger"),
	}
}

// LogAuthEvent logs authentication-related events
func (l *Logger) LogAuthEvent(ctx context.Context, event *models.AuditEvent) error {
	ctx, span := l.tracer.Start(ctx, "audit_logger.log_auth_event")
	defer span.End()

	event.Category = models.AuditCategoryAuth
	event.Severity = l.getEventSeverity(event.Action)
	
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	span.SetAttributes(
		attribute.String("audit.action", event.Action),
		attribute.String("audit.category", string(event.Category)),
		attribute.String("audit.severity", string(event.Severity)),
		attribute.Int("audit.user_id", event.UserID),
	)

	return l.repository.CreateAuditEvent(ctx, event)
}

// LogMarketingEvent logs marketing-related events
func (l *Logger) LogMarketingEvent(ctx context.Context, userID int, action string, resourceType string, resourceID int, details map[string]interface{}, ipAddress, userAgent string) error {
	ctx, span := l.tracer.Start(ctx, "audit_logger.log_marketing_event")
	defer span.End()

	event := &models.AuditEvent{
		UserID:       userID,
		Action:       action,
		Category:     models.AuditCategoryMarketing,
		Severity:     l.getEventSeverity(action),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Timestamp:    time.Now(),
	}

	span.SetAttributes(
		attribute.String("audit.action", action),
		attribute.String("audit.resource_type", resourceType),
		attribute.Int("audit.resource_id", resourceID),
		attribute.Int("audit.user_id", userID),
	)

	return l.repository.CreateAuditEvent(ctx, event)
}

// LogSystemEvent logs system-related events
func (l *Logger) LogSystemEvent(ctx context.Context, action string, details map[string]interface{}) error {
	ctx, span := l.tracer.Start(ctx, "audit_logger.log_system_event")
	defer span.End()

	event := &models.AuditEvent{
		Action:    action,
		Category:  models.AuditCategorySystem,
		Severity:  l.getEventSeverity(action),
		Details:   details,
		Timestamp: time.Now(),
	}

	span.SetAttributes(
		attribute.String("audit.action", action),
		attribute.String("audit.category", string(event.Category)),
	)

	return l.repository.CreateAuditEvent(ctx, event)
}

// LogDataEvent logs data access and modification events
func (l *Logger) LogDataEvent(ctx context.Context, userID int, action string, tableName string, recordID int, oldData, newData interface{}, ipAddress string) error {
	ctx, span := l.tracer.Start(ctx, "audit_logger.log_data_event")
	defer span.End()

	details := map[string]interface{}{
		"table_name": tableName,
		"record_id":  recordID,
	}

	if oldData != nil {
		details["old_data"] = oldData
	}
	if newData != nil {
		details["new_data"] = newData
	}

	event := &models.AuditEvent{
		UserID:       userID,
		Action:       action,
		Category:     models.AuditCategoryData,
		Severity:     l.getEventSeverity(action),
		ResourceType: tableName,
		ResourceID:   recordID,
		Details:      details,
		IPAddress:    ipAddress,
		Timestamp:    time.Now(),
	}

	span.SetAttributes(
		attribute.String("audit.action", action),
		attribute.String("audit.table_name", tableName),
		attribute.Int("audit.record_id", recordID),
		attribute.Int("audit.user_id", userID),
	)

	return l.repository.CreateAuditEvent(ctx, event)
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(ctx context.Context, userID int, action string, severity models.AuditSeverity, details map[string]interface{}, ipAddress, userAgent string) error {
	ctx, span := l.tracer.Start(ctx, "audit_logger.log_security_event")
	defer span.End()

	event := &models.AuditEvent{
		UserID:    userID,
		Action:    action,
		Category:  models.AuditCategorySecurity,
		Severity:  severity,
		Details:   details,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: time.Now(),
	}

	span.SetAttributes(
		attribute.String("audit.action", action),
		attribute.String("audit.severity", string(severity)),
		attribute.Int("audit.user_id", userID),
	)

	return l.repository.CreateAuditEvent(ctx, event)
}

// GetAuditTrail retrieves audit events with filters
func (l *Logger) GetAuditTrail(ctx context.Context, filters AuditFilters) ([]models.AuditEvent, error) {
	ctx, span := l.tracer.Start(ctx, "audit_logger.get_audit_trail")
	defer span.End()

	if filters.Limit == 0 {
		filters.Limit = 100 // Default limit
	}

	events, err := l.repository.GetAuditEvents(ctx, filters)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get audit events: %w", err)
	}

	span.SetAttributes(
		attribute.Int("audit.events_count", len(events)),
		attribute.Int("audit.limit", filters.Limit),
	)

	return events, nil
}

// GetUserAuditTrail retrieves audit events for a specific user
func (l *Logger) GetUserAuditTrail(ctx context.Context, userID int, limit int) ([]models.AuditEvent, error) {
	ctx, span := l.tracer.Start(ctx, "audit_logger.get_user_audit_trail")
	defer span.End()

	span.SetAttributes(
		attribute.Int("audit.user_id", userID),
		attribute.Int("audit.limit", limit),
	)

	events, err := l.repository.GetAuditEventsByUser(ctx, userID, limit)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user audit events: %w", err)
	}

	return events, nil
}

// CleanupOldEvents removes old audit events
func (l *Logger) CleanupOldEvents(ctx context.Context, retentionDays int) error {
	ctx, span := l.tracer.Start(ctx, "audit_logger.cleanup_old_events")
	defer span.End()

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	span.SetAttributes(
		attribute.Int("audit.retention_days", retentionDays),
		attribute.String("audit.cutoff_date", cutoffDate.Format(time.RFC3339)),
	)

	err := l.repository.DeleteOldAuditEvents(ctx, cutoffDate)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to cleanup old audit events: %w", err)
	}

	return nil
}

// Helper methods

func (l *Logger) getEventSeverity(action string) models.AuditSeverity {
	switch action {
	case "login_failed", "unauthorized_access", "permission_denied", "token_expired":
		return models.AuditSeverityHigh
	case "login_success", "logout", "token_refreshed":
		return models.AuditSeverityMedium
	case "campaign_created", "campaign_updated", "campaign_deleted":
		return models.AuditSeverityMedium
	case "content_generated", "integration_connected", "integration_disconnected":
		return models.AuditSeverityMedium
	case "user_created", "user_updated", "user_deleted", "role_changed":
		return models.AuditSeverityHigh
	case "data_exported", "bulk_operation":
		return models.AuditSeverityMedium
	case "system_startup", "system_shutdown", "configuration_changed":
		return models.AuditSeverityLow
	default:
		return models.AuditSeverityLow
	}
}

// AuditMiddleware provides audit logging middleware
type AuditMiddleware struct {
	logger *Logger
	tracer trace.Tracer
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(logger *Logger) *AuditMiddleware {
	return &AuditMiddleware{
		logger: logger,
		tracer: otel.Tracer("audit.middleware"),
	}
}

// LogRequest logs HTTP requests for audit purposes
func (am *AuditMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := am.tracer.Start(r.Context(), "audit_middleware.log_request")
		defer span.End()

		// Skip logging for health checks and metrics
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Get user from context if available
		userID := 0
		if user := getUserFromContext(r.Context()); user != nil {
			userID = user.ID
		}

		// Log the request
		details := map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"referer":    r.Header.Get("Referer"),
		}

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Process request
		start := time.Now()
		next.ServeHTTP(wrapper, r)
		duration := time.Since(start)

		// Add response details
		details["status_code"] = wrapper.statusCode
		details["duration_ms"] = duration.Milliseconds()

		// Determine action based on method and path
		action := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

		// Log the event
		am.logger.LogSystemEvent(ctx, action, details)

		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.path", r.URL.Path),
			attribute.Int("http.status_code", wrapper.statusCode),
			attribute.Int64("http.duration_ms", duration.Milliseconds()),
			attribute.Int("user.id", userID),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper function to get user from context
func getUserFromContext(ctx context.Context) *models.User {
	// This would use the same context key as defined in auth middleware
	if user, ok := ctx.Value("user").(*models.User); ok {
		return user
	}
	return nil
}

// AuditEventBuilder helps build audit events
type AuditEventBuilder struct {
	event *models.AuditEvent
}

// NewAuditEventBuilder creates a new audit event builder
func NewAuditEventBuilder() *AuditEventBuilder {
	return &AuditEventBuilder{
		event: &models.AuditEvent{
			Timestamp: time.Now(),
			Details:   make(map[string]interface{}),
		},
	}
}

// WithUser sets the user for the audit event
func (b *AuditEventBuilder) WithUser(userID int, userEmail string) *AuditEventBuilder {
	b.event.UserID = userID
	b.event.UserEmail = userEmail
	return b
}

// WithAction sets the action for the audit event
func (b *AuditEventBuilder) WithAction(action string) *AuditEventBuilder {
	b.event.Action = action
	return b
}

// WithCategory sets the category for the audit event
func (b *AuditEventBuilder) WithCategory(category models.AuditCategory) *AuditEventBuilder {
	b.event.Category = category
	return b
}

// WithSeverity sets the severity for the audit event
func (b *AuditEventBuilder) WithSeverity(severity models.AuditSeverity) *AuditEventBuilder {
	b.event.Severity = severity
	return b
}

// WithResource sets the resource information for the audit event
func (b *AuditEventBuilder) WithResource(resourceType string, resourceID int) *AuditEventBuilder {
	b.event.ResourceType = resourceType
	b.event.ResourceID = resourceID
	return b
}

// WithDetails adds details to the audit event
func (b *AuditEventBuilder) WithDetails(key string, value interface{}) *AuditEventBuilder {
	b.event.Details[key] = value
	return b
}

// WithRequest adds request information to the audit event
func (b *AuditEventBuilder) WithRequest(ipAddress, userAgent string) *AuditEventBuilder {
	b.event.IPAddress = ipAddress
	b.event.UserAgent = userAgent
	return b
}

// Build returns the constructed audit event
func (b *AuditEventBuilder) Build() *models.AuditEvent {
	return b.event
}
