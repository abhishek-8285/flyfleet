package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/domains/auth/handlers"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/models"
	"github.com/flyfleet/fleet-core-service/internal/domains/auth/service"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func TestInviteValidationRateLimiter(t *testing.T) {
	// Setup Redis (use in-memory miniredis or standard local redis)
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis is not available, skipping rate limiter test")
	}

	// Flush DB to ensure clean state
	rdb.FlushDB(ctx)

	logger := zap.NewNop()

	// Create a dummy service to prevent panics when rate limits allow the request through
	mockRepo := &mockAuthRepo{}
	svc := service.NewAuthService(nil, mockRepo, nil, "secret", rdb, nil)
	h := handlers.NewHandlers(svc, logger, rdb, "")

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		h.RegisterRoutes(r)
	})

	// Simulate 6 attempts from the same IP
	testIP := "192.168.1.55"

	for i := 1; i <= 6; i++ {
		reqBody := `{"code":"test-code-123"}`
		req, _ := http.NewRequest("POST", "/api/v1/auth/invites/validate", bytes.NewBufferString(reqBody))
		req.Header.Set("X-Forwarded-For", testIP)
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if i <= 5 {
			// Actually the handler returns 500/200 when svc is nil and rate limit allows, but we only care it's NOT 429
			if rr.Code == http.StatusTooManyRequests {
				t.Fatalf("Attempt %d was unexpectedly rate limited", i)
			}
		} else {
			if rr.Code != http.StatusTooManyRequests {
				t.Fatalf("Attempt %d should have been rate limited with 429, got %d", i, rr.Code)
			}
		}
	}
}

type mockAuthRepo struct {
	service.AuthRepositoryInterface
}

func (m *mockAuthRepo) GetInviteByCode(ctx context.Context, code string) (*models.Invite, error) {
	return &models.Invite{}, nil // Return empty to bypass validation logic safely
}
