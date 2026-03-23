package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts settings endpoints on a chi router.
func (h *SettingsHandlers) RegisterRoutes(r chi.Router) {
	r.Get("/export", h.GetSettingsExport)
	r.Get("/audit", h.GetAuditLogs)
	r.Post("/import", h.ImportSettings)

	r.Route("/platform", func(r chi.Router) {
		r.Get("/", h.GetPlatformSettings)
		r.Put("/", h.UpdatePlatformSettings)
	})

	r.Route("/flags", func(r chi.Router) {
		r.Get("/", h.ListFeatureFlags)
		r.Post("/", h.CreateFeatureFlag)
		r.Patch("/{id}/toggle", func(w http.ResponseWriter, req *http.Request) {
			// Keep existing handler behavior by remapping URL path shape.
			id := chi.URLParam(req, "id")
			clone := req.Clone(req.Context())
			clone.URL.Path = "/settings/flags/" + strings.TrimSpace(id) + "/toggle"
			h.ToggleFeatureFlag(w, clone)
		})
	})

	r.Route("/localization", func(r chi.Router) {
		r.Get("/", h.GetLocalizationSettings)
		r.Put("/", h.UpdateLocalizationSettings)
	})

	r.Route("/templates", func(r chi.Router) {
		r.Get("/", h.ListTemplates)
		r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
			id := chi.URLParam(req, "id")
			clone := req.Clone(req.Context())
			clone.URL.Path = "/settings/templates/" + strings.TrimSpace(id)
			h.UpdateTemplate(w, clone)
		})
	})

	r.Route("/channels", func(r chi.Router) {
		r.Get("/", h.ListNotificationChannels)
		r.Put("/", h.UpdateNotificationChannel)
	})

	r.Route("/limits", func(r chi.Router) {
		r.Get("/", h.GetSystemLimits)
		r.Put("/", h.UpdateSystemLimits)
	})

	r.Post("/audit/cleanup", h.CleanupAuditLogs)
}
