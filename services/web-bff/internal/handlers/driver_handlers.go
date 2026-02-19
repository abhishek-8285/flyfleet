package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type DriverHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewDriverHandlers(cfg *config.Config, logger *zap.Logger) *DriverHandlers {
	return &DriverHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// proxyToDriver is a helper to forward requests to the fleet-core-service
func (h *DriverHandlers) proxyToDriver(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
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
	
	h.logger.Info("Proxying driver mobile request", 
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

// ProxyDriverMobile forwards driver mobile app requests
// /api/v1/driver/* -> /api/v1/drivers/* (Mapping singular driver to plural drivers service endpoint)
func (h *DriverHandlers) ProxyMobileDriver(w http.ResponseWriter, r *http.Request) {
	h.proxyToDriver(w, r, "/api/v1/driver", "/api/v1/drivers")
}
