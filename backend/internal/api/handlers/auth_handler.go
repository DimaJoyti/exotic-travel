package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/exotic-travel-booking/backend/internal/auth"
	"github.com/exotic-travel-booking/backend/internal/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AuthHandler handles authentication-related endpoints
type AuthHandler struct {
	jwtManager     *auth.JWTManager
	passwordHasher auth.PasswordHasher
	userService    UserService
	auditLogger    AuditLogger
	tracer         trace.Tracer
}

// UserService interface for user operations
type UserService interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, userID int) error
	ValidateUserCredentials(ctx context.Context, email, password string) (*models.User, error)
}

// AuditLogger interface for audit logging
type AuditLogger interface {
	LogAuthEvent(ctx context.Context, event *models.AuditEvent) error
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	CompanyID int    `json:"company_id,omitempty"`
}

// RefreshRequest represents token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Success bool              `json:"success"`
	User    *UserResponse     `json:"user,omitempty"`
	Tokens  *auth.TokenPair   `json:"tokens,omitempty"`
	Message string            `json:"message,omitempty"`
}

// UserResponse represents user data in responses
type UserResponse struct {
	ID          int                `json:"id"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	Email       string             `json:"email"`
	Role        models.UserRole    `json:"role"`
	Status      models.UserStatus  `json:"status"`
	CompanyID   int                `json:"company_id,omitempty"`
	Permissions []string           `json:"permissions"`
	CreatedAt   time.Time          `json:"created_at"`
	LastLogin   *time.Time         `json:"last_login,omitempty"`
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	jwtManager *auth.JWTManager,
	passwordHasher auth.PasswordHasher,
	userService UserService,
	auditLogger AuditLogger,
) *AuthHandler {
	return &AuthHandler{
		jwtManager:     jwtManager,
		passwordHasher: passwordHasher,
		userService:    userService,
		auditLogger:    auditLogger,
		tracer:         otel.Tracer("api.auth_handler"),
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth_handler.login")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateLoginRequest(&req); err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetAttributes(attribute.String("login.email", req.Email))

	// Validate credentials
	user, err := h.userService.ValidateUserCredentials(ctx, req.Email, req.Password)
	if err != nil {
		span.RecordError(err)
		// Log failed login attempt
		h.auditLogger.LogAuthEvent(ctx, &models.AuditEvent{
			Action:    "login_failed",
			UserEmail: req.Email,
			IPAddress: h.getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
			Details:   map[string]interface{}{"reason": "invalid_credentials"},
			Timestamp: time.Now(),
		})
		h.writeError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		h.auditLogger.LogAuthEvent(ctx, &models.AuditEvent{
			Action:    "login_failed",
			UserID:    user.ID,
			UserEmail: user.Email,
			IPAddress: h.getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
			Details:   map[string]interface{}{"reason": "account_inactive"},
			Timestamp: time.Now(),
		})
		h.writeError(w, "Account is not active", http.StatusUnauthorized)
		return
	}

	// Generate session ID
	sessionID, err := h.jwtManager.GenerateSessionID()
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	tokens, err := h.jwtManager.GenerateTokenPair(ctx, user, sessionID)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Update last login
	h.userService.UpdateLastLogin(ctx, user.ID)

	// Log successful login
	h.auditLogger.LogAuthEvent(ctx, &models.AuditEvent{
		Action:    "login_success",
		UserID:    user.ID,
		UserEmail: user.Email,
		IPAddress: h.getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Details:   map[string]interface{}{"session_id": sessionID},
		Timestamp: time.Now(),
	})

	span.SetAttributes(
		attribute.Int("user.id", user.ID),
		attribute.String("user.role", string(user.Role)),
		attribute.String("session.id", sessionID),
	)

	// Return response
	response := AuthResponse{
		Success: true,
		User:    h.userToResponse(user),
		Tokens:  tokens,
		Message: "Login successful",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth_handler.register")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRegisterRequest(&req); err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("register.email", req.Email),
		attribute.String("register.first_name", req.FirstName),
		attribute.String("register.last_name", req.LastName),
	)

	// Check if user already exists
	existingUser, _ := h.userService.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		h.writeError(w, "User with this email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := h.passwordHasher.Hash(req.Password)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         models.UserRoleMarketer, // Default role
		Status:       models.UserStatusActive,
		CompanyID:    req.CompanyID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.userService.CreateUser(ctx, user); err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Log registration
	h.auditLogger.LogAuthEvent(ctx, &models.AuditEvent{
		Action:    "user_registered",
		UserID:    user.ID,
		UserEmail: user.Email,
		IPAddress: h.getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
		Details:   map[string]interface{}{"role": user.Role},
		Timestamp: time.Now(),
	})

	span.SetAttributes(
		attribute.Int("user.id", user.ID),
		attribute.String("user.role", string(user.Role)),
	)

	// Return response (without tokens - user needs to login)
	response := AuthResponse{
		Success: true,
		User:    h.userToResponse(user),
		Message: "Registration successful. Please login to continue.",
	}

	h.writeJSON(w, response, http.StatusCreated)
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth_handler.refresh_token")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		h.writeError(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Refresh tokens
	tokens, err := h.jwtManager.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		span.RecordError(err)
		h.writeError(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Log token refresh
	claims, _ := h.jwtManager.ValidateToken(ctx, req.RefreshToken)
	if claims != nil {
		h.auditLogger.LogAuthEvent(ctx, &models.AuditEvent{
			Action:    "token_refreshed",
			UserID:    claims.UserID,
			UserEmail: claims.Email,
			IPAddress: h.getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
			Details:   map[string]interface{}{"session_id": claims.SessionID},
			Timestamp: time.Now(),
		})

		span.SetAttributes(
			attribute.Int("user.id", claims.UserID),
			attribute.String("session.id", claims.SessionID),
		)
	}

	response := AuthResponse{
		Success: true,
		Tokens:  tokens,
		Message: "Token refreshed successfully",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth_handler.logout")
	defer span.End()

	if r.Method != http.MethodPost {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	token, err := h.extractToken(r)
	if err != nil {
		h.writeError(w, "Missing authorization token", http.StatusBadRequest)
		return
	}

	// Revoke token
	if err := h.jwtManager.RevokeToken(ctx, token); err != nil {
		span.RecordError(err)
		h.writeError(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Log logout
	claims, _ := h.jwtManager.ValidateToken(ctx, token)
	if claims != nil {
		h.auditLogger.LogAuthEvent(ctx, &models.AuditEvent{
			Action:    "logout",
			UserID:    claims.UserID,
			UserEmail: claims.Email,
			IPAddress: h.getClientIP(r),
			UserAgent: r.Header.Get("User-Agent"),
			Details:   map[string]interface{}{"session_id": claims.SessionID},
			Timestamp: time.Now(),
		})

		span.SetAttributes(
			attribute.Int("user.id", claims.UserID),
			attribute.String("session.id", claims.SessionID),
		)
	}

	response := AuthResponse{
		Success: true,
		Message: "Logout successful",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Me handles GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "auth_handler.me")
	defer span.End()

	if r.Method != http.MethodGet {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	user := h.getUserFromContext(r.Context())
	if user == nil {
		h.writeError(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	span.SetAttributes(
		attribute.Int("user.id", user.ID),
		attribute.String("user.email", user.Email),
	)

	response := AuthResponse{
		Success: true,
		User:    h.userToResponse(user),
		Message: "User information retrieved successfully",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Helper methods

func (h *AuthHandler) validateLoginRequest(req *LoginRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func (h *AuthHandler) validateRegisterRequest(req *RegisterRequest) error {
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func (h *AuthHandler) userToResponse(user *models.User) *UserResponse {
	// Get user permissions based on role
	permissions := h.getUserPermissions(user.Role)

	return &UserResponse{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Role:        user.Role,
		Status:      user.Status,
		CompanyID:   user.CompanyID,
		Permissions: permissions,
		CreatedAt:   user.CreatedAt,
		LastLogin:   user.LastLogin,
	}
}

func (h *AuthHandler) getUserPermissions(role models.UserRole) []string {
	// This would typically come from a permissions service
	switch role {
	case models.UserRoleAdmin:
		return []string{"marketing:*", "campaigns:*", "users:*", "company:*"}
	case models.UserRoleManager:
		return []string{"marketing:read", "marketing:write", "campaigns:*", "users:read"}
	case models.UserRoleMarketer:
		return []string{"marketing:read", "marketing:write", "campaigns:read", "campaigns:write"}
	case models.UserRoleViewer:
		return []string{"marketing:read", "campaigns:read"}
	default:
		return []string{}
	}
}

func (h *AuthHandler) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func (h *AuthHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func (h *AuthHandler) getUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(auth.UserContextKey).(*models.User); ok {
		return user
	}
	return nil
}

func (h *AuthHandler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]interface{}{
		"success":   false,
		"error":     message,
		"timestamp": time.Now().Unix(),
	}
	h.writeJSON(w, response, statusCode)
}
