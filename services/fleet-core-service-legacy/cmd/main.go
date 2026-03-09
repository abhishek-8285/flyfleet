package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	shareddb "github.com/fleet-platform/shared/pkg/database"
	sharedmessaging "github.com/fleet-platform/shared/pkg/messaging"
	"github.com/flyfleet/fleet-core-service/internal/config"
	"github.com/flyfleet/fleet-core-service/internal/control"
	"github.com/flyfleet/fleet-core-service/internal/database"
	"github.com/flyfleet/fleet-core-service/internal/docs"
	"github.com/flyfleet/fleet-core-service/internal/gateway"
	// "github.com/flyfleet/fleet-core-service/internal/logadapter"
	coremiddleware "github.com/flyfleet/fleet-core-service/internal/middleware"
	// dashHandlers "github.com/flyfleet/fleet-core-service/internal/handlers"
	"github.com/fleet-platform/shared/pkg/cache"
	// monitorSvc "github.com/flyfleet/fleet-core-service/internal/domains/monitoring/service"

	// Domains
	governanceHandlers "github.com/flyfleet/fleet-core-service/internal/domains/governance/handlers"
	governanceRepo     "github.com/flyfleet/fleet-core-service/internal/domains/governance/repository"
	governanceService  "github.com/flyfleet/fleet-core-service/internal/domains/governance/service"
	govMiddleware      "github.com/flyfleet/fleet-core-service/internal/domains/governance/middleware"
	govUtils           "github.com/flyfleet/fleet-core-service/internal/domains/governance/utils"
	govWorkers         "github.com/flyfleet/fleet-core-service/internal/domains/governance/workers"
	indiav2handlers    "github.com/flyfleet/fleet-core-service/internal/domains/india_v2/handlers"
	indiav2repo        "github.com/flyfleet/fleet-core-service/internal/domains/india_v2/repository"
	indiav2service     "github.com/flyfleet/fleet-core-service/internal/domains/india_v2/service"
	integrationsv2handlers "github.com/flyfleet/fleet-core-service/internal/domains/integrations_v2/handlers"
	integrationsv2repo     "github.com/flyfleet/fleet-core-service/internal/domains/integrations_v2/repository"
	integrationsv2service  "github.com/flyfleet/fleet-core-service/internal/domains/integrations_v2/service"
	reportsv2handlers  "github.com/flyfleet/fleet-core-service/internal/domains/reports_v2/handlers"
	reportsv2queue     "github.com/flyfleet/fleet-core-service/internal/domains/reports_v2/queue"
	reportsv2repo      "github.com/flyfleet/fleet-core-service/internal/domains/reports_v2/repository"
	reportsv2service   "github.com/flyfleet/fleet-core-service/internal/domains/reports_v2/service"
	settingsHandlers   "github.com/flyfleet/fleet-core-service/internal/domains/settings/handlers"
	settingsMiddleware "github.com/flyfleet/fleet-core-service/internal/domains/settings/middleware"
	settingsRepo       "github.com/flyfleet/fleet-core-service/internal/domains/settings/repository"
	settingsService    "github.com/flyfleet/fleet-core-service/internal/domains/settings/service"

	vehHandlers "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/handlers"
	vehProcessor "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/processor"
	vehRepo "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	vehService "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"

	// Finance Domain
	financeHandlers "github.com/flyfleet/fleet-core-service/internal/domains/finance/handlers"
	financeRepo "github.com/flyfleet/fleet-core-service/internal/domains/finance/repository"
	financeService "github.com/flyfleet/fleet-core-service/internal/domains/finance/service"

	geoHandlers "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/handlers"
	geoRepo     "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/repository"
	geoService  "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/service"
	geoEngine   "github.com/flyfleet/fleet-core-service/internal/domains/geofencing/engine"

	emergencyHandlers "github.com/flyfleet/fleet-core-service/internal/domains/emergency/handlers"
	emergencyRepo "github.com/flyfleet/fleet-core-service/internal/domains/emergency/repository"
	emergencyService "github.com/flyfleet/fleet-core-service/internal/domains/emergency/service"
	trackingRepo "github.com/flyfleet/fleet-core-service/internal/domains/tracking_v2/repository"

	maintenHandlers "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/handlers"
	maintenRepo     "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/repository"
	maintenService  "github.com/flyfleet/fleet-core-service/internal/domains/maintenance/service"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"

	// Video Domain
	/*
	videoHandlers "github.com/flyfleet/fleet-core-service/internal/domains/video/handlers"
	videoRepo "github.com/flyfleet/fleet-core-service/internal/domains/video/repository"
	videoService "github.com/flyfleet/fleet-core-service/internal/domains/video/service"
	*/

	// Analytics Domain
	/*
	analyticsHandlers "github.com/flyfleet/fleet-core-service/internal/domains/analytics/handlers"
	analyticsRepo "github.com/flyfleet/fleet-core-service/internal/domains/analytics/repository"
	analyticsService "github.com/flyfleet/fleet-core-service/internal/domains/analytics/service"
	*/
)

