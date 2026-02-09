package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type IntegrationsHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewIntegrationsHandlers(cfg *config.Config, logger *zap.Logger) *IntegrationsHandlers {
	return &IntegrationsHandlers{cfg: cfg, logger: logger}
}

func (h *IntegrationsHandlers) ProxyIntegrations(w http.ResponseWriter, r *http.Request) {
	h.proxyToService(w, r, "/api/integrations/", "/api/v1/integrations/")
}

func (h *IntegrationsHandlers) ProxyApiClients(w http.ResponseWriter, r *http.Request) {
	h.proxyToService(w, r, "/api/api-clients/", "/api/v1/api-clients/")
}

func (h *IntegrationsHandlers) ProxyWebhooks(w http.ResponseWriter, r *http.Request) {
	h.proxyToService(w, r, "/api/webhooks/", "/api/v1/webhooks/")
}

func (h *IntegrationsHandlers) proxyToService(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)
	
	targetURL, err := url.Parse(h.cfg.IntegrationsServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse integrations service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying integration request", 
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
	
	// Copy headers (Authorization, X-Organization-ID, etc.)
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call integrations service", zap.Error(err))
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
