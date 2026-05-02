package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// SessionHandler handles session HTTP endpoints.
type SessionHandler struct {
	sessionSvc sessionService
	log        *zap.Logger
}

func NewSessionHandler(sessionSvc sessionService, log *zap.Logger) *SessionHandler {
	return &SessionHandler{
		sessionSvc: sessionSvc,
		log:        log,
	}
}

// ListSessions returns all active sessions for the authenticated user.
// GET /api/v1/auth/sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	tenantID := GetTenantUUID(r)

	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	sessions, err := h.sessionSvc.ListSessions(r.Context(), userID, tenantID)
	if err != nil {
		h.log.Error("failed to list sessions", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list sessions", r)
		return
	}

	if sessions == nil {
		sessions = []models.Session{}
	}

	h.writeJSON(w, http.StatusOK, sessions)
}

// RevokeSession revokes a single session.
// DELETE /api/v1/auth/sessions/{id}
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		h.writeError(w, http.StatusBadRequest, "missing_id", "Session ID is required", r)
		return
	}

	userID := GetUserIDUUID(r)
	tenantID := GetTenantUUID(r)

	if err := h.sessionSvc.RevokeSession(r.Context(), userID, tenantID, sessionID); err != nil {
		h.log.Error("failed to revoke session", zap.Error(err), zap.String("session_id", sessionID))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke session", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "session_revoked"})
}

// RevokeAllSessions revokes all sessions for the authenticated user.
// DELETE /api/v1/auth/sessions/all
func (h *SessionHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	if err := h.sessionSvc.RevokeAllSessions(r.Context(), userID); err != nil {
		h.log.Error("failed to revoke all sessions", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke all sessions", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "all_sessions_revoked"})
}

func (h *SessionHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *SessionHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
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
