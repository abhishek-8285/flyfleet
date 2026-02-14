package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/fleet-core-service/internal/config"
	"github.com/flyfleet/fleet-core-service/internal/database"
	"github.com/flyfleet/fleet-core-service/internal/middleware"

	// Domains
	"github.com/flyfleet/fleet-core-service/internal/domains/integrations"
	"github.com/flyfleet/fleet-core-service/internal/domains/reports"
	vehHandlers "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/handlers"
	vehRepo "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/repository"
	vehService "github.com/flyfleet/fleet-core-service/internal/domains/vehicle/service"

	driverHandlers "github.com/flyfleet/fleet-core-service/internal/domains/driver/handlers"
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

	"github.com/gin-gonic/gin"
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
		// Continue anyway for now, or handle based on requirement
	} else {
		defer nc.Close()
	}

	// 5. Initialize Router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware(logger))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "fleet-core-service"})
	})

	api := r.Group("/api/v1")

	// 6. Initialize Domains

	// Vehicle Domain
	vRepo := vehRepo.NewVehicleRepository(db, logger)
	vService := vehService.NewVehicleService(vRepo, logger)
	vHandlers := vehHandlers.NewVehicleHandlers(vService, logger)

	vehicles := api.Group("/vehicles")
	{
		vehicles.GET("", vHandlers.ListVehicles)
		vehicles.POST("", vHandlers.CreateVehicle)
		vehicles.GET("/stats", vHandlers.GetStats)
		v := vehicles.Group("/:id")
		{
			v.GET("", vHandlers.GetVehicle)
			v.PUT("", vHandlers.UpdateVehicle)
			v.DELETE("", vHandlers.DeleteVehicle)
			v.POST("/assign-driver", vHandlers.AssignDriver)
		}
	}

	// Driver Domain
	dRepo := driverRepo.NewDriverRepository(db)
	dService := driverService.NewDriverService(dRepo, logger)
	dHandlers := driverHandlers.NewDriverHandler(dService, logger)

	drivers := api.Group("/drivers")
	{
		drivers.GET("", dHandlers.ListDrivers)
		drivers.POST("", dHandlers.CreateDriver)
		d := drivers.Group("/:id")
		{
			d.GET("", dHandlers.GetDriver)
			d.PUT("", dHandlers.UpdateDriver)
			d.DELETE("", dHandlers.DeleteDriver)
			d.POST("/assign-vehicle", dHandlers.AssignVehicle)
			d.POST("/unassign-vehicle", dHandlers.UnassignVehicle)
			d.POST("/update-status", dHandlers.UpdateStatus)
			d.GET("/behavior-score", dHandlers.GetBehaviorScore)
			d.POST("/update-behavior", dHandlers.UpdateBehaviorScore)
			d.GET("/analytics", dHandlers.GetAnalytics)
			d.POST("/documents", dHandlers.UploadDocument)
			d.GET("/documents", dHandlers.ListDocuments)
			d.DELETE("/documents/:doc_id", dHandlers.DeleteDocument)
			d.POST("/attendance", dHandlers.MarkAttendance)
			d.GET("/attendance", dHandlers.GetAttendance)
		}
	}

	// Trip Domain
	tRepo := tripRepo.NewTripRepository(db)
	tService := tripService.NewTripService(tRepo, logger)
	tHandlers := tripHandlers.NewTripHandler(tService, logger)

	trips := api.Group("/trips")
	{
		trips.GET("", tHandlers.ListTrips)
		trips.POST("", tHandlers.CreateTrip)
		t := trips.Group("/:id")
		{
			t.GET("", tHandlers.GetTrip)
			t.PUT("", tHandlers.UpdateTrip)
			t.POST("/start", tHandlers.StartTrip)
			t.POST("/pause", tHandlers.PauseTrip)
			t.POST("/resume", tHandlers.ResumeTrip)
			t.POST("/complete", tHandlers.CompleteTrip)
			t.POST("/cancel", tHandlers.CancelTrip)
			t.POST("/events", tHandlers.AddEvent)
			t.POST("/pod", tHandlers.SubmitPOD)
			t.GET("/optimize-route", tHandlers.OptimizeRoute)
		}
		trips.GET("/driver-stats", tHandlers.GetDriverStats)
		trips.GET("/trip-stats", tHandlers.GetTripStats)
		trips.GET("/stats", tHandlers.GetStats)
	}

	// Maintenance Domain
	mRepo := maintenRepo.NewMaintenanceRepository(db)
	mService := maintenService.NewMaintenanceService(mRepo, logger)
	mHandlers := maintenHandlers.NewMaintenanceHandler(mService, logger)

	maintenance := api.Group("/maintenance")
	{
		maintenance.POST("/schedules", mHandlers.CreateSchedule)
		maintenance.GET("/schedules", mHandlers.ListSchedules)
		maintenance.GET("/schedules/:id", mHandlers.GetSchedule)
		maintenance.PUT("/schedules/:id", mHandlers.UpdateSchedule)
		maintenance.DELETE("/schedules/:id", mHandlers.DeleteSchedule)

		maintenance.POST("/logs", mHandlers.CreateLog)
		maintenance.GET("/logs", mHandlers.ListLogs)
		maintenance.GET("/logs/:id", mHandlers.GetLog)

		maintenance.GET("/upcoming", mHandlers.GetUpcomingMaintenance)
		maintenance.GET("/overdue", mHandlers.GetOverdueMaintenance)
		maintenance.GET("/summary", mHandlers.GetMaintenanceSummary)
		maintenance.GET("/vehicles/:vehicle_id/history", mHandlers.GetVehicleMaintenanceHistory)
	}

	// Work Order Management (Maintenance Sub-domain)
	woRepo := maintenRepo.NewWorkOrderRepository(db)
	woPartsRepo := maintenRepo.NewWorkOrderPartsRepository(db)
	woInvRepo := maintenRepo.NewInventoryRepository(db)
	woSettingsRepo := maintenRepo.NewSettingsRepository(db)
	
	woService := maintenService.NewWorkOrderService(woRepo, woPartsRepo, woInvRepo, woSettingsRepo, logger)
	woHandlers := maintenHandlers.NewWorkOrderHandler(woService, logger)

	workOrders := api.Group("/work-orders")
	{
		workOrders.GET("", woHandlers.ListWorkOrders)
		workOrders.POST("", woHandlers.CreateWorkOrder)
		workOrders.GET("/stats", woHandlers.GetStats)
		wo := workOrders.Group("/:id")
		{
			wo.GET("", woHandlers.GetWorkOrder)
			wo.PUT("", woHandlers.UpdateWorkOrder)
			wo.PATCH("/status", woHandlers.ChangeStatus)
			wo.POST("/close", woHandlers.CloseWorkOrder)
			wo.GET("/parts", woHandlers.GetWorkOrderParts)
			wo.POST("/parts", woHandlers.AssignPart)
			wo.DELETE("/parts/:partId", woHandlers.RemovePart)
		}
	}

	// Inventory Management (Maintenance Sub-domain)
	invRepo := maintenRepo.NewInventoryRepository(db)
	invHandlers := maintenHandlers.NewInventoryHandler(invRepo, logger)

	inventory := api.Group("/inventory")
	{
		inventory.GET("", invHandlers.ListInventory)
		inventory.POST("", invHandlers.CreatePart)
		inv := inventory.Group("/:id")
		{
			inv.GET("", invHandlers.GetPart)
			inv.PUT("", invHandlers.UpdatePart)
			inv.DELETE("", invHandlers.DeletePart)
		}
	}

	// Geofencing Domain
	gRepo := geoRepo.NewGeofenceRepository(db)
	gService := geoService.NewGeofenceService(gRepo, nc, logger)
	gHandlers := geoHandlers.NewGeofenceHandler(gService, logger)

	geofencing := api.Group("/geofencing")
	{
		geofencing.POST("", gHandlers.CreateGeofence)
		geofencing.GET("", gHandlers.ListGeofences)
		geofencing.GET("/:id", gHandlers.GetGeofence)
		geofencing.PUT("/:id", gHandlers.UpdateGeofence)
		geofencing.DELETE("/:id", gHandlers.DeleteGeofence)
		geofencing.GET("/events", gHandlers.GetGeofenceEvents)
	}

	// Fuel Domain
	fLogRepo := fuelRepo.NewFuelLogRepository(db)
	fSettingsRepo := fuelRepo.NewSettingsRepository(db)
	fService := fuelService.NewFuelService(fLogRepo, fSettingsRepo, logger)
	fHandlers := fuelHandlers.NewFuelHandler(fService, logger)

	fuel := api.Group("/fuel")
	{
		fuel.POST("/logs", fHandlers.CreateLog)
		fuel.GET("/logs", fHandlers.ListLogs)
		fuel.GET("/logs/:id", fHandlers.GetLog)
		fuel.PUT("/logs/:id", fHandlers.UpdateLog)
		fuel.DELETE("/logs/:id", fHandlers.DeleteLog)
		fuel.PATCH("/logs/:id/status", fHandlers.ChangeStatus)
		fuel.GET("/stats", fHandlers.GetStats)
		fuel.GET("/settings", fHandlers.GetSettings)
		fuel.PUT("/settings", fHandlers.UpdateSettings)
	}

	// Route Optimization Domain
	rRepo := routeRepo.NewRouteOptimizationRepository(db)
	rService := routeService.NewRouteOptimizationService(rRepo)
	rHandlers := routeHandlers.NewRouteOptimizationHandler(rService, logger)

	routes := api.Group("/routes")
	{
		routes.POST("/requests", rHandlers.CreateOptimizationRequest)
		routes.GET("/requests/:id", rHandlers.GetOptimizationRequest)
		routes.PUT("/requests/:id", rHandlers.UpdateOptimizationRequest)
		routes.DELETE("/requests/:id", rHandlers.DeleteOptimizationRequest)
		routes.GET("/requests", rHandlers.ListOptimizationRequests)
		routes.GET("/results/:id", rHandlers.GetOptimizationResult)
		routes.GET("/results/request/:request_id", rHandlers.GetOptimizationResultByRequestID)
		routes.GET("/results", rHandlers.ListOptimizationResults)
		routes.POST("/jobs", rHandlers.SubmitOptimizationJob)
		routes.GET("/jobs/:id", rHandlers.GetOptimizationJob)
		routes.GET("/jobs", rHandlers.ListOptimizationJobs)
	}

	// Diagnostics Domain
	diagR := diagRepo.NewRepository(dbSQL, logger)
	diagS := diagService.NewDeviceDiagnosticsService(diagR, logger)
	diagH := diagHandlers.NewHandlers(diagS, logger)

	// Initialize Reports Domain
	reportsRepo := reports.NewRepository(db, logger)
	reportsService := reports.NewService(reportsRepo, logger)
	reportsHandler := reports.NewHandler(reportsService, logger)

	// Initialize Integrations Domain
	integrationsRepo := integrations.NewRepository(db, logger)
	integrationsService := integrations.NewService(integrationsRepo, logger)
	integrationsHandler := integrations.NewHandler(integrationsService, logger)

	// Initialize Analytics Domain
	aRepo := analyticsRepo.NewAnalyticsRepository(db)
	aService := analyticsService.NewAnalyticsService(aRepo)
	aHandlers := analyticsHandlers.NewAnalyticsHandler(aService)

	// Register Routes for new/refactored domains

	// Diagnostics Domain
	diagnostics := api.Group("/diagnostics")
	{
		diagnostics.GET("/summary", gin.WrapF(diagH.GetSummary))
		diagnostics.GET("/devices", gin.WrapF(diagH.ListDevices))
		diagnostics.GET("/devices/:id", gin.WrapF(diagH.GetDevice))
		diagnostics.GET("/devices/:id/sensors", gin.WrapF(diagH.GetDeviceSensors))
		diagnostics.GET("/devices/:id/logs", gin.WrapF(diagH.GetDeviceLogs))
		diagnostics.GET("/devices/:id/telemetry", gin.WrapF(diagH.GetDeviceTelemetry))
		diagnostics.POST("/devices/:id/command", gin.WrapF(diagH.SendCommand))
		diagnostics.GET("/devices/:id/commands", gin.WrapF(diagH.GetCommandHistory))
		diagnostics.GET("/firmware", gin.WrapF(diagH.ListFirmware))
		diagnostics.POST("/firmware", gin.WrapF(diagH.CreateFirmware))
		diagnostics.GET("/rollouts", gin.WrapF(diagH.ListRollouts))
		diagnostics.POST("/rollouts", gin.WrapF(diagH.CreateRollout))
	}

	// Reports Domain
	reportsHandler.RegisterRoutes(api)

	// Integrations Domain
	integrationsHandler.RegisterRoutes(api)

	// Analytics Domain
	aHandlers.RegisterRoutes(api)

	// Video Domain
	vidRepo := videoRepo.NewVideoRepository(db)
	vidService := videoService.NewVideoService(vidRepo)
	vidHandlers := videoHandlers.NewVideoHandler(vidService)
	vidHandlers.RegisterRoutes(r)

	// 7. Initialize and Start Background Workers
	if nc != nil {
		// Geofence Engine
		ge := geoEngine.NewGeofenceEngine(db, nc, logger)
		go func() {
			if err := ge.Start(context.Background()); err != nil {
				logger.Error("Geofence engine failed", zap.Error(err))
			}
		}()

		// Telemetry Worker (Maintenance)
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
