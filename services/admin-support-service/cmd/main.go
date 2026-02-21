package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	google_grpc "google.golang.org/grpc"

	"github.com/flyfleet/admin-support-service/internal/config"
	admingrpc "github.com/flyfleet/admin-support-service/internal/grpc"
	pb "github.com/flyfleet/proto/admin/v1"
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

	emergencyHandlers "github.com/flyfleet/admin-support-service/internal/domains/emergency/handlers"
	emergencyService "github.com/flyfleet/admin-support-service/internal/domains/emergency/service"

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
	logger.Info("Connected to database")

	// NATS connection
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Warn("Failed to connect to NATS, emergency halt will be disabled", zap.Error(err))
	} else {
		defer nc.Close()
		logger.Info("Connected to NATS")
	}
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

	// Initialize Emergency Domain
	var eHandler *emergencyHandlers.EmergencyHandler
	if nc != nil {
		eService := emergencyService.NewEmergencyService(nc, logger)
		eHandler = emergencyHandlers.NewEmergencyHandler(eService, logger)
	}

	// Start gRPC Server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50056"
	}
	go func() {
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			logger.Fatal("Failed to listen for gRPC", zap.Error(err))
		}
		
		grpcServer := google_grpc.NewServer()
		adminServer := admingrpc.NewAdminSupportServer(sService, oService, logger)
		pb.RegisterAdminSupportServiceServer(grpcServer, adminServer)
		
		logger.Info("Starting admin-support gRPC server", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Setup Chi router
	r := chi.NewRouter()
	
	// Global Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(sharedMiddleware.RequestIDMiddleware())
	r.Use(sharedMiddleware.LoggingMiddleware(logger))
	
	// Public Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "admin-support-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// Authenticated API endpoints
	r.Group(func(r chi.Router) {
		r.Use(sharedMiddleware.AuthMiddleware(&envSecretProvider{secret: cfg.JWTSecret}))

		r.Route("/api/v1", func(r chi.Router) {
			sHandler.RegisterRoutes(r)
			setHandler.RegisterRoutes(r)
			gHandler.RegisterRoutes(r)
			dHandler.RegisterRoutes(r)
			oHandler.RegisterRoutes(r)
			if eHandler != nil {
				eHandler.RegisterRoutes(r)
			}
		})
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
