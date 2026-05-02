package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
)

// Legacy constants for local reference, though we prefer ctxkeys now
const (
	CtxKeyUserID   = "user_id"
	CtxKeyTenantID = "tenant_id"
	CtxKeyRole     = "role"
	CtxKeyPermMask = "permissions_mask"
	CtxKeyTokenJTI = "token_jti"
)

func GetUserID(r *http.Request) string {
	id, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		return ""
	}
	return id.String()
}

func GetTenantID(r *http.Request) string {
	id, ok := ctxkeys.GetTenantID(r.Context())
	if !ok {
		return ""
	}
	return id.String()
}

func GetTenantUUID(r *http.Request) uuid.UUID {
	id, ok := ctxkeys.GetTenantID(r.Context())
	if !ok {
		return uuid.Nil
	}
	return id
}

func GetUserIDUUID(r *http.Request) uuid.UUID {
	id, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		return uuid.Nil
	}
	return id
}

func GetRole(r *http.Request) string {
	return ctxkeys.GetRole(r.Context())
}

func GetPermissionMask(r *http.Request) int64 {
	switch v := r.Context().Value(ctxkeys.PermissionMaskKey).(type) {
	case uint64:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func GetTokenJTI(r *http.Request) string {
	// JTI might still be in the context as a string if set by middleware
	v, _ := r.Context().Value(CtxKeyTokenJTI).(string)
	return v
}

func SetAuthClaims(r *http.Request, userID, tenantID, role string, permMask int64, jti string) *http.Request {
	ctx := r.Context()
	// Set both standardized keys and legacy string keys for safety during transition
	ctx = ctxkeys.WithTenantID(ctx, tenantID)
	// We use the raw values for the local constants to ensure internal IAM consistency
	ctx = ctxkeys.WithContextValue(ctx, ctxkeys.UserIDKey, userID)
	ctx = ctxkeys.WithContextValue(ctx, ctxkeys.RoleKey, role)
	ctx = ctxkeys.WithContextValue(ctx, ctxkeys.PermissionMaskKey, uint64(permMask))
	ctx = ctxkeys.WithContextValue(ctx, CtxKeyTokenJTI, jti)
	
	// Legacy string keys for old handlers/libs
	ctx = ctxkeys.WithContextValue(ctx, CtxKeyUserID, userID)
	ctx = ctxkeys.WithContextValue(ctx, CtxKeyTenantID, tenantID)

	return r.WithContext(ctx)
}
