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

// RBACHandler handles RBAC HTTP endpoints.
type RBACHandler struct {
	rbacSvc rbacService
	log     *zap.Logger
}

func NewRBACHandler(rbacSvc rbacService, log *zap.Logger) *RBACHandler {
	return &RBACHandler{
		rbacSvc: rbacSvc,
		log:     log,
	}
}

// ListRoles returns all roles for the current organization.
// GET /api/v1/admin/rbac/roles
func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	tenantID := GetTenantUUID(r)

	roles, err := h.rbacSvc.ListRoles(r.Context(), tenantID)
	if err != nil {
		h.log.Error("failed to list roles", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list roles", r)
		return
	}

	if roles == nil {
		roles = []models.Role{}
	}

	h.writeJSON(w, http.StatusOK, roles)
}

// CreateRole creates a new role.
// POST /api/v1/admin/rbac/roles
func (h *RBACHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "missing_name", "Role name is required", r)
		return
	}

	tenantID := GetTenantUUID(r)
	createdBy := GetUserIDUUID(r)

	// Parse permissions from query string or body
	// For simplicity, read from query param
	permStrs := r.URL.Query()["permissions"]

	role, err := h.rbacSvc.CreateRole(r.Context(), tenantID, req.Name, req.Description, permStrs, createdBy)
	if err != nil {
		h.log.Error("failed to create role", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to create role", r)
		return
	}

	h.writeJSON(w, http.StatusCreated, role)
}

// UpdateRole updates an existing role.
// PUT /api/v1/admin/rbac/roles/{id}
func (h *RBACHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID format", r)
		return
	}

	var req models.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	tenantID := GetTenantUUID(r)
	if err := h.rbacSvc.UpdateRole(r.Context(), roleID, tenantID, req.Name, req.Description); err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			h.writeError(w, http.StatusNotFound, "role_not_found", "Role not found", r)
		case errors.Is(err, service.ErrRoleIsSystem):
			h.writeError(w, http.StatusBadRequest, "role_is_system", "Cannot modify system roles", r)
		default:
			h.log.Error("failed to update role", zap.Error(err), zap.String("role_id", idStr))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update role", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "role_updated"})
}

// CloneRole creates a copy of an existing role.
// POST /api/v1/admin/rbac/roles/{id}/clone
func (h *RBACHandler) CloneRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	sourceID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID format", r)
		return
	}

	var req models.CloneRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "missing_name", "New role name is required", r)
		return
	}

	createdBy := GetUserIDUUID(r)
	tenantID := GetTenantUUID(r)

	role, err := h.rbacSvc.CloneRole(r.Context(), sourceID, tenantID, req.Name, req.Description, createdBy)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			h.writeError(w, http.StatusNotFound, "role_not_found", "Source role not found", r)
		default:
			h.log.Error("failed to clone role", zap.Error(err), zap.String("role_id", idStr))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to clone role", r)
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, role)
}

// DeleteRole deletes a role.
// DELETE /api/v1/admin/rbac/roles/{id}
func (h *RBACHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID format", r)
		return
	}

	tenantID := GetTenantUUID(r)
	if err := h.rbacSvc.DeleteRole(r.Context(), roleID, tenantID); err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			h.writeError(w, http.StatusNotFound, "role_not_found", "Role not found", r)
		case errors.Is(err, service.ErrRoleIsSystem):
			h.writeError(w, http.StatusBadRequest, "role_is_system", "Cannot delete system roles", r)
		default:
			h.log.Error("failed to delete role", zap.Error(err), zap.String("role_id", idStr))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete role", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "role_deleted"})
}

// GetPermissionMatrix returns the full permission matrix for all roles.
// GET /api/v1/admin/rbac/matrix
func (h *RBACHandler) GetPermissionMatrix(w http.ResponseWriter, r *http.Request) {
	tenantID := GetTenantUUID(r)

	matrix, err := h.rbacSvc.GetPermissionMatrix(r.Context(), tenantID)
	if err != nil {
		h.log.Error("failed to get permission matrix", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get permission matrix", r)
		return
	}

	if matrix == nil {
		matrix = []models.RolePermissionMatrixEntry{}
	}

	h.writeJSON(w, http.StatusOK, models.PermissionMatrixResponse{Roles: matrix})
}

// UpdateRolePermissions updates a role's permissions.
// PUT /api/v1/admin/rbac/roles/{id}/permissions
func (h *RBACHandler) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid role ID format", r)
		return
	}

	var req models.UpdateRolePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	tenantID := GetTenantUUID(r)
	if err := h.rbacSvc.UpdateRolePermissions(r.Context(), roleID, tenantID, req.Permissions); err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			h.writeError(w, http.StatusNotFound, "role_not_found", "Role not found", r)
		case errors.Is(err, service.ErrRoleIsSystem):
			h.writeError(w, http.StatusBadRequest, "role_is_system", "Cannot modify system roles", r)
		default:
			h.log.Error("failed to update role permissions", zap.Error(err), zap.String("role_id", idStr))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to update role permissions", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "role_permissions_updated"})
}

