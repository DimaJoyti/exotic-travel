package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/scrypt"
)

// CryptoManager handles encryption, decryption, and hashing operations
type CryptoManager struct {
	encryptionKey []byte
	gcm           cipher.AEAD
}

// PasswordHasher handles secure password hashing
type PasswordHasher struct {
	algorithm string
	params    HashParams
}

// HashParams contains parameters for password hashing
type HashParams struct {
	// Argon2 parameters
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32

	// Scrypt parameters (alternative)
	N     int
	R     int
	P     int
	DkLen int
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Data      []byte `json:"data"`
	Nonce     []byte `json:"nonce"`
	Algorithm string `json:"algorithm"`
}

// HashedPassword represents a hashed password with metadata
type HashedPassword struct {
	Hash      string     `json:"hash"`
	Salt      []byte     `json:"salt"`
	Algorithm string     `json:"algorithm"`
	Params    HashParams `json:"params"`
}

// NewCryptoManager creates a new crypto manager with AES-GCM encryption
func NewCryptoManager(encryptionKey []byte) (*CryptoManager, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &CryptoManager{
		encryptionKey: encryptionKey,
		gcm:           gcm,
	}, nil
}

// NewPasswordHasher creates a new password hasher with secure defaults
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		algorithm: "argon2id",
		params: HashParams{
			Time:    3,         // 3 iterations
			Memory:  64 * 1024, // 64 MB
			Threads: 4,         // 4 threads
			KeyLen:  32,        // 32 bytes output
			SaltLen: 16,        // 16 bytes salt
		},
	}
}

// Encrypt encrypts data using AES-GCM
func (cm *CryptoManager) Encrypt(plaintext []byte) (*EncryptedData, error) {
	nonce := make([]byte, cm.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := cm.gcm.Seal(nil, nonce, plaintext, nil)

	return &EncryptedData{
		Data:      ciphertext,
		Nonce:     nonce,
		Algorithm: "AES-256-GCM",
	}, nil
}

// Decrypt decrypts data using AES-GCM
func (cm *CryptoManager) Decrypt(encrypted *EncryptedData) ([]byte, error) {
	if encrypted.Algorithm != "AES-256-GCM" {
		return nil, fmt.Errorf("unsupported algorithm: %s", encrypted.Algorithm)
	}

	plaintext, err := cm.gcm.Open(nil, encrypted.Nonce, encrypted.Data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (cm *CryptoManager) EncryptString(plaintext string) (string, error) {
	encrypted, err := cm.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}

	// Combine nonce and ciphertext for storage
	combined := append(encrypted.Nonce, encrypted.Data...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptString decrypts a base64 encoded string
func (cm *CryptoManager) DecryptString(encryptedStr string) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := cm.gcm.NonceSize()
	if len(combined) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	encrypted := &EncryptedData{
		Nonce:     combined[:nonceSize],
		Data:      combined[nonceSize:],
		Algorithm: "AES-256-GCM",
	}

	plaintext, err := cm.Decrypt(encrypted)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// HashPassword hashes a password using Argon2id
func (ph *PasswordHasher) HashPassword(password string) (*HashedPassword, error) {
	salt := make([]byte, ph.params.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	var hash []byte
	switch ph.algorithm {
	case "argon2id":
		hash = argon2.IDKey(
			[]byte(password),
			salt,
			ph.params.Time,
			ph.params.Memory,
			ph.params.Threads,
			ph.params.KeyLen,
		)
	case "scrypt":
		var err error
		hash, err = scrypt.Key(
			[]byte(password),
			salt,
			ph.params.N,
			ph.params.R,
			ph.params.P,
			ph.params.DkLen,
		)
		if err != nil {
			return nil, fmt.Errorf("scrypt hashing failed: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", ph.algorithm)
	}

	return &HashedPassword{
		Hash:      base64.StdEncoding.EncodeToString(hash),
		Salt:      salt,
		Algorithm: ph.algorithm,
		Params:    ph.params,
	}, nil
}

// VerifyPassword verifies a password against its hash
func (ph *PasswordHasher) VerifyPassword(password string, hashedPassword *HashedPassword) (bool, error) {
	var hash []byte
	var err error

	switch hashedPassword.Algorithm {
	case "argon2id":
		hash = argon2.IDKey(
			[]byte(password),
			hashedPassword.Salt,
			hashedPassword.Params.Time,
			hashedPassword.Params.Memory,
			hashedPassword.Params.Threads,
			hashedPassword.Params.KeyLen,
		)
	case "scrypt":
		hash, err = scrypt.Key(
			[]byte(password),
			hashedPassword.Salt,
			hashedPassword.Params.N,
			hashedPassword.Params.R,
			hashedPassword.Params.P,
			hashedPassword.Params.DkLen,
		)
		if err != nil {
			return false, fmt.Errorf("scrypt hashing failed: %w", err)
		}
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", hashedPassword.Algorithm)
	}

	expectedHash, err := base64.StdEncoding.DecodeString(hashedPassword.Hash)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, expectedHash) == 1, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateEncryptionKey generates a new 256-bit encryption key
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32) // 256 bits
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

// HashSHA256 creates a SHA-256 hash of the input
func HashSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashSHA256String creates a SHA-256 hash of a string and returns base64 encoded result
func HashSHA256String(data string) string {
	hash := HashSHA256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash)
}

// SecureCompare performs constant-time comparison of two byte slices
func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

// SecureCompareString performs constant-time comparison of two strings
func SecureCompareString(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// HMAC creates an HMAC-SHA256 of the data with the given key
func HMAC(key, data []byte) []byte {
	h := sha256.New()
	h.Write(key)
	h.Write(data)
	return h.Sum(nil)
}

// VerifyHMAC verifies an HMAC-SHA256 signature
func VerifyHMAC(key, data, signature []byte) bool {
	expected := HMAC(key, data)
	return SecureCompare(signature, expected)
}

// DeriveKey derives a key from a password using PBKDF2
func DeriveKey(password, salt []byte, iterations, keyLength int) []byte {
	return argon2.Key(password, salt, 3, 32*1024, 4, uint32(keyLength))
}

// SecureRandom generates cryptographically secure random bytes
func SecureRandom(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// ZeroBytes securely zeros out a byte slice
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// ZeroString securely zeros out a string (by converting to bytes)
func ZeroString(s *string) {
	if s == nil {
		return
	}
	// Convert string to byte slice and zero it
	bytes := []byte(*s)
	ZeroBytes(bytes)
	*s = ""
}

// ConstantTimeSelect returns x if v == 1 and y if v == 0
func ConstantTimeSelect(v int, x, y []byte) []byte {
	if subtle.ConstantTimeByteEq(uint8(v), 1) == 1 {
		return x
	}
	return y
}

// IsValidBase64 checks if a string is valid base64
func IsValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsValidBase64URL checks if a string is valid base64 URL encoding
func IsValidBase64URL(s string) bool {
	_, err := base64.URLEncoding.DecodeString(s)
	return err == nil
}
