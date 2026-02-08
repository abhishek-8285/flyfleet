package main

import (
	"net/http"
	"testing"

	"github.com/flyfleet/web-bff/internal/config"
	"github.com/flyfleet/web-bff/internal/handlers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNotificationProxyRouteRegistration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		NotificationServiceURL: "http://notification-service:8080",
	}

	assetHandlers := handlers.NewAssetHandlers(cfg, logger)
	mux := http.NewServeMux()

	// Mimic registration in main.go
	mux.HandleFunc("/api/v1/admin/notifications/", assetHandlers.ProxyNotification)

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{"List Notifications", "/api/v1/admin/notifications/", "GET"},
		{"Get Notification Detail", "/api/v1/admin/notifications/123", "GET"},
		{"Resend Notification", "/api/v1/admin/notifications/123/resend", "POST"},
		{"Get Settings", "/api/v1/admin/notifications/settings", "GET"},
		{"Update Settings", "/api/v1/admin/notifications/settings", "PUT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			// We check if the mux finds a handler for the path
			_, pattern := mux.Handler(req)
			assert.NotEmpty(t, pattern, "Route should be registered")
			assert.Contains(t, pattern, "/api/v1/admin/notifications/", "Pattern should match our registration")
		})
	}
}
