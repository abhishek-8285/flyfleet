package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/security"
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
	auditHandlers := handlers.NewAuditHandlers(authService)
	
	// Setup routes
	router := setupRoutes(authHandlers, adminHandlers, securityHandlers, auditHandlers, monitoringHandler, cfg, logger)
	
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

func setupRoutes(auth *handlers.AuthHandlers, admin *handlers.AdminHandlers, securityHandlers *handlers.SecurityHandlers, audit *handlers.AuditHandlers, monitor *handlers.MonitoringHandler, cfg *config.Config, logger *zap.Logger) *http.ServeMux {
	router := http.NewServeMux()
	
	// Middleware Chains
	adminChain := middleware.AdminMiddleware(logger)
	
	// Security Middleware & Policy Centralization
	limiter := middleware.NewRateLimiter(10, 1*time.Minute, logger)
	sp := &security.EnvSecretProvider{Key: cfg.JWTSecret}

	// Health check endpoint
	router.HandleFunc("/health", monitor.HealthCheck)
	
	// Monitoring endpoints
	router.HandleFunc("/metrics/all", monitor.GetAllMetrics)

	// Security Policy Version (for automated validation)
	router.HandleFunc("/health/security", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Manually constructing JSON to avoid importing encoding/json if not already used in this scope, 
		// but main.go imports usually include it. 
		// Checking imports... main.go imports: context, fmt, net/http, os, os/signal, syscall, time, ...
		// It does NOT import encoding/json effectively in the top block shown in previous step (it wasn't in list).
		// Wait, looking at step 1947, line 3-20: "encoding/json" is NOT in the import list.
		// So I should use fmt.Fprintf or add the import. Adding import is cleaner but multi-replace is hard with imports.
		// I will use fmt.Sprintf for simplicity as it's just a small JSON.
		fmt.Fprintf(w, `{"service": "auth-service", "security_policy_hash": "%s"}`, security.GetPolicyVersion())
	})
	
	// Enterprise Auth endpoints
	router.HandleFunc("/api/v1/auth/login", limiter.Limit(auth.Login))
	router.HandleFunc("/api/v1/auth/register", limiter.Limit(auth.Register))
	router.HandleFunc("/api/v1/user/theme", auth.UpdateTheme)
	router.HandleFunc("/api/v1/auth/refresh-token", auth.RefreshToken)

	// MFA Security APIs
	router.HandleFunc("/api/v1/security/mfa/dashboard", securityHandlers.GetMFADashboard)
	router.HandleFunc("/api/v1/security/mfa/adoption-trend", securityHandlers.GetMFAAdoptionTrend)
	router.HandleFunc("/api/v1/security/mfa/method-distribution", securityHandlers.GetMFAMethodDistribution)
	router.HandleFunc("/api/v1/security/mfa/events", securityHandlers.GetMFAEvents)
	router.HandleFunc("/api/v1/security/mfa/failures/locations", securityHandlers.GetTopFailureLocations)
	router.HandleFunc("/api/v1/security/mfa/failures/users", securityHandlers.GetTopFailureUsers)
	router.HandleFunc("/api/v1/security/mfa/locked-users", securityHandlers.GetLockedUsers)
	
	router.HandleFunc("/api/v1/security/mfa/unlock", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { securityHandlers.UnlockUser(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	router.HandleFunc("/api/v1/security/mfa/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { securityHandlers.ResetMFA(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/policy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { securityHandlers.GetMFAPolicy(w, r) } else if r.Method == http.MethodPut { securityHandlers.UpdateMFAPolicy(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/policy/history", securityHandlers.GetMFAPolicyHistory)
	router.HandleFunc("/api/v1/security/mfa/policy/rollback", securityHandlers.RollbackMFAPolicy)
	
	router.HandleFunc("/api/v1/security/mfa/overrides", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { securityHandlers.GetMFAOrgOverride(w, r) } else if r.Method == http.MethodPut { securityHandlers.UpdateMFAOrgOverride(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/exceptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { securityHandlers.GetMFAExceptions(w, r) } else if r.Method == http.MethodPost { securityHandlers.AddMFAException(w, r) } else { http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) }
	})
	
	router.HandleFunc("/api/v1/security/mfa/exceptions/", securityHandlers.DeleteMFAException) // Supports DELETE /api/v1/security/mfa/exceptions/{id}
	router.HandleFunc("/api/v1/security/mfa/stream", securityHandlers.StreamMFAEvents)

	// Admin Portal Endpoints (Standardized and Centralized)
	// Organizations
	router.Handle("/api/v1/admin/organizations", adminChain(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("id") != "" {
				admin.GetOrganization(w, r)
			} else {
				admin.ListOrganizations(w, r)
			}
		case http.MethodPost:
			admin.CreateOrganization(w, r)
		case http.MethodPut:
			admin.UpdateOrganization(w, r)
		case http.MethodDelete:
			admin.DeleteOrganization(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	// Users (Org Scoped)
	router.Handle("/api/v1/admin/users", adminChain(middleware.RBACMiddleware(sp, security.ResourceUsers)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListUsers(w, r)
		} else if r.Method == http.MethodDelete {
			admin.DeleteUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Vehicles (Org Scoped)
	router.Handle("/api/v1/admin/vehicles", adminChain(middleware.RBACMiddleware(sp, security.ResourceVehicles)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("id") != "" {
				admin.GetVehicle(w, r)
			} else {
				admin.ListVehicles(w, r)
			}
		case http.MethodPost:
			admin.CreateVehicle(w, r)
		case http.MethodPut:
			admin.UpdateVehicle(w, r)
		case http.MethodDelete:
			admin.DeleteVehicle(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Documents (Org Scoped)
	router.Handle("/api/v1/admin/documents", adminChain(middleware.RBACMiddleware(sp, security.ResourceDocuments)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Check if vehicle_id is present for List
			if r.URL.Query().Get("vehicle_id") != "" {
				admin.ListDocuments(w, r)
			} else {
				http.Error(w, "Missing vehicle_id", http.StatusBadRequest)
			}
		case http.MethodPost:
			admin.CreateDocument(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/documents/", adminChain(middleware.RBACMiddleware(sp, security.ResourceDocuments)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			admin.GetDocument(w, r)
		case http.MethodPut:
			admin.UpdateDocument(w, r)
		case http.MethodDelete:
			admin.DeleteDocument(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/vehicles/status", adminChain(middleware.RBACMiddleware(sp, security.ResourceVehicles)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			admin.UpdateVehicleStatus(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/vehicles/assign-driver", adminChain(middleware.RBACMiddleware(sp, security.ResourceVehicles)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			admin.AssignDriver(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))
	
	// Devices (Org Scoped)
	router.Handle("/api/v1/admin/devices", adminChain(middleware.RBACMiddleware(sp, security.ResourceVehicles)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListDevices(w, r)
		} else if r.Method == http.MethodPost {
			admin.CreateDevice(w, r)
		} else if r.Method == http.MethodDelete {
			admin.DeleteDevice(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/devices/assign", adminChain(middleware.RBACMiddleware(sp, security.ResourceVehicles)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			admin.AssignDevice(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/devices/unassign", adminChain(middleware.RBACMiddleware(sp, security.ResourceVehicles)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			admin.UnassignDevice(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Trips (Org Scoped)
	router.Handle("/api/v1/admin/trips", adminChain(middleware.RBACMiddleware(sp, security.ResourceTrips)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListTrips(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Alerts (Org Scoped)
	router.Handle("/api/v1/admin/alerts", adminChain(middleware.RBACMiddleware(sp, security.ResourceAlerts)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.ListAlerts(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/alerts/stats/summary", adminChain(middleware.RBACMiddleware(sp, security.ResourceAlerts)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.GetAlertSummary(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Analytics (Org Scoped)
	router.Handle("/api/v1/admin/analytics/drivers", adminChain(middleware.RBACMiddleware(sp, security.ResourceAnalytics)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.GetDriverStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/analytics/trips", adminChain(middleware.RBACMiddleware(sp, security.ResourceAnalytics)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			admin.GetTripStats(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// Audit Logs (Org Scoped)
	router.Handle("/api/v1/admin/audit/logs", adminChain(middleware.RBACMiddleware(sp, security.ResourceAuditLogs)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			audit.ListAuditLogs(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/audit/logs/", adminChain(middleware.RBACMiddleware(sp, security.ResourceAuditLogs)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Manual routing for /{id}
		if r.Method == http.MethodGet {
			audit.GetAuditLogByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/audit/settings", adminChain(middleware.RBACMiddleware(sp, security.ResourceAuditSettings)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			audit.GetAuditSettings(w, r)
		} else if r.Method == http.MethodPut {
			audit.UpdateAuditSettings(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))
	
	return router
}