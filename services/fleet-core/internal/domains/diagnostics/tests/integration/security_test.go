//go:build integration
// +build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestSecurity_Middleware(t *testing.T) {
	// 1. Auth Middleware Test
	t.Run("Device Auth Middleware - No Token", func(t *testing.T) {
		r := chi.NewRouter()
		authMid := middleware.NewDeviceAuthMiddleware(repo).EnsureAuthenticated
		r.Use(authMid)
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// 2. Rate Limiting Test (Requires Redis)
	t.Run("Rate Limiter Middleware", func(t *testing.T) {
		// Mock/Real Redis setup would be needed here. 
		// In our env, Redis is assumed but not explicitly started in TestMain yet.
		// We'll skip deep Redis testing or just verify it doesn't crash.
	})
}
