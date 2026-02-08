package main

import (
	"net/http"
	"testing"

	"github.com/flyfleet/web-bff/internal/config"
	"github.com/flyfleet/web-bff/internal/handlers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestBillingProxyRouteRegistration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{
		BillingServiceURL: "http://billing-service:8080",
	}

	assetHandlers := handlers.NewAssetHandlers(cfg, logger)
	mux := http.NewServeMux()

	// Mimic registration in main.go
	mux.HandleFunc("/api/v1/admin/billing/", assetHandlers.ProxyBilling)
	mux.HandleFunc("/api/v1/admin/billing-settings/", assetHandlers.ProxyBillingSettings)
	mux.HandleFunc("/api/v1/admin/billing-metrics/", assetHandlers.ProxyBillingMetrics)

	tests := []struct {
		name   string
		path   string
		method string
	}{
		{"Get Invoices", "/api/v1/admin/billing/invoices", "GET"},
		{"Get Settings", "/api/v1/admin/billing-settings", "GET"},
		{"Get Metrics", "/api/v1/admin/billing-metrics", "GET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			// We check if the mux finds a handler for the path
			_, pattern := mux.Handler(req)
			assert.NotEmpty(t, pattern, "Route should be registered")
		})
	}
}
