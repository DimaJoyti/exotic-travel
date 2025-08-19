package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token operations with enhanced security
type JWTManager struct {
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	issuer      string
	audience    string
	accessTTL   time.Duration
	refreshTTL  time.Duration
	blacklist   TokenBlacklist
	keyRotation *KeyRotationManager
}

// TokenClaims represents JWT claims with security enhancements
type TokenClaims struct {
	UserID      int64    `json:"user_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	SessionID   string   `json:"session_id"`
	TokenType   string   `json:"token_type"` // "access" or "refresh"
	DeviceID    string   `json:"device_id"`
	IPAddress   string   `json:"ip_address"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// TokenBlacklist interface for token revocation
type TokenBlacklist interface {
	Add(tokenID string, expiry time.Time) error
	IsBlacklisted(tokenID string) (bool, error)
	Cleanup() error
}

// KeyRotationManager handles automatic key rotation
type KeyRotationManager struct {
	currentKeyID string
	keys         map[string]*rsa.PrivateKey
	rotationTTL  time.Duration
	lastRotation time.Time
}

// NewJWTManager creates a new JWT manager with security features
func NewJWTManager(privateKeyPEM, publicKeyPEM, issuer, audience string, accessTTL, refreshTTL time.Duration, blacklist TokenBlacklist) (*JWTManager, error) {
	privateKey, err := parsePrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := parsePublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	keyRotation := &KeyRotationManager{
		currentKeyID: generateKeyID(),
		keys:         make(map[string]*rsa.PrivateKey),
		rotationTTL:  24 * time.Hour, // Rotate keys daily
		lastRotation: time.Now(),
	}
	keyRotation.keys[keyRotation.currentKeyID] = privateKey

	return &JWTManager{
		privateKey:  privateKey,
		publicKey:   publicKey,
		issuer:      issuer,
		audience:    audience,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		blacklist:   blacklist,
		keyRotation: keyRotation,
	}, nil
}

// GenerateTokenPair creates a new access and refresh token pair
func (jm *JWTManager) GenerateTokenPair(userID int64, email, role string, permissions []string, sessionID, deviceID, ipAddress string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(jm.accessTTL)
	refreshExpiry := now.Add(jm.refreshTTL)

	// Generate unique token IDs
	accessTokenID := generateTokenID()
	refreshTokenID := generateTokenID()

	// Create access token claims
	accessClaims := &TokenClaims{
		UserID:      userID,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		SessionID:   sessionID,
		TokenType:   "access",
		DeviceID:    deviceID,
		IPAddress:   ipAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenID,
			Issuer:    jm.issuer,
			Audience:  jwt.ClaimStrings{jm.audience},
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Create refresh token claims
	refreshClaims := &TokenClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
		TokenType: "refresh",
		DeviceID:  deviceID,
		IPAddress: ipAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenID,
			Issuer:    jm.issuer,
			Audience:  jwt.ClaimStrings{jm.audience},
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	// Sign tokens
	accessToken, err := jm.signToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken, err := jm.signToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates and parses a JWT token
func (jm *JWTManager) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check if token is blacklisted
	if jm.blacklist != nil {
		blacklisted, err := jm.blacklist.IsBlacklisted(claims.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check blacklist: %w", err)
		}
		if blacklisted {
			return nil, fmt.Errorf("token is revoked")
		}
	}

	// Validate audience
	if len(claims.Audience) == 0 || claims.Audience[0] != jm.audience {
		return nil, fmt.Errorf("invalid audience")
	}

	// Validate issuer
	if claims.Issuer != jm.issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}

// RefreshToken creates a new access token from a valid refresh token
func (jm *JWTManager) RefreshToken(refreshTokenString, deviceID, ipAddress string) (*TokenPair, error) {
	claims, err := jm.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Validate device and IP for additional security
	if claims.DeviceID != deviceID {
		return nil, fmt.Errorf("device mismatch")
	}

	if claims.IPAddress != ipAddress {
		return nil, fmt.Errorf("IP address mismatch")
	}

	// Generate new token pair
	return jm.GenerateTokenPair(
		claims.UserID,
		claims.Email,
		claims.Role,
		claims.Permissions,
		claims.SessionID,
		deviceID,
		ipAddress,
	)
}

// RevokeToken adds a token to the blacklist
func (jm *JWTManager) RevokeToken(tokenString string) error {
	claims, err := jm.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("cannot revoke invalid token: %w", err)
	}

	if jm.blacklist == nil {
		return fmt.Errorf("blacklist not configured")
	}

	return jm.blacklist.Add(claims.ID, claims.ExpiresAt.Time)
}

// RevokeAllUserTokens revokes all tokens for a specific user
func (jm *JWTManager) RevokeAllUserTokens(userID int64) error {
	// This would typically involve marking all tokens for the user as revoked
	// Implementation depends on the blacklist storage mechanism
	return fmt.Errorf("not implemented")
}

// signToken signs a token with the current private key
func (jm *JWTManager) signToken(claims *TokenClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Add key ID to header for key rotation support
	token.Header["kid"] = jm.keyRotation.currentKeyID

	return token.SignedString(jm.privateKey)
}

// RotateKeys rotates the signing keys if needed
func (jm *JWTManager) RotateKeys() error {
	if time.Since(jm.keyRotation.lastRotation) < jm.keyRotation.rotationTTL {
		return nil // No rotation needed
	}

	// Generate new key pair
	newPrivateKey, err := generateRSAKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate new key pair: %w", err)
	}

	// Update key rotation manager
	newKeyID := generateKeyID()
	jm.keyRotation.keys[newKeyID] = newPrivateKey
	jm.keyRotation.currentKeyID = newKeyID
	jm.keyRotation.lastRotation = time.Now()

	// Update JWT manager
	jm.privateKey = newPrivateKey
	jm.publicKey = &newPrivateKey.PublicKey

	// Clean up old keys (keep last 2 for grace period)
	if len(jm.keyRotation.keys) > 2 {
		for keyID := range jm.keyRotation.keys {
			if keyID != newKeyID && keyID != jm.keyRotation.currentKeyID {
				delete(jm.keyRotation.keys, keyID)
				break
			}
		}
	}

	return nil
}

// Helper functions

func parsePrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

func parsePublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPublicKey, nil
}

func generateRSAKeyPair() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func generateKeyID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
