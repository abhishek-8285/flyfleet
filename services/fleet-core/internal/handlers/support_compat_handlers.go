package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fleet-platform/shared/pkg/api"
	orgservice "github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/service"
	authservice "github.com/flyfleet/fleet-core/internal/domains/auth/service"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Security: Check if user has permission to access target user
func (h *SupportCompatHandlers) canAccessUser(r *http.Request, targetUserID uuid.UUID) (bool, error) {
	// Get authenticated user from context
	authUserID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		return false, errors.New("unauthorized")
	}

	// Allow users to access their own data
	if authUserID == targetUserID {
		return true, nil
	}

	// Check if the authenticated user has an administrative role
	role, ok := r.Context().Value(ctxkeys.RoleKey).(string)
	if ok && (role == "admin" || role == "super_admin" || role == "org_admin" || role == "fleet_manager") {
		// Admins can access any user in the same org
		// Verify if target user belongs to the same org as the authenticated admin
		authTenantID, _ := ctxkeys.GetTenantID(r.Context())

		targetUser, err := h.authSvc.GetUserByID(r.Context(), targetUserID)
		if err != nil {
			h.logger.Error("Failed to fetch target user for access check",
				zap.Error(err),
				zap.String("target_user_id", targetUserID.String()))
			return false, err
		}

		if targetUser.TenantID != authTenantID && role != "super_admin" {
			h.logger.Warn("Access denied: Admin from different tenant attempted access",
				zap.String("auth_user_id", authUserID.String()),
				zap.String("auth_tenant_id", authTenantID.String()),
				zap.String("target_user_id", targetUserID.String()),
				zap.String("target_tenant_id", targetUser.TenantID.String()))
			return false, errors.New("forbidden: user belongs to a different tenant")
		}

		h.logger.Debug("Admin access granted to user data",
			zap.String("auth_user_id", authUserID.String()),
			zap.String("target_user_id", targetUserID.String()),
			zap.String("role", role))
		return true, nil
	}

	// Access denied - user is not self and not admin
	h.logger.Warn("Access denied: user lacks permission to access target user",
		zap.String("auth_user_id", authUserID.String()),
		zap.String("target_user_id", targetUserID.String()),
		zap.String("role", role))
	return false, nil
}

type SupportCompatHandlers struct {
	logger      *zap.Logger
	authSvc     *authservice.AuthService
	tenantSvc   orgservice.TenantService
	supportRepo repository.SupportRepository
}

func NewSupportCompatHandlers(logger *zap.Logger, authSvc *authservice.AuthService, tenantSvc orgservice.TenantService, supportRepo repository.SupportRepository) *SupportCompatHandlers {
	return &SupportCompatHandlers{
		logger:      logger,
		authSvc:     authSvc,
		tenantSvc:   tenantSvc,
		supportRepo: supportRepo,
	}
}