func main() {
	// 1. Initialize Signal-aware Context for Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 2. Initialize Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 3. Load Config
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

	// Shared abstractions for migrated tracking v2 handlers.
	sharedDB, err := shareddb.NewPostgres(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Warn("Failed to initialize shared database adapter", zap.Error(err))
	}
	if sharedDB != nil {
		defer sharedDB.Close()
	}
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

	// 5. Initialize Redis Cache
	redisCache, err := cache.NewRedisCache(cfg.RedisURL, "", 0, logger)
	var sharedRedisClient *redis.Client
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
	// pulseService := monitorSvc.NewPulseService(sharedDB, redisCache, nc, logger)

	// 5. Initialize Router
	r := chi.NewRouter()

	// Standard Middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(coremiddleware.TraceIDMiddleware)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ENFORCE_BFF_AUTH_GATE")), "true") {
		r.Use(coremiddleware.NewBFFAuthGate(cfg.JWTSecret, []string{
			"/health",
			"/swagger",
			"/api/control/",
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/validate-token",
			"/api/v1/ws",
			"/api/v1/dashboard/stats",
			"/api/v1/admin/dashboard/data",
		}))
	}

	// Pre-initialize Repositories (needed by multiple routes and workers)
	vRepo := vehRepo.NewVehicleRepository(db, logger)
	vBulkRepo := vehRepo.NewBulkRepository(db)
	mRepo := maintenRepo.NewMaintenanceRepository(db)
	// tRepo := tripRepo.NewTripRepository(db)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
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
		vBulkProcessor := vehProcessor.NewBulkProcessor(vBulkRepo, vRepo, logger)
		vBulkService := vehService.NewBulkService(vBulkRepo, vBulkProcessor, logger)
		vBulkHandlers := vehHandlers.NewBulkHandlers(vBulkService, logger)
		r.Route("/bulk-jobs", func(r chi.Router) {
			vBulkHandlers.RegisterRoutes(r)
		})
		// vTripRepo := vehRepo.NewTripRepository(db)
		// vTripService := vehService.NewTripService(vTripRepo, logger)
		// vTripHandlers := vehHandlers.NewTripHandlers(vTripService, logger)



		// Maintenance Domain
		mService := maintenService.NewMaintenanceService(mRepo, logger)
		mHandlers := maintenHandlers.NewMaintenanceHandler(mService, logger)
		r.Route("/maintenance", func(r chi.Router) {
			mHandlers.RegisterRoutes(r)
		})
		// Compatibility alias for maintenance-service route shape.
		r.Get("/vehicles/{vehicle_id}/maintenance-history", mHandlers.GetVehicleMaintenanceHistory)

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

