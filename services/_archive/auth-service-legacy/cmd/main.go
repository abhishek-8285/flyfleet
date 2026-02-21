package main

// @title Fleet Platform Auth Service API
// @version 1.0
// @description API for authentication, user management, and security
// @host localhost:8080
// @BasePath /

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fleet-platform/shared/pkg/security"
	"github.com/flyfleet/auth-service/internal/config"
	"github.com/flyfleet/auth-service/internal/database"
	"github.com/flyfleet/auth-service/internal/handlers"
	"github.com/flyfleet/auth-service/internal/middleware"
	"github.com/flyfleet/auth-service/internal/repository"
	"github.com/flyfleet/auth-service/internal/service"

	// _ "github.com/flyfleet/auth-service/docs"
	// httpSwagger "github.com/swaggo/http-swagger"
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
	authService := service.NewAuthService(repo, cfg.JWTSecret, cfg.FleetCoreServiceURL, cfg.AlertServiceURL, logger)
	
	// Seed RBAC Data
	if err := authService.EnsureSeedData(context.Background()); err != nil {
		logger.Error("Failed to seed RBAC data", zap.Error(err))
		// continue anyway, don't crash info
	}
	
	// Initialize CSRF protection
	csrfStore := middleware.NewCSRFStore(logger)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, logger)
	adminHandlers := handlers.NewAdminHandlers(authService, logger)
	monitoringHandler := handlers.NewMonitoringHandler(db, logger)
	securityHandlers := handlers.NewSecurityHandlers(authService, logger)
	auditHandlers := handlers.NewAuditHandlers(authService)
	rbacHandlers := handlers.NewRBACOperationsHandlers(authService, logger)

	// Setup routes
	router := setupRoutes(authService, authHandlers, adminHandlers, securityHandlers, rbacHandlers, auditHandlers, monitoringHandler, cfg, logger)
	
	// ... (rest of main unchanged) ...
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: middleware.SecurityHeadersMiddleware(middleware.CORSMiddleware(middleware.CSRFMiddleware(csrfStore)(router))),
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

