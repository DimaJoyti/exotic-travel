package security

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// AuditLogger handles security event logging and monitoring
type AuditLogger struct {
	events   chan AuditEvent
	storage  AuditStorage
	rules    []SecurityRule
	alerts   AlertManager
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
}

// AuditEvent represents a security-related event
type AuditEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Severity    string                 `json:"severity"`
	UserID      *int64                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Details     map[string]interface{} `json:"details"`
	RiskScore   int                    `json:"risk_score"`
	Fingerprint string                 `json:"fingerprint"`
}

// SecurityRule defines conditions for security alerts
type SecurityRule struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	EventTypes  []string      `json:"event_types"`
	Conditions  []Condition   `json:"conditions"`
	Threshold   int           `json:"threshold"`
	TimeWindow  time.Duration `json:"time_window"`
	Severity    string        `json:"severity"`
	Actions     []string      `json:"actions"`
	Enabled     bool          `json:"enabled"`
}

// Condition represents a rule condition
type Condition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// AuditStorage interface for storing audit events
type AuditStorage interface {
	Store(event AuditEvent) error
	Query(filters map[string]interface{}, limit int) ([]AuditEvent, error)
	GetEventsByTimeRange(start, end time.Time) ([]AuditEvent, error)
	GetEventsByUser(userID int64, limit int) ([]AuditEvent, error)
	GetEventsByIP(ipAddress string, limit int) ([]AuditEvent, error)
}

// AlertManager interface for handling security alerts
type AlertManager interface {
	SendAlert(alert SecurityAlert) error
	GetAlerts(filters map[string]interface{}) ([]SecurityAlert, error)
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	RuleID       string                 `json:"rule_id"`
	RuleName     string                 `json:"rule_name"`
	Severity     string                 `json:"severity"`
	Message      string                 `json:"message"`
	Events       []AuditEvent           `json:"events"`
	Details      map[string]interface{} `json:"details"`
	Status       string                 `json:"status"`
	Acknowledged bool                   `json:"acknowledged"`
}

// ThreatDetector analyzes events for security threats
type ThreatDetector struct {
	patterns    []ThreatPattern
	ipTracker   *IPTracker
	userTracker *UserTracker
}

// ThreatPattern defines patterns for threat detection
type ThreatPattern struct {
	Name        string
	Description string
	Pattern     func(event AuditEvent, context *DetectionContext) (bool, int)
	Severity    string
}

// DetectionContext provides context for threat detection
type DetectionContext struct {
	RecentEvents []AuditEvent
	IPHistory    map[string][]AuditEvent
	UserHistory  map[int64][]AuditEvent
}

// IPTracker tracks IP-based security metrics
type IPTracker struct {
	mu           sync.RWMutex
	ipEvents     map[string][]AuditEvent
	ipRiskScores map[string]int
	blacklist    map[string]time.Time
}

