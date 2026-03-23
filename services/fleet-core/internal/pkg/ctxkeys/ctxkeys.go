package ctxkeys

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	OrgIDKey          contextKey = "org_id"
	TenantIDKey       contextKey = "tenant_id"
	RoleKey           contextKey = "role"
	PermissionMaskKey contextKey = "permission_mask"
)
