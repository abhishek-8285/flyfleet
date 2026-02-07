package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyfleet/auth-service/internal/config"
	"github.com/flyfleet/auth-service/internal/database"
	"github.com/flyfleet/auth-service/internal/handlers"
	"github.com/flyfleet/auth-service/internal/middleware"
	"github.com/flyfleet/auth-service/internal/repository"
	"github.com/flyfleet/auth-service/internal/service"
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
	
	// Initialize database connection
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	
	// Initialize repository
	repo := repository.NewRepository(db, logger)
	
	// Initialize service
	authService := service.NewAuthService(repo, cfg.JWTSecret, cfg.DriverServiceURL, cfg.VehicleServiceURL, cfg.TripServiceURL, cfg.AlertServiceURL, logger)
	
	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, logger)
	adminHandlers := handlers.NewAdminHandlers(authService, logger)
	monitoringHandler := handlers.NewMonitoringHandler(db, logger)
	securityHandlers := handlers.NewSecurityHandlers(authService, logger)
	
	// Setup routes
	router := setupRoutes(authHandlers, adminHandlers, securityHandlers, monitoringHandler, logger)
	
	// ... (rest of main unchanged) ...
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: middleware.CORSMiddleware(router),
	}
	
	// Run server in goroutine
	go func() {
		logger.Info("Starting enterprise auth service", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down auth service...")
	
	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	
	logger.Info("Auth service exited")
}

func setupRoutes(auth *handlers.AuthHandlers, admin *handlers.AdminHandlers, security *handlers.SecurityHandlers, monitor *handlers.MonitoringHandler, logger *zap.Logger) *http.ServeMux {
	router := http.NewServeMux()
	
	// Middleware Chains
	adminChain := middleware.AdminMiddleware(logger)
	
	// Security Middleware
	limiter := middleware.NewRateLimiter(10, 1*time.Minute, logger)

	// Health check endpoint
	router.HandleFunc("/health", monitor.HealthCheck)
	
	// Monitoring endpoints
	router.HandleFunc("/metrics/all", monitor.GetAllMetrics)
	
	// Enterprise Auth endpoints
	router.HandleFunc("/api/v1/auth/login", limiter.Limit(auth.Login))
	router.HandleFunc("/api/v1/auth/register", limiter.Limit(auth.Register))
	router.HandleFunc("/api/v1/user/theme", auth.UpdateTheme)
	router.HandleFunc("/api/v1/auth/refresh-token", auth.RefreshToken)

	// MFA Security APIs
	router.HandleFunc("/api/v1/security/mfa/dashboard", security.GetMFADashboard)
	router.HandleFunc("/api/v1/security/mfa/adoption-trend", security.GetMFAAdoptionTrend)
	router.HandleFunc("/api/v1/security/mfa/method-distribution", security.GetMFAMethodDistribution)
	router.HandleFunc("/api/v1/security/mfa/events", security.GetMFAEvents)
	router.HandleFunc("/api/v1/security/mfa/failures/locations", security.GetTopFailureLocations)
	router.HandleFunc("/api/v1/security/mfa/failures/users", security.GetTopFailureUsers)
	router.HandleFunc("/api/v1/security/mfa/locked-users", security.GetLockedUsers)
	
	router.HandleFunc("/api/v1/security/mfa/unlock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { security.UnlockUser(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	router.HandleFunc("/api/v1/security/mfa/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { security.ResetMFA(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/policy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { security.GetMFAPolicy(w, r) } else if r.Method == http.MethodPut { security.UpdateMFAPolicy(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/exceptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { security.GetMFAExceptions(w, r) } else if r.Method == http.MethodPost { security.AddMFAException(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/exceptions/", security.DeleteMFAException) // Supports DELETE /api/v1/security/mfa/exceptions/{id}
	router.HandleFunc("/api/v1/security/mfa/stream", security.StreamMFAEvents)

	// Admin Portal Endpoints (Standardized)
	// Organizations
	router.Handle("/api/v1/admin/organizations", adminChain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListOrganizations(w, r)
		} else if r.Method == http.MethodDelete {
			admin.DeleteOrganization(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Users (Org Scoped)
	router.Handle("/api/v1/admin/users", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListUsers(w, r)
		} else if r.Method == http.MethodDelete {
			admin.DeleteUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Vehicles (Org Scoped)
	router.Handle("/api/v1/admin/vehicles", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListVehicles(w, r)
		} else if r.Method == http.MethodPost {
			admin.CreateVehicle(w, r)
		} else if r.Method == http.MethodDelete {
			admin.DeleteVehicle(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Trips (Org Scoped)
	router.Handle("/api/v1/admin/trips", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListTrips(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Alerts (Org Scoped)
	router.Handle("/api/v1/admin/alerts", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListAlerts(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	router.Handle("/api/v1/admin/alerts/stats/summary", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.GetAlertSummary(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Analytics (Org Scoped)
	router.Handle("/api/v1/admin/analytics/drivers", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.GetDriverStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	router.Handle("/api/v1/admin/analytics/trips", adminChain(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.GetTripStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))
	
	return router
}