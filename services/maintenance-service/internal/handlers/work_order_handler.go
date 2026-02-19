package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/flyfleet/maintenance-service/internal/models"
	"github.com/flyfleet/maintenance-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WorkOrderHandler handles HTTP requests for work orders
type WorkOrderHandler struct {
	service *service.WorkOrderService
	logger  *zap.Logger
}

// NewWorkOrderHandler creates a new work order handler
func NewWorkOrderHandler(svc *service.WorkOrderService, logger *zap.Logger) *WorkOrderHandler {
	return &WorkOrderHandler{
		service: svc,
		logger:  logger,
	}
}

// Helper methods

func (h *WorkOrderHandler) decodeJSON(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

func (h *WorkOrderHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			h.logger.Error("failed to encode JSON response", zap.Error(err))
		}
	}
}

func (h *WorkOrderHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	h.writeJSONResponse(w, status, map[string]string{"error": message})
}

// ListWorkOrders godoc
func (h *WorkOrderHandler) ListWorkOrders(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)

	query := r.URL.Query()
	filter := models.WorkOrderFilter{
		OrganizationID: orgID,
		Search:         query.Get("search"),
		SortBy:         query.Get("sort_by"),
		SortDir:        query.Get("sort_dir"),
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortDir == "" {
		filter.SortDir = "desc"
	}

	// Parse status filter
	if statuses := query["status"]; len(statuses) > 0 {
		for _, s := range statuses {
			filter.Status = append(filter.Status, models.WorkOrderStatus(s))
		}
	}

	// Parse priority filter
	if priorities := query["priority"]; len(priorities) > 0 {
		for _, p := range priorities {
			filter.Priority = append(filter.Priority, models.WorkOrderPriority(p))
		}
	}

	// Parse pagination
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

	result, err := h.service.ListWorkOrders(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list work orders", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// CreateWorkOrder godoc
func (h *WorkOrderHandler) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)
	userID := getUserID(r)

	var req models.CreateWorkOrderRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	wo, err := h.service.CreateWorkOrder(r.Context(), orgID, userID, &req)
	if err != nil {
		h.logger.Error("failed to create work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, wo)
}

// GetWorkOrder godoc
func (h *WorkOrderHandler) GetWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	wo, err := h.service.GetWorkOrder(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if wo == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "work order not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, wo)
}

// UpdateWorkOrder godoc
func (h *WorkOrderHandler) UpdateWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	var req models.UpdateWorkOrderRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	wo, err := h.service.UpdateWorkOrder(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to update work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if wo == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "work order not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, wo)
}

// ChangeStatus godoc
func (h *WorkOrderHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	var req models.ChangeStatusRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.ChangeStatus(r.Context(), id, req.Status); err != nil {
		h.logger.Error("failed to change status", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "status updated"})
}

// CloseWorkOrder godoc
func (h *WorkOrderHandler) CloseWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	if err := h.service.CloseWorkOrder(r.Context(), id); err != nil {
		h.logger.Error("failed to close work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "work order closed"})
}

// GetWorkOrderParts godoc
func (h *WorkOrderHandler) GetWorkOrderParts(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	wo, err := h.service.GetWorkOrder(r.Context(), id)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if wo == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "work order not found")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, wo.Parts)
}

// AssignPart godoc
func (h *WorkOrderHandler) AssignPart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	var req models.AssignPartRequest
	if err := h.decodeJSON(r.Body, &req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	part, err := h.service.AssignPart(r.Context(), id, &req)
	if err != nil {
		h.logger.Error("failed to assign part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, part)
}

// RemovePart godoc
func (h *WorkOrderHandler) RemovePart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID")
		return
	}

	partID, err := uuid.Parse(chi.URLParam(r, "partId"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID")
		return
	}

	if err := h.service.RemovePart(r.Context(), id, partID); err != nil {
		h.logger.Error("failed to remove part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "part removed"})
}

// GetStats godoc
func (h *WorkOrderHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	orgID := getOrgID(r)

	stats, err := h.service.GetStats(r.Context(), orgID)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}
