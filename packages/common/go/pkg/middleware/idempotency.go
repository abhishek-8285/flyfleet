package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// IdempotencyMiddleware ensures that mutating operations with the same Idempotency-Key are processed exactly once.
func IdempotencyMiddleware(client *redis.Client, logger *zap.Logger, expiration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to mutating requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodDelete {
				next.ServeHTTP(w, r)
				return
			}

			idemKey := r.Header.Get("Idempotency-Key")
			if idemKey == "" {
				// No key provided, bypass idempotency check
				next.ServeHTTP(w, r)
				return
			}

			// Generate a unique cache key based on the Idempotency-Key and User/Org to prevent cross-tenant collisions
			userID := ""
			if uid, ok := r.Context().Value("user_id").(string); ok {
				userID = uid
			}
			if userID == "" { // fallback if auth unparsed
				userID = "anonymous:" + r.RemoteAddr
			}
			
			// Include request path in hash to prevent reuse of key across different endpoints
			hashInput := idemKey + ":" + userID + ":" + r.URL.Path
			hash := sha256.Sum256([]byte(hashInput))
			cacheKey := "idem:" + hex.EncodeToString(hash[:])

			// Check if the key exists
			ctx := r.Context()
			cachedResp, err := client.Get(ctx, cacheKey).Result()
			if err == redis.Nil {
				// Key does not exist. Try to lock it to indicate we are processing.
				// SetNX returns true if key didn't exist and was set.
				ok, err := client.SetNX(ctx, cacheKey, "PROCESSING", expiration).Result()
				if err != nil {
					logger.Error("Failed to set idempotency lock", zap.Error(err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				if !ok {
					// Someone else set it right before us (race condition), conflict.
					http.Error(w, "Conflict: Request already processing", http.StatusConflict)
					return
				}

				// We hold the lock. Process the request.
				// Capture the response using a ResponseRecorder.
				rec := httptest.NewRecorder()
				
				// Read body, then reset so next handler can read it
				var bodyBytes []byte
				if r.Body != nil {
					bodyBytes, _ = io.ReadAll(r.Body)
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}

				next.ServeHTTP(rec, r)

				// After processing, store the result if successful or leave it for retry if 5xx?
				// Usually we cache everything, including 4xx, but maybe not 5xx if it's considered retryable.
				// For simplicity, we cache the response text.
				respBody := rec.Body.Bytes()
				err = client.Set(ctx, cacheKey, respBody, expiration).Err()
				if err != nil {
					logger.Error("Failed to store idempotency response", zap.Error(err))
					// We still return we processed it
				}

				// Copy headers and status
				for k, v := range rec.Header() {
					w.Header()[k] = v
				}
				w.WriteHeader(rec.Code)
				w.Write(respBody)

			} else if err != nil {
				// Redis error
				logger.Error("Redis error during idempotency check", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			} else {
				// Key exists
				if cachedResp == "PROCESSING" {
					// Concurrent request
					w.Header().Set("Retry-After", "2")
					http.Error(w, "Conflict: Request is currently processing", http.StatusConflict)
					return
				}

				// Return cached response
				w.Header().Set("X-Idempotency-Restored", "true")
				w.Header().Set("Content-Type", "application/json") // Assuming JSON
				w.WriteHeader(http.StatusOK) // In a fully robust system, you'd cache the status code too.
				w.Write([]byte(cachedResp))
			}
		})
	}
}
