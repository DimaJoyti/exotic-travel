package security

import (
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Validator provides comprehensive input validation and sanitization
type Validator struct {
	maxStringLength int
	allowedTags     map[string]bool
	sqlPatterns     []*regexp.Regexp
	xssPatterns     []*regexp.Regexp
	pathPatterns    []*regexp.Regexp
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d validation errors: %s", len(e), e[0].Message)
}

// NewValidator creates a new validator with security rules
func NewValidator() *Validator {
	return &Validator{
		maxStringLength: 10000, // 10KB max string length
		allowedTags: map[string]bool{
			"b": true, "i": true, "u": true, "strong": true, "em": true,
			"p": true, "br": true, "ul": true, "ol": true, "li": true,
		},
		sqlPatterns: compileSQLPatterns(),
		xssPatterns: compileXSSPatterns(),
		pathPatterns: compilePathPatterns(),
	}
}

// ValidateEmail validates email format and security
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return ValidationError{Field: "email", Message: "Email is required", Code: "REQUIRED"}
	}

	if len(email) > 254 {
		return ValidationError{Field: "email", Message: "Email is too long", Code: "TOO_LONG"}
	}

	// Parse email
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ValidationError{Field: "email", Message: "Invalid email format", Code: "INVALID_FORMAT"}
	}

	// Additional security checks
	if strings.Contains(addr.Address, "..") {
		return ValidationError{Field: "email", Message: "Invalid email format", Code: "INVALID_FORMAT"}
	}

	// Check for suspicious patterns
	if v.containsSuspiciousPatterns(addr.Address) {
		return ValidationError{Field: "email", Message: "Email contains suspicious content", Code: "SUSPICIOUS_CONTENT"}
	}

	return nil
}

// ValidatePassword validates password strength and security
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return ValidationError{Field: "password", Message: "Password is required", Code: "REQUIRED"}
	}

	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "Password must be at least 8 characters", Code: "TOO_SHORT"}
	}

	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "Password is too long", Code: "TOO_LONG"}
	}

	// Check character requirements
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ValidationError{Field: "password", Message: "Password must contain uppercase letters", Code: "MISSING_UPPERCASE"}
	}
	if !hasLower {
		return ValidationError{Field: "password", Message: "Password must contain lowercase letters", Code: "MISSING_LOWERCASE"}
	}
	if !hasDigit {
		return ValidationError{Field: "password", Message: "Password must contain digits", Code: "MISSING_DIGIT"}
	}
	if !hasSpecial {
		return ValidationError{Field: "password", Message: "Password must contain special characters", Code: "MISSING_SPECIAL"}
	}

	// Check for common weak patterns
	if v.isWeakPassword(password) {
		return ValidationError{Field: "password", Message: "Password is too weak", Code: "WEAK_PASSWORD"}
	}

	return nil
}

// ValidateString validates and sanitizes general string input
func (v *Validator) ValidateString(field, value string, minLength, maxLength int, required bool) error {
	if required && value == "" {
		return ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field), Code: "REQUIRED"}
	}

	if value == "" && !required {
		return nil
	}

	if !utf8.ValidString(value) {
		return ValidationError{Field: field, Message: "Invalid UTF-8 encoding", Code: "INVALID_ENCODING"}
	}

	if len(value) < minLength {
		return ValidationError{Field: field, Message: fmt.Sprintf("%s must be at least %d characters", field, minLength), Code: "TOO_SHORT"}
	}

	if maxLength > 0 && len(value) > maxLength {
		return ValidationError{Field: field, Message: fmt.Sprintf("%s must be at most %d characters", field, maxLength), Code: "TOO_LONG"}
	}

	if len(value) > v.maxStringLength {
		return ValidationError{Field: field, Message: "String is too long", Code: "TOO_LONG"}
	}

	// Check for malicious content
	if v.containsSQLInjection(value) {
		return ValidationError{Field: field, Message: "Input contains suspicious SQL patterns", Code: "SQL_INJECTION"}
	}

	if v.containsXSS(value) {
		return ValidationError{Field: field, Message: "Input contains suspicious script content", Code: "XSS_ATTEMPT"}
	}

	if v.containsPathTraversal(value) {
		return ValidationError{Field: field, Message: "Input contains path traversal patterns", Code: "PATH_TRAVERSAL"}
	}

	return nil
}

// SanitizeString sanitizes string input for safe storage and display
func (v *Validator) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Normalize whitespace
	input = strings.TrimSpace(input)
	input = regexp.MustCompile(`\s+`).ReplaceAllString(input, " ")
	
	// HTML escape
	input = html.EscapeString(input)
	
	return input
}

// SanitizeHTML sanitizes HTML content allowing only safe tags
func (v *Validator) SanitizeHTML(input string) string {
	// Remove script tags and their content
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	input = scriptRegex.ReplaceAllString(input, "")
	
	// Remove dangerous attributes
	attrRegex := regexp.MustCompile(`(?i)\s+(on\w+|javascript:|data:|vbscript:)[^>]*`)
	input = attrRegex.ReplaceAllString(input, "")
	
	// Remove dangerous tags
	dangerousTagRegex := regexp.MustCompile(`(?i)<(script|iframe|object|embed|form|input|meta|link)[^>]*>.*?</\1>`)
	input = dangerousTagRegex.ReplaceAllString(input, "")
	
	return input
}

