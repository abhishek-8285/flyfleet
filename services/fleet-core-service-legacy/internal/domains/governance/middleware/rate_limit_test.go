package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestDSARRateLimit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	logger := zap.NewNop()
	middleware := DSARRateLimit(rdb, nil, logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	userID := "user-123"

	// Perform 5 requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/governance/requests", nil)
		req.Header.Set("X-User-ID", userID)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// 6th request should fail
	req := httptest.NewRequest("POST", "/api/v1/governance/requests", nil)
	req.Header.Set("X-User-ID", userID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)

	// Another user should still be able to request
	req2 := httptest.NewRequest("POST", "/api/v1/governance/requests", nil)
	req2.Header.Set("X-User-ID", "user-456")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}
