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
	databaseURL := getEnv("DATABASE_URL", "postgres://fleet:fleet@localhost:5433/fleetdb")

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

	// Initialize service
	vehicleService := service.NewVehicleService(vehicleRepo, logger)

	// Initialize handlers
	vehicleHandlers := handlers.NewVehicleHandlers(vehicleService, logger)

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
