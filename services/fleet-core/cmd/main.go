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
	"github.com/flyfleet/fleet-core/internal/middleware"
	"github.com/flyfleet/fleet-core/internal/tracking/handlers"
	"github.com/flyfleet/fleet-core/internal/tracking/repository"
	"github.com/flyfleet/fleet-core/internal/tracking/service"
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
	telemRepo := repository.NewTelemetryRepository(pool, logger)
	liveCache := repository.NewLiveCacheRepository(rdb)
	geoRepo := repository.NewGeofenceRepository(pool)
	vehicleRepo := repository.NewVehicleRepository(pool)

	// --- Services ---
	tenantChecker := &pgTenantChecker{pool: pool}
	ingestionSvc := service.NewIngestionService(telemRepo, liveCache, rdb, logger)
	querySvc := service.NewTelemetryQueryService(telemRepo, liveCache, tenantChecker, logger)
	batchSvc := service.NewBatchLocationService(vehicleRepo, liveCache, telemRepo, logger, cfg.Flags.MaxVehiclesOnMap)

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
	ingestHandler := handlers.NewIngestHandler(ingestionSvc, logger)
	locationsHandler := handlers.NewLocationsHandler(batchSvc, logger)
	historyHandler := handlers.NewHistoryHandler(querySvc, logger)
	wsHandler := handlers.NewWebSocketHandler(rdb, cfg.JWTSecret, vehicleRepo, liveCache, logger)
	geofenceHandler := handlers.NewGeofenceHandler(geoRepo, logger)

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
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Org-ID"},
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
	r.Group(func(r chi.Router) {
		// All API routes require authentication
		r.Use(middleware.JWKSAuthGate(cfg.JWTSecret, logger))
		r.Use(middleware.TenantMiddleware(logger))

		api := chi.NewRouter()

		// IAM routes (public + authenticated) — registered BEFORE the auth gate
		// because login/register/etc. must be public
		if iamModule != nil {
			// Public IAM routes (no auth required)
			r.Route("/api/v1", func(r chi.Router) {
				iamModule.RegisterPublicRoutes(r, logger)
			})

			// Authenticated IAM routes (JWT required — already gated by outer JWKSAuthGate)
			r.Route("/api/v1", func(r chi.Router) {
				iamModule.RegisterAuthRoutes(r, logger)
			})
		}

		// Tracking endpoints (always enabled if feature flag is on)
		if cfg.Flags.VehicleMapEnabled {
			// POST /api/v1/telemetry — GPS position ingest
			api.Post("/v1/telemetry", ingestHandler.ServeHTTP)

			// GET /api/v1/vehicles/locations — bulk org vehicle positions
			api.Get("/v1/vehicles/locations", locationsHandler.ServeHTTP)

			// GET /api/v1/vehicles/{id}/telemetry/history — 24h history
			if cfg.Flags.HistoryTrail {
				api.Get("/v1/vehicles/{id}/telemetry/history", historyHandler.ServeHTTP)
			}

			// WebSocket /ws/telemetry — real-time position push
			if cfg.Flags.RealTimeUpdates {
				api.Get("/ws/telemetry", wsHandler.ServeHTTP)
			}

			// GET/POST /api/v1/geofences — geofence management
			api.Get("/v1/geofences", geofenceHandler.ServeHTTP)
			api.Post("/v1/geofences", geofenceHandler.ServeHTTP)
		}

		r.Mount("/api", api)
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
