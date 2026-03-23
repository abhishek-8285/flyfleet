package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGateway_RequestIDMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	g := NewGateway(logger, nil)
	handler := g.RequestIDMiddleware

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}

func TestGateway_LoggingMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	g := NewGateway(logger, nil)
	handler := g.LoggingMiddleware

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "test-id")
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGateway_RateLimitMiddleware_NoRedis(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	g := NewGateway(logger, nil)
	handler := g.RateLimitMiddleware(100, time.Minute)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler(next).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
