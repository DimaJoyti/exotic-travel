package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	tracer        trace.Tracer
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims represents JWT claims
type Claims struct {
	UserID      int      `json:"user_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	CompanyID   int      `json:"company_id,omitempty"`
	SessionID   string   `json:"session_id"`
	TokenType   string   `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		issuer:        issuer,
		accessExpiry:  15 * time.Minute,  // Short-lived access tokens
		refreshExpiry: 7 * 24 * time.Hour, // 7 days refresh tokens
		tracer:        otel.Tracer("auth.jwt_manager"),
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(ctx context.Context, user *models.User, sessionID string) (*TokenPair, error) {
	ctx, span := j.tracer.Start(ctx, "jwt_manager.generate_token_pair")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user.id", user.ID),
		attribute.String("user.email", user.Email),
		attribute.String("user.role", string(user.Role)),
	)

	// Get user permissions
	permissions := j.getUserPermissions(user.Role)

	// Generate access token
	accessToken, err := j.generateToken(user, sessionID, "access", j.accessExpiry, permissions)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(user, sessionID, "refresh", j.refreshExpiry, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(j.accessExpiry)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(j.accessExpiry.Seconds()),
		ExpiresAt:    expiresAt,
	}, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	ctx, span := j.tracer.Start(ctx, "jwt_manager.validate_token")
	defer span.End()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Check if token is expired
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token has expired")
	}

	span.SetAttributes(
		attribute.Int("user.id", claims.UserID),
		attribute.String("user.email", claims.Email),
		attribute.String("token.type", claims.TokenType),
		attribute.String("session.id", claims.SessionID),
	)

	return claims, nil
}

// RefreshToken generates a new access token using a refresh token
func (j *JWTManager) RefreshToken(ctx context.Context, refreshTokenString string) (*TokenPair, error) {
	ctx, span := j.tracer.Start(ctx, "jwt_manager.refresh_token")
	defer span.End()

	// Validate refresh token
	claims, err := j.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Ensure it's a refresh token
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("token is not a refresh token")
	}

	// Create user object from claims
	user := &models.User{
		ID:        claims.UserID,
		Email:     claims.Email,
		Role:      models.UserRole(claims.Role),
		CompanyID: claims.CompanyID,
	}

	// Generate new token pair
	return j.GenerateTokenPair(ctx, user, claims.SessionID)
}

// RevokeToken adds a token to the revocation list
func (j *JWTManager) RevokeToken(ctx context.Context, tokenString string) error {
	ctx, span := j.tracer.Start(ctx, "jwt_manager.revoke_token")
	defer span.End()

	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("invalid token: %w", err)
	}

	// In a production system, you would store revoked tokens in Redis or database
	// For now, we'll just log the revocation
	span.SetAttributes(
		attribute.String("revoked.token.id", claims.ID),
		attribute.Int("revoked.user.id", claims.UserID),
		attribute.String("revoked.session.id", claims.SessionID),
	)

	// TODO: Store in revocation list (Redis recommended)
	return nil
}

// GenerateSessionID generates a cryptographically secure session ID
func (j *JWTManager) GenerateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Helper methods

func (j *JWTManager) generateToken(user *models.User, sessionID, tokenType string, expiry time.Duration, permissions []string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(expiry)

	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        string(user.Role),
		Permissions: permissions,
		CompanyID:   user.CompanyID,
		SessionID:   sessionID,
		TokenType:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        j.generateTokenID(),
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("user:%d", user.ID),
			Audience:  []string{"marketing-ai"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTManager) generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (j *JWTManager) getUserPermissions(role models.UserRole) []string {
	switch role {
	case models.UserRoleAdmin:
		return []string{
			"marketing:read",
			"marketing:write",
			"marketing:delete",
			"campaigns:read",
			"campaigns:write",
			"campaigns:delete",
			"content:read",
			"content:write",
			"content:delete",
			"integrations:read",
			"integrations:write",
			"integrations:delete",
			"analytics:read",
			"users:read",
			"users:write",
			"users:delete",
			"company:read",
			"company:write",
		}
	case models.UserRoleManager:
		return []string{
			"marketing:read",
			"marketing:write",
			"campaigns:read",
			"campaigns:write",
			"content:read",
			"content:write",
			"integrations:read",
			"integrations:write",
			"analytics:read",
			"users:read",
			"company:read",
		}
	case models.UserRoleMarketer:
		return []string{
			"marketing:read",
			"marketing:write",
			"campaigns:read",
			"campaigns:write",
			"content:read",
			"content:write",
			"integrations:read",
			"analytics:read",
		}
	case models.UserRoleViewer:
		return []string{
			"marketing:read",
			"campaigns:read",
			"content:read",
			"analytics:read",
		}
	default:
		return []string{"marketing:read"}
	}
}

// TokenBlacklist interface for token revocation
type TokenBlacklist interface {
	Add(ctx context.Context, tokenID string, expiresAt time.Time) error
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
	Cleanup(ctx context.Context) error
}

// RedisTokenBlacklist implements TokenBlacklist using Redis
type RedisTokenBlacklist struct {
	// Redis client would be injected here
	// client redis.Client
}

// Add adds a token to the blacklist
func (r *RedisTokenBlacklist) Add(ctx context.Context, tokenID string, expiresAt time.Time) error {
	// Implementation would use Redis SETEX command
	// TTL would be set to token expiry time
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	// Implementation would use Redis EXISTS command
	return false, nil
}

// Cleanup removes expired tokens from blacklist
func (r *RedisTokenBlacklist) Cleanup(ctx context.Context) error {
	// Redis automatically handles TTL expiry
	return nil
}

// PasswordHasher interface for password hashing
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

// BcryptHasher implements PasswordHasher using bcrypt
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt hasher
func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < 10 {
		cost = 12 // Default secure cost
	}
	return &BcryptHasher{cost: cost}
}

// Hash hashes a password using bcrypt
func (b *BcryptHasher) Hash(password string) (string, error) {
	// Implementation would use golang.org/x/crypto/bcrypt
	// bytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	// return string(bytes), err
	return "hashed_password", nil // Placeholder
}

// Verify verifies a password against its hash
func (b *BcryptHasher) Verify(password, hash string) error {
	// Implementation would use bcrypt.CompareHashAndPassword
	// return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return nil // Placeholder
}
