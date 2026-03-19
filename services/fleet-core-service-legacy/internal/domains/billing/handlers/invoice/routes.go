package invoice

import (
	"github.com/flyfleet/fleet-core-service/internal/domains/billing/middleware"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/api/v1/billing/invoices", func(r chi.Router) {
		r.Use(middleware.Tenant)

		r.Post("/", h.Create)
		r.Get("/", h.List)
		
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Post("/void", h.Void)
		})
	})
}

