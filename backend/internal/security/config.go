package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	// JWT Configuration
	JWTPrivateKey    string        `json:"jwt_private_key"`
	JWTPublicKey     string        `json:"jwt_public_key"`
	JWTIssuer        string        `json:"jwt_issuer"`
	JWTAudience      string        `json:"jwt_audience"`
	JWTAccessTTL     time.Duration `json:"jwt_access_ttl"`
	JWTRefreshTTL    time.Duration `json:"jwt_refresh_ttl"`
	
	// Encryption Configuration
	EncryptionKey    []byte `json:"-"` // Don't serialize encryption key
	
	// Password Hashing Configuration
	PasswordAlgorithm string `json:"password_algorithm"`
	ArgonTime         uint32 `json:"argon_time"`
	ArgonMemory       uint32 `json:"argon_memory"`
	ArgonThreads      uint8  `json:"argon_threads"`
	ArgonKeyLen       uint32 `json:"argon_key_len"`
	ArgonSaltLen      uint32 `json:"argon_salt_len"`
	
	// Rate Limiting Configuration
	RateLimitRPS     float64 `json:"rate_limit_rps"`
	RateLimitBurst   int     `json:"rate_limit_burst"`
	RateLimitEnabled bool    `json:"rate_limit_enabled"`
	
	// Security Headers Configuration
	CSPEnabled       bool     `json:"csp_enabled"`
	CSPDirectives    string   `json:"csp_directives"`
	HSTSEnabled      bool     `json:"hsts_enabled"`
	HSTSMaxAge       int      `json:"hsts_max_age"`
	
	// CORS Configuration
	CORSEnabled      bool     `json:"cors_enabled"`
	CORSOrigins      []string `json:"cors_origins"`
	CORSMethods      []string `json:"cors_methods"`
	CORSHeaders      []string `json:"cors_headers"`
	CORSCredentials  bool     `json:"cors_credentials"`
	
	// IP Filtering Configuration
	TrustedProxies   []string `json:"trusted_proxies"`
	BlockedIPs       []string `json:"blocked_ips"`
	IPWhitelistOnly  bool     `json:"ip_whitelist_only"`
	AllowedIPs       []string `json:"allowed_ips"`
	
	// Session Configuration
	SessionTimeout   time.Duration `json:"session_timeout"`
	SessionSecure    bool          `json:"session_secure"`
	SessionHTTPOnly  bool          `json:"session_http_only"`
	SessionSameSite  string        `json:"session_same_site"`
	
	// Audit Configuration
	AuditEnabled     bool   `json:"audit_enabled"`
	AuditLevel       string `json:"audit_level"`
	AuditRetention   int    `json:"audit_retention_days"`
	
	// Validation Configuration
	MaxRequestSize   int64 `json:"max_request_size"`
	MaxStringLength  int   `json:"max_string_length"`
	
	// TLS Configuration
	TLSEnabled       bool   `json:"tls_enabled"`
	TLSCertFile      string `json:"tls_cert_file"`
	TLSKeyFile       string `json:"tls_key_file"`
	TLSMinVersion    string `json:"tls_min_version"`
	
	// Development/Debug Configuration
	DebugMode        bool `json:"debug_mode"`
	SecurityTesting  bool `json:"security_testing"`
}

