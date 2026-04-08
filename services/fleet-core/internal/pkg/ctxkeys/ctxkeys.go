package ctxkeys

import "context"

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	OrgIDKey          contextKey = "org_id"
	TenantIDKey       contextKey = "tenant_id"
	RoleKey           contextKey = "role"
	PermissionMaskKey contextKey = "permission_mask"
)

// WithOrgID returns a new context with the org ID set.
func WithOrgID(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, OrgIDKey, orgID)
}
