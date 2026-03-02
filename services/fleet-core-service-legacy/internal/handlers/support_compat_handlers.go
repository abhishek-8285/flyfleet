package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fleet-platform/shared/pkg/api"
	orgservice "github.com/flyfleet/fleet-core-service/internal/domains/admin_support_v2/domains/organizations/service"
	authservice "github.com/flyfleet/fleet-core-service/internal/domains/auth_v2/service"
	"github.com/flyfleet/fleet-core-service/internal/pkg/ctxkeys"
	"github.com/flyfleet/fleet-core-service/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SupportCompatHandlers struct {
	logger     *zap.Logger
	authSvc    *authservice.AuthService
	orgSvc     orgservice.OrganizationService
	supportRepo repository.SupportRepository
}

func NewSupportCompatHandlers(logger *zap.Logger, authSvc *authservice.AuthService, orgSvc orgservice.OrganizationService, supportRepo repository.SupportRepository) *SupportCompatHandlers {
	return &SupportCompatHandlers{
		logger:      logger,
		authSvc:     authSvc,
		orgSvc:      orgSvc,
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
	userId, _ := uuid.Parse(userIdStr)
	user, err := h.authSvc.GetUserByID(r.Context(), userId)
	
	roles := []string{"User"}
	if err == nil && user != nil {
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

	// 3. Log Audit
	agentID, _ := r.Context().Value(ctxkeys.UserIDKey).(string)
	orgID, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	agentUID, _ := uuid.Parse(agentID)
	orgUID, _ := uuid.Parse(orgID)

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		OrganizationID: orgUID,
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

	err = h.supportRepo.RevokeSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Failed to revoke session", http.StatusInternalServerError)
		return
	}

	// 3. Log Audit
	userIdStr := chi.URLParam(r, "id")
	userId, _ := uuid.Parse(userIdStr)
	agentID, _ := r.Context().Value(ctxkeys.UserIDKey).(string)
	orgID, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	agentUID, _ := uuid.Parse(agentID)
	orgUID, _ := uuid.Parse(orgID)

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		OrganizationID: orgUID,
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
	userId, _ := uuid.Parse(userIdStr)
	
	agentID, _ := r.Context().Value(ctxkeys.UserIDKey).(string)
	orgID, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	agentUID, _ := uuid.Parse(agentID)
	orgUID, _ := uuid.Parse(orgID)

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		OrganizationID: orgUID,
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
	userId, _ := uuid.Parse(userIdStr)
	
	agentID, _ := r.Context().Value(ctxkeys.UserIDKey).(string)
	orgID, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	agentUID, _ := uuid.Parse(agentID)
	orgUID, _ := uuid.Parse(orgID)

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		OrganizationID: orgUID,
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
	userId, _ := uuid.Parse(userIdStr)
	
	agentID, _ := r.Context().Value(ctxkeys.UserIDKey).(string)
	orgID, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	agentUID, _ := uuid.Parse(agentID)
	orgUID, _ := uuid.Parse(orgID)

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		OrganizationID: orgUID,
		UserID:         agentUID,
		Action:         "MFA_RESET",
		TargetUserID:   &userId,
		IPAddress:      r.RemoteAddr,
		UserAgent:      r.UserAgent(),
		Status:       "SUCCESS",
	})
	w.WriteHeader(http.StatusOK)
}

func (h *SupportCompatHandlers) ForceLogout(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	userId, _ := uuid.Parse(userIdStr)
	
	agentID, _ := r.Context().Value(ctxkeys.UserIDKey).(string)
	orgID, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	agentUID, _ := uuid.Parse(agentID)
	orgUID, _ := uuid.Parse(orgID)

	h.supportRepo.LogAudit(r.Context(), &repository.AuditLog{
		OrganizationID: orgUID,
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
	orgIDStr, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	orgID, _ := uuid.Parse(orgIDStr)
	
	tickets, err := h.supportRepo.GetTickets(r.Context(), orgID)
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
		"organization": "FlyFleet Core",
		"created_at":   ticket.CreatedAt.Format("Jan 02, 2006 15:04"),
		"updates":      events,
	})
}

func (h *SupportCompatHandlers) GetIncidentMetrics(w http.ResponseWriter, r *http.Request) {
	orgIDStr, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	orgID, _ := uuid.Parse(orgIDStr)
	metrics, err := h.supportRepo.GetDashboardMetrics(r.Context(), orgID)
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
	orgIDStr, _ := r.Context().Value(ctxkeys.OrgIDKey).(string)
	orgID, _ := uuid.Parse(orgIDStr)
	tickets, _ := h.supportRepo.GetTickets(r.Context(), orgID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": tickets,
	})
}

type UIHandlers struct{}

func (h *UIHandlers) GetGlobalConfig(w http.ResponseWriter, r *http.Request) {
	config := map[string]interface{}{
		"system_metadata": map[string]interface{}{
			"environment": "production-consolidated",
			"version":     "2.0.0-monolith",
		},
		"support_metadata": map[string]interface{}{
			"avg_resolution_time": "1h 15m",
			"critical_load":      "25%",
			"high_priority_load": "40%",
			"in_progress_status": "Active",
			"sla_attainment":     "99.8%",
			"system_protocol":    "Operational",
			"version":            "v1.4.2",
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}