// LoadSecurityConfig loads security configuration from environment variables
func LoadSecurityConfig() (*SecurityConfig, error) {
	config := &SecurityConfig{
		// JWT defaults
		JWTIssuer:     getEnvString("JWT_ISSUER", "exotic-travel-booking"),
		JWTAudience:   getEnvString("JWT_AUDIENCE", "exotic-travel-api"),
		JWTAccessTTL:  getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL: getEnvDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		
		// Password hashing defaults
		PasswordAlgorithm: getEnvString("PASSWORD_ALGORITHM", "argon2id"),
		ArgonTime:         uint32(getEnvInt("ARGON_TIME", 3)),
		ArgonMemory:       uint32(getEnvInt("ARGON_MEMORY", 64*1024)),
		ArgonThreads:      uint8(getEnvInt("ARGON_THREADS", 4)),
		ArgonKeyLen:       uint32(getEnvInt("ARGON_KEY_LEN", 32)),
		ArgonSaltLen:      uint32(getEnvInt("ARGON_SALT_LEN", 16)),
		
		// Rate limiting defaults
		RateLimitRPS:     getEnvFloat("RATE_LIMIT_RPS", 10.0),
		RateLimitBurst:   getEnvInt("RATE_LIMIT_BURST", 20),
		RateLimitEnabled: getEnvBool("RATE_LIMIT_ENABLED", true),
		
		// Security headers defaults
		CSPEnabled:    getEnvBool("CSP_ENABLED", true),
		CSPDirectives: getEnvString("CSP_DIRECTIVES", "default-src 'self'"),
		HSTSEnabled:   getEnvBool("HSTS_ENABLED", true),
		HSTSMaxAge:    getEnvInt("HSTS_MAX_AGE", 31536000),
		
		// CORS defaults
		CORSEnabled:     getEnvBool("CORS_ENABLED", true),
		CORSOrigins:     getEnvStringSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
		CORSMethods:     getEnvStringSlice("CORS_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		CORSHeaders:     getEnvStringSlice("CORS_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With"}),
		CORSCredentials: getEnvBool("CORS_CREDENTIALS", true),
		
		// IP filtering defaults
		TrustedProxies:  getEnvStringSlice("TRUSTED_PROXIES", []string{"127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}),
		BlockedIPs:      getEnvStringSlice("BLOCKED_IPS", []string{}),
		IPWhitelistOnly: getEnvBool("IP_WHITELIST_ONLY", false),
		AllowedIPs:      getEnvStringSlice("ALLOWED_IPS", []string{}),
		
		// Session defaults
		SessionTimeout:  getEnvDuration("SESSION_TIMEOUT", 24*time.Hour),
		SessionSecure:   getEnvBool("SESSION_SECURE", true),
		SessionHTTPOnly: getEnvBool("SESSION_HTTP_ONLY", true),
		SessionSameSite: getEnvString("SESSION_SAME_SITE", "Strict"),
		
		// Audit defaults
		AuditEnabled:   getEnvBool("AUDIT_ENABLED", true),
		AuditLevel:     getEnvString("AUDIT_LEVEL", "INFO"),
		AuditRetention: getEnvInt("AUDIT_RETENTION_DAYS", 90),
		
		// Validation defaults
		MaxRequestSize:  getEnvInt64("MAX_REQUEST_SIZE", 10*1024*1024), // 10MB
		MaxStringLength: getEnvInt("MAX_STRING_LENGTH", 10000),
		
		// TLS defaults
		TLSEnabled:    getEnvBool("TLS_ENABLED", false),
		TLSCertFile:   getEnvString("TLS_CERT_FILE", ""),
		TLSKeyFile:    getEnvString("TLS_KEY_FILE", ""),
		TLSMinVersion: getEnvString("TLS_MIN_VERSION", "1.2"),
		
		// Debug defaults
		DebugMode:       getEnvBool("DEBUG_MODE", false),
		SecurityTesting: getEnvBool("SECURITY_TESTING", false),
	}
	
	// Load or generate JWT keys
	if err := config.loadJWTKeys(); err != nil {
		return nil, fmt.Errorf("failed to load JWT keys: %w", err)
	}
	
	// Load or generate encryption key
	if err := config.loadEncryptionKey(); err != nil {
		return nil, fmt.Errorf("failed to load encryption key: %w", err)
	}
	
	return config, nil
}

// loadJWTKeys loads or generates JWT RSA key pair
func (c *SecurityConfig) loadJWTKeys() error {
	privateKeyPEM := os.Getenv("JWT_PRIVATE_KEY")
	publicKeyPEM := os.Getenv("JWT_PUBLIC_KEY")
	
	// If keys are not provided, generate them
	if privateKeyPEM == "" || publicKeyPEM == "" {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate RSA key: %w", err)
		}
		
		// Encode private key
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyPEM = string(pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}))
		
		// Encode public key
		publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshal public key: %w", err)
		}
		publicKeyPEM = string(pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		}))
		
		// In production, these should be stored securely
		fmt.Println("Generated new JWT keys. Store these securely:")
		fmt.Printf("JWT_PRIVATE_KEY=%s\n", privateKeyPEM)
		fmt.Printf("JWT_PUBLIC_KEY=%s\n", publicKeyPEM)
	}
	
	c.JWTPrivateKey = privateKeyPEM
	c.JWTPublicKey = publicKeyPEM
	
	return nil
}

// loadEncryptionKey loads or generates encryption key
func (c *SecurityConfig) loadEncryptionKey() error {
	encryptionKeyHex := os.Getenv("ENCRYPTION_KEY")
	
	if encryptionKeyHex == "" {
		// Generate new encryption key
		key, err := GenerateEncryptionKey()
		if err != nil {
			return fmt.Errorf("failed to generate encryption key: %w", err)
		}
		c.EncryptionKey = key
		
		// In production, this should be stored securely
		fmt.Printf("Generated new encryption key. Store this securely:\nENCRYPTION_KEY=%x\n", key)
	} else {
		// Decode hex key
		key := make([]byte, 32)
		n, err := fmt.Sscanf(encryptionKeyHex, "%x", &key)
		if err != nil || n != 1 {
			return fmt.Errorf("invalid encryption key format")
		}
		c.EncryptionKey = key
	}
	
	return nil
}

// Validate validates the security configuration
func (c *SecurityConfig) Validate() error {
	// Validate JWT configuration
	if c.JWTIssuer == "" {
		return fmt.Errorf("JWT issuer is required")
	}
	if c.JWTAudience == "" {
		return fmt.Errorf("JWT audience is required")
	}
	if c.JWTAccessTTL <= 0 {
		return fmt.Errorf("JWT access TTL must be positive")
	}
	if c.JWTRefreshTTL <= 0 {
		return fmt.Errorf("JWT refresh TTL must be positive")
	}
	
	// Validate encryption key
	if len(c.EncryptionKey) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes")
	}
	
	// Validate rate limiting
	if c.RateLimitRPS <= 0 {
		return fmt.Errorf("rate limit RPS must be positive")
	}
	if c.RateLimitBurst <= 0 {
		return fmt.Errorf("rate limit burst must be positive")
	}
	
	// Validate request size limits
	if c.MaxRequestSize <= 0 {
		return fmt.Errorf("max request size must be positive")
	}
	if c.MaxStringLength <= 0 {
		return fmt.Errorf("max string length must be positive")
	}
	
	return nil
}

// Helper functions for environment variable parsing

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
