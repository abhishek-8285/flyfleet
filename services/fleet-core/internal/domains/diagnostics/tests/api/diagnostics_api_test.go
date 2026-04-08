package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flyfleet/fleet-core/internal/domains/diagnostics/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestDiagnosticsAPI_Endpoints(t *testing.T) {
	// Mock handler dependencies or use real ones if shared pool is available
	// For API tests, we'll use a new router
	r := chi.NewRouter()

	// We need a way to mock the services or use a test setup.
	// Since this is in diagnostics_v2/tests/api, we can't easily share the integration pool
	// unless we move it to a shared place or use a different strategy.

	// For now, let's just verify the routes are wired correctly with a focus on structure.
	h := handlers.NewDiagnosticsHandler(nil, nil, nil)
	h.RegisterRoutes(r)

	t.Run("GET /diagnostics/devices returns 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/diagnostics/devices", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST /diagnostics/devices returns 201", func(t *testing.T) {
		body := `{"device_imei": "12345"}`
		req := httptest.NewRequest("POST", "/diagnostics/devices", strings.NewReader(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
