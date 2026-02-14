package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/maintenance-service/internal/handlers"
	"github.com/flyfleet/maintenance-service/internal/repository"
	"github.com/flyfleet/maintenance-service/internal/service"
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

	// Initialize components
	maintenanceRepo := repository.NewMaintenanceRepository(db)
	maintenanceService := service.NewMaintenanceService(maintenanceRepo, logger)
	maintenanceHandler := handlers.NewMaintenanceHandler(maintenanceService, logger)

	// Initialize and start Telemetry Worker
	telemetryWorker := service.NewTelemetryWorker(maintenanceRepo, nc, logger)
	if err := telemetryWorker.Start(context.Background()); err != nil {
		logger.Error("Failed to start telemetry worker", zap.Error(err))
	}

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "maintenance-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Maintenance API endpoints
	api := r.Group("/api/v1")
	{
		maintenance := api.Group("/maintenance")
		{
			// Maintenance schedules
			maintenance.POST("/schedules", maintenanceHandler.CreateSchedule)
			maintenance.GET("/schedules", maintenanceHandler.ListSchedules)
			maintenance.GET("/schedules/:id", maintenanceHandler.GetSchedule)
			maintenance.PUT("/schedules/:id", maintenanceHandler.UpdateSchedule)
			maintenance.DELETE("/schedules/:id", maintenanceHandler.DeleteSchedule)

			// Maintenance logs
			maintenance.POST("/logs", maintenanceHandler.CreateLog)
			maintenance.GET("/logs", maintenanceHandler.ListLogs)
			maintenance.GET("/logs/:id", maintenanceHandler.GetLog)
			maintenance.PUT("/logs/:id", maintenanceHandler.UpdateLog)
			maintenance.DELETE("/logs/:id", maintenanceHandler.DeleteLog)

			// Maintenance summaries and reports
			maintenance.GET("/upcoming", maintenanceHandler.GetUpcomingMaintenance)
			maintenance.GET("/overdue", maintenanceHandler.GetOverdueMaintenance)
			maintenance.GET("/summary", maintenanceHandler.GetMaintenanceSummary)
		}

		// Vehicle-specific maintenance endpoints
		vehicles := api.Group("/vehicles")
		{
			vehicles.GET("/:vehicle_id/maintenance-history", maintenanceHandler.GetVehicleMaintenanceHistory)
		}
	}

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Maintenance service exited")
}

func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		status := c.Writer.Status()

		logger.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
