package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flyfleet/web-bff/internal/config"
	"go.uber.org/zap"
)

type SettingsHandlers struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewSettingsHandlers(cfg *config.Config, logger *zap.Logger) *SettingsHandlers {
	return &SettingsHandlers{cfg: cfg, logger: logger}
}

func (h *SettingsHandlers) ProxySettings(w http.ResponseWriter, r *http.Request) {
	// Source: /api/settings/* -> Target: /settings/*
	sourcePrefix := "/api/settings/"
	targetPrefix := "/settings/"
	
	path := r.URL.Path
	if !strings.HasPrefix(path, sourcePrefix) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	
	targetPath := strings.Replace(path, sourcePrefix, targetPrefix, 1)
	
	targetURL, err := url.Parse(h.cfg.SettingsServiceURL)
	if err != nil {
		h.logger.Error("Failed to parse settings service URL", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	targetURL.Path = targetPath
	targetURL.RawQuery = r.URL.RawQuery
	
	h.logger.Info("Proxying settings request", 
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
		h.logger.Error("Failed to call settings service", zap.Error(err))
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
