package ctxkeys

import (
	"context"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	TenantIDKey       contextKey = "tenant_id"
	RoleKey           contextKey = "role"
	RolesKey          contextKey = "user_roles"
	PermissionMaskKey contextKey = "permission_mask"
)

// GetUserID safely extracts UserID from context, handling string and UUID types
// across both the standardized internal key and the shared platform key.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	keys := []any{UserIDKey, middleware.UserIDKey, "user_id"}
	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if id, ok := val.(uuid.UUID); ok && id != uuid.Nil {
				return id, true
			}
			if s, ok := val.(string); ok && s != "" {
				if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
					return id, true
				}
			}
		}
	}
	return uuid.Nil, false
}

// GetTenantID safely extracts TenantID from context, handling string and UUID types
// across both the standardized internal key and the shared platform key.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	keys := []any{TenantIDKey, middleware.TenantIDKey, "tenant_id", "org_id"}
	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if id, ok := val.(uuid.UUID); ok && id != uuid.Nil {
				return id, true
			}
			if s, ok := val.(string); ok && s != "" {
				if id, err := uuid.Parse(s); err == nil && id != uuid.Nil {
					return id, true
				}
			}
		}
	}
	return uuid.Nil, false
}

// GetRole extracts the user role from context.
func GetRole(ctx context.Context) string {
	keys := []any{RoleKey, middleware.RoleKey}
	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}
	return ""
}

// WithTenantID returns a new context with the tenant ID set.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// WithContextValue returns a new context with the given key and value set.
func WithContextValue(ctx context.Context, key any, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetPermissionMask safely extracts the permission mask from context.
func GetPermissionMask(ctx context.Context) uint64 {
	keys := []any{PermissionMaskKey, "permissions_mask", "permission_mask"}
	for _, k := range keys {
		if val := ctx.Value(k); val != nil {
			if m, ok := val.(uint64); ok {
				return m
			}
			if m, ok := val.(int64); ok {
				return uint64(m)
			}
			if m, ok := val.(int); ok {
				return uint64(m)
			}
			if f, ok := val.(float64); ok {
				return uint64(f)
			}
		}
	}
	return 0
}

