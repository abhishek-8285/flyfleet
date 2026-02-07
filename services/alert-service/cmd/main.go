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

	"github.com/flyfleet/alert-service/internal/handlers"
	"github.com/flyfleet/alert-service/internal/notifications"
	"github.com/flyfleet/alert-service/internal/repository"
	"github.com/flyfleet/alert-service/internal/rule_engine"
	"github.com/flyfleet/alert-service/internal/service"
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
	port := getEnv("PORT", "8085")
	dbURL := getEnv("DATABASE_URL", "postgres://fleet:fleet123@localhost:5432/fleetdb")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)
	twilioAccountSID := getEnv("TWILIO_ACCOUNT_SID", "")
	twilioAuthToken := getEnv("TWILIO_AUTH_TOKEN", "")
	twilioPhoneNumber := getEnv("TWILIO_PHONE_NUMBER", "")
	sendgridAPIKey := getEnv("SENDGRID_API_KEY", "")
	_ = getEnv("AWS_REGION", "ap-south-1")

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

	// Initialize notification clients
	smsClient := notifications.NewTwilioClient(twilioAccountSID, twilioAuthToken, twilioPhoneNumber)
	emailClient := notifications.NewSendGridClient(sendgridAPIKey)
	pushClient := notifications.NewFirebaseClient() // Will be implemented later

	// Initialize components
	alertRepo := repository.NewAlertRepository(db)
	ruleEngine := rule_engine.NewRuleEngine(alertRepo, logger)
	notificationService := notifications.NewNotificationService(smsClient, emailClient, pushClient, logger)
	
	alertService := service.NewAlertService(alertRepo, ruleEngine, notificationService, logger)
	alertHandler := handlers.NewAlertHandler(alertService, logger)

	// Subscribe to NATS topics for real-time alert processing
	subscribeToTelemetryEvents(nc, ruleEngine, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "alert-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Alert API endpoints
	api := r.Group("/api/v1")
	{
		alerts := api.Group("/alerts")
		{
			// Alert rules management
			alerts.POST("/rules", alertHandler.CreateRule)
			alerts.GET("/rules", alertHandler.ListRules)
			alerts.GET("/rules/:id", alertHandler.GetRule)
			alerts.PUT("/rules/:id", alertHandler.UpdateRule)
			alerts.DELETE("/rules/:id", alertHandler.DeleteRule)
			alerts.POST("/rules/:id/enable", alertHandler.EnableRule)
			alerts.POST("/rules/:id/disable", alertHandler.DisableRule)

			// Alert instances
			alerts.GET("", alertHandler.ListAlerts)
			alerts.GET("/:id", alertHandler.GetAlert)
			alerts.POST("/:id/acknowledge", alertHandler.AcknowledgeAlert)
			alerts.POST("/:id/resolve", alertHandler.ResolveAlert)
			alerts.POST("/:id/escalate", alertHandler.EscalateAlert)

			// Alert statistics and analytics
			alerts.GET("/stats/summary", alertHandler.GetAlertSummary)
			alerts.GET("/stats/trends", alertHandler.GetAlertTrends)
		}

		// Geofence endpoints
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting alert service", zap.String("port", port))
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

func subscribeToTelemetryEvents(nc *nats.Conn, ruleEngine *rule_engine.RuleEngine, logger *zap.Logger) {
	// Implementation for subscribing to telemetry events
	// This would typically subscribe to NATS topics for vehicle telemetry data
	logger.Info("Subscribed to telemetry events")
}