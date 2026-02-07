package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/india-integration-service/internal/handlers"
	"github.com/flyfleet/india-integration-service/internal/repository"
	"github.com/flyfleet/india-integration-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Fatal("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database connection
	db, err := initDB(dbURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize repositories
	indiaRepo := repository.NewIndiaIntegrationRepository(db)

	// Initialize services
	indiaService := service.NewIndiaIntegrationService(indiaRepo)

	// Initialize handlers
	indiaHandlers := handlers.NewIndiaIntegrationHandlers(indiaService)

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))

	// Health check endpoint
	router.GET("/health", indiaHandlers.HealthCheck)

	// FASTag Transaction routes
	fastagGroup := router.Group("/api/fastag")
	{
		fastagGroup.POST("/transactions", indiaHandlers.CreateFASTagTransaction)
            fastagGroup.POST("/sync", indiaHandlers.SyncFASTag)
		fastagGroup.GET("/transactions/:id", indiaHandlers.GetFASTagTransaction)
		fastagGroup.GET("/transactions", indiaHandlers.ListFASTagTransactions)
	}

	// E-Way Bill routes
	ewayGroup := router.Group("/api/ewaybill")
	{
		ewayGroup.POST("/bills", indiaHandlers.CreateEWayBill)
            ewayGroup.POST("/validate", indiaHandlers.ValidateEWayBill)
		ewayGroup.GET("/bills/:id", indiaHandlers.GetEWayBill)
		ewayGroup.GET("/bills/number/:number", indiaHandlers.GetEWayBillByNumber)
		ewayGroup.GET("/bills", indiaHandlers.ListEWayBills)
	}

	// Vehicle Verification routes (Vahan integration)
	vahanGroup := router.Group("/api/vahan")
	{
		vahanGroup.POST("/verifications", indiaHandlers.CreateVehicleVerification)
            vahanGroup.POST("/verify", indiaHandlers.VerifyVehicle)
		vahanGroup.GET("/verifications/:id", indiaHandlers.GetVehicleVerification)
		vahanGroup.GET("/verifications/reg/:reg_number", indiaHandlers.GetVehicleVerificationByRegNumber)
		vahanGroup.GET("/verifications", indiaHandlers.ListVehicleVerifications)
	}

	// Driver Verification routes (Sarathi integration)
	sarathiGroup := router.Group("/api/sarathi")
	{
		sarathiGroup.POST("/verifications", indiaHandlers.CreateDriverVerification)
            sarathiGroup.POST("/verify", indiaHandlers.VerifyDriver)
		sarathiGroup.GET("/verifications/:id", indiaHandlers.GetDriverVerification)
		sarathiGroup.GET("/verifications/license/:license_number", indiaHandlers.GetDriverVerificationByLicenseNumber)
		sarathiGroup.GET("/verifications", indiaHandlers.ListDriverVerifications)
	}

	// Local Permit routes
	permitGroup := router.Group("/api/permits")
	{
		permitGroup.POST("", indiaHandlers.CreateLocalPermit)
		permitGroup.GET("/:id", indiaHandlers.GetLocalPermit)
		permitGroup.GET("/number/:permit_number", indiaHandlers.GetLocalPermitByNumber)
		permitGroup.GET("", indiaHandlers.ListLocalPermits)
	}

	// GST Compliance routes
	gstGroup := router.Group("/api/gst")
	{
		gstGroup.POST("/compliances", indiaHandlers.CreateGSTCompliance)
		gstGroup.GET("/compliances/:id", indiaHandlers.GetGSTCompliance)
		gstGroup.GET("/compliances/gstin/:gstin", indiaHandlers.GetGSTComplianceByGSTIN)
		gstGroup.GET("/compliances", indiaHandlers.ListGSTCompliances)
	}

	// Integration Config routes
	configGroup := router.Group("/api/config")
	{
		configGroup.POST("", indiaHandlers.CreateIntegrationConfig)
		configGroup.GET("/:id", indiaHandlers.GetIntegrationConfig)
		configGroup.GET("/org/:org_id/type/:service_type", indiaHandlers.GetIntegrationConfigByServiceType)
		configGroup.GET("", indiaHandlers.ListIntegrationConfigs)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Channel to listen for interrupt signal
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Info("Starting India Integration Service", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	<-stopChan

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
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

// initDB initializes the database connection pool
func initDB(dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings
	config.MinConns = 2
	config.MaxConns = 20
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 10 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}