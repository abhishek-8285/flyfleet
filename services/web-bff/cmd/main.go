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

	"github.com/fleet-platform/shared/pkg/security"
	"github.com/flyfleet/web-bff/internal/config"
	"github.com/flyfleet/web-bff/internal/handlers"
	"github.com/flyfleet/web-bff/internal/middleware"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()
	
	// Load configuration
	cfg := config.Load()
	
	// Initialize handlers
	mfaHandlers := handlers.NewMFAHandlers(cfg, logger)
	assetHandlers := handlers.NewAssetHandlers(cfg, logger)
	sp := &security.EnvSecretProvider{Key: cfg.JWTSecret}
	
	// Setup routes
	mux := http.NewServeMux()
	
	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Security Policy Version (for automated validation)
	mux.HandleFunc("/health/security", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service":              "web-bff",
			"security_policy_hash": security.GetPolicyVersion(),
		})
	})
	
	// Security endpoints (RBAC Enforcement - Centralized)
	mux.Handle("/api/mfa/", middleware.RBACMiddleware(sp, security.ResourceAlerts)(http.HandlerFunc(mfaHandlers.ProxyMFA)))
	mux.Handle("/api/audit/", middleware.RBACMiddleware(sp, security.ResourceAuditLogs)(http.HandlerFunc(mfaHandlers.ProxyAudit)))
	mux.Handle("/api/organizations/", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(mfaHandlers.ProxyOrganization)))

	// Asset Management (RBAC Enforcement)
	mux.Handle("/api/v1/admin/vehicles/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(assetHandlers.ProxyVehicle)))
	mux.Handle("/api/v1/admin/devices/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(assetHandlers.ProxyDevice))) // Devices share vehicle resource for now
	mux.Handle("/api/v1/admin/drivers/", middleware.RBACMiddleware(sp, security.ResourceTrips)(http.HandlerFunc(assetHandlers.ProxyDriver)))   // Drivers share trips resource for now
	mux.Handle("/api/v1/admin/documents/", middleware.RBACMiddleware(sp, security.ResourceDocuments)(http.HandlerFunc(assetHandlers.ProxyDocument)))
	mux.Handle("/api/v1/admin/bulk-jobs/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(assetHandlers.ProxyBulkJobs))) // Bulk Jobs share vehicle resource
	mux.Handle("/api/v1/admin/trips/", middleware.RBACMiddleware(sp, security.ResourceTrips)(http.HandlerFunc(assetHandlers.ProxyTrips)))
	
	// Server configuration
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: middleware.CORSMiddleware(mux),
	}
	
	// Run server in goroutine
	go func() {
		logger.Info("Starting Web BFF service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down Web BFF service...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	
	logger.Info("Web BFF service exited")
}
