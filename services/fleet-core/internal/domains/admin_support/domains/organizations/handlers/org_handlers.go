package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/models"
	"github.com/flyfleet/fleet-core/internal/domains/admin_support/domains/organizations/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DTO for create organization request (separate from models to prevent mass assignment)
type createOrganizationRequestDTO struct {
	Name             string `json:"name" validate:"required"`
	Domain           string `json:"domain"`
	DefaultTheme     string `json:"default_theme"`
	ContactEmail     string `json:"contact_email"`
	SubscriptionTier string `json:"subscription_tier"`
}

type OrganizationHandlers struct {
	service service.OrganizationService
	logger  *zap.Logger
}

func NewOrganizationHandlers(service service.OrganizationService, logger *zap.Logger) *OrganizationHandlers {
	return &OrganizationHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *OrganizationHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/organizations", func(r chi.Router) {
		r.Get("/", h.ListOrganizations)
		r.Get("/stats", h.GetOrganizationStats)
		r.Post("/", h.CreateOrganization)
		r.Delete("/{id}", h.DeleteOrganization)
	})

	r.Get("/dashboard/stats", h.GetDashboardStats)
	r.Get("/branches", h.ListBranches)
	r.Get("/cost-centers", h.ListCostCenters)
}

func (h *OrganizationHandlers) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 {
		limit = 10
	}
	search := query.Get("search")

	orgs, total, err := h.service.GetOrganizations(r.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("Failed to list organizations", zap.Error(err))
		h.errorJSON(w, "Failed to fetch organizations", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data": orgs,
		"meta": map[string]interface{}{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}, http.StatusOK)
}

func (h *OrganizationHandlers) GetOrganizationStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetOrganizationStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get organization stats", zap.Error(err))
		h.errorJSON(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": stats}, http.StatusOK)
}

func (h *OrganizationHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", zap.Error(err))
		h.errorJSON(w, "Failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": stats}, http.StatusOK)
}

func (h *OrganizationHandlers) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var dto createOrganizationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.errorJSON(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Map DTO to model
	req := &models.CreateOrganizationRequest{
		Name:             dto.Name,
		Domain:           dto.Domain,
		DefaultTheme:     dto.DefaultTheme,
		ContactEmail:     dto.ContactEmail,
		SubscriptionTier: dto.SubscriptionTier,
	}

	resp, err := h.service.CreateOrganization(r.Context(), *req)
	if err != nil {
		h.logger.Error("Failed to create organization", zap.Error(err))
		h.errorJSON(w, "Failed to create organization", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": resp}, http.StatusCreated)
}

func (h *OrganizationHandlers) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid organization ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteOrganization(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete organization", zap.String("id", idStr), zap.Error(err))
		h.errorJSON(w, "Failed to delete organization", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListBranches returns a list of branches for the organization
func (h *OrganizationHandlers) ListBranches(w http.ResponseWriter, r *http.Request) {
	branches, err := h.service.ListBranches(r.Context())
	if err != nil {
		h.logger.Error("Failed to list branches", zap.Error(err))
		h.errorJSON(w, "Failed to fetch branches", http.StatusInternalServerError)
		return
	}

	// Map to frontend expectation { value, label }
	var result []map[string]string
	for _, b := range branches {
		result = append(result, map[string]string{
			"value": b.ID.String(),
			"label": b.Name,
		})
	}

	h.json(w, map[string]interface{}{"data": result}, http.StatusOK)
}

// ListCostCenters returns a list of cost centers for the organization
func (h *OrganizationHandlers) ListCostCenters(w http.ResponseWriter, r *http.Request) {
	costCenters, err := h.service.ListCostCenters(r.Context())
	if err != nil {
		h.logger.Error("Failed to list cost centers", zap.Error(err))
		h.errorJSON(w, "Failed to fetch cost centers", http.StatusInternalServerError)
		return
	}

	// Map to frontend expectation { value, label }
	var result []map[string]string
	for _, c := range costCenters {
		result = append(result, map[string]string{
			"value": c.ID.String(),
			"label": c.Name,
		})
	}

	h.json(w, map[string]interface{}{"data": result}, http.StatusOK)
}

// Helpers
func (h *OrganizationHandlers) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *OrganizationHandlers) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}
