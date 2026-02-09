package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type MFAHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewMFAHandlers(cfg *config.Config, logger *zap.Logger) *MFAHandlers {
	return &MFAHandlers{cfg: cfg, logger: logger}
}

func (h *MFAHandlers) ProxyMFA(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/mfa/", "/api/v1/security/mfa/")
}

func (h *MFAHandlers) ProxyAudit(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/audit/", "/api/v1/admin/audit/")
}

func (h *MFAHandlers) ProxyOrganization(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/organizations/", "/api/v1/admin/organizations/")
}

func (h *MFAHandlers) ProxyRBAC(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/rbac/", "/api/v1/admin/rbac/")
}

func (h *MFAHandlers) proxyToAuth(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
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
	
	h.logger.Info("Proxying request", 
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
