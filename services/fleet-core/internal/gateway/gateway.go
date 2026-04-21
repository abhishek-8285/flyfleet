package gateway

import (
	"net/http"
	"time"

	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/platform/observability"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// APIBusGateway implements centralized handling for the platform infrastructure
type APIBusGateway struct {
	logger *zap.Logger
}

// NewAPIBusGateway creates a new APIBusGateway
func NewAPIBusGateway(logger *zap.Logger) *APIBusGateway {
	return &APIBusGateway{
		logger: logger,
	}
}

// Handler returns a middleware that handles centralized management
func (g *APIBusGateway) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Request Logging & Metrics
		defer func() {
			latency := time.Since(start)
			l := observability.WithContext(r.Context(), g.logger)
			l.Info("gateway_intercept",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("latency", latency),
				zap.String("remote_ip", r.RemoteAddr),
			)
		}()

		// 2. Tenant Context Enforcement
		tenantID := middleware.GetTenantID(r.Context())
		if tenantID == uuid.Nil && !isPublicRoute(r.URL.Path) {
			// In a strict SaaS architecture, we would fail here.
			// Currently, we'll log it as a compliance warning.
			l := observability.WithContext(r.Context(), g.logger)
			l.Warn("saas_compliance_warning: missing tenant context",
				zap.String("path", r.URL.Path))
		} else if tenantID != uuid.Nil {
			// Ensure response identifies the tenant context
			w.Header().Set("X-Tenant-ID", tenantID.String())
		}

		// 3. API Versioning Header
		w.Header().Set("X-API-Version", "v1.platform")

		next.ServeHTTP(w, r)
	})
}

func isPublicRoute(path string) bool {
	publicPrefixes := []string{"/health", "/swagger", "/api/v1/auth/login", "/api/v1/auth/register", "/api/v1/auth/verify"}
	for _, p := range publicPrefixes {
		if path == p || (len(path) > len(p) && path[:len(p)] == p) {
			return true
		}
	}
	return false
}
