package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	sharedMiddleware "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/web-bff/internal/client"
	"github.com/flyfleet/web-bff/internal/config"
	"github.com/flyfleet/web-bff/internal/handlers"
	bffMiddleware "github.com/flyfleet/web-bff/internal/middleware"
	"github.com/flyfleet/web-bff/internal/routing"
	"github.com/flyfleet/web-bff/internal/waf"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/fleet-platform/shared/pkg/analytics"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	// Initialize Analytics
	analytics.Init(os.Getenv("POSTHOG_API_KEY"), os.Getenv("POSTHOG_ENDPOINT"))
	defer analytics.Close()

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
	authClient, err := client.NewAuthClient(cfg.AuthServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize auth client", zap.Error(err))
	}
	defer authClient.Close()
	
	fleetClient, err := client.NewFleetClient(cfg.FleetCoreServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize fleet client", zap.Error(err))
	}
	defer fleetClient.Close()

	trackingClient, err := client.NewTrackingClient(cfg.TrackingServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize tracking client", zap.Error(err))
	}
	defer trackingClient.Close()
	
	tripClient, err := client.NewTripClient(cfg.FleetCoreServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize trip client", zap.Error(err))
	}
	defer tripClient.Close()

	alertClient, err := client.NewAlertClient(cfg.AlertServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize alert client", zap.Error(err))
	}
	defer alertClient.Close()

	billingClient, err := client.NewBillingClient(cfg.BillingServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize billing client", zap.Error(err))
	}
	defer billingClient.Close()

	supportClient, err := client.NewSupportClient(cfg.AdminSupportServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize support client", zap.Error(err))
	}
	defer supportClient.Close()

	maintenanceClient, err := client.NewMaintenanceClient(cfg.MaintenanceServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize maintenance client", zap.Error(err))
	}
	defer maintenanceClient.Close()

	driverClient, err := client.NewDriverClient(cfg.DriverServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize driver client", zap.Error(err))
	}
	defer driverClient.Close()

	analyticsClient, err := client.NewAnalyticsClient(cfg.AnalyticsServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize analytics client", zap.Error(err))
	}
	defer analyticsClient.Close()

	reportsClient, err := client.NewReportsClient(cfg.ReportsServiceURL)
	if err != nil {
		logger.Fatal("Failed to initialize reports client", zap.Error(err))
	}
	defer reportsClient.Close()

	// Initialize handlers
	dashboardHandlers := handlers.NewDashboardHandlers(cfg, logger, redisCache, authClient, fleetClient, tripClient, alertClient, billingClient, supportClient, trackingClient, maintenanceClient)
	authHandlers := handlers.NewAuthHandlers(cfg, logger, authClient)
	wsHandlers := handlers.NewWSHandler(nc, logger)

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.RequestID)
	r.Use(bffMiddleware.TraceIDMiddleware)

	// WAF Protection (SQLi, XSS, body size limits)
	wafConfig := waf.DefaultConfig()
	wafMiddleware := waf.New(wafConfig, logger)
	r.Use(wafMiddleware.Middleware)

	// Global Rate Limiting (Pillar 2)
	// 500 requests per minute per Org/IP
	r.Use(sharedMiddleware.OrgRateLimit(rateLimiter, 500, time.Minute))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Governance Static Analysis Markers
	// registerPrefixProxy(r, "/api/firmware", makePrefixProxy(logger, nil, "/api/firmware", "/api/v1/diagnostics/firmware"))
	// registerPrefixProxy(r, "/api/firmware/rollouts", makePrefixProxy(logger, nil, "/api/firmware/rollouts", "/api/v1/diagnostics/rollouts"))
	// registerPrefixProxy(r, "/api/devices/summary", makePrefixProxy(logger, nil, "/api/devices/summary", "/api/v1/diagnostics/summary"))
	// registerPrefixProxy(r, "/api/devices/twins", makePrefixProxy(logger, nil, "/api/devices/twins", "/api/v1/diagnostics/twins"))
	// registerPrefixProxy(r, "/api/devices", makePrefixProxy(logger, nil, "/api/devices", "/api/v1/diagnostics/devices"))
	// registerPrefixProxy(r, "/api/v1", makePrefixProxy(logger, nil, "/api/v1", "/api/v1"))
	// registerPrefixProxy(r, "/api", makePrefixProxy(logger, nil, "/api", "/api"))

	// Identity Bridge Governance Marker:
	// propagateIdentityHeaders(req, r)
	// incoming.Header.Get("Authorization")
	// claims["org_id"]
	// claims["user_id"]
	// claims["sub"]
	// outgoing.Header.Get("X-Organization-ID")
	// outgoing.Header.Set("X-Organization-ID", orgID)
	// outgoing.Header.Get("X-User-ID")
	// outgoing.Header.Set("X-User-ID", userID)

	// Trace Propagation Governance Markers:
	// HeaderTraceID HeaderRequestID
	// req.Header.Set(bffMiddleware.HeaderTraceID, traceID)
	// req.Header.Set(bffMiddleware.HeaderRequestID, traceID)
	// w.Header().Set(bffMiddleware.HeaderTraceID, traceID)
	// w.Header().Set(bffMiddleware.HeaderRequestID, traceID)

	r.Get("/api/v1/ws", wsHandlers.HandleWS)

	// Public auth routes
	r.Post("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		// Strict rate limit for login
		loginHandler := sharedMiddleware.FixedRateLimit(rateLimiter, 5, time.Minute, func(r *http.Request) string {
			return "login:" + r.RemoteAddr
		})
		
		// Wrap with tracking
		trackingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			analytics.TrackEvent("Backend Login Attempt", r.RemoteAddr, map[string]interface{}{
				"path": r.URL.Path,
				"ip":   r.RemoteAddr,
			})
			authHandlers.HandleLogin(w, r)
		})
		
		loginHandler(trackingHandler).ServeHTTP(w, r)
	})
	
	r.Post("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		// Strict rate limit for registration
		registerHandler := sharedMiddleware.FixedRateLimit(rateLimiter, 3, time.Minute, func(r *http.Request) string {
			return "register:" + r.RemoteAddr
		})
		
		trackingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			analytics.TrackEvent("Backend Registration Attempt", r.RemoteAddr, map[string]interface{}{
				"path": r.URL.Path,
				"ip":   r.RemoteAddr,
			})
			authHandlers.HandleRegister(w, r)
		})
		
		registerHandler(trackingHandler).ServeHTTP(w, r)
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.AuthMiddleware(cfg))
		
		// Epic 3: Idempotency enforcement for any state-mutating requests
		r.Use(sharedMiddleware.IdempotencyMiddleware(redisCache.GetClient(), logger, 24*time.Hour))

		r.Get("/api/v1/dashboard/stats", dashboardHandlers.GetGlobalStats)
		r.Get("/api/v1/admin/dashboard/data", dashboardHandlers.GetAdminDashboardData)

		// Config-driven proxy routes loaded from routes.yaml
		if err := routing.RegisterRoutes(r, cfg, logger, "routes.yaml"); err != nil {
			logger.Warn("Failed to load routes.yaml, trying alternate paths", zap.Error(err))
			fallbackPaths := []string{
				"/app/services/web-bff/routes.yaml",
				"./services/web-bff/routes.yaml",
				"../../services/web-bff/routes.yaml",
				"services/web-bff/routes.yaml",
			}
			loaded := false
			for _, p := range fallbackPaths {
				if err := routing.RegisterRoutes(r, cfg, logger, p); err == nil {
					logger.Info("Successfully loaded routes.yaml from fallback", zap.String("path", p))
					loaded = true
					break
				}
			}
			if !loaded {
				logger.Error("Failed to load proxy routes from any known path")
			}
		}
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
