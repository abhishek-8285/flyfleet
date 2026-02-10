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

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/flyfleet/tracking-service/internal/handlers"
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
	port := getEnv("PORT", "8082")
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

	logger.Info("Connected to NATS", zap.String("url", natsURL))

	// Initialize handlers
	telemetryHandler := handlers.NewTelemetryHandler(db, nc, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "tracking-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Device telemetry endpoints (API key auth)
	r.POST("/api/v1/device/telemetry/gps", telemetryHandler.IngestGPS)
	r.POST("/api/v1/device/telemetry/obd", telemetryHandler.IngestOBD)

	// Web admin endpoints (JWT auth)
	r.GET("/api/v1/vehicles/:id/history", telemetryHandler.GetVehicleHistory)

	// Legacy endpoints for backward compatibility
	r.POST("/telemetry/gps", func(c *gin.Context) {
		telemetryHandler.IngestGPS(c)
	})
	r.POST("/telemetry/obd", func(c *gin.Context) {
		telemetryHandler.IngestOBD(c)
	})

	// Analytics endpoints
	r.POST("/telemetry/fuel", func(c *gin.Context) {
		var payload map[string]interface{}
		if err := c.BindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		// Add timestamp if not present
		if _, exists := payload["timestamp"]; !exists {
			payload["timestamp"] = time.Now().Unix()
		}

		// Publish to NATS for processing
		data, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
			return
		}

		err = nc.Publish("fleet.telemetry.fuel", data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Publish failed"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status": "accepted",
			"message": "Fuel telemetry received and queued for processing",
		})
	})

	// Batch telemetry endpoint
	r.POST("/telemetry/batch", func(c *gin.Context) {
		var batch struct {
			GPSData  []map[string]interface{} `json:"gps_data,omitempty"`
			OBData   []map[string]interface{} `json:"obd_data,omitempty"`
			FuelData []map[string]interface{} `json:"fuel_data,omitempty"`
			DeviceID string                   `json:"device_id"`
		}

		if err := c.BindJSON(&batch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		// Process GPS data
		for _, gps := range batch.GPSData {
			gps["device_id"] = batch.DeviceID
			gps["timestamp"] = time.Now().Unix()
			
			data, _ := json.Marshal(gps)
			nc.Publish("fleet.telemetry.gps", data)
		}

		// Process OBD data
		for _, obd := range batch.OBData {
			obd["device_id"] = batch.DeviceID
			obd["timestamp"] = time.Now().Unix()
			
			data, _ := json.Marshal(obd)
			nc.Publish("fleet.telemetry.obd", data)
		}

		// Process fuel data
		for _, fuel := range batch.FuelData {
			fuel["device_id"] = batch.DeviceID
			fuel["timestamp"] = time.Now().Unix()
			
			data, _ := json.Marshal(fuel)
			nc.Publish("fleet.telemetry.fuel", data)
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status": "accepted",
			"message": fmt.Sprintf("Batch processed: %%d GPS, %%d OBD, %%d Fuel records",
				len(batch.GPSData), len(batch.OBData), len(batch.FuelData)),
		})
	})

	// Device heartbeat endpoint
	r.POST("/devices/heartbeat", func(c *gin.Context) {
		var heartbeat map[string]interface{}
		if err := c.BindJSON(&heartbeat); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
			return
		}

		heartbeat["timestamp"] = time.Now().Unix()
		
		data, err := json.Marshal(heartbeat)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
			return
		}

		err = nc.Publish("fleet.devices.heartbeat", data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Publish failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "received",
			"message": "Device heartbeat acknowledged",
		})
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	// Run server in goroutine
	go func() {
		logger.Info("Starting tracking service", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down tracking service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Tracking service exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
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