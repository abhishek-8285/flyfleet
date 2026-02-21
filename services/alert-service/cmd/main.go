package main

import (
	"context"
	"encoding/json"
	"fmt"
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

	sharedMiddleware "github.com/fleet-platform/shared/pkg/middleware"
	"github.com/flyfleet/alert-service/internal/config"
	"github.com/flyfleet/alert-service/internal/providers"

	alertgrpc "github.com/flyfleet/alert-service/internal/grpc"
	alertHandlers "github.com/flyfleet/alert-service/internal/domains/alerts/handlers"
	alertRepo "github.com/flyfleet/alert-service/internal/domains/alerts/repository"
	alertRule "github.com/flyfleet/alert-service/internal/domains/alerts/rule_engine"
	alertService "github.com/flyfleet/alert-service/internal/domains/alerts/service"
	pb "github.com/flyfleet/proto/alert/v1"

	notifyHandlers "github.com/flyfleet/alert-service/internal/domains/notifications/handlers"
	notifyRepo "github.com/flyfleet/alert-service/internal/domains/notifications/repository"
	notifyService "github.com/flyfleet/alert-service/internal/domains/notifications/service"
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

	// Database connection
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
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
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", zap.Error(err))
	}
	defer nc.Close()
	logger.Info("Connected to NATS")

	// Initialize notification clients (Providers)
	smsClient := providers.NewTwilioClient(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioPhoneNumber)
	emailClient := providers.NewSendGridClient(cfg.SendGridAPIKey)
	pushClient := providers.NewFirebaseClient() // Will be implemented later

	// Initialize Alert Domain
	aRepo := alertRepo.NewAlertRepository(db)
	aRuleEngine := alertRule.NewRuleEngine(aRepo, logger)
	aNotificationSvc := providers.NewNotificationService(smsClient, emailClient, pushClient, logger)
	aService := alertService.NewAlertService(aRepo, aRuleEngine, aNotificationSvc, logger)
	aHandler := alertHandlers.NewAlertHandler(aService, logger)

	// Initialize Notification Domain
	nRepo := notifyRepo.NewNotificationRepository(db)
	nService := notifyService.NewNotificationService(nRepo, logger)
	nHandler := notifyHandlers.NewNotificationHandlers(nService, logger)

	// Start gRPC Server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50054"
	}
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}
		
		grpcServer := google_grpc.NewServer()
		alertServer := alertgrpc.NewAlertServer(aService, logger)
		pb.RegisterAlertServiceServer(grpcServer, alertServer)
		
		logger.Info("Starting alert gRPC server", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Subscribe to NATS topics for real-time alert processing
	subscribeToTelemetryEvents(nc, aRuleEngine, logger)

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "alert-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API endpoints
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(sharedMiddleware.AuthMiddleware(cfg))
		aHandler.RegisterRoutes(r)
		nHandler.RegisterRoutes(r)
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting alert service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down alert service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Alert service exited gracefully")
}

func subscribeToTelemetryEvents(nc *nats.Conn, ruleEngine *alertRule.RuleEngine, logger *zap.Logger) {
	// Implementation for subscribing to telemetry events
	// This would typically subscribe to NATS topics for vehicle telemetry data
	logger.Info("Subscribed to telemetry events")
}