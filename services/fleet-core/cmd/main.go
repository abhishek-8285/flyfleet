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
	tripHandlers "github.com/flyfleet/fleet-core/internal/domains/trip/handlers"
	tripRepo "github.com/flyfleet/fleet-core/internal/domains/trip/repository"
	tripService "github.com/flyfleet/fleet-core/internal/domains/trip/service"
	"github.com/flyfleet/fleet-core/internal/domains/billing"
	customerHandlers "github.com/flyfleet/fleet-core/internal/domains/customer/handlers"
	customerRepo "github.com/flyfleet/fleet-core/internal/domains/customer/repository"
	customerSvc "github.com/flyfleet/fleet-core/internal/domains/customer/service"
	invoicingHandlers "github.com/flyfleet/fleet-core/internal/domains/invoicing/handlers"
	invoicingRepo "github.com/flyfleet/fleet-core/internal/domains/invoicing/repository/postgres"
	invoicingSvc "github.com/flyfleet/fleet-core/internal/domains/invoicing/service"
)

func main() {
	// --- Bootstrap ---
	cfg := config.Load()

	// Initialize Logger based on config
	var zapCfg zap.Config
	if cfg.Environment == "development" {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	if level, err := zap.ParseAtomicLevel(cfg.LogLevel); err == nil {
		zapCfg.Level = level
	}

	logger, _ := zapCfg.Build()
	defer logger.Sync()

	logger.Info("starting fleet-core service",
		zap.String("port", cfg.Port),
		zap.String("environment", cfg.Environment),
		zap.Int("jwt_secret_len", len(cfg.JWTSecret)))

	ctx := context.Background()

	// --- Database ---
	pool, err := database.NewPool(ctx, cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("database connection failed", zap.Error(err))
	}
	defer pool.Close()

	// --- SYSTEM RESTORATION (Unified Schema Sync) ---
	syncAllSchemas(ctx, pool, logger)

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
	ingestionSvc.SetHMACSecret(cfg.TelemetryHMACSecret)
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
	
	// --- Trip Domain ---
	tRepo := tripRepo.NewTripRepository(pool)
	tService := tripService.NewTripService(tRepo, logger)
	tHandler := tripHandlers.NewTripHandler(tService, logger)

	// --- Custom Customer V2 Domain (Injected as needed) ---
	cV2Repo := customerRepo.NewCustomerV2Repository(pool)
	custV2Svc := customerSvc.NewCustomerService(cV2Repo)
	// Other services stubs (as required by NewCustomerV2Handler)
	bookV2Svc := customerSvc.NewBookingService(cV2Repo)
	shipV2Svc := customerSvc.NewShipmentService(cV2Repo, nil)
	priceV2Svc := customerSvc.NewPricingService(cV2Repo)
	creditV2Svc := customerSvc.NewCreditService(cV2Repo)
	slaV2Svc := customerSvc.NewSLAService(cV2Repo)

	custV2Handler := customerHandlers.NewCustomerV2Handler(
		custV2Svc, 
		bookV2Svc, 
		shipV2Svc, 
		priceV2Svc, 
		creditV2Svc, 
		slaV2Svc, 
		cV2Repo, 
		logger,
	)

	// --- New Invoicing Domain ---
	invRepo := invoicingRepo.NewInvoiceRepository(pool)
	invoicingService := invoicingSvc.NewInvoicingService(invRepo)
	invoicingHandler := invoicingHandlers.NewHandler(invoicingService)

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
		w.Write([]byte(`{"status":"healthy","service":"fleet-core"}`))
	})

	// PUBLIC Hardware Telemetry (No User Token Required)
	r.Post("/api/v1/telemetry", ingestHandler.ServeHTTP)

	// --- API Routes ---

	// --- API Routes ---

	// Combined API v1 Route (Public & Authenticated)
	r.Route("/api/v1", func(r chi.Router) {
		// 1. IAM public routes
		if iamModule != nil {
			iamModule.RegisterPublicRoutes(r, logger)
		}

		// Menu endpoints (static)
		menuHandler := th.NewMenuHandler()
		r.Get("/ui/menus/*", menuHandler.ServeHTTP)

		// UI config / strings stubs
		uiConfigHandler := th.NewUIConfigHandler()
		uiStringsHandler := th.NewUIStringsHandler()
		r.Get("/ui/global_config", uiConfigHandler.ServeHTTP)
		r.Get("/ui/strings/*", uiStringsHandler.ServeHTTP)
		r.Get("/ui/misc/legal_docs", uiConfigHandler.GetLegalDocs)

		// Admin stub endpoints
		adminHandler := th.NewAdminHandler()
		r.Get("/ui/settings/columns", adminHandler.ServeHTTP)
		r.Get("/ui/settings/gps-types", adminHandler.ServeHTTP)

		// Dashboard stub endpoints
		dashboardHandler := th.NewDashboardHandler()
		r.Get("/dashboard/overview", dashboardHandler.ServeHTTP)
		r.Get("/dashboard/data", dashboardHandler.ServeHTTP)

		// Refresh endpoint
		if iamModule != nil {
			r.Post("/auth/refresh", iamModule.AuthHandler.Refresh)
		}

		// Landing page data
		landingHandler := th.NewLandingHandler()
		r.Get("/data", landingHandler.ServeHTTP)
		
		// 1.5 Authenticated API routes
		r.Group(func(r chi.Router) {
			authGate := middleware.NewJWKSAuthGate("FlyFleet", "fleet-dashboard", cfg.JWTSecret, logger)
			r.Use(authGate.Handler)
			r.Use(middleware.TenantMiddleware(logger))

			// Authenticated IAM routes
			if iamModule != nil {
				iamModule.RegisterAuthRoutes(r, logger)
			}

			// Tracking endpoints
			if cfg.Flags.VehicleMapEnabled {
				r.Get("/vehicles/locations", locationsHandler.ServeHTTP)

				if cfg.Flags.HistoryTrail {
					r.Get("/vehicles/{id}/telemetry/history", historyHandler.ServeHTTP)
				}
				
				// WebSockets moved to main router if starting with /ws
				// Handled below if needed
			}

			// Geofences
			r.Get("/geofences", geofenceHandler.ServeHTTP)
			r.Post("/geofences", geofenceHandler.ServeHTTP)

			// Administrative Unified Router
			r.Route("/admin", func(r chi.Router) {
				// Vehicle CRUD
				r.Route("/vehicles", func(r chi.Router) {
					vehicleHandlers.RegisterRoutes(r)
				})

				// Driver CRUD
				r.Route("/drivers", func(r chi.Router) {
					driverHandler.RegisterRoutes(r)
				})

				// Trip CRUD
				r.Route("/trips", func(r chi.Router) {
					tHandler.RegisterRoutes(r)
				})

				// Customer Module
				r.Route("/customers", func(r chi.Router) {
					custV2Handler.RegisterRoutes(r)
				})

				// Invoicing Domain
				r.Route("/invoicing", func(r chi.Router) {
					invoicingHandler.RegisterRoutes(r)
				})

				// Organization & Dashboard Intelligence (Registers /organizations, /branches, etc.)
				organizationHandler.RegisterRoutes(r)
			})

			// Billing Module
			billing.RegisterModule(r, pool, billing.Config{
				Enabled:              cfg.Billing.Enabled,
				ChromeTimeoutSeconds: cfg.Billing.ChromeTimeoutSeconds,
			}, logger)
		})
	})

	// WebSockets (outside /api/v1)
	if cfg.Flags.RealTimeUpdates {
		r.Get("/ws/telemetry", wsHandler.ServeHTTP)
	}

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

