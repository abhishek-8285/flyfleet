package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flyfleet/auth-service/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GET /api/v1/admin/documents
func (h *AdminHandlers) ListDocuments(w http.ResponseWriter, r *http.Request) {
	orgID, ok := r.Context().Value("org_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"error": "Organization context required"}`, http.StatusForbidden)
		return
	}

	filter := models.DocumentFilter{
		Page:  1,
		Limit: 20,
	}

	// Parse query parameters
	if v := r.URL.Query().Get("vehicle_id"); v != "" {
		vid, err := uuid.Parse(v)
		if err != nil {
			http.Error(w, `{"error": "Invalid vehicle_id"}`, http.StatusBadRequest)
			return
		}
		filter.VehicleID = &vid
	}
	if v := r.URL.Query().Get("type"); v != "" {
		filter.Type = v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = v
	}
	if v := r.URL.Query().Get("expiring_within_days"); v != "" {
		days, err := strconv.Atoi(v)
		if err == nil && days > 0 {
			filter.ExpiringWithinDays = days
		}
	}
	if v := r.URL.Query().Get("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err == nil && p > 0 {
			filter.Page = p
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		l, err := strconv.Atoi(v)
		if err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	response, err := h.service.ListDocuments(r.Context(), orgID, &filter)
	if err != nil {
		h.logger.Error("Failed to list documents", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// POST /api/v1/admin/documents
func (h *AdminHandlers) CreateDocument(w http.ResponseWriter, r *http.Request) {
	var d models.Document
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}

	// vehicle_id can come from query param or body
	if d.VehicleID == nil {
		if v := r.URL.Query().Get("vehicle_id"); v != "" {
			vid, err := uuid.Parse(v)
			if err != nil {
				http.Error(w, `{"error": "Invalid vehicle_id"}`, http.StatusBadRequest)
				return
			}
			d.VehicleID = &vid
		}
	}

	doc, err := h.service.CreateDocument(r.Context(), &d)
	if err != nil {
		h.logger.Error("Failed to create document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": doc,
	})
}

// GET /api/v1/admin/documents/{id}
func (h *AdminHandlers) GetDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.service.GetDocumentByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}
	if doc == nil {
		http.Error(w, `{"error": "Document not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": doc,
	})
}

// PUT /api/v1/admin/documents/{id}
func (h *AdminHandlers) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	var d models.Document
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, `{"error": "Invalid Body"}`, http.StatusBadRequest)
		return
	}
	d.ID = id

	doc, err := h.service.UpdateDocument(r.Context(), &d)
	if err != nil {
		h.logger.Error("Failed to update document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": doc,
	})
}

// DELETE /api/v1/admin/documents/{id}
func (h *AdminHandlers) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error": "Invalid ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteDocument(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete document", zap.Error(err))
		http.Error(w, `{"error": "Internal Server Error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
