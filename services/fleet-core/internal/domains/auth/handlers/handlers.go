package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	authservice "github.com/flyfleet/fleet-core-service/internal/domains/auth/service"
	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	internalmw "github.com/flyfleet/fleet-core-service/internal/middleware"
	pb "github.com/flyfleet/proto/auth/v1"
	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Handlers struct {
	svc         *authservice.AuthService
	logger      *zap.Logger
	redis       *redis.Client
	rateLimiter *internalmw.RedisRateLimiter
}

func NewHandlers(svc *authservice.AuthService, logger *zap.Logger, redis *redis.Client) *Handlers {
	return &Handlers{
		svc:         svc,
		logger:      logger,
		redis:       redis,
		rateLimiter: internalmw.NewRedisRateLimiter(redis, logger),
	}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	// Standard limits
	loginLimit := h.rateLimiter.Handler("login", 1000, time.Minute)
	registerLimit := h.rateLimiter.Handler("register", 3, time.Minute)
	inviteLimit := h.rateLimiter.Handler("invite_check", 5, 15*time.Minute)

	r.Route("/auth", func(r chi.Router) {
		r.With(loginLimit).Post("/login", h.Login)
		r.Post("/refresh", h.RefreshToken)
		r.Post("/revoke", h.RevokeToken)
		r.With(registerLimit).Post("/register", h.Register)
		r.Post("/validate-token", h.ValidateToken)
		r.Post("/login/mfa/verify", h.VerifyLoginMFA)
		r.Post("/login/mfa/setup", h.SetupAdminMFA)
		r.Post("/login/oidc/{provider}", h.LoginOIDC)
		r.Post("/password/reset/request", h.RequestPasswordReset)
		r.Post("/password/reset/confirm", h.ConfirmPasswordReset)
		r.Post("/resolve-orgs", h.ResolveOrgs)
		r.With(inviteLimit).Post("/invites/validate", h.ValidateInvite)

		// Protected auth endpoints
		r.Group(func(r chi.Router) {
			// Extract JWKS endpoint for middleware
			// JWKS URL for middleware - configurable for tests
			jwksURL := os.Getenv("JWKS_URL")
			if jwksURL == "" {
				jwksURL = "http://localhost:8084/api/v1/auth/.well-known/jwks.json"
			}
			authGate := internalmw.NewJWKSAuthGate(jwksURL, "FlyFleet", "fleet-dashboard", h.logger)
			r.Use(authGate.Handler)

			// MFA Management
			r.Post("/mfa/generate-secret", h.GenerateMFASecret)
			r.Post("/mfa/setup", h.GenerateMFASecret) // Frontend Alias for Init
			r.Post("/mfa/enable", h.EnableMFA)
			r.Post("/mfa/disable", h.DisableMFA)
			r.Post("/mfa/verify", h.VerifyMFAAuthenticated)
			r.Post("/switch-context", h.SwitchContext)

			// Session endpoints
			r.Get("/users/{id}/sessions", h.GetSessions)
			r.Post("/users/{id}/sessions/{session_id}/revoke", h.RevokeSession)

			r.Get("/users/{id}", h.GetUser)
			r.Get("/organizations/{id}", h.GetOrganization)
			r.Get("/organizations", h.ListOrganizations)
			r.Get("/stats", h.GetStats)
			r.Get("/security-telemetry", h.GetSecurityTelemetry)
			r.Get("/audit-logs", h.GetAuditLogs)
			
			// Admin Console endpoints
			r.Post("/api-keys", h.CreateAPIKey)
			r.Get("/api-keys", h.ListAPIKeys)
			r.Post("/api-keys/{id}/revoke", h.RevokeAPIKey)
			
			r.Post("/service-accounts", h.CreateServiceAccount)
			r.Get("/service-accounts", h.ListServiceAccounts)
			
			r.Post("/emergency-revoke", h.EmergencyRevoke)
		})

		r.Get("/.well-known/jwks.json", h.GetJWKS)
	})
}

func (h *Handlers) json(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handlers) errorJSON(w http.ResponseWriter, status int, msg string) {
	h.json(w, status, map[string]string{"error": msg})
}

// Rate limiting refactored to middleware



