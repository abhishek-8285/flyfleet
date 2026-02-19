package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	sharedMiddleware "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/web-bff/internal/client"
	"github.com/flyfleet/web-bff/internal/config"
	"github.com/flyfleet/web-bff/internal/handlers"
	bffMiddleware "github.com/flyfleet/web-bff/internal/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	// Load configuration
	cfg := config.Load()

	// Initialize Redis Cache
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, "", 0, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redisCache.Close()

	// Initialize Redis Rate Limiter
	rateLimiter := sharedMiddleware.NewRedisRateLimiter(redisCache.GetClient(), logger)

	// Initialize NATS
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Warn("Failed to connect to NATS", zap.Error(err))
	} else {
		defer nc.Close()
	}

	// Initialize clients
	authClient := client.NewAuthClient(cfg.AuthServiceURL)
	vehicleClient := client.NewVehicleClient(cfg.FleetCoreServiceURL)
	tripClient := client.NewTripClient(cfg.FleetCoreServiceURL)
	alertClient := client.NewAlertClient(cfg.AlertServiceURL)
	billingClient := client.NewBillingClient(cfg.BillingServiceURL)
	notificationClient := client.NewNotificationClient(cfg.AlertServiceURL)
	supportClient := client.NewSupportClient(cfg.AdminSupportServiceURL)

	// Initialize handlers
	dashboardHandlers := handlers.NewDashboardHandlers(cfg, logger, redisCache, authClient, vehicleClient, tripClient, alertClient, billingClient, supportClient)
	monitoringHandlers := handlers.NewMonitoringHandlers(cfg, logger)
	authHandlers := handlers.NewAuthHandlers(cfg, logger)
	searchHandlers := handlers.NewSearchHandlers(cfg, logger)
	timelineHandlers := handlers.NewTimelineHandlers(cfg, logger, notificationClient)
	wsHandlers := handlers.NewWSHandler(nc, logger)

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.RequestID)
	r.Use(bffMiddleware.TraceIDMiddleware)

	// Global Rate Limiting (Pillar 2)
	// 500 requests per minute per Org/IP
	r.Use(sharedMiddleware.OrgRateLimit(rateLimiter, 500, time.Minute))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Get("/api/v1/ws", wsHandlers.HandleWS)

	// Public auth routes
	r.Post("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Strict rate limit for login
		loginHandler := sharedMiddleware.FixedRateLimit(rateLimiter, 5, time.Minute, func(r *http.Request) string {
			return "login:" + r.RemoteAddr
		})
		loginHandler(http.HandlerFunc(authHandlers.ProxyLogin)).ServeHTTP(w, r)
	})
	registerPrefixProxy(r, "/api/v1/auth", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/v1/auth", "/api/v1/auth", nil))

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.AuthMiddleware(cfg))

		r.Get("/api/v1/dashboard/stats", dashboardHandlers.GetGlobalStats)
		r.Get("/api/v1/admin/dashboard/data", dashboardHandlers.GetAdminDashboardData)

		// Monitoring & SLA (Pillar 7)
		r.Get("/api/v1/monitoring/sla", monitoringHandlers.GetSLAMetrics)
		r.Get("/api/v1/monitoring/health", monitoringHandlers.GetSystemHealth)
		r.Get("/api/v1/admin/monitoring/sla", monitoringHandlers.GetSLAMetrics)
		r.Get("/api/v1/admin/monitoring/health", monitoringHandlers.GetSystemHealth)
		r.Get("/api/v1/admin/monitoring/metrics", monitoringHandlers.GetSystemHealth)
		r.Get("/api/v1/admin/monitoring/queues", monitoringHandlers.GetSystemHealth)
		r.Get("/api/v1/admin/monitoring/logs", monitoringHandlers.GetSystemHealth)

		// Search and timeline APIs used by dashboard.
		r.Get("/api/search", searchHandlers.GlobalSearch)
		r.Get("/api/search/facets", searchHandlers.GetSearchFacets)
		r.Get("/api/search/suggestions", searchHandlers.GetSearchSuggestions)
		r.Get("/api/search/recent", searchHandlers.GetRecentSearches)
		r.Get("/api/search/saved", searchHandlers.GetSavedSearches)
		r.Get("/api/v1/search", searchHandlers.GlobalSearch)
		r.Get("/api/v1/search/facets", searchHandlers.GetSearchFacets)
		r.Get("/api/v1/search/suggestions", searchHandlers.GetSearchSuggestions)
		r.Get("/api/v1/search/recent", searchHandlers.GetRecentSearches)
		r.Get("/api/v1/search/saved", searchHandlers.GetSavedSearches)
		r.Get("/api/timeline", timelineHandlers.GetTimeline)
		r.Get("/api/timeline/stats", timelineHandlers.GetTimelineStats)
		r.Get("/api/timeline/filters", timelineHandlers.GetTimelineFilters)
		r.Get("/api/v1/timeline", timelineHandlers.GetTimeline)
		r.Get("/api/v1/timeline/stats", timelineHandlers.GetTimelineStats)
		r.Get("/api/v1/timeline/filters", timelineHandlers.GetTimelineFilters)

		// High-priority admin/service proxies (specific prefixes first).
		registerPrefixProxy(r, "/api/v1/admin/governance", makePrefixProxy(logger, cfg.AdminSupportServiceURL, "/api/v1/admin/governance", "/api/v1/governance", nil))
		registerPrefixProxy(r, "/api/v1/admin/dr", makePrefixProxy(logger, cfg.AdminSupportServiceURL, "/api/v1/admin/dr", "/api/v1/dr", rewriteDRPaths))
		registerPrefixProxy(r, "/api/v1/admin/reports", makePrefixProxy(logger, cfg.ReportsServiceURL, "/api/v1/admin/reports", "/api/v1/reports", nil))
		registerPrefixProxy(r, "/api/v1/admin/notifications", makePrefixProxy(logger, cfg.AlertServiceURL, "/api/v1/admin/notifications", "/api/v1/notifications", nil))
		registerPrefixProxy(r, "/api/v1/admin/preferences", makePrefixProxy(logger, cfg.AlertServiceURL, "/api/v1/admin/preferences", "/api/v1/notifications/preferences", nil))
		registerPrefixProxy(r, "/api/v1/admin/billing", makePrefixProxy(logger, cfg.BillingServiceURL, "/api/v1/admin/billing", "/api/v1/billing", nil))
		registerPrefixProxy(r, "/api/v1/admin/billing-settings", makePrefixProxy(logger, cfg.BillingServiceURL, "/api/v1/admin/billing-settings", "/api/v1/billing/settings", nil))
		registerPrefixProxy(r, "/api/v1/admin/billing-metrics", makePrefixProxy(logger, cfg.BillingServiceURL, "/api/v1/admin/billing-metrics", "/api/v1/billing/metrics", nil))
		registerPrefixProxy(r, "/api/v1/admin/settings", makePrefixProxy(logger, cfg.SettingsServiceURL, "/api/v1/admin/settings", "/settings", nil))
		registerPrefixProxy(r, "/api/v1/settings", makePrefixProxy(logger, cfg.SettingsServiceURL, "/api/v1/settings", "/settings", nil))

		// Fleet core routed admin modules.
		registerPrefixProxy(r, "/api/v1/admin/work-orders", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/work-orders", "/api/v1/work-orders", nil))
		registerPrefixProxy(r, "/api/v1/admin/devices", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/devices", "/api/v1/diagnostics/devices", nil))
		registerPrefixProxy(r, "/api/v1/admin/inventory", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/inventory", "/api/v1/inventory", nil))
		registerPrefixProxy(r, "/api/v1/admin/vendors", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/vendors", "/api/v1/vendors", nil))
		registerPrefixProxy(r, "/api/v1/admin/inspections", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/inspections", "/api/v1/inspections", nil))
		registerPrefixProxy(r, "/api/v1/admin/fuel", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/fuel", "/api/v1/fuel", nil))
		registerPrefixProxy(r, "/api/v1/admin/finance", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/finance", "/api/v1/finance", nil))
		registerPrefixProxy(r, "/api/v1/admin/revenue", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/revenue", "/api/v1/finance/revenue", nil))
		registerPrefixProxy(r, "/api/v1/admin/fuel-logs", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/fuel-logs", "/api/v1/fuel", nil))
		registerPrefixProxy(r, "/api/v1/admin/video", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/video", "/api/v1/video", nil))
		registerPrefixProxy(r, "/api/v1/admin/emergency", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/admin/emergency", "/api/v1/emergency", nil))
		registerPrefixProxy(r, "/api/v1/routes", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/routes", "/api/v1/routes", nil))
		registerPrefixProxy(r, "/api/v1/customer", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/customer", "/api/v1/customer", nil))
		registerPrefixProxy(r, "/api/v1/emergency", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/emergency", "/api/v1/emergency", nil))

		// Device diagnostics legacy paths.
		registerPrefixProxy(r, "/api/devices/summary", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/devices/summary", "/api/v1/diagnostics/summary", nil))
		registerPrefixProxy(r, "/api/devices/status", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/devices/status", "/api/v1/diagnostics/devices", nil))
		registerPrefixProxy(r, "/api/devices/twins", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/devices/twins", "/api/v1/diagnostics/twins", nil))
		registerPrefixProxy(r, "/api/devices", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/devices", "/api/v1/diagnostics/devices", nil))
		registerPrefixProxy(r, "/api/firmware/rollouts", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/firmware/rollouts", "/api/v1/diagnostics/rollouts", nil))
		registerPrefixProxy(r, "/api/firmware", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/firmware", "/api/v1/diagnostics/firmware", nil))

		// Security path harmonization.
		registerPrefixProxy(r, "/api/mfa", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/mfa", "/api/v1/security/mfa", nil))
		registerPrefixProxy(r, "/api/audit", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/audit", "/api/v1/admin/audit", nil))
		registerPrefixProxy(r, "/api/security", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/security", "/api/v1/security", nil))
		registerPrefixProxy(r, "/api/v1/security", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/v1/security", "/api/v1/security", nil))
		registerPrefixProxy(r, "/api/v1/rbac", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/v1/rbac", "/api/v1/admin/rbac", nil))

		// Integrations legacy prefixes.
		registerPrefixProxy(r, "/api/integrations", makePrefixProxy(logger, cfg.IntegrationsServiceURL, "/api/integrations", "/api/v1/integrations", nil))
		registerPrefixProxy(r, "/api/api-clients", makePrefixProxy(logger, cfg.IntegrationsServiceURL, "/api/api-clients", "/api/v1/api-clients", nil))
		registerPrefixProxy(r, "/api/webhooks", makePrefixProxy(logger, cfg.IntegrationsServiceURL, "/api/webhooks", "/api/v1/webhooks", nil))

		// Trip legacy and direct routes used by older pages.
		registerPrefixProxy(r, "/api/v1/trips", makePrefixProxy(logger, cfg.FleetCoreServiceURL, "/api/v1/trips", "/api/v1/trips", nil))

		// Support APIs currently consumed by dashboard.
		registerPrefixProxy(r, "/api/support", makePrefixProxy(logger, cfg.AdminSupportServiceURL, "/api/support", "/api/v1/support", rewriteSupportPaths))
		registerPrefixProxy(r, "/bff/support", makePrefixProxy(logger, cfg.AdminSupportServiceURL, "/bff/support", "/api/v1/support", rewriteSupportPaths))

		// Catch-all admin proxy goes last.
		registerPrefixProxy(r, "/api/v1/admin", makePrefixProxy(logger, cfg.AuthServiceURL, "/api/v1/admin", "/api/v1/admin", nil))
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		logger.Info("Starting Web BFF service with Redis Scaling", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down Web BFF service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Web BFF service exited")
}

func registerPrefixProxy(r chi.Router, prefix string, handler http.HandlerFunc) {
	r.Handle(prefix, handler)
	if strings.HasSuffix(prefix, "/") {
		r.Handle(prefix+"*", handler)
		return
	}
	r.Handle(prefix+"/*", handler)
}

func makePrefixProxy(logger *zap.Logger, serviceBaseURL, sourcePrefix, targetPrefix string, rewrite func(string) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, sourcePrefix) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		targetPath := targetPrefix + strings.TrimPrefix(r.URL.Path, sourcePrefix)
		if rewrite != nil {
			targetPath = rewrite(targetPath)
		}

		baseURL, err := url.Parse(serviceBaseURL)
		if err != nil {
			logger.Error("Failed to parse service URL", zap.String("service_url", serviceBaseURL), zap.Error(err))
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
		baseURL.Path = targetPath
		baseURL.RawQuery = r.URL.RawQuery

		req, err := http.NewRequestWithContext(r.Context(), r.Method, baseURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for name, values := range r.Header {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}
		propagateIdentityHeaders(req, r)
		if traceID := bffMiddleware.GetTraceID(r.Context()); traceID != "" {
			req.Header.Set(bffMiddleware.HeaderTraceID, traceID)
			req.Header.Set(bffMiddleware.HeaderRequestID, traceID)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error("Proxy upstream request failed", zap.String("target", baseURL.String()), zap.Error(err))
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		if traceID := bffMiddleware.GetTraceID(r.Context()); traceID != "" {
			w.Header().Set(bffMiddleware.HeaderTraceID, traceID)
			w.Header().Set(bffMiddleware.HeaderRequestID, traceID)
		}

		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func propagateIdentityHeaders(outgoing *http.Request, incoming *http.Request) {
	if outgoing.Header.Get("X-Organization-ID") != "" && outgoing.Header.Get("X-User-ID") != "" {
		return
	}

	authHeader := strings.TrimSpace(incoming.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenString == "" {
		return
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}

	if outgoing.Header.Get("X-Organization-ID") == "" {
		if orgID, ok := claims["org_id"].(string); ok && strings.TrimSpace(orgID) != "" {
			outgoing.Header.Set("X-Organization-ID", orgID)
		}
	}

	if outgoing.Header.Get("X-User-ID") == "" {
		if userID, ok := claims["user_id"].(string); ok && strings.TrimSpace(userID) != "" {
			outgoing.Header.Set("X-User-ID", userID)
			return
		}
		if subject, ok := claims["sub"].(string); ok && strings.TrimSpace(subject) != "" {
			outgoing.Header.Set("X-User-ID", subject)
		}
	}
}

func rewriteDRPaths(path string) string {
	path = strings.Replace(path, "/restores/", "/restore-jobs/", 1)
	path = strings.Replace(path, "/restores", "/restore-jobs", 1)
	return path
}

func rewriteSupportPaths(path string) string {
	if strings.HasSuffix(path, "/dashboard/stats") {
		return strings.TrimSuffix(path, "/dashboard/stats") + "/dashboard/summary"
	}
	if strings.HasSuffix(path, "/incidents/feed") {
		return strings.TrimSuffix(path, "/incidents/feed") + "/incidents"
	}
	if strings.HasSuffix(path, "/incidents/metrics") {
		return strings.TrimSuffix(path, "/incidents/metrics") + "/dashboard/summary"
	}
	if strings.HasSuffix(path, "/incidents/trends") {
		return strings.TrimSuffix(path, "/incidents/trends") + "/dashboard/recent-activity"
	}
	return path
}
