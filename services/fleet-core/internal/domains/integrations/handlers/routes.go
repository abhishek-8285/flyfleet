package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type orgCtxKey string

const tenantIDKey orgCtxKey = "tenant_id"

func orgScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			tenantIDStr = r.Header.Get("X-Organization-ID")
		}
		if tenantIDStr == "" {
			tenantIDStr = r.URL.Query().Get("tenantId")
		}
		if tenantIDStr == "" {
			tenantIDStr = r.URL.Query().Get("orgId")
		}
		if tenantIDStr == "" {
			http.Error(w, "MISSING_TENANT_ID", http.StatusBadRequest)
			return
		}
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			http.Error(w, "INVALID_TENANT_ID", http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RegisterRoutes mounts integrations v2 API routes.
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/integrations", func(r chi.Router) {
		// JWT middleware should be applied globally; tenantID will come from validated context
		r.Get("/", h.ListIntegrations)
		r.Post("/", h.CreateIntegration)
		r.Get("/logs", h.ListLogs)
		r.Route("/{id}", func(r chi.Router) {
			r.Put("/", h.UpdateIntegration)
			r.Post("/test", h.TestConnection)
		})
	})

	r.Route("/api-clients", func(r chi.Router) {
		r.Get("/", h.ListAPIClients)
		r.Post("/", h.CreateAPIClient)
	})

	r.Route("/webhooks", func(r chi.Router) {
		r.Get("/", h.ListWebhooks)
		r.Post("/", h.CreateWebhook)
		r.Delete("/{id}", h.DeleteWebhook)
	})
}
