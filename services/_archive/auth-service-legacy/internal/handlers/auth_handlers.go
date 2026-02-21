package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/flyfleet/auth-service/internal/service"
	"github.com/flyfleet/auth-service/internal/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthHandlers struct {
	service *service.AuthService
	logger  *zap.Logger
}

func NewAuthHandlers(s *service.AuthService, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{
		service: s,
		logger:  logger,
	}
}

// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	ip := utils.GetClientIP(r) // Use validated IP extraction from shared utils
	ua := r.UserAgent()

	resp, err := h.service.Login(r.Context(), req.Email, req.Password, ip, ua)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err), zap.String("ip", ip))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Additional password strength validation
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	for _, char := range req.Password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case containsRune(specialChars, char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		http.Error(w, "password must contain uppercase, lowercase, digit, and special character", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if len(req.Name) > 100 {
		http.Error(w, "name must be less than 100 characters", http.StatusBadRequest)
		return
	}

	ip := utils.GetClientIP(r) // Use validated IP extraction from shared utils
	ua := r.UserAgent()

	user, err := h.service.Register(r.Context(), req.Email, req.Password, req.Name, ip, ua)
	if err != nil {
		h.logger.Error("Registration failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registration successful. Please check your email for verification.",
		"user":    user,
	})
}

// isValidEmail validates email format using a simple regex
func isValidEmail(email string) bool {
	// Simple email validation - in production, use a more robust solution
	if len(email) > 254 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) == 0 || len(localPart) > 64 || len(domainPart) == 0 || len(domainPart) > 253 {
		return false
	}

	return true
}

// containsRune checks if a string contains a specific rune
func containsRune(s string, r rune) bool {
	for _, char := range s {
		if char == r {
			return true
		}
	}
	return false
}

func (h *AuthHandlers) SendMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Method string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	if req.Method == "" {
		http.Error(w, "MFA method is required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	if err := h.service.SendMFA(r.Context(), userID, req.Method); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string `json:"userId"`
		OTPCode string `json:"otpCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	if req.OTPCode == "" {
		http.Error(w, "OTP code is required", http.StatusBadRequest)
		return
	}

	if len(req.OTPCode) != 6 {
		http.Error(w, "OTP code must be 6 digits", http.StatusBadRequest)
		return
	}

	ip := utils.GetClientIP(r) // Use validated IP extraction from shared utils
	ua := r.UserAgent()

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user ID format", http.StatusBadRequest)
		return
	}

	resp, err := h.service.VerifyMFA(r.Context(), userID, req.OTPCode, ip, ua)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandlers) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	userID, _ := uuid.Parse(userIDStr)

	orgs, err := h.service.GetUserOrganizations(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

func (h *AuthHandlers) SelectOrganization(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		OrgID  string `json:"orgId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ip := utils.GetClientIP(r) // Use validated IP extraction from shared utils
	ua := r.UserAgent()

	uID, _ := uuid.Parse(req.UserID)
	oID, _ := uuid.Parse(req.OrgID)

	resp, err := h.service.SelectOrganization(r.Context(), uID, oID, ip, ua)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ip := utils.GetClientIP(r) // Use validated IP extraction from shared utils
	ua := r.UserAgent()

	resp, err := h.service.RefreshToken(r.Context(), req.RefreshToken, ip, ua)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	_ = h.service.Logout(r.Context(), accessToken) // Need Logout method in service
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) GetSessions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	userID, _ := uuid.Parse(userIDStr)
	sessions, err := h.service.GetActiveSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

func (h *AuthHandlers) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	userID, _ := uuid.Parse(userIDStr)
	if err := h.service.LogoutAll(r.Context(), userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandlers) GetSecurityAudit(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	userID, _ := uuid.Parse(userIDStr)
	events, err := h.service.GetSecurityEvents(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(events)
}

func (h *AuthHandlers) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Theme  string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.UserID == "" {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	if req.Theme == "" {
		http.Error(w, "theme is required", http.StatusBadRequest)
		return
	}

	// Validate theme values
	if req.Theme != "light" && req.Theme != "dark" && req.Theme != "system" {
		http.Error(w, "theme must be one of: light, dark, system", http.StatusBadRequest)
		return
	}

	uID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ip := utils.GetClientIP(r) // Use validated IP extraction from shared utils
	ua := r.UserAgent()

	if err := h.service.UpdateTheme(r.Context(), uID, req.Theme, ip, ua); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Theme updated successfully"})
}

func (h *AuthHandlers) SSOCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	ssoErr := r.URL.Query().Get("error")

	if provider == "" {
		http.Error(w, "missing provider", http.StatusBadRequest)
		return
	}

	if ssoErr != "" {
		h.logger.Warn("SSO callback returned provider error", zap.String("provider", provider), zap.String("error", ssoErr))
		http.Error(w, "sso callback failed", http.StatusUnauthorized)
		return
	}

	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	// Provider exchange/session provisioning is delegated to enterprise IdP integration.
	h.logger.Info("SSO callback accepted", zap.String("provider", provider), zap.String("state", state))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"provider": provider,
		"status":   "callback_received",
		"state":    state,
	})
}

