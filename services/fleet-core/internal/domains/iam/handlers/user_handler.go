package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core/internal/domains/iam/models"
	"github.com/flyfleet/fleet-core/internal/domains/iam/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserHandler handles admin user management HTTP endpoints.
type UserHandler struct {
	userSvc userService
	log     *zap.Logger
}

func NewUserHandler(userSvc userService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		log:     log,
	}
}

// ListUsers returns a paginated list of users.
// GET /api/v1/admin/users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	params := models.ListUsersParams{
		TenantID: GetTenantUUID(r),
		Page:     1,
		PageSize: 20,
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			params.Page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			params.PageSize = v
		}
	}
	if search := r.URL.Query().Get("search"); search != "" {
		params.Search = search
	}
	if status := r.URL.Query().Get("status"); status != "" {
		params.Status = status
	}
	if roleID := r.URL.Query().Get("role_id"); roleID != "" {
		if id, err := uuid.Parse(roleID); err == nil {
			params.RoleID = &id
		}
	}

	resp, err := h.userSvc.ListUsers(r.Context(), params)
	if err != nil {
		h.log.Error("failed to list users", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list users", r)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// CreateUser handles direct user creation by an admin.
// POST /api/v1/admin/users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Email == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "Email and password are required", r)
		return
	}

	tenantID := GetTenantUUID(r)
	actorID := GetUserIDUUID(r)

	user, err := h.userSvc.CreateUser(r.Context(), tenantID, req, actorID)
	if err != nil {
		h.log.Error("failed to create user", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create user", r)
		return
	}

	h.writeJSON(w, http.StatusCreated, user)
}

// GetUser returns a single user by ID.
// GET /api/v1/admin/users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	user, err := h.userSvc.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			h.writeError(w, http.StatusNotFound, "user_not_found", "User not found", r)
			return
		}
		h.log.Error("failed to get user", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get user", r)
		return
	}

	h.writeJSON(w, http.StatusOK, user)
}

// UpdateStatus updates a user's status (active, disabled, locked).
// PUT /api/v1/admin/users/{id}/status
func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	var req models.UpdateUserStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Status == "" {
		h.writeError(w, http.StatusBadRequest, "missing_status", "Status is required", r)
		return
	}

	actorID := GetUserIDUUID(r)
	tenantID := GetTenantUUID(r)

	if err := h.userSvc.UpdateUserStatus(r.Context(), userID, tenantID, req.Status, actorID); err != nil {
		if errors.Is(err, service.ErrInvalidStatus) {
			h.writeError(w, http.StatusBadRequest, "invalid_status", "Valid statuses: active, disabled, locked", r)
			return
		}
		h.log.Error("failed to update user status", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update user status", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "user_status_updated"})
}

// DeleteUser soft-deletes a user.
// DELETE /api/v1/admin/users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	actorID := GetUserIDUUID(r)
	tenantID := GetTenantUUID(r)

	if err := h.userSvc.DeleteUser(r.Context(), userID, tenantID, actorID); err != nil {
		h.log.Error("failed to delete user", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete user", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "user_deleted"})
}

// UpdatePreferences updates user UI preferences (sidebar state, language).
// PUT /api/v1/admin/users/{id}/preferences
func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	var req models.UpdateUserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.SidebarCollapsed == nil && req.Language == nil {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "At least one preference field must be provided", r)
		return
	}

	actorID := GetUserIDUUID(r)
	tenantID := GetTenantUUID(r)

	if err := h.userSvc.UpdateUserPreferences(r.Context(), userID, tenantID, req.SidebarCollapsed, req.Language, actorID); err != nil {
		h.log.Error("failed to update user preferences", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update user preferences", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "preferences_updated"})
}

// GetUserAuditLog returns the audit log for a specific user.
// GET /api/v1/admin/users/{id}/audit-log
func (h *UserHandler) GetUserAuditLog(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	logs, err := h.userSvc.GetUserAuditLog(r.Context(), userID, limit)
	if err != nil {
		h.log.Error("failed to get user audit log", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get audit log", r)
		return
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	h.writeJSON(w, http.StatusOK, logs)
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *UserHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
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
