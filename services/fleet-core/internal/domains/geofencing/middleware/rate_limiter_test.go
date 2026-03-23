package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestGeofenceRateLimiter_ConfiguredRoutes(t *testing.T) {
	rl := NewGeofenceRateLimiter(zap.NewNop())
	if got := rl.getLimiter(http.MethodPost, "/api/v1/geofencing").Limit(); got != 10 {
		t.Fatalf("expected POST /geofencing limit 10, got %v", got)
	}
	if got := rl.getLimiter(http.MethodGet, "/api/v1/geofencing").Limit(); got != 50 {
		t.Fatalf("expected GET /geofencing limit 50, got %v", got)
	}
	if got := rl.getLimiter(http.MethodDelete, "/api/v1/geofencing/123").Limit(); got != 5 {
		t.Fatalf("expected DELETE /geofencing/{id} limit 5, got %v", got)
	}
}

func TestGeofenceRateLimiter_Handler429(t *testing.T) {
	rl := NewGeofenceRateLimiter(zap.NewNop())
	h := rl.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/geofencing", nil)
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if i >= 10 && res.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 after limit, got %d at request %d", res.Code, i)
		}
	}
}
