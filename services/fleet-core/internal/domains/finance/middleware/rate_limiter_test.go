package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestFinanceRateLimiter_PerEndpointLimits(t *testing.T) {
	rl := NewFinanceRateLimiter(zap.NewNop())

	if got := rl.getLimiter("/api/v1/finance/fuel/ingest").Limit(); got != 50 {
		t.Fatalf("fuel ingest limit = %v, want 50", got)
	}
	if got := rl.getLimiter("/api/v1/finance/expenses/reports").Limit(); got != 20 {
		t.Fatalf("expense submit limit = %v, want 20", got)
	}
	if got := rl.getLimiter("/api/v1/unknown/path").Limit(); got != 100 {
		t.Fatalf("default limit = %v, want 100", got)
	}
}

func TestFinanceRateLimiter_AllowsConfiguredBurst(t *testing.T) {
	rl := NewFinanceRateLimiter(zap.NewNop())
	limiter := rl.getLimiter("/api/v1/finance/fuel/ingest")

	allowed := 0
	for i := 0; i < 50; i++ {
		if limiter.Allow() {
			allowed++
		}
	}
	if allowed != 50 {
		t.Fatalf("allowed = %d, want 50", allowed)
	}

	time.Sleep(25 * time.Millisecond)
	if !limiter.Allow() {
		t.Fatal("expected limiter to refill after wait")
	}
}

func TestFinanceRateLimiter_HandlerReturns429(t *testing.T) {
	rl := NewFinanceRateLimiter(zap.NewNop())
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := rl.Handler(next)

	// Consume burst capacity.
	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/finance/fuel/ingest", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/finance/fuel/ingest", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}
