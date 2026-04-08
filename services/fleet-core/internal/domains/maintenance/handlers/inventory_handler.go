package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fleet-platform/shared/pkg/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/flyfleet/fleet-core/internal/domains/maintenance/models"
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/repository"
)

// InventoryHandler handles HTTP requests for inventory
type InventoryHandler struct {
	repo   *repository.InventoryRepository
	logger *zap.Logger
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(repo *repository.InventoryRepository, logger *zap.Logger) *InventoryHandler {
	return &InventoryHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *InventoryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListInventory)
	r.Post("/", h.CreatePart)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetPart)
		r.Put("/", h.UpdatePart)
		r.Delete("/", h.DeletePart)
	})
}

// ListInventory godoc
func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	query := r.URL.Query()
	filter := models.InventoryFilter{
		OrgID:         orgID,
		Search:        query.Get("search"),
		StockLocation: query.Get("stock_location"),
	}

	filter.LowStockOnly, _ = strconv.ParseBool(query.Get("low_stock_only"))
	filter.AvailableOnly, _ = strconv.ParseBool(query.Get("available_only"))

	if page := query.Get("page"); page != "" {
		filter.Page, _ = strconv.Atoi(page)
	} else {
		filter.Page = 1
	}
	if pageSize := query.Get("page_size"); pageSize != "" {
		filter.PageSize, _ = strconv.Atoi(pageSize)
	} else {
		filter.PageSize = 20
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	parts, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list inventory", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, models.InventoryListResponse{
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		Parts:    parts,
	})
}

// CreatePart godoc
func (h *InventoryHandler) CreatePart(w http.ResponseWriter, r *http.Request) {
	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	var req models.CreateInventoryPartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	part := &models.InventoryPart{
		OrgID:             orgID,
		PartNumber:        req.PartNumber,
		Name:              req.Name,
		Description:       &req.Description,
		StockLocation:     &req.StockLocation,
		QuantityAvailable: req.QuantityAvailable,
		ReorderLevel:      req.ReorderLevel,
		UnitCost:          req.UnitCost,
	}

	if err := h.repo.Create(r.Context(), part); err != nil {
		h.logger.Error("failed to create part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, part)
}

// GetPart godoc
func (h *InventoryHandler) GetPart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID", nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	part, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if part == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "part not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, part)
}

// UpdatePart godoc
func (h *InventoryHandler) UpdatePart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID", nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	part, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if part == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "part not found", nil)
		return
	}

	var req models.UpdateInventoryPartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Apply updates
	if req.PartNumber != nil {
		part.PartNumber = *req.PartNumber
	}
	if req.Name != nil {
		part.Name = *req.Name
	}
	if req.Description != nil {
		part.Description = req.Description
	}
	if req.StockLocation != nil {
		part.StockLocation = req.StockLocation
	}
	if req.QuantityAvailable != nil {
		part.QuantityAvailable = *req.QuantityAvailable
	}
	if req.ReorderLevel != nil {
		part.ReorderLevel = *req.ReorderLevel
	}
	if req.UnitCost != nil {
		part.UnitCost = *req.UnitCost
	}

	if err := h.repo.Update(r.Context(), part); err != nil {
		h.logger.Error("failed to update part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, part)
}

// DeletePart godoc
func (h *InventoryHandler) DeletePart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID", nil)
		return
	}

	orgID := h.getOrgID(r)
	if orgID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.logger.Error("failed to delete part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *InventoryHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *InventoryHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *InventoryHandler) getOrgID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.OrgIDKey); val != nil {
		if orgIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(orgIDStr); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed insecure fallback to prevent unauthorized access
	// The org ID must come from validated JWT context only
	h.logger.Warn("Missing org_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}
