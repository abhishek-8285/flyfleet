package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flyfleet/fleet-core/internal/platform/iam/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type IAMHandlers struct {
	svc service.IAMService
}

func NewIAMHandlers(svc service.IAMService) *IAMHandlers {
	return &IAMHandlers{svc: svc}
}

func (h *IAMHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string    `json:"name"`
		Email    string    `json:"email"`
		Password string    `json:"password"`
		TenantID    uuid.UUID `json:"organization_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Name, req.TenantID)
	if err != nil {
		http.Error(w, "registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *IAMHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *IAMHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
	})
}
