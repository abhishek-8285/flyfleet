package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/database"
	"github.com/fleet-platform/shared/pkg/messaging"
	sharedMiddleware "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/tracking-service/internal/analytics"
	"github.com/flyfleet/tracking-service/internal/config"
	"github.com/flyfleet/tracking-service/internal/handlers"
	"github.com/flyfleet/tracking-service/internal/processor"
	"github.com/flyfleet/tracking-service/internal/repository"
	"github.com/flyfleet/tracking-service/internal/service"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logger.Sync()
	}()

	// Load configuration
	cfg := config.Load()

	// Database connection (using the standardized Postgres wrapper)
	db, err := database.NewPostgres(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Connected to database")

	// NATS connection via common messaging bridge (JetStream enabled)
	mb, err := messaging.NewNATSClient(cfg.NatsURL, logger)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer mb.Close()

	// Initialize Analytics Service
	analyticsService := analytics.NewTelemetryAnalytics(db, logger)

	// Initialize Device Repository and Service
	pg := db.(*database.Postgres)
	deviceRepo := repository.NewDeviceRepository(pg.GetPool(), logger)
	deviceService := service.NewDeviceService(deviceRepo, cfg.JWTSecret, nil, logger)
	deviceHandler := handlers.NewDeviceHandler(deviceService, logger)

	// Initialize Processor
	telemetryProcessor := processor.NewTelemetryProcessor(db, analyticsService, logger)

	// Subscribe to NATS subjects (Core NATS for now, migrate to JetStream consumer next)
	if err := mb.Subscribe("telemetry.>", func(data []byte) {
		// Shim to match HandleTelemetry signature
		// Since mb.Subscribe abstract away the nats.Msg, we create a mock msg if needed
		// or just call the processor logic directly if we refactor it.
		// For now, let's just make sure it's used.
		msg := &nats.Msg{
			Subject: "telemetry.gps", // This is a bit brittle, better to refactor processor
			Data:    data,
		}
		telemetryProcessor.HandleTelemetry(msg)
	}); err != nil {
		logger.Error("Failed to subscribe to telemetry", zap.Error(err))
	}

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(db, mb, logger)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, logger)

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	// Health check endpoint (public)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "tracking-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Public device routes (registration doesn't require auth)
	r.Post("/device/register", deviceHandler.Register)

	// Authenticated and Tenant-aware routes
	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.AuthMiddleware(cfg))
		r.Use(sharedMiddleware.TenantMiddleware)

		// Device routes
		deviceHandler.RegisterRoutes(r)

		// Telemetry routes
		telemetryHandler.RegisterRoutes(r)

		// Analytics routes
		analyticsHandler.RegisterRoutes(r)
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	go func() {
		logger.Info("Starting tracking service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down tracking service...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Tracking service exited")
}