package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/config"
	"github.com/flyfleet/fleet-core/internal/database"
	"github.com/flyfleet/fleet-core/internal/domains/iam"
	iamcfg "github.com/flyfleet/fleet-core/internal/domains/iam/config"
	driverMiddleware "github.com/flyfleet/fleet-core/internal/domains/driver/middleware"
	drvRepo "github.com/flyfleet/fleet-core/internal/domains/driver/repository"
	driverService "github.com/flyfleet/fleet-core/internal/domains/driver/service"
	driverHandlers "github.com/flyfleet/fleet-core/internal/domains/driver/handlers"
	vh "github.com/flyfleet/fleet-core/internal/domains/vehicle/handlers"
	vrepo "github.com/flyfleet/fleet-core/internal/domains/vehicle/repository"
	vs "github.com/flyfleet/fleet-core/internal/domains/vehicle/service"
	orgHandlers "github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/handlers"
	orgRepo "github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/repository"
	orgSvc "github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/service"
	"github.com/flyfleet/fleet-core/internal/middleware"
	th "github.com/flyfleet/fleet-core/internal/tracking/handlers"
	tr "github.com/flyfleet/fleet-core/internal/tracking/repository"
	ts "github.com/flyfleet/fleet-core/internal/tracking/service"
	"github.com/flyfleet/fleet-core/internal/tracking/worker"
)

