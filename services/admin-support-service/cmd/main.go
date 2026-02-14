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

	"github.com/fleet-platform/shared/pkg/middleware"
)

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

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	
	// Enable Auth Middleware
	r.Use(middleware.AuthMiddleware(&envSecretProvider{secret: cfg.JWTSecret}))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "admin-support-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// API endpoints
	api := r.Group("/api/v1")
	{
		sHandler.RegisterRoutes(api)
		setHandler.RegisterRoutes(api)
		gHandler.RegisterRoutes(api)

		dHandler.RegisterRoutes(api)

		// Organization routes
		oHandler.RegisterRoutes(api)
	}

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