// AssignRole assigns a role to a user.
// POST /api/v1/admin/rbac/users/assign
func (h *RBACHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req models.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.UserID.String() == "00000000-0000-0000-0000-000000000000" ||
		req.RoleID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "user_id and role_id are required", r)
		return
	}

	assignedBy := GetUserIDUUID(r)

	if err := h.rbacSvc.AssignRole(r.Context(), req.UserID, req.RoleID, req.TenantID, assignedBy); err != nil {
		switch {
		case errors.Is(err, service.ErrRoleNotFound):
			h.writeError(w, http.StatusNotFound, "role_not_found", "Role not found", r)
		default:
			h.log.Error("failed to assign role", zap.Error(err))
			h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to assign role", r)
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "role_assigned"})
}

// BulkAssignRole assigns roles to multiple users at once.
// POST /api/v1/admin/rbac/users/assign/bulk
func (h *RBACHandler) BulkAssignRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserIDs []uuid.UUID `json:"user_ids"`
		RoleIDs []uuid.UUID `json:"role_ids"`
		TenantID   uuid.UUID   `json:"tenant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if len(req.UserIDs) == 0 || len(req.RoleIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "user_ids and role_ids are required", r)
		return
	}

	assignedBy := GetUserIDUUID(r)

	result := h.rbacSvc.BulkAssignRoles(r.Context(), req.UserIDs, req.RoleIDs, req.TenantID, assignedBy)

	h.writeJSON(w, http.StatusOK, result)
}

// UnassignRole removes a role from a user.
// DELETE /api/v1/admin/rbac/users/unassign
func (h *RBACHandler) UnassignRole(w http.ResponseWriter, r *http.Request) {
	var req models.UnassignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", r)
		return
	}

	if req.UserID.String() == "00000000-0000-0000-0000-000000000000" ||
		req.RoleID.String() == "00000000-0000-0000-0000-000000000000" {
		h.writeError(w, http.StatusBadRequest, "missing_fields", "user_id and role_id are required", r)
		return
	}

	if err := h.rbacSvc.UnassignRole(r.Context(), req.UserID, req.RoleID, req.TenantID); err != nil {
		h.log.Error("failed to unassign role", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to unassign role", r)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "role_unassigned"})
}

// GetUserRoles returns all roles assigned to a user.
// GET /api/v1/admin/rbac/users/{id}/roles
func (h *RBACHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	tenantID := GetTenantUUID(r)

	roles, err := h.rbacSvc.GetUserRoles(r.Context(), userID, tenantID)
	if err != nil {
		h.log.Error("failed to get user roles", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get user roles", r)
		return
	}

	if roles == nil {
		roles = []models.Role{}
	}

	h.writeJSON(w, http.StatusOK, roles)
}

// GetUserPermissions returns the effective permissions for a user.
// GET /api/v1/admin/rbac/users/{id}/permissions
func (h *RBACHandler) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_id", "Invalid user ID format", r)
		return
	}

	tenantID := GetTenantUUID(r)

	perms, err := h.rbacSvc.GetUserPermissions(r.Context(), userID, tenantID)
	if err != nil {
		h.log.Error("failed to get user permissions", zap.Error(err), zap.String("user_id", idStr))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get user permissions", r)
		return
	}

	if perms == nil {
		perms = []string{}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":     userID,
		"permissions": perms,
	})
}

// ListAllPermissions returns all available permissions in the system.
// GET /api/v1/admin/rbac/permissions
func (h *RBACHandler) ListAllPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.rbacSvc.ListAllPermissions(r.Context())
	if err != nil {
		h.log.Error("failed to list permissions", zap.Error(err))
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list permissions", r)
		return
	}

	if perms == nil {
		perms = []models.Permission{}
	}

	h.writeJSON(w, http.StatusOK, perms)
}

func (h *RBACHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *RBACHandler) writeError(w http.ResponseWriter, status int, code string, message string, r *http.Request) {
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
