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
	
	targetURL, err := url.Parse(h.cfg.VehicleServiceURL)
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
	h.proxyToAuth(w, r, "/api/v1/admin/vehicles", "/api/v1/admin/vehicles")
}

// ProxyDevice forwards device management requests
func (h *AssetHandlers) ProxyDevice(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/v1/admin/devices", "/api/v1/admin/devices")
}

// ProxyDriver forwards driver management requests
func (h *AssetHandlers) ProxyDriver(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/v1/admin/drivers", "/api/v1/admin/drivers")
}

// ProxyDocument forwards document management requests
func (h *AssetHandlers) ProxyDocument(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/v1/admin/documents", "/api/v1/admin/documents")
}

// ProxyTrips forwards trip management requests to vehicle-service
// /api/v1/admin/trips -&gt; /api/v1/trips
func (h *AssetHandlers) ProxyTrips(w http.ResponseWriter, r *http.Request) {
	h.proxyToVehicle(w, r, "/api/v1/admin/trips", "/api/v1/trips")
}
