package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyfleet/fleet-core-service/internal/domains/governance/middleware"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func TestDSARRateLimit(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	logger := zap.NewNop()

	handler := middleware.DSARRateLimit(rdb, nil, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("POST", "/api/v1/governance/requests", nil)
	req.Header.Set("X-User-ID", "user-1")

	// First 5 requests should pass
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	}

	// 6th request should fail
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr.Code)
	}

	// Fast forward time by 1 hour + 1 second
	s.FastForward(time.Hour + time.Second)

	// Should pass again
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status 200 after time advance, got %d", rr2.Code)
	}
}
