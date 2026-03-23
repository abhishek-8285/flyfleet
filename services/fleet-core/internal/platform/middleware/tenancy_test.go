package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core-service/internal/platform/iam/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTenantContextMiddleware(t *testing.T) {
	handler := TenantContextMiddleware()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Case: Valid Org ID
	orgID := uuid.New()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middleware.OrgIDKey, orgID)
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req.WithContext(ctx))
	assert.Equal(t, http.StatusOK, rr.Code)

	// Case: Missing Org ID
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestGetOrganizationID(t *testing.T) {
	orgID := uuid.New()
	ctx := context.WithValue(context.Background(), middleware.OrgIDKey, orgID)

	assert.Equal(t, orgID, GetOrganizationID(ctx))
	assert.Equal(t, uuid.Nil, GetOrganizationID(context.Background()))
}
