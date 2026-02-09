package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyfleet/auth-service/internal/service"
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

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
	ua := r.UserAgent()

	resp, err := h.service.Login(r.Context(), req.Email, req.Password, ip, ua)
	if err != nil {
		h.logger.Error("Login failed", zap.Error(err), zap.String("ip", ip))
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	ip := r.RemoteAddr
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

func (h *AuthHandlers) SendMFA(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Method string `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	userID, _ := uuid.Parse(req.UserID)
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

	ip := r.RemoteAddr
	ua := r.UserAgent()

	userID, _ := uuid.Parse(req.UserID)
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

	ip := r.RemoteAddr
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

	ip := r.RemoteAddr
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

	uID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
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
	// TODO: Implement OIDC/SAML callback logic
	h.logger.Info("SSO Callback received", zap.String("provider", provider))
	w.WriteHeader(http.StatusOK)
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
	
	if err := h.service.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

func (h *AuthHandlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "use register for now", http.StatusNotImplemented)
}

func (h *AuthHandlers) GetUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}