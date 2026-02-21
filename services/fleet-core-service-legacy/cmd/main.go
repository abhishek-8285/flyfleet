package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/config"
	"github.com/flyfleet/fleet-core-service/internal/database"

	// Domains
	"github.com/flyfleet/fleet-core-service/internal/domains/integrations"
	"github.com/flyfleet/fleet-core-service/internal/domains/reports"
	vehHandlers "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/handlers"
	vehRepo "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	vehService "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"

	// Finance Domain
	financeHandlers "github.com/flyfleet/fleet-core-service/internal/domains/finance/handlers"
	financeRepo "github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	financeService "github.com/flyfleet/fleet-core-service/internal/domains/finance/service"

	// Emergency Domain
	emergencyHandlers "github.com/flyfleet/fleet-core-service/internal/domains/emergency/handlers"
	emergencyRepo "github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	emergencyService "github.com/flyfleet/fleet-core-service/internal/domains/emergency/service"

	// Customer Domain
	customerHandlers "github.com/flyfleet/fleet-core-service/internal/domains/customer/handlers"
	customerRepo "github.com/flyfleet/fleet-core-service/internal/domains/customer/repository"
	customerService "github.com/flyfleet/fleet-core-service/internal/domains/customer/service"

	// Driver Domain
	driverHandlers "github.com/flyfleet/fleet-core-service/internal/domains/driver/handlers"
	driverMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/driver/middleware"
	driverRepo "github.com/flyfleet/fleet-core-service/internal/domains/driver/repository"
	driverService "github.com/flyfleet/fleet-core-service/internal/domains/driver/service"

	tripHandlers "github.com/flyfleet/fleet-core-service/internal/domains/trip/handlers"
	tripRepo "github.com/flyfleet/fleet-core-service/internal/domains/trip/repository"
	tripService "github.com/flyfleet/fleet-core-service/internal/domains/trip/service"

	maintenHandlers "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/handlers"
	maintenRepo "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
	maintenService "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/service"

	geoHandlers "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/handlers"
	geoRepo "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/repository"
	geoService "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/service"

	fuelHandlers "github.com/flyfleet/fleet-core-service/internal/domains/fuel/handlers"
	fuelRepo "github.com/flyfleet/fleet-core-service/internal/domains/fuel/repository"
	fuelService "github.com/flyfleet/fleet-core-service/internal/domains/fuel/service"

	routeHandlers "github.com/flyfleet/fleet-core-service/internal/domains/route/handlers"
	routeRepo "github.com/flyfleet/fleet-core-service/internal/domains/route/repository"
	routeService "github.com/flyfleet/fleet-core-service/internal/domains/route/service"

	diagHandlers "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/handlers"
	diagRepo "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/repository"
	diagService "github.com/flyfleet/fleet-core-service/internal/domains/diagnostics/service"

	geoEngine "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/engine"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	// Postgres driver for database/sql (needed for diagnostics domain)
	_ "github.com/jackc/pgx/v5/stdlib"

	// Video Domain
	videoHandlers "github.com/flyfleet/fleet-core-service/internal/domains/video/handlers"
	videoRepo "github.com/flyfleet/fleet-core-service/internal/domains/video/repository"
	videoService "github.com/flyfleet/fleet-core-service/internal/domains/video/service"

	// Analytics Domain
	analyticsHandlers "github.com/flyfleet/fleet-core-service/internal/domains/analytics/handlers"
	analyticsRepo "github.com/flyfleet/fleet-core-service/internal/domains/analytics/repository"
	analyticsService "github.com/flyfleet/fleet-core-service/internal/domains/analytics/service"
)

