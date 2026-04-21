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
	handler := TenantContextMiddleware(logger)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Case: Valid Org ID
	tenantID := uuid.New()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), ctxkeys.TenantIDKey, tenantID.String())
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Missing Org ID
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Case: Super Admin Bypass - should pass without tenant_id
	req = httptest.NewRequest("GET", "/", nil)
	ctx = context.WithValue(req.Context(), ctxkeys.RoleKey, "super_admin")
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGetOrganizationID(t *testing.T) {
	tenantID := uuid.New()
	ctx := context.WithValue(context.Background(), ctxkeys.TenantIDKey, tenantID.String())

	assert.Equal(t, tenantID, GetOrganizationID(ctx))
	assert.Equal(t, uuid.Nil, GetOrganizationID(context.Background()))
}
