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
	settingsHandlers := handlers.NewSettingsHandlers(cfg, logger)
	integrationsHandlers := handlers.NewIntegrationsHandlers(cfg, logger)
	monitoringHandlers := handlers.NewMonitoringHandlers(cfg, logger)
	searchHandlers := handlers.NewSearchHandlers(cfg, logger)
	timelineHandlers := handlers.NewTimelineHandlers(cfg, logger)
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
	
	// Audit Compliance Extensions (Granular RBAC)
	mux.Handle("/api/audit/entity-history", middleware.RBACMiddleware(sp, security.ResourceEntityHistory)(http.HandlerFunc(mfaHandlers.ProxyAudit)))
	mux.Handle("/api/audit/evidence-exports", middleware.RBACMiddleware(sp, security.ResourceEvidenceExports)(http.HandlerFunc(mfaHandlers.ProxyAudit)))
	mux.Handle("/api/audit/evidence-exports/", middleware.RBACMiddleware(sp, security.ResourceEvidenceExports)(http.HandlerFunc(mfaHandlers.ProxyAudit)))
	mux.Handle("/api/audit/policy", middleware.RBACMiddleware(sp, security.ResourceAuditPolicy)(http.HandlerFunc(mfaHandlers.ProxyAudit)))
	mux.Handle("/api/audit/", middleware.RBACMiddleware(sp, security.ResourceAuditLogs)(http.HandlerFunc(mfaHandlers.ProxyAudit)))
	
	mux.Handle("/api/organizations/", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(mfaHandlers.ProxyOrganization)))
	mux.Handle("/api/rbac/", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(mfaHandlers.ProxyRBAC)))

	// Asset Management (RBAC Enforcement)
	mux.Handle("/api/v1/admin/vehicles/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(assetHandlers.ProxyVehicle)))
	mux.Handle("/api/v1/admin/devices/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(assetHandlers.ProxyDevice))) // Devices share vehicle resource for now
	mux.Handle("/api/v1/admin/drivers/", middleware.RBACMiddleware(sp, security.ResourceTrips)(http.HandlerFunc(assetHandlers.ProxyDriver)))   // Drivers share trips resource for now
	mux.Handle("/api/v1/admin/documents/", middleware.RBACMiddleware(sp, security.ResourceDocuments)(http.HandlerFunc(assetHandlers.ProxyDocument)))
	mux.Handle("/api/v1/admin/bulk-jobs/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(assetHandlers.ProxyBulkJobs))) // Bulk Jobs share vehicle resource
	mux.Handle("/api/v1/admin/trips/", middleware.RBACMiddleware(sp, security.ResourceTrips)(http.HandlerFunc(assetHandlers.ProxyTrips)))

	// Maintenance Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/work-orders/", middleware.RBACMiddleware(sp, security.ResourceMaintenance)(http.HandlerFunc(assetHandlers.ProxyWorkOrders)))
	mux.Handle("/api/v1/admin/inventory/", middleware.RBACMiddleware(sp, security.ResourceInventory)(http.HandlerFunc(assetHandlers.ProxyInventory)))
	mux.Handle("/api/v1/admin/maintenance-settings/", middleware.RBACMiddleware(sp, security.ResourceMaintenance)(http.HandlerFunc(assetHandlers.ProxyMaintenanceSettings)))

	// Fuel Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/fuel/", middleware.RBACMiddleware(sp, security.ResourceFuel)(http.HandlerFunc(assetHandlers.ProxyFuelSettings)))
	
	// Governance Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/governance/requests/", middleware.RBACMiddleware(sp, security.ResourceGovernance)(http.HandlerFunc(assetHandlers.ProxyGovernanceRequests)))
	mux.Handle("/api/v1/admin/governance/settings", middleware.RBACMiddleware(sp, security.ResourceGovernance)(http.HandlerFunc(assetHandlers.ProxyGovernanceSettings)))
	
	// Billing Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/billing/", middleware.RBACMiddleware(sp, security.ResourceBilling)(http.HandlerFunc(assetHandlers.ProxyBilling)))
	mux.Handle("/api/v1/admin/billing-settings/", middleware.RBACMiddleware(sp, security.ResourceBilling)(http.HandlerFunc(assetHandlers.ProxyBillingSettings)))
	mux.Handle("/api/v1/admin/billing-metrics/", middleware.RBACMiddleware(sp, security.ResourceBilling)(http.HandlerFunc(assetHandlers.ProxyBillingMetrics)))
	
	// Notification Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/notifications/", middleware.RBACMiddleware(sp, security.ResourceAlerts)(http.HandlerFunc(assetHandlers.ProxyNotification)))
	mux.Handle("/api/v1/admin/preferences/", middleware.RBACMiddleware(sp, security.ResourceNotificationsPreferences)(http.HandlerFunc(assetHandlers.ProxyNotificationPreferences)))
	mux.Handle("/api/v1/admin/preferences/global", middleware.RBACMiddleware(sp, security.ResourceNotificationsPreferences)(http.HandlerFunc(assetHandlers.ProxyNotificationGlobalDefaults)))

	// Reports Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/reports/", middleware.RBACMiddleware(sp, security.ResourceDocuments)(http.HandlerFunc(assetHandlers.ProxyReports)))
	
	// Disaster Recovery Module (RBAC Enforcement)
	mux.Handle("/api/v1/admin/dr/", middleware.RBACMiddleware(sp, security.ResourceDisasterRecovery)(http.HandlerFunc(assetHandlers.ProxyDR)))

	// Global Settings Module (RBAC Enforcement)
	mux.Handle("/api/settings/", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(settingsHandlers.ProxySettings)))
	
	// Security Operations Module (RBAC Enforcement + IP Firewall)
	securityHandlers := handlers.NewSecurityHandlers(cfg, logger)
	fm := middleware.NewFirewallManager(cfg, logger)
	fw := middleware.IPFirewallMiddleware(fm)

	mux.Handle("/api/security/ip-rules", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyIPRules))))
	mux.Handle("/api/security/ip-rules/", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyIPRules))))
	mux.Handle("/api/security/access-reviews", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyAccessReviews))))
	mux.Handle("/api/security/access-reviews/", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyAccessReviews))))
	mux.Handle("/api/security/sessions", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxySessions))))
	mux.Handle("/api/security/sessions/", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxySessions))))
	mux.Handle("/api/security/review-items/", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyReviewItems))))
	
	// New Expansion Routes
	mux.Handle("/api/security/logins", fw(middleware.RBACMiddleware(sp, security.ResourceAuditLogs)(http.HandlerFunc(securityHandlers.ProxyLoginActivityLogs))))
	mux.Handle("/api/security/logins/", fw(middleware.RBACMiddleware(sp, security.ResourceAuditLogs)(http.HandlerFunc(securityHandlers.ProxyLoginActivityLogs))))
	mux.Handle("/api/security/alerts", fw(middleware.RBACMiddleware(sp, security.ResourceAlerts)(http.HandlerFunc(securityHandlers.ProxyThreatAlerts))))
	mux.Handle("/api/security/alerts/", fw(middleware.RBACMiddleware(sp, security.ResourceAlerts)(http.HandlerFunc(securityHandlers.ProxyThreatAlerts))))
	mux.Handle("/api/security/mfa", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyMFACenter))))
	mux.Handle("/api/security/mfa/", fw(middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(securityHandlers.ProxyMFACenter))))
	
	// Integrations Management Module
	mux.Handle("/api/integrations/", middleware.RBACMiddleware(sp, security.ResourceIntegrations)(http.HandlerFunc(integrationsHandlers.ProxyIntegrations)))
	mux.Handle("/api/api-clients/", middleware.RBACMiddleware(sp, security.ResourceApiClients)(http.HandlerFunc(integrationsHandlers.ProxyApiClients)))
	mux.Handle("/api/webhooks/", middleware.RBACMiddleware(sp, security.ResourceWebhooks)(http.HandlerFunc(integrationsHandlers.ProxyWebhooks)))

	// Device Diagnostics Module
	deviceDiagnosticsHandlers := handlers.NewDeviceDiagnosticsHandler(cfg)
	mux.Handle("/api/devices/summary", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(deviceDiagnosticsHandlers.ProxyDeviceRequest)))
	mux.Handle("/api/devices/status", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(deviceDiagnosticsHandlers.ProxyDeviceRequest)))
	mux.Handle("/api/devices/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(deviceDiagnosticsHandlers.ProxyDeviceRequest)))
	mux.Handle("/api/firmware", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(deviceDiagnosticsHandlers.ProxyFirmwareRequest)))
	mux.Handle("/api/firmware/", middleware.RBACMiddleware(sp, security.ResourceVehicles)(http.HandlerFunc(deviceDiagnosticsHandlers.ProxyFirmwareRequest)))

	// Monitoring Module (Mock Data - Phase 1)
	// Using ResourceSystem for now, or fallback to ResourceOrganizations if System not available in security package
	// TODO: Create specific ResourceMonitoring constant
	mux.Handle("/api/v1/admin/monitoring/health", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(monitoringHandlers.GetSystemHealth)))
	mux.Handle("/api/v1/admin/monitoring/metrics", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(monitoringHandlers.GetMetrics)))
	mux.Handle("/api/v1/admin/monitoring/logs", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(monitoringHandlers.GetLogs)))
	mux.Handle("/api/v1/admin/monitoring/queues", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(monitoringHandlers.GetQueues)))
	mux.Handle("/api/v1/admin/monitoring/sla", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(monitoringHandlers.GetSLA)))

	// Global Search Module
	mux.Handle("/api/search", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(searchHandlers.GlobalSearch)))
	mux.Handle("/api/search/suggestions", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(searchHandlers.GetSearchSuggestions)))
	mux.Handle("/api/search/facets", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(searchHandlers.GetSearchFacets)))
	mux.Handle("/api/search/recent", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(searchHandlers.GetRecentSearches)))
	mux.Handle("/api/search/saved", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(searchHandlers.GetSavedSearches)))

	// Activity Timeline Module
	mux.Handle("/api/timeline", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(timelineHandlers.GetTimeline)))
	mux.Handle("/api/timeline/stats", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(timelineHandlers.GetTimelineStats)))
	mux.Handle("/api/timeline/filters", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(timelineHandlers.GetTimelineFilters)))

	// FleetSupport Module (New)
	supportHandlers := handlers.NewSupportHandlers(cfg, logger)
	mux.Handle("/api/support/dashboard/stats", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetDashboardStats)))
	mux.Handle("/api/support/dashboard/trends", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetDashboardTrends)))
	mux.Handle("/api/support/dashboard/recent-activity", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetRecentActivity)))

	mux.Handle("/api/support/users/profile", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetUserProfile)))
	mux.Handle("/api/support/users/sessions", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetUserSessions)))

	mux.Handle("/api/support/orgs/overview", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetOrgOverview)))

	mux.Handle("/api/support/incidents", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetIncidents)))
	mux.Handle("/api/support/incidents/create", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.CreateIncident)))
	mux.Handle("/api/support/incidents/detail", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetIncidentDetail)))
	mux.Handle("/api/support/incidents/comment", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.PostComment)))
	mux.Handle("/api/support/incidents/update", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.UpdateIncident)))
	mux.Handle("/api/support/logs", middleware.RBACMiddleware(sp, security.ResourceOrganizations)(http.HandlerFunc(supportHandlers.GetSystemLogs)))

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