func (t *pgTenantChecker) VehicleBelongsToOrg(ctx context.Context, tenantID, vehicleID uuid.UUID) (bool, error) {
	var exists bool
	err := t.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM vehicles WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL)`,
		vehicleID, tenantID).Scan(&exists)
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

func syncAllSchemas(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Starting Total System Schema Restoration...")
	
	ensureIAMSchema(ctx, pool, logger)
	ensureVehicleSchema(ctx, pool, logger)
	ensureTripSchema(ctx, pool, logger)
	ensureDriverSchema(ctx, pool, logger)
	ensureCustomerSchema(ctx, pool, logger)
	ensureInvoicingSchema(ctx, pool, logger)
	ensureBillingSchema(ctx, pool, logger)
	ensureHardwareBillingSchema(ctx, pool, logger)
	ensureAnalyticsSchema(ctx, pool, logger)
	ensureTrackingSchema(ctx, pool, logger)

	logger.Info("Total System Restoration Complete.")
}

func ensureIAMSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing IAM Schema...")
	statements := []string{
		"ALTER TABLE roles ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE roles ADD COLUMN IF NOT EXISTS permission_mask BIGINT DEFAULT 0",
		"ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT FALSE",
		"ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE organizations ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS must_reset_password BOOLEAN DEFAULT FALSE",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}

	// Backfill tenant_id from org_id for IAM tables
	backfills := []string{
		"UPDATE roles SET tenant_id = org_id WHERE tenant_id IS NULL AND org_id IS NOT NULL",
		"UPDATE user_roles SET tenant_id = org_id WHERE tenant_id IS NULL AND org_id IS NOT NULL",
		"UPDATE organizations SET tenant_id = id WHERE tenant_id IS NULL",
	}
	for _, bf := range backfills {
		pool.Exec(ctx, bf)
	}
}

func ensureVehicleSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Vehicle Schema...")
	statements := []string{
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS registration_number VARCHAR(50)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS vin_number VARCHAR(100)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS chassis_number VARCHAR(100)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS engine_number VARCHAR(100)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS make VARCHAR(100)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS model VARCHAR(100)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS year INTEGER",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS color VARCHAR(50)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active'",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_odometer FLOAT8 DEFAULT 0",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS last_service_date TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS last_service_odometer FLOAT8 DEFAULT 0",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS next_service_date TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS next_service_odometer FLOAT8 DEFAULT 0",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS insurance_expiry TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS permit_expiry TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fitness_expiry TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS pollution_expiry TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_enabled BOOLEAN DEFAULT FALSE",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS gps_device_id VARCHAR(100)",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS last_gps_update TIMESTAMP",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_latitude FLOAT8",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_longitude FLOAT8",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS current_speed FLOAT8 DEFAULT 0",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS fuel_level FLOAT8 DEFAULT 0",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1",
		"ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
	// Safely backfill tenant_id from org_id
	pool.Exec(ctx, "DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'vehicles' AND column_name = 'org_id') THEN UPDATE vehicles SET tenant_id = org_id WHERE tenant_id IS NULL; END IF; END $$;")
}

func ensureTripSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Trip Schema...")
	statements := []string{
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS planned_distance FLOAT8 DEFAULT 0",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS total_distance FLOAT8 DEFAULT 0",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS actual_duration_min INTEGER DEFAULT 0",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS fuel_used FLOAT8 DEFAULT 0",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS idling_time_minutes INTEGER DEFAULT 0",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS eway_bill_no VARCHAR(100)",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS gst_classification VARCHAR(100)",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS customer_name VARCHAR(200)",
		"ALTER TABLE trips ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
	pool.Exec(ctx, "DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'trips' AND column_name = 'org_id') THEN UPDATE trips SET tenant_id = org_id WHERE tenant_id IS NULL; END IF; END $$;")
}

func ensureDriverSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Driver Schema...")
	statements := []string{
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS license_number VARCHAR(100)",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active'",
		"ALTER TABLE drivers ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
	pool.Exec(ctx, "DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'drivers' AND column_name = 'org_id') THEN UPDATE drivers SET tenant_id = org_id WHERE tenant_id IS NULL; END IF; END $$;")
}

func ensureCustomerSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Customer Schema...")
	statements := []string{
		"ALTER TABLE customers ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE customers ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'ACTIVE'",
		"ALTER TABLE customers ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
	pool.Exec(ctx, "DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'org_id') THEN UPDATE customers SET tenant_id = org_id WHERE tenant_id IS NULL; END IF; END $$;")
}

func ensureInvoicingSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Invoicing Schema...")
	statements := []string{
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS template_type TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS trip_id UUID",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS vehicle_no TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS driver_name TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS driver_mobile TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS pickup_location TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS drop_location TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS pickup_time TIMESTAMP WITH TIME ZONE",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS drop_time TIMESTAMP WITH TIME ZONE",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS map_url TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS billed_distance DECIMAL",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS place_of_supply TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS reverse_charge BOOLEAN DEFAULT FALSE",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS gstin_seller TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS pan_seller TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS eway_bill_no TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS eway_bill_date TIMESTAMP WITH TIME ZONE",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS irn TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS ack_no TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS ack_date TIMESTAMP WITH TIME ZONE",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS bank_name TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS account_no TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS ifsc_code TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS branch_name TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS beneficiary_name TEXT",
		"ALTER TABLE invoices ADD COLUMN IF NOT EXISTS currency VARCHAR(10) DEFAULT 'INR'",
		"ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
		"ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS hsn_sac TEXT",
		"ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS unit TEXT",
		"ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS list_price DECIMAL",
		"ALTER TABLE invoice_items ADD COLUMN IF NOT EXISTS discount_amount DECIMAL",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
	pool.Exec(ctx, "DO $$ BEGIN IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'invoices' AND column_name = 'org_id') THEN UPDATE invoices SET tenant_id = org_id WHERE tenant_id IS NULL; END IF; END $$;")
}

func ensureBillingSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Billing Schema...")
	statements := []string{
		"ALTER TABLE billing_settings ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE billing_settings ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
		"ALTER TABLE payment_records ADD COLUMN IF NOT EXISTS tenant_id UUID",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
}

func ensureHardwareBillingSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Hardware Billing Schema...")
	
	// Create hardware products table
	pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS billing_hardware_products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			model_code VARCHAR(100) UNIQUE NOT NULL,
			category VARCHAR(50) NOT NULL,
			one_time_price DECIMAL(15, 2) NOT NULL,
			lease_monthly_price DECIMAL(15, 2),
			currency VARCHAR(3) NOT NULL DEFAULT 'USD',
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)

	// Create device assignments table
	pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS billing_device_assignments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			device_id VARCHAR(100) NOT NULL,
			product_id UUID NOT NULL REFERENCES billing_hardware_products(id),
			billing_type VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'PENDING_ACTIVATION',
			activated_at TIMESTAMPTZ,
			invoice_id UUID REFERENCES invoices(id), 
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(device_id)
		)
	`)

	// Seed data if empty
	var count int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM billing_hardware_products").Scan(&count)
	if count == 0 {
		pool.Exec(ctx, `
			INSERT INTO billing_hardware_products (name, model_code, category, one_time_price, lease_monthly_price)
			VALUES 
			('FlyFleet Vision Pro Dashcam', 'FF-CAM-V2', 'DASHCAM', 399.00, 25.00),
			('FlyFleet Standard GPS Tracker', 'FF-GPS-100', 'TRACKER', 99.00, 10.00),
			('FlyFleet OBD-II Telematics Node', 'FF-OBD-V1', 'TRACKER', 149.00, 15.00)
		`)
	}

	// Ensure indexes
	pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_device_assignments_tenant_id ON billing_device_assignments(tenant_id)")
	pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_device_assignments_device_id ON billing_device_assignments(device_id)")

	// BACKWARD COMPATIBILITY: Auto-populate assignments for existing vehicles
	// This ensures that existing vehicles with GPS devices don't throw "not in inventory" errors.
	pool.Exec(ctx, `
		INSERT INTO billing_device_assignments (tenant_id, device_id, product_id, billing_type, status)
		SELECT 
			v.tenant_id, 
			v.gps_device_id, 
			(SELECT id FROM billing_hardware_products WHERE model_code = 'FF-GPS-100' LIMIT 1),
			'ONE_TIME',
			'ACTIVE'
		FROM vehicles v
		WHERE v.gps_device_id IS NOT NULL 
		  AND v.gps_device_id != ''
		  AND v.deleted_at IS NULL
		ON CONFLICT (device_id) DO NOTHING
	`)

	// Ensure they also exist in tracking registry
	pool.Exec(ctx, `
		INSERT INTO device_registry (device_id, status, created_at, updated_at)
		SELECT 
			gps_device_id, 
			'active', 
			NOW(), 
			NOW()
		FROM vehicles
		WHERE gps_device_id IS NOT NULL 
		  AND gps_device_id != ''
		  AND deleted_at IS NULL
		ON CONFLICT (device_id) DO NOTHING
	`)
}


func ensureAnalyticsSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Analytics Schema...")
	statements := []string{
		"ALTER TABLE predictive_models ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE prediction_results ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE analytics_jobs ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE fleet_analytics ADD COLUMN IF NOT EXISTS tenant_id UUID",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
}

func ensureTrackingSchema(ctx context.Context, pool *pgxpool.Pool, logger *zap.Logger) {
	logger.Info("Syncing Tracking Schema...")
	
	// Ensure device_registry exists for tracking control plane
	pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS device_registry (
			device_id         VARCHAR(100) PRIMARY KEY,
			vehicle_id        UUID,
			firmware_version  VARCHAR(50),
			sim_iccid         VARCHAR(30),
			carrier           VARCHAR(50),
			last_heartbeat    TIMESTAMPTZ,
			status            VARCHAR(20) NOT NULL DEFAULT 'active',
			certificate_hash  VARCHAR(64),
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)

	// Ensure telemetry_points exists for tracking data plane
	pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS telemetry_points (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			device_id    VARCHAR(100) NOT NULL,
			lat          DOUBLE PRECISION NOT NULL,
			lng          DOUBLE PRECISION NOT NULL,
			speed        DOUBLE PRECISION DEFAULT 0,
			recorded_at  TIMESTAMP WITH TIME ZONE NOT NULL,
			event_id     UUID NOT NULL,
			tenant_id    UUID NOT NULL
		)
	`)

	// Ensure extended columns exist
	pool.Exec(ctx, "ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS received_at TIMESTAMP WITH TIME ZONE")
	pool.Exec(ctx, "ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS vehicle_id UUID")
	pool.Exec(ctx, "ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS heading DOUBLE PRECISION")
	pool.Exec(ctx, "ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS ignition BOOLEAN")
	pool.Exec(ctx, "ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS fuel_level DOUBLE PRECISION")
	pool.Exec(ctx, "ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS raw_data JSONB")

	statements := []string{
		"ALTER TABLE geofences ADD COLUMN IF NOT EXISTS tenant_id UUID",
		"ALTER TABLE geofences ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE",
		"ALTER TABLE telemetry_points ADD COLUMN IF NOT EXISTS tenant_id UUID",
	}
	for _, stmt := range statements {
		pool.Exec(ctx, stmt)
	}
}