// ValidateID validates numeric IDs
func (v *Validator) ValidateID(field string, id int64) error {
	if id <= 0 {
		return ValidationError{Field: field, Message: fmt.Sprintf("%s must be a positive integer", field), Code: "INVALID_ID"}
	}
	return nil
}

// ValidateEnum validates enum values
func (v *Validator) ValidateEnum(field, value string, allowedValues []string) error {
	if value == "" {
		return ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field), Code: "REQUIRED"}
	}

	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return ValidationError{Field: field, Message: fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowedValues, ", ")), Code: "INVALID_ENUM"}
}

// ValidateURL validates URL format and security
func (v *Validator) ValidateURL(field, url string) error {
	if url == "" {
		return ValidationError{Field: field, Message: fmt.Sprintf("%s is required", field), Code: "REQUIRED"}
	}

	// Basic URL validation
	if !regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`).MatchString(url) {
		return ValidationError{Field: field, Message: "Invalid URL format", Code: "INVALID_FORMAT"}
	}

	// Security checks
	if strings.Contains(strings.ToLower(url), "javascript:") ||
		strings.Contains(strings.ToLower(url), "data:") ||
		strings.Contains(strings.ToLower(url), "vbscript:") {
		return ValidationError{Field: field, Message: "URL contains suspicious protocol", Code: "SUSPICIOUS_PROTOCOL"}
	}

	return nil
}

// Helper methods

func (v *Validator) containsSQLInjection(input string) bool {
	lowerInput := strings.ToLower(input)
	for _, pattern := range v.sqlPatterns {
		if pattern.MatchString(lowerInput) {
			return true
		}
	}
	return false
}

func (v *Validator) containsXSS(input string) bool {
	lowerInput := strings.ToLower(input)
	for _, pattern := range v.xssPatterns {
		if pattern.MatchString(lowerInput) {
			return true
		}
	}
	return false
}

func (v *Validator) containsPathTraversal(input string) bool {
	for _, pattern := range v.pathPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

func (v *Validator) containsSuspiciousPatterns(input string) bool {
	return v.containsSQLInjection(input) || v.containsXSS(input) || v.containsPathTraversal(input)
}

func (v *Validator) isWeakPassword(password string) bool {
	weakPatterns := []string{
		"password", "123456", "qwerty", "abc123", "admin", "letmein",
		"welcome", "monkey", "dragon", "master", "shadow", "superman",
	}

	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPatterns {
		if strings.Contains(lowerPassword, weak) {
			return true
		}
	}

	// Check for repeated characters
	if regexp.MustCompile(`(.)\1{3,}`).MatchString(password) {
		return true
	}

	// Check for sequential characters
	if regexp.MustCompile(`(abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz|012|123|234|345|456|567|678|789)`).MatchString(strings.ToLower(password)) {
		return true
	}

	return false
}

// Pattern compilation functions

func compileSQLPatterns() []*regexp.Regexp {
	patterns := []string{
		`\b(union|select|insert|update|delete|drop|create|alter|exec|execute)\b`,
		`\b(or|and)\s+\d+\s*=\s*\d+`,
		`\b(or|and)\s+['"]\w+['"]?\s*=\s*['"]\w+['"]?`,
		`['"]\s*(or|and)\s+['"]\w+['"]?\s*=\s*['"]\w+['"]?`,
		`\b(sleep|benchmark|waitfor)\s*\(`,
		`\b(load_file|into\s+outfile|into\s+dumpfile)\b`,
		`\b(information_schema|mysql|sys|performance_schema)\b`,
		`['"]\s*;\s*(drop|delete|update|insert)`,
		`\b(0x[0-9a-f]+|char\s*\(|ascii\s*\(|hex\s*\()`,
	}

	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, regex)
		}
	}
	return compiled
}

func compileXSSPatterns() []*regexp.Regexp {
	patterns := []string{
		`<\s*script[^>]*>`,
		`<\s*iframe[^>]*>`,
		`<\s*object[^>]*>`,
		`<\s*embed[^>]*>`,
		`<\s*link[^>]*>`,
		`<\s*meta[^>]*>`,
		`javascript\s*:`,
		`vbscript\s*:`,
		`data\s*:.*base64`,
		`on\w+\s*=`,
		`expression\s*\(`,
		`@import`,
		`<\s*style[^>]*>.*expression`,
	}

	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(`(?i)`+pattern); err == nil {
			compiled = append(compiled, regex)
		}
	}
	return compiled
}

func compilePathPatterns() []*regexp.Regexp {
	patterns := []string{
		`\.\.[\\/]`,
		`[\\/]\.\.[\\/]`,
		`^\.\.[\\/]`,
		`[\\/]\.\.[^\\/.]*$`,
		`%2e%2e[\\/]`,
		`%2e%2e%2f`,
		`%2e%2e%5c`,
		`\.\.%2f`,
		`\.\.%5c`,
	}

	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(`(?i)`+pattern); err == nil {
			compiled = append(compiled, regex)
		}
	}
	return compiled
}
