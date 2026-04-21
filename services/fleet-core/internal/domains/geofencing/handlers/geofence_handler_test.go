package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	sharedmw "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func routerForAuthTests() chi.Router {
	h := NewGeofenceHandler(nil, zap.NewNop())
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func TestCreateGeofence_UnauthorizedWithoutHeaders(t *testing.T) {
	r := routerForAuthTests()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"A","type":"circle"}`))
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestListGeofences_RequiresUserAndOrg(t *testing.T) {
	r := routerForAuthTests()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), sharedmw.TenantIDKey, uuid.New())
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)
	require.Equal(t, http.StatusUnauthorized, res.Code)
}