/*
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
		platformCompatHandlers := platformcompathandlers.NewHandlers()
		platformCompatHandlers.RegisterRoutes(r)

		// Diagnostics Domain
		diagR := diagRepo.NewRepository(dbSQL, logger)
		diagS := diagService.NewDeviceDiagnosticsService(diagR, logger)
		diagH := diagHandlers.NewHandlers(diagS, logger)
		r.Route("/diagnostics", func(r chi.Router) {
			diagH.RegisterRoutes(r)
		})
*/

		// Reports Domain (migrated v2 in monolith)
		reportsRepo := reportsv2repo.NewPostgresRepository(db)
		reportsQueue := reportsv2queue.NewQueueClient(cfg.RedisURL)
		reportsService := reportsv2service.NewReportService(reportsRepo, reportsQueue, logger)
		reportsHandler := reportsv2handlers.NewReportHandler(reportsService, logger)
		r.Route("/reports", func(r chi.Router) {
			r.Use(reportsv2handlers.ContextMiddleware)
			reportsHandler.RegisterRoutes(r)
		})

		// Integrations Domain (migrated v2 in monolith)
		intRepo := integrationsv2repo.NewRepository(dbSQL, logger)
		intService := integrationsv2service.NewIntegrationsService(intRepo, cfg.JWTSecret, logger)
		intHandler := integrationsv2handlers.NewHandlers(intService)
		r.Route("/", func(r chi.Router) {
			intHandler.RegisterRoutes(r)
		})

		// India Integration Domain (migrated in monolith)
		indiaRepo := indiav2repo.NewIndiaIntegrationRepository(db)
		indiaService := indiav2service.NewIndiaIntegrationService(indiaRepo)
		indiaHandlers := indiav2handlers.NewIndiaIntegrationHandlers(indiaService)
		r.Route("/india", func(r chi.Router) {
			r.Route("/fastag", func(r chi.Router) {
				r.Post("/transactions", indiaHandlers.CreateFASTagTransaction)
				r.Post("/sync", indiaHandlers.SyncFASTag)
				r.Get("/transactions/{id}", indiaHandlers.GetFASTagTransaction)
				r.Get("/transactions", indiaHandlers.ListFASTagTransactions)
			})
			r.Route("/ewaybill", func(r chi.Router) {
				r.Post("/bills", indiaHandlers.CreateEWayBill)
				r.Post("/validate", indiaHandlers.ValidateEWayBill)
				r.Get("/bills/{id}", indiaHandlers.GetEWayBill)
				r.Get("/bills/number/{number}", indiaHandlers.GetEWayBillByNumber)
				r.Get("/bills", indiaHandlers.ListEWayBills)
			})
			r.Route("/vahan", func(r chi.Router) {
				r.Post("/verifications", indiaHandlers.CreateVehicleVerification)
				r.Post("/verify", indiaHandlers.VerifyVehicle)
				r.Get("/verifications/{id}", indiaHandlers.GetVehicleVerification)
				r.Get("/verifications/reg/{reg_number}", indiaHandlers.GetVehicleVerificationByRegNumber)
				r.Get("/verifications", indiaHandlers.ListVehicleVerifications)
			})
			r.Route("/sarathi", func(r chi.Router) {
				r.Post("/verifications", indiaHandlers.CreateDriverVerification)
				r.Post("/verify", indiaHandlers.VerifyDriver)
				r.Get("/verifications/{id}", indiaHandlers.GetDriverVerification)
				r.Get("/verifications/license/{license_number}", indiaHandlers.GetDriverVerificationByLicenseNumber)
				r.Get("/verifications", indiaHandlers.ListDriverVerifications)
			})
			r.Route("/permits", func(r chi.Router) {
				r.Post("/", indiaHandlers.CreateLocalPermit)
				r.Get("/{id}", indiaHandlers.GetLocalPermit)
				r.Get("/number/{permit_number}", indiaHandlers.GetLocalPermitByNumber)
				r.Get("/", indiaHandlers.ListLocalPermits)
			})
			r.Route("/gst", func(r chi.Router) {
				r.Post("/compliances", indiaHandlers.CreateGSTCompliance)
				r.Get("/compliances/{id}", indiaHandlers.GetGSTCompliance)
				r.Get("/compliances/gstin/{gstin}", indiaHandlers.GetGSTComplianceByGSTIN)
				r.Get("/compliances", indiaHandlers.ListGSTCompliances)
			})
			r.Route("/config", func(r chi.Router) {
				r.Post("/", indiaHandlers.CreateIntegrationConfig)
				r.Get("/{id}", indiaHandlers.GetIntegrationConfig)
				r.Get("/org/{org_id}/type/{service_type}", indiaHandlers.GetIntegrationConfigByServiceType)
				r.Get("/", indiaHandlers.ListIntegrationConfigs)
			})
		})

		// Settings Domain (migrated in monolith)
		setRepo := settingsRepo.NewRepository(db, logger)
		setService := settingsService.NewSettingsService(setRepo, logger)
		setHandlers := settingsHandlers.NewSettingsHandlers(setService, logger)
		r.With(settingsMiddleware.UserContextMiddleware).Route("/settings", func(r chi.Router) {
			setHandlers.RegisterRoutes(r)
		})

		// Governance Domain (migrated in monolith)
		govRepo := governanceRepo.NewGovernanceRepository(db)
		govService := governanceService.NewGovernanceService(govRepo, nc, logger)
		
		// Enterprise Governance Extensions
		govAggregator := governanceService.NewGovernanceDataAggregator(db, logger)
		govMasker := governanceService.NewMaskingService()
		govGenerator := govUtils.NewExportFileGenerator()
		govExportSvc := governanceService.NewExportService(govRepo, govAggregator, govMasker, govGenerator, nc, logger)
		govDeletionSvc := governanceService.NewDataDeletionService(govRepo, db, nc, logger)
		govAccessReviewSvc := governanceService.NewAccessReviewService(govRepo, db, logger)

		// Workers
		govExpiryWorker := govWorkers.NewGovernanceExpiryWorker(govService, logger)
		go govExpiryWorker.Start(ctx)

		govExportWorker := govWorkers.NewExportProcessingWorker(govExportSvc, nc, logger)
		go govExportWorker.Start(ctx)

		govDeletionWorker := govWorkers.NewDeletionWorker(govDeletionSvc, nc, logger)
		go govDeletionWorker.Start(ctx)

		govRetentionWorker := govWorkers.NewGovernanceRetentionWorker(govService, logger)
		go govRetentionWorker.Start(ctx)

		govHandler := governanceHandlers.NewGovernanceHandler(govService, govExportSvc, govAccessReviewSvc)
		r.Route("/governance", func(r chi.Router) {
			govRateLimiter := govMiddleware.DSARRateLimit(sharedRedisClient, nc, logger)
			govHandler.RegisterRoutes(r, govRateLimiter)
		})

