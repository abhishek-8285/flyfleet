package maintenance_schedule

import (
	"github.com/flyfleet/maintenance-service/internal/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/api/v1/maintenance/schedules", func(r chi.Router) {
		r.Use(middleware.Tenant) // Ensure org_id isolation

		r.Post("/", h.Create)
		r.Get("/", h.List)
		
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
			r.Post("/complete", h.Complete)
		})
	})
}

