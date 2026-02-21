package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyfleet/auth-service/internal/middleware"
	"github.com/flyfleet/auth-service/internal/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RBACOperationsHandlers struct {
	service *service.AuthService
	logger  *zap.Logger
}

func NewRBACOperationsHandlers(s *service.AuthService, logger *zap.Logger) *RBACOperationsHandlers {
	return &RBACOperationsHandlers{
		service: s,
		logger:  logger,
	}
}

// GET /api/v1/admin/rbac/roles
func (h *RBACOperationsHandlers) ListRoles(w http.ResponseWriter, r *http.Request) {
	orgID := r.Context().Value("org_id").(uuid.UUID)
	roles, err := h.service.ListRolesByOrg(r.Context(), orgID)
	if err != nil {
		h.logger.Error("Failed to list roles", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, roles, 1, len(roles), len(roles))
}

// POST /api/v1/admin/rbac/roles
func (h *RBACOperationsHandlers) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	orgID := r.Context().Value("org_id").(uuid.UUID)
	userID := r.Context().Value("user_id").(uuid.UUID)
	role, err := h.service.CreateRole(r.Context(), req.Name, req.Description, orgID, false, userID)
	if err != nil {
		h.logger.Error("Failed to create role", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, role, 1, 1, 1)
}

// PUT /api/v1/admin/rbac/roles/{id}
func (h *RBACOperationsHandlers) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/rbac/roles/")
	if idStr == "" {
		http.Error(w, `{"error": "Missing ID"}`, http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	role, err := h.service.UpdateRole(r.Context(), id, req.Name, req.Description, userID)
	if err != nil {
		h.logger.Error("Failed to update role", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, role, 1, 1, 1)
}

// DELETE /api/v1/admin/rbac/roles/{id}
func (h *RBACOperationsHandlers) DeleteRole(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/rbac/roles/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	if err := h.service.DeleteRole(r.Context(), id, userID); err != nil {
		h.logger.Error("Failed to delete role", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/admin/rbac/permissions
func (h *RBACOperationsHandlers) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.service.ListPermissions(r.Context())
	if err != nil {
		h.logger.Error("Failed to list permissions", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, perms, 1, len(perms), len(perms))
}

// GET /api/v1/admin/rbac/roles/{id}/permissions
func (h *RBACOperationsHandlers) GetRolePermissions(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// Expected: /api/v1/admin/rbac/roles/{id}/permissions
	parts := strings.Split(path, "/")
	if len(parts) < 7 {
		http.Error(w, `{"error": "Invalid path"}`, http.StatusBadRequest)
		return
	}
	idStr := parts[6]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	perms, err := h.service.GetRolePermissions(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get role permissions", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, perms, 1, len(perms), len(perms))
}

// PUT /api/v1/admin/rbac/roles/{id}/permissions
func (h *RBACOperationsHandlers) UpdateRolePermissions(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 7 {
		http.Error(w, `{"error": "Invalid path"}`, http.StatusBadRequest)
		return
	}
	idStr := parts[6]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		PermissionIDs []uuid.UUID `json:"permission_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	if err := h.service.UpdateRolePermissions(r.Context(), id, req.PermissionIDs, userID); err != nil {
		h.logger.Error("Failed to update role permissions", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, nil, 1, 1, 1)
}

// POST /api/v1/admin/rbac/users/assign
func (h *RBACOperationsHandlers) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uuid.UUID `json:"user_id"`
		RoleID uuid.UUID `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	adminID := r.Context().Value("user_id").(uuid.UUID)
	if err := h.service.AssignRoleToUser(r.Context(), req.UserID, req.RoleID, adminID); err != nil {
		h.logger.Error("Failed to assign role", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, nil, 1, 1, 1)
}

// DELETE /api/v1/admin/rbac/users/unassign
func (h *RBACOperationsHandlers) UnassignRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uuid.UUID `json:"user_id"`
		RoleID uuid.UUID `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID) // Admin ID
	if err := h.service.RemoveRoleFromUser(r.Context(), req.UserID, req.RoleID, userID); err != nil {
		h.logger.Error("Failed to unassign role", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/admin/rbac/users/{id}/roles
func (h *RBACOperationsHandlers) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/rbac/users/")
	idStr = strings.TrimSuffix(idStr, "/roles")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	roles, err := h.service.GetUserRoles(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get user roles", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, roles, 1, len(roles), len(roles))
}

// POST /api/v1/admin/rbac/roles/{id}/clone
func (h *RBACOperationsHandlers) CloneRole(w http.ResponseWriter, r *http.Request) {
	// Pattern: /api/v1/admin/rbac/roles/{id}/clone
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 7 {
		http.Error(w, `{"error": "Invalid path"}`, http.StatusBadRequest)
		return
	}
	idStr := parts[6]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid body"}`, http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(uuid.UUID)
	role, err := h.service.CloneRole(r.Context(), id, req.Name, req.Description, userID)
	if err != nil {
		h.logger.Error("Failed to clone role", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, role, 1, 1, 1)
}

// GET /api/v1/admin/rbac/users/{id}/permissions
func (h *RBACOperationsHandlers) GetUserEffectivePermissions(w http.ResponseWriter, r *http.Request) {
	// Pattern: /api/v1/admin/rbac/users/{id}/permissions
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 7 {
		http.Error(w, `{"error": "Invalid path"}`, http.StatusBadRequest)
		return
	}
	idStr := parts[6]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	perms, err := h.service.GetEffectivePermissions(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get effective permissions", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	middleware.WriteSuccess(w, perms, 1, len(perms), len(perms))
}
