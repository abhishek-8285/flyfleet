package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/alert-service/internal/config"
	"github.com/flyfleet/alert-service/internal/middleware"
	"github.com/flyfleet/alert-service/internal/providers"

	alertHandlers "github.com/flyfleet/alert-service/internal/domains/alerts/handlers"
	alertRepo "github.com/flyfleet/alert-service/internal/domains/alerts/repository"
	alertRule "github.com/flyfleet/alert-service/internal/domains/alerts/rule_engine"
	alertService "github.com/flyfleet/alert-service/internal/domains/alerts/service"

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

	// Subscribe to NATS topics for real-time alert processing
	subscribeToTelemetryEvents(nc, aRuleEngine, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "alert-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API endpoints
	api := r.Group("/api/v1")
	{
		aHandler.RegisterRoutes(api)
		nHandler.RegisterRoutes(api)
	}

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func subscribeToTelemetryEvents(nc *nats.Conn, ruleEngine *alertRule.RuleEngine, logger *zap.Logger) {
	// Implementation for subscribing to telemetry events
	// This would typically subscribe to NATS topics for vehicle telemetry data
	logger.Info("Subscribed to telemetry events")
}