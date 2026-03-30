package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/cache"
	shareddb "github.com/fleet-platform/shared/pkg/database"
	sharedmessaging "github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service/internal/config"
	"github.com/flyfleet/fleet-core-service/internal/control"
	"github.com/flyfleet/fleet-core-service/internal/database"
	"github.com/flyfleet/fleet-core-service/internal/docs"
	adminsupportdrhandlers "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/handlers"
	adminsupportdrrepo "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/repository"
	adminsupportdrservice "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/dr/service"
	adminsupportemergencyhandlers "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/emergency/handlers"
	adminsupportemergencyservice "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/emergency/service"
	adminsupportorghandlers "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/organizations/handlers"
	adminsupportorgrepo "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/organizations/repository"
	adminsupportorgservice "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/organizations/service"
	adminsupportsupporthandlers "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/support/handlers"
	adminsupportsupportrepo "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/support/repository"
	adminsupportsupportservice "github.com/flyfleet/fleet-core-service/internal/domains/admin_support/domains/support/service"
	monitoringrepo "github.com/flyfleet/fleet-core-service/internal/domains/monitoring/repository"
	monitorSvc "github.com/flyfleet/fleet-core-service/internal/domains/monitoring/service"
	"github.com/flyfleet/fleet-core-service/internal/gateway"
	dashHandlers "github.com/flyfleet/fleet-core-service/internal/handlers"
	"github.com/flyfleet/fleet-core-service/internal/logadapter"
	coremiddleware "github.com/flyfleet/fleet-core-service/internal/middleware"
	"github.com/flyfleet/fleet-core-service/internal/platform/metrics"
	"github.com/flyfleet/fleet-core-service/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"

	// Domains
	governanceHandlers "github.com/flyfleet/fleet-core-service/internal/domains/governance/handlers"
	govMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/governance/middleware"
	governanceRepo "github.com/flyfleet/fleet-core-service/internal/domains/governance/repository"
	governanceService "github.com/flyfleet/fleet-core-service/internal/domains/governance/service"
	govUtils "github.com/flyfleet/fleet-core-service/internal/domains/governance/utils"
	govWorkers "github.com/flyfleet/fleet-core-service/internal/domains/governance/workers"
	indiav2handlers "github.com/flyfleet/fleet-core-service/internal/domains/india/handlers"
	indiav2repo "github.com/flyfleet/fleet-core-service/internal/domains/india/repository"
	indiav2service "github.com/flyfleet/fleet-core-service/internal/domains/india/service"
	indiav2workers "github.com/flyfleet/fleet-core-service/internal/domains/india/workers"
	integrationsv2handlers "github.com/flyfleet/fleet-core-service/internal/domains/integrations/handlers"
	integrationsv2repo "github.com/flyfleet/fleet-core-service/internal/domains/integrations/repository"
	integrationsv2service "github.com/flyfleet/fleet-core-service/internal/domains/integrations/service"
	reportsv2handlers "github.com/flyfleet/fleet-core-service/internal/domains/reports/handlers"
	reportsv2queue "github.com/flyfleet/fleet-core-service/internal/domains/reports/queue"
	reportsv2repo "github.com/flyfleet/fleet-core-service/internal/domains/reports/repository"
	reportsv2service "github.com/flyfleet/fleet-core-service/internal/domains/reports/service"
	settingsHandlers "github.com/flyfleet/fleet-core-service/internal/domains/settings/handlers"
	settingsMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/settings/middleware"
	settingsRepo "github.com/flyfleet/fleet-core-service/internal/domains/settings/repository"
	settingsService "github.com/flyfleet/fleet-core-service/internal/domains/settings/service"

	alertalerthandlers "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/alerts/handlers"
	alertalertrepo "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/alerts/repository"
	alertalertrule "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/alerts/rule_engine"
	alertalertservice "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/alerts/service"
	alertnotifyhandlers "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/handlers"
	alertnotifyrepo "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/repository"
	alertnotifyservice "github.com/flyfleet/fleet-core-service/internal/domains/alert/domains/notifications/service"
	alertproviders "github.com/flyfleet/fleet-core-service/internal/domains/alert/providers"

	vehHandlers "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/handlers"
	vehProcessor "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/processor"
	vehRepo "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	vehService "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"

	// Finance Domain
	financeHandlers "github.com/flyfleet/fleet-core-service/internal/domains/finance/handlers"
	financeRepo "github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	financeService "github.com/flyfleet/fleet-core-service/internal/domains/finance/service"

	geoEngine "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/engine"
	geoHandlers "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/handlers"
	geoRepo "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/repository"
	geoService "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/service"

	emergencyHandlers "github.com/flyfleet/fleet-core-service/internal/domains/emergency/handlers"
	emergencyRepo "github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	emergencyService "github.com/flyfleet/fleet-core-service/internal/domains/emergency/service"

	trackingv2analytics "github.com/flyfleet/fleet-core-service/internal/domains/tracking/analytics"
	trackingv2handlers "github.com/flyfleet/fleet-core-service/internal/domains/tracking/handlers"
	trackingv2repo "github.com/flyfleet/fleet-core-service/internal/domains/tracking/repository"
	trackingv2service "github.com/flyfleet/fleet-core-service/internal/domains/tracking/service"

	maintenHandlers "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/handlers"
	maintenRepo "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
	maintenService "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/service"

	fuelHandlers "github.com/flyfleet/fleet-core-service/internal/domains/fuel/handlers"
	fuelRepo "github.com/flyfleet/fleet-core-service/internal/domains/fuel/repository"
	fuelService "github.com/flyfleet/fleet-core-service/internal/domains/fuel/service"

	routeHandlers "github.com/flyfleet/fleet-core-service/internal/domains/route/handlers"
	routeRepo "github.com/flyfleet/fleet-core-service/internal/domains/route/repository"
	routeService "github.com/flyfleet/fleet-core-service/internal/domains/route/service"

	diagHandlers "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/handlers"
	diagRepo "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	diagService "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"

	platformcompat "github.com/flyfleet/fleet-core-service/internal/domains/platform_compat"
	platformcompathandlers "github.com/flyfleet/fleet-core-service/internal/domains/platform_compat/handlers"
	platformcompatservice "github.com/flyfleet/fleet-core-service/internal/domains/platform_compat/service"

	analyticsPackage "github.com/flyfleet/fleet-core-service/internal/domains/analytics"
	analyticsCache "github.com/flyfleet/fleet-core-service/internal/domains/analytics/cache"
	analyticsHandlers "github.com/flyfleet/fleet-core-service/internal/domains/analytics/handlers"
	analyticsMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/analytics/middleware"
	analyticsQueue "github.com/flyfleet/fleet-core-service/internal/domains/analytics/queue"
	analyticsRepo "github.com/flyfleet/fleet-core-service/internal/domains/analytics/repository"
	analyticsService "github.com/flyfleet/fleet-core-service/internal/domains/analytics/service"

	billingv2service "github.com/flyfleet/fleet-core-service/internal/domains/billing/service"

	compliancev2handlers "github.com/flyfleet/fleet-core-service/internal/domains/compliance/handlers"
	compliancev2repo "github.com/flyfleet/fleet-core-service/internal/domains/compliance/repository"
	compliancev2service "github.com/flyfleet/fleet-core-service/internal/domains/compliance/service"

	authv2handlers "github.com/flyfleet/fleet-core-service/internal/domains/auth/handlers"
	authv2repo "github.com/flyfleet/fleet-core-service/internal/domains/auth/repository"
	authv2service "github.com/flyfleet/fleet-core-service/internal/domains/auth/service"

	driverHandlers "github.com/flyfleet/fleet-core-service/internal/domains/driver/handlers"
	driverMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/driver/middleware"
	driverRepo "github.com/flyfleet/fleet-core-service/internal/domains/driver/repository"
	driverService "github.com/flyfleet/fleet-core-service/internal/domains/driver/service"

	"github.com/flyfleet/fleet-core-service/internal/domains/tracking"
	trackingRepo "github.com/flyfleet/fleet-core-service/internal/domains/tracking/repository"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	videoHandlers "github.com/flyfleet/fleet-core-service/internal/domains/video/handlers"
	videoRepo "github.com/flyfleet/fleet-core-service/internal/domains/video/repository"
	videoService "github.com/flyfleet/fleet-core-service/internal/domains/video/service"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// OpenTelemetry Observability
	platformmw "github.com/flyfleet/fleet-core-service/internal/platform/middleware"
	observability "github.com/flyfleet/fleet-core-service/internal/platform/observability"

	// Metrics Registration - Import all domain metrics packages to register them with Prometheus
	_ "github.com/flyfleet/fleet-core-service/internal/domains/admin_support"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/alert"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/alerts"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/analytics"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/auth"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/billing"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/compliance"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/customer"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/driver"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/emergency"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/fuel"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/geofencing"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/india"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/integrations"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/maintenance"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/monitoring"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/platform_compat"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/reports"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/route"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/settings"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/tracking"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/trip"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/vehicle"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/video"
	// Sub-package metrics
	_ "github.com/flyfleet/fleet-core-service/internal/domains/billing/pdf"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/metrics"
	_ "github.com/flyfleet/fleet-core-service/internal/domains/governance/metrics"
	// Platform cost metrics
	_ "github.com/flyfleet/fleet-core-service/internal/platform/cost"
)

