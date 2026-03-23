package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Gateway struct {
	logger *zap.Logger
	redis  *redis.Client
}

func NewGateway(logger *zap.Logger, redisClient *redis.Client) *Gateway {
	return &Gateway{
		logger: logger,
		redis:  redisClient,
	}
}

func (g *Gateway) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *Gateway) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		
		g.logger.Info("request processed",
			zap.String("request_id", r.Header.Get("X-Request-ID")),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.status),
			zap.Duration("latency", time.Since(start)),
		)
	})
}

func (g *Gateway) RateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if g.redis == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("ratelimit:%s", r.RemoteAddr)
			count, err := g.redis.Incr(r.Context(), key).Result()
			if err != nil {
				g.logger.Error("rate limit error", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				g.redis.Expire(r.Context(), key, window)
			}

			if count > int64(limit) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (g *Gateway) RegisterRoutes(r chi.Router, v1Handler, v2Handler http.Handler) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/", v1Handler)
	})
	r.Route("/api/v2", func(r chi.Router) {
		r.Mount("/", v2Handler)
	})
}
