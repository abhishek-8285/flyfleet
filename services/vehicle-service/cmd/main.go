package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/vehicle-service/internal/handlers"
	"github.com/flyfleet/vehicle-service/internal/processor"
	"github.com/flyfleet/vehicle-service/internal/repository"
	"github.com/flyfleet/vehicle-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration from environment
	port := getEnv("PORT", "8084")
	databaseURL := getEnv("DATABASE_URL", "")

	// Initialize database connection
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}

	logger.Info("Connected to database successfully")

	// Initialize repository
	vehicleRepo := repository.NewVehicleRepository(db, logger)
	bulkRepo := repository.NewBulkRepository(db)

	// Initialize processor
	bulkProcessor := processor.NewBulkProcessor(bulkRepo, vehicleRepo, logger)

	// Initialize service
	vehicleService := service.NewVehicleService(vehicleRepo, logger)
	bulkService := service.NewBulkService(bulkRepo, bulkProcessor, logger)
	
	// Trip management
	tripRepo := repository.NewTripRepository(db)
	tripService := service.NewTripService(tripRepo, logger)
	tripHandlers := handlers.NewTripHandlers(tripService, logger)

	// Initialize handlers
	vehicleHandlers := handlers.NewVehicleHandlers(vehicleService, logger)
	bulkHandlers := handlers.NewBulkHandlers(bulkService, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "vehicle-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Vehicle routes
	vehicles := r.Group("/api/v1/vehicles")
	{
		vehicles.GET("", func(c *gin.Context) { vehicleHandlers.ListVehicles(c.Writer, c.Request) })
		vehicles.POST("", func(c *gin.Context) { vehicleHandlers.CreateVehicle(c.Writer, c.Request) })
		
		vehicle := vehicles.Group("/:id")
		{
			vehicle.GET("", func(c *gin.Context) { vehicleHandlers.GetVehicle(c.Writer, c.Request) })
			vehicle.PUT("", func(c *gin.Context) { vehicleHandlers.UpdateVehicle(c.Writer, c.Request) })
			vehicle.DELETE("", func(c *gin.Context) { vehicleHandlers.DeleteVehicle(c.Writer, c.Request) })
			vehicle.POST("/assign-driver", func(c *gin.Context) { vehicleHandlers.AssignDriver(c.Writer, c.Request) })
		}
	}

	// Bulk Job routes
	bulkJobs := r.Group("/api/v1/bulk-jobs")
	{
		bulkJobs.POST("", func(c *gin.Context) { bulkHandlers.CreateBulkJob(c.Writer, c.Request) })
		bulkJobs.GET("", func(c *gin.Context) { bulkHandlers.ListBulkJobs(c.Writer, c.Request) })
		bulkJobs.GET("/:id", func(c *gin.Context) { bulkHandlers.GetBulkJob(c.Writer, c.Request) })
		bulkJobs.GET("/:id/records", func(c *gin.Context) { bulkHandlers.GetJobRecords(c.Writer, c.Request) })
		bulkJobs.POST("/:id/retry", func(c *gin.Context) { bulkHandlers.RetryBulkJob(c.Writer, c.Request) })
	}

	// Trip routes
	trips := r.Group("/api/v1/trips")
	{
		trips.GET("", func(c *gin.Context) { tripHandlers.ListTrips(c.Writer, c.Request) })
		trips.POST("", func(c *gin.Context) { tripHandlers.CreateTrip(c.Writer, c.Request) })
		trips.GET("/settings", func(c *gin.Context) { tripHandlers.GetSettings(c.Writer, c.Request) })
		trips.PUT("/settings", func(c *gin.Context) { tripHandlers.UpdateSettings(c.Writer, c.Request) })
		trips.GET("/:id", func(c *gin.Context) { tripHandlers.GetTrip(c.Writer, c.Request) })
		trips.PUT("/:id", func(c *gin.Context) { tripHandlers.UpdateTrip(c.Writer, c.Request) })
		trips.GET("/:id/telemetry", func(c *gin.Context) { tripHandlers.GetTelemetry(c.Writer, c.Request) })
	}

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in goroutine
	go func() {
		logger.Info("Starting vehicle service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down vehicle service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Vehicle service exited gracefully")
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
