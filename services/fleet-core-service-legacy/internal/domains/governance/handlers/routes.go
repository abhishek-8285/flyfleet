package handlers

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts governance endpoints under /api/v1/governance.
func (h *GovernanceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/requests", func(r chi.Router) {
		r.Post("/", h.CreateRequest)
		r.Get("/", h.ListRequests)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetRequest)
			r.Post("/approve", h.ApproveRequest)
			r.Post("/reject", h.RejectRequest)
		})
	})

	r.Get("/settings", h.GetSettings)
	r.Put("/settings", h.UpdateSettings)
}
