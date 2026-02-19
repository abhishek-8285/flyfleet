package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/flyfleet/admin-support-service/internal/config"
	supportHandlers "github.com/flyfleet/admin-support-service/internal/domains/support/handlers"
	supportRepo "github.com/flyfleet/admin-support-service/internal/domains/support/repository"
	supportService "github.com/flyfleet/admin-support-service/internal/domains/support/service"

	settingsHandlers "github.com/flyfleet/admin-support-service/internal/domains/settings/handlers"
	settingsRepo "github.com/flyfleet/admin-support-service/internal/domains/settings/repository"
	settingsService "github.com/flyfleet/admin-support-service/internal/domains/settings/service"

	governanceHandlers "github.com/flyfleet/admin-support-service/internal/domains/governance/handlers"
	governanceRepo "github.com/flyfleet/admin-support-service/internal/domains/governance/repository"
	governanceService "github.com/flyfleet/admin-support-service/internal/domains/governance/service"

	drHandlers "github.com/flyfleet/admin-support-service/internal/domains/dr/handlers"
	drRepo "github.com/flyfleet/admin-support-service/internal/domains/dr/repository"
	drService "github.com/flyfleet/admin-support-service/internal/domains/dr/service"

	orgHandlers "github.com/flyfleet/admin-support-service/internal/domains/organizations/handlers"
	orgRepo "github.com/flyfleet/admin-support-service/internal/domains/organizations/repository"
	orgService "github.com/flyfleet/admin-support-service/internal/domains/organizations/service"

	sharedMiddleware "github.com/fleet-platform/shared/pkg/middleware"
)

// envSecretProvider implements sharedMiddleware.SecretProvider
type envSecretProvider struct {
	secret string
}

func (p *envSecretProvider) GetJWTSecret() string {
	return p.secret
}

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Database connection
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Support Domain
	sRepo := supportRepo.NewTicketRepository(db, logger)
	sService := supportService.NewTicketService(sRepo, logger)
	sHandler := supportHandlers.NewTicketHandlers(sService, logger)

	// Initialize Settings Domain
	setRepo := settingsRepo.NewRepository(db, logger)
	setService := settingsService.NewSettingsService(setRepo, logger)
	setHandler := settingsHandlers.NewSettingsHandlers(setService, logger)

	// Initialize Governance Domain
	gRepo := governanceRepo.NewGovernanceRepository(db)
	gService := governanceService.NewGovernanceService(gRepo, logger)
	gHandler := governanceHandlers.NewGovernanceHandler(gService)

	// Initialize DR Domain
	dRepo := drRepo.NewDrRepository(db, logger)
	dService := drService.NewDrService(dRepo, logger)
	dHandler := drHandlers.NewDrHandlers(dService, logger)

	// Initialize Organizations Domain
	oRepo := orgRepo.NewOrganizationRepository(db, logger)
	oService := orgService.NewOrganizationService(oRepo, logger)
	oHandler := orgHandlers.NewOrganizationHandlers(oService, logger)

	// Setup Chi router
	r := chi.NewRouter()
	
	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(sharedMiddleware.RequestIDMiddleware())
	r.Use(sharedMiddleware.LoggingMiddleware(logger))
	
	// Enable Auth Middleware
	r.Use(sharedMiddleware.AuthMiddleware(&envSecretProvider{secret: cfg.JWTSecret}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "admin-support-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API endpoints
	r.Route("/api/v1", func(r chi.Router) {
		sHandler.RegisterRoutes(r)
		setHandler.RegisterRoutes(r)
		gHandler.RegisterRoutes(r)
		dHandler.RegisterRoutes(r)
		oHandler.RegisterRoutes(r)
	})

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting admin-support service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down admin-support service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Admin-support service exited gracefully")
}