func main() {
	// 1. Initialize Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 2. Load Config
	cfg := config.Load()

	// 3. Initialize Database (Main pgxpool)
	db, err := database.NewPostgresPool(context.Background(), cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize sql.DB for domains that still use database/sql (e.g., diagnostics)
	dbSQL, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to initialize sql.DB", zap.Error(err))
	}
	defer dbSQL.Close()

	// 4. Initialize NATS (for Geofencing, etc.)
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Warn("Failed to connect to NATS, geofencing features might be limited", zap.Error(err))
	} else {
		defer nc.Close()
	}

	// 5. Initialize Router
	r := chi.NewRouter()

	// Standard Middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))

	// Pre-initialize Repositories (needed by multiple routes and workers)
	vRepo := vehRepo.NewVehicleRepository(db, logger)
	mRepo := maintenRepo.NewMaintenanceRepository(db)
	tRepo := tripRepo.NewTripRepository(db)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"service": "fleet-core-service",
		})
	})

	// 6. Initialize Domains & Register Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Vehicle Domain
		vService := vehService.NewVehicleService(vRepo, logger)
		vHandlers := vehHandlers.NewVehicleHandlers(vService, logger)
		r.Route("/vehicles", func(r chi.Router) {
			vHandlers.RegisterRoutes(r)
		})

		// Driver Domain
		dRepo := driverRepo.NewDriverRepository(db)
		dSecurity := driverMiddleware.NewDriverSecurityMiddleware(logger)
		dService := driverService.NewDriverService(dRepo, logger)
		dHandlers := driverHandlers.NewDriverHandler(dService, dSecurity, logger)
		r.Route("/drivers", func(r chi.Router) {
			dHandlers.RegisterRoutes(r)
		})


		// Customer Domain
		cRepo := customerRepo.NewCustomerRepository(db)
		cService := customerService.NewCustomerService(cRepo, logger, cfg.JWTSecret)
		cHandlers := customerHandlers.NewCustomerHandlers(cService, logger, cfg.JWTSecret)
		r.Route("/customer", func(r chi.Router) {
			cHandlers.RegisterRoutes(r)
		})

		// Trip Domain
		tService := tripService.NewTripService(tRepo, logger)
		tHandlers := tripHandlers.NewTripHandler(tService, logger)
		r.Route("/trips", func(r chi.Router) {
			tHandlers.RegisterRoutes(r)
		})

		// Maintenance Domain
		mService := maintenService.NewMaintenanceService(mRepo, logger)
		mHandlers := maintenHandlers.NewMaintenanceHandler(mService, logger)
		r.Route("/maintenance", func(r chi.Router) {
			mHandlers.RegisterRoutes(r)
		})

		// Work Order Domain
		woRepo := maintenRepo.NewWorkOrderRepository(db)
		woPartsRepo := maintenRepo.NewWorkOrderPartsRepository(db)
		woInvRepo := maintenRepo.NewInventoryRepository(db)
		woSettingsRepo := maintenRepo.NewSettingsRepository(db)
		woVendorRepo := maintenRepo.NewVendorRepository(db)
		woInspectionRepo := maintenRepo.NewInspectionRepository(db)
		woService := maintenService.NewWorkOrderService(woRepo, woPartsRepo, woInvRepo, woSettingsRepo, woVendorRepo, woInspectionRepo, logger)
		woHandlers := maintenHandlers.NewWorkOrderHandler(woService, logger)
		r.Route("/work-orders", func(r chi.Router) {
			woHandlers.RegisterRoutes(r)
		})

		// Vendors Domain
		vHandler := maintenHandlers.NewVendorHandler(woVendorRepo, logger)
		r.Route("/vendors", func(r chi.Router) {
			vHandler.RegisterRoutes(r)
		})

		// Inspections Domain
		iHandler := maintenHandlers.NewInspectionHandler(woInspectionRepo, logger)
		r.Route("/inspections", func(r chi.Router) {
			iHandler.RegisterRoutes(r)
		})

		// Inventory Domain
		invRepo := maintenRepo.NewInventoryRepository(db)
		invHandlers := maintenHandlers.NewInventoryHandler(invRepo, logger)
		r.Route("/inventory", func(r chi.Router) {
			invHandlers.RegisterRoutes(r)
		})

		// Geofencing Domain
		gRepo := geoRepo.NewGeofenceRepository(db)
		gService := geoService.NewGeofenceService(gRepo, nc, logger)
		gHandlers := geoHandlers.NewGeofenceHandler(gService, logger)
		r.Route("/geofencing", func(r chi.Router) {
			gHandlers.RegisterRoutes(r)
		})

		// Fuel Domain
		fLogRepo := fuelRepo.NewFuelLogRepository(db)
		fSettingsRepo := fuelRepo.NewSettingsRepository(db)
		fService := fuelService.NewFuelService(fLogRepo, fSettingsRepo, logger)
		fHandlers := fuelHandlers.NewFuelHandler(fService, logger)
		r.Route("/fuel", func(r chi.Router) {
			fHandlers.RegisterRoutes(r)
		})

		// Route Optimization Domain
		roRepo := routeRepo.NewRouteOptimizationRepository(db)
		roService := routeService.NewRouteOptimizationService(roRepo)
		roHandlers := routeHandlers.NewRouteOptimizationHandler(roService, logger)
		r.Route("/routes", func(r chi.Router) {
			roHandlers.RegisterRoutes(r)
		})

		// Diagnostics Domain
		diagR := diagRepo.NewRepository(dbSQL, logger)
		diagS := diagService.NewDeviceDiagnosticsService(diagR, logger)
		diagH := diagHandlers.NewHandlers(diagS, logger)
		r.Route("/diagnostics", func(r chi.Router) {
			diagH.RegisterRoutes(r)
		})

		// Reports Domain
		reportsRepo := reports.NewRepository(db, logger)
		reportsService := reports.NewService(reportsRepo, logger)
		reportsHandler := reports.NewHandler(reportsService, logger)
		r.Route("/reports", func(r chi.Router) {
			reportsHandler.RegisterRoutes(r)
		})

		// Integrations Domain
		integrationsRepo := integrations.NewRepository(db, logger)
		integrationsService := integrations.NewService(integrationsRepo, logger)
		integrationsHandler := integrations.NewHandler(integrationsService, logger)
		r.Route("/integrations", func(r chi.Router) {
			integrationsHandler.RegisterRoutes(r)
		})

		// Analytics Domain
		aRepo := analyticsRepo.NewAnalyticsRepository(db)
		aService := analyticsService.NewAnalyticsService(aRepo)
		aHandlers := analyticsHandlers.NewAnalyticsHandler(aService, logger)
		aHandlers.RegisterRoutes(r)

		// Finance Domain
		finRepo := financeRepo.NewFinanceRepository(db)
		finService := financeService.NewFinanceService(finRepo, logger)
		finHandlers := financeHandlers.NewFinanceHandlers(finService, logger)
		r.Route("/finance", func(r chi.Router) {
			finHandlers.RegisterRoutes(r)
		})

		// Emergency Domain
		emRepo := emergencyRepo.NewEmergencyRepository(db)
		emService := emergencyService.NewEmergencyService(emRepo, nc, logger)
		emHandlers := emergencyHandlers.NewEmergencyHandlers(emService, logger)
		r.Route("/emergency", func(r chi.Router) {
			emHandlers.RegisterRoutes(r)
		})
	})

	// Video Domain
	vidRepo := videoRepo.NewVideoRepository(db)
	vidService := videoService.NewVideoService(vidRepo)
	vidHandlers := videoHandlers.NewVideoHandler(vidService)
	r.Route("/api/v1/video", func(r chi.Router) {
		vidHandlers.RegisterRoutes(r)
	})

	// 7. Initialize and Start Background Workers
	if nc != nil {
		ge := geoEngine.NewGeofenceEngine(db, nc, logger)
		go func() {
			if err := ge.Start(context.Background()); err != nil {
				logger.Error("Geofence engine failed", zap.Error(err))
			}
		}()

		tw := maintenService.NewTelemetryWorker(mRepo, nc, logger)
		go func() {
			if err := tw.Start(context.Background()); err != nil {
				logger.Error("Telemetry worker failed", zap.Error(err))
			}
		}()
	}

	// 8. Start Server
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

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down fleet-core-service...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	logger.Info("Fleet-core-service exited")
}