/*
		// Admin Support Domains (migrated in monolith, non-overlapping surfaces)
		adminSupportRepo := adminsupportsupportrepo.NewTicketRepository(db, logger)
		adminSupportService := adminsupportsupportservice.NewTicketService(adminSupportRepo, logger)
		adminSupportHandler := adminsupportsupporthandlers.NewTicketHandlers(adminSupportService, logger)
		adminSupportHandler.RegisterRoutes(r)

		adminDrRepo := adminsupportdrrepo.NewDrRepository(db, logger)
		adminDrService := adminsupportdrservice.NewDrService(adminDrRepo, logger)
		adminDrHandler := adminsupportdrhandlers.NewDrHandlers(adminDrService, logger)
		adminDrHandler.RegisterRoutes(r)

		adminOrgRepo := adminsupportorgrepo.NewOrganizationRepository(db, logger)
		adminOrgService := adminsupportorgservice.NewOrganizationService(adminOrgRepo, logger)
		adminOrgHandler := adminsupportorghandlers.NewOrganizationHandlers(adminOrgService, logger)
		adminOrgHandler.RegisterRoutes(r)

		if nc != nil {
			adminEmergencySvc := adminsupportemergencyservice.NewEmergencyService(nc, logger)
			adminEmergencyHandler := adminsupportemergencyhandlers.NewEmergencyHandler(adminEmergencySvc, logger)
			adminEmergencyHandler.RegisterRoutes(r)
		}
*/

