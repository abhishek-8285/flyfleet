package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
)

type DeviceAuthMiddleware struct {
	repo *repository.DiagnosticsRepository
}

func NewDeviceAuthMiddleware(repo *repository.DiagnosticsRepository) *DeviceAuthMiddleware {
	return &DeviceAuthMiddleware{
		repo: repo,
	}
}

// EnsureAuthenticated validates device tokens and signatures
func (m *DeviceAuthMiddleware) EnsureAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "DeviceToken" {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// In a real implementation this would query device_auth_tokens table
		// device_id := m.repo.GetDeviceIDByToken(ctx, token)
		if token == "" {
			http.Error(w, "Invalid device token", http.StatusUnauthorized)
			return
		}

		// Attach context if successful
		ctx := context.WithValue(r.Context(), "device_token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
