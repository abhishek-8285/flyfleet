package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/flyfleet/maintenance-service/internal/models"
	"github.com/flyfleet/maintenance-service/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

// Helper methods

func (h *InventoryHandler) decodeJSON(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

func (h *InventoryHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("failed to encode JSON response", zap.Error(err))
		}
	}
}

func (h *InventoryHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}

// ListInventory godoc
func (h *InventoryHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)

	query := r.URL.Query()
	filter := models.InventoryFilter{
		OrganizationID: orgID,
		Search:         query.Get("search"),
		StockLocation:  query.Get("stock_location"),
	}

	filter.LowStockOnly, _ = strconv.ParseBool(query.Get("low_stock_only"))
	filter.AvailableOnly, _ = strconv.ParseBool(query.Get("available_only"))
	
	filter.Page = 1
	if p := query.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			filter.Page = v
		}
	}
	
	filter.PageSize = 20
	if ps := query.Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			filter.PageSize = v
		}
	}

	parts, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list inventory", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
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
	orgID := getOrgID(r)

	var req models.CreateInventoryPartRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	part := &models.InventoryPart{
		OrganizationID:    orgID,
		PartNumber:        req.PartNumber,
		Name:              req.Name,
		Description:       req.Description,
		StockLocation:     req.StockLocation,
		QuantityAvailable: req.QuantityAvailable,
		ReorderLevel:      req.ReorderLevel,
		UnitCost:          req.UnitCost,
	}

	if err := h.repo.Create(r.Context(), part); err != nil {
		h.logger.Error("failed to create part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, part)
}

// GetPart godoc
func (h *InventoryHandler) GetPart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID")
		return
	}

	part, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if part == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "part not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, part)
}

// UpdatePart godoc
func (h *InventoryHandler) UpdatePart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID")
		return
	}

	part, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if part == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "part not found")
		return
	}

	var req models.UpdateInventoryPartRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
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
		part.Description = *req.Description
	}
	if req.StockLocation != nil {
		part.StockLocation = *req.StockLocation
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
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, part)
}

// DeletePart godoc
func (h *InventoryHandler) DeletePart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.logger.Error("failed to delete part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
