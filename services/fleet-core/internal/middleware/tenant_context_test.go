package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/pkg/ctxkeys"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTenantContextMiddleware(t *testing.T) {
	logger := zap.NewNop()
	middleware := TenantContextMiddleware(logger)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Case: Valid Org ID
	tenantID := uuid.New()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ctxkeys.TenantIDKey, tenantID.String())
	rr := httptest.NewRecorder()
	middleware(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Missing Org ID
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	middleware(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Case: Super Admin without tenant_id - bypass, should pass
	req = httptest.NewRequest("GET", "/", nil)
	ctx = context.WithValue(req.Context(), ctxkeys.RoleKey, "super_admin")
	rr = httptest.NewRecorder()
	middleware(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Super Admin with tenant_id - should pass and tenant_id remains set
	req = httptest.NewRequest("GET", "/", nil)
	ctx = context.WithValue(req.Context(), ctxkeys.TenantIDKey, tenantID.String())
	ctx = context.WithValue(ctx, ctxkeys.RoleKey, "super_admin")
	rr = httptest.NewRecorder()
	middleware(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)
}
