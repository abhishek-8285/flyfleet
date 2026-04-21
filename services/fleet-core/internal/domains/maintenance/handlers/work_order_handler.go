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
	"github.com/flyfleet/fleet-core/internal/domains/maintenance/service"
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

func (h *WorkOrderHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.ListWorkOrders)
	r.Post("/", h.CreateWorkOrder)
	r.Get("/stats", h.GetStats)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.GetWorkOrder)
		r.Put("/", h.UpdateWorkOrder)
		r.Patch("/status", h.ChangeStatus)
		r.Patch("/approve", h.ApproveWorkOrder)
		r.Patch("/reject", h.RejectWorkOrder)
		r.Post("/close", h.CloseWorkOrder)
		r.Route("/parts", func(r chi.Router) {
			r.Get("/", h.GetWorkOrderParts)
			r.Post("/", h.AssignPart)
			r.Delete("/{partId}", h.RemovePart)
		})
	})
}

// ListWorkOrders godoc
func (h *WorkOrderHandler) ListWorkOrders(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	query := r.URL.Query()
	filter := models.WorkOrderFilter{
		TenantID:   tenantID,
		Search:  query.Get("search"),
		SortBy:  query.Get("sort_by"),
		SortDir: query.Get("sort_dir"),
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

	result, err := h.service.ListWorkOrders(r.Context(), filter)
	if err != nil {
		h.logger.Error("failed to list work orders", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// CreateWorkOrder godoc
func (h *WorkOrderHandler) CreateWorkOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "user ID required", nil)
		return
	}

	var req models.CreateWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	wo, err := h.service.CreateWorkOrder(r.Context(), tenantID, userID, &req)
	if err != nil {
		h.logger.Error("failed to create work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, wo)
}

// GetWorkOrder godoc
func (h *WorkOrderHandler) GetWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	wo, err := h.service.GetWorkOrder(r.Context(), id, tenantID)
	if err != nil {
		h.logger.Error("failed to get work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if wo == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "work order not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, wo)
}

// UpdateWorkOrder godoc
func (h *WorkOrderHandler) UpdateWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	var req models.UpdateWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	wo, err := h.service.UpdateWorkOrder(r.Context(), id, tenantID, &req)
	if err != nil {
		h.logger.Error("failed to update work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if wo == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "work order not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, wo)
}

// ChangeStatus godoc
func (h *WorkOrderHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	var req models.ChangeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.service.ChangeStatus(r.Context(), id, tenantID, req.Status); err != nil {
		h.logger.Error("failed to change status", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "status updated"})
}

// CloseWorkOrder handles closing a work order
func (h *WorkOrderHandler) CloseWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	if err := h.service.CloseWorkOrder(r.Context(), id, tenantID); err != nil {
		h.logger.Error("failed to close work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "work order closed"})
}

// ApproveWorkOrder handles work order approval
func (h *WorkOrderHandler) ApproveWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "user ID required", nil)
		return
	}

	role := h.getUserRole(r)

	if err := h.service.ApproveWorkOrder(r.Context(), id, tenantID, userID, role); err != nil {
		h.logger.Error("failed to approve work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "work order approved"})
}

// RejectWorkOrder handles work order rejection
func (h *WorkOrderHandler) RejectWorkOrder(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	userID := h.getUserID(r)
	if userID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "user ID required", nil)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "reason is required", nil)
		return
	}

	if err := h.service.RejectWorkOrder(r.Context(), id, tenantID, userID, req.Reason); err != nil {
		h.logger.Error("failed to reject work order", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "work order rejected"})
}

// GetWorkOrderParts godoc
func (h *WorkOrderHandler) GetWorkOrderParts(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	wo, err := h.service.GetWorkOrder(r.Context(), id, tenantID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if wo == nil {
		h.writeErrorResponse(w, http.StatusNotFound, "work order not found", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, wo.Parts)
}

// AssignPart godoc
func (h *WorkOrderHandler) AssignPart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	var req models.AssignPartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	part, err := h.service.AssignPart(r.Context(), id, tenantID, &req)
	if err != nil {
		h.logger.Error("failed to assign part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, part)
}

// RemovePart godoc
func (h *WorkOrderHandler) RemovePart(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid work order ID", nil)
		return
	}

	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	partID, err := uuid.Parse(chi.URLParam(r, "partId"))
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid part ID", nil)
		return
	}

	if err := h.service.RemovePart(r.Context(), id, tenantID, partID); err != nil {
		h.logger.Error("failed to remove part", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "part removed"})
}

// GetStats godoc
func (h *WorkOrderHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID := h.getTenantID(r)
	if tenantID == uuid.Nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "organization ID required", nil)
		return
	}

	stats, err := h.service.GetStats(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// Helper methods

func (h *WorkOrderHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", zap.Error(err))
	}
}

func (h *WorkOrderHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	resp := map[string]interface{}{
		"error": message,
	}
	if err != nil && statusCode == http.StatusInternalServerError {
		resp["details"] = err.Error()
	}
	h.writeJSONResponse(w, statusCode, resp)
}

func (h *WorkOrderHandler) getTenantID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.TenantIDKey); val != nil {
		if tenantIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(tenantIDStr); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed insecure fallback to prevent unauthorized access
	// The org ID must come from validated JWT context only
	h.logger.Warn("Missing tenant_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}

func (h *WorkOrderHandler) getUserID(r *http.Request) uuid.UUID {
	if val := r.Context().Value(middleware.UserIDKey); val != nil {
		if userIDStr, ok := val.(string); ok {
			if id, err := uuid.Parse(userIDStr); err == nil {
				return id
			}
		}
	}
	// SECURITY FIX: Removed insecure default - user ID must come from validated JWT context
	h.logger.Warn("Missing user_id in context - request likely bypassed authentication middleware")
	return uuid.Nil
}

func (h *WorkOrderHandler) getUserRole(r *http.Request) string {
	if val := r.Context().Value(middleware.RoleKey); val != nil {
		if role, ok := val.(string); ok {
			return role
		}
	}
	// SECURITY FIX: Removed insecure default - role must come from validated JWT context
	h.logger.Warn("Missing role in context - request likely bypassed authentication middleware")
	return ""
}
