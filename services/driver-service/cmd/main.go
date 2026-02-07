package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/driver-service/internal/handlers"
	"github.com/flyfleet/driver-service/internal/repository"
	"github.com/flyfleet/driver-service/internal/service"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	port := getEnv("PORT", "8083")
	dbURL := getEnv("DATABASE_URL", "postgres://fleet:fleet123@localhost:5432/fleetdb")
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
	driverRepo := repository.NewDriverRepository(db)
	driverService := service.NewDriverService(driverRepo, logger)
	driverHandler := handlers.NewDriverHandler(driverService, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "driver-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Driver API endpoints
	api := r.Group("/api/v1")
	{
		drivers := api.Group("/drivers")
		{
			drivers.POST("", driverHandler.CreateDriver)
			drivers.GET("", driverHandler.ListDrivers)
			drivers.GET("/:id", driverHandler.GetDriver)
			drivers.PUT("/:id", driverHandler.UpdateDriver)
			drivers.DELETE("/:id", driverHandler.DeleteDriver)
			
			// Driver assignment and status
			drivers.POST("/:id/assign-vehicle", driverHandler.AssignVehicle)
			drivers.POST("/:id/unassign-vehicle", driverHandler.UnassignVehicle)
			drivers.POST("/:id/update-status", driverHandler.UpdateStatus)
			
			// Driver behavior and scoring
			
			// Driver documents and licensing
			drivers.POST("/:id/documents", driverHandler.UploadDocument)
			drivers.GET("/:id/documents", driverHandler.ListDocuments)
			drivers.DELETE("/:id/documents/:doc_id", driverHandler.DeleteDocument)
			
			// Driver attendance and scheduling
			drivers.POST("/:id/attendance", driverHandler.MarkAttendance)
			drivers.GET("/:id/attendance", driverHandler.GetAttendance)
			drivers.POST("/:id/schedule", driverHandler.SetSchedule)
			drivers.GET("/:id/schedule", driverHandler.GetSchedule)
		}
	}

	// Start server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting driver service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down driver service...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Driver service exited")
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

// Helper function to parse pagination parameters
func parseIntParam(param string, defaultValue int) int {
	if param == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(param)
	if err != nil || value < 1 {
		return defaultValue
	}
	
	return value
}