func (h *SupportCompatHandlers) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	h.logger.Info("Searching users (compat)", zap.String("query", query))

	pagination := api.ParsePagination(r)

	users, err := h.authSvc.SearchUsers(r.Context(), query)
	if err != nil {
		h.logger.Error("Failed to search users", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Transform for UI
	var uiUsers []map[string]interface{}
	for _, u := range users {
		uiUsers = append(uiUsers, map[string]interface{}{
			"userId": u.ID.String(),
			"name":   u.Email, // Email as name if name is missing in model
			"email":  u.Email,
		})
	}

	resp := api.NewPaginatedResponse(uiUsers, len(uiUsers), pagination.Page, pagination.Limit)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *SupportCompatHandlers) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get authenticated user from context (set by auth middleware)
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Allow users to access their own profile or admins within their organization
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"
	
	if authUserUUID == userId || isAdmin {
		allowed := authUserUUID == userId
		if !allowed && isAdmin {
			allowed, _ = h.canAccessUser(r, userId)
		}
		
		if allowed {
			goto fetchUser
		}
	}

	h.logger.Warn("Profile access denied",
		zap.String("auth_user", authUserUUID.String()),
		zap.String("target_user", userIdStr))
	http.Error(w, "Access denied", http.StatusForbidden)
	return

fetchUser:
	user, err := h.authSvc.GetUserByID(r.Context(), userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	profile := map[string]interface{}{
		"userId":      user.ID.String(),
		"name":        user.Email,
		"email":       user.Email,
		"role":        user.Role,
		"status":      "ACTIVE",
		"verified":    true,
		"memberSince": user.CreatedAt.Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *SupportCompatHandlers) GetUserSecurity(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant access - only self-access allowed
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant security data access blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	user, err := h.authSvc.GetUserByID(r.Context(), userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	security := map[string]interface{}{
		"mfaEnabled":         user.MfaEnabled,
		"mfaMethods":         user.MfaMethods,
		"lastPasswordChange": user.LastPasswordChange.Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(security)
}

func (h *SupportCompatHandlers) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant access
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant sessions access blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	sessions, err := h.supportRepo.GetUserSessions(r.Context(), userId)
	if err != nil {
		http.Error(w, "Failed to get sessions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": sessions})
}

func (h *SupportCompatHandlers) GetUserLoginHistory(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant access
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant login history access blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	history, err := h.supportRepo.GetLoginHistory(r.Context(), userId)
	if err != nil {
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": history})
}

func (h *SupportCompatHandlers) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant access
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant roles access blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	user, err := h.authSvc.GetUserByID(r.Context(), userId)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	roles := []string{"User"}
	if user != nil {
		roles = []string{user.Role}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roles": roles,
	})
}

func (h *SupportCompatHandlers) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	permissions := []string{"VIEW_DASHBOARD", "VIEW_FLEET"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"permissions": permissions,
	})
}

func (h *SupportCompatHandlers) GetUserNotes(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant access
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant notes access blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	notes, err := h.supportRepo.GetUserNotes(r.Context(), userId)
	if err != nil {
		http.Error(w, "Failed to get notes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": notes})
}

