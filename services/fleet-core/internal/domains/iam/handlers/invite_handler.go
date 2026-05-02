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

// InviteHandler handles invitation HTTP endpoints.
type InviteHandler struct {
	inviteSvc invitationService
	log       *zap.Logger
}

func NewInviteHandler(inviteSvc invitationService, log *zap.Logger) *InviteHandler {
	return &InviteHandler{
		inviteSvc: inviteSvc,
		log:       log,
	}
}

// CreateInvite creates a new user invitation.
// POST /api/v1/admin/users/invite
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	var req models.InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Email == "" {
		h.writeError(w, http.StatusBadRequest, "missing_email", "Email is required", r)
		return
	}

	tenantID := GetTenantUUID(r)
	invitedBy := GetUserIDUUID(r)

	if tenantID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusBadRequest, "missing_tenant", "Tenant context is required", r)
		return
	}

	invite, err := h.inviteSvc.CreateInvite(r.Context(), tenantID, invitedBy, req.Email, req.RoleIDs, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicateInvite):
			h.writeError(w, http.StatusConflict, "duplicate_invite", "An invitation for this email already exists", r)
		default:
			h.log.Error("failed to create invite", zap.Error(err), zap.String("email", req.Email))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create invitation", r)
		}
		return
	}

	resp := models.InviteResponse{
		ID:        invite.ID,
		Email:     invite.Email,
		Status:    string(invite.Status),
		ExpiresAt: invite.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		InvitedBy: invite.InvitedBy,
		InvitedAt: invite.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	h.writeJSON(w, http.StatusCreated, resp)
}

// ResendInvite resends a pending invitation.
// POST /api/v1/admin/users/invite/{id}/resend
func (h *InviteHandler) ResendInvite(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	inviteID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid invitation ID format", r)
		return
	}

	actorID := GetUserIDUUID(r)

	if err := h.inviteSvc.ResendInvite(r.Context(), inviteID, actorID); err != nil {
		switch {
		case errors.Is(err, service.ErrInviteNotFound):
			h.writeError(w, http.StatusNotFound, "invite_not_found", "Invitation not found", r)
		default:
			h.log.Error("failed to resend invite", zap.Error(err), zap.String("invite_id", idStr))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to resend invitation", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "invite_resent"})
}

// RevokeInvite revokes a pending invitation.
// DELETE /api/v1/admin/users/invite/{id}
func (h *InviteHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	inviteID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid invitation ID format", r)
		return
	}

	actorID := GetUserIDUUID(r)

	if err := h.inviteSvc.RevokeInvite(r.Context(), inviteID, actorID); err != nil {
		switch {
		case errors.Is(err, service.ErrInviteNotFound):
			h.writeError(w, http.StatusNotFound, "invite_not_found", "Invitation not found", r)
		default:
			h.log.Error("failed to revoke invite", zap.Error(err), zap.String("invite_id", idStr))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke invitation", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "invite_revoked"})
}

// ValidateInviteToken validates an invitation token and returns invite details.
// GET /api/v1/auth/invite/{token}
func (h *InviteHandler) ValidateInviteToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.writeError(w, http.StatusBadRequest, "missing_token", "Invitation token is required", r)
		return
	}

	invite, err := h.inviteSvc.ValidateInviteToken(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteNotFound):
			h.writeError(w, http.StatusNotFound, "invite_not_found", "Invalid or expired invitation", r)
		case errors.Is(err, service.ErrInviteExpired):
			h.writeError(w, http.StatusGone, "invite_expired", "Invitation has expired", r)
		case errors.Is(err, service.ErrInviteAlreadyUsed):
			h.writeError(w, http.StatusConflict, "invite_already_used", "Invitation has already been used", r)
		case errors.Is(err, service.ErrInviteRevoked):
			h.writeError(w, http.StatusGone, "invite_revoked", "Invitation has been revoked", r)
		default:
			h.log.Error("failed to validate invite token", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to validate invitation", r)
		}
		return
	}

	// Return public-safe invite info (no token)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"email":      invite.Email,
		"tenant_id":     invite.TenantID,
		"message":    invite.Message,
		"expires_at": invite.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

// AcceptInvite accepts an invitation and creates the user account.
// POST /api/v1/auth/invite/accept
func (h *InviteHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req models.InviteAcceptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Token == "" || req.Name == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "token, name, and password are required", r)
		return
	}

	user, err := h.inviteSvc.AcceptInvite(r.Context(), req.Token, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteNotFound):
			h.writeError(w, http.StatusNotFound, "invite_not_found", "Invalid or expired invitation", r)
		case errors.Is(err, service.ErrInviteExpired):
			h.writeError(w, http.StatusGone, "invite_expired", "Invitation has expired", r)
		case errors.Is(err, service.ErrInviteAlreadyUsed):
			h.writeError(w, http.StatusConflict, "invite_already_used", "Invitation has already been used", r)
		case errors.Is(err, service.ErrInviteRevoked):
			h.writeError(w, http.StatusGone, "invite_revoked", "Invitation has been revoked", r)
		default:
			h.log.Error("failed to accept invite", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to accept invitation", r)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "invite_accepted",
		"user_id": user.ID,
		"email":   user.Email,
	})
}

func (h *InviteHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *InviteHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
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