/*
		// Alert Domains (migrated in monolith)
		alertRepo := alertalertrepo.NewAlertRepository(db)
		alertRuleEngine := alertalertrule.NewRuleEngine(alertRepo, logger)
		alertNotificationService := alertproviders.NewNotificationService(
			alertproviders.NewTwilioClient("", "", ""),
			alertproviders.NewSendGridClient(""),
			alertproviders.NewFirebaseClient(),
			logger,
		)
		alertSvc := alertalertservice.NewAlertService(alertRepo, alertRuleEngine, alertNotificationService, logger)
		alertHandler := alertalerthandlers.NewAlertHandler(alertSvc, logger)
		alertHandler.RegisterRoutes(r)

		notifyRepo := alertnotifyrepo.NewNotificationRepository(db)
		notifySvc := alertnotifyservice.NewNotificationService(notifyRepo, logger)
		notifyHandler := alertnotifyhandlers.NewNotificationHandlers(notifySvc, logger)
		notifyHandler.RegisterRoutes(r)
*/

/*
		// Tracking Domain (stage-1 local migration: device APIs)
		trkRepo := trackingv2repo.NewDeviceRepository(db, logger)
		trkService := trackingv2service.NewDeviceService(trkRepo, cfg.JWTSecret, nil, logger)
		trkHandler := trackingv2handlers.NewDeviceHandler(trkService, logger)
		trkHandler.RegisterRoutes(r)
			if sharedDB != nil {
				trkTelemetryHandler := trackingv2handlers.NewTelemetryHandler(sharedDB, sharedMB, logger)
				trkTelemetryHandler.RegisterRoutes(r)

				trkAnalytics := trackingv2analytics.NewTelemetryAnalytics(sharedDB, logger)
				trkAnalyticsHandler := trackingv2handlers.NewAnalyticsHandler(trkAnalytics, logger)
				// Avoid clashing with analytics domain routes mounted at /analytics.
				r.Route("/tracking", func(r chi.Router) {
					trkAnalyticsHandler.RegisterRoutes(r)
				})
			}
*/

/*
		// Auth Domain (migrated in monolith via REST gateway)
		authLog := logadapter.New(logger)
		authRepo := authv2repo.NewAuthRepository(dbSQL, authLog)
		authSvc := authv2service.NewAuthService(authLog, authRepo, cfg.JWTSecret)
		authHandler := authv2handlers.NewHandlers(authSvc, logger)
		authHandler.RegisterRoutes(r)

		// Dashboard & Support Domain (consolidated from BFF)
		supportRepo := repository.NewSupportRepository(db, logger)
		monitorHandlers := dashHandlers.NewMonitoringHandlers(logger, pulseService)
		dashboardHandlers := dashHandlers.NewDashboardHandlers(logger, authSvc, vService, adminOrgService, pulseService, supportRepo)
		supportHandlers := dashHandlers.NewSupportCompatHandlers(logger, authSvc, adminOrgService, supportRepo)

		r.Get("/dashboard/stats", dashboardHandlers.GetGlobalStats)
		r.Get("/admin/dashboard/data", dashboardHandlers.GetAdminDashboardData)

		r.Route("/support", func(r chi.Router) {
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
				r.Patch("/update", supportHandlers.AddUserNote)  // Compat for update
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

		// UI Configuration
		uiHandlers := &dashHandlers.UIHandlers{}
		r.Get("/ui/global_config", uiHandlers.GetGlobalConfig)

		// Monitoring Domain
		r.Get("/monitoring/health", monitorHandlers.GetSystemHealth)
		r.Get("/monitoring/sla", monitorHandlers.GetSLAMetrics)

		// ML Service Proxy (to Sidecar)
		if cfg.MLServiceURL != "" {
			mlProxy, err := gateway.NewServiceProxy("/api/v1/ml", cfg.MLServiceURL, logger)
			if err == nil {
				r.Handle("/ml", mlProxy)
				r.Handle("/ml/*", mlProxy)
			}
		}
*/

/*
		// Compliance Domain (migrated in monolith via REST gateway)
		compLog := logadapter.New(logger)
		compRepo := compliancev2repo.NewComplianceRepository(dbSQL, compLog)
		compSvc := compliancev2service.NewComplianceService(compLog, compRepo)
		compHandler := compliancev2handlers.NewHandlers(compSvc, logger)
		compHandler.RegisterRoutes(r)
*/

