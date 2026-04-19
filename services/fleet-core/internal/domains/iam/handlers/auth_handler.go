package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authSvc     authService
	passwordSvc passwordService
	log         *zap.Logger
}

func NewAuthHandler(authSvc authService, passwordSvc passwordService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc:     authSvc,
		passwordSvc: passwordSvc,
		log:         log,
	}
}

// Identify checks if a user exists and returns org info + MFA status.
// POST /api/v1/auth/identify
func (h *AuthHandler) Identify(w http.ResponseWriter, r *http.Request) {
	var req models.IdentifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Email == "" {
		h.writeError(w, http.StatusBadRequest, "missing_email", "Email is required", r)
		return
	}

	resp, err := h.authSvc.Identify(r.Context(), req.Email)
	if err != nil {
		h.log.Error("identify failed", zap.Error(err), zap.String("email", req.Email))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Login authenticates a user with email + password.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Email == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "missing_credentials", "Email and password are required", r)
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	tokens, mfaResp, err := h.authSvc.Login(r.Context(), req, ip, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			h.writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password", r)
		case errors.Is(err, service.ErrAccountLocked):
			h.writeError(w, http.StatusForbidden, "account_locked", "Account is locked. Try again later", r)
		case errors.Is(err, service.ErrUserNotFound):
			h.writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password", r)
		default:
			h.log.Error("login failed", zap.Error(err), zap.String("email", req.Email))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	if mfaResp != nil {
		h.writeJSON(w, http.StatusAccepted, mfaResp)
		return
	}

	h.writeJSON(w, http.StatusOK, tokens)
}

// VerifyMFA verifies an MFA code and returns tokens.
// POST /api/v1/auth/mfa/verify
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req models.MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.TempToken == "" || req.Code == "" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "temp_token and code are required", r)
		return
	}

	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	tokens, err := h.authSvc.VerifyMFA(r.Context(), req, ip, userAgent)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMFACode):
			h.writeError(w, http.StatusUnauthorized, "invalid_mfa_code", "Invalid MFA code", r)
		case errors.Is(err, service.ErrUserNotFound):
			h.writeError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token", r)
		default:
			h.log.Error("mfa verification failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, tokens)
}

// Refresh generates new access tokens using a refresh token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.RefreshToken == "" {
		h.writeError(w, http.StatusBadRequest, "missing_token", "refresh_token is required", r)
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	resp, err := h.authSvc.Refresh(r.Context(), req.RefreshToken, ip, ua)
	if err != nil {
		switch {
		case err.Error() == "invalid_token" || err.Error() == "invalid_token_type":
			h.writeError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired refresh token", r)
		case errors.Is(err, service.ErrAccountLocked):
			h.writeError(w, http.StatusForbidden, "account_locked", "Account is locked", r)
		default:
			h.log.Error("token refresh failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// ChangePassword changes the authenticated user's password.
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "current_password and new_password are required", r)
		return
	}

	if err := h.passwordSvc.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			h.writeError(w, http.StatusUnauthorized, "invalid_current_password", "Current password is incorrect", r)
		case errors.Is(err, service.ErrPasswordTooShort),
			errors.Is(err, service.ErrPasswordNoUpper),
			errors.Is(err, service.ErrPasswordNoLower),
			errors.Is(err, service.ErrPasswordNoDigit),
			errors.Is(err, service.ErrPasswordNoSpecial):
			h.writeError(w, http.StatusBadRequest, "password_too_weak", "Password does not meet policy requirements", r)
		case errors.Is(err, service.ErrPasswordReused):
			h.writeError(w, http.StatusBadRequest, "password_reused", "Password was recently used", r)
		default:
			h.log.Error("password change failed", zap.Error(err), zap.String("user_id", userID.String()))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "password_changed"})
}

// GetPolicy returns the password policy.
// GET /api/v1/auth/password-policy
func (h *AuthHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	policy := h.passwordSvc.GetPolicy()
	h.writeJSON(w, http.StatusOK, policy)
}