func setupRoutes(authService *service.AuthService, auth *handlers.AuthHandlers, admin *handlers.AdminHandlers, securityHandlers *handlers.SecurityHandlers, rbac *handlers.RBACOperationsHandlers, audit *handlers.AuditHandlers, monitor *handlers.MonitoringHandler, cfg *config.Config, logger *zap.Logger) *http.ServeMux {
	router := http.NewServeMux()
	
	// Middleware Chains
	adminChain := middleware.AdminMiddleware(logger)
	
	// Security Middleware & Policy Centralization
	limiter := middleware.NewRateLimiter(10, 1*time.Minute, logger)
	authLimiter := middleware.NewAuthRateLimiter(5, 15*time.Minute, logger) // 5 attempts per 15 minutes for auth endpoints
	sp := &security.EnvSecretProvider{Key: cfg.JWTSecret}

	// Health check endpoint
	router.HandleFunc("/health", monitor.HealthCheck)
	
	// Swagger documentation
	// router.Handle("/swagger/", httpSwagger.WrapHandler)

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
	router.HandleFunc("/api/v1/auth/login", authLimiter.LimitLogin(auth.Login))  // Enhanced auth rate limiting
	router.HandleFunc("/api/v1/auth/register", limiter.Limit(auth.Register))  // Standard rate limiting for registration
	router.HandleFunc("/api/v1/user/theme", auth.UpdateTheme)
	router.HandleFunc("/api/v1/auth/refresh-token", auth.RefreshToken)
	
	// Profile Endpoint (Authenticated)
	router.Handle("/api/v1/auth/profile", middleware.AuthMiddleware(sp)(http.HandlerFunc(auth.GetProfile)))
	router.Handle("/api/v1/auth/profile/", middleware.AuthMiddleware(sp)(http.HandlerFunc(auth.GetProfile)))

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

	// Security Expansion Endpoints
	extHandlers := handlers.NewSecurityExtHandlers(authService, logger)
	router.HandleFunc("/api/v1/security/ip-rules/all", extHandlers.ListIPRules) // For BFF Firewall Manager
	router.HandleFunc("/api/v1/security/ip-rules", extHandlers.ListIPRules)
	router.HandleFunc("/api/v1/security/ip-rules/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { 
			extHandlers.CreateIPRule(w, r) 
		} else if r.Method == http.MethodDelete {
			extHandlers.DeleteIPRule(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	router.HandleFunc("/api/v1/security/logins", extHandlers.ListLoginEvents)
	router.HandleFunc("/api/v1/security/alerts", extHandlers.ListThreatAlerts)
	router.HandleFunc("/api/v1/security/alerts/", extHandlers.UpdateThreatAlertStatus) // /alerts/{id}/status handled inside
	router.HandleFunc("/api/v1/security/mfa/users", extHandlers.ListMFAUsers)
	router.HandleFunc("/api/v1/security/telemetry", extHandlers.GetSecurityTelemetry)

	// Access Reviews
	router.HandleFunc("/api/v1/security/access-reviews", extHandlers.ListAccessReviews)
	router.HandleFunc("/api/v1/security/access-reviews/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			extHandlers.GetAccessReview(w, r)
		} else if r.Method == http.MethodPost {
			extHandlers.CreateAccessReview(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/api/v1/security/access-reviews/status", extHandlers.UpdateAccessReviewStatus)
	router.HandleFunc("/api/v1/security/review-items", extHandlers.UpdateReviewItem)

	// Sessions
	router.HandleFunc("/api/v1/security/sessions", extHandlers.ListSessions)
	router.HandleFunc("DELETE /api/v1/security/sessions/{id}", extHandlers.RevokeSession)
	router.HandleFunc("DELETE /api/v1/security/sessions/revoke-all", extHandlers.RevokeAllSessions)

	// Security Configuration & Metrics
	router.HandleFunc("/api/v1/security/metrics", extHandlers.GetSecurityMetrics)
	router.HandleFunc("/api/v1/security/posture-score", extHandlers.GetSecurityPostureScore)
	router.HandleFunc("/api/v1/security/events", extHandlers.ListSecurityEvents)
	router.HandleFunc("/api/v1/security/compliance", extHandlers.GetSecurityComplianceReport)
	router.HandleFunc("/api/v1/security/vulnerabilities", extHandlers.GetVulnerabilityReport)
	router.HandleFunc("/api/v1/security/settings", extHandlers.GetSecuritySettings)
	router.HandleFunc("/api/v1/security/settings/update", extHandlers.ManageSecuritySettings)


	// Admin Portal Endpoints (Standardized and Centralized)
	// Organizations
	router.Handle("/api/v1/admin/organizations", adminChain(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(organizationsHandler(admin)))))
	router.Handle("/api/v1/admin/organizations/", adminChain(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(organizationsHandler(admin)))))

	// Users (Org Scoped)
	router.Handle("/api/v1/admin/users/stats", adminChain(middleware.RBACMiddleware(sp, security.ResourceUsers)(http.HandlerFunc(admin.GetStats))))

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
			admin.ListDocuments(w, r)
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

	router.Handle("/api/v1/admin/audit/entity-history", adminChain(middleware.RBACMiddleware(sp, security.ResourceEntityHistory)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			audit.ListEntityHistory(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/audit/evidence-exports", adminChain(middleware.RBACMiddleware(sp, security.ResourceEvidenceExports)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			audit.ListEvidenceExports(w, r)
		} else if r.Method == http.MethodPost {
			audit.CreateEvidenceExport(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/audit/evidence-exports/", adminChain(middleware.RBACMiddleware(sp, security.ResourceEvidenceExports)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			audit.GetEvidenceExport(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/audit/policy", adminChain(middleware.RBACMiddleware(sp, security.ResourceAuditPolicy)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			audit.GetAuditPolicy(w, r)
		} else if r.Method == http.MethodPut {
			audit.UpdateAuditPolicy(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	// RBAC Suite
	router.Handle("/api/v1/admin/rbac/roles", adminChain(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rbac.ListRoles(w, r)
		} else if r.Method == http.MethodPost {
			rbac.CreateRole(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))))

	router.Handle("/api/v1/admin/rbac/roles/", adminChain(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(middleware.OrgScopeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/permissions") {
			if r.Method == http.MethodGet {
				rbac.GetRolePermissions(w, r)
			} else if r.Method == http.MethodPut {
				rbac.UpdateRolePermissions(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(r.URL.Path, "/clone") {
			if r.Method == http.MethodPost {
				rbac.CloneRole(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			if r.Method == http.MethodPut {
				rbac.UpdateRole(w, r)
			} else if r.Method == http.MethodDelete {
				rbac.DeleteRole(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})))))

	router.Handle("/api/v1/admin/rbac/permissions", adminChain(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			rbac.ListPermissions(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	router.Handle("/api/v1/admin/rbac/users/assign", adminChain(middleware.RBACMiddleware(sp, security.ResourceUsers)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			rbac.AssignRole(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	router.Handle("/api/v1/admin/rbac/users/unassign", adminChain(middleware.RBACMiddleware(sp, security.ResourceUsers)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			rbac.UnassignRole(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))))

	router.Handle("/api/v1/admin/rbac/users/", adminChain(middleware.RBACMiddleware(sp, security.ResourceUsers)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/roles") {
			if r.Method == http.MethodGet {
				rbac.GetUserRoles(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasSuffix(r.URL.Path, "/permissions") {
			if r.Method == http.MethodGet {
				rbac.GetUserEffectivePermissions(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))))
	
	return router
}
func organizationsHandler(admin *handlers.AdminHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
