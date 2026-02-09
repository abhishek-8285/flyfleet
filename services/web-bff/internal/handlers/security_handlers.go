package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

// SecurityHandlers handles proxying for security endpoints:
// - IP Rules
// - Access Reviews
// - Sessions
type SecurityHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewSecurityHandlers(cfg *config.Config, logger *zap.Logger) *SecurityHandlers {
	return &SecurityHandlers{cfg: cfg, logger: logger}
}

// ProxyIPRules proxies /api/security/ip-rules/* to auth-service
func (h *SecurityHandlers) ProxyIPRules(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/ip-rules", "/api/v1/security/ip-rules")
}

// ProxyAccessReviews proxies /api/security/access-reviews/* to auth-service
func (h *SecurityHandlers) ProxyAccessReviews(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/access-reviews", "/api/v1/security/access-reviews")
}

// ProxySessions proxies /api/security/sessions/* to auth-service
func (h *SecurityHandlers) ProxySessions(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/sessions", "/api/v1/security/sessions")
}

// ProxyReviewItems proxies /api/security/review-items/* to auth-service
func (h *SecurityHandlers) ProxyReviewItems(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/review-items", "/api/v1/security/review-items")
}

// ProxyLoginActivityLogs proxies /api/security/logins/* to auth-service
func (h *SecurityHandlers) ProxyLoginActivityLogs(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/logins", "/api/v1/security/logins")
}

// ProxyThreatAlerts proxies /api/security/alerts/* to auth-service
func (h *SecurityHandlers) ProxyThreatAlerts(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/alerts", "/api/v1/security/alerts")
}

// ProxyMFACenter proxies /api/security/mfa/* to auth-service
func (h *SecurityHandlers) ProxyMFACenter(w http.ResponseWriter, r *http.Request) {
	h.proxyToAuth(w, r, "/api/security/mfa", "/api/v1/security/mfa")
}

func (h *SecurityHandlers) proxyToAuth(w http.ResponseWriter, r *http.Request, sourcePrefix, targetPrefix string) {
	path := r.URL.Path
	
	// Handle both exact match and subpath match
	targetPath := path
	if strings.HasPrefix(path, sourcePrefix) {
		suffix := strings.TrimPrefix(path, sourcePrefix)
		targetPath = targetPrefix + suffix
	} else {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	
	targetURL, err := url.Parse(h.cfg.AuthServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse auth service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying security request", 
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