// SwitchOrg switches the user's active organization.
// POST /api/v1/auth/orgs/{orgId}/switch
func (h *AuthHandler) SwitchOrg(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "orgId")
	if orgIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "missing_org", "Organization ID is required", r)
		return
	}

	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_org_id", "Invalid organization ID", r)
		return
	}

	resp, err := h.authSvc.SwitchOrg(r.Context(), userID, orgID, r.RemoteAddr, r.UserAgent())
	if err != nil {
		switch {
		case err.Error() == "org_switch_failed":
			h.writeError(w, http.StatusBadRequest, "org_switch_failed", "You do not have access to this organization", r)
		default:
			h.log.Error("org switch failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Register creates a new user account.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		OrgID    string `json:"tenant_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "email, password, and name are required", r)
		return
	}

	var orgID uuid.UUID
	if req.OrgID != "" {
		var err error
		orgID, err = uuid.Parse(req.OrgID)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "invalid_org_id", "Invalid organization ID", r)
			return
		}
	}

	tokens, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.Name, orgID)
	if err != nil {
		switch {
		case err.Error() == "email_already_exists":
			h.writeError(w, http.StatusConflict, "email_already_exists", "An account with this email already exists", r)
		default:
			h.log.Error("registration failed", zap.Error(err), zap.String("email", req.Email))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, tokens)
}

// Logout revokes the current session.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	tokenJTI := GetTokenJTI(r)

	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	if err := h.authSvc.Logout(r.Context(), userID, tokenJTI); err != nil {
		h.log.Error("logout failed", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "logged_out"})
}

// VerifyEmail verifies a user's email using a verification token.
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Token == "" {
		h.writeError(w, http.StatusBadRequest, "missing_token", "Verification token is required", r)
		return
	}

	if err := h.authSvc.VerifyEmail(r.Context(), req.Token); err != nil {
		switch {
		case err.Error() == "invalid_token":
			h.writeError(w, http.StatusBadRequest, "invalid_token", "Invalid or expired verification token", r)
		default:
			h.log.Error("email verification failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "email_verified"})
}

// ResetPassword resets a user's password using a reset token.
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "token and new_password are required", r)
		return
	}

	if err := h.authSvc.ResetPasswordWithToken(r.Context(), req.Token, req.NewPassword); err != nil {
		switch {
		case err.Error() == "invalid_token":
			h.writeError(w, http.StatusBadRequest, "invalid_token", "Invalid or expired reset token", r)
		case err.Error() == "token_already_used":
			h.writeError(w, http.StatusBadRequest, "token_already_used", "This reset link has already been used", r)
		case errors.Is(err, service.ErrPasswordTooShort),
			errors.Is(err, service.ErrPasswordNoUpper),
			errors.Is(err, service.ErrPasswordNoLower),
			errors.Is(err, service.ErrPasswordNoDigit),
			errors.Is(err, service.ErrPasswordNoSpecial):
			h.writeError(w, http.StatusBadRequest, "password_too_weak", "Password does not meet policy requirements", r)
		case errors.Is(err, service.ErrPasswordReused):
			h.writeError(w, http.StatusBadRequest, "password_reused", "Password was recently used", r)
		default:
			h.log.Error("password reset failed", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "password_reset"})
}

// ForgotPassword initiates a password reset flow.
// POST /api/v1/auth/forgot-password
// Anti-enumeration: always returns 200 with generic message regardless of whether email exists.
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Email == "" {
		h.writeError(w, http.StatusBadRequest, "missing_email", "Email is required", r)
		return
	}

	// Always returns success — anti-enumeration
	if err := h.authSvc.ForgotPassword(r.Context(), req.Email); err != nil {
		h.log.Error("forgot password failed", zap.Error(err), zap.String("email", req.Email))
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "If an account exists with that email, a password reset link has been sent.",
	})
}

// ListOrgs returns the user's organizations.
// GET /api/v1/auth/orgs
func (h *AuthHandler) ListOrgs(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	orgs, err := h.authSvc.GetUserOrganizations(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to list orgs", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to fetch organizations", r)
		return
	}

	h.writeJSON(w, http.StatusOK, orgs)
}

// Health returns the service health status.
// GET /health
func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, models.HealthResponse{
		Status:    "ok",
		Services:  map[string]string{"auth": "healthy"},
		Timestamp: "now",
	})
}

// Ready returns whether the service is ready to accept traffic.
// GET /health/ready
func (h *AuthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *AuthHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	correlationID := r.Header.Get("X-Correlation-ID")
	resp := models.ErrorResponse{
		Error:     code,
		Message:   message,
		RequestID: correlationID,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
