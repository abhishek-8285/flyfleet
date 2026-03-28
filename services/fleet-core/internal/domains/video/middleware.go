package video

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ─── Device Authentication Middleware ────────────────────────────────────────

// DeviceAuthMiddleware authenticates dashcam edge device requests.
// Devices must include an HMAC-SHA256 signature in the X-Device-Token header.
//
// Header format:
//
//	X-Device-Token: {device_id}:{timestamp}:{hmac_signature}
//
// Where signature = HMAC-SHA256("{device_id}:{timestamp}", deviceMasterSecret).
// Timestamp must be within ±5 minutes to prevent replay attacks.
func DeviceAuthMiddleware(secret string, logger *zap.Logger) func(http.Handler) http.Handler {
	secretBytes := []byte(secret)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Device-Token")
			if token == "" {
				http.Error(w, "Device authentication required", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(token, ":", 3)
			if len(parts) != 3 {
				http.Error(w, "Invalid device token format", http.StatusUnauthorized)
				return
			}

			deviceID, tsStr, sig := parts[0], parts[1], parts[2]

			// Replay protection: verify timestamp is recent
			var ts int64
			if _, err := fmt.Sscanf(tsStr, "%d", &ts); err != nil {
				http.Error(w, "Invalid token timestamp", http.StatusUnauthorized)
				return
			}
			tokenTime := time.Unix(ts, 0)
			if diff := time.Since(tokenTime); diff > 5*time.Minute || diff < -5*time.Minute {
				http.Error(w, "Token expired or future-dated", http.StatusUnauthorized)
				return
			}

			// HMAC verification
			payload := deviceID + ":" + tsStr
			mac := hmac.New(sha256.New, secretBytes)
			mac.Write([]byte(payload))
			expected := hex.EncodeToString(mac.Sum(nil))
			if !hmac.Equal([]byte(expected), []byte(sig)) {
				logger.Warn("Device authentication failed",
					zap.String("device_id", deviceID),
					zap.String("remote_addr", r.RemoteAddr))
				http.Error(w, "Invalid device signature", http.StatusUnauthorized)
				return
			}

			logger.Debug("Device authenticated", zap.String("device_id", deviceID))
			next.ServeHTTP(w, r)
		})
	}
}

// ─── Streaming Rate Limiter ───────────────────────────────────────────────────

// StreamRateLimiter limits concurrent video streaming sessions per tenant.
// Prevents a single tenant from monopolizing bandwidth/CDN capacity.
type StreamRateLimiter struct {
	mu           sync.Mutex
	counts       map[string]int // tenantID → active stream count
	maxPerTenant int
	logger       *zap.Logger
}

func NewStreamRateLimiter(maxPerTenant int, logger *zap.Logger) *StreamRateLimiter {
	return &StreamRateLimiter{
		counts:       make(map[string]int),
		maxPerTenant: maxPerTenant,
		logger:       logger,
	}
}

// Acquire attempts to claim a streaming slot for the tenant.
// Returns true if allowed, false if rate limit exceeded.
func (l *StreamRateLimiter) Acquire(tenantID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.counts[tenantID] >= l.maxPerTenant {
		l.logger.Warn("Stream rate limit exceeded",
			zap.String("tenant_id", tenantID),
			zap.Int("active_streams", l.counts[tenantID]),
			zap.Int("limit", l.maxPerTenant))
		return false
	}
	l.counts[tenantID]++
	return true
}

// Release frees a streaming slot for the tenant.
func (l *StreamRateLimiter) Release(tenantID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.counts[tenantID] > 0 {
		l.counts[tenantID]--
	}
}

// Middleware wraps the stream endpoint with the rate limiter.
func (l *StreamRateLimiter) Middleware(tenantIDFn func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := tenantIDFn(r)
			if !l.Acquire(tenantID) {
				http.Error(w, "Too many concurrent streams for this tenant", http.StatusTooManyRequests)
				return
			}
			defer l.Release(tenantID)
			next.ServeHTTP(w, r)
		})
	}
}
