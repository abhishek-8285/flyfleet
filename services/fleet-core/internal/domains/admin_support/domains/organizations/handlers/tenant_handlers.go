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

// DTO for create tenant request (separate from models to prevent mass assignment)
type createTenantRequestDTO struct {
	Name             string `json:"name" validate:"required"`
	Domain           string `json:"domain"`
	DefaultTheme     string `json:"default_theme"`
	ContactEmail     string `json:"contact_email"`
	SubscriptionTier string `json:"subscription_tier"`
}

type TenantHandlers struct {
	service service.TenantService
	logger  *zap.Logger
}

func NewTenantHandlers(service service.TenantService, logger *zap.Logger) *TenantHandlers {
	return &TenantHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *TenantHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/tenants", func(r chi.Router) {
		r.Get("/", h.ListTenants)
		r.Get("/stats", h.GetTenantStats)
		r.Post("/", h.CreateTenant)
		r.Delete("/{id}", h.DeleteTenant)
	})

	r.Get("/dashboard/stats", h.GetDashboardStats)
	r.Get("/branches", h.ListBranches)
	r.Get("/cost-centers", h.ListCostCenters)
}

func (h *TenantHandlers) ListTenants(w http.ResponseWriter, r *http.Request) {
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

	tenants, total, err := h.service.GetTenants(r.Context(), page, limit, search)
	if err != nil {
		h.logger.Error("Failed to list tenants", zap.Error(err))
		h.errorJSON(w, "Failed to fetch tenants", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{
		"data": tenants,
		"meta": map[string]interface{}{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}, http.StatusOK)
}

func (h *TenantHandlers) GetTenantStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetTenantStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get tenant stats", zap.Error(err))
		h.errorJSON(w, "Failed to fetch stats", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": stats}, http.StatusOK)
}

func (h *TenantHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetDashboardStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get dashboard stats", zap.Error(err))
		h.errorJSON(w, "Failed to fetch dashboard stats", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": stats}, http.StatusOK)
}

func (h *TenantHandlers) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var dto createTenantRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		h.errorJSON(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Map DTO to model
	req := &models.CreateTenantRequest{
		Name:             dto.Name,
		Domain:           dto.Domain,
		DefaultTheme:     dto.DefaultTheme,
		ContactEmail:     dto.ContactEmail,
		SubscriptionTier: dto.SubscriptionTier,
	}

	resp, err := h.service.CreateTenant(r.Context(), *req)
	if err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		h.errorJSON(w, "Failed to create tenant", http.StatusInternalServerError)
		return
	}

	h.json(w, map[string]interface{}{"data": resp}, http.StatusCreated)
}

func (h *TenantHandlers) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.errorJSON(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteTenant(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete tenant", zap.String("id", idStr), zap.Error(err))
		h.errorJSON(w, "Failed to delete tenant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListBranches returns a list of branches for the tenant
func (h *TenantHandlers) ListBranches(w http.ResponseWriter, r *http.Request) {
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

// ListCostCenters returns a list of cost centers for the tenant
func (h *TenantHandlers) ListCostCenters(w http.ResponseWriter, r *http.Request) {
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
func (h *TenantHandlers) json(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TenantHandlers) errorJSON(w http.ResponseWriter, message string, status int) {
	h.json(w, map[string]string{"error": message}, status)
}
