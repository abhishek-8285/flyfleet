package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	iamCtx "github.com/flyfleet/fleet-core/internal/platform/iam/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTenantContextMiddleware(t *testing.T) {
	logger := zap.NewNop()
	handler := TenantContextMiddleware(logger)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Case: Valid Org ID
	orgID := uuid.New()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), iamCtx.OrgIDKey, orgID)
	rr := httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Missing Org ID
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Case: Super Admin without org_id - bypass, should pass
	req = httptest.NewRequest("GET", "/", nil)
	ctx = context.WithValue(req.Context(), ctxkeys.RoleKey, "super_admin")
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Super Admin with org_id - should pass and org_id remains set
	req = httptest.NewRequest("GET", "/", nil)
	ctx = context.WithValue(req.Context(), iamCtx.OrgIDKey, orgID)
	ctx = context.WithValue(ctx, ctxkeys.RoleKey, "super_admin")
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)
}
