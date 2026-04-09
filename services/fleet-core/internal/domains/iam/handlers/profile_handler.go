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

	// Fetch user's organizations (includes role per org)
	orgs, err := h.userRepo.GetOrganizations(r.Context(), userID)
	if err != nil {
		h.log.Warn("failed to get user organizations", zap.Error(err))
	} else {
		resp.Organizations = orgs
		// Set current_org to first org and extract role
		if len(orgs) > 0 {
			resp.CurrentOrg = &orgs[0]
			resp.Role = orgs[0].Role
		}
	}

	// Fallback role from JWT claims
	if resp.Role == "" {
		resp.Role = GetRole(r)
	}

	// Compute permissions based on role (admin gets everything, others get defaults)
	resp.Permissions = computePermissionsForRole(resp.Role)

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

// computePermissionsForRole returns a list of permission strings for the given role.
func computePermissionsForRole(role string) []string {
	switch role {
	case "admin", "super_admin", "org_admin":
		return []string{
			"vehicle:read", "vehicle:create", "vehicle:update", "vehicle:delete",
			"driver:read", "driver:create", "driver:update", "driver:delete",
			"trip:read", "trip:create", "trip:update", "trip:delete",
			"user:read", "user:create", "user:update", "user:delete",
			"system:read", "system:update",
			"reports:read", "reports:export",
			"settings:read", "settings:update",
			"alert:read", "alert:manage",
			"finance:read", "finance:manage",
		}
	case "fleet_manager":
		return []string{
			"vehicle:read", "vehicle:create", "vehicle:update",
			"driver:read", "driver:create", "driver:update",
			"trip:read", "trip:create", "trip:update",
			"user:read",
			"system:read",
			"reports:read", "reports:export",
			"alert:read", "alert:manage",
		}
	case "dispatcher":
		return []string{
			"vehicle:read",
			"driver:read",
			"trip:read", "trip:create", "trip:update",
			"user:read",
			"system:read",
			"reports:read",
			"alert:read",
		}
	default:
		return []string{
			"vehicle:read",
			"driver:read",
			"trip:read",
			"system:read",
			"reports:read",
		}
	}
}