func (h *AuthHandlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	if err := h.service.VerifyEmail(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email verified successfully"})
}

func (h *AuthHandlers) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	if err := h.service.ResendVerification(r.Context(), req.Email); err != nil {
		// We don't return error to prevent user enumeration
		h.logger.Warn("Resend verification failed", zap.Error(err), zap.String("email", req.Email))
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If the account exists and is unverified, a new link has been sent."})
}

func (h *AuthHandlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	// Rate limiting for password reset attempts to prevent enumeration
	ip := r.RemoteAddr
	if !h.service.IsPasswordResetAllowed(r.Context(), req.Email, ip) {
		h.logger.Warn("Password reset rate limited", zap.String("email", req.Email), zap.String("ip", ip))
		// Still return success message to prevent user enumeration
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If the account exists, a reset link has been sent."})
		return
	}

	if err := h.service.ForgotPassword(r.Context(), req.Email); err != nil {
		h.logger.Error("Forgot password failed", zap.Error(err))
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If the account exists, a reset link has been sent."})
}

func (h *AuthHandlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	if req.NewPassword == "" {
		http.Error(w, "new password is required", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, "password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	if err := h.service.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

func (h *AuthHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Admin/CreateUser currently follows the same registration contract.
	h.Register(w, r)
}

func (h *AuthHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	// Extract access token from request to update last active time
	authHeader := r.Header.Get("Authorization")
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Update token's last active time for session management
	if accessToken != "" {
		go func() {
			// Run in a goroutine to avoid blocking the response
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.service.UpdateTokenLastActive(ctx, accessToken); err != nil {
				h.logger.Error("Failed to update token last active time", zap.Error(err))
			}
		}()
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch user", zap.Error(err))
		http.Error(w, "Failed to fetch user profile", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	permissions, err := h.service.GetEffectivePermissions(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch permissions", zap.Error(err))
		// continue without permissions or return error?
		// Return empty permissions to allow partial load
		permissions = []models.PermissionDefinition{}
	}

	orgs, err := h.service.GetUserOrganizations(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to fetch organizations", zap.Error(err))
		orgs = []models.Organization{}
	}

	// Transform Permissions to simple string list if needed, or keep as definitions
	// Frontend likely expects a list of permission codes strings or objects
	// Let's return the full object for flexibility

	if len(orgs) > 0 {
		user.OrgID = orgs[0].ID // Default to first org for profile
	}

	// Fetch role name if RoleID is present
	if user.RoleID != nil {
		role, err := h.service.GetRoleByID(r.Context(), *user.RoleID)
		if err == nil && role != nil {
			user.Role = role.Name
		}
	}

	resp := map[string]interface{}{
		"user":        user,
		"permissions": permissions,
		"orgs":        orgs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
