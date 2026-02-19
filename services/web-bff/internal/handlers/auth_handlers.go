package handlers

import (
	"io"
	"net/http"
	"net/url"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

// AuthHandlers handles authentication-related proxying
type AuthHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewAuthHandlers(cfg *config.Config, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{
		cfg:    cfg,
		logger: logger,
	}
}

// ProxyLogin proxies /api/v1/auth/login to auth-service
func (h *AuthHandlers) ProxyLogin(w http.ResponseWriter, r *http.Request) {
	targetURL, err := url.Parse(h.cfg.AuthServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse auth service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	targetURL.Path = "/api/v1/auth/login"
	targetURL.RawQuery = r.URL.RawQuery

	h.logger.Info("Proxying login request",
		zap.String("method", r.Method),
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