/*
		// Analytics Domain
		aRepo := analyticsRepo.NewAnalyticsRepository(db)
		aService := analyticsService.NewAnalyticsService(aRepo)
		aHandlers := analyticsHandlers.NewAnalyticsHandler(aService, logger)
		aHandlers.RegisterRoutes(r)
		// Compatibility alias for standalone analytics-service route shape.
		r.Route("/analytics", func(r chi.Router) {
			// (Truncated during previous edit, needs restoration if analytics is needed)
		})
*/

		// Finance Domain
		finRepo := financeRepo.NewFinanceRepository(db)
		finService := financeService.NewFinanceService(finRepo, logger)
		finHandlers := financeHandlers.NewFinanceHandlers(finService, logger)
		r.Route("/finance", func(r chi.Router) {
			finHandlers.RegisterRoutes(r)
		})
/*
		// Billing Domain (migrated v2 in monolith)
		billRepo := billingv2repo.NewBillingRepository(db)
		billService := billingv2service.NewBillingService(billRepo, logger)
		billHandler := billingv2handlers.NewBillingHandler(billService, logger)
		billHandler.RegisterRoutes(r)
*/

		// Emergency Domain
		emRepo := emergencyRepo.NewEmergencyRepository(db)
		emAnalyticsSvc := emergencyService.NewEmergencyAnalyticsService(emRepo)
		emService := emergencyService.NewEmergencyService(emRepo, liveCacheRepo, nc, sharedRedisClient, logger)
		emHandlers := emergencyHandlers.NewEmergencyHandlers(emService, emAnalyticsSvc, logger)
		r.Route("/emergency", func(r chi.Router) {
			emHandlers.RegisterRoutes(r)
		})
	})

/*
	// Video Domain
	vidRepo := videoRepo.NewVideoRepository(db)
	vidService := videoService.NewVideoService(vidRepo)
	vidHandlers := videoHandlers.NewVideoHandler(vidService)
	r.Route("/api/v1/video", func(r chi.Router) {
		vidHandlers.RegisterRoutes(r)
	})
*/

	// India legacy compatibility routes (/api/*) for existing Web-BFF mappings.
	indiaRepoLegacy := indiav2repo.NewIndiaIntegrationRepository(db)
	indiaServiceLegacy := indiav2service.NewIndiaIntegrationService(indiaRepoLegacy)
	indiaHandlersLegacy := indiav2handlers.NewIndiaIntegrationHandlers(indiaServiceLegacy)
	r.Route("/api", func(r chi.Router) {
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
	localMigratedServices := []string{"admin", "alert", "analytics", "auth", "billing", "compliance", "governance", "india", "integrations", "maintenance", "reports", "settings", "tracking", "trip", "vehicle"}
	upstreamNames := make([]string, 0, len(upstreamTargets))
	for name := range upstreamTargets {
		upstreamNames = append(upstreamNames, name)
	}
	sort.Strings(upstreamNames)
	for _, name := range upstreamNames {
		// locally migrated routes are served by this monolith.
		if name == "settings" || name == "governance" || name == "integrations" || name == "reports" || name == "maintenance" || name == "analytics" || name == "trip" || name == "vehicle" || name == "billing" || name == "admin" || name == "alert" || name == "auth" || name == "compliance" || name == "tracking" {
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "fleet-core-service-legacy",
			"report":  report,
		})
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		})
	})

	// 7. Initialize and Start Background Workers
	if nc != nil {
		ge := geoEngine.NewGeofenceEngine(db, nc, logger)
		go func() {
			if err := ge.Start(context.Background()); err != nil {
				logger.Error("Geofence engine failed", zap.Error(err))
			}
		}()

/*
		tw := maintenService.NewTelemetryWorker(mRepo, nc, logger)
		go func() {
			if err := tw.Start(context.Background()); err != nil {
				logger.Error("Telemetry worker failed", zap.Error(err))
			}
		}()
*/
	}

	// 8. Start Server
	r.Get("/swagger/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
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
	r.Get("/swagger/openapi.combined.json", func(w http.ResponseWriter, _ *http.Request) {
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
	r.Get("/swagger", func(w http.ResponseWriter, _ *http.Request) {
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
      url: "/swagger/openapi.combined.json",
      dom_id: "#swagger-ui"
    });
  </script>
</body>
</html>`))
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