func main() {
	// --- Bootstrap ---
	cfg := config.Load()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("starting fleet-core service",
		zap.String("port", cfg.Port),
		zap.String("environment", cfg.Environment))

	ctx := context.Background()

	// --- Database ---
	pool, err := database.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}
	defer pool.Close()

	// --- Redis ---
	rdb, err := database.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		logger.Fatal("redis connection failed", zap.Error(err))
	}
	defer rdb.Close()

	// --- Repositories ---
	telemRepo := tr.NewTelemetryRepository(pool, logger)
	liveCache := tr.NewLiveCacheRepository(rdb)
	geoRepo := tr.NewGeofenceRepository(pool)
	vehicleRepo := tr.NewVehicleRepository(pool)               // tracking layer
	vehicleDomainRepo := vrepo.NewVehicleRepository(pool, logger) // vehicle domain (CRUD)

	// --- Services ---
	tenantChecker := &pgTenantChecker{pool: pool}
	ingestionSvc := ts.NewIngestionService(telemRepo, liveCache, rdb, logger)
	querySvc := ts.NewTelemetryQueryService(telemRepo, liveCache, tenantChecker, logger)
	batchSvc := ts.NewBatchLocationService(vehicleRepo, liveCache, telemRepo, logger, cfg.Flags.MaxVehiclesOnMap)

	// --- IAM Domain (conditional) ---
	var iamModule *iam.Module
	if cfg.IAM.Enabled {
		iamCfg := iamcfg.IAMConfig{
			JWTSecret:             cfg.JWTSecret,
			JWTAccessTTL:          cfg.IAM.JWTAccessTTL,
			JWTRefreshTTL:         cfg.IAM.JWTRefreshTTL,
			BcryptCost:            cfg.IAM.BcryptCost,
			MaxLoginAttempts:      cfg.IAM.MaxLoginAttempts,
			LockoutDuration:       cfg.IAM.LockoutDuration,
			PasswordHistoryLength: cfg.IAM.PasswordHistoryLen,
			Email: iamcfg.EmailConfig{
				Provider:       cfg.IAM.EmailProvider,
				SendGridAPIKey: cfg.IAM.EmailSendGridKey,
				SMTPHost:       cfg.IAM.EmailSMTPHost,
				SMTPPort:       cfg.IAM.EmailSMTPPort,
				SMTPUser:       cfg.IAM.EmailSMTPUser,
				SMTPPassword:   cfg.IAM.EmailSMTPPassword,
				FromAddress:    cfg.IAM.EmailFromAddress,
				FromName:       cfg.IAM.EmailFromName,
			},
		}
		iamModule = iam.NewModule(pool, rdb, cfg.JWTSecret, iamCfg, logger)
		logger.Info("IAM domain initialized",
			zap.Int("bcrypt_cost", cfg.IAM.BcryptCost),
			zap.String("email_provider", cfg.IAM.EmailProvider),
		)
	}

	// --- Handlers ---
	ingestHandler := th.NewIngestHandler(ingestionSvc, logger)
	locationsHandler := th.NewLocationsHandler(batchSvc, logger)
	historyHandler := th.NewHistoryHandler(querySvc, logger)
	wsHandler := th.NewWebSocketHandler(rdb, cfg.JWTSecret, vehicleRepo, liveCache, logger)
	geofenceHandler := th.NewGeofenceHandler(geoRepo, logger)

	// --- Vehicle Domain ---
	vehicleService := vs.NewVehicleService(vehicleDomainRepo, logger)
	vehicleHandlers := vh.NewVehicleHandlers(vehicleService, logger)

	// --- Driver Domain ---
	driverRepo := drvRepo.NewDriverRepository(pool)
	driverOnboardingRepo := drvRepo.NewOnboardingRepository(pool)
	driverShiftRepo := drvRepo.NewShiftRepository(pool)
	driverTrainingRepo := drvRepo.NewTrainingRepository(pool)
	driverPayrollRepo := drvRepo.NewPayrollRepository(pool)
	driverRatingRepo := drvRepo.NewRatingRepository(pool)
	driverSafetyRepo := drvRepo.NewSafetyRepository(pool)
	driverMobileSyncRepo := drvRepo.NewMobileSyncRepository(pool)

	driverSvc := driverService.NewDriverService(driverRepo, logger)
	driverOnboardingSvc := driverService.NewOnboardingService(driverOnboardingRepo, logger)
	driverShiftSvc := driverService.NewShiftService(driverShiftRepo, logger)
	driverTrainingSvc := driverService.NewTrainingService(driverTrainingRepo, logger)
	driverPayrollSvc := driverService.NewPayrollService(driverPayrollRepo, logger)
	driverRatingSvc := driverService.NewRatingService(driverRatingRepo, logger)
	driverSafetySvc := driverService.NewSafetyEventService(driverSafetyRepo, logger)
	driverMobileSyncSvc := driverService.NewMobileSyncService(driverMobileSyncRepo, logger)
	driverSecurity := driverMiddleware.NewDriverSecurityMiddleware(logger)

	driverHandler := driverHandlers.NewDriverHandler(
		driverSvc,
		driverOnboardingSvc,
		driverShiftSvc,
		driverTrainingSvc,
		driverPayrollSvc,
		driverRatingSvc,
		driverSafetySvc,
		driverMobileSyncSvc,
		driverSecurity,
		logger,
	)

	// --- Organization Domain ---
	organizationRepo := orgRepo.NewOrganizationRepository(pool, logger)
	organizationService := orgSvc.NewOrganizationService(organizationRepo, logger)
	organizationHandler := orgHandlers.NewOrganizationHandlers(organizationService, logger)

	// --- Geofence Worker (conditional) ---
	var geofenceChecker *worker.GeofenceChecker
	if cfg.Flags.GeofenceAlerts {
		geofenceChecker = worker.NewGeofenceChecker(
			telemRepo, geoRepo, liveCache, rdb, logger, 10*time.Second)
		go geofenceChecker.Start()
		logger.Info("geofence alerts enabled, checker started")
	}

	// --- Router ---
	r := chi.NewRouter()

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8081", "https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Org-ID", "X-CSRF-Token", "X-Trace-ID", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Request logging
	r.Use(requestLogger(logger))

	// Health check (no auth)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"fleet-core"}`))
	})

	// --- API Routes ---

	// 1. IAM public routes (no auth)
	if iamModule != nil {
		r.Route("/api/v1", func(r chi.Router) {
			iamModule.RegisterPublicRoutes(r, logger)
		})
	}

	// Menu endpoints (static, no auth needed for sidebar)
	menuHandler := th.NewMenuHandler()
	r.Get("/api/v1/ui/menus/*", menuHandler.ServeHTTP)

	// UI config / strings stubs (frontend falls back to defaults)
	uiConfigHandler := th.NewUIConfigHandler()
	uiStringsHandler := th.NewUIStringsHandler()
	r.Get("/api/v1/ui/global_config", uiConfigHandler.ServeHTTP)
	r.Get("/api/v1/ui/strings/*", uiStringsHandler.ServeHTTP)
	r.Get("/api/v1/ui/misc/legal_docs", uiConfigHandler.GetLegalDocs)

	// Admin stub endpoints (columns — drivers moved to real CRUD)
	adminHandler := th.NewAdminHandler()
	r.Get("/api/v1/ui/settings/columns", adminHandler.ServeHTTP)

	// Dashboard stub endpoints
	dashboardHandler := th.NewDashboardHandler()
	r.Get("/api/v1/dashboard/overview", dashboardHandler.ServeHTTP)
	r.Get("/api/v1/dashboard/data", dashboardHandler.ServeHTTP)

	// Refresh endpoint (uses refresh_token from body, not JWT)
	if iamModule != nil {
		r.Post("/api/v1/auth/refresh", iamModule.AuthHandler.Refresh)
	}

	// 2. All authenticated API routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWKSAuthGate(cfg.JWTSecret, logger))
		r.Use(middleware.TenantMiddleware(logger))

		// Authenticated IAM routes (profile, logout, etc.)
		if iamModule != nil {
			iamModule.RegisterAuthRoutes(r, logger)
		}

		// Tracking endpoints
		if cfg.Flags.VehicleMapEnabled {
			r.Post("/api/v1/telemetry", ingestHandler.ServeHTTP)
			r.Get("/api/v1/vehicles/locations", locationsHandler.ServeHTTP)

			if cfg.Flags.HistoryTrail {
				r.Get("/api/v1/vehicles/{id}/telemetry/history", historyHandler.ServeHTTP)
			}

			if cfg.Flags.RealTimeUpdates {
				r.Get("/ws/telemetry", wsHandler.ServeHTTP)
			}

			r.Get("/api/v1/geofences", geofenceHandler.ServeHTTP)
			r.Post("/api/v1/geofences", geofenceHandler.ServeHTTP)
		}

		// Vehicle CRUD (always available)
		r.Route("/api/v1/admin/vehicles", func(r chi.Router) {
			vehicleHandlers.RegisterRoutes(r)
		})

		// Driver CRUD + extended endpoints
		r.Route("/api/v1/admin/drivers", func(r chi.Router) {
			driverHandler.RegisterRoutes(r)
		})

		// Organization CRUD
		r.Route("/api/v1/admin", func(r chi.Router) {
			organizationHandler.RegisterRoutes(r)
		})
	})

	// --- Start HTTP Server ---
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	if geofenceChecker != nil {
		geofenceChecker.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}

// pgTenantChecker implements repository.TenantChecker using the live DB.
type pgTenantChecker struct{ pool *pgxpool.Pool }

func (t *pgTenantChecker) VehicleBelongsToOrg(ctx context.Context, orgID, vehicleID uuid.UUID) (bool, error) {
	var exists bool
	err := t.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL)`,
		vehicleID, orgID).Scan(&exists)
	return exists, err
}

// requestLogger is a simple chi middleware that logs each request.
func requestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Debug("request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("elapsed", time.Since(start)),
			)
		})
	}
}
