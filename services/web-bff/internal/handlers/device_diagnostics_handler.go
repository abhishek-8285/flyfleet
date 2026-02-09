package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"github.com/go-chi/chi/v5"
)

type DeviceDiagnosticsHandler struct {
	serviceURL string
	client     *http.Client
}

func NewDeviceDiagnosticsHandler(cfg *config.Config) *DeviceDiagnosticsHandler {
	return &DeviceDiagnosticsHandler{
		serviceURL: cfg.DeviceDiagnosticsServiceURL,
		client:     &http.Client{},
	}
}

// ProxyDeviceRequest forwards requests to device-diagnostics-service
func (h *DeviceDiagnosticsHandler) ProxyDeviceRequest(w http.ResponseWriter, r *http.Request) {
	// Build target URL
	path := strings.TrimPrefix(r.URL.Path, "/api")
	targetURL := h.serviceURL + "/api/v1" + path

	// Add query params
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Create proxy request
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Forward org ID and user ID from context
	if orgID := r.Header.Get("X-Organization-ID"); orgID != "" {
		proxyReq.Header.Set("X-Organization-ID", orgID)
	}
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		proxyReq.Header.Set("X-User-ID", userID)
	}

	// Execute request
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status and body
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// ProxyFirmwareRequest forwards firmware requests
func (h *DeviceDiagnosticsHandler) ProxyFirmwareRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	targetURL := h.serviceURL + "/api/v1" + path

	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	if orgID := r.Header.Get("X-Organization-ID"); orgID != "" {
		proxyReq.Header.Set("X-Organization-ID", orgID)
	}

	resp, err := h.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// RegisterRoutes adds device diagnostics routes to the router
func (h *DeviceDiagnosticsHandler) RegisterRoutes(r chi.Router) {
	// Device endpoints
	r.Get("/devices/summary", h.ProxyDeviceRequest)
	r.Get("/devices/status", h.ProxyDeviceRequest)
	r.Get("/devices/{id}", h.ProxyDeviceRequest)
	r.Get("/devices/{id}/sensors", h.ProxyDeviceRequest)
	r.Get("/devices/{id}/logs", h.ProxyDeviceRequest)
	r.Get("/devices/{id}/telemetry", h.ProxyDeviceRequest)
	r.Post("/devices/{id}/commands", h.ProxyDeviceRequest)
	r.Get("/devices/{id}/commands", h.ProxyDeviceRequest)

	// Firmware endpoints
	r.Get("/firmware", h.ProxyFirmwareRequest)
	r.Post("/firmware", h.ProxyFirmwareRequest)
	r.Get("/firmware/rollouts", h.ProxyFirmwareRequest)
	r.Post("/firmware/rollouts", h.ProxyFirmwareRequest)
}
