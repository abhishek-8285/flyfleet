package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	CtxKeyUserID   = "user_id"
	CtxKeyOrgID    = "org_id"
	CtxKeyRole     = "role"
	CtxKeyPermMask = "permissions_mask"
	CtxKeyTokenJTI = "token_jti"
)

func GetUserID(r *http.Request) string {
	v, _ := r.Context().Value(CtxKeyUserID).(string)
	return v
}

func GetOrgID(r *http.Request) string {
	v, _ := r.Context().Value(CtxKeyOrgID).(string)
	return v
}

func GetOrgUUID(r *http.Request) uuid.UUID {
	v, _ := r.Context().Value(CtxKeyOrgID).(string)
	id, _ := uuid.Parse(v)
	return id
}

func GetUserIDUUID(r *http.Request) uuid.UUID {
	v, _ := r.Context().Value(CtxKeyUserID).(string)
	id, _ := uuid.Parse(v)
	return id
}

func GetRole(r *http.Request) string {
	v, _ := r.Context().Value(CtxKeyRole).(string)
	return v
}

func GetPermissionMask(r *http.Request) int64 {
	v, _ := r.Context().Value(CtxKeyPermMask).(int64)
	return v
}

func GetTokenJTI(r *http.Request) string {
	v, _ := r.Context().Value(CtxKeyTokenJTI).(string)
	return v
}

func SetAuthClaims(r *http.Request, userID, orgID, role string, permMask int64, jti string) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, CtxKeyUserID, userID)
	ctx = context.WithValue(ctx, CtxKeyOrgID, orgID)
	ctx = context.WithValue(ctx, CtxKeyRole, role)
	ctx = context.WithValue(ctx, CtxKeyPermMask, permMask)
	ctx = context.WithValue(ctx, CtxKeyTokenJTI, jti)
	return r.WithContext(ctx)
}