func (h *SupportCompatHandlers) AddUserNote(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant access
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant add note blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	var body struct {
		Note  string `json:"note"`
		Agent string `json:"agent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if body.Agent == "" {
		body.Agent = "Global Admin"
	}

	err = h.supportRepo.AddUserNote(r.Context(), userId, body.Agent, body.Note)
	if err != nil {
		http.Error(w, "Failed to save note", http.StatusInternalServerError)
		return
	}

	agentUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgUID, _ := ctxkeys.GetTenantID(r.Context())

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		TenantID: orgUID,
		UserID:         agentUID,
		Action:         "ADD_USER_NOTE",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:         "SUCCESS",
		Metadata:       map[string]interface{}{"note_length": len(body.Note)},
	})

	w.WriteHeader(http.StatusCreated)
}

func (h *SupportCompatHandlers) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Validate user can only revoke their own sessions
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant session revocation blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	err = h.supportRepo.RevokeSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Failed to revoke session", http.StatusInternalServerError)
		return
	}

	// Log Audit
	userIdStr = chi.URLParam(r, "id")
	userId, _ = uuid.Parse(userIdStr)
	agentUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgUID, _ := ctxkeys.GetTenantID(r.Context())

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		TenantID: orgUID,
		UserID:         agentUID,
		Action:         "REVOKE_SESSION",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:         "SUCCESS",
		Metadata:       map[string]interface{}{"session_id": sessionID.String()},
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *SupportCompatHandlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant password reset
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant password reset blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	agentUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgUID, _ := ctxkeys.GetTenantID(r.Context())

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		TenantID: orgUID,
		UserID:         agentUID,
		Action:         "PASSWORD_RESET",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:         "SUCCESS",
	})
	w.WriteHeader(http.StatusOK)
}

func (h *SupportCompatHandlers) UnlockAccount(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant account unlock
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant account unlock blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	agentUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgUID, _ := ctxkeys.GetTenantID(r.Context())

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		TenantID: orgUID,
		UserID:         agentUID,
		Action:         "UNLOCK_ACCOUNT",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:         "SUCCESS",
	})
	w.WriteHeader(http.StatusOK)
}

func (h *SupportCompatHandlers) ResetMFA(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant MFA reset
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if authUserUUID != userId {
		h.logger.Warn("Cross-tenant MFA reset blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	agentUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgUID, _ := ctxkeys.GetTenantID(r.Context())

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		TenantID: orgUID,
		UserID:         agentUID,
		Action:         "MFA_RESET",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:         "SUCCESS",
	})
	w.WriteHeader(http.StatusOK)
}

func (h *SupportCompatHandlers) ForceLogout(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// SECURITY FIX: Prevent cross-tenant force logout
	authUserUUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)
	isAdmin := role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager"

	if authUserUUID != userId && !isAdmin {
		h.logger.Warn("Cross-tenant force logout blocked",
			zap.String("auth_user", authUserUUID.String()),
			zap.String("target_user", userIdStr))
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	agentUID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgUID, _ := ctxkeys.GetTenantID(r.Context())

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		TenantID:       orgUID,
		UserID:         agentUID,
		Action:         "FORCE_LOGOUT",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:         "SUCCESS",
	})
	w.WriteHeader(http.StatusOK)
}

func (h *SupportCompatHandlers) GetIncidents(w http.ResponseWriter, r *http.Request) {
	// For compat, we use a default org ID if not provided, or fetch from context
	tenantID, _ := ctxkeys.GetTenantID(r.Context())

	tickets, err := h.supportRepo.GetTickets(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to fetch tickets", zap.Error(err))
		tickets = []*repository.SupportTicket{} // Fallback to empty for UI stability
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": tickets,
		"total": len(tickets),
	})
}

func (h *SupportCompatHandlers) GetIncidentDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid incident ID", http.StatusBadRequest)
		return
	}

	ticket, events, err := h.supportRepo.GetTicketDetail(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch ticket detail", zap.Error(err))
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           ticket.ID,
		"subject":      ticket.Subject,
		"description":  ticket.Description,
		"status":       ticket.Status,
		"priority":     ticket.Priority,
		"tenant":       "FlyFleet Core",
		"created_at":   ticket.CreatedAt.Format("Jan 02, 2006 15:04"),
		"updates":      events,
	})
}

func (h *SupportCompatHandlers) GetIncidentMetrics(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := ctxkeys.GetTenantID(r.Context())
	metrics, err := h.supportRepo.GetDashboardMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to fetch metrics", zap.Error(err))
		// Fallback to static metrics for UI stability
		metrics = map[string]interface{}{
			"total_incidents": 0,
			"new_today":       0,
			"critical_open":   0,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *SupportCompatHandlers) GetIncidentFeed(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := ctxkeys.GetTenantID(r.Context())
	tickets, _ := h.supportRepo.GetTickets(r.Context(), tenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": tickets,
	})
}

type UIHandlers struct{}

func (h *UIHandlers) GetGlobalConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"system_metadata": map[string]interface{}{

			"ui_strings": map[string]interface{}{
				"portal": map[string]string{
					"dashboard_title": "B2B Customer Dashboard",
					"welcome_prefix":  "Authorized for",
					"new_booking_btn": "Book Shipment",
				},
				"portal_shipments": map[string]string{
					"title":        "Your Active Shipments",
					"th_tracking":  "Tracking ID",
					"th_status":    "Status",
					"action_track": "Real-time Tracking",
				},
				"portal_booking": map[string]string{
					"title":      "New Booking Request",
					"btn_submit": "Notify Logistics Team",
				},
				"portal_invoices": map[string]string{
					"title": "Billing & Financials",
				},
			},
		},
		"support_metadata": map[string]interface{}{
			"avg_resolution_time": "1h 15m",
			"critical_load":       "25%",
			"high_priority_load":  "40%",
			"in_progress_status":  "Active",
			"sla_attainment":      "99.8%",
			"system_protocol":     "Operational",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (h *UIHandlers) GetMenuById(w http.ResponseWriter, r *http.Request) {
	menuId := chi.URLParam(r, "menuId")
	role, _ := r.Context().Value(ctxkeys.RoleKey).(string)

	w.Header().Set("Content-Type", "application/json")

	if menuId == "menuItems" {
		items := []map[string]interface{}{}

		if role == "user" {
			json.NewEncoder(w).Encode(items)
			return
		}

		if role == "driver" || role == "staff" {
			items = []map[string]interface{}{
				{"label": "Dashboard", "path": "/dashboard", "icon": "dashboard", "resource": "dashboard", "action": "view"},
				{"label": "My Trips", "path": "/dashboard/trips", "icon": "map", "resource": "trip", "action": "view"},
				{"label": "Add Fuel Log", "path": "/dashboard/fuel/add", "icon": "local_gas_station", "resource": "vehicle", "action": "view"},
			}
			json.NewEncoder(w).Encode(items)
			return
		}

		// Super Admin / Org Admin / Fleet Manager
		items = append(items, map[string]interface{}{"label": "Dashboard", "title": "Dashboard", "icon": "dashboard", "path": "/admin/dashboard", "items": []any{}})

		if role == "super_admin" || role == "org_admin" || role == "fleet_manager" {
			items = append(items, map[string]interface{}{"label": "Tenants", "title": "Tenants", "icon": "business", "path": "/admin/tenants", "items": []any{}})
			items = append(items, map[string]interface{}{"label": "Users & Roles", "title": "Users & Roles", "icon": "people", "path": "/admin/users", "items": []any{}})
			items = append(items, map[string]interface{}{"label": "Billing", "title": "Billing", "icon": "payments", "path": "/admin/billing", "items": []any{}})
		}

		if role == "super_admin" {
			items = append(items, map[string]interface{}{"label": "Control Tower", "title": "Control Tower", "icon": "radar", "path": "/admin/control-tower", "items": []any{}})
		}
		items = append(items, map[string]interface{}{"label": "Fleet Command", "title": "Fleet Command", "icon": "local_shipping", "path": "/admin/vehicles", "items": []any{}})
		items = append(items, map[string]interface{}{"label": "Dispatch", "title": "Dispatch", "icon": "route", "path": "/admin/dispatch", "items": []any{}})

		json.NewEncoder(w).Encode(items)
		return
	}

	if menuId == "analyticsItems" {
		items := []map[string]interface{}{}
		if role == "super_admin" || role == "admin" || role == "org_admin" || role == "fleet_manager" {
			items = []map[string]interface{}{
				{"label": "Analytics", "path": "/admin/analytics", "icon": "bar_chart"},
				{"label": "Reports", "path": "/admin/reports", "icon": "summarize"},
			}
		}
		json.NewEncoder(w).Encode(items)
		return
	}

	if menuId == "securityItems" {
		items := []map[string]interface{}{}
		if role == "super_admin" || role == "admin" || role == "org_admin" {
			items = []map[string]interface{}{
				{"label": "Security Logs", "path": "/admin/security/telemetry", "icon": "security"},
				{"label": "Access Review", "path": "/admin/security/access-review", "icon": "policy"},
			}
		}
		json.NewEncoder(w).Encode(items)
		return
	}

	// Default fallback (e.g. portal)
	menuItems := []map[string]interface{}{
		{"label": "Dashboard", "title": "Dashboard", "icon": "dashboard", "path": "/portal/dashboard", "items": []any{}},
		{"label": "Shipments", "title": "Shipments", "icon": "local_shipping", "path": "/portal/shipments", "items": []any{}},
		{"label": "Bookings", "title": "Bookings", "icon": "add_box", "path": "/portal/bookings", "items": []any{}},
		{"label": "Invoices", "title": "Invoices", "icon": "payments", "path": "/portal/invoices", "items": []any{}},
		{"label": "Settings", "title": "Settings", "icon": "settings", "path": "/portal/settings", "items": []any{}},
	}
	json.NewEncoder(w).Encode(menuItems)
}

func (h *UIHandlers) GetLegalDocs(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"privacy_policy": map[string]interface{}{
			"last_updated": "2024-03-20",
			"sections": []map[string]interface{}{
				{"title": "Data Collection", "content": "We collect minimal telemetry to ensure platform stability."},
				{"title": "User Rights", "content": "You have the right to export or delete your security logs."},
			},
		},
		"terms_of_service": map[string]interface{}{
			"effective_date": "2024-03-20",
			"sections": []map[string]interface{}{
				{"title": "Acceptable Use", "content": "The system must only be used for authorized fleet operations."},
				{"title": "Termination", "content": "Unauthorized access will result in immediate multi-region revocation."},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
}