// UserTracker tracks user-based security metrics
type UserTracker struct {
	mu              sync.RWMutex
	userEvents      map[int64][]AuditEvent
	userRiskScores  map[int64]int
	suspiciousUsers map[int64]time.Time
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(storage AuditStorage, alerts AlertManager) *AuditLogger {
	return &AuditLogger{
		events:   make(chan AuditEvent, 1000),
		storage:  storage,
		alerts:   alerts,
		rules:    getDefaultSecurityRules(),
		stopChan: make(chan struct{}),
	}
}

// Start begins the audit logging process
func (al *AuditLogger) Start() {
	al.mu.Lock()
	if al.running {
		al.mu.Unlock()
		return
	}
	al.running = true
	al.mu.Unlock()

	go al.processEvents()
	log.Println("Audit logger started")
}

// Stop stops the audit logging process
func (al *AuditLogger) Stop() {
	al.mu.Lock()
	if !al.running {
		al.mu.Unlock()
		return
	}
	al.running = false
	al.mu.Unlock()

	close(al.stopChan)
	log.Println("Audit logger stopped")
}

// LogEvent logs a security event
func (al *AuditLogger) LogEvent(eventType, severity, resource, action, result string, userID *int64, sessionID, ipAddress, userAgent string, details map[string]interface{}) {
	event := AuditEvent{
		ID:          generateEventID(),
		Timestamp:   time.Now(),
		EventType:   eventType,
		Severity:    severity,
		UserID:      userID,
		SessionID:   sessionID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Resource:    resource,
		Action:      action,
		Result:      result,
		Details:     details,
		RiskScore:   calculateRiskScore(eventType, severity, result),
		Fingerprint: generateFingerprint(eventType, resource, action, ipAddress),
	}

	select {
	case al.events <- event:
	default:
		log.Printf("Audit event queue full, dropping event: %s", event.ID)
	}
}

// LogAuthenticationEvent logs authentication-related events
func (al *AuditLogger) LogAuthenticationEvent(action, result string, userID *int64, email, ipAddress, userAgent string, details map[string]interface{}) {
	severity := "INFO"
	if result == "FAILURE" {
		severity = "WARNING"
	}

	if details == nil {
		details = make(map[string]interface{})
	}
	details["email"] = email

	al.LogEvent("AUTHENTICATION", severity, "auth", action, result, userID, "", ipAddress, userAgent, details)
}

// LogAuthorizationEvent logs authorization-related events
func (al *AuditLogger) LogAuthorizationEvent(action, resource, result string, userID *int64, sessionID, ipAddress, userAgent string, details map[string]interface{}) {
	severity := "INFO"
	if result == "DENIED" {
		severity = "WARNING"
	}

	al.LogEvent("AUTHORIZATION", severity, resource, action, result, userID, sessionID, ipAddress, userAgent, details)
}

// LogDataAccessEvent logs data access events
func (al *AuditLogger) LogDataAccessEvent(action, resource string, userID *int64, sessionID, ipAddress, userAgent string, details map[string]interface{}) {
	al.LogEvent("DATA_ACCESS", "INFO", resource, action, "SUCCESS", userID, sessionID, ipAddress, userAgent, details)
}

// LogSecurityEvent logs general security events
func (al *AuditLogger) LogSecurityEvent(eventType, action, result string, userID *int64, sessionID, ipAddress, userAgent string, details map[string]interface{}) {
	severity := "WARNING"
	if result == "BLOCKED" || result == "FAILURE" {
		severity = "ERROR"
	}

	al.LogEvent(eventType, severity, "security", action, result, userID, sessionID, ipAddress, userAgent, details)
}

// processEvents processes audit events in the background
func (al *AuditLogger) processEvents() {
	detector := NewThreatDetector()

	for {
		select {
		case event := <-al.events:
			// Store the event
			if err := al.storage.Store(event); err != nil {
				log.Printf("Failed to store audit event: %v", err)
			}

			// Analyze for threats
			if threats := detector.AnalyzeEvent(event); len(threats) > 0 {
				for _, threat := range threats {
					al.handleThreat(threat, event)
				}
			}

			// Check security rules
			al.checkSecurityRules(event)

		case <-al.stopChan:
			return
		}
	}
}

// checkSecurityRules evaluates security rules against events
func (al *AuditLogger) checkSecurityRules(event AuditEvent) {
	for _, rule := range al.rules {
		if !rule.Enabled {
			continue
		}

		if al.eventMatchesRule(event, rule) {
			// Get recent events for threshold checking
			recentEvents, err := al.storage.GetEventsByTimeRange(
				time.Now().Add(-rule.TimeWindow),
				time.Now(),
			)
			if err != nil {
				log.Printf("Failed to get recent events for rule %s: %v", rule.ID, err)
				continue
			}

			// Count matching events
			matchCount := 0
			for _, recentEvent := range recentEvents {
				if al.eventMatchesRule(recentEvent, rule) {
					matchCount++
				}
			}

			// Trigger alert if threshold exceeded
			if matchCount >= rule.Threshold {
				alert := SecurityAlert{
					ID:        generateAlertID(),
					Timestamp: time.Now(),
					RuleID:    rule.ID,
					RuleName:  rule.Name,
					Severity:  rule.Severity,
					Message:   fmt.Sprintf("Security rule '%s' triggered: %d events in %v", rule.Name, matchCount, rule.TimeWindow),
					Events:    recentEvents[:min(matchCount, 10)], // Include up to 10 events
					Status:    "OPEN",
				}

				if err := al.alerts.SendAlert(alert); err != nil {
					log.Printf("Failed to send security alert: %v", err)
				}
			}
		}
	}
}

// eventMatchesRule checks if an event matches a security rule
func (al *AuditLogger) eventMatchesRule(event AuditEvent, rule SecurityRule) bool {
	// Check event type
	eventTypeMatch := false
	for _, eventType := range rule.EventTypes {
		if event.EventType == eventType {
			eventTypeMatch = true
			break
		}
	}
	if !eventTypeMatch {
		return false
	}

	// Check conditions
	for _, condition := range rule.Conditions {
		if !al.evaluateCondition(event, condition) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition against an event
func (al *AuditLogger) evaluateCondition(event AuditEvent, condition Condition) bool {
	var fieldValue interface{}

	switch condition.Field {
	case "severity":
		fieldValue = event.Severity
	case "result":
		fieldValue = event.Result
	case "ip_address":
		fieldValue = event.IPAddress
	case "risk_score":
		fieldValue = event.RiskScore
	default:
		if val, exists := event.Details[condition.Field]; exists {
			fieldValue = val
		} else {
			return false
		}
	}

	return evaluateOperator(fieldValue, condition.Operator, condition.Value)
}

// handleThreat handles detected security threats
func (al *AuditLogger) handleThreat(threat ThreatPattern, event AuditEvent) {
	log.Printf("Security threat detected: %s for event %s", threat.Name, event.ID)

	// Create security alert
	alert := SecurityAlert{
		ID:        generateAlertID(),
		Timestamp: time.Now(),
		RuleID:    "THREAT_" + strings.ToUpper(strings.ReplaceAll(threat.Name, " ", "_")),
		RuleName:  threat.Name,
		Severity:  threat.Severity,
		Message:   fmt.Sprintf("Threat detected: %s", threat.Description),
		Events:    []AuditEvent{event},
		Status:    "OPEN",
		Details: map[string]interface{}{
			"threat_pattern": threat.Name,
			"detection_time": time.Now(),
		},
	}

	if err := al.alerts.SendAlert(alert); err != nil {
		log.Printf("Failed to send threat alert: %v", err)
	}
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector() *ThreatDetector {
	return &ThreatDetector{
		patterns:    getDefaultThreatPatterns(),
		ipTracker:   NewIPTracker(),
		userTracker: NewUserTracker(),
	}
}

// AnalyzeEvent analyzes an event for security threats
func (td *ThreatDetector) AnalyzeEvent(event AuditEvent) []ThreatPattern {
	var threats []ThreatPattern

	context := &DetectionContext{
		RecentEvents: []AuditEvent{event}, // In a real implementation, this would include recent events
		IPHistory:    td.ipTracker.GetIPHistory(event.IPAddress),
		UserHistory:  td.userTracker.GetUserHistory(event.UserID),
	}

	for _, pattern := range td.patterns {
		if matches, riskScore := pattern.Pattern(event, context); matches {
			threats = append(threats, pattern)

			// Update risk scores
			td.ipTracker.UpdateRiskScore(event.IPAddress, riskScore)
			if event.UserID != nil {
				td.userTracker.UpdateRiskScore(*event.UserID, riskScore)
			}
		}
	}

	return threats
}

// Helper functions

func generateEventID() string {
	return fmt.Sprintf("evt_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

func generateAlertID() string {
	return fmt.Sprintf("alt_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

func generateFingerprint(eventType, resource, action, ipAddress string) string {
	data := fmt.Sprintf("%s:%s:%s:%s", eventType, resource, action, ipAddress)
	return HashSHA256String(data)[:16]
}

func calculateRiskScore(eventType, severity, result string) int {
	score := 0

	// Base score by event type
	switch eventType {
	case "AUTHENTICATION":
		score = 10
	case "AUTHORIZATION":
		score = 15
	case "DATA_ACCESS":
		score = 20
	case "SECURITY_VIOLATION":
		score = 50
	default:
		score = 5
	}

	// Severity multiplier
	switch severity {
	case "ERROR":
		score *= 3
	case "WARNING":
		score *= 2
	case "INFO":
		score *= 1
	}

	// Result modifier
	if result == "FAILURE" || result == "DENIED" || result == "BLOCKED" {
		score += 20
	}

	return score
}

func evaluateOperator(fieldValue interface{}, operator string, conditionValue interface{}) bool {
	switch operator {
	case "equals":
		return fieldValue == conditionValue
	case "not_equals":
		return fieldValue != conditionValue
	case "contains":
		if str, ok := fieldValue.(string); ok {
			if substr, ok := conditionValue.(string); ok {
				return strings.Contains(str, substr)
			}
		}
	case "greater_than":
		if num, ok := fieldValue.(int); ok {
			if threshold, ok := conditionValue.(int); ok {
				return num > threshold
			}
		}
	case "less_than":
		if num, ok := fieldValue.(int); ok {
			if threshold, ok := conditionValue.(int); ok {
				return num < threshold
			}
		}
	}
	return false
}

func generateRandomString(length int) string {
	token, _ := GenerateSecureToken(length)
	return token[:length]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewIPTracker creates a new IP tracker
func NewIPTracker() *IPTracker {
	return &IPTracker{
		ipEvents:     make(map[string][]AuditEvent),
		ipRiskScores: make(map[string]int),
		blacklist:    make(map[string]time.Time),
	}
}

// GetIPHistory returns recent events for an IP address
func (it *IPTracker) GetIPHistory(ipAddress string) map[string][]AuditEvent {
	it.mu.RLock()
	defer it.mu.RUnlock()

	result := make(map[string][]AuditEvent)
	if events, exists := it.ipEvents[ipAddress]; exists {
		result[ipAddress] = events
	}
	return result
}

// UpdateRiskScore updates the risk score for an IP address
func (it *IPTracker) UpdateRiskScore(ipAddress string, additionalRisk int) {
	it.mu.Lock()
	defer it.mu.Unlock()

	it.ipRiskScores[ipAddress] += additionalRisk
}

// NewUserTracker creates a new user tracker
func NewUserTracker() *UserTracker {
	return &UserTracker{
		userEvents:      make(map[int64][]AuditEvent),
		userRiskScores:  make(map[int64]int),
		suspiciousUsers: make(map[int64]time.Time),
	}
}

// GetUserHistory returns recent events for a user
func (ut *UserTracker) GetUserHistory(userID *int64) map[int64][]AuditEvent {
	if userID == nil {
		return make(map[int64][]AuditEvent)
	}

	ut.mu.RLock()
	defer ut.mu.RUnlock()

	result := make(map[int64][]AuditEvent)
	if events, exists := ut.userEvents[*userID]; exists {
		result[*userID] = events
	}
	return result
}

// UpdateRiskScore updates the risk score for a user
func (ut *UserTracker) UpdateRiskScore(userID int64, additionalRisk int) {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	ut.userRiskScores[userID] += additionalRisk
}

// getDefaultSecurityRules returns default security rules
func getDefaultSecurityRules() []SecurityRule {
	return []SecurityRule{
		{
			ID:          "FAILED_LOGIN_ATTEMPTS",
			Name:        "Multiple Failed Login Attempts",
			Description: "Detects multiple failed login attempts from the same IP",
			EventTypes:  []string{"AUTHENTICATION"},
			Conditions: []Condition{
				{Field: "result", Operator: "equals", Value: "FAILURE"},
				{Field: "action", Operator: "equals", Value: "LOGIN"},
			},
			Threshold:  5,
			TimeWindow: 15 * time.Minute,
			Severity:   "WARNING",
			Actions:    []string{"ALERT", "RATE_LIMIT"},
			Enabled:    true,
		},
		{
			ID:          "PRIVILEGE_ESCALATION",
			Name:        "Privilege Escalation Attempt",
			Description: "Detects attempts to access unauthorized resources",
			EventTypes:  []string{"AUTHORIZATION"},
			Conditions: []Condition{
				{Field: "result", Operator: "equals", Value: "DENIED"},
				{Field: "risk_score", Operator: "greater_than", Value: 30},
			},
			Threshold:  3,
			TimeWindow: 10 * time.Minute,
			Severity:   "ERROR",
			Actions:    []string{"ALERT", "BLOCK_IP"},
			Enabled:    true,
		},
	}
}

// getDefaultThreatPatterns returns default threat detection patterns
func getDefaultThreatPatterns() []ThreatPattern {
	return []ThreatPattern{
		{
			Name:        "Brute Force Attack",
			Description: "Multiple failed authentication attempts",
			Severity:    "HIGH",
			Pattern: func(event AuditEvent, context *DetectionContext) (bool, int) {
				if event.EventType == "AUTHENTICATION" && event.Result == "FAILURE" {
					// Check for multiple failures from same IP
					failureCount := 0
					for _, recentEvent := range context.RecentEvents {
						if recentEvent.IPAddress == event.IPAddress &&
							recentEvent.EventType == "AUTHENTICATION" &&
							recentEvent.Result == "FAILURE" {
							failureCount++
						}
					}
					return failureCount >= 5, 50
				}
				return false, 0
			},
		},
		{
			Name:        "SQL Injection Attempt",
			Description: "Potential SQL injection in request parameters",
			Severity:    "CRITICAL",
			Pattern: func(event AuditEvent, context *DetectionContext) (bool, int) {
				if details, ok := event.Details["request_params"].(string); ok {
					sqlPatterns := []string{"union", "select", "drop", "insert", "update", "delete", "'", "--", "/*"}
					lowerDetails := strings.ToLower(details)
					for _, pattern := range sqlPatterns {
						if strings.Contains(lowerDetails, pattern) {
							return true, 100
						}
					}
				}
				return false, 0
			},
		},
	}
}
