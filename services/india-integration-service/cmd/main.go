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
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	// Setup Chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(loggingMiddleware(logger))

	// Health check endpoint
	r.Get("/health", indiaHandlers.HealthCheck)

	r.Route("/api", func(r chi.Router) {
		// FASTag Transaction routes
		r.Route("/fastag", func(r chi.Router) {
			r.Post("/transactions", indiaHandlers.CreateFASTagTransaction)
			r.Post("/sync", indiaHandlers.SyncFASTag)
			r.Get("/transactions/{id}", indiaHandlers.GetFASTagTransaction)
			r.Get("/transactions", indiaHandlers.ListFASTagTransactions)
		})

		// E-Way Bill routes
		r.Route("/ewaybill", func(r chi.Router) {
			r.Post("/bills", indiaHandlers.CreateEWayBill)
			r.Post("/validate", indiaHandlers.ValidateEWayBill)
			r.Get("/bills/{id}", indiaHandlers.GetEWayBill)
			r.Get("/bills/number/{number}", indiaHandlers.GetEWayBillByNumber)
			r.Get("/bills", indiaHandlers.ListEWayBills)
		})

		// Vehicle Verification routes (Vahan integration)
		r.Route("/vahan", func(r chi.Router) {
			r.Post("/verifications", indiaHandlers.CreateVehicleVerification)
			r.Post("/verify", indiaHandlers.VerifyVehicle)
			r.Get("/verifications/{id}", indiaHandlers.GetVehicleVerification)
			r.Get("/verifications/reg/{reg_number}", indiaHandlers.GetVehicleVerificationByRegNumber)
			r.Get("/verifications", indiaHandlers.ListVehicleVerifications)
		})

		// Driver Verification routes (Sarathi integration)
		r.Route("/sarathi", func(r chi.Router) {
			r.Post("/verifications", indiaHandlers.CreateDriverVerification)
			r.Post("/verify", indiaHandlers.VerifyDriver)
			r.Get("/verifications/{id}", indiaHandlers.GetDriverVerification)
			r.Get("/verifications/license/{license_number}", indiaHandlers.GetDriverVerificationByLicenseNumber)
			r.Get("/verifications", indiaHandlers.ListDriverVerifications)
		})

		// Local Permit routes
		r.Route("/permits", func(r chi.Router) {
			r.Post("/", indiaHandlers.CreateLocalPermit)
			r.Get("/{id}", indiaHandlers.GetLocalPermit)
			r.Get("/number/{permit_number}", indiaHandlers.GetLocalPermitByNumber)
			r.Get("/", indiaHandlers.ListLocalPermits)
		})

		// GST Compliance routes
		r.Route("/gst", func(r chi.Router) {
			r.Post("/compliances", indiaHandlers.CreateGSTCompliance)
			r.Get("/compliances/{id}", indiaHandlers.GetGSTCompliance)
			r.Get("/compliances/gstin/{gstin}", indiaHandlers.GetGSTComplianceByGSTIN)
			r.Get("/compliances", indiaHandlers.ListGSTCompliances)
		})

		// Integration Config routes
		r.Route("/config", func(r chi.Router) {
			r.Post("/", indiaHandlers.CreateIntegrationConfig)
			r.Get("/{id}", indiaHandlers.GetIntegrationConfig)
			r.Get("/org/{org_id}/type/{service_type}", indiaHandlers.GetIntegrationConfigByServiceType)
			r.Get("/", indiaHandlers.ListIntegrationConfigs)
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
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

func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Custom response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			
			next.ServeHTTP(ww, r)
			
			duration := time.Since(start)
			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("duration", duration),
				zap.String("client_ip", r.RemoteAddr),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
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