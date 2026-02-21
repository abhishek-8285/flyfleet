package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"

	maintenancegrpc "github.com/flyfleet/maintenance-service/internal/grpc"
	"github.com/flyfleet/maintenance-service/internal/handlers"
	"github.com/flyfleet/maintenance-service/internal/repository"
	"github.com/flyfleet/maintenance-service/internal/service"
	pb "github.com/flyfleet/proto/maintenance/v1"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	port := getEnv("PORT", "8086")
	dbURL := getEnv("DATABASE_URL", "")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)

	// Database connection
	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Connected to database")

	// NATS connection
	nc, err := nats.Connect(natsURL)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()
	logger.Info("Connected to NATS")

	// Initialize repositories
	maintenanceRepo := repository.NewMaintenanceRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	workOrderRepo := repository.NewWorkOrderRepository(db)
	partsRepo := repository.NewWorkOrderPartsRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Initialize services
	maintenanceService := service.NewMaintenanceService(maintenanceRepo, logger)
	workOrderService := service.NewWorkOrderService(workOrderRepo, partsRepo, inventoryRepo, settingsRepo, logger)

	// Initialize handlers
	maintenanceHandler := handlers.NewMaintenanceHandler(maintenanceService, logger)
	inventoryHandler := handlers.NewInventoryHandler(inventoryRepo, logger)
	workOrderHandler := handlers.NewWorkOrderHandler(workOrderService, logger)

	// Initialize and start Telemetry Worker
	telemetryWorker := service.NewTelemetryWorker(maintenanceRepo, nc, logger)
	if err := telemetryWorker.Start(context.Background()); err != nil {
		logger.Error("Failed to start telemetry worker", zap.Error(err))
	}

	// Start gRPC Server
	grpcPort := getEnv("GRPC_PORT", "50057")
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}
		
		grpcServer := google_grpc.NewServer()
		mSrv := maintenancegrpc.NewMaintenanceServer(maintenanceService, workOrderService, logger)
		pb.RegisterMaintenanceServiceServer(grpcServer, mSrv)
		
		logger.Info("Starting maintenance gRPC server", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "maintenance-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Maintenance routes
		r.Route("/maintenance", func(r chi.Router) {
			r.Post("/schedules", maintenanceHandler.CreateSchedule)
			r.Get("/schedules", maintenanceHandler.ListSchedules)
			r.Route("/schedules/{id}", func(r chi.Router) {
				r.Get("/", maintenanceHandler.GetSchedule)
				r.Put("/", maintenanceHandler.UpdateSchedule)
				r.Delete("/", maintenanceHandler.DeleteSchedule)
			})

			r.Post("/logs", maintenanceHandler.CreateLog)
			r.Get("/logs", maintenanceHandler.ListLogs)
			r.Route("/logs/{id}", func(r chi.Router) {
				r.Get("/", maintenanceHandler.GetLog)
				r.Put("/", maintenanceHandler.UpdateLog)
				r.Delete("/", maintenanceHandler.DeleteLog)
			})

			r.Get("/upcoming", maintenanceHandler.GetUpcomingMaintenance)
			r.Get("/overdue", maintenanceHandler.GetOverdueMaintenance)
			r.Get("/summary", maintenanceHandler.GetMaintenanceSummary)
		})

		// Inventory routes
		r.Route("/inventory", func(r chi.Router) {
			r.Get("/", inventoryHandler.ListInventory)
			r.Post("/", inventoryHandler.CreatePart)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", inventoryHandler.GetPart)
				r.Put("/", inventoryHandler.UpdatePart)
				r.Delete("/", inventoryHandler.DeletePart)
			})
		})

		// Work Order routes
		r.Route("/work-orders", func(r chi.Router) {
			r.Get("/", workOrderHandler.ListWorkOrders)
			r.Post("/", workOrderHandler.CreateWorkOrder)
			r.Get("/stats", workOrderHandler.GetStats)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", workOrderHandler.GetWorkOrder)
				r.Put("/", workOrderHandler.UpdateWorkOrder)
				r.Patch("/status", workOrderHandler.ChangeStatus)
				r.Post("/close", workOrderHandler.CloseWorkOrder)
				
				r.Route("/parts", func(r chi.Router) {
					r.Get("/", workOrderHandler.GetWorkOrderParts)
					r.Post("/", workOrderHandler.AssignPart)
					r.Delete("/{partId}", workOrderHandler.RemovePart)
				})
			})
		})

		// Vehicle-specific maintenance endpoints
		r.Get("/vehicles/{vehicle_id}/maintenance-history", maintenanceHandler.GetVehicleMaintenanceHistory)
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting maintenance service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down maintenance service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Maintenance service exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
