package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts governance endpoints under /api/v1/governance.
func (h *GovernanceHandler) RegisterRoutes(r chi.Router, rateLimiter func(http.Handler) http.Handler) {
	r.Route("/requests", func(r chi.Router) {
		if rateLimiter != nil {
			r.With(rateLimiter).Post("/", h.CreateRequest)
		} else {
			r.Post("/", h.CreateRequest)
		}
		r.Get("/", h.ListRequests)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetRequest)
			r.Post("/approve", h.ApproveRequest)
			r.Post("/reject", h.RejectRequest)
			r.Post("/verify", h.VerifyRequest)
			r.Get("/download-link", h.GetDownloadLink)
		})
	})

	r.Get("/download/{token}", h.DownloadFile)
	r.Get("/reports/access-review", h.GenerateAccessReview)
	r.Get("/settings", h.GetSettings)
	r.Put("/settings", h.UpdateSettings)
}
