package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type TrackingHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewTrackingHandlers(cfg *config.Config, logger *zap.Logger) *TrackingHandlers {
	return &TrackingHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// ProxyTelemetry proxies telemetry ingestion requests to the tracking service
// POST /api/v1/tracking/telemetry/gps -> POST /api/v1/device/telemetry/gps
// POST /api/v1/tracking/telemetry/obd -> POST /api/v1/device/telemetry/obd
func (h *TrackingHandlers) ProxyTelemetry(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// Map external paths to internal paths
	targetPath := path
	if strings.HasPrefix(path, "/api/v1/tracking/telemetry/gps") {
		targetPath = "/api/v1/device/telemetry/gps"
	} else if strings.HasPrefix(path, "/api/v1/tracking/telemetry/obd") {
		targetPath = "/api/v1/device/telemetry/obd"
	} else if strings.HasPrefix(path, "/api/v1/tracking/telemetry/fuel") {
		// Mapping legacy fuel telemetry path if needed, or new standard
		targetPath = "/telemetry/fuel"
	}

	h.proxyRequest(w, r, targetPath)
}

// ProxyVehicleHistory proxies vehicle history requests
// GET /api/v1/tracking/vehicles/{id}/history -> GET /api/v1/vehicles/{id}/history
func (h *TrackingHandlers) ProxyVehicleHistory(w http.ResponseWriter, r *http.Request) {
	// Extract the ID from the URL using standard string manipulation since we don't have chi/gorilla here in the handler
	// path: /api/v1/tracking/vehicles/{id}/history
	path := r.URL.Path
	parts := strings.Split(path, "/")
	// Expected: ["", "api", "v1", "tracking", "vehicles", "{id}", "history"]
	
	if len(parts) >= 7 && parts[4] == "vehicles" && parts[6] == "history" {
		id := parts[5]
		targetPath := fmt.Sprintf("/api/v1/vehicles/%s/history", id)
		
		// Append query parameters
		if r.URL.RawQuery != "" {
			targetPath += "?" + r.URL.RawQuery
		}
		
		h.proxyRequest(w, r, targetPath)
		return
	}

	// Fallback for direct proxy if structure matches
	targetPath := strings.Replace(path, "/api/v1/tracking", "/api/v1", 1)
	h.proxyRequest(w, r, targetPath)
}

// ProxyDeviceCommand proxies device command requests (forwarding to tracking for execution if needed)
// Uses fleet-core-service for commands usually, but checking if tracking needs direct access
func (h *TrackingHandlers) proxyRequest(w http.ResponseWriter, r *http.Request, targetPath string) {
	targetURL, err := url.Parse(h.cfg.TrackingServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse tracking service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying tracking request",
		zap.String("method", r.Method),
		zap.String("source", r.URL.Path),
		zap.String("target", targetURL.String()),
	)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	
	// Update director to set correct host and headers
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host
		req.URL.Path = targetPath
		// Maintain auth headers
	}
	
	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		h.logger.Error("Tracking proxy error", zap.Error(err))
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}