func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		OrgID    string `json:"org_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// We pass empty string for IP and UA as they are captured by middleware now
	// Or we can extract them if needed by the service, but let's keep it simple.
	resp, err := h.svc.Login(r.Context(), req.Email, req.Password, req.OrgID, "", "")
	if err != nil {
		h.logger.Error("auth login failed", zap.Error(err))
		if strings.Contains(err.Error(), "locked") {
			h.errorJSON(w, http.StatusForbidden, "account_locked")
			return
		}
		h.errorJSON(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.RefreshToken(r.Context(), req.RefreshToken, "", "")
	if err != nil {
		h.logger.Error("token refresh failed", zap.Error(err))
		h.errorJSON(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}
	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.RevokeToken(r.Context(), req.RefreshToken); err != nil {
		h.logger.Error("token revoke failed", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to revoke token")
		return
	}
	h.json(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req authservice.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.RegisterUser(r.Context(), req)
	if err != nil {
		h.logger.Error("auth register failed", zap.Error(err))

		// Check if it's a validation error
		var ve authservice.ValidationError
		if errors.As(err, &ve) {
			h.errorJSON(w, http.StatusBadRequest, ve.Error())
			return
		}

		h.errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.json(w, http.StatusCreated, resp)
}

func (h *Handlers) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req pb.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.ValidateToken(r.Context(), req.Token)
	if err != nil {
		h.json(w, http.StatusOK, map[string]any{"valid": false})
		return
	}
	h.json(w, http.StatusOK, map[string]any{"valid": true, "user": user})
}

func (h *Handlers) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.ValidateInvite(r.Context(), req.Code)
	if err != nil {
		// Log internal errors, but return ambiguous response to attackers
		h.logger.Error("failed to validate invite natively", zap.Error(err))
		h.json(w, http.StatusOK, map[string]any{"valid": false})
		return
	}

	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.svc.GetUserByID(r.Context(), id)
	if err != nil {
		h.errorJSON(w, http.StatusNotFound, "user not found")
		return
	}
	h.json(w, http.StatusOK, user)
}

func (h *Handlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid organization id")
		return
	}
	org, err := h.svc.GetOrganizationByID(r.Context(), id)
	if err != nil {
		h.errorJSON(w, http.StatusNotFound, "organization not found")
		return
	}
	h.json(w, http.StatusOK, org)
}

func (h *Handlers) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := int32((page - 1) * limit)
	orgs, total, err := h.svc.ListOrganizations(r.Context(), int32(limit), offset)
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}
	h.json(w, http.StatusOK, map[string]any{"organizations": orgs, "total": total, "page": page, "limit": limit})
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to get stats")
		return
	}
	h.json(w, http.StatusOK, stats)
}

func (h *Handlers) GetSecurityTelemetry(w http.ResponseWriter, r *http.Request) {
	telemetry, err := h.svc.GetSecurityTelemetry(r.Context())
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to get security telemetry")
		return
	}
	h.json(w, http.StatusOK, telemetry)
}

func (h *Handlers) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := int32((page - 1) * limit)
	logs, total, err := h.svc.GetAuditLogs(r.Context(), int32(limit), offset)
	if err != nil {
		h.errorJSON(w, http.StatusInternalServerError, "failed to get audit logs")
		return
	}
	h.json(w, http.StatusOK, map[string]any{"logs": logs, "total": total, "page": page, "limit": limit})
}

// MFA Handlers

func (h *Handlers) GenerateMFASecret(w http.ResponseWriter, r *http.Request) {
	var userID uuid.UUID
	var err error

	// Try extracting from context first (modern)
	if val := r.Context().Value(ctxkeys.UserIDKey); val != nil {
		if id, ok := val.(uuid.UUID); ok {
			userID = id
		} else if idStr, ok := val.(string); ok {
			userID, _ = uuid.Parse(idStr)
		}
	}

	// Fallback to header (legacy/manual)
	if userID == uuid.Nil {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr != "" {
			userID, err = uuid.Parse(userIDStr)
		}
	}

	if userID == uuid.Nil {
		h.errorJSON(w, http.StatusBadRequest, "user ID required")
		return
	}

	secret, uri, err := h.svc.GenerateMFASecret(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to generate MFA secret", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to generate secret")
		return
	}

	h.json(w, http.StatusOK, map[string]any{
		"secret":     secret,
		"qr_uri":     uri,
		"issuer":     "FlyFleet",
		"algorithm":  "SHA1",
		"digits":     6,
		"period":     30,
	})
}

func (h *Handlers) EnableMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Try extracting userID from context
	var userID uuid.UUID
	if val := r.Context().Value(ctxkeys.UserIDKey); val != nil {
		if id, ok := val.(uuid.UUID); ok {
			userID = id
		} else if idStr, ok := val.(string); ok {
			userID, _ = uuid.Parse(idStr)
		}
	}

	if userID == uuid.Nil {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr != "" {
			userID, _ = uuid.Parse(userIDStr)
		}
	}

	if userID == uuid.Nil {
		h.errorJSON(w, http.StatusBadRequest, "user ID required")
		return
	}

	// If secret is missing, fetch from user profile
	if req.Secret == "" {
		user, err := h.svc.GetUserByID(r.Context(), userID)
		if err != nil || user.MfaSecret.String == "" {
			h.errorJSON(w, http.StatusBadRequest, "MFA secret not found for user")
			return
		}
		req.Secret = user.MfaSecret.String
	}

	codes, err := h.svc.EnableMFA(r.Context(), userID, req.Secret, req.Code)
	if err != nil {
		h.logger.Error("failed to enable MFA", zap.Error(err))
		h.errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]any{
		"status":         "mfa_enabled",
		"recovery_codes": codes,
	})
}

func (h *Handlers) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		h.errorJSON(w, http.StatusBadRequest, "user ID required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.DisableMFA(r.Context(), userID, req.Code); err != nil {
		h.logger.Error("failed to disable MFA", zap.Error(err))
		h.errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]string{"status": "mfa_disabled"})
}

func (h *Handlers) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	valid, err := h.svc.VerifyTOTP(req.Secret, req.Code)
	if err != nil {
		h.logger.Error("MFA verification error", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "verification error")
		return
	}

	if !valid {
		h.errorJSON(w, http.StatusBadRequest, "invalid code")
		return
	}

	h.json(w, http.StatusOK, map[string]bool{"valid": true})
}

// VerifyMFA (authenticated)
func (h *Handlers) VerifyMFAAuthenticated(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Try extracting userID from context
	var userID uuid.UUID
	if val := r.Context().Value(ctxkeys.UserIDKey); val != nil {
		if id, ok := val.(uuid.UUID); ok {
			userID = id
		} else if idStr, ok := val.(string); ok {
			userID, _ = uuid.Parse(idStr)
		}
	}

	if userID == uuid.Nil {
		h.errorJSON(w, http.StatusBadRequest, "user ID required")
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), userID)
	if err != nil || user.MfaSecret.String == "" {
		h.errorJSON(w, http.StatusBadRequest, "MFA secret not found")
		return
	}

	valid, err := h.svc.VerifyTOTP(user.MfaSecret.String, req.Code)
	if err != nil || !valid {
		h.errorJSON(w, http.StatusUnauthorized, "invalid MFA code")
		return
	}

	h.json(w, http.StatusOK, map[string]bool{"valid": true})
}

func (h *Handlers) VerifyLoginMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TempToken string `json:"temp_token"`
		Code      string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.VerifyLoginMFA(r.Context(), req.TempToken, req.Code, "", "")
	if err != nil {
		h.logger.Error("MFA login verification failed", zap.Error(err))
		h.errorJSON(w, http.StatusUnauthorized, "invalid MFA code or token")
		return
	}

	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) SetupAdminMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TempToken string `json:"temp_token"`
		Secret    string `json:"secret"`
		Code      string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.SetupAdminMFA(r.Context(), req.TempToken, req.Secret, req.Code, "", "")
	if err != nil {
		h.logger.Error("MFA setup failed", zap.Error(err))
		h.errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	h.json(w, http.StatusOK, resp)
}

// Password Reset Handlers

func (h *Handlers) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.RequestPasswordReset(r.Context(), req.Email); err != nil {
		h.logger.Error("password reset request failed", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to process request")
		return
	}

	// Always return success to prevent email enumeration
	h.json(w, http.StatusOK, map[string]string{
		"message": "If an account exists with this email, a password reset link has been sent",
	})
}

func (h *Handlers) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		h.logger.Error("password reset failed", zap.Error(err))

		var ve authservice.ValidationError
		if errors.As(err, &ve) {
			h.errorJSON(w, http.StatusBadRequest, ve.Error())
			return
		}

		h.errorJSON(w, http.StatusBadRequest, "invalid or expired reset token")
		return
	}

	h.json(w, http.StatusOK, map[string]string{"status": "password_reset"})
}

func (h *Handlers) ResolveOrgs(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	orgs, err := h.svc.ResolveOrgs(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("failed to resolve orgs", zap.Error(err), zap.String("email", req.Email))
		// Return generic success to prevent enumeration as per requirement
		h.json(w, http.StatusOK, map[string]any{"organizations": []any{}})
		return
	}

	type orgResp struct {
		OrgID   uuid.UUID `json:"org_id"`
		OrgName string    `json:"org_name"`
		OrgLogo string    `json:"org_logo"`
	}

	result := make([]orgResp, 0, len(orgs))
	for _, o := range orgs {
		result = append(result, orgResp{
			OrgID:   o.ID,
			OrgName: o.Name,
			OrgLogo: o.LogoURL,
		})
	}

	h.json(w, http.StatusOK, map[string]any{"organizations": result})
}

// Session Handlers

func (h *Handlers) GetSessions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	sessions, err := h.svc.ListUserSessions(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to list sessions", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to get sessions")
		return
	}

	h.json(w, http.StatusOK, sessions)
}

func (h *Handlers) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid user id")
		return
	}

	sessionID, err := uuid.Parse(chi.URLParam(r, "session_id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid session id")
		return
	}

	if err := h.svc.RevokeUserSession(r.Context(), sessionID, userID); err != nil {
		h.logger.Error("failed to revoke session", zap.Error(err))
		h.errorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]string{"status": "session_revoked"})
}

func (h *Handlers) LoginOIDC(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	var req struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.LoginWithOIDC(r.Context(), provider, req.Code, req.State, "", "")
	if err != nil {
		h.logger.Error("OIDC login failed", zap.Error(err), zap.String("provider", provider))
		h.errorJSON(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.json(w, http.StatusOK, resp)
}

func (h *Handlers) GetJWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := h.svc.GetJWKS(r.Context())
	if err != nil {
		h.logger.Error("failed to get JWKS", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to fetch JWKS")
		return
	}
	h.json(w, http.StatusOK, jwks)
}

func (h *Handlers) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
	if !ok {
		// Fallback to legacy string if needed, but primary should be UUID
		orgIDStr, okStr := r.Context().Value(ctxkeys.OrgIDKey).(string)
		if !okStr {
			h.errorJSON(w, http.StatusUnauthorized, "organization context missing")
			return
		}
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.errorJSON(w, http.StatusUnauthorized, "invalid organization id")
			return
		}
	}
	
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		userIDStr, okStr := r.Context().Value(ctxkeys.UserIDKey).(string)
		if !okStr {
			h.errorJSON(w, http.StatusUnauthorized, "user context missing")
			return
		}
		var err error
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			h.errorJSON(w, http.StatusUnauthorized, "invalid user id")
			return
		}
	}

	var req struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	key, err := h.svc.CreateAPIKey(r.Context(), userID, orgID, req.Name, req.ExpiresAt)
	if err != nil {
		h.logger.Error("failed to create api key", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	h.json(w, http.StatusCreated, map[string]string{"api_key": key})
}

func (h *Handlers) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
	if !ok {
		orgIDStr, okStr := r.Context().Value(ctxkeys.OrgIDKey).(string)
		if !okStr {
			h.errorJSON(w, http.StatusUnauthorized, "organization context missing")
			return
		}
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.errorJSON(w, http.StatusUnauthorized, "invalid organization id")
			return
		}
	}

	keys, err := h.svc.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list api keys", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to list api keys")
		return
	}

	h.json(w, http.StatusOK, keys)
}

func (h *Handlers) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid api key id")
		return
	}

	if err := h.svc.RevokeAPIKey(r.Context(), id); err != nil {
		h.logger.Error("failed to revoke api key", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to revoke api key")
		return
	}

	h.json(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (h *Handlers) CreateServiceAccount(w http.ResponseWriter, r *http.Request) {
	orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
	if !ok {
		orgIDStr, okStr := r.Context().Value(ctxkeys.OrgIDKey).(string)
		if !okStr {
			h.errorJSON(w, http.StatusUnauthorized, "organization context missing")
			return
		}
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.errorJSON(w, http.StatusUnauthorized, "invalid organization id")
			return
		}
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sa, secret, err := h.svc.CreateServiceAccount(r.Context(), orgID, req.Name, req.Description)
	if err != nil {
		h.logger.Error("failed to create service account", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to create service account")
		return
	}

	h.json(w, http.StatusCreated, map[string]any{
		"service_account": sa,
		"client_secret":   secret, // Only shown once
	})
}

func (h *Handlers) ListServiceAccounts(w http.ResponseWriter, r *http.Request) {
	orgID, ok := r.Context().Value(middleware.OrgIDKey).(uuid.UUID)
	if !ok {
		orgIDStr, okStr := r.Context().Value(ctxkeys.OrgIDKey).(string)
		if !okStr {
			h.errorJSON(w, http.StatusUnauthorized, "organization context missing")
			return
		}
		var err error
		orgID, err = uuid.Parse(orgIDStr)
		if err != nil {
			h.errorJSON(w, http.StatusUnauthorized, "invalid organization id")
			return
		}
	}

	accounts, err := h.svc.ListServiceAccounts(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to list service accounts", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to list service accounts")
		return
	}

	h.json(w, http.StatusOK, accounts)
}

func (h *Handlers) EmergencyRevoke(w http.ResponseWriter, r *http.Request) {
	// Only superadmins should be able to do this, handled by RBAC middleware ideally
	if err := h.svc.EmergencyRevoke(r.Context()); err != nil {
		h.logger.Error("emergency revoke failed", zap.Error(err))
		h.errorJSON(w, http.StatusInternalServerError, "failed to trigger emergency revoke")
		return
	}

	h.json(w, http.StatusOK, map[string]string{"status": "emergency_revoke_triggered"})
}
