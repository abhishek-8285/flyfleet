package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"go.uber.org/zap"
)

// ProfileHandler handles profile HTTP endpoints.
type ProfileHandler struct {
	userRepo userRepo
	log      *zap.Logger
}

func NewProfileHandler(userRepo userRepo, log *zap.Logger) *ProfileHandler {
	return &ProfileHandler{
		userRepo: userRepo,
		log:      log,
	}
}

// GetProfile returns the authenticated user's profile.
// GET /api/v1/auth/profile
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to get user", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", r)
		return
	}

	if user == nil {
		h.writeError(w, http.StatusNotFound, "user_not_found", "User not found", r)
		return
	}

	resp := models.ProfileResponse{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		MFAEnabled:    user.MFAEnabled,
		EmailVerified: user.EmailVerified,
		Organizations: []models.OrgInfo{},
		Permissions:   []string{},
	}

	if user.LastLoginAt != nil {
		ts := user.LastLoginAt.Format("2006-01-02T15:04:05Z")
		resp.LastLoginAt = &ts
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// UpdateProfile updates the authenticated user's profile.
// PUT /api/v1/auth/profile
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := GetUserIDUUID(r)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusUnauthorized, "unauthenticated", "Authentication required", r)
		return
	}

	var req models.ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if err := h.userRepo.UpdateProfile(r.Context(), userID, req.Name, req.Phone, req.AvatarURL); err != nil {
		h.log.Error("failed to update profile", zap.Error(err), zap.String("user_id", userID.String()))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update profile", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "profile_updated"})
}

func (h *ProfileHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *ProfileHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
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
