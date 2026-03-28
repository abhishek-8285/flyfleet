package middleware

import (
	"net/http"
	"sync"
	"time"
)

// idempotencyEntry caches the response for idempotent requests.
type idempotencyEntry struct {
	statusCode int
	body       []byte
	expiresAt  time.Time
}

// idempotencyStore is an in-memory store for idempotency keys.
// In production, replace with Redis:
//
//	GET idempotency:<key>  → return cached response
//	SET idempotency:<key> EX 86400 → store for 24h
type idempotencyStore struct {
	mu      sync.RWMutex
	entries map[string]*idempotencyEntry
}

var store = &idempotencyStore{
	entries: make(map[string]*idempotencyEntry),
}

// captureResponseWriter intercepts response writes so we can cache them.
type captureResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (c *captureResponseWriter) WriteHeader(code int) {
	c.statusCode = code
	c.ResponseWriter.WriteHeader(code)
}

func (c *captureResponseWriter) Write(b []byte) (int, error) {
	c.body = append(c.body, b...)
	return c.ResponseWriter.Write(b)
}

// IdempotencyMiddleware enforces the Idempotency-Key header contract:
//   - If the key has been seen within 24h, return the cached response directly.
//   - Otherwise execute the handler and cache the result.
//
// Spec: All mutating trip lifecycle APIs require Idempotency-Key header.
func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to mutating methods
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, `{"error":"Idempotency-Key header is required for mutating requests"}`, http.StatusUnprocessableEntity)
			return
		}

		// Check for cached response
		store.mu.RLock()
		entry, found := store.entries[key]
		store.mu.RUnlock()

		if found && time.Now().Before(entry.expiresAt) {
			// Replay cached response
			w.Header().Set("X-Idempotency-Replayed", "true")
			w.WriteHeader(entry.statusCode)
			w.Write(entry.body)
			return
		}

		// Execute and capture response
		cw := &captureResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(cw, r)

		// Only cache successful/client-error responses (not 5xx)
		if cw.statusCode < 500 {
			store.mu.Lock()
			store.entries[key] = &idempotencyEntry{
				statusCode: cw.statusCode,
				body:       cw.body,
				expiresAt:  time.Now().Add(24 * time.Hour),
			}
			store.mu.Unlock()
		}
	})
}
