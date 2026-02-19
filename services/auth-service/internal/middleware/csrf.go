package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CSRFTokenHeader is the header name for CSRF tokens
const CSRFTokenHeader = "X-CSRF-Token"

// CSRFStore holds the CSRF tokens with expiration
type CSRFStore struct {
	tokens map[string]time.Time
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewCSRFStore creates a new CSRF store
func NewCSRFStore(logger *zap.Logger) *CSRFStore {
	store := &CSRFStore{
		tokens: make(map[string]time.Time),
		logger: logger,
	}
	
	// Start cleanup goroutine to remove expired tokens
	go store.cleanupExpiredTokens()
	
	return store
}

// cleanupExpiredTokens removes expired tokens periodically
func (s *CSRFStore) cleanupExpiredTokens() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for token, expiry := range s.tokens {
			if expiry.Before(now) {
				delete(s.tokens, token)
			}
		}
		s.mutex.Unlock()
	}
}

// CSRFMiddleware provides CSRF protection
func CSRFMiddleware(store *CSRFStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF for safe methods
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE" {
				// Generate a new CSRF token for the client
				token := generateCSRFToken()
				expiry := time.Now().Add(24 * time.Hour) // Token valid for 24 hours
				
				store.mutex.Lock()
				store.tokens[token] = expiry
				store.mutex.Unlock()
				
				// Add the token to the response headers
				w.Header().Set(CSRFTokenHeader, token)
				
				next.ServeHTTP(w, r)
				return
			}

			// For unsafe methods, validate the CSRF token
			clientToken := r.Header.Get(CSRFTokenHeader)
			if clientToken == "" {
				store.logger.Warn("Missing CSRF token", zap.String("method", r.Method), zap.String("path", r.URL.Path))
				http.Error(w, "Missing CSRF token", http.StatusForbidden)
				return
			}

			// Check if the token exists in our store
			store.mutex.RLock()
			expiry, exists := store.tokens[clientToken]
			store.mutex.RUnlock()
			
			if !exists || expiry.Before(time.Now()) {
				store.logger.Warn("Invalid or expired CSRF token", zap.String("token", clientToken), zap.String("method", r.Method), zap.String("path", r.URL.Path))
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}

			// Consume the token to prevent replay attacks
			store.mutex.Lock()
			delete(store.tokens, clientToken)
			store.mutex.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// generateCSRFToken generates a random CSRF token
func generateCSRFToken() string {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		// This should not happen in practice, but if it does, we'll panic
		panic(fmt.Sprintf("Failed to generate CSRF token: %v", err))
	}
	return base64.URLEncoding.EncodeToString(bytes)
}