package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type CustomerHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewCustomerHandlers(cfg *config.Config, logger *zap.Logger) *CustomerHandlers {
	return &CustomerHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// ProxyCustomerPortal forwards customer portal requests to fleet-core-service
// /api/v1/customer/* -> /api/v1/customer/*
func (h *CustomerHandlers) ProxyCustomerPortal(w http.ResponseWriter, r *http.Request) {
	h.proxyToFleetCore(w, r, "/api/v1/customer", "/api/v1/customer")
}

// proxyToFleetCore is a helper to forward requests
func (h *CustomerHandlers) proxyToFleetCore(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
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
		h.logger.Error("Failed to parse fleet core service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying customer portal request", 
		zap.String("method", r.Method),
		zap.String("source", path),
		zap.String("target", targetURL.String()),
	)
	
	client := &http.Client{}
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
	
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call fleet core service", zap.Error(err))
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