// updateDatabasePoolMetrics periodically collects stats from each pool and updates Prometheus gauges.
func updateDatabasePoolMetrics(pools *database.DatabasePools) {
	// Helper to update a specific pool's metrics
	updatePool := func(poolName string, pool *pgxpool.Pool) {
		if pool == nil {
			return
		}
		stat := pool.Stat()
		active := float64(stat.AcquiredConns())
		idle := float64(stat.IdleConns())
		max := float64(stat.MaxConns())

		metrics.DatabaseConnectionsActive.WithLabelValues(poolName).Set(active)
		metrics.DatabaseConnectionsIdle.WithLabelValues(poolName).Set(idle)
		metrics.DatabaseConnectionsMax.WithLabelValues(poolName).Set(max)

		if max > 0 {
			util := (active / max) * 100
			metrics.DatabaseConnectionUtilization.WithLabelValues(poolName).Set(util)
		} else {
			metrics.DatabaseConnectionUtilization.WithLabelValues(poolName).Set(0)
		}
	}

	updatePool("critical", pools.Critical)
	updatePool("high", pools.High)
	updatePool("medium", pools.Medium)
	updatePool("low", pools.Low)
}

func main() {
	var sharedRedisClient *redis.Client

	// 1. Initialize Signal-aware Context for Graceful Shutdown
	// Termux/Android Fix: We explicitly do NOT include syscall.SIGHUP here
	// so the process survives when the parent shell closes.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Ignore SIGHUP explicitly
	signal.Ignore(syscall.SIGHUP)

	// 2. Initialize Logger with sensitive data scrubbing
	baseLogger, _ := zap.NewDevelopment()
	logger := zap.New(logadapter.NewScrubberCore(baseLogger.Core()), zap.AddCaller())
	defer logger.Sync()

	// 3. Load Config
	cfg := config.Load()
	logger.Info("Configuration loaded", zap.String("db_url", cfg.DatabaseURL), zap.String("port", cfg.Port))

	// 4. Initialize OpenTelemetry Tracer (if endpoint configured)
	if otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); otelEndpoint != "" {
		shutdown, err := observability.InitTracer(ctx, "fleet-core-service", otelEndpoint)
		if err != nil {
			logger.Warn("Failed to initialize OpenTelemetry tracer", zap.Error(err))
		} else {
			defer shutdown(ctx)
			logger.Info("OpenTelemetry tracer initialized", zap.String("endpoint", otelEndpoint))
		}
	}

	// 3. Initialize Database (Main pgxpool)
	// Termux/Android Hardening: Strictly limit connections to protect phone memory
	pools, err := database.NewDatabasePools(context.Background(), database.PoolConfig{
		DatabaseURL: cfg.DatabaseURL,
		CriticalMax: 25, // Increased for better dev concurrency (auth, emergency, tracking)
		HighMax:     20, // Reduced from 15
		MediumMax:   10, // Reduced from 10
		LowMax:      10,
	})
	if err != nil {
		logger.Fatal("Failed to initialize database pools", zap.Error(err))
	}
	defer pools.Close()

	// 3.1 Start Performance Monitoring
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Update database pool metrics
				updateDatabasePoolMetrics(pools)

				if pools.Critical != nil {
					s := pools.Critical.Stat()
					logger.Debug("DB Pool Saturation Audit", zap.String("pool", "CRITICAL"), zap.Int32("acquired", s.AcquiredConns()), zap.Int32("idle", s.IdleConns()), zap.Int32("max", s.MaxConns()))

					// Auto Pool Shedding logic
					if sharedRedisClient != nil {
						if float64(s.AcquiredConns()) > float64(s.MaxConns())*0.90 {
							if setErr := sharedRedisClient.Set(ctx, coremiddleware.PoolSheddingKey, "1", 30*time.Second).Err(); setErr != nil {
								logger.Debug("Failed to set pool shedding key", zap.Error(setErr))
							}
						} else {
							if delErr := sharedRedisClient.Del(ctx, coremiddleware.PoolSheddingKey).Err(); delErr != nil {
								logger.Debug("Failed to delete pool shedding key", zap.Error(delErr))
							}
						}
					}
				}
				if pools.High != nil {
					s := pools.High.Stat()
					logger.Debug("DB Pool Saturation Audit", zap.String("pool", "HIGH"), zap.Int32("acquired", s.AcquiredConns()), zap.Int32("idle", s.IdleConns()), zap.Int32("max", s.MaxConns()))
				}
				if pools.Medium != nil {
					s := pools.Medium.Stat()
					logger.Debug("DB Pool Saturation Audit", zap.String("pool", "MEDIUM"), zap.Int32("acquired", s.AcquiredConns()), zap.Int32("idle", s.IdleConns()), zap.Int32("max", s.MaxConns()))
				}
				if pools.Low != nil {
					s := pools.Low.Stat()
					logger.Debug("DB Pool Saturation Audit", zap.String("pool", "LOW"), zap.Int32("acquired", s.AcquiredConns()), zap.Int32("idle", s.IdleConns()), zap.Int32("max", s.MaxConns()))
				}
			}
		}
	}()

	// Shims for legacy repositories that still require *sql.DB
	criticalSQL := stdlib.OpenDBFromPool(pools.Critical)
	highSQL := stdlib.OpenDBFromPool(pools.High)
	mediumSQL := stdlib.OpenDBFromPool(pools.Medium)
	lowSQL := stdlib.OpenDBFromPool(pools.Low)
	defer criticalSQL.Close()
	defer highSQL.Close()
	defer mediumSQL.Close()
	defer lowSQL.Close()

	sharedMB, err := sharedmessaging.NewNATSClient(cfg.NatsURL, logger)
	if err != nil {
		logger.Warn("Failed to initialize shared message broker adapter", zap.Error(err))
	}
	if sharedMB != nil {
		defer sharedMB.Close()
	}

	// 4. Initialize NATS (for Geofencing, etc.)
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Warn("Failed to connect to NATS, geofencing features might be limited", zap.Error(err))
	} else {
		defer nc.Close()
	}

	// Initialize global detector for reused across domains
	indiaRepoGlobal := indiav2repo.NewIndiaIntegrationRepository(pools.Low)
	doubleSwipeDetector := billingv2service.NewDoubleSwipeDetector(indiaRepoGlobal, nc, logger)

	// 5. Initialize Redis Cache
	redisCache, err := cache.NewRedisCache(cfg.RedisURL, "", 0, logger)

	var liveCacheRepo trackingRepo.LiveCacheRepository
	if err != nil {
		logger.Warn("Failed to initialize Redis cache, some features might be limited", zap.Error(err))
	} else {
		defer redisCache.Close()

		rOpt, err := redis.ParseURL(cfg.RedisURL)
		if err == nil {
			sharedRedisClient = redis.NewClient(rOpt)
		} else {
			sharedRedisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
		}
		liveCacheRepo = trackingRepo.NewRedisLiveCacheRepository(sharedRedisClient)
	}

	// 6. Initialize Global Pulse Service (Observability)
	monitoringRepo := monitoringrepo.NewMonitoringRepository(pools.Critical)
	// Pulse service uses the internal Postgres wrapper, we inject the Critical pool
	pulseDB := shareddb.WrapPool(pools.Critical, logger)
	pulseService := monitorSvc.NewPulseService(pulseDB, redisCache, nc, logger, monitoringRepo)

	// 6.1 Initialize Domain Metrics
	trackingMetrics := trackingv2.NewMetrics()

	// 6.2 Start pprof server for performance profiling (if enabled)
	if os.Getenv("ENABLE_PPROF") == "true" {
		pprofAddr := os.Getenv("PPROF_ADDR")
		if pprofAddr == "" {
			pprofAddr = "127.0.0.1:6060"
		}
		pprofSrv := &http.Server{
			Addr:    pprofAddr,
			Handler: http.DefaultServeMux,
		}
		go func() {
			logger.Info("Starting pprof server", zap.String("addr", pprofAddr))
			if err := pprofSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("pprof server failed", zap.Error(err))
			}
		}()
		// Ensure pprof server shuts down gracefully
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			pprofSrv.Shutdown(shutdownCtx)
		}()
	}

	// 5. Initialize Router
	r := chi.NewRouter()

	// 5.1 System Guard (Global Kill Switch)
	systemGuard := coremiddleware.NewSystemGuard(sharedRedisClient)

	// Standard Middlewares
	r.Use(systemGuard.Middleware)
	r.Use(chiMiddleware.RequestID)
	r.Use(coremiddleware.TraceIDMiddleware)
	r.Use(platformmw.Observability("fleet-core-service"))
	r.Use(metrics.HTTPMetricsMiddleware(logger))
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(coremiddleware.CORSMiddleware)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Embedder-Policy", "credentialless") // use credentialless instead of require-corp to avoid blocking external assets if needed, or require-corp
			next.ServeHTTP(w, r)
		})
	})
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ENFORCE_BFF_AUTH_GATE")), "true") {
		r.Use(coremiddleware.NewBFFAuthGate(cfg.JWTSecret, []string{
			"/health",
			// "/swagger" removed - admin documentation now requires authentication
			"/api/control/",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/validate-token",
			"/api/v1/ws",
			"/api/v1/ui/",
			"/api/v1/legal/",
			"/api/v1/dashboard/stats",
			// SECURITY FIX: Removed "/api/v1/admin/dashboard/data" from public routes
			// Admin dashboard data must require authentication
		}))
	}

	// Pre-initialize Repositories (needed by multiple routes and workers)
	vRepo := vehRepo.NewVehicleRepository(pools.High, logger)
	vBulkRepo := vehRepo.NewBulkRepository(pools.High)

	// Driver Domain Repositories
	dRepo := driverRepo.NewDriverRepository(pools.High)
	dOnboardingRepo := driverRepo.NewOnboardingRepository(pools.High)
	dShiftRepo := driverRepo.NewShiftRepository(pools.High)
	dTrainingRepo := driverRepo.NewTrainingRepository(pools.High)
	dPayrollRepo := driverRepo.NewPayrollRepository(pools.High)
	dRatingRepo := driverRepo.NewRatingRepository(pools.High)
	dSafetyRepo := driverRepo.NewSafetyRepository(pools.High)
	dMobileSyncRepo := driverRepo.NewMobileSyncRepository(pools.High)

	// Maintenance Domain Repositories
	maintenRepository := maintenRepo.NewMaintenanceRepository(pools.Medium)
	workOrderRepo := maintenRepo.NewWorkOrderRepository(pools.Medium)
	partsRepo := maintenRepo.NewWorkOrderPartsRepository(pools.Medium)
	inventoryRepo := maintenRepo.NewInventoryRepository(pools.Medium)
	maintenSettingsRepo := maintenRepo.NewSettingsRepository(pools.Medium)
	vendorRepo := maintenRepo.NewVendorRepository(pools.Medium)
	inspectionRepo := maintenRepo.NewInspectionRepository(pools.Medium)

	// Driver Domain Services
	dOnboardingSvc := driverService.NewOnboardingService(dOnboardingRepo, logger)
	dShiftSvc := driverService.NewShiftService(dShiftRepo, logger)
	dTrainingSvc := driverService.NewTrainingService(dTrainingRepo, logger)
	dPayrollSvc := driverService.NewPayrollService(dPayrollRepo, logger)
	dRatingSvc := driverService.NewRatingService(dRatingRepo, logger)
	dSafetySvc := driverService.NewSafetyEventService(dSafetyRepo, logger)
	dMobileSyncSvc := driverService.NewMobileSyncService(dMobileSyncRepo, logger)
	dSecurityMid := driverMiddleware.NewDriverSecurityMiddleware(logger)
	dService := driverService.NewDriverService(dRepo, logger, driverService.WithNATS(nc))

	// Driver Domain Handler
	dHandler := driverHandlers.NewDriverHandler(
		dService, dOnboardingSvc, dShiftSvc, dTrainingSvc,
		dPayrollSvc, dRatingSvc, dSafetySvc, dMobileSyncSvc,
		dSecurityMid, logger,
	)

	// Health Check Endpoints (Kubernetes-ready)
	// Simple liveness probe
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "pass",
			"service": "fleet-core-service",
		})
	})

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Comprehensive readiness probe with dependencies
	// SECURITY: Sanitized to not leak internal error details
	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		checks := map[string]interface{}{
			"status":    "pass",
			"service":   "fleet-core-service",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"checks":    map[string]string{},
		}

		// Check database (generic messages only - no error details)
		if pools.Critical != nil {
			if err := pools.Critical.Ping(ctx); err != nil {
				checks["checks"].(map[string]string)["database"] = "unavailable"
				checks["status"] = "fail"
			} else {
				checks["checks"].(map[string]string)["database"] = "ready"
			}
		} else {
			checks["checks"].(map[string]string)["database"] = "unavailable (pool nil)"
			checks["status"] = "fail"
		}

		// Check Redis
		if sharedRedisClient != nil {
			if _, err := sharedRedisClient.Ping(ctx).Result(); err != nil {
				checks["checks"].(map[string]string)["redis"] = "unavailable"
				checks["status"] = "fail"
			} else {
				checks["checks"].(map[string]string)["redis"] = "ready"
			}
		} else {
			checks["checks"].(map[string]string)["redis"] = "not configured"
		}

		// Check NATS
		if nc != nil && nc.IsConnected() {
			checks["checks"].(map[string]string)["nats"] = "connected"
		} else {
			checks["checks"].(map[string]string)["nats"] = "disconnected"
		}

		// Overall status
		statusCode := http.StatusOK
		if checks["status"] == "fail" {
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(checks)
	})

	// Domain-specific readiness (optional query param ?domain=analytics)
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Same as /ready but more compact for k8s
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Has("domain") {
			// Could add domain-specific checks here
			json.NewEncoder(w).Encode(map[string]string{
				"status": "pass",
				"domain": r.URL.Query().Get("domain"),
			})
		} else {
			json.NewEncoder(w).Encode(map[string]string{
				"status": "pass",
			})
		}
	})

	// 6. Initialize Domains & Register Routes
	// Setup Global JWKS Auth Gate for protected routes
	jwksURL := os.Getenv("JWKS_URL")
	if jwksURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8081"
		}
		// Fallback to localhost relative to the port tests run on natively
		jwksURL = "http://127.0.0.1:" + port + "/api/v1/auth/.well-known/jwks.json"
	}
	authGate := coremiddleware.NewJWKSAuthGate(jwksURL, "FlyFleet", "fleet-dashboard", logger)

	r.Route("/api/v1", func(r chi.Router) {
		// IAM Domain (Superseded by auth_v2)
		/*
			orgR := iamrepo.NewOrganizationRepository(dbSQL)
			orgS := iamservice.NewOrganizationService(orgR)
			iamR := iamrepo.NewIAMRepository(db)
			iamS := iamservice.NewIAMService(iamR, orgS, cfg.JWTSecret)
			iamH := iamhandlers.NewIAMHandlers(iamS)
			iamH.RegisterRoutes(r)
		*/

		// Driver Domain
		r.Route("/drivers", func(r chi.Router) {
			// Require elevated roles for mutating operations
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))
			dHandler.RegisterRoutes(r)
		})

		// Vehicle Domain
		vService := vehService.NewVehicleService(vRepo, logger)
		vHandlers := vehHandlers.NewVehicleHandlers(vService, logger)
		r.Route("/vehicles", func(r chi.Router) {
			// Require elevated roles for mutating operations
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))
			vHandlers.RegisterRoutes(r)
		})
		vBulkProcessor := vehProcessor.NewBulkProcessor(vBulkRepo, vRepo, logger)
		vBulkService := vehService.NewBulkService(vBulkRepo, vBulkProcessor, logger)
		vBulkHandlers := vehHandlers.NewBulkHandlers(vBulkService, logger)
		r.Route("/bulk-jobs", func(r chi.Router) {
			// Bulk jobs are mutating operations
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager"))
			vBulkHandlers.RegisterRoutes(r)
		})
		// vTripRepo := vehRepo.NewTripRepository(db)
		// vTripService := vehService.NewTripService(vTripRepo, logger)
		// vTripHandlers := vehHandlers.NewTripHandlers(vTripService, logger)

		// Route Optimization Domain
		roRepo := routeRepo.NewRouteOptimizationRepository(pools.High)
		roService := routeService.NewRouteOptimizationService(roRepo, "") // osrmURL is currently empty as per existing pattern
		roHandlers := routeHandlers.NewRouteOptimizationHandler(roService, logger)
		r.Route("/routes", func(r chi.Router) {
			// Route creation/modification requires elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))
			roHandlers.RegisterRoutes(r)
		})

		// Initialize Platform Compatibility Service
		compatHub := platformcompatservice.NewHub(logger)
		go compatHub.Run(ctx)

		// Bridge NATS alerts to WebSockets
		nc.Subscribe("billing.v2.alerts", func(m *nats.Msg) {
			var alert struct {
				OrgID uuid.UUID `json:"organization_id"`
			}
			if err := json.Unmarshal(m.Data, &alert); err == nil {
				compatHub.BroadcastToOrg(alert.OrgID, m.Data)
			}
		})

		wsGateway := platformcompat.NewWebsocketGateway(nc, compatHub, logger)
		if err := wsGateway.Start(); err != nil {
			logger.Warn("Failed to start websocket gateway", zap.Error(err))
		}

		compatService := platformcompatservice.NewPlatformCompatService(logger)
		platformCompatHandlers := platformcompathandlers.NewHandlers(compatService, compatHub, logger)
		platformCompatHandlers.RegisterRoutes(r)

		// Diagnostics Domain
		diagR := diagRepo.NewDiagnosticsRepository(pools.Low)
		diagHealth := diagService.NewDeviceHealthService(diagR)
		diagTelemetry := diagService.NewTelemetryService(diagR, diagHealth)
		diagCommand := diagService.NewDeviceCommandService(diagR)
		diagRollout := diagService.NewFirmwareRolloutService(diagR)
		diagH := diagHandlers.NewDiagnosticsHandler(diagTelemetry, diagCommand, diagRollout)
		r.Route("/diagnostics", func(r chi.Router) {
			r.Use(authGate.Handler)
			r.Use(coremiddleware.TenantContextMiddleware(logger))
			r.Use(coremiddleware.RequireRole(logger, "super_admin", "org_admin", "fleet_manager", "dispatcher"))
			diagH.RegisterRoutes(r)
		})

		// Reports Domain (migrated v2 in monolith)
		reportsRepo := reportsv2repo.NewPostgresRepository(pools.Medium)
		reportsQueue := reportsv2queue.NewQueueClient(cfg.RedisURL)
		reportsService := reportsv2service.NewReportService(reportsRepo, reportsQueue, logger)
		reportsHandler := reportsv2handlers.NewReportHandler(reportsService, logger)
		r.Route("/reports", func(r chi.Router) {
			// Reports generation may require elevated roles for operations like create/delete
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "reporter"))
			r.Use(systemGuard.LowPriorityShedder)
			r.Use(coremiddleware.TenantContextMiddleware(logger))
			reportsHandler.RegisterRoutes(r)
		})

		// Integrations Domain (migrated v2 in monolith)
		intRepo := integrationsv2repo.NewRepository(lowSQL, logger)
		intService := integrationsv2service.NewIntegrationsService(intRepo, cfg.JWTSecret, logger)
		intHandler := integrationsv2handlers.NewHandlers(intService)
		r.Route("/", func(r chi.Router) {
			// Integrations configuration changes require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin"))
			intHandler.RegisterRoutes(r)
		})

		// India Integration Domain (migrated in monolith)
		indiaRepo := indiav2repo.NewIndiaIntegrationRepository(pools.Low)
		indiaService := indiav2service.NewIndiaIntegrationService(indiaRepo, sharedRedisClient)

		// Initialize India V2 sub-services
		gstService := indiav2service.NewGSTIntegrationService(indiaRepo, logger)

		fastagService := indiav2service.NewFastagIntegrationService(indiaRepo, doubleSwipeDetector, logger)
		vahanService := indiav2service.NewVahanIntegrationService(indiaRepo, sharedRedisClient, logger)
		sarathiService := indiav2service.NewSarathiIntegrationService(indiaRepo, sharedRedisClient, logger)
		ewaybillService := indiav2service.NewEWayBillIntegrationService(indiaRepo, logger)
		bulkService := indiav2service.NewBulkVerificationService(vahanService)
		encryptionService := indiav2service.NewEncryptionService()

		indiaHandlersLegacy := indiav2handlers.NewIndiaIntegrationHandlers(
			indiaService,
			fastagService,
			ewaybillService,
			vahanService,
			sarathiService,
			gstService,
			bulkService,
			encryptionService,
			sharedRedisClient,
		)
		r.Route("/india", func(r chi.Router) {
			// India compliance operations require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "compliance_manager", "fleet_manager"))
			r.Route("/fastag", func(r chi.Router) {
				r.Get("/transactions/{id}", indiaHandlersLegacy.GetFASTagTransaction)
				r.Get("/transactions", indiaHandlersLegacy.ListFASTagTransactions)
			})
			r.Route("/ewaybill", func(r chi.Router) {
				r.Post("/bills", indiaHandlersLegacy.CreateEWayBill)
				r.Post("/validate", indiaHandlersLegacy.ValidateEWayBill)
				r.Get("/bills/{id}", indiaHandlersLegacy.GetEWayBill)
				r.Get("/bills/number/{number}", indiaHandlersLegacy.GetEWayBillByNumber)
				r.Get("/bills", indiaHandlersLegacy.ListEWayBills)
			})
			r.Route("/vahan", func(r chi.Router) {
				r.Post("/verifications", indiaHandlersLegacy.CreateVehicleVerification)
				r.Post("/verify", indiaHandlersLegacy.VerifyVehicle)
				r.Get("/verifications/{id}", indiaHandlersLegacy.GetVehicleVerification)
				r.Get("/verifications/reg/{reg_number}", indiaHandlersLegacy.GetVehicleVerificationByRegNumber)
				r.Get("/verifications", indiaHandlersLegacy.ListVehicleVerifications)
			})
			r.Route("/sarathi", func(r chi.Router) {
				r.Post("/verifications", indiaHandlersLegacy.CreateDriverVerification)
				r.Post("/verify", indiaHandlersLegacy.VerifyDriver)
				r.Get("/verifications/{id}", indiaHandlersLegacy.GetDriverVerification)
				r.Get("/verifications/license/{license_number}", indiaHandlersLegacy.GetDriverVerificationByLicenseNumber)
				r.Get("/verifications", indiaHandlersLegacy.ListDriverVerifications)
			})
			r.Route("/permits", func(r chi.Router) {
				r.Post("/", indiaHandlersLegacy.CreateLocalPermit)
				r.Get("/{id}", indiaHandlersLegacy.GetLocalPermit)
				r.Get("/number/{permit_number}", indiaHandlersLegacy.GetLocalPermitByNumber)
				r.Get("/", indiaHandlersLegacy.ListLocalPermits)
			})
			r.Route("/gst", func(r chi.Router) {
				r.Post("/compliances", indiaHandlersLegacy.CreateGSTCompliance)
				r.Get("/compliances/{id}", indiaHandlersLegacy.GetGSTCompliance)
				r.Get("/compliances/gstin/{gstin}", indiaHandlersLegacy.GetGSTComplianceByGSTIN)
				r.Get("/compliances", indiaHandlersLegacy.ListGSTCompliances)
			})
			r.Route("/config", func(r chi.Router) {
				r.Post("/", indiaHandlersLegacy.CreateIntegrationConfig)
				r.Get("/{id}", indiaHandlersLegacy.GetIntegrationConfig)
				r.Get("/org/{org_id}/type/{service_type}", indiaHandlersLegacy.GetIntegrationConfigByServiceType)
				r.Get("/", indiaHandlersLegacy.ListIntegrationConfigs)
			})
		})

		// Settings Domain (migrated in monolith)
		setRepo := settingsRepo.NewRepository(pools.Low, logger)
		setService := settingsService.NewSettingsService(setRepo, logger)
		setHandlers := settingsHandlers.NewSettingsHandlers(setService, logger)
		r.With(settingsMiddleware.UserContextMiddleware).Route("/settings", func(r chi.Router) {
			// Settings changes require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin"))
			setHandlers.RegisterRoutes(r)
		})

		// Governance Domain (migrated in monolith)
		govRepo := governanceRepo.NewGovernanceRepository(pools.Low)
		govService := governanceService.NewGovernanceService(govRepo, nc, logger)

		// Enterprise Governance Extensions
		govAggregator := governanceService.NewGovernanceDataAggregator(pools.Low, logger)
		govMasker := governanceService.NewMaskingService()
		govGenerator := govUtils.NewExportFileGenerator()
		govExportSvc := governanceService.NewExportService(govRepo, govAggregator, govMasker, govGenerator, nc, logger)
		govDeletionSvc := governanceService.NewDataDeletionService(govRepo, pools.Low, nc, sharedRedisClient, logger)
		govAccessReviewSvc := governanceService.NewAccessReviewService(govRepo, pools.Low, logger)

		// Workers
		govExpiryWorker := govWorkers.NewGovernanceExpiryWorker(govService, logger)
		go govExpiryWorker.Start(ctx)

		govExportWorker := govWorkers.NewExportProcessingWorker(govExportSvc, nc, logger)
		go govExportWorker.Start(ctx)

		govDeletionWorker := govWorkers.NewDeletionWorker(govDeletionSvc, nc, logger)
		go govDeletionWorker.Start(ctx)

		govRetentionWorker := govWorkers.NewGovernanceRetentionWorker(govService, logger)
		go govRetentionWorker.Start(ctx)

		govHandler := governanceHandlers.NewGovernanceHandler(govService, govExportSvc, govAccessReviewSvc, logger)
		r.Route("/governance", func(r chi.Router) {
			// Governance operations require elevated roles (compliance, data deletion, export)
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "compliance_manager", "data_protection_officer"))
			govRateLimiter := govMiddleware.DSARRateLimit(sharedRedisClient, nc, logger)
			govHandler.RegisterRoutes(r, govRateLimiter)
		})

		// Admin Support Domains (migrated in monolith, non-overlapping surfaces)
		adminTicketRepo := adminsupportsupportrepo.NewTicketRepository(pools.Low, logger)
		notificationCh := make(chan adminsupportsupportservice.NotificationEvent, 100)
		adminTicketService := adminsupportsupportservice.NewTicketService(adminTicketRepo, logger, notificationCh)
		adminTicketHandler := adminsupportsupporthandlers.NewTicketHandlers(adminTicketService, logger)
		r.Group(func(r chi.Router) {
			// Admin support endpoints require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin"))
			adminTicketHandler.RegisterRoutes(r)
		})

		adminDRRepo := adminsupportdrrepo.NewDrRepository(pools.Low, logger)
		adminDRService := adminsupportdrservice.NewDrService(adminDRRepo, logger)
		adminDRHandler := adminsupportdrhandlers.NewDrHandlers(adminDRService, logger)
		r.Group(func(r chi.Router) {
			// Admin support endpoints require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin"))
			adminDRHandler.RegisterRoutes(r)
		})

		adminOrgRepo := adminsupportorgrepo.NewOrganizationRepository(pools.Low, logger)
		adminOrgService := adminsupportorgservice.NewOrganizationService(adminOrgRepo, logger)
		adminOrgHandler := adminsupportorghandlers.NewOrganizationHandlers(adminOrgService, logger)
		r.Group(func(r chi.Router) {
			// Admin support endpoints require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin"))
			adminOrgHandler.RegisterRoutes(r)
		})

		adminEmergencyService, err := adminsupportemergencyservice.NewEmergencyService(nc, logger)
		if err != nil {
			logger.Fatal("Failed to initialize emergency service", zap.Error(err))
		}
		adminEmergencyHandler := adminsupportemergencyhandlers.NewEmergencyHandler(adminEmergencyService, logger)
		r.Group(func(r chi.Router) {
			// Admin support endpoints require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin"))
			adminEmergencyHandler.RegisterRoutes(r)
		})

		// Alert Domains (migrated in monolith)
		alertRepo := alertalertrepo.NewAlertRepository(pools.Low)
		alertRuleEngine := alertalertrule.NewRuleEngine(alertRepo, logger)
		alertNotificationService := alertproviders.NewNotificationService(
			alertproviders.NewTwilioClient("", "", ""),
			alertproviders.NewSendGridClient(""),
			alertproviders.NewFirebaseClient(),
			logger,
		)
		alertSvc := alertalertservice.NewAlertService(alertRepo, alertRuleEngine, alertNotificationService, logger)
		alertHandler := alertalerthandlers.NewAlertHandler(alertSvc, logger)
		r.Group(func(r chi.Router) {
			// Alert operations require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "safety_manager", "compliance_manager"))
			alertHandler.RegisterRoutes(r)
		})

		notifyRepo := alertnotifyrepo.NewNotificationRepository(pools.Low)
		notifySvc := alertnotifyservice.NewNotificationService(notifyRepo, logger)
		notifyHandler := alertnotifyhandlers.NewNotificationHandlers(notifySvc, logger)
		r.Group(func(r chi.Router) {
			// Notification operations require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "safety_manager", "compliance_manager"))
			notifyHandler.RegisterRoutes(r)
		})

		// Analytics Domain
		aMetrics := analyticsPackage.NewMetrics()
		aCache := analyticsCache.NewAnalyticsCache(sharedRedisClient, aMetrics, logger)
		aQueue := analyticsQueue.NewAnalyticsQueueClient("localhost:6379") // Placeholder
		aRepo := analyticsRepo.NewAnalyticsRepository(pools.Medium, logger)
		aService := analyticsService.NewAnalyticsService(aRepo, aCache, aQueue, aMetrics, logger)

		aGuard := analyticsMiddleware.NewTenantGuard(pools.Medium, aMetrics, logger)
		aRateLimiter := analyticsMiddleware.NewRateLimiter(aCache, aService, logger)
		aQualityValidator := analyticsMiddleware.NewDataQualityValidator(logger)

		aHandlers := analyticsHandlers.NewAnalyticsHandler(aService, aGuard, logger, aRateLimiter, aQualityValidator)
		r.Group(func(r chi.Router) {
			// Analytics mutations require elevated roles (data stewardship)
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "data_analyst", "analytics_manager"))
			r.Use(systemGuard.LowPriorityShedder)
			aHandlers.RegisterRoutes(r)
		})

		// Tracking Domain (stage-1 local migration: device APIs)
		trkRepo := trackingv2repo.NewDeviceRepository(pools.Critical, logger)
		trkService := trackingv2service.NewDeviceService(trkRepo, cfg.JWTSecret, nil, aService, nc, logger)
		trkHandler := trackingv2handlers.NewDeviceHandler(trkService, logger)
		trkTelemetryHandler := trackingv2handlers.NewTelemetryHandler(shareddb.WrapPool(pools.Critical, logger), sharedMB, sharedRedisClient, logger, trackingMetrics)
		// Protect tracking device management and telemetry routes with role-based mutations
		r.Group(func(r chi.Router) {
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "device_manager"))
			trkHandler.RegisterRoutes(r)
		})
		r.Group(func(r chi.Router) {
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "device_manager"))
			trkTelemetryHandler.RegisterRoutes(r)
		})

		trkAnalytics := trackingv2analytics.NewTelemetryAnalytics(shareddb.WrapPool(pools.Critical, logger), logger)
		trkAnalyticsHandler := trackingv2handlers.NewAnalyticsHandler(trkAnalytics, logger)
		// Avoid clashing with analytics domain routes mounted at /analytics.
		r.Route("/tracking", func(r chi.Router) {
			r.Use(systemGuard.LowPriorityShedder)
			trkAnalyticsHandler.RegisterRoutes(r)
		})

		// Auth Domain (migrated in monolith via REST gateway)
		authLog := logadapter.New(logger)
		authRepo := authv2repo.NewAuthRepository(pools.Critical, authLog)
		identitySvc := authv2service.NewIdentityService(authRepo, logger)
		authSvc := authv2service.NewAuthService(authLog, authRepo, identitySvc, cfg.JWTSecret, sharedRedisClient, nc)
		authHandler := authv2handlers.NewHandlers(authSvc, logger, sharedRedisClient, jwksURL)
		authHandler.RegisterRoutes(r)

		// UI Configuration
		uiHandlers := &dashHandlers.UIHandlers{}
		r.Get("/ui/global_config", uiHandlers.GetGlobalConfig)
		r.Get("/ui/misc/legal_docs", uiHandlers.GetLegalDocs)
		r.Get("/ui/menus/{menuId}", uiHandlers.GetMenuById)

		// Dashboard & Support Domain (consolidated from BFF)
		supportRepo := repository.NewSupportRepository(pools.Critical, logger)
		monitorHandlers := dashHandlers.NewMonitoringHandlers(logger, pulseService)
		dashboardHandlers := dashHandlers.NewDashboardHandlers(logger, authSvc, vService, adminOrgService, pulseService, supportRepo)
		supportHandlers := dashHandlers.NewSupportCompatHandlers(logger, authSvc, adminOrgService, supportRepo)

		// Protected Dashboard & Admin routes
		r.Group(func(r chi.Router) {
			r.Use(authGate.Handler)
			r.Use(coremiddleware.TenantContextMiddleware(logger))
			r.Use(coremiddleware.RequireRole(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))

			r.Get("/dashboard/stats", dashboardHandlers.GetGlobalStats)
			r.Get("/dashboard/overview", dashboardHandlers.GetAdminDashboardData)
			r.Get("/dashboard/data", dashboardHandlers.GetAdminDashboardData)
			r.Get("/admin/dashboard/data", dashboardHandlers.GetAdminDashboardData)
		})

		// Protected User Profile
		r.Group(func(r chi.Router) {
			r.Use(authGate.Handler)
			r.Use(coremiddleware.TenantContextMiddleware(logger))

			r.Get("/auth/profile", dashboardHandlers.GetProfile)
		})
		r.Route("/support", func(r chi.Router) {
			r.Use(coremiddleware.RequireRole(logger, "super_admin", "org_admin", "support_agent"))

			r.Get("/dashboard/trends", dashboardHandlers.GetDashboardTrends)
			r.Get("/dashboard/top-errors", dashboardHandlers.GetTopErrors)
			r.Get("/dashboard/summary", dashboardHandlers.GetGlobalStats)
			r.Get("/dashboard/recent-activity", dashboardHandlers.GetAdminDashboardData)

			r.Get("/slas", monitorHandlers.GetSLAMetrics)
			r.Get("/tickets", supportHandlers.GetIncidents)

			r.Route("/incidents", func(r chi.Router) {
				r.Get("/", supportHandlers.GetIncidents)
				r.Get("/metrics", supportHandlers.GetIncidentMetrics)
				r.Get("/trends", dashboardHandlers.GetDashboardTrends)
				r.Get("/feed", supportHandlers.GetIncidentFeed)
				r.Get("/detail", supportHandlers.GetIncidentDetail)
				r.Post("/comment", supportHandlers.AddUserNote) // Compat for comment
				r.Patch("/update", supportHandlers.AddUserNote) // Compat for update
			})

			r.Route("/users", func(r chi.Router) {
				r.Get("/", supportHandlers.SearchUsers)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/profile", supportHandlers.GetUserProfile)
					r.Get("/security", supportHandlers.GetUserSecurity)
					r.Get("/sessions", supportHandlers.GetUserSessions)
					r.Post("/sessions/{sessionId}/revoke", supportHandlers.RevokeSession)
					r.Get("/logins", supportHandlers.GetUserLoginHistory)
					r.Get("/roles", supportHandlers.GetUserRoles)
					r.Get("/permissions", supportHandlers.GetUserPermissions)
					r.Get("/notes", supportHandlers.GetUserNotes)
					r.Post("/notes", supportHandlers.AddUserNote)

					// Admin Actions
					r.Post("/password-reset", supportHandlers.ResetPassword)
					r.Post("/unlock", supportHandlers.UnlockAccount)
					r.Post("/mfa-reset", supportHandlers.ResetMFA)
					r.Post("/logout", supportHandlers.ForceLogout)
				})
			})
		})

		// Monitoring Domain
		r.Get("/monitoring/health", monitorHandlers.GetSystemHealth)
		r.Get("/monitoring/sla", monitorHandlers.GetSLAMetrics)

		// ML Service Proxy (to Sidecar) - SECURITY FIX: Require authentication
		if cfg.MLServiceURL != "" {
			mlProxy, err := gateway.NewServiceProxy("/api/v1/ml", cfg.MLServiceURL, logger)
			if err == nil {
				// Wrap ML routes with authGate to prevent unauthenticated access
				r.Group(func(r chi.Router) {
					r.Use(authGate.Handler)
					r.Use(coremiddleware.TenantContextMiddleware(logger))
					r.Handle("/ml", mlProxy)
					r.Handle("/ml/*", mlProxy)
				})
			}
		}

		// Compliance Domain (migrated in monolith via REST gateway)
		compLog := logadapter.New(logger)
		compRepo := compliancev2repo.NewComplianceRepository(pools.High, compLog)
		compSvc := compliancev2service.NewComplianceService(compLog, compRepo)
		compHandler := compliancev2handlers.NewHandlers(compSvc, logger)
		compHandler.RegisterRoutes(r)

		// Finance Domain
		finRepo := financeRepo.NewFinanceRepository(pools.High)
		finService := financeService.NewFinanceService(finRepo, logger)
		finHandlers := financeHandlers.NewFinanceHandlers(finService, logger)
		r.Route("/finance", func(r chi.Router) {
			// Finance operations require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "finance_manager"))
			finHandlers.RegisterRoutes(r)
		})

		// Fuel Domain
		fuelLogRepo := fuelRepo.NewFuelLogRepository(pools.High)
		fuelSettingsRepo := fuelRepo.NewSettingsRepository(pools.High)
		anomalySvc := fuelService.NewAnomalyDetectionService(nil, logger)
		taxSvc := financeService.NewTaxCalculationService()
		fuelSvc := fuelService.NewFuelService(fuelLogRepo, fuelSettingsRepo, anomalySvc, taxSvc, logger)
		fuelHandler := fuelHandlers.NewFuelHandler(fuelSvc, logger)
		r.Route("/fuel", func(r chi.Router) {
			// Fuel logs require elevated roles for modifications
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))
			fuelHandler.RegisterRoutes(r)
		})

		// Maintenance Domain
		workOrderService := maintenService.NewWorkOrderService(
			workOrderRepo,
			partsRepo,
			inventoryRepo,
			maintenSettingsRepo,
			vendorRepo,
			inspectionRepo,
			maintenRepository,
			nc,
			logger,
		)
		maintenServiceSvc := maintenService.NewMaintenanceService(maintenRepository, logger)
		maintenHandler := maintenHandlers.NewMaintenanceHandler(maintenServiceSvc, logger)
		workOrderHandler := maintenHandlers.NewWorkOrderHandler(workOrderService, logger)
		r.Route("/maintenance", func(r chi.Router) {
			// Maintenance records require elevated roles for changes
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))
			maintenHandler.RegisterRoutes(r)
		})
		r.Route("/work-orders", func(r chi.Router) {
			// Work order management requires elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "mechanic"))
			workOrderHandler.RegisterRoutes(r)
		})
		// Emergency Domain
		emRepo := emergencyRepo.NewEmergencyRepository(pools.Critical)
		emAnalyticsSvc := emergencyService.NewEmergencyAnalyticsService(emRepo)
		emRetry := emergencyService.NewNotificationRetryService(nc, logger)
		emGeo := emergencyService.NewEmergencyGeoService()
		emAuth := emergencyService.NewEmergencyAuthorityService(emRepo, nc, logger)
		emSeverity := emergencyService.NewEmergencySeverityEngine()
		emHealth := emergencyService.NewProviderHealthService(sharedRedisClient, logger)
		emFallback := emergencyService.NewProviderFallbackManager(emHealth, logger)

		emService := emergencyService.NewEmergencyService(
			emRepo,
			liveCacheRepo,
			nc,
			sharedRedisClient,
			logger,
			emRetry,
			emGeo,
			emAuth,
			emSeverity,
			emFallback,
		)
		// Geofencing Domain
		gRepo := geoRepo.NewGeofenceRepository(pools.High)
		gService := geoService.NewGeofenceService(gRepo, nc, logger)
		gHandlers := geoHandlers.NewGeofenceHandler(gService, logger)
		r.Route("/geofencing", func(r chi.Router) {
			// Geofence management requires elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher"))
			gHandlers.RegisterRoutes(r)
		})
		emHandlers := emergencyHandlers.NewEmergencyHandlers(emService, emAnalyticsSvc, logger)
		r.Route("/emergency", func(r chi.Router) {
			// Emergency response endpoints require elevated roles
			r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "emergency_responder"))
			emHandlers.RegisterRoutes(r)
		})
	})

	// Video Domain
	vidRepo := videoRepo.NewVideoRepository(pools.High)
	vidService := videoService.NewVideoService(vidRepo, nc, logger)
	vidHandlers := videoHandlers.NewVideoHandler(vidService, logger)
	r.Route("/api/v1/video", func(r chi.Router) {
		// Video access and management requires elevated roles
		r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "fleet_manager", "dispatcher", "safety_manager"))
		vidHandlers.RegisterRoutes(r)
	})

	// India legacy compatibility routes (/api/*) for existing Web-BFF mappings.
	indiaRepoLegacy := indiav2repo.NewIndiaIntegrationRepository(pools.Low)
	indiaServiceLegacy := indiav2service.NewIndiaIntegrationService(indiaRepoLegacy, sharedRedisClient)
	fastagServiceLegacy := indiav2service.NewFastagIntegrationService(indiaRepoLegacy, doubleSwipeDetector, logger)
	ewayBillServiceLegacy := indiav2service.NewEWayBillIntegrationService(indiaRepoLegacy, logger)
	vahanServiceLegacy := indiav2service.NewVahanIntegrationService(indiaRepoLegacy, sharedRedisClient, logger)
	sarathiServiceLegacy := indiav2service.NewSarathiIntegrationService(indiaRepoLegacy, sharedRedisClient, logger)
	gstServiceLegacy := indiav2service.NewGSTIntegrationService(indiaRepoLegacy, logger)
	bulkServiceLegacy := indiav2service.NewBulkVerificationService(vahanServiceLegacy)
	encryptServiceLegacy := indiav2service.NewEncryptionService()

	// 4. Pluggable Domain Initialization (Non-blocking)

	indiaComplianceWorker := indiav2workers.NewIndiaComplianceWorker(indiaRepoLegacy, sharedMB, sharedRedisClient, logger)
	go indiaComplianceWorker.Start(ctx)

	indiaHandlersLegacy := indiav2handlers.NewIndiaIntegrationHandlers(
		indiaServiceLegacy,
		fastagServiceLegacy,
		ewayBillServiceLegacy,
		vahanServiceLegacy,
		sarathiServiceLegacy,
		gstServiceLegacy,
		bulkServiceLegacy,
		encryptServiceLegacy,
		sharedRedisClient,
	)
	r.Route("/api", func(r chi.Router) {
		// India compliance endpoints require elevated roles (regulatory)
		r.Use(coremiddleware.RequireRoleForMutations(logger, "super_admin", "org_admin", "admin", "compliance_manager", "india_compliance_officer"))
		r.Route("/fastag", func(r chi.Router) {
			r.Post("/transactions", indiaHandlersLegacy.CreateFASTagTransaction)
			r.Post("/sync", indiaHandlersLegacy.SyncFASTag)
			r.Get("/transactions/{id}", indiaHandlersLegacy.GetFASTagTransaction)
			r.Get("/transactions", indiaHandlersLegacy.ListFASTagTransactions)
		})
		r.Route("/ewaybill", func(r chi.Router) {
			r.Post("/bills", indiaHandlersLegacy.CreateEWayBill)
			r.Post("/validate", indiaHandlersLegacy.ValidateEWayBill)
			r.Get("/bills/{id}", indiaHandlersLegacy.GetEWayBill)
			r.Get("/bills/number/{number}", indiaHandlersLegacy.GetEWayBillByNumber)
			r.Get("/bills", indiaHandlersLegacy.ListEWayBills)
		})
		r.Route("/vahan", func(r chi.Router) {
			r.Post("/verifications", indiaHandlersLegacy.CreateVehicleVerification)
			r.Post("/verify", indiaHandlersLegacy.VerifyVehicle)
			r.Get("/verifications/{id}", indiaHandlersLegacy.GetVehicleVerification)
			r.Get("/verifications/reg/{reg_number}", indiaHandlersLegacy.GetVehicleVerificationByRegNumber)
			r.Get("/verifications", indiaHandlersLegacy.ListVehicleVerifications)
		})
		r.Route("/sarathi", func(r chi.Router) {
			r.Post("/verifications", indiaHandlersLegacy.CreateDriverVerification)
			r.Post("/verify", indiaHandlersLegacy.VerifyDriver)
			r.Get("/verifications/{id}", indiaHandlersLegacy.GetDriverVerification)
			r.Get("/verifications/license/{license_number}", indiaHandlersLegacy.GetDriverVerificationByLicenseNumber)
			r.Get("/verifications", indiaHandlersLegacy.ListDriverVerifications)
		})
		r.Route("/permits", func(r chi.Router) {
			r.Post("/", indiaHandlersLegacy.CreateLocalPermit)
			r.Get("/{id}", indiaHandlersLegacy.GetLocalPermit)
			r.Get("/number/{permit_number}", indiaHandlersLegacy.GetLocalPermitByNumber)
			r.Get("/", indiaHandlersLegacy.ListLocalPermits)
		})
		r.Route("/gst", func(r chi.Router) {
			r.Post("/compliances", indiaHandlersLegacy.CreateGSTCompliance)
			r.Get("/compliances/{id}", indiaHandlersLegacy.GetGSTCompliance)
			r.Get("/compliances/gstin/{gstin}", indiaHandlersLegacy.GetGSTComplianceByGSTIN)
			r.Get("/compliances", indiaHandlersLegacy.ListGSTCompliances)
		})
		r.Route("/config", func(r chi.Router) {
			r.Post("/", indiaHandlersLegacy.CreateIntegrationConfig)
			r.Get("/{id}", indiaHandlersLegacy.GetIntegrationConfig)
			r.Get("/org/{org_id}/type/{service_type}", indiaHandlersLegacy.GetIntegrationConfigByServiceType)
			r.Get("/", indiaHandlersLegacy.ListIntegrationConfigs)
		})
	})

	// Web-BFF compatibility prefix rewrites now handled in monolith.
	type bffRewrite struct {
		From string
		To   string
	}
	bffRewrites := []bffRewrite{
		{From: "/api/v1/admin/governance", To: "/api/v1/governance"},
		// DR rewrite compatibility (web-bff rewrite: /restores -> /restore-jobs)
		{From: "/api/v1/admin/dr/restores", To: "/api/v1/dr/restore-jobs"},
		{From: "/api/v1/admin/dr", To: "/api/v1/dr"},
		{From: "/api/v1/admin/reports", To: "/api/v1/reports"},
		{From: "/api/v1/admin/notifications", To: "/api/v1/notifications"},
		{From: "/api/v1/admin/preferences", To: "/api/v1/notifications/preferences"},
		{From: "/api/v1/admin/billing-settings", To: "/api/v1/billing/settings"},
		{From: "/api/v1/admin/billing-metrics", To: "/api/v1/billing/metrics"},
		{From: "/api/v1/admin/billing", To: "/api/v1/billing"},
		{From: "/api/v1/admin/settings", To: "/api/v1/settings"},
		{From: "/api/v1/admin/settings/platform", To: "/api/v1/settings"},
		{From: "/api/v1/admin/work-orders", To: "/api/v1/work-orders"},
		{From: "/api/v1/admin/inventory", To: "/api/v1/inventory"},
		{From: "/api/v1/admin/vendors", To: "/api/v1/vendors"},
		{From: "/api/v1/admin/inspections", To: "/api/v1/inspections"},
		{From: "/api/v1/admin/fuel", To: "/api/v1/fuel"},
		{From: "/api/v1/admin/finance", To: "/api/v1/finance"},
		{From: "/api/v1/admin/revenue", To: "/api/v1/finance/revenue"},
		{From: "/api/v1/admin/fuel-logs", To: "/api/v1/fuel"},
		{From: "/api/v1/admin/video", To: "/api/v1/video"},
		{From: "/api/v1/admin/emergency", To: "/api/v1/emergency"},
		{From: "/api/v1/admin/organizations", To: "/api/v1/organizations"},
		{From: "/api/v1/admin/devices", To: "/api/v1/diagnostics/devices"},
		{From: "/api/v1/emergency", To: "/api/v1/emergency"},
		// Legacy non-v1 aliases intentionally removed to keep monolith surface canonical on /api/v1.
		{From: "/api/search", To: "/api/v1/search"},
		{From: "/api/timeline", To: "/api/v1/timeline"},
		{From: "/api/v1/search", To: "/api/v1/search"},
		{From: "/api/v1/timeline", To: "/api/v1/timeline"},
		{From: "/api/v1/monitoring", To: "/api/v1/monitoring"},
		{From: "/api/v1/devices/summary", To: "/api/v1/diagnostics/summary"},
		{From: "/api/v1/devices/status", To: "/api/v1/diagnostics/status"},
		{From: "/api/v1/devices/twins", To: "/api/v1/diagnostics/twins"},
		{From: "/api/v1/devices", To: "/api/v1/diagnostics/devices"},
		{From: "/api/v1/firmware/rollouts", To: "/api/v1/diagnostics/rollouts"},
		{From: "/api/v1/firmware", To: "/api/v1/diagnostics/firmware"},
		{From: "/api/v1/geofences", To: "/api/v1/geofencing"},
		{From: "/api/v1/driver", To: "/api/v1/drivers"},
		{From: "/api/v1/tracking/telemetry/gps", To: "/api/v1/device/telemetry/gps"},
		{From: "/api/v1/tracking/telemetry/obd", To: "/api/v1/device/telemetry/obd"},
		{From: "/api/v1/tracking/vehicles", To: "/api/v1/vehicles"},
		// Support rewrite compatibility (web-bff rewrite function)
		{From: "/api/v1/support/dashboard/stats", To: "/api/v1/support/dashboard/summary"},
		{From: "/api/v1/support/incidents/feed", To: "/api/v1/support/incidents"},
		{From: "/api/v1/support/incidents/metrics", To: "/api/v1/support/dashboard/summary"},
		{From: "/api/v1/support/incidents/trends", To: "/api/v1/support/dashboard/recent-activity"},
		{From: "/api/v1/support", To: "/api/v1/support"},
	}
	for _, rw := range bffRewrites {
		if rw.From == rw.To {
			continue
		}
		h := gateway.NewRewriteProxy(rw.From, rw.To, logger)
		r.Handle(rw.From, h)
		r.Handle(rw.From+"/*", h)
	}

	// Optional upstream proxy routes let clients use this service as a single control-plane entrypoint.
	upstreamTargets := map[string]string{
		"auth":         cfg.AuthServiceURL,
		"alert":        cfg.AlertServiceURL,
		"admin":        cfg.AdminSupportServiceURL,
		"analytics":    cfg.AnalyticsServiceURL,
		"billing":      cfg.BillingServiceURL,
		"compliance":   cfg.ComplianceServiceURL,
		"governance":   cfg.GovernanceServiceURL,
		"integrations": cfg.IntegrationsServiceURL,
		"maintenance":  cfg.MaintenanceServiceURL,
		"reports":      cfg.ReportsServiceURL,
		"settings":     cfg.SettingsServiceURL,
		"tracking":     cfg.TrackingServiceURL,
		"trip":         cfg.TripServiceURL,
		"vehicle":      cfg.VehicleServiceURL,
	}
	enabledUpstreams := map[string]string{}
	localMigratedServices := []string{"admin", "alert", "analytics", "auth", "billing", "compliance", "driver", "governance", "india", "integrations", "maintenance", "reports", "settings", "tracking", "trip", "vehicle"}
	upstreamNames := make([]string, 0, len(upstreamTargets))
	for name := range upstreamTargets {
		upstreamNames = append(upstreamNames, name)
	}
	sort.Strings(upstreamNames)
	for _, name := range upstreamNames {
		// locally migrated routes are served by this monolith.
		if name == "settings" || name == "governance" || name == "integrations" || name == "reports" || name == "maintenance" || name == "analytics" || name == "trip" || name == "vehicle" || name == "billing" || name == "admin" || name == "alert" || name == "auth" || name == "compliance" || name == "tracking" || name == "driver" {
			continue
		}
		target := strings.TrimSpace(upstreamTargets[name])
		if target == "" {
			continue
		}
		if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
			logger.Warn("Skipping non-http upstream target", zap.String("name", name), zap.String("target", target))
			continue
		}
		prefix := "/api/edge/" + name
		proxy, err := gateway.NewServiceProxy(prefix, target, logger)
		if err != nil {
			logger.Warn("Skipping invalid upstream target", zap.String("name", name), zap.String("target", target), zap.Error(err))
			continue
		}
		r.Handle(prefix, proxy)
		r.Handle(prefix+"/*", proxy)
		enabledUpstreams[name] = target
	}
	r.Get("/api/control/upstreams", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"mode":           "monolith-control-plane",
			"local":          "fleet-core-service-legacy",
			"local_migrated": localMigratedServices,
			"upstreams":      enabledUpstreams,
		})
	})
	r.Get("/api/control/migrations", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"service":          "fleet-core-service-legacy",
			"local_migrated":   localMigratedServices,
			"upstream_pending": enabledUpstreams,
		})
	})
	r.Get("/api/control/upstreams/health", func(w http.ResponseWriter, req *http.Request) {
		names := make([]string, 0, len(enabledUpstreams))
		for name := range enabledUpstreams {
			names = append(names, name)
		}
		sort.Strings(names)

		items := make([]control.UpstreamHealth, 0, len(names))
		okCount := 0
		for _, name := range names {
			item := control.CheckUpstreamHealth(req.Context(), name, enabledUpstreams[name])
			if item.Reachable {
				okCount++
			}
			items = append(items, item)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"service":        "fleet-core-service-legacy",
			"upstream_total": len(names),
			"reachable":      okCount,
			"upstreams":      items,
		})
	})
	r.Get("/api/control/apis/routes", func(w http.ResponseWriter, _ *http.Request) {
		routes, err := control.LocalRoutes(r)
		if err != nil {
			http.Error(w, "failed to enumerate local routes", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "fleet-core-service-legacy",
			"count":   len(routes),
			"routes":  routes,
		})
	})
	r.Get("/api/control/apis/counts", func(w http.ResponseWriter, req *http.Request) {
		localRoutes, err := control.LocalRoutes(r)
		if err != nil {
			http.Error(w, "failed to enumerate local routes", http.StatusInternalServerError)
			return
		}
		localUnique := control.UniqueMethodPathCount(localRoutes)

		upstreamNames := make([]string, 0, len(enabledUpstreams))
		for name := range enabledUpstreams {
			upstreamNames = append(upstreamNames, name)
		}
		sort.Strings(upstreamNames)

		upstreamCounts := make([]control.UpstreamCount, 0, len(upstreamNames))
		upstreamTotal := 0
		for _, name := range upstreamNames {
			target := enabledUpstreams[name]
			count := control.CountUpstreamOpenAPI(req.Context(), name, target)
			upstreamCounts = append(upstreamCounts, count)
			upstreamTotal += count.Count
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"service":                   "fleet-core-service-legacy",
			"local_declared_routes":     len(localRoutes),
			"local_unique_method_paths": localUnique,
			"upstream_openapi_total":    upstreamTotal,
			"upstream_counts":           upstreamCounts,
			"combined_estimated_total":  localUnique + upstreamTotal,
		})
	})
	r.Get("/api/control/apis/combined-report", func(w http.ResponseWriter, _ *http.Request) {
		serverURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
		spec, err := docs.BuildCombinedOpenAPIJSON(r, serverURL, enabledUpstreams)
		if err != nil {
			http.Error(w, "failed to build combined spec", http.StatusInternalServerError)
			return
		}
		var payload struct {
			Paths map[string]map[string]interface{} `json:"paths"`
		}
		if err := json.Unmarshal(spec, &payload); err != nil {
			http.Error(w, "failed to parse combined spec", http.StatusInternalServerError)
			return
		}
		report := control.AnalyzeOpenAPIPaths(payload.Paths)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "fleet-core-service-legacy",
			"report":  report,
		}); err != nil {
			logger.Debug("Failed to encode response", zap.Error(err))
		}
	})
	r.Get("/api/control/overview", func(w http.ResponseWriter, req *http.Request) {
		localRoutes, err := control.LocalRoutes(r)
		if err != nil {
			http.Error(w, "failed to enumerate local routes", http.StatusInternalServerError)
			return
		}
		localUnique := control.UniqueMethodPathCount(localRoutes)

		names := make([]string, 0, len(enabledUpstreams))
		for name := range enabledUpstreams {
			names = append(names, name)
		}
		sort.Strings(names)

		upstreamCounts := make([]control.UpstreamCount, 0, len(names))
		upstreamHealth := make([]control.UpstreamHealth, 0, len(names))
		upstreamTotal := 0
		reachable := 0
		for _, name := range names {
			target := enabledUpstreams[name]
			c := control.CountUpstreamOpenAPI(req.Context(), name, target)
			h := control.CheckUpstreamHealth(req.Context(), name, target)
			upstreamCounts = append(upstreamCounts, c)
			upstreamHealth = append(upstreamHealth, h)
			upstreamTotal += c.Count
			if h.Reachable {
				reachable++
			}
		}

		serverURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
		spec, err := docs.BuildCombinedOpenAPIJSON(r, serverURL, enabledUpstreams)
		if err != nil {
			http.Error(w, "failed to build combined spec", http.StatusInternalServerError)
			return
		}
		var payload struct {
			Paths map[string]map[string]interface{} `json:"paths"`
		}
		if err := json.Unmarshal(spec, &payload); err != nil {
			http.Error(w, "failed to parse combined spec", http.StatusInternalServerError)
			return
		}
		report := control.AnalyzeOpenAPIPaths(payload.Paths)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"service":                   "fleet-core-service-legacy",
			"local_declared_routes":     len(localRoutes),
			"local_unique_method_paths": localUnique,
			"upstream_openapi_total":    upstreamTotal,
			"combined_estimated_total":  localUnique + upstreamTotal,
			"upstream_total":            len(names),
			"upstream_reachable":        reachable,
			"upstream_health":           upstreamHealth,
			"upstream_counts":           upstreamCounts,
			"combined_duplicate_report": report,
		}); err != nil {
			logger.Debug("Failed to encode response", zap.Error(err))
		}
	})

	// 7. Initialize and Start Background Workers
	if nc != nil {
		ge := geoEngine.NewGeofenceEngine(pools.High, nc, logger)
		go func() {
			if err := ge.Start(context.Background()); err != nil {
				logger.Error("Geofence engine failed", zap.Error(err))
			}
		}()

		tw := maintenService.NewTelemetryWorker(maintenRepository, nc, logger)
		go func() {
			if err := tw.Start(context.Background()); err != nil {
				logger.Error("Telemetry worker failed", zap.Error(err))
			}
		}()
	}

	// 8. Start Server
	// Admin API Documentation - RESTRICTED TO ADMIN ONLY
	r.Route("/api/swagger", func(r chi.Router) {
		// Require admin authentication
		r.Use(coremiddleware.RequireRole(logger, "super_admin", "admin"))
		r.Get("/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
			serverURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
			spec, err := docs.BuildOpenAPIJSON(r, serverURL)
			if err != nil {
				logger.Error("Failed to generate OpenAPI", zap.Error(err))
				http.Error(w, "failed to generate openapi", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(spec)
		})
		r.Get("/openapi.combined.json", func(w http.ResponseWriter, _ *http.Request) {
			serverURL := fmt.Sprintf("http://localhost:%s", cfg.Port)
			spec, err := docs.BuildCombinedOpenAPIJSON(r, serverURL, enabledUpstreams)
			if err != nil {
				logger.Error("Failed to generate combined OpenAPI", zap.Error(err))
				http.Error(w, "failed to generate combined openapi", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(spec)
		})
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>FlyFleet Monolith Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/api/swagger/openapi.combined.json",
      dom_id: "#swagger-ui"
    });
  </script>
</body>
</html>`))
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		logger.Info("Starting fleet-core-service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for Signal
	<-ctx.Done()
	logger.Info("Context triggered shutdown", zap.String("reason", fmt.Sprintf("%v", ctx.Err())))

	logger.Info("Shutting down fleet-core-service...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Fleet-core-service exited")
}
