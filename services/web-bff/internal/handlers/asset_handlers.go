package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type AssetHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewAssetHandlers(cfg *config.Config, logger *zap.Logger) *AssetHandlers {
	return &AssetHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// proxyToAuth is a helper to forward requests to the auth-service
// We might want to move this to a shared middleware/utils if used often
func (h *AssetHandlers) proxyToAuth(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	
	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)
	
	targetURL, err := url.Parse(h.cfg.AuthServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse auth service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying asset request", 
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)
	
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call auth service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// proxyToVehicle is a helper to forward requests to the vehicle-service
func (h *AssetHandlers) proxyToVehicle(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	
	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Strip trailing slash if it exists
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}
	
	targetURL, err := url.Parse(h.cfg.FleetCoreServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse vehicle service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying vehicle request", 
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)
	
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call vehicle service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyVehicle forwards vehicle management requests
// Maps /api/v1/admin/vehicles/* -> /api/v1/vehicles/* (Mapping admin path to service path)
// WAIT: auth-service has /api/v1/admin/vehicles, vehicle-service has /api/v1/vehicles
// The existing ProxyVehicle maps /api/v1/admin/vehicles -> /api/v1/admin/vehicles which goes to auth-service?
// Let's check auth-service routes. Auth service might be proxying or handling admin stuff.
// If auth-service handles vehicles, then why do we have vehicle-service? 
// Maybe vehicle-service is new and we are migrating? 
// Or maybe Auth service handles authz logic?

// Looking at handlers.go in auth-service (from memory or previous view), it seemed to act as a gateway/policy point.
// But we are adding DIRECT proxy to vehicle-service for Bulk Jobs for now as per plan? 
// Plan said: "BFF: Register Bulk Upload proxy routes".
// Use vehicle-service for bulk uploads.

// ProxyBulkJobs forwards bulk job requests to vehicle-service
// /api/v1/admin/bulk-jobs -> /api/v1/bulk-jobs
func (h *AssetHandlers) ProxyBulkJobs(w http.ResponseWriter, r *http.Request) {
	h.proxyToVehicle(w, r, "/api/v1/admin/bulk-jobs", "/api/v1/bulk-jobs")
}

// ProxyVehicle forwards vehicle management requests
// Maps /api/v1/admin/vehicles/* -> /api/v1/admin/vehicles/* (Direct forwarding)
func (h *AssetHandlers) ProxyVehicle(w http.ResponseWriter, r *http.Request) {
	h.proxyToVehicle(w, r, "/api/v1/admin/vehicles", "/api/v1/vehicles")
}

// ProxyDevice forwards device management requests
func (h *AssetHandlers) ProxyDevice(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/v1/admin/devices", "/api/v1/admin/devices")
}

// ProxyDriver forwards driver management requests
func (h *AssetHandlers) ProxyDriver(w http.ResponseWriter, r *http.Request) {
	h.proxyToDriver(w, r, "/api/v1/admin/drivers", "/api/v1/drivers")
}

// proxyToDriver is a helper to forward requests to the driver-service
func (h *AssetHandlers) proxyToDriver(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	
	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Strip trailing slash if it exists
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}
	
	targetURL, err := url.Parse(h.cfg.FleetCoreServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse driver service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying driver request", 
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)
	
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call driver service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyDocument forwards document management requests
func (h *AssetHandlers) ProxyDocument(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/v1/admin/documents", "/api/v1/admin/documents")
}

// ProxyTrips forwards trip management requests to vehicle-service
// proxyToTrip is a helper to forward requests to the trip-service
func (h *AssetHandlers) proxyToTrip(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	
	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	targetURL, err := url.Parse(h.cfg.FleetCoreServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse trip service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying trip request", 
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)
	
	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call trip service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyTrips forwards trip management requests
func (h *AssetHandlers) ProxyTrips(w http.ResponseWriter, r *http.Request) {
	h.proxyToTrip(w, r, "/api/v1/admin/trips", "/api/v1/trips")
}

// proxyToMaintenance is a helper to forward requests to the maintenance-service
func (h *AssetHandlers) proxyToMaintenance(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Trim trailing slash to be consistent with core service expectations
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}

	targetURL, err := url.Parse(h.cfg.FleetCoreServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse maintenance service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying maintenance request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call maintenance service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyWorkOrders forwards work order requests to maintenance-service
// /api/v1/admin/work-orders/* -> /api/v1/work-orders/*
func (h *AssetHandlers) ProxyWorkOrders(w http.ResponseWriter, r *http.Request) {
	h.proxyToMaintenance(w, r, "/api/v1/admin/work-orders", "/api/v1/work-orders")
}

// ProxyInventory forwards inventory requests to maintenance-service
// /api/v1/admin/inventory/* -> /api/v1/inventory/*
func (h *AssetHandlers) ProxyInventory(w http.ResponseWriter, r *http.Request) {
	h.proxyToMaintenance(w, r, "/api/v1/admin/inventory", "/api/v1/inventory")
}

// ProxyMaintenanceSettings forwards settings requests to maintenance-service
// /api/v1/admin/maintenance-settings/* -> /api/v1/settings/*
func (h *AssetHandlers) ProxyMaintenanceSettings(w http.ResponseWriter, r *http.Request) {
	h.proxyToMaintenance(w, r, "/api/v1/admin/maintenance-settings", "/api/v1/settings")
}
// proxyToFuel is a helper to forward requests to the fuel-service
func (h *AssetHandlers) proxyToFuel(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Trim trailing slash
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}

	targetURL, err := url.Parse(h.cfg.FleetCoreServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse fuel service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying fuel request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call fuel service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyFuelLogs forwards fuel log requests to fuel-service
func (h *AssetHandlers) ProxyFuelLogs(w http.ResponseWriter, r *http.Request) {
	h.proxyToFuel(w, r, "/api/v1/admin/fuel-logs", "/api/v1/fuel-logs")
}

// ProxyFuelSettings forwards fuel settings requests to fuel-service
func (h *AssetHandlers) ProxyFuelSettings(w http.ResponseWriter, r *http.Request) {
	h.proxyToFuel(w, r, "/api/v1/admin/fuel", "/api/v1/fuel")
}

// proxyToGovernance is a helper to forward requests to the governance-service
func (h *AssetHandlers) proxyToGovernance(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Trim trailing slash
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}

	targetURL, err := url.Parse(h.cfg.AdminSupportServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse governance service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying governance request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call governance service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyGovernanceRequests forwards governance requests to governance-service
func (h *AssetHandlers) ProxyGovernanceRequests(w http.ResponseWriter, r *http.Request) {
	h.proxyToGovernance(w, r, "/api/v1/admin/governance/requests", "/api/v1/governance/requests")
}

// ProxyGovernanceSettings forwards governance settings requests to governance-service
func (h *AssetHandlers) ProxyGovernanceSettings(w http.ResponseWriter, r *http.Request) {
	h.proxyToGovernance(w, r, "/api/v1/admin/governance/settings", "/api/v1/governance/settings")
}
// proxyToBilling is a helper to forward requests to the billing-service
func (h *AssetHandlers) proxyToBilling(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Trim trailing slash
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}

	targetURL, err := url.Parse(h.cfg.BillingServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse billing service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying billing request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call billing service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyBilling forwards invoice requests to billing-service
func (h *AssetHandlers) ProxyBilling(w http.ResponseWriter, r *http.Request) {
	h.proxyToBilling(w, r, "/api/v1/admin/billing", "/api/v1/billing")
}

// ProxyBillingSettings forwards settings requests to billing-service
func (h *AssetHandlers) ProxyBillingSettings(w http.ResponseWriter, r *http.Request) {
	h.proxyToBilling(w, r, "/api/v1/admin/billing-settings", "/api/v1/billing/settings")
}

// ProxyBillingMetrics forwards metrics requests to billing-service
func (h *AssetHandlers) ProxyBillingMetrics(w http.ResponseWriter, r *http.Request) {
	h.proxyToBilling(w, r, "/api/v1/admin/billing-metrics", "/api/v1/billing/metrics")
}
// proxyToNotification is a helper to forward requests to the notification-service
func (h *AssetHandlers) proxyToNotification(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	targetURL, err := url.Parse(h.cfg.AlertServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse notification service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying notification request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call notification service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyNotification forwards notification management requests
func (h *AssetHandlers) ProxyNotification(w http.ResponseWriter, r *http.Request) {
	h.proxyToNotification(w, r, "/api/v1/admin/notifications", "/api/v1/admin/notifications")
}

// ProxyNotificationPreferences forwards notification preference requests
func (h *AssetHandlers) ProxyNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	h.proxyToNotification(w, r, "/api/v1/admin/preferences", "/api/v1/admin/notifications/preferences")
}

// ProxyNotificationGlobalDefaults forwards global notification default requests
func (h *AssetHandlers) ProxyNotificationGlobalDefaults(w http.ResponseWriter, r *http.Request) {
	h.proxyToNotification(w, r, "/api/v1/admin/preferences/global", "/api/v1/admin/notifications/preferences/global")
}

// proxyToReports is a helper to forward requests to the reports-service
func (h *AssetHandlers) proxyToReports(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	targetURL, err := url.Parse(h.cfg.ReportsServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse reports service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying reports request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call reports service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyReports forwards reports management requests
func (h *AssetHandlers) ProxyReports(w http.ResponseWriter, r *http.Request) {
	h.proxyToReports(w, r, "/api/v1/admin/reports", "/api/v1/reports")
}

// proxyToDR is a helper to forward requests to the dr-service
func (h *AssetHandlers) proxyToDR(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	targetURL, err := url.Parse(h.cfg.DRServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse DR service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying DR request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call DR service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyDR forwards disaster recovery requests
func (h *AssetHandlers) ProxyDR(w http.ResponseWriter, r *http.Request) {
	h.proxyToDR(w, r, "/api/v1/admin/dr", "/api/v1/admin/dr")
}

// proxyToAlert is a helper to forward requests to the alert-service
func (h *AssetHandlers) proxyToAlert(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	targetURL, err := url.Parse(h.cfg.AlertServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse alert service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying alert request", 
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call alert service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyAlerts forwards alert management requests
func (h *AssetHandlers) ProxyAlerts(w http.ResponseWriter, r *http.Request) {
	h.proxyToAlert(w, r, "/api/v1/admin/alerts", "/api/v1/alerts")
}


// proxyToVideo is a helper to forward requests to the video service (part of fleet-core)
func (h *AssetHandlers) proxyToVideo(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Trim trailing slash
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}

	targetURL, err := url.Parse(h.cfg.FleetCoreServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse fleet core service URL for video", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying video request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call fleet core service for video", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyVideo forwards video management requests to fleet-core-service
// /api/v1/admin/video/* -> /api/v1/video/*
func (h *AssetHandlers) ProxyVideo(w http.ResponseWriter, r *http.Request) {
	h.proxyToVideo(w, r, "/api/v1/admin/video", "/api/v1/video")
}

// ProxyAnalytics forwards analytics requests to fleet-core-service
// /api/v1/admin/analytics/* -> /api/v1/analytics/*
func (h *AssetHandlers) ProxyAnalytics(w http.ResponseWriter, r *http.Request) {
	h.proxyToVehicle(w, r, "/api/v1/admin/analytics", "/api/v1/analytics")
}

// ProxyRouteOptimization forwards route optimization requests to fleet-core-service
// /api/v1/routes/* -> /api/v1/routes/* (Direct proxy, no path rewriting needed if prefix matches)
func (h *AssetHandlers) ProxyRouteOptimization(w http.ResponseWriter, r *http.Request) {
    // Using proxyToVehicle as it targets fleet-core-service
    // Source: /api/v1/routes, Target: /api/v1/routes
	h.proxyToVehicle(w, r, "/api/v1/routes", "/api/v1/routes")
}

// proxyToIndia is a helper to forward requests to the india-integration-service
func (h *AssetHandlers) proxyToIndia(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)

	// Trim trailing slash
	if strings.HasSuffix(targetPath, "/") && len(targetPath) > 1 {
		targetPath = strings.TrimSuffix(targetPath, "/")
	}

	targetURL, err := url.Parse(h.cfg.IndiaIntegrationServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse india integration service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying india integration request",
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		h.logger.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call india integration service", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyFASTag forwards FASTag requests
func (h *AssetHandlers) ProxyFASTag(w http.ResponseWriter, r *http.Request) {
	h.proxyToIndia(w, r, "/api/v1/india/fastag", "/api/fastag")
}

// ProxyEWayBill forwards E-Way Bill requests
func (h *AssetHandlers) ProxyEWayBill(w http.ResponseWriter, r *http.Request) {
	h.proxyToIndia(w, r, "/api/v1/india/ewaybill", "/api/ewaybill")
}

// ProxyVahan forwards Vahan requests
func (h *AssetHandlers) ProxyVahan(w http.ResponseWriter, r *http.Request) {
	h.proxyToIndia(w, r, "/api/v1/india/vahan", "/api/vahan")
}

// ProxySarathi forwards Sarathi requests
func (h *AssetHandlers) ProxySarathi(w http.ResponseWriter, r *http.Request) {
	h.proxyToIndia(w, r, "/api/v1/india/sarathi", "/api/sarathi")
}

// ProxyGST forwards GST requests
func (h *AssetHandlers) ProxyGST(w http.ResponseWriter, r *http.Request) {
	h.proxyToIndia(w, r, "/api/v1/india/gst", "/api/gst")
}
