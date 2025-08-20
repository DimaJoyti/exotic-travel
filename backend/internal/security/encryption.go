package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

// Encryptor provides data encryption and decryption functionality
type Encryptor struct {
	key []byte
	gcm cipher.AEAD
}

// NewEncryptor creates a new encryptor with the given key
func NewEncryptor(key string) (*Encryptor, error) {
	// Derive a 32-byte key from the provided key using SHA-256
	hash := sha256.Sum256([]byte(key))
	derivedKey := hash[:]

	// Create AES cipher
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Encryptor{
		key: derivedKey,
		gcm: gcm,
	}, nil
}

// Encrypt encrypts plaintext and returns base64 encoded ciphertext
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Generate a random nonce
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64 encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64 encoded ciphertext and returns plaintext
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length
	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptBytes encrypts byte data and returns encrypted bytes
func (e *Encryptor) EncryptBytes(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Generate a random nonce
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := e.gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptBytes decrypts encrypted bytes and returns original data
func (e *Encryptor) DecryptBytes(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	// Check minimum length
	nonceSize := e.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext_bytes := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// PasswordHasher provides secure password hashing
type PasswordHasher struct {
	saltLength int
	keyLength  int
	iterations int
}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		saltLength: 32,
		keyLength:  64,
		iterations: 100000, // PBKDF2 iterations
	}
}

// HashPassword hashes a password using PBKDF2
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, ph.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with PBKDF2
	hash := pbkdf2.Key([]byte(password), salt, ph.iterations, ph.keyLength, sha256.New)

	// Combine salt and hash
	combined := append(salt, hash...)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword verifies a password against its hash
func (ph *PasswordHasher) VerifyPassword(password, hashedPassword string) error {
	// Decode the hash
	combined, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	// Check minimum length
	if len(combined) < ph.saltLength+ph.keyLength {
		return fmt.Errorf("invalid hash format")
	}

	// Extract salt and hash
	salt := combined[:ph.saltLength]
	hash := combined[ph.saltLength:]

	// Hash the provided password with the same salt
	newHash := pbkdf2.Key([]byte(password), salt, ph.iterations, ph.keyLength, sha256.New)

	// Compare hashes
	if !constantTimeCompare(hash, newHash) {
		return fmt.Errorf("password verification failed")
	}

	return nil
}

// ScryptHasher provides scrypt-based password hashing (alternative to PBKDF2)
type ScryptHasher struct {
	saltLength int
	keyLength  int
	n          int // CPU/memory cost parameter
	r          int // block size parameter
	p          int // parallelization parameter
}

// NewScryptHasher creates a new scrypt hasher
func NewScryptHasher() *ScryptHasher {
	return &ScryptHasher{
		saltLength: 32,
		keyLength:  64,
		n:          32768, // 2^15
		r:          8,
		p:          1,
	}
}

// HashPassword hashes a password using scrypt
func (sh *ScryptHasher) HashPassword(password string) (string, error) {
	// Generate random salt
	salt := make([]byte, sh.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with scrypt
	hash, err := scrypt.Key([]byte(password), salt, sh.n, sh.r, sh.p, sh.keyLength)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Combine salt and hash
	combined := append(salt, hash...)

	// Return base64 encoded result
	return base64.StdEncoding.EncodeToString(combined), nil
}

// VerifyPassword verifies a password against its scrypt hash
func (sh *ScryptHasher) VerifyPassword(password, hashedPassword string) error {
	// Decode the hash
	combined, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to decode hash: %w", err)
	}

	// Check minimum length
	if len(combined) < sh.saltLength+sh.keyLength {
		return fmt.Errorf("invalid hash format")
	}

	// Extract salt and hash
	salt := combined[:sh.saltLength]
	hash := combined[sh.saltLength:]

	// Hash the provided password with the same salt
	newHash, err := scrypt.Key([]byte(password), salt, sh.n, sh.r, sh.p, sh.keyLength)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Compare hashes
	if !constantTimeCompare(hash, newHash) {
		return fmt.Errorf("password verification failed")
	}

	return nil
}

// TokenGenerator generates cryptographically secure tokens
type TokenGenerator struct{}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{}
}

// GenerateToken generates a random token of specified length
func (tg *TokenGenerator) GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateAPIKey generates a secure API key
func (tg *TokenGenerator) GenerateAPIKey() (string, error) {
	return tg.GenerateToken(32) // 256-bit key
}

// GenerateSessionID generates a secure session ID
func (tg *TokenGenerator) GenerateSessionID() (string, error) {
	return tg.GenerateToken(24) // 192-bit session ID
}

// GenerateCSRFToken generates a CSRF token
func (tg *TokenGenerator) GenerateCSRFToken() (string, error) {
	return tg.GenerateToken(16) // 128-bit CSRF token
}

// DataMasker provides data masking functionality for sensitive information
type DataMasker struct{}

// NewDataMasker creates a new data masker
func NewDataMasker() *DataMasker {
	return &DataMasker{}
}

// MaskEmail masks an email address
func (dm *DataMasker) MaskEmail(email string) string {
	if email == "" {
		return ""
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email // Invalid email format
	}

	username := parts[0]
	domain := parts[1]

	// Mask username (show first and last character)
	if len(username) <= 2 {
		return "*@" + domain
	}

	maskedUsername := string(username[0]) + strings.Repeat("*", len(username)-2) + string(username[len(username)-1])
	return maskedUsername + "@" + domain
}

// MaskCreditCard masks a credit card number
func (dm *DataMasker) MaskCreditCard(cardNumber string) string {
	if len(cardNumber) < 4 {
		return strings.Repeat("*", len(cardNumber))
	}

	// Show only last 4 digits
	return strings.Repeat("*", len(cardNumber)-4) + cardNumber[len(cardNumber)-4:]
}

// MaskPhone masks a phone number
func (dm *DataMasker) MaskPhone(phone string) string {
	if len(phone) < 4 {
		return strings.Repeat("*", len(phone))
	}

	// Show only last 4 digits
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

// MaskString masks a string showing only first and last characters
func (dm *DataMasker) MaskString(str string, showChars int) string {
	if len(str) <= showChars*2 {
		return strings.Repeat("*", len(str))
	}

	prefix := str[:showChars]
	suffix := str[len(str)-showChars:]
	middle := strings.Repeat("*", len(str)-showChars*2)

	return prefix + middle + suffix
}

// Helper functions

// constantTimeCompare performs constant-time comparison of two byte slices
func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// SecureHeaders provides security-related HTTP headers
type SecureHeaders struct{}

// NewSecureHeaders creates a new secure headers provider
func NewSecureHeaders() *SecureHeaders {
	return &SecureHeaders{}
}

// GetSecurityHeaders returns a map of security headers
func (sh *SecureHeaders) GetSecurityHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:; frame-ancestors 'none';",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
	}
}

// GetCORSHeaders returns CORS headers for API endpoints
func (sh *SecureHeaders) GetCORSHeaders(allowedOrigins []string) map[string]string {
	origin := "*"
	if len(allowedOrigins) > 0 {
		origin = strings.Join(allowedOrigins, ", ")
	}

	return map[string]string{
		"Access-Control-Allow-Origin":      origin,
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, Authorization, X-Requested-With, X-CSRF-Token",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Max-Age":           "86400",
	}
}
